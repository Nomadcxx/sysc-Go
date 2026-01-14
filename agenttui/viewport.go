package agenttui

import (
	"strings"
	"sync"
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

	// Performance optimization: cache history content to avoid reformatting on every token
	cachedHistoryContent string
	cachedHistoryMu     sync.RWMutex
	lastHistoryUpdate   time.Time
	tokenCount          int  // Tokens since last viewport update
	tickCount           int  // Ticks since last viewport update
}

// Viewport update throttling - only update viewport every N tokens or M milliseconds
const (
	tokensPerViewportUpdate = 5     // Update viewport every 5 tokens during streaming
	msBetweenViewportUpdates = 100  // Or every 100ms max
	maxHistoryLinesToShow   = 500   // Limit history to prevent slowdown
)

// NewViewportComponent creates a new viewport component
func NewViewportComponent() ViewportComponent {
	// Create viewport with default non-zero dimensions
	// IMPORTANT: HighPerformanceRendering is disabled because it causes View() to return
	// only newlines instead of actual content. We need the full content for streaming display.
	vp := viewport.New(80, 20)
	vp.HighPerformanceRendering = false

	return ViewportComponent{
		Model:             vp,
		history:           NewHistoryStore(10000),
		builder:           &strings.Builder{},
		streaming:         false,
		autoScroll:        true,
		lineCount:         0,
		cachedHistoryContent: "",
		lastHistoryUpdate:  time.Now(),
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

	// Invalidate cache and force final viewport update
	v.invalidateCache()
	v.Model.SetContent(v.rebuildHistoryContent())

	if v.autoScroll {
		v.Model.GotoBottom()
	}

	// Reset token counter
	v.tokenCount = 0
}

// AppendToken adds a token to the streaming content
func (v *ViewportComponent) AppendToken(token string) {
	if v.builder == nil {
		v.builder = &strings.Builder{}
	}

	v.lastToken = token
	v.builder.WriteString(token)
	v.tokenCount++

	// Throttle viewport updates - only update every N tokens or after time threshold
	shouldUpdate := v.tokenCount >= tokensPerViewportUpdate ||
		time.Since(v.lastHistoryUpdate) > time.Duration(msBetweenViewportUpdates)*time.Millisecond

	if shouldUpdate {
		v.updateViewport()
	}
}

// updateViewport updates the viewport content (call directly when forcing an update)
func (v *ViewportComponent) updateViewport() {
	v.Model.SetContent(v.getDisplayContentOptimized())
	v.tokenCount = 0
	v.lastHistoryUpdate = time.Now()

	if v.autoScroll {
		v.Model.GotoBottom()
	}
}

// AppendMessage adds a complete message to the viewport
func (v *ViewportComponent) AppendMessage(msg Message) {
	v.history.Append(msg)
	// Invalidate cache when new message is added
	v.invalidateCache()
	v.Model.SetContent(v.getDisplayContentOptimized())

	if v.autoScroll {
		v.Model.GotoBottom()
	}
}

// invalidateCache clears the cached history content
func (v *ViewportComponent) invalidateCache() {
	v.cachedHistoryMu.Lock()
	v.cachedHistoryContent = ""
	v.lastHistoryUpdate = time.Now()
	v.cachedHistoryMu.Unlock()
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

// getDisplayContent returns the current display content (deprecated - use getDisplayContentOptimized)
func (v *ViewportComponent) getDisplayContent() string {
	return v.getDisplayContentOptimized()
}

// getDisplayContentOptimized returns the display content with caching and history limiting
func (v *ViewportComponent) getDisplayContentOptimized() string {
	v.cachedHistoryMu.RLock()
	cached := v.cachedHistoryContent
	v.cachedHistoryMu.RUnlock()

	// If we have cached history and we're streaming, just append the builder
	if v.streaming && cached != "" && v.builder != nil && v.builder.Len() > 0 {
		return cached + "│ " + v.builder.String() + "\n"
	}

	// Need to rebuild history content
	return v.rebuildHistoryContent()
}

// rebuildHistoryContent rebuilds the history content from scratch and caches it
func (v *ViewportComponent) rebuildHistoryContent() string {
	var sb strings.Builder

	// Build history content, limiting to recent messages
	lineCount := 0
	for msg := range v.history.Iter() {
		formatted := msg.Formatted()
		sb.WriteString(formatted)
		sb.WriteString("\n")
		lineCount += strings.Count(formatted, "\n") + 1

		// Limit history to prevent slowdown
		if lineCount >= maxHistoryLinesToShow {
			sb.WriteString("│ ... (earlier messages truncated for performance)\n")
			break
		}
	}

	// Cache the result
	result := sb.String()

	v.cachedHistoryMu.Lock()
	v.cachedHistoryContent = result
	v.cachedHistoryMu.Unlock()

	return result
}

// RefreshContent rebuilds the viewport content from history
func (v *ViewportComponent) RefreshContent() {
	v.invalidateCache()
	v.Model.SetContent(v.getDisplayContentOptimized())
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
