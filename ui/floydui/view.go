package floydui

import (
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// View renders the UI
func (m Model) View() string {
	if m.Quitting {
		return ""
	}

	// Check for minimum height to prevent crashes
	if m.Height < 10 {
		return Styles.Error.
			Width(m.Width).
			Height(m.Height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("Terminal too small.")
	}

	// 1. Render Header (Restored ASCII Art)
	// Calculate banner text based on width
	var bannerText string
	if m.Width < 250 {
		bannerText = SmallBanner
	} else {
		bannerText = PRISMBanner
	}
	header := ColorizeBanner(bannerText, m.CurrentTheme, m.AnimationFrame/2)

	// 2. Render Main Content (Viewport or Welcome)
	var middleContent string
	if len(m.Messages) == 0 {
		// Welcome Screen
		tips := []string{
			"FLOYD Intelligent Terminal",
			"Type /help for commands",
			"",
			"Using: " + m.CurrentTheme.Name,
		}
		middleContent = Styles.Help.Render(strings.Join(tips, "\n"))
	} else {
		// Chat Viewport
		// We rely on updateViewportContent to keep the buffer fresh
		middleContent = Styles.Viewport.Render(m.Viewport.View())
	}

	// 3. Render Input
	styledInput := Styles.Input.Width(m.Width - 4).Render(m.Input.View())

	// 4. Render Footer (Managed by component)
	footer := m.Footer.View()

	// 5. Progress Bar (if active)
	var progressBar string
	if m.ProgressActive {
		progressBar = Styles.Progress.Render(m.Progress.View())
	}

	// Calculate layout
	// Header + Footer + Input are fixed. Viewport fills the rest.
	// We handle this sizing in handleWindowSize, so here we just Join.

	mainContent := lipgloss.JoinVertical(lipgloss.Left,
		header,
		middleContent,
		progressBar,
		styledInput,
		footer, // Replaces old Status Bar
	)

	// 6. Overlay Layer (Command Palette & Model Select)
	if m.Mode == ModeCommand || m.Mode == ModeModelSelect {
		// Render Palette roughly centered
		// For simplicity in this iteration, we append it or use a separate view return
		// To truly overlay, we'd need complex whitespace calculation.
		// A decent TUI trick: Just return the Palette View if it's "Modal" enough.
		// But better: Render it at the bottom anchor or top anchor.
		// Let's render it on top of the input area for now.
		return lipgloss.JoinVertical(lipgloss.Left,
			header,
			middleContent,
			m.Palette.View(), // Overlay input
			footer,
		)
	}

	return Styles.Base.Render(mainContent)
}

// buildStatusText is deprecated, replaced by Footer component
func (m Model) buildStatusText() string {
	return ""
}

// SimpleView returns a simplified view for CLI mode (no TUI)
func (m Model) SimpleView(input string) string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("=", 60))
	sb.WriteString("\n")

	if input != "" {
		sb.WriteString(Styles.UserMsg.Render("YOU: " + input))
		sb.WriteString("\n")
	}

	for _, msg := range m.Messages {
		switch msg.Role {
		case "user":
			sb.WriteString(Styles.UserMsg.Render("YOU: " + msg.Content))
		case "assistant":
			sb.WriteString(Styles.FloydMsg.Render("FLOYD: " + msg.Content))
		case "system":
			if msg.ToolUse != nil {
				sb.WriteString(Styles.ToolUse.Render("▶ Tool: " + msg.ToolUse.Name))
			} else if msg.ToolResult != nil {
				prefix := "✓ "
				if msg.ToolResult.IsError {
					prefix = "✗ "
				}
				sb.WriteString(prefix + msg.ToolResult.Name)
			} else {
				sb.WriteString(Styles.SystemMsg.Render(msg.Content))
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString(strings.Repeat("=", 60))
	sb.WriteString("\n> ")

	return sb.String()
}

// ColorizeBanner applies dynamic, monochrome + blue shimmer to the ASCII art
func ColorizeBanner(text string, t Theme, frame int) string {
	var sb strings.Builder

	// Palette for the shimmering monolith effect
	palette := []lipgloss.Color{
		t.Accent,
		t.Muted,
		t.Cyan,
		t.Subtle,
	}

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		// Render all lines, even empty ones, to preserve banner shape/spacing

		for j, ch := range line {
			char := string(ch)
			if char == " " {
				sb.WriteString(" ")
				continue
			}

			// Diagonal pulse logic
			pulseIdx := (i + j/2 + frame) % (len(palette) * 4)
			colorIdx := pulseIdx / 4
			if colorIdx >= len(palette) {
				colorIdx = 0
			}

			style := lipgloss.NewStyle().Foreground(palette[colorIdx])

			// Character-specific shading overrides
			switch char {
			case "\\", "/":
				sb.WriteString(style.Bold(true).Render(char))
			case "_":
				sb.WriteString(lipgloss.NewStyle().Foreground(t.Muted).Render(char))
			case "|":
				sb.WriteString(lipgloss.NewStyle().Foreground(t.Cyan).Bold(true).Render(char))
			default:
				sb.WriteString(style.Render(char))
			}
		}
		sb.WriteString("\n")
	}

	return Styles.Header.Render(sb.String())
}

// RenderMarkdown renders markdown text using the current theme
func RenderMarkdown(text string, width int, theme Theme) string {
	// Custom style JSON for high contrast on dark slate
	const styleJSON = `{
		"document": {
			"block_prefix": "",
			"block_suffix": "",
			"color": "#FFFFFF" 
		},
		"block_quote": {
			"indent": 2,
			"indent_token": "│ "
		},
		"list": {
			"level_01": "• ",
			"level_02": "+ ",
			"level_03": "* "
		},
		"heading": {
			"block_prefix": "\n",
			"block_suffix": "\n",
			"color": "#58a6ff",
			"bold": true
		},
		"code_block": {
			"margin": 2
		}
	}`

	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
		glamour.WithStylesFromJSONBytes([]byte(styleJSON)),
	)
	if err != nil {
		return text // Fallback
	}

	out, err := r.Render(text)
	if err != nil {
		return text // Fallback
	}

	return strings.TrimSpace(out)
}
