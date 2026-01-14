package floydui

import (
	"github.com/charmbracelet/lipgloss"
)

// StyleSet contains all dynamically generated styles based on the current theme
type StyleSet struct {
	Base             lipgloss.Style
	GlassPanel       lipgloss.Style
	Header           lipgloss.Style
	Tagline          lipgloss.Style
	Viewport         lipgloss.Style
	UserMsg          lipgloss.Style
	FloydMsg         lipgloss.Style
	SystemMsg        lipgloss.Style
	Error            lipgloss.Style
	Thinking         lipgloss.Style
	MicroInteraction lipgloss.Style
	Input            lipgloss.Style
	StatusBar        lipgloss.Style
	Help             lipgloss.Style
	ToolUse          lipgloss.Style
	ToolResult       lipgloss.Style
	ToolResultError  lipgloss.Style
	Progress         lipgloss.Style
}

// GenerateStyles creates all styles based on the given theme
func GenerateStyles(t Theme) StyleSet {
	// Clean, no-background base style
	base := lipgloss.NewStyle().Foreground(t.White)

	return StyleSet{
		Base: lipgloss.NewStyle().
			Background(t.Black).
			Foreground(t.White),

		// GlassPanel replaced with transparent container
		GlassPanel: lipgloss.NewStyle().
			Padding(0, 1),

		Header: base.Copy().
			Foreground(t.Accent).
			Align(lipgloss.Left).
			Padding(3, 0, 1, 1),

		// Tagline removed/simplified in new layout or used as subtle text
		Tagline: base.Copy().
			Foreground(t.Cyan).
			Align(lipgloss.Left).
			Padding(0, 0, 1, 1),

		Viewport: lipgloss.NewStyle().
			Padding(0, 1),

		UserMsg: lipgloss.NewStyle().
			Foreground(t.White).
			Padding(0, 0).
			Bold(true).
			MarginTop(1),

		FloydMsg: lipgloss.NewStyle().
			Foreground(t.White).
			Padding(0, 0),

		SystemMsg: lipgloss.NewStyle().
			Foreground(t.Cyan),

		Error: lipgloss.NewStyle().
			Foreground(t.Error).
			Padding(1),

		Thinking: base.Copy().
			Foreground(t.Yellow).
			Italic(true),

		MicroInteraction: base.Copy().
			Foreground(t.Accent),

		// Gemini-style Input Box - Modern, padded, and themed
		Input: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.Accent).
			Padding(1, 2).
			MarginTop(0),

		StatusBar: lipgloss.NewStyle().
			Foreground(t.Cyan).
			Padding(0, 1),

		Help: lipgloss.NewStyle().
			Foreground(t.Cyan).
			Padding(1, 2),

		ToolUse: lipgloss.NewStyle().
			Foreground(t.Yellow).
			Bold(true).
			MarginTop(1),

		ToolResult: lipgloss.NewStyle().
			Foreground(t.Green),

		ToolResultError: lipgloss.NewStyle().
			Foreground(t.Error).
			Bold(true),

		Progress: lipgloss.NewStyle().
			Foreground(t.Accent).
			Width(40),
	}
}

// Global styles (will be set when theme changes)
var Styles StyleSet

// SetTheme updates the global styles based on the given theme
func SetTheme(t Theme) {
	Styles = GenerateStyles(t)
}

// InitStyles initializes styles with the default theme
func InitStyles() {
	Styles = GenerateStyles(DefaultTheme)
}

// InitStylesWithTheme initializes styles with a specific theme
func InitStylesWithTheme(theme Theme) {
	Styles = GenerateStyles(theme)
}
