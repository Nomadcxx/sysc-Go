package floydui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Message types for the FLOYD UI

// tickMsg is sent periodically to drive animations
type tickMsg time.Time

// thinkingFinishedMsg is sent when the agent finishes thinking
type thinkingFinishedMsg struct{ response string }

// thinkingStartedMsg is sent when the agent starts thinking
type thinkingStartedMsg struct{}

// streamTokenMsg is sent for each token from the stream
type streamTokenMsg struct{ token string }

// streamErrorMsg is sent when there's a stream error
type streamErrorMsg struct{ err error }

// toolUseMsg is sent when a tool is being used
type toolUseMsg struct{ toolName, toolID string }

// toolExecutingMsg is sent when a tool is executing
type toolExecutingMsg struct{ toolName string }

// toolResultMsg is sent when a tool returns a result
type toolResultMsg struct{ toolName, content string }

// toolErrorMsg is sent when a tool returns an error
type toolErrorMsg struct{ toolName, content string }

// saveSessionMsg is sent to save the session
type saveSessionMsg struct{}

// progressMsg is sent to update progress
type progressMsg float64

// themeChangedMsg is sent when theme is changed
type themeChangedMsg struct{ themeName string }

// sessionLoadedMsg is sent when session is loaded
type sessionLoadedMsg struct{ messages []string; history []string }

// SubAgent messages for agent spawning
type subAgentSpawnMsg struct {
	agentType string
	task      string
}

type subAgentStatusMsg struct {
	status string
	detail string
}

type subAgentCompleteMsg struct {
	result string
}

// Tick returns a tea.Cmd that sends tickMsg at the given interval
func Tick(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// ThinkingFinished returns a tea.Cmd that sends thinkingFinishedMsg after a delay
func ThinkingFinished(delay time.Duration, response string) tea.Cmd {
	return tea.Tick(delay, func(t time.Time) tea.Msg {
		return thinkingFinishedMsg{response: response}
	})
}

// Progress returns a tea.Cmd that sends progressMsg
func Progress(percent float64) tea.Cmd {
	return func() tea.Msg {
		return progressMsg(percent)
	}
}
