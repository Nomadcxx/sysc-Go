package floydui

import (
	"fmt"
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

	// Add Right Corner Logo (CRUSH style)
	rightLogo := lipgloss.NewStyle().
		Foreground(m.CurrentTheme.Muted).
		Align(lipgloss.Right).
		Padding(1, 2).
		Render(fmt.Sprintf("FLOYD v2.0\n%s\nCRUSH", strings.ToUpper(m.CurrentTheme.Name)))

	// Combine Header and Logo
	headerWidth := lipgloss.Width(header)
	logoWidth := lipgloss.Width(rightLogo)
	padding := m.Width - headerWidth - logoWidth

	var fullHeader string
	if padding > 0 {
		fullHeader = lipgloss.JoinHorizontal(lipgloss.Top, header, strings.Repeat(" ", padding), rightLogo)
	} else {
		fullHeader = header
	}

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
		// Render whimsical thinking phrase
		// Ensure index is valid
		idx := m.CurrentThinking
		if idx < 0 || idx >= len(ThinkingStates) {
			idx = 0
		}

		phrase := ThinkingStates[idx].Phrase
		// We can include the description/sound effect if we want, or just the phrase.
		// User asked for "raunchy whimsy", so likely the full text or phrase.
		// Let's do: Phrase + " " + Description (dimmed)

		text := Styles.Thinking.Render(phrase)
		// desc := Styles.Subtle.Render(ThinkingStates[idx].Description) // Optional

		bar := Styles.Progress.Render(m.Progress.View())
		progressBar = lipgloss.JoinVertical(lipgloss.Left, text, bar)
	}

	// Calculate layout
	// Header + Footer + Input are fixed. Viewport fills the rest.
	// We handle this sizing in handleWindowSize, so here we just Join.

	mainContent := lipgloss.JoinVertical(lipgloss.Left,
		fullHeader,
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
			fullHeader,
			middleContent,
			m.Palette.View(), // Overlay input
			footer,
		)
	}

	// 7. Audit Exit Summary
	if m.Mode == ModeExitSummary {
		return m.RenderAuditSummary()
	}

	return Styles.Base.Render(mainContent)
}

// RenderAuditSummary renders the session summary
func (m Model) RenderAuditSummary() string {
	// Calculate stats
	totalMessages := len(m.Messages)
	userMessages := 0
	assistantMessages := 0
	toolUses := 0
	toolErrors := 0

	for _, msg := range m.Messages {
		if msg.Role == "user" {
			userMessages++
		}
		if msg.Role == "assistant" {
			assistantMessages++
		}
		if msg.ToolUse != nil {
			toolUses++
		}
		if msg.ToolResult != nil && msg.ToolResult.IsError {
			toolErrors++
		}
	}

	successRate := 100.0
	if toolUses > 0 {
		successRate = float64(toolUses-toolErrors) / float64(toolUses) * 100.0
	}

	// Mock tokens for now (avg 4 chars/token)
	totalChars := 0
	for _, msg := range m.Messages {
		totalChars += len(msg.Content)
	}
	estimatedTokens := totalChars / 4
	estimatedCost := float64(estimatedTokens) * 0.000015 // rough blended rate

	// Render box
	// Using the floating frame style (AuditBox)
	// We want a "Crush" aesthetic - maybe add the Right Corner Logo if we can position it.
	// For now, a clean centered box.

	content := fmt.Sprintf(
		"SESSION AUDIT\n\n"+
			"• Messages:  %d\n"+
			"  - User:    %d\n"+
			"  - Floyd:   %d\n"+
			"• Tools:     %d\n"+
			"  - Errors:  %d\n"+
			"• Success:   %.1f%%\n"+
			"• Est. Cost: $%.5f (%d tokens)\n\n"+
			"Press Ctrl+C again to exit.",
		totalMessages,
		userMessages,
		assistantMessages,
		toolUses, toolErrors,
		successRate,
		estimatedCost, estimatedTokens,
	)

	summaryBox := Styles.AuditBox.Render(content)

	// Place in center of screen
	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, summaryBox)
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

// ColorizeBanner applies dynamic colors to the ASCII art
func ColorizeBanner(text string, t Theme, frame int) string {
	// Fallback for SmallBanner or non-standard text
	if !strings.Contains(text, ".----------------.") {
		return lipgloss.NewStyle().Foreground(t.Accent).Render(text)
	}

	var sb strings.Builder
	lines := strings.Split(text, "\n")

	// Color grid from user (Rows x Cols)
	// Rows map to vertical height (approx 2 lines per row)
	// Cols map to the 5 cards (F, L, O, Y, D)
	grid := [][]string{
		{"17", "18", "19", "20", "21"},      // Dark Blues -> Blue
		{"53", "54", "55", "56", "57"},      // Purples -> Blue-Purple
		{"89", "90", "91", "92", "93"},      // Red-Purple -> Purple
		{"125", "126", "127", "128", "129"}, // Red-Pink -> Purple
		{"161", "162", "163", "164", "165"}, // Red -> Pink
		{"197", "198", "199", "200", "201"}, // Bright Red -> Hot Pink
	}

	// Each card is roughly 22 chars wide (including spacing)
	const cardWidth = 22

	for i, line := range lines {
		// Map line index (0-10) to grid row (0-5)
		// 11 lines total.
		// Lines 0,1 -> Row 0
		// Lines 2,3 -> Row 1
		// ...
		gridRow := i / 2
		if gridRow >= len(grid) {
			gridRow = len(grid) - 1
		}

		for j, ch := range line {
			char := string(ch)
			if char == " " || char == "\t" || char == "\n" {
				sb.WriteString(char)
				continue
			}

			// Map column index to grid column (0-4)
			gridCol := j / cardWidth
			if gridCol >= len(grid[0]) {
				gridCol = len(grid[0]) - 1
			}

			colorCode := grid[gridRow][gridCol]
			style := lipgloss.NewStyle().Foreground(lipgloss.Color(colorCode))

			sb.WriteString(style.Render(char))
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
