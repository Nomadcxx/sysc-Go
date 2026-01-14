package tui

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Nomadcxx/sysc-Go/tui/floydtools"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const floydHeaderCommand = "syscgo -effect matrix-art -theme dracula -file ~/Volumes/Storage/FLOYD_CLI/assets/floyd2.txt -duration 0"

type floydStreamMsg struct {
	chunk     string
	status    string
	done      bool
	logAction string
	logResult string
	logNext   string
}

const (
	floydMasterPlanTemplate = `# Master Plan & Objectives (FLOYD)
## Primary Goal
[User: Enter the main goal here, e.g., "Build the MVP web app from docs/prd.md"]

## Definition of Done (Near-Complete)
- [ ] App compiles/runs end-to-end
- [ ] Core flows implemented per PRD
- [ ] Tests added/updated where appropriate
- [ ] Lint/typecheck passes (or issues clearly documented)
- [ ] Deployment notes + env vars documented
- [ ] Final output is PR-ready (NO direct main pushes)
- [ ] Handoff checklist written for Douglas (exact commands + what to verify)

## Strategic Steps
- [ ] Phase 1: Stack Definition (Update .floyd/stack.md from PRD)
- [ ] Phase 2: Core Logic Implementation
- [ ] Phase 3: Testing & Refinement
- [ ] Phase 4: PR Packaging (notes, checklist, handoff)

## Context & Constraints
- Hard Rule: NEVER push to main/master
- Preferred Workflow: feature branch -> PR -> Douglas approves merge
`

	floydScratchTemplate = `# Current Thinking & Scratchpad (FLOYD)
Use this file to store temporary code snippets, error logs, investigation notes, decisions, and “things we must remember” so we don’t clog chat context.

## Current Focus
initializing...

## Decisions (persistent)
- (none yet)

## Errors / Warnings Log
(none yet)
`

	floydProgressTemplate = `# Execution Log (FLOYD)
| Timestamp | Action Taken | Result/Status | Next Step |
`

	floydBranchTemplate = `# Working Branch & PR Checklist (FLOYD)

## Working Branch
- Name: [e.g., feat/auth-onboarding]
- Base: main
- Status: [in progress | ready for review]

## PR Checklist (Non-Negotiable)
- [ ] No direct commits to main/master
- [ ] PR branch created and used for all changes
- [ ] Lint/typecheck pass (or documented exceptions)
- [ ] Tests added/updated (or documented)
- [ ] Migration steps documented (if any)
- [ ] Env vars documented
- [ ] How to run locally documented
- [ ] Known issues / TODOs captured
- [ ] PR summary written for Douglas
`

	floydStackTemplate = `# Technology Stack (FLOYD)
*To be populated by Agent from the PRD/Blueprint immediately upon start.*

## Core Frameworks
- [ ]

## Database / Auth
- [ ]

## Styling / UI
- [ ]

## DevOps / Infra
- [ ]
`

	floydAgentInstructions = `<floyd_protocol>
  <identity>
    You are FLOYD: professional, capable, concise.
    You orchestrate specialists to build software nearly ready to ship.
  </identity>

  <file_system_authority>
    The .floyd/ directory is persistent memory.
    1) Read .floyd/master_plan.md before acting.
    2) Update checkboxes immediately after completing a task.
    3) Append one-line summaries to .floyd/progress.md after each command or edit.
    4) Populate .floyd/stack.md from the PRD and follow it strictly.
  </file_system_authority>

  <self_correction_loop>
    When a tool fails or output is wrong:
    - Log the error in .floyd/scratchpad.md
    - Update .floyd/master_plan.md with the fix approach
    - Re-run with a tighter plan
    - If lost, run 'floyd status' for bearings.
  </self_correction_loop>

  <safety_shipping_rules>
    <rule id="1">Never push directly to main or master.</rule>
    <rule id="2">Work on a feature branch.</rule>
    <rule id="3">Prepare PR-ready changes only; Douglas approves merges.</rule>
  </safety_shipping_rules>

  <orchestration>
    Spawn planner, implementer, tester, and reviewer specialists as needed.
    Prefer concise checklists, commands, and diffs.
  </orchestration>

  <repo_layout>
    .floyd/ holds control-plane files only. Modify application code in normal repo directories.
  </repo_layout>
</floyd_protocol>
`
)

// FloydToolKind represents a FLOYD tool type (string name for registry lookup).
type FloydToolKind string

const (
	ToolBash     FloydToolKind = "bash"
	ToolRead     FloydToolKind = "read"
	ToolWrite    FloydToolKind = "write"
	ToolLS       FloydToolKind = "ls"
	ToolGlob     FloydToolKind = "glob"
	ToolTemplate FloydToolKind = "template"
)

// FloydTool describes an available tool in Floyd mode.
type FloydTool struct {
	Name        string
	Description string
	Kind        FloydToolKind
	Template    string
}

// defaultFloydTools returns the built-in FLOYD-inspired tools from the registry.
func defaultFloydTools() []FloydTool {
	// Get tools from the floydtools registry
	registryTools := floydtools.AllTools()

	tools := make([]FloydTool, 0, len(registryTools))
	for _, t := range registryTools {
		tools = append(tools, FloydTool{
			Name:        strings.ToUpper(t.Name()), // "bash" -> "Bash" for display
			Description: t.Description(),
			Kind:        FloydToolKind(t.Name()),
		})
	}

	return append(tools, floydChromeTools...)
}

var floydChromeTools = []FloydTool{
	{
		Name:        "Chrome: tabs_context",
		Description: "List all tabs in the current Chrome group.",
		Kind:        ToolTemplate,
		Template: `{
  "name": "tabs_context",
  "input": {}
}`,
	},
	{
		Name:        "Chrome: tabs_create",
		Description: "Create a new empty tab in the active group.",
		Kind:        ToolTemplate,
		Template: `{
  "name": "tabs_create",
  "input": {}
}`,
	},
	{
		Name:        "Chrome: navigate",
		Description: "Navigate a tab to a URL or move in history.",
		Kind:        ToolTemplate,
		Template: `{
  "name": "navigate",
  "input": {
    "url": "https://example.com",
    "tabId": 1
  }
}`,
	},
	{
		Name:        "Chrome: computer",
		Description: "Drive the browser with mouse/keyboard actions.",
		Kind:        ToolTemplate,
		Template: `// Required: action, tabId
// Optional: coordinate, text, duration, scroll_direction, scroll_amount, start_coordinate, region, repeat, ref, modifiers
{
  "name": "computer",
  "input": {
    "action": "screenshot",
    "tabId": 1,
    "coordinate": [100, 200],
    "text": "",
    "duration": 1,
    "scroll_direction": "down",
    "scroll_amount": 3,
    "start_coordinate": [0, 0],
    "region": [0, 0, 1280, 720],
    "repeat": 1,
    "ref": "ref_1",
    "modifiers": ""
  }
}`,
	},
	{
		Name:        "Chrome: find",
		Description: "Locate elements by natural language query.",
		Kind:        ToolTemplate,
		Template: `{
  "name": "find",
  "input": {
    "query": "search bar",
    "tabId": 1
  }
}`,
	},
	{
		Name:        "Chrome: read_page",
		Description: "Fetch an accessibility tree snapshot.",
		Kind:        ToolTemplate,
		Template: `{
  "name": "read_page",
  "input": {
    "tabId": 1,
    "filter": "all",
    "depth": 15,
    "ref_id": ""
  }
}`,
	},
	{
		Name:        "Chrome: form_input",
		Description: "Set the value of a form element.",
		Kind:        ToolTemplate,
		Template: `{
  "name": "form_input",
  "input": {
    "ref": "ref_1",
    "value": "",
    "tabId": 1
  }
}`,
	},
	{
		Name:        "Chrome: get_page_text",
		Description: "Extract readable text from the active page.",
		Kind:        ToolTemplate,
		Template: `{
  "name": "get_page_text",
  "input": {
    "tabId": 1
  }
}`,
	},
	{
		Name:        "Chrome: javascript_tool",
		Description: "Run JavaScript in the page context.",
		Kind:        ToolTemplate,
		Template: `{
  "name": "javascript_tool",
  "input": {
    "action": "javascript_exec",
    "text": "window.location.href",
    "tabId": 1
  }
}`,
	},
	{
		Name:        "Chrome: gif_creator",
		Description: "Record, stop, and export GIFs of browser sessions.",
		Kind:        ToolTemplate,
		Template: `// Actions: start_recording, stop_recording, export, clear
{
  "name": "gif_creator",
  "input": {
    "action": "start_recording",
    "tabId": 1,
    "coordinate": [600, 400],
    "download": false,
    "filename": "recording.gif",
    "options": {
      "showClickIndicators": true,
      "showDragPaths": true,
      "showActionLabels": true,
      "showProgressBar": true,
      "showWatermark": true,
      "quality": 10
    }
  }
}`,
	},
	{
		Name:        "Chrome: read_console_messages",
		Description: "Retrieve console logs from the tab (filter strongly!).",
		Kind:        ToolTemplate,
		Template: `{
  "name": "read_console_messages",
  "input": {
    "tabId": 1,
    "pattern": "error|warning",
    "onlyErrors": false,
    "clear": false,
    "limit": 50
  }
}`,
	},
	{
		Name:        "Chrome: read_network_requests",
		Description: "List recent network requests from the tab.",
		Kind:        ToolTemplate,
		Template: `{
  "name": "read_network_requests",
  "input": {
    "tabId": 1,
    "urlPattern": "/api/",
    "clear": false,
    "limit": 50
  }
}`,
	},
	{
		Name:        "Chrome: resize_window",
		Description: "Resize the Chrome window to specific dimensions.",
		Kind:        ToolTemplate,
		Template: `{
  "name": "resize_window",
  "input": {
    "tabId": 1,
    "width": 1280,
    "height": 720
  }
}`,
	},
	{
		Name:        "Chrome: upload_image",
		Description: "Upload or drag/drop a screenshot into the tab.",
		Kind:        ToolTemplate,
		Template: `{
  "name": "upload_image",
  "input": {
    "imageId": "screenshot_001",
    "tabId": 1,
    "ref": "ref_3",
    "coordinate": [500, 400],
    "filename": "image.png"
  }
}`,
	},
	{
		Name:        "Chrome: update_plan",
		Description: "Outline domains and approach before browsing.",
		Kind:        ToolTemplate,
		Template: `{
  "name": "update_plan",
  "input": {
    "domains": ["example.com"],
    "approach": [
      "Navigate to homepage",
      "Search for documentation",
      "Summarize findings"
    ]
  }
}`,
	},
	{
		Name:        "Chrome: turn_answer_start",
		Description: "Mark the beginning of a user-facing response.",
		Kind:        ToolTemplate,
		Template: `{
  "name": "turn_answer_start",
  "input": {}
}`,
	},
}

func waitForFloydChunk(ch <-chan floydStreamMsg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return floydStreamMsg{done: true}
		}
		return msg
	}
}

func (m Model) ensureFloydViewportSize() Model {
	width := max(32, m.width-6)
	height := max(8, m.height-22)

	if m.floydViewport.Width != width || m.floydViewport.Height != height {
		m.floydViewport.Width = width
		m.floydViewport.Height = height
		m.floydViewport.SetContent(m.floydBuilder.String())
		if m.floydPinned {
			m.floydViewport.GotoBottom()
		}
	}
	return m
}

func (m Model) ensureFloydIntro() Model {
	if m.floydBuilder.Len() == 0 {
		m.floydBuilder.WriteString("FLOYD console ready. Select a tool and press Enter to launch it.\n")
		m.floydViewport.SetContent(m.floydBuilder.String())
		m.floydViewport.GotoBottom()
		m.floydPinned = true
	}
	return m
}

func (m Model) appendFloydChunk(chunk string) Model {
	if chunk == "" {
		return m
	}
	pinned := m.floydPinned || m.floydViewport.AtBottom()
	m.floydBuilder.WriteString(chunk)
	m.floydViewport.SetContent(m.floydBuilder.String())
	if pinned {
		m.floydViewport.GotoBottom()
		m.floydPinned = true
	} else {
		m.floydPinned = m.floydViewport.AtBottom()
	}
	return m
}

func (m Model) appendFloydTemplate(tool FloydTool) Model {
	m = m.ensureFloydIntro()
	label := fmt.Sprintf("\nTemplate: %s\n", tool.Name)
	m = m.appendFloydChunk(label)
	if tool.Template != "" {
		m = m.appendFloydChunk("```json\n" + tool.Template + "\n```\n")
	}
	m.floydStatus = fmt.Sprintf("Template \"%s\" appended. Copy and adapt as needed.", tool.Name)
	return m
}

func (m Model) beginFloydStream(title string, runner func(chan<- floydStreamMsg)) (Model, tea.Cmd) {
	if m.floydStreaming && m.floydStream != nil {
		m.floydStatus = "A task is already running. Please wait for it to finish."
		return m, nil
	}

	m = m.ensureFloydIntro()
	if m.floydBuilder.Len() > 0 {
		m = m.appendFloydChunk("\n")
	}
	stamp := time.Now().Format("15:04:05")
	header := fmt.Sprintf("[%s] %s\n", stamp, title)
	m = m.appendFloydChunk(header)

	ch := make(chan floydStreamMsg)
	m.floydStream = ch
	m.floydStreaming = true
	m.floydStatus = fmt.Sprintf("Running %s ...", title)

	go func() {
		runner(ch)
		close(ch)
	}()

	return m, waitForFloydChunk(ch)
}

func (m Model) startFloydAction(kind FloydToolKind, value string) (Model, tea.Cmd) {
	input := strings.TrimSpace(value)

	// Handle template tools (Chrome tools) - not yet implemented via registry
	if kind == ToolTemplate {
		m.floydStatus = "Template tools require manual implementation."
		return m, nil
	}

	// Look up tool from registry
	tool, err := floydtools.Get(string(kind))
	if err != nil {
		m.floydStatus = fmt.Sprintf("Tool not found: %s", kind)
		return m, nil
	}

	// Validate input
	if err := tool.Validate(input); err != nil {
		m.floydStatus = err.Error()
		return m, nil
	}

	// Special case for LS: default to "."
	if kind == ToolLS && input == "" {
		input = "."
	}

	// Run the tool via the registry
	return m.beginFloydStream(string(kind)+" "+input, adaptToolRunner(tool.Run(input)))
}

// adaptToolRunner converts a floydtools runner to the legacy floydStreamMsg format.
func adaptToolRunner(runner func(chan<- floydtools.StreamMsg)) func(chan<- floydStreamMsg) {
	return func(ch chan<- floydStreamMsg) {
		toolCh := make(chan floydtools.StreamMsg, 10)
		go runner(toolCh)

		for msg := range toolCh {
			ch <- floydStreamMsg{
				chunk:     msg.Chunk,
				status:    msg.Status,
				done:      msg.Done,
				logAction: msg.LogAction,
				logResult: msg.LogResult,
				logNext:   msg.LogNext,
			}
			if msg.Done {
				close(toolCh)
				return
			}
		}
	}
}

func (m Model) handleFloydStream(msg floydStreamMsg) Model {
	if msg.chunk != "" {
		m = m.appendFloydChunk(msg.chunk)
	}
	if msg.status != "" {
		m.floydStatus = msg.status
	}
	if msg.done {
		m.floydStreaming = false
		m.floydStream = nil
		if msg.logAction != "" {
			m.logFloyd(msg.logAction, msg.logResult, msg.logNext)
		}
	}
	return m
}

// prepareFloydMode ensures the workspace exists and snapshots core files.
func (m Model) prepareFloydMode() Model {
	if err := m.ensureFloydState(); err != nil {
		m.floydStatus = fmt.Sprintf("FLOYD setup failed: %v", err)
		return m
	}
	m = m.ensureFloydViewportSize()
	m = m.ensureFloydIntro()
	m.floydViewport.SetContent(m.floydBuilder.String())
	if m.floydPinned {
		m.floydViewport.GotoBottom()
	} else {
		m.floydPinned = m.floydViewport.AtBottom()
	}
	if strings.TrimSpace(m.floydStatus) == "" {
		m.floydStatus = "Select a tool and press Enter to run it."
	}
	return m
}

// ensureFloydState makes sure .floyd/ exists and templates are present.
func (m *Model) ensureFloydState() error {
	dir, err := ensureFloydWorkspace()
	if err != nil {
		return err
	}
	m.floydDir = dir
	m.refreshFloydSnapshot()
	return nil
}

// refreshFloydSnapshot reloads key FLOYD files for display.
func (m *Model) refreshFloydSnapshot() {
	if m.floydDir == "" {
		return
	}
	masterPath := filepath.Join(m.floydDir, "master_plan.md")
	stackPath := filepath.Join(m.floydDir, "stack.md")
	progressPath := filepath.Join(m.floydDir, "progress.md")

	m.floydMasterPlan = readFileOrPlaceholder(masterPath, "(master_plan.md not found)")
	m.floydStack = readFileOrPlaceholder(stackPath, "(stack.md not found)")
	m.floydRecentLog = readLastLines(progressPath, 5)
}

// logFloyd appends to the execution log.
func (m *Model) logFloyd(action, result, next string) {
	if m.floydDir == "" {
		return
	}

	progressPath := filepath.Join(m.floydDir, "progress.md")
	if err := appendProgress(progressPath, action, result, next); err != nil {
		m.floydStatus = fmt.Sprintf("Failed to log progress: %v", err)
		return
	}
	m.refreshFloydSnapshot()
}

func ensureFloydWorkspace() (string, error) {
	root, err := findRepoRoot()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(root, ".floyd")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	files := map[string]string{
		"master_plan.md":        floydMasterPlanTemplate,
		"scratchpad.md":         floydScratchTemplate,
		"progress.md":           floydProgressTemplate,
		"branch.md":             floydBranchTemplate,
		"stack.md":              floydStackTemplate,
		"AGENT_INSTRUCTIONS.md": floydAgentInstructions,
	}

	for name, content := range files {
		target := filepath.Join(dir, name)
		if err := ensureFileWithContent(target, content); err != nil {
			return "", err
		}
	}

	// Ensure progress header exists if the file already had content
	if err := ensureProgressHeader(filepath.Join(dir, "progress.md")); err != nil {
		return "", err
	}

	return dir, nil
}

func ensureFileWithContent(path, content string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func ensureProgressHeader(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return os.WriteFile(path, []byte(floydProgressTemplate), 0o644)
	}
	if !strings.Contains(string(data), "| Timestamp | Action Taken | Result/Status | Next Step |") {
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := f.Write([]byte(floydProgressTemplate)); err != nil {
			return err
		}
	}
	return nil
}

func appendProgress(path, action, result, next string) error {
	if err := ensureProgressHeader(path); err != nil {
		return err
	}

	entry := fmt.Sprintf("| %s | %s | %s | %s |\n",
		time.Now().Format("2006-01-02 15:04:05"),
		sanitizeCell(action),
		sanitizeCell(result),
		sanitizeCell(next),
	)

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(entry)
	return err
}

func sanitizeCell(s string) string {
	s = strings.ReplaceAll(s, "|", "/")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	if s == "" {
		return "-"
	}
	return s
}

func readFileOrPlaceholder(path, placeholder string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return placeholder
	}
	return string(data)
}

func readLastLines(path string, n int) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return "(progress log unavailable)"
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) <= n {
		return strings.Join(lines, "\n")
	}
	return strings.Join(lines[len(lines)-n:], "\n")
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if dir == "/" {
			return dir, nil
		}
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return dir, nil
		}
		dir = parent
	}
}

func floydSection(title, body string, maxLines int) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#88C0D0")).
		Bold(true)
	bodyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ECEFF4"))

	clipped := clipLines(body, maxLines)
	return lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render(title),
		bodyStyle.Render(clipped),
	)
}

func clipLines(text string, maxLines int) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return "(empty)"
	}
	lines := strings.Split(text, "\n")
	if len(lines) <= maxLines {
		return text
	}
	lines = append(lines[:maxLines], "...")
	return strings.Join(lines, "\n")
}

// renderFloydMode renders the FLOYD agent view.
func (m Model) renderFloydMode() string {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#88C0D0")).
		Padding(1, 2)

	header := headerStyle.Render(floydHeaderCommand)

	modeTitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ECEFF4")).
		Bold(true).
		Padding(0, 2).
		Render("FLOYD Agent Console")

	divider := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4C566A")).
		Render(strings.Repeat("─", max(10, m.width-4)))

	toolItems := make([]string, len(m.floydTools))
	for i, tool := range m.floydTools {
		style := lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(lipgloss.Color("#ECEFF4"))
		if i == m.floydSelectedTool {
			style = style.Bold(true).
				Foreground(lipgloss.Color("#2E3440")).
				Background(lipgloss.Color("#88C0D0"))
			toolItems[i] = style.Render("▸ " + tool.Name)
		} else {
			toolItems[i] = style.Render("  " + tool.Name)
		}
	}

	toolList := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4C566A")).
		Padding(0, 1).
		Width(32).
		Render(lipgloss.JoinVertical(lipgloss.Left, toolItems...))

	var promptSection string
	if m.floydPromptActive {
		label := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#88C0D0")).
			Bold(true).
			Render(m.floydPromptLabel)
		promptSection = lipgloss.JoinVertical(lipgloss.Left, label, m.renderFloydPrompt())
	} else {
		promptSection = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4C566A")).
			Render("Press Enter to run the selected tool. Use ↑/↓ to change tools, Ctrl+K or Esc to exit.")
	}

	toolDescription := "(no tool selected)"
	if len(m.floydTools) > 0 {
		toolDescription = m.floydTools[m.floydSelectedTool].Description
	}

	statusSections := []string{
		floydSection("Tool Details", toolDescription, 4),
		floydSection("Master Plan (.floyd/master_plan.md)", m.floydMasterPlan, 8),
		floydSection("Tech Stack (.floyd/stack.md)", m.floydStack, 6),
		floydSection("Recent Progress (.floyd/progress.md)", m.floydRecentLog, 6),
	}

	descBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4C566A")).
		Padding(1, 2).
		Width(m.width - 40).
		Render(lipgloss.JoinVertical(lipgloss.Left, statusSections...))

	outputBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4C566A")).
		Padding(0, 1).
		Height(max(8, m.height-22)).
		Width(m.width - 6).
		Render(m.floydViewport.View())

	status := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ECEFF4")).
		Render(m.floydStatus)

	mainRow := lipgloss.JoinHorizontal(lipgloss.Top, toolList, descBox)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		modeTitle,
		divider,
		mainRow,
		promptSection,
		outputBox,
		status,
	)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Left, lipgloss.Top).
		Padding(1, 2).
		Render(content)
}

// renderFloydPrompt renders the active prompt input or editor.
func (m Model) renderFloydPrompt() string {
	if m.floydAwaitContent {
		return m.floydContent.View()
	}
	return m.floydInput.View()
}

// handleFloydKeyPress processes key events while in FLOYD mode.
func (m Model) handleFloydKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if !m.floydPromptActive {
		switch msg.String() {
		case "pgup", "pgdown", "home", "end", "shift+up", "shift+down":
			var cmd tea.Cmd
			m.floydViewport, cmd = m.floydViewport.Update(msg)
			m.floydPinned = m.floydViewport.AtBottom()
			return m, cmd
		}
	}

	if m.floydPromptActive {
		// Handle multiline content entry (Write tool)
		if m.floydAwaitContent {
			switch {
			case msg.Type == tea.KeyCtrlD:
				content := m.floydContent.Value()
				if err := m.executeWrite(content); err != nil {
					m.floydStatus = fmt.Sprintf("Write failed: %v", err)
					m.logFloyd("write "+m.floydWritePath, "error: "+err.Error(), "check path or permissions")
				} else {
					m.floydStatus = "Write completed successfully."
					summary := fmt.Sprintf("Saved %s (%d bytes)\n", m.floydWritePath, len(content))
					m = m.appendFloydChunk(summary)
					m.logFloyd("write "+m.floydWritePath, fmt.Sprintf("%d bytes", len(content)), "review file")
				}
				m = m.resetFloydPrompt()
				return m, nil
			case msg.Type == tea.KeyEsc:
				m.floydStatus = "Write cancelled."
				m = m.resetFloydPrompt()
				return m, nil
			default:
				var cmd tea.Cmd
				m.floydContent, cmd = m.floydContent.Update(msg)
				return m, cmd
			}
		}

		// Handle single-line prompt entry
		switch msg.Type {
		case tea.KeyEsc:
			m.floydStatus = "Cancelled."
			m = m.resetFloydPrompt()
			return m, nil
		case tea.KeyEnter:
			value := strings.TrimSpace(m.floydInput.Value())
			if value == "" {
				m.floydStatus = "Input cannot be empty."
				return m, nil
			}
			switch m.floydPromptKind {
			case ToolWrite:
				m.floydWritePath = value
				m.floydAwaitContent = true
				m.floydPromptLabel = fmt.Sprintf("Editing %s (Ctrl+D to save, Esc to cancel)", value)
				m.floydContent.SetValue(m.prefillWriteContent(value))
				m.floydContent.CursorStart()
				m.floydContent.Focus()
				return m, nil
			default:
				var cmd tea.Cmd
				m, cmd = m.startFloydAction(m.floydPromptKind, value)
				m = m.resetFloydPrompt()
				return m, cmd
			}
		default:
			var cmd tea.Cmd
			m.floydInput, cmd = m.floydInput.Update(msg)
			return m, cmd
		}
	}

	switch msg.String() {
	case "esc", "ctrl+k":
		m = m.resetFloydPrompt()
		m.floydMode = false
		return m, nil
	case "up", "k":
		if m.floydSelectedTool > 0 {
			m.floydSelectedTool--
		}
		return m, nil
	case "down", "j":
		if m.floydSelectedTool < len(m.floydTools)-1 {
			m.floydSelectedTool++
		}
		return m, nil
	case "enter":
		if len(m.floydTools) == 0 {
			return m, nil
		}
		selected := m.floydTools[m.floydSelectedTool]
		prevPinned := m.floydViewport.AtBottom()
		m = m.prepareFloydPrompt(selected)
		if prevPinned {
			m.floydViewport.GotoBottom()
			m.floydPinned = true
		} else {
			m.floydPinned = m.floydViewport.AtBottom()
		}
		return m, nil
	}

	return m, nil
}

// prepareFloydPrompt initializes the prompt state for the selected tool.
func (m Model) prepareFloydPrompt(tool FloydTool) Model {
	if tool.Kind == ToolTemplate {
		return m.appendFloydTemplate(tool)
	}

	m.floydPromptActive = true
	m.floydPromptKind = tool.Kind
	m.floydInput.SetValue("")
	m.floydAwaitContent = false

	switch tool.Kind {
	case ToolBash:
		m.floydPromptLabel = "Enter bash command"
		m.floydInput.Placeholder = "e.g., ls -la"
	case ToolRead:
		m.floydPromptLabel = "Enter path to read"
		m.floydInput.Placeholder = "/path/to/file.txt"
	case ToolLS:
		m.floydPromptLabel = "Enter directory to list"
		m.floydInput.Placeholder = "."
	case ToolGlob:
		m.floydPromptLabel = "Enter glob pattern [optional base path]"
		m.floydInput.Placeholder = "**/*.go"
	case ToolWrite:
		m.floydPromptLabel = "Enter file path to write"
		m.floydInput.Placeholder = "./output.txt"
	}

	m.floydInput.CursorStart()
	m.floydInput.Focus()
	m.floydStatus = "Awaiting input..."
	return m
}

// resetFloydPrompt clears the prompt/input state.
func (m Model) resetFloydPrompt() Model {
	m.floydPromptActive = false
	m.floydAwaitContent = false
	m.floydPromptLabel = ""
	m.floydWritePath = ""
	m.floydInput.Blur()
	m.floydInput.SetValue("")
	m.floydContent.Blur()
	m.floydContent.SetValue("")
	return m
}

// executeWrite saves content to the requested file.
func (m *Model) executeWrite(content string) error {
	if m.floydWritePath == "" {
		return fmt.Errorf("no target path set")
	}

	// Ensure parent directory exists
	dir := filepath.Dir(m.floydWritePath)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	return os.WriteFile(m.floydWritePath, []byte(content), fs.FileMode(0o644))
}

// prefillWriteContent loads existing file content for editing.
func (m *Model) prefillWriteContent(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

// max helper to avoid importing math for ints.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
