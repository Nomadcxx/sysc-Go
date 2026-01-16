package agenttui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles messages and updates the model
func (m AgentModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		cmd := m.HandleResize(msg.Width, msg.Height)
		return m, cmd

	case TickMsg:
		// Handle periodic updates
		m.lastTick = time.Time(msg)
		// If we're streaming, also wait for stream chunks
		if m.streaming {
			return m, tea.Batch(m.tick(), m.waitForStream())
		}
		return m, m.tick()

	case StreamChunkMsg:
		return m.handleStreamChunk(msg)

	case waitForStreamRetryMsg:
		// waitForStream timed out - retry waiting without creating a tick loop
		return m, m.waitForStream()

	case ToolStartMsg:
		return m.handleToolStart(msg)

	case ToolCompleteMsg:
		return m.handleToolComplete(msg)

	case ToolProgressMsg:
		return m.handleToolProgress(msg)

	case UserSubmitMsg:
		return m, m.SubmitMessage(msg.Content)

	case StatusMsg:
		m.SetStatus(msg.Message, msg.Timeout)
		return m, nil

	case ConnErrorMsg:
		//m.lastError = msg.Err
		m.SetStatus("Connection error", 3*time.Second)
		return m, nil

	case CancelStreamMsg:
		m.Cancel()
		return m, nil

	case HistoryNavigateMsg:
		m.input.HistoryNavigate(msg.Direction)
		return m, nil

	case FocusChangeMsg:
		return m.handleFocusChange(msg)

	case AgentModeMsg:
		return m.handleAgentModeChange(msg)

	case SubAgentSpawnMsg:
		return m.handleSubAgentSpawn(msg)

	case SubAgentCompleteMsg:
		return m.handleSubAgentComplete(msg)

	case SubAgentStatusMsg:
		return m.handleSubAgentStatus(msg)
	}

	// Default: pass through to components
	var cmd tea.Cmd

	// Update input
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	// Update viewport ONLY when streaming or for specific messages
	// This prevents rebuilding history on every keystroke
	if m.streaming {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// handleKeyPress processes keyboard input
func (m AgentModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global quit
	switch msg.String() {
	case "ctrl+c":
		if m.streaming || m.toolRunning {
			// Cancel current operation first
			m.Cancel()
			return m, nil
		}
		return m, tea.Quit

	case "ctrl+l":
		// Clear conversation
		m.history.Clear()
		m.viewport.ClearAll()
		m.SetStatus("Conversation cleared", 2*time.Second)
		return m, nil

	case "ctrl+r":
		// Refresh viewport
		m.viewport.RefreshContent()
		return m, nil

	case "ctrl+s":
		// Toggle auto-scroll
		m.viewport.SetAutoScroll(!m.viewport.autoScroll)
		state := "enabled"
		if !m.viewport.autoScroll {
			state = "disabled"
		}
		m.SetStatus("Auto-scroll "+state, 2*time.Second)
		return m, nil
	}

	// Check for escape in normal mode
	if m.input.Mode() == ModeNormal {
		switch msg.String() {
		case "q":
			return m, tea.Quit
		}
	}

	// Update input component
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)

	// If user submitted, check for UserSubmitMsg
	// (This is handled in the main Update loop)

	return m, cmd
}

// handleStreamChunk processes streaming tokens
func (m AgentModel) handleStreamChunk(msg StreamChunkMsg) (tea.Model, tea.Cmd) {
	if msg.Error != nil {
		//m.lastError = msg.Error
		m.viewport.AppendError(msg.Error.Error())
		m.EndStream()
		// Don't return waitForStream - the stream is done!
		return m, m.waitForTool()
	}

	if msg.Done {
		m.EndStream()
		// Don't return waitForStream - the stream is done!
		return m, m.waitForTool()
	}

	// Append token to viewport
	m.viewport.AppendToken(msg.Token)

	// Continue waiting for more chunks
	return m, m.waitForStream()
}

// handleToolStart processes tool execution start
func (m AgentModel) handleToolStart(msg ToolStartMsg) (tea.Model, tea.Cmd) {
	m.toolRunning = true
	m.SetStatus("Running tool: "+msg.ToolID, 0)

	// Add tool call message to history
	toolMsg := Message{
		Role:      RoleTool,
		ToolName:  msg.ToolID,
		Timestamp: currentTime(),
	}
	m.history.Append(toolMsg)
	m.viewport.AppendMessage(toolMsg)

	return m, m.waitForTool()
}

// handleToolComplete processes tool execution completion
func (m AgentModel) handleToolComplete(msg ToolCompleteMsg) (tea.Model, tea.Cmd) {
	m.toolRunning = false

	if msg.Error != nil {
		m.viewport.AppendError(msg.ToolID + ": " + msg.Error.Error())
		m.SetStatus("Tool failed", 2*time.Second)
	} else {
		// Add result to viewport
		resultMsg := Message{
			Role:       RoleTool,
			ToolName:   msg.ToolID,
			ToolOutput: msg.Output,
			Timestamp:  currentTime(),
		}
		m.history.Append(resultMsg)
		m.viewport.AppendMessage(resultMsg)
		m.SetStatus("Tool completed", 2*time.Second)
	}

	return m, m.waitForTool()
}

// handleToolProgress processes tool execution progress
func (m AgentModel) handleToolProgress(msg ToolProgressMsg) (tea.Model, tea.Cmd) {
	status := msg.ToolID
	if msg.Progress > 0 {
		status += " (" + itoa(msg.Progress) + "%)"
	}
	if msg.Message != "" {
		status += ": " + msg.Message
	}
	m.SetStatus(status, 0)

	return m, nil
}

// handleFocusChange handles focus changes
func (m AgentModel) handleFocusChange(msg FocusChangeMsg) (tea.Model, tea.Cmd) {
	switch msg.Component {
	case "input":
		m.input.Focus()
	case "viewport":
		m.input.Blur()
	}
	return m, nil
}

// handleAgentModeChange handles agent mode changes
func (m AgentModel) handleAgentModeChange(msg AgentModeMsg) (tea.Model, tea.Cmd) {
	// For now, just log the mode change
	m.SetStatus("Mode: "+msg.Mode, 2*time.Second)
	return m, nil
}

// handleSubAgentSpawn handles sub-agent spawn events
func (m AgentModel) handleSubAgentSpawn(msg SubAgentSpawnMsg) (tea.Model, tea.Cmd) {
	// Add notification to viewport
	m.viewport.AppendSystem(fmt.Sprintf("[Spawned %s agent: %s]", msg.AgentType, msg.AgentID))
	m.viewport.AppendSystem(fmt.Sprintf("Task: %s", msg.Task))

	// Start a goroutine to wait for completion and collect results
	go func() {
		result, err := m.orchestrator.WaitForCompletion(msg.AgentID, 5*time.Minute)
		if err != nil {
			// Send error message
			// In a real implementation, we'd send this through a channel
			return
		}
		// Add to viewport when complete
		m.viewport.AppendSystem(fmt.Sprintf("[%s agent completed]", msg.AgentType))
		m.viewport.AppendSystem(result.Output)
	}()

	return m, nil
}

// handleSubAgentComplete handles sub-agent completion events
func (m AgentModel) handleSubAgentComplete(msg SubAgentCompleteMsg) (tea.Model, tea.Cmd) {
	if msg.Error != nil {
		m.viewport.AppendError(fmt.Sprintf("Sub-Agent %s failed: %v", msg.AgentID, msg.Error))
		m.SetStatus("Sub-agent failed", 2*time.Second)
	} else {
		m.SetStatus("Sub-agent completed", 2*time.Second)
	}
	return m, nil
}

// handleSubAgentStatus handles sub-agent status updates
func (m AgentModel) handleSubAgentStatus(msg SubAgentStatusMsg) (tea.Model, tea.Cmd) {
	status := fmt.Sprintf("[%s] %s", msg.AgentID, msg.Status)
	if msg.Info != "" {
		status += ": " + msg.Info
	}
	m.SetStatus(status, 0)
	return m, nil
}

// currentTime returns the current time
func currentTime() time.Time {
	return time.Now()
}

// itoa converts int to string (simple implementation)
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	sign := ""
	if i < 0 {
		sign = "-"
		i = -i
	}
	digits := []byte{}
	for i > 0 {
		digits = append(digits, byte('0'+i%10))
		i /= 10
	}
	// Reverse digits
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}
	return sign + string(digits)
}
