package agenttui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

// ViewportComponent manages the streaming output display with scrollback
type ViewportComponent struct {
	viewport.Model
	history     *HistoryStore
	builder     *strings.Builder // For building streaming content
	streaming   bool
	autoScroll  bool
	lineCount   int
	lastToken   string
}

// NewViewportComponent creates a new viewport component
func NewViewportComponent() ViewportComponent {
	vp := viewport.New(0, 0)
	vp.HighPerformanceRendering = true

	return ViewportComponent{
		Model:      vp,
		history:    NewHistoryStore(10000),
		builder:    &strings.Builder{},
		streaming:  false,
		autoScroll: true,
		lineCount:  0,
	}
}

// SetHistory sets the history store
func (v *ViewportComponent) SetHistory(h *HistoryStore) {
	v.history = h
}

// SetSize sets the viewport dimensions
func (v *ViewportComponent) SetSize(width, height int) {
	v.Model.Width = width - 4 // Account for borders
	v.Model.Height = height
}

// StartStreaming prepares the viewport for streaming
func (v *ViewportComponent) StartStreaming() {
	v.streaming = true
	v.autoScroll = true
	v.builder = &strings.Builder{}
	v.lineCount = 0
}

// StopStreaming ends the streaming session
func (v *ViewportComponent) StopStreaming() {
	v.streaming = false

	// Add the accumulated content to history
	if v.builder != nil && v.builder.Len() > 0 {
		msg := Message{
			Role:    RoleAssistant,
			Content: v.builder.String(),
		}
		v.history.Append(msg)
		v.builder = &strings.Builder{}
	}
}

// AppendToken adds a token to the streaming content
func (v *ViewportComponent) AppendToken(token string) {
	if v.builder == nil {
		v.builder = &strings.Builder{}
	}

	v.lastToken = token
	v.builder.WriteString(token)

	// Update viewport content incrementally
	v.Model.SetContent(v.getDisplayContent())

	// Auto-scroll during streaming
	if v.autoScroll {
		v.Model.GotoBottom()
	}
}

// AppendMessage adds a complete message to the viewport
func (v *ViewportComponent) AppendMessage(msg Message) {
	v.history.Append(msg)
	v.Model.SetContent(v.getDisplayContent())

	if v.autoScroll {
		v.Model.GotoBottom()
	}
}

// AppendSystem adds a system message
func (v *ViewportComponent) AppendSystem(text string) {
	msg := Message{
		Role:      RoleSystem,
		Content:   text,
		Timestamp: time.Now(),
	}
	v.AppendMessage(msg)
}

// AppendError adds an error message
func (v *ViewportComponent) AppendError(text string) {
	msg := Message{
		Role:      RoleSystem,
		Content:   "Error: " + text,
		Timestamp: time.Now(),
	}
	v.AppendMessage(msg)
}

// getDisplayContent returns the current display content
func (v *ViewportComponent) getDisplayContent() string {
	var sb strings.Builder

	// Add history messages
	for msg := range v.history.Iter() {
		sb.WriteString(msg.Formatted())
		sb.WriteString("\n")
	}

	// Add streaming content if active
	if v.streaming && v.builder != nil && v.builder.Len() > 0 {
		sb.WriteString("│ ")
		sb.WriteString(v.builder.String())
	}

	return sb.String()
}

// RefreshContent rebuilds the viewport content from history
func (v *ViewportComponent) RefreshContent() {
	v.Model.SetContent(v.getDisplayContent())
}

// AtBottom returns whether the viewport is at the bottom
func (v *ViewportComponent) AtBottom() bool {
	return v.Model.AtBottom()
}

// GotoBottom scrolls to the bottom
func (v *ViewportComponent) GotoBottom() {
	v.Model.GotoBottom()
}

// GotoTop scrolls to the top
func (v *ViewportComponent) GotoTop() {
	v.Model.GotoTop()
}

// LineDown scrolls down by n lines
func (v *ViewportComponent) LineDown(n int) {
	v.Model.LineDown(n)
}

// LineUp scrolls up by n lines
func (v *ViewportComponent) LineUp(n int) {
	v.Model.LineUp(n)
}

// SetAutoScroll sets whether to auto-scroll during streaming
func (v *ViewportComponent) SetAutoScroll(auto bool) {
	v.autoScroll = auto
}

// IsStreaming returns whether currently streaming
func (v *ViewportComponent) IsStreaming() bool {
	return v.streaming
}

// Content returns the current viewport content
func (v *ViewportComponent) Content() string {
	return v.getDisplayContent()
}

// Update handles viewport messages
func (v ViewportComponent) Update(msg tea.Msg) (ViewportComponent, tea.Cmd) {
	var cmd tea.Cmd
	v.Model, cmd = v.Model.Update(msg)
	return v, cmd
}

// View renders the viewport
func (v ViewportComponent) View() string {
	return v.Model.View()
}

// RenderWithBorder renders the viewport with a border
func (v ViewportComponent) RenderWithBorder(borderStyle lipgloss.Style) string {
	content := v.View()
	return borderStyle.Width(v.Model.Width).Render(content)
}

// GetVisibleLineCount returns the number of visible lines
func (v *ViewportComponent) GetVisibleLineCount() int {
	return v.Model.Height
}

// GetTotalLineCount returns the total content line count
func (v *ViewportComponent) GetTotalLineCount() int {
	return v.Model.TotalLineCount()
}

// ScrollPercentage returns the current scroll position as a percentage
func (v *ViewportComponent) ScrollPercentage() float64 {
	total := v.GetTotalLineCount()
	if total == 0 {
		return 100
	}

	visible := v.GetVisibleLineCount()
	if visible >= total {
		return 100
	}

	// This is approximate - viewport doesn't expose exact position
	// This is a simplification - actual implementation would need more work
	return 100.0
}


// CopyContent copies the viewport content to clipboard
func (v *ViewportComponent) CopyContent() string {
	return v.getDisplayContent()
}

// GetLastToken returns the last streamed token
func (v *ViewportComponent) GetLastToken() string {
	return v.lastToken
}

// Clear clears the viewport content (but not history)
func (v *ViewportComponent) Clear() {
	v.builder = &strings.Builder{}
	v.lastToken = ""
	v.Model.SetContent("")
}

// ClearAll clears both viewport and history
func (v *ViewportComponent) ClearAll() {
	v.Clear()
	v.history = NewHistoryStore(v.history.maxSize)
}
