package agenttui

import "time"

// StreamChunkMsg represents a single token from the LLM stream
type StreamChunkMsg struct {
	Token string
	Done  bool
	Error error
}

// ToolStartMsg indicates a tool execution has begun
type ToolStartMsg struct {
	ToolID string
	Input  any
}

// ToolCompleteMsg indicates a tool execution has finished
type ToolCompleteMsg struct {
	ToolID string
	Output string
	Error  error
}

// ToolProgressMsg indicates progress during long-running tool execution
type ToolProgressMsg struct {
	ToolID   string
	Progress int    // 0-100
	Message  string
}

// UserSubmitMsg is sent when the user submits their input
type UserSubmitMsg struct {
	Content string
}

// ResizePreserveMsg is a specialized resize message that preserves state
type ResizePreserveMsg struct {
	Width  int
	Height int
}

// TickMsg is sent on each animation/frame update
type TickMsg time.Time

// StatusMsg updates the status bar
type StatusMsg struct {
	Message string
	Timeout time.Duration // If > 0, clear after this duration
}

// ConnErrorMsg indicates a connection/communication error
type ConnErrorMsg struct {
	Err error
}

// CancelStreamMsg requests cancellation of the current stream
type CancelStreamMsg struct{}

// HistoryNavigateMsg requests navigation through conversation history
type HistoryNavigateMsg struct {
	Direction int // -1 for older, 1 for newer, 0 for current
}

// FocusChangeMsg indicates focus has changed between components
type FocusChangeMsg struct {
	Component string // "input", "viewport", etc.
}

// AgentModeMsg indicates a mode change in the agent interaction
type AgentModeMsg struct {
	Mode string // "chat", "tool", "config"
}
