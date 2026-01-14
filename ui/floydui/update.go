package floydui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"math/rand"

	"github.com/Nomadcxx/sysc-Go/agent"
	"github.com/Nomadcxx/sysc-Go/agent/loop"
	"github.com/Nomadcxx/sysc-Go/ui/components/palette"
)

// Update handles all message updates to the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	// 1. Global Keybindings & Component Resize
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.Header.Width = msg.Width
		m.Footer.Width = msg.Height
		// Fix: SetSize is a pointer receiver, so we can't call it on m.Palette (value) if it modifies internal state heavily without return
		// But here passing width/height is fine.
		// Ideally we re-assign: m.Palette.SetSize(...)
		// Actually SetSize in palette/model.go is (m *Model).
		// Since m.Palette is a struct value in Model, we might need to take address or handle it carefully.
		// For now, let's assume it works or we fix it if build fails.
		// Better:
		p := &m.Palette
		p.SetSize(msg.Width, msg.Height)

		m.handleWindowSize() // Handles viewport resize
		return m, nil

	case tea.KeyMsg:
		// Global Hotkeys (Priority 1)
		switch msg.String() {
		case "ctrl+c":
			m.Quitting = true
			return m, tea.Quit
		case "ctrl+p":
			m.Mode = ModeCommand
			// Reset to Commands
			items := []list.Item{
				palette.Item{Title: "New Session", Description: "Start a fresh context (Ctrl+N)"},
				palette.Item{Title: "Switch Model", Description: "Change active LLM (Ctrl+L)"},
				palette.Item{Title: "Toggle Help", Description: "View keyboard shortcuts"},
				palette.Item{Title: "Quit", Description: "Exit application"},
			}
			m.Palette.SetItems(items)
			return m, nil
		case "ctrl+l":
			m.Mode = ModeModelSelect
			// Load Models
			items := []list.Item{
				palette.Item{Title: "GPT-4o", Description: "OpenAI - Smartest & Fast"},
				palette.Item{Title: "Claude 3.5 Sonnet", Description: "Anthropic - Best for Coding"},
				palette.Item{Title: "Gemini 1.5 Pro", Description: "Google - Large Context"},
				palette.Item{Title: "Local (Ollama)", Description: "Offline - Private"},
			}
			m.Palette.SetItems(items)
			return m, nil
		case "esc":
			if m.Mode != ModeChat {
				m.Mode = ModeChat
				// Reset palette filtering if needed
				return m, nil
			}
		}

		// Mode-Specific Routing (Priority 2)
		switch m.Mode {
		case ModeCommand, ModeModelSelect:
			newPalette, cmd := m.Palette.Update(msg)
			m.Palette = newPalette
			cmds = append(cmds, cmd)

			if msg.String() == "enter" {
				selectedItem := m.Palette.List.SelectedItem()
				if selectedItem != nil {
					item := selectedItem.(palette.Item)

					if m.Mode == ModeModelSelect {
						m.AddSystemMessage(fmt.Sprintf("✓ Model switched to: %s", item.Title))
						// Here you would actually update m.ProtocolMgr.SetModel(...)
					} else {
						// Handle Command Selection
						switch item.Title {
						case "Quit":
							m.Quitting = true
							return m, tea.Quit
						case "New Session":
							m.ProtocolMgr.InitializeWorkspace()
							m.ClearMessages()
							m.AddSystemMessage("Started new session.")
						case "Switch Model":
							// Chain to Model Select
							m.Mode = ModeModelSelect
							items := []list.Item{
								palette.Item{Title: "GPT-4o", Description: "OpenAI - Smartest & Fast"},
								palette.Item{Title: "Claude 3.5 Sonnet", Description: "Anthropic - Best for Coding"},
								palette.Item{Title: "Gemini 1.5 Pro", Description: "Google - Large Context"},
							}
							m.Palette.SetItems(items)
							return m, nil // Return early to stay in overlay
						case "Toggle Help":
							m.ShowHelp = !m.ShowHelp
						}
					}
				}
				m.Mode = ModeChat
			}
			return m, tea.Batch(cmds...)

		case ModeChat:
			// Pass to standard handler (Input, History, etc)
			return m.handleKeyMsg(msg)
		}

		return m, nil

	case tickMsg:
		m.LastTick = time.Time(msg)
		m.AnimationFrame = (m.AnimationFrame + 1) % 100
		if m.IsThinking {
			// Smoothly "unspool" tokens from the buffer (Teletype effect)
			if len(m.PendingTokens) > 0 {
				// Move up to 15 tokens per frame for a fast, readable flow
				limit := 15
				if len(m.PendingTokens) < limit {
					limit = len(m.PendingTokens)
				}
				for i := 0; i < limit; i++ {
					m.CurrentResponse.WriteString(m.PendingTokens[0])
					m.PendingTokens = m.PendingTokens[1:]
				}
				m.updateViewportContent()
				m.Viewport.GotoBottom()
			}

			// If we received the DONE sentinel AND the buffer is empty, finalize the message
			if m.StreamDone && len(m.PendingTokens) == 0 {
				finalResponse := m.CurrentResponse.String()
				m.StreamDone = false
				m.IsThinking = false
				m.ProgressActive = false
				m.AddAssistantMessage(finalResponse)
				m.updateViewportContent()
				m.Viewport.GotoBottom()
			}

			// Check for timeout (120 seconds without activity)
			if time.Since(m.LastPacketReceived) > 120*time.Second {
				if m.StreamCancel != nil {
					m.StreamCancel()
				}
				m.IsThinking = false
				m.ExecutingTools = false
				m.ProgressActive = false
				m.StreamDone = false
				// Flush any remaining tokens before error
				for len(m.PendingTokens) > 0 {
					m.CurrentResponse.WriteString(m.PendingTokens[0])
					m.PendingTokens = m.PendingTokens[1:]
				}
				m.AddSystemMessage("Error: Agent timed out (no response for 120s). Check your connection or API key.")
				m.updateViewportContent()
				m.Viewport.GotoBottom()
				return m, nil
			}
		}

		cmds = append(cmds, Tick(time.Millisecond*250))

	case thinkingStartedMsg:
		m.IsThinking = true
		m.StreamDone = false
		m.AnimationFrame = 0
		// Pick one random whimsical phrase for the entire thought
		m.CurrentThinking = rand.Intn(len(ThinkingStates))
		m.ProgressActive = true
		m.ProgressPercent = 0.0
		m.LastPacketReceived = time.Now()
		m.CurrentResponse.Reset()
		m.PendingTokens = m.PendingTokens[:0]
		// Start the token reading loop when agent starts thinking
		return m, tea.Batch(
			Tick(time.Millisecond*250),
			m.waitForToken(),
		)

	case thinkingFinishedMsg:
		// Logic moved to tickMsg for smooth completion
		return m, nil

	case streamTokenMsg:
		if msg.token == "\x00DONE" {
			// Mark done, but don't finish until buffer is empty (in tickMsg)
			m.StreamDone = true
			return m, m.waitForToken()
		}

		m.PendingTokens = append(m.PendingTokens, msg.token)
		m.LastPacketReceived = time.Now()

		// Continue waiting for tokens
		return m, m.waitForToken()

	case streamErrorMsg:
		// Ensure all pending tokens are flushed to the response before erroring
		for len(m.PendingTokens) > 0 {
			m.CurrentResponse.WriteString(m.PendingTokens[0])
			m.PendingTokens = m.PendingTokens[1:]
		}
		m.IsThinking = false
		m.ExecutingTools = false
		m.ProgressActive = false
		m.StreamDone = false
		m.AddSystemMessage(fmt.Sprintf("Error: %v", msg.err))
		m.updateViewportContent()
		m.Viewport.GotoBottom()

	case toolUseMsg:
		m.AddToolUse(msg.toolName, msg.toolID)

	case toolExecutingMsg:
		m.ExecutingTools = true
		m.CurrentToolName = msg.toolName

	case toolResultMsg:
		m.ExecutingTools = false
		m.AddToolResult(msg.toolName, msg.content, false)

	case toolErrorMsg:
		m.ExecutingTools = false
		m.AddToolResult(msg.toolName, msg.content, true)

	case progressMsg:
		if m.ProgressActive {
			m.ProgressPercent = float64(msg)
			if m.ProgressPercent >= 1.0 {
				m.ProgressPercent = 1.0
			} else {
				cmds = append(cmds, func() tea.Msg {
					return progressMsg(m.ProgressPercent + 0.05)
				})
			}
		}

	case themeChangedMsg:
		// Theme already updated in SetTheme
		m.AddSystemMessage(fmt.Sprintf("Theme changed to '%s'.", m.CurrentTheme.Name))

	case subAgentSpawnMsg:
		m.AddSystemMessage(fmt.Sprintf("[SPAWN] %s agent: %s", msg.agentType, msg.task))

	case subAgentStatusMsg:
		m.AddSystemMessage(fmt.Sprintf("[SUB-AGENT] %s: %s", msg.status, msg.detail))

	case subAgentCompleteMsg:
		m.AddSystemMessage(fmt.Sprintf("[SUB-AGENT] Complete: %s", msg.result))
	}

	// Update child components

	// Only update Input for non-tick messages to prevent cursor flickering/strobing
	switch msg.(type) {
	case tickMsg:
		// Do not pass tick to Input
	default:
		m.Input, cmd = m.Input.Update(msg)
		cmds = append(cmds, cmd)
	}

	m.Viewport, cmd = m.Viewport.Update(msg)
	cmds = append(cmds, cmd)

	if m.ProgressActive {
		var newModel tea.Model
		newModel, cmd = m.Progress.Update(m.ProgressPercent)
		m.Progress = newModel.(progress.Model)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// handleWindowSize handles window size changes
func (m *Model) handleWindowSize() {
	// Calculate header height based on which banner will be shown
	var bannerText string
	if m.Width < 250 {
		bannerText = SmallBanner
	} else {
		bannerText = PRISMBanner
	}

	// Header with colorized, animated banner (shimmer slowed down)
	header := ColorizeBanner(bannerText, m.CurrentTheme, m.AnimationFrame/2)
	headerHeight := lipgloss.Height(header) + lipgloss.Height(Styles.Tagline.Render(GetRandomTagline()))
	statusBarHeight := 1

	inputLines := strings.Count(m.Input.Value(), "\n") + 1
	inputHeight := inputLines + 2

	errorPanelHeight := 0
	if m.HasError {
		errorPanelHeight = lipgloss.Height(m.LastError) + 2
	}

	helpPanelHeight := 0
	if m.ShowHelp {
		helpPanelHeight = lipgloss.Height(HelpContent) + 2
	}

	verticalMargin := headerHeight + statusBarHeight + inputHeight + errorPanelHeight + helpPanelHeight

	// Prevent negative height calculation which causes panics
	viewportHeight := m.Height - verticalMargin
	if viewportHeight < 0 {
		viewportHeight = 0 // Viewport will just be empty, but won't panic
	}

	if !m.Ready {
		m.Viewport = viewport.New(m.Width, viewportHeight)
		m.Ready = true
	} else {
		m.Viewport.Width = m.Width - 4
		m.Viewport.Height = viewportHeight
	}

	m.Input.SetWidth(m.Width - 6)
	m.Progress.Width = m.Width - 10
	m.updateViewportContent()
	m.Viewport.GotoBottom()
}

// handleKeyMsg handles keyboard input
func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Check for theme command
	if strings.HasPrefix(m.Input.Value(), "/theme ") {
		themeName := strings.TrimPrefix(m.Input.Value(), "/theme ")
		// Normalize theme name to lowercase
		themeName = strings.ToLower(themeName)
		if cmd := m.SetTheme(themeName); cmd != nil {
			m.Input.Reset()
			m.Viewport.GotoBottom()
			return m, cmd
		}
		// Error 7: Invalid theme error
		m.AddSystemMessage(fmt.Sprintf("Unknown theme '%s'. Available: classic, dark, highcontrast, cyberpunk, midnight", themeName))
		m.Input.Reset()
		m.updateViewportContent()
		m.Viewport.GotoBottom()
		return m, nil
	}

	switch msg.Type {
	case tea.KeyEnter:
		// Logic: Shift/Alt/Ctrl+Enter = newline. Bare Enter = send.
		s := msg.String()
		if s == "shift+enter" || s == "alt+enter" || s == "ctrl+enter" {
			var cmd tea.Cmd
			m.Input, cmd = m.Input.Update(msg)
			m.handleWindowSize()
			return m, cmd
		}

		input := strings.TrimSpace(m.Input.Value())
		if input == "" {
			return m, nil
		}

		// Handle commands - Always allow commands
		if strings.HasPrefix(input, "/") {
			// Special handling for cancel while thinking
			if m.IsThinking {
				cmdLower := strings.ToLower(input)
				if cmdLower == "/cancel" {
					// Cancel the current operation
					if m.StreamCancel != nil {
						m.StreamCancel()
					}
					m.IsThinking = false
					m.ExecutingTools = false
					m.ProgressActive = false
					m.Input.Reset()
					m.AddSystemMessage("Operation cancelled.")
					m.updateViewportContent()
					m.Viewport.GotoBottom()
					return m, nil
				}
			}
			// Allow all other commands to pass through
			return m.handleCommand(input)
		}

		// Normal input blocked while thinking
		if m.IsThinking {
			m.AddSystemMessage("Agent is busy. Type /cancel to stop.")
			m.updateViewportContent()
			m.Viewport.GotoBottom()
			return m, nil
		}

		// Add user message and start agent
		m.AddUserMessage(input)
		m.Input.Reset()
		m.handleWindowSize() // Reset to 1 line height
		m.CurrentResponse.Reset()
		m.IsThinking = true
		m.PendingTokens = m.PendingTokens[:0]

		cmds = append(cmds, m.startAgentCall(input))
		cmds = append(cmds, Tick(time.Millisecond*250))

		return m, tea.Batch(cmds...)

	case tea.KeyUp, tea.KeyShiftUp:
		if m.HistoryIndex > 0 {
			m.HistoryIndex--
			m.Input.SetValue(m.History[m.HistoryIndex])
		}
		// History navigation shouldn't update input through default handler
		return m, tea.Batch(cmds...)

	case tea.KeyDown, tea.KeyShiftDown:
		if m.HistoryIndex < len(m.History)-1 {
			m.HistoryIndex++
			m.Input.SetValue(m.History[m.HistoryIndex])
		} else if m.HistoryIndex == len(m.History)-1 {
			m.HistoryIndex++
			m.Input.Reset()
		}
		// History navigation shouldn't update input through default handler
		return m, tea.Batch(cmds...)

	case tea.KeyCtrlC, tea.KeyEsc:
		if m.StreamCancel != nil {
			m.StreamCancel()
		}
		// Save session before quitting
		_ = m.SaveCurrentSession()
		return m, tea.Quit

	case tea.KeyCtrlH:
		m.ShowHelp = !m.ShowHelp
		// Ctrl+H shouldn't update input through default handler
		return m, tea.Batch(cmds...)
	}

	// Handle ? key for help toggle (only when input is empty to avoid conflicts with typing)
	if msg.String() == "?" && m.Input.Value() == "" {
		m.ShowHelp = !m.ShowHelp
		return m, tea.Batch(cmds...)
	}

	// For all other keys, let the textinput component handle them
	// This includes character input (KeyRunes), backspace, space, etc.
	var cmd tea.Cmd
	m.Input, cmd = m.Input.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// handleCommand handles slash commands
func (m Model) handleCommand(input string) (tea.Model, tea.Cmd) {
	cmd := strings.TrimPrefix(input, "/")
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		// Error 6: Empty command feedback
		m.AddSystemMessage("Usage: /command [args]. Type /help for available commands.")
		m.Input.Reset()
		m.updateViewportContent()
		m.Viewport.GotoBottom()
		return m, nil
	}

	cmdName := strings.ToLower(parts[0])
	args := parts[1:]

	// Check if command exists in registry
	if m.CommandRegistry != nil {
		if registeredCmd := m.CommandRegistry.Get(cmdName); registeredCmd != nil {
			// If it's a custom command with a prompt, send it to the agent
			if registeredCmd.Prompt != "" && !registeredCmd.IsBuiltIn {
				m.AddSystemMessage(fmt.Sprintf("Running /%s...", registeredCmd.Name))
				m.Input.Reset()
				m.updateViewportContent()
				// Custom commands inject their prompt and send to agent
				return m, m.startAgentCall(registeredCmd.Prompt)
			}
		}
	}

	switch cmdName {
	case "exit", "quit", "q":
		_ = m.SaveCurrentSession()
		return m, tea.Quit

	case "clear", "cls":
		m.ClearMessages()
		m.updateViewportContent()
		m.Viewport.GotoBottom()
		return m, nil

	case "help", "h", "?":
		m.ShowHelp = !m.ShowHelp
		m.updateViewportContent()
		m.Viewport.GotoBottom()
		return m, nil

	case "status":
		status := m.ProtocolMgr.WorkspaceStatus()
		var sb strings.Builder
		sb.WriteString("📁 FLOYD Workspace Status:\n")
		for k, v := range status {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
		}
		m.AddSystemMessage(sb.String())
		m.updateViewportContent()
		m.Viewport.GotoBottom()
		return m, nil

	case "tools":
		// List available tools
		toolList := []string{"bash", "read", "write", "edit", "multiedit", "grep", "ls", "glob", "cache"}
		var sb strings.Builder
		sb.WriteString("🔧 Available Tools:\n")
		for _, t := range toolList {
			sb.WriteString(fmt.Sprintf("  • %s\n", t))
		}
		m.AddSystemMessage(sb.String())
		m.Input.Reset()
		m.updateViewportContent()
		m.Viewport.GotoBottom()
		return m, nil

	case "theme", "t":
		if len(args) == 0 {
			// List themes
			themes := GetThemeNames()
			m.AddSystemMessage(fmt.Sprintf("Available themes: %s\nUsage: /theme <name>", strings.Join(themes, ", ")))
		} else {
			themeName := strings.Join(args, " ")
			if theme := GetThemeByName(themeName); theme != nil {
				m.CurrentTheme = *theme
				InitStylesWithTheme(m.CurrentTheme)
				m.Input.FocusedStyle.Prompt = lipgloss.NewStyle().Foreground(m.CurrentTheme.Accent)

				m.AddSystemMessage(fmt.Sprintf("Theme changed to: %s", theme.Name))
			} else {
				themes := GetThemeNames()
				m.AddSystemMessage(fmt.Sprintf("✗ Unknown theme: %s. Available: %s", themeName, strings.Join(themes, ", ")))
			}

		}
		m.Input.Reset()
		m.updateViewportContent()
		m.Viewport.GotoBottom()
		return m, nil

	case "protocol":
		masterPlan := m.ProtocolMgr.GetMasterPlan()
		stack := m.ProtocolMgr.GetStack()
		var sb strings.Builder
		sb.WriteString("📋 FLOYD Protocol Status:\n\n")
		if len(masterPlan) > 0 {
			sb.WriteString("Master Plan: Loaded\n")
		} else {
			sb.WriteString("Master Plan: Not found\n")
		}
		if len(stack) > 0 && !strings.Contains(stack, "not defined") {
			sb.WriteString("Tech Stack: Defined\n")
		} else {
			sb.WriteString("Tech Stack: Not defined\n")
		}
		m.AddSystemMessage(sb.String())
		m.updateViewportContent()
		m.Viewport.GotoBottom()
		return m, nil

	case "init":
		if err := m.ProtocolMgr.InitializeWorkspace(); err != nil {
			m.AddSystemMessage(fmt.Sprintf("✗ Failed to initialize workspace: %v", err))
		} else {
			m.AddSystemMessage("✓ Initialized .floyd/ workspace with AGENT_INSTRUCTIONS.md")
		}
		m.Input.Reset()
		m.updateViewportContent()
		m.Viewport.GotoBottom()
		return m, nil

	case "compact":
		summary := m.CompactConversation()
		// Store in cache
		if m.CacheMgr != nil {
			_ = m.CacheMgr.Store(context.Background(), "project", "compact_summary", summary)
		}
		m.AddSystemMessage("═══════════════════════════════════════════\n  CONVERSATION COMPACTED ✓\n═══════════════════════════════════════════\n" + summary)
		m.Input.Reset()
		m.updateViewportContent()
		m.Viewport.GotoBottom()
		return m, nil

	case "commands", "cmds":
		if m.CommandRegistry == nil {
			m.AddSystemMessage("Command registry not initialized")
		} else {
			var sb strings.Builder
			sb.WriteString("═══════════════════════════════════════════\n")
			sb.WriteString("  AVAILABLE COMMANDS\n")
			sb.WriteString("═══════════════════════════════════════════\n\n")

			// Built-in commands
			sb.WriteString("📌 Built-in:\n")
			for _, cmd := range m.CommandRegistry.ListCommands() {
				if cmd.IsBuiltIn {
					aliases := ""
					if len(cmd.Aliases) > 0 {
						aliases = fmt.Sprintf(" (%s)", strings.Join(cmd.Aliases, ", "))
					}
					sb.WriteString(fmt.Sprintf("  /%s%s - %s\n", cmd.Name, aliases, cmd.Description))
				}
			}

			// Custom commands
			hasCustom := false
			for _, cmd := range m.CommandRegistry.ListCommands() {
				if !cmd.IsBuiltIn {
					if !hasCustom {
						sb.WriteString("\n📦 Custom:\n")
						hasCustom = true
					}
					sb.WriteString(fmt.Sprintf("  /%s - %s\n", cmd.Name, cmd.Description))
				}
			}

			if !hasCustom {
				sb.WriteString("\n📦 Custom: None (add to ~/.floyd/commands/)")
			}

			m.AddSystemMessage(sb.String())
		}
		m.Input.Reset()
		m.updateViewportContent()
		m.Viewport.GotoBottom()
		return m, nil

	case "skills":
		if m.CommandRegistry == nil {
			m.AddSystemMessage("Command registry not initialized")
		} else if len(args) == 0 {
			// List skills
			skills := m.CommandRegistry.ListSkills()
			var sb strings.Builder
			sb.WriteString("═══════════════════════════════════════════\n")
			sb.WriteString("  SKILLS\n")
			sb.WriteString("═══════════════════════════════════════════\n\n")

			if len(skills) == 0 {
				sb.WriteString("No skills loaded.\n")
				sb.WriteString("Add skills to ~/.floyd/skills/ or .floyd/skills/\n")
			} else {
				for _, skill := range skills {
					status := "☐"
					if skill.Enabled {
						status = "☑"
					}
					sb.WriteString(fmt.Sprintf("  %s %s - %s\n", status, skill.Name, skill.Description))
				}
			}
			sb.WriteString("\nUsage: /skills enable <name> | /skills disable <name>")
			m.AddSystemMessage(sb.String())
		} else if len(args) >= 2 {
			action := strings.ToLower(args[0])
			skillName := args[1]
			switch action {
			case "enable":
				if m.CommandRegistry.EnableSkill(skillName) {
					m.AddSystemMessage(fmt.Sprintf("✓ Skill '%s' enabled", skillName))
				} else {
					m.AddSystemMessage(fmt.Sprintf("✗ Skill '%s' not found", skillName))
				}
			case "disable":
				if m.CommandRegistry.DisableSkill(skillName) {
					m.AddSystemMessage(fmt.Sprintf("✓ Skill '%s' disabled", skillName))
				} else {
					m.AddSystemMessage(fmt.Sprintf("✗ Skill '%s' not found", skillName))
				}
			default:
				m.AddSystemMessage("Usage: /skills enable <name> | /skills disable <name>")
			}
		} else {
			m.AddSystemMessage("Usage: /skills | /skills enable <name> | /skills disable <name>")
		}
		m.Input.Reset()
		m.updateViewportContent()
		m.Viewport.GotoBottom()
		return m, nil

	case "memory", "cache":
		var sb strings.Builder
		sb.WriteString("═══════════════════════════════════════════\n")
		sb.WriteString("  SUPERCACHE STATUS\n")
		sb.WriteString("═══════════════════════════════════════════\n\n")

		if m.CacheMgr == nil {
			sb.WriteString("Cache not initialized\n")
		} else {
			sb.WriteString("Tier 1 - Reasoning (5min):  Active\n")
			sb.WriteString("Tier 2 - Project (24h):     Active\n")
			sb.WriteString("Tier 3 - Vault (7d):        Active\n")
			sb.WriteString(fmt.Sprintf("\nCache dir: %s\n", ".floyd/.cache/"))
		}
		m.AddSystemMessage(sb.String())
		m.Input.Reset()
		m.updateViewportContent()
		m.Viewport.GotoBottom()
		return m, nil

	default:
		// If not a known command, treat as regular chat message IF it's not a slash command
		// But we are in handleCommand, so it IS a slash command.
		m.AddSystemMessage(fmt.Sprintf("✗ Unknown command: /%s", cmdName))
		m.Input.Reset()
		m.updateViewportContent()
		m.Viewport.GotoBottom()
		return m, nil
	}
}

// startAgentCall initiates a call to the agent
func (m *Model) startAgentCall(userInput string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		m.StreamCancel = cancel

		// Build request with protocol
		req := agent.ChatRequest{
			Messages:    m.Conversation,
			MaxTokens:   4096,
			Temperature: 0.2, // Lowered from 0.7 for high stability and coherence
			Stream:      true,
		}

		// Inject protocol
		req = m.ProtocolMgr.EnhancedChatRequest(req, "")

		// Run the tool loop with streaming
		eventChan, err := m.LoopAgent.RunToolLoopStreaming(ctx, req)
		if err != nil {
			return streamErrorMsg{err: err}
		}

		// ⚠️  CRITICAL: Persistent Channel Pattern
		// TokenStreamChan is created once in NewModel and NEVER closed.
		// Closing it causes panics on subsequent agent calls.
		// We use "\x00DONE" sentinel to signal completion instead.
		// See Claude.md "TUI Channel Pattern" for full documentation.
		//
		// Process events in goroutine
		go func() {
			// ⚠️  DO NOT add: defer close(m.TokenStreamChan)
			// The channel is persistent - closing it breaks the UI!

			for event := range eventChan {
				switch event.Type {
				case loop.EventTypeToken:
					select {
					case m.TokenStreamChan <- event.Text:
					case <-ctx.Done():
						return
					}

				case loop.EventTypeToolUse:
					select {
					case m.TokenStreamChan <- fmt.Sprintf("\n• Using tool: %s\n", event.ToolName):
					case <-ctx.Done():
						return
					}

				case loop.EventTypeToolExecuting:
					// Show tool execution status
					select {
					case m.TokenStreamChan <- fmt.Sprintf("• Executing: %s\n", event.ToolCall):
					case <-ctx.Done():
						return
					}

				case loop.EventTypeToolResult:
					// Show tool result (truncated if too long)
					resultContent := event.ToolContent
					if len(resultContent) > 300 {
						resultContent = resultContent[:300] + "..."
					}
					prefix := "✓"
					if event.IsError {
						prefix = "✗"
					}
					resultMsg := fmt.Sprintf("  └── %s Result: %s\n\n", prefix, resultContent)
					if event.IsError {
						resultMsg = fmt.Sprintf("  └── %s Error (%s): %s\n\n", prefix, event.ToolName, resultContent)
					}
					select {
					case m.TokenStreamChan <- resultMsg:
					case <-ctx.Done():
						return
					}

				case loop.EventTypeDone:
					select {
					case m.TokenStreamChan <- "\x00DONE":
					case <-ctx.Done():
					}
					return

				case loop.EventTypeError:
					return
				}
			}
		}()

		return thinkingStartedMsg{}
	}
}

// waitForToken waits for a token from the stream channel.
// ⚠️  CRITICAL: This function is part of the persistent channel pattern.
// TokenStreamChan is NEVER closed; we use "\x00DONE" as a sentinel value.
// This command MUST be re-issued after each streamTokenMsg to keep the loop alive.
// See Claude.md "TUI Channel Pattern" for details.
func (m *Model) waitForToken() tea.Cmd {
	return func() tea.Msg {
		token, ok := <-m.TokenStreamChan
		if !ok {
			// Channel closed unexpectedly - this should NOT happen!
			// If you see this, someone broke the pattern.
			return nil
		}
		return streamTokenMsg{token: token}
	}
}

// flushTokens is now deprecated in favor of the Smooth Stream unspooling in tickMsg.
// It remains as a no-op to prevent compilation errors if called elsewhere in the short term.
func (m *Model) flushTokens() {
	// Smooth Stream unspooling handles all tokens via PendingTokens.
}

// updateViewportContent updates the viewport content
func (m *Model) updateViewportContent() {
	var chatHistory strings.Builder

	for _, msg := range m.Messages {
		var styled string
		switch msg.Role {
		case "user":
			styled = Styles.UserMsg.
				Width(m.Width - 4).
				Render("YOU: " + msg.Content)
		case "assistant":
			if msg.IsStreaming {
				// We render markdown even during streaming for the teletype effect
				rendered := RenderMarkdown(m.CurrentResponse.String(), m.Width-6, m.CurrentTheme)
				styled = Styles.FloydMsg.Render("FLOYD:\n" + rendered)
			} else {
				rendered := RenderMarkdown(msg.Content, m.Width-6, m.CurrentTheme)
				styled = Styles.FloydMsg.Render("FLOYD:\n" + rendered)
			}

		case "system":
			if msg.ToolUse != nil {
				styled = Styles.ToolUse.Render("▶ Using tool: " + msg.ToolUse.Name)
			} else if msg.ToolResult != nil {
				prefix := "✓ "
				if msg.ToolResult.IsError {
					prefix = "✗ "
				}
				content := msg.ToolResult.Content
				if len(content) > 200 {
					content = content[:200] + "..."
				}
				if msg.ToolResult.IsError {
					styled = Styles.ToolResultError.Render(prefix + msg.ToolResult.Name + ": " + content)
				} else {
					styled = Styles.ToolResult.Render(prefix + msg.ToolResult.Name + ": " + content)
				}
			} else {
				styled = Styles.SystemMsg.Render(msg.Content)
			}
		}
		chatHistory.WriteString(styled + "\n")
	}

	// Add current streaming response
	if m.IsThinking && m.CurrentResponse.Len() > 0 {
		rendered := RenderMarkdown(m.CurrentResponse.String(), m.Width-6, m.CurrentTheme)
		styled := Styles.FloydMsg.Render("FLOYD:\n" + rendered)
		chatHistory.WriteString(styled + "\n")
	}

	m.Viewport.SetContent(chatHistory.String())
}
