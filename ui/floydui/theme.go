package floydui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Theme defines a color scheme for the FLOYD UI
type Theme struct {
	Name string

	// Core Surfaces
	Black lipgloss.Color // Main Terminal Background
	Glass lipgloss.Color // Secondary/Modal Background ("Glass")

	// Text
	White lipgloss.Color // Primary/Emphasis Text
	Gray  lipgloss.Color // Body/Standard Text (The "Quiet" default)

	// Accents
	Accent lipgloss.Color // Primary Brand Accent
	Muted  lipgloss.Color // Secondary/Meta info (Dimmed)
	Subtle lipgloss.Color // Very faint background/details

	// Borders
	Border       lipgloss.Color // Inactive/Standard border
	ActiveBorder lipgloss.Color // Focused/Glow border

	// Functional Semantics
	Error  lipgloss.Color // Critical Errors
	Yellow lipgloss.Color // Warning/Pending
	Cyan   lipgloss.Color // Info/Structure
	Green  lipgloss.Color // Success
}

// All available themes
var Themes = map[string]Theme{
	"classic": {
		Name:         "Legacy Silver",
		Black:        "#000000",
		Glass:        "#111111",
		White:        "#FFFFFF",
		Gray:         "#DDDDDD",
		Accent:       "#FAFAFA",
		Muted:        "#BDBDBD",
		Subtle:       "#333333",
		Border:       "#444444",
		ActiveBorder: "#FFFFFF",
		Error:        "#FF4D4D",
		Yellow:       "#90A4AE",
		Cyan:         "#90CAF9",
		Green:        "#A5D6A7",
	},
	"dark": {
		Name:         "Legacy Noir",
		Black:        "#000000",
		Glass:        "#121212",
		White:        "#FFFFFF",
		Gray:         "#E0E0E0",
		Accent:       "#E0E0E0",
		Muted:        "#757575",
		Subtle:       "#212121",
		Border:       "#333333",
		ActiveBorder: "#888888",
		Error:        "#D32F2F",
		Yellow:       "#B0BEC5",
		Cyan:         "#64B5F6",
		Green:        "#4DB6AC",
	},
	"highcontrast": {
		Name:         "High Contrast",
		Black:        "#000000",
		Glass:        "#000000",
		White:        "#FFFFFF",
		Gray:         "#FFFFFF",
		Accent:       "#FFFFFF",
		Muted:        "#CCCCCC",
		Subtle:       "#444444",
		Border:       "#FFFFFF",
		ActiveBorder: "#FFFF00",
		Error:        "#FF0000",
		Yellow:       "#FFFF00",
		Cyan:         "#00FFFF",
		Green:        "#00FF00",
	},
	"monolith": {
		Name:         "Monolith",
		Black:        "#000000",
		Glass:        "#101010",
		White:        "#FFFFFF",
		Gray:         "#EEEEEE",
		Accent:       "#E0E0E0",
		Muted:        "#9E9E9E",
		Subtle:       "#101010",
		Border:       "#333333",
		ActiveBorder: "#FFFFFF",
		Error:        "#FF5252",
		Yellow:       "#FFD740",
		Cyan:         "#40C4FF",
		Green:        "#69F0AE",
	},

	// "Quiet Confidence" - The New Standard (Deep Slate)
	"midnight": {
		Name:         "Showcase Slate", // Renamed to reflect its status
		Black:        "#0d1117",        // Deep Slate (GitHub Dimmed)
		Glass:        "#161b22",        // Lighter Slate (Panels)
		White:        "#f0f6fc",        // Bright White (Headers only)
		Gray:         "#c9d1d9",        // Soft Gray (Body text)
		Accent:       "#58a6ff",        // Neon Blue (Focus)
		Muted:        "#8b949e",        // Muted Gray (Meta)
		Subtle:       "#21262d",        // Subtle backgrounds
		Border:       "#30363d",        // Subtle border
		ActiveBorder: "#58a6ff",        // Active Blue Glow
		Error:        "#f85149",
		Yellow:       "#d29922",
		Cyan:         "#58a6ff", // Unifying Cyan with Accent for consistency
		Green:        "#3fb950",
	},
}

// DefaultTheme is the default theme to use
var DefaultTheme = Themes["midnight"]

// GetTheme returns a theme by name, or default if not found
func GetTheme(name string) Theme {
	if t, ok := Themes[name]; ok {
		return t
	}
	return DefaultTheme
}

// ThemeNames returns a list of all available theme names
func ThemeNames() []string {
	names := make([]string, 0, len(Themes))
	for name := range Themes {
		names = append(names, name)
	}
	return names
}

// GetThemeNames returns a sorted list of theme names
func GetThemeNames() []string {
	return ThemeNames()
}

// GetThemeByName returns a pointer to a theme by name, or nil if not found
func GetThemeByName(name string) *Theme {
	// Try exact match first
	if t, ok := Themes[name]; ok {
		return &t
	}
	// Try case-insensitive match
	for key, theme := range Themes {
		if strings.EqualFold(key, name) || strings.EqualFold(theme.Name, name) {
			return &theme
		}
	}
	return nil
}
