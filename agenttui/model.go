package agenttui

import (
	"context"
	"fmt"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/Nomadcxx/sysc-Go/agent"
)

// AgentModel is the root model for the agent TUI
type AgentModel struct {
	// Dimensions
	width  int
	height int

	// UI Components
	input    InputComponent
	viewport ViewportComponent

	// State
	streaming   bool
	toolRunning bool

	// Cancellation
	ctx        context.Context
	cancelFunc context.CancelFunc

	// Async channels (buffered for backpressure)
	// NOTE: streamChan is created per-request to avoid concurrency issues
	streamChan   chan StreamChunkMsg
	streamChanMu sync.Mutex
	toolChan     chan ToolCompleteMsg

	// Data
	history *HistoryStore
	styles  Styles

	// Status
	status      string
	statusTimer *time.Timer
	statusClear bool

	// Configuration
	theme      string
	maxHistory int

	// Agent integration
	client      agent.GLMClient
	processor   *agent.StreamProcessor
	orchestrator *agent.Orchestrator

	// Header
	headerText string
	headerSub  string

	// Tick rate for frame updates
	tickInterval time.Duration
	lastTick     time.Time
}

// NewAgentModel creates a new agent TUI model
func NewAgentModel() AgentModel {
	ctx, cancel := context.WithCancel(context.Background())

	history := NewHistoryStore(10000)

	// Add initial system message with FLOYD persona
	history.Append(Message{
		Role:      RoleSystem,
		Content:   "You are FLOYD, a professional coding assistant. You are concise, helpful, and accurate. You write working code, avoid word salad, and get straight to the point.",
		Timestamp: time.Now(),
	})

	// Initialize the GLM proxy client (reads from ~/.claude/settings.json)
	// This uses the Anthropic-compatible endpoint which maps Claude models to GLM
	client := agent.NewProxyClient("", "")

	return AgentModel{
		width:        80,
		height:       24,
		input:        NewInputComponent(),
		viewport:     NewViewportComponent(),
		streaming:    false,
		toolRunning:  false,
		ctx:          ctx,
		cancelFunc:   cancel,
		streamChan:   nil, // Created per-request in sendToAgent
		toolChan:     make(chan ToolCompleteMsg, 100),
		history:      history,
		styles:       DefaultStyles(),
		status:       "Ready",
		theme:        "catppuccin",
		maxHistory:   10000,
		headerText:   "FLOYD Agent Console",
		headerSub:    "GLM Coding Plan (via Anthropic proxy)",
		tickInterval: 33 * time.Millisecond, // ~30 FPS (sufficient for TUI, reduces overhead)
		lastTick:     time.Now(),
		client:       client,
		processor:    agent.NewStreamProcessor(client),
		orchestrator: agent.NewOrchestrator(client),
	}
}

// Init initializes the model
func (m AgentModel) Init() tea.Cmd {
	// Focus input (doesn't return a command)
	m.input.Focus()

	// Don't waitForStream here - streamChan is nil until sendToAgent is called
	// The TickMsg loop will keep the UI alive until user submits a message
	return tea.Batch(
		m.waitForTool(),
		m.tick(),
	)
}

// tick returns a command that sends TickMsg at regular intervals
func (m AgentModel) tick() tea.Cmd {
	return tea.Tick(m.tickInterval, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// waitForStream returns a command that waits for stream chunks
func (m *AgentModel) waitForStream() tea.Cmd {
	return func() tea.Msg {
		// Get the current streamChan safely
		m.streamChanMu.Lock()
		ch := m.streamChan
		m.streamChanMu.Unlock()

		if ch == nil {
			// No active stream channel - this shouldn't happen during streaming
			// Return a special message that won't cause issues
			return StreamChunkMsg{Done: false}
		}

		// Wait for data with a longer timeout to avoid busy loops
		select {
		case <-m.ctx.Done():
			return StreamChunkMsg{Done: true}
		case msg, ok := <-ch:
			if !ok {
				// Channel closed - clear the reference to prevent other in-flight
				// waitForStream calls from also trying to read the closed channel
				m.streamChanMu.Lock()
				if m.streamChan == ch {
					m.streamChan = nil
				}
				m.streamChanMu.Unlock()
				return StreamChunkMsg{Done: true}
			}
			return msg
		case <-time.After(500 * time.Millisecond):
			// Timeout - return self-retry command to continue waiting
			// This avoids creating a tick loop
			return waitForStreamRetryMsg{}
		}
	}
}

// waitForStreamRetryMsg is a special message that causes waitForStream to retry
type waitForStreamRetryMsg struct{}

func (waitForStreamRetryMsg) retry() {} // Marker method

// waitForTool returns a command that waits for tool completion
func (m AgentModel) waitForTool() tea.Cmd {
	return func() tea.Msg {
		select {
		case <-m.ctx.Done():
			return ToolCompleteMsg{}
		case msg, ok := <-m.toolChan:
			if !ok {
				return ToolCompleteMsg{}
			}
			return msg
		case <-time.After(100 * time.Millisecond):
			// Timeout - tick again
			return TickMsg(time.Now())
		}
	}
}

// SetSize sets the terminal dimensions
func (m *AgentModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Configure component sizes
	// Header: 4 lines
	// Viewport: remaining - input (5 lines) - padding
	viewportHeight := height - 9 // 4 header + 5 input + borders
	if viewportHeight < 5 {
		viewportHeight = 5
	}

	m.input.SetSize(width, 5)
	m.viewport.SetSize(width, viewportHeight)
}

// SetTheme sets the UI theme
func (m *AgentModel) SetTheme(theme string) {
	m.theme = theme
	m.styles = NewStyles(theme)
}

// SetStatus sets the status message with optional timeout
func (m *AgentModel) SetStatus(msg string, timeout time.Duration) {
	m.status = msg
	if timeout > 0 {
		m.statusClear = true
		m.statusTimer = time.NewTimer(timeout)
	} else {
		m.statusClear = false
		if m.statusTimer != nil {
			m.statusTimer.Stop()
		}
	}
}

// ClearStatus clears the status message
func (m *AgentModel) ClearStatus() {
	m.status = ""
	if m.statusTimer != nil {
		m.statusTimer.Stop()
	}
	m.statusClear = false
}

// StartStream prepares the model for streaming
func (m *AgentModel) StartStream() {
	m.streaming = true
	m.viewport.StartStreaming()
	m.SetStatus("Receiving response...", 0)
}

// EndStream finalizes the streaming session
func (m *AgentModel) EndStream() {
	m.streaming = false
	m.viewport.StopStreaming()
	m.SetStatus("Ready", 3*time.Second)
}

// Cancel cancels the current operation
func (m *AgentModel) Cancel() {
	if m.cancelFunc != nil {
		m.cancelFunc()
	}
	m.streaming = false
	m.toolRunning = false
	m.viewport.StopStreaming()
	m.SetStatus("Cancelled", 2*time.Second)
}

// SubmitMessage submits a user message
func (m *AgentModel) SubmitMessage(content string) tea.Cmd {
	// Add user message to history
	msg := Message{
		Role:      RoleUser,
		Content:   content,
		Timestamp: time.Now(),
	}
	m.history.Append(msg)

	// Add to viewport
	m.viewport.AppendMessage(msg)

	// Add to input history
	m.input.AddToHistory(content)
	m.input.SetValue("")

	// Trigger the agent response
	return m.sendToAgent(content)
}

// sendToAgent sends a message to the GLM agent
func (m *AgentModel) sendToAgent(content string) tea.Cmd {
	m.SetStatus("Thinking...", 0)
	m.viewport.StartStreaming()
	m.streaming = true

	// Build message history for API
	messages := m.history.GetMessagesForAPI(true)  // Include system prompts!

	// Add the new user message
	messages = append(messages, agent.Message{
		Role:    "user",
		Content: content,
	})

	// Build request
	req := agent.ChatRequest{
		Messages:    messages,
		MaxTokens:   8192,
		Temperature: 0.7,
	}

	// Create a NEW channel for this request to avoid concurrency issues
	streamChan := make(chan StreamChunkMsg, 100)

	// Lock and update the streamChan reference
	m.streamChanMu.Lock()
	m.streamChan = streamChan
	m.streamChanMu.Unlock()

	// Start streaming in background
	return func() tea.Msg {
		chunkChan, err := m.client.StreamChat(m.ctx, req)
		if err != nil {
			close(streamChan)
			return ConnErrorMsg{Err: err}
		}

		// Bridge chunks to TUI messages
		go func() {
			defer close(streamChan)

			for chunk := range chunkChan {
				// Check context before attempting send
				select {
				case <-m.ctx.Done():
					return
				default:
				}

				// Prepare the message
				var msg StreamChunkMsg
				if chunk.Error != nil {
					msg = StreamChunkMsg{Error: chunk.Error, Done: true}
				} else if chunk.Done {
					msg = StreamChunkMsg{Done: true}
				} else if chunk.Token != "" {
					msg = StreamChunkMsg{Token: chunk.Token}
				} else {
					continue
				}

				// Safe send with context check
				select {
				case streamChan <- msg:
				case <-m.ctx.Done():
					return
				}
			}
		}()

		// Return first chunk immediately to trigger UI update
		// After this is processed by handleStreamChunk, waitForStream() will be returned
		return StreamChunkMsg{Token: "", Done: false}
	}
}

// HandleResize handles terminal resize
func (m *AgentModel) HandleResize(width, height int) tea.Cmd {
	// Store scroll position
	wasAtBottom := m.viewport.AtBottom()

	// Update sizes
	m.SetSize(width, height)
	m.viewport.RefreshContent()

	// Restore scroll position
	if wasAtBottom {
		m.viewport.GotoBottom()
	}

	return nil
}

// GetHistory returns the history store
func (m *AgentModel) GetHistory() *HistoryStore {
	return m.history
}

// GetViewport returns the viewport component
func (m *AgentModel) GetViewport() *ViewportComponent {
	return &m.viewport
}

// GetInput returns the input component
func (m *AgentModel) GetInput() *InputComponent {
	return &m.input
}

// IsStreaming returns whether currently streaming
func (m *AgentModel) IsStreaming() bool {
	return m.streaming
}

// IsToolRunning returns whether a tool is running
func (m *AgentModel) IsToolRunning() bool {
	return m.toolRunning
}

// GetContext returns the model's context
func (m *AgentModel) GetContext() context.Context {
	return m.ctx
}

// SpawnSubAgent creates a new sub-agent with the given type and task
func (m *AgentModel) SpawnSubAgent(agentType agent.AgentType, task string) tea.Cmd {
	return func() tea.Msg {
		// Build context from current history
		contextData := map[string]any{
			"history_size": m.history.Len(),
			"streaming":    m.streaming,
			"theme":        m.theme,
		}

		subAgent, err := m.orchestrator.Spawn(agentType, task, contextData)
		if err != nil {
			return StatusMsg{Message: "Failed to spawn sub-agent: " + err.Error(), Timeout: 3 * time.Second}
		}

		m.SetStatus(fmt.Sprintf("Spawned %s agent: %s", agentType, subAgent.ID), 2*time.Second)

		return SubAgentSpawnMsg{
			AgentID:   subAgent.ID,
			AgentType: string(agentType),
			Task:      task,
		}
	}
}

// CollectSubAgentResults retrieves results from a completed sub-agent
func (m *AgentModel) CollectSubAgentResults(agentID string) tea.Cmd {
	return func() tea.Msg {
		agent, err := m.orchestrator.CollectResults(agentID)
		if err != nil {
			return SubAgentCompleteMsg{
				AgentID: agentID,
				Error:   err,
			}
		}

		// Add the output as a system message to the viewport
		m.viewport.AppendSystem(fmt.Sprintf("[Sub-Agent %s - %s]", agent.Type, agent.ID))
		m.viewport.AppendSystem(agent.Output)

		return SubAgentCompleteMsg{
			AgentID: agentID,
			Output:  agent.Output,
		}
	}
}

// ListSubAgents returns all active sub-agents
func (m *AgentModel) ListSubAgents() []*agent.SubAgent {
	return m.orchestrator.ListAgents()
}

// CleanupOldSubAgents removes agents older than the specified duration
func (m *AgentModel) CleanupOldSubAgents(olderThan time.Duration) int {
	return m.orchestrator.Cleanup(olderThan)
}
