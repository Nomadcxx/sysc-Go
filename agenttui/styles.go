package agenttui

import (
	"github.com/Nomadcxx/sysc-Go/animations"
	"github.com/charmbracelet/lipgloss"
)

// Styles contains all lipgloss styles for the agent TUI
type Styles struct {
	// Persistent header styles
	Header      lipgloss.Style
	HeaderTitle lipgloss.Style
	HeaderSub   lipgloss.Style

	// Message styles
	UserMsg      lipgloss.Style
	AssistantMsg lipgloss.Style
	SystemMsg    lipgloss.Style
	ToolCall     lipgloss.Style
	ToolResult   lipgloss.Style

	// Status and error styles
	Status     lipgloss.Style
	StatusErr  lipgloss.Style
	Error      lipgloss.Style
	Warning    lipgloss.Style

	// Input area styles
	InputActive   lipgloss.Style
	InputInactive lipgloss.Style
	InputPlaceholder lipgloss.Style

	// Viewport styles
	Viewport      lipgloss.Style
	ViewportBorder lipgloss.Style

	// Component borders
	Border       lipgloss.Style
	BorderFocused lipgloss.Style

	// Utility styles
	Dim      lipgloss.Style
	Muted    lipgloss.Style
	Highlight lipgloss.Style
}

// NewStyles creates a new Styles registry for the given theme
func NewStyles(theme string) Styles {
	palette := animations.GetFirePalette(theme)

	// Extract colors from palette
	// Palettes are ordered from dark to light (coolest to hottest)
	var bg, primary, secondary, accent, hot string
	if len(palette) > 0 {
		bg = palette[0]
	}
	if len(palette) > 1 {
		primary = palette[1]
	}
	if len(palette) > 2 {
		secondary = palette[2]
	}
	if len(palette) > 3 {
		accent = palette[3]
	}
	if len(palette) > 4 {
		hot = palette[len(palette)-1] // Hottest color
	}

	// Fallback colors
	if bg == "" {
		bg = "#1e1e2e"
	}
	if primary == "" {
		primary = "#313244"
	}
	if secondary == "" {
		secondary = "#45475a"
	}
	if accent == "" {
		accent = "#89dceb"
	}
	if hot == "" {
		hot = "#f38ba8"
	}

	return Styles{
		// Persistent header
		Header: lipgloss.NewStyle().
			Background(lipgloss.Color(primary)).
			Foreground(lipgloss.Color(accent)).
			Bold(true).
			Padding(0, 1).
			Width(80),

		HeaderTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(hot)).
			Bold(true).
			Height(1),

		HeaderSub: lipgloss.NewStyle().
			Foreground(lipgloss.Color(accent)).
			Faint(true).
			Height(1),

		// Messages
		UserMsg: lipgloss.NewStyle().
			Foreground(lipgloss.Color(hot)).
			Bold(true).
			MarginTop(1).
			Padding(0, 1),

		AssistantMsg: lipgloss.NewStyle().
			Foreground(lipgloss.Color(accent)).
			MarginTop(1).
			Padding(0, 1),

		SystemMsg: lipgloss.NewStyle().
			Foreground(lipgloss.Color(secondary)).
			Faint(true).
			Italic(true).
			MarginTop(1).
			Padding(0, 1),

		ToolCall: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EBCB8B")).
			Italic(true).
			MarginTop(1).
			Padding(0, 1),

		ToolResult: lipgloss.NewStyle().
			Foreground(lipgloss.Color(secondary)).
			Background(lipgloss.Color(primary)).
			MarginTop(0).
			MarginBottom(1).
			Padding(0, 2),

		// Status and errors
		Status: lipgloss.NewStyle().
			Foreground(lipgloss.Color(accent)).
			Padding(0, 1),

		StatusErr: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#BF616A")).
			Bold(true).
			Padding(0, 1),

		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#BF616A")).
			Bold(true).
			Padding(0, 1),

		Warning: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EBCB8B")).
			Padding(0, 1),

		// Input area
		InputActive: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(accent)).
			Background(lipgloss.Color(primary)).
			Padding(0, 1),

		InputInactive: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(secondary)).
			Background(lipgloss.Color(bg)).
			Padding(0, 1),

		InputPlaceholder: lipgloss.NewStyle().
			Foreground(lipgloss.Color(secondary)).
			Faint(true),

		// Viewport
		Viewport: lipgloss.NewStyle().
			Background(lipgloss.Color(bg)).
			Height(20),

		ViewportBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(secondary)).
			Padding(0, 1),

		// Component borders
		Border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(secondary)).
			Padding(0, 1),

		BorderFocused: lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color(accent)).
			Padding(0, 1),

		// Utility
		Dim: lipgloss.NewStyle().
			Foreground(lipgloss.Color(secondary)).
			Padding(0, 1),

		Muted: lipgloss.NewStyle().
			Foreground(lipgloss.Color(secondary)).
			Faint(true).
			Padding(0, 1),

		Highlight: lipgloss.NewStyle().
			Foreground(lipgloss.Color(hot)).
			Bold(true).
			Padding(0, 1),
	}
}

// DefaultStyles returns styles using the default theme
func DefaultStyles() Styles {
	return NewStyles("catppuccin")
}
