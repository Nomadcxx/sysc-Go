package agenttui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the TUI
func (m AgentModel) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	// Check for minimum dimensions
	if m.width < 60 || m.height < 15 {
		return m.renderTooSmall()
	}

	// Build the main layout
	var sections []string

	// 1. Persistent header (4 lines)
	sections = append(sections, m.renderHeader())

	// 2. Viewport (main content area)
	sections = append(sections, m.renderViewport())

	// 3. Status bar
	sections = append(sections, m.renderStatus())

	// 4. Input area
	sections = append(sections, m.renderInput())

	// 5. Help footer
	sections = append(sections, m.renderHelp())

	// Join and render
	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Render(content)
}

// renderHeader renders the persistent header
func (m AgentModel) renderHeader() string {
	// Header background
	headerStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#313244")).
		Foreground(lipgloss.Color("#89dceb")).
		Bold(true).
		Padding(0, 2).
		Width(m.width)

	var sb strings.Builder

	// Title
	sb.WriteString(m.headerText)

	// Right-side info (if space allows)
	if m.width > 50 {
		info := " GLM4.7 "
		if m.streaming {
			info = " ● Streaming"
		} else if m.toolRunning {
			info = " ● Tool Running"
		}
		sb.WriteString(strings.Repeat(" ", m.width-40-len(m.headerText)))
		sb.WriteString(info)
	}

	header := headerStyle.Render(sb.String())

	// Subtitle line
	subStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6c7086")).
		Faint(true).
		Padding(0, 2)

	sub := subStyle.Width(m.width).Render(m.headerSub)

	// Divider
	dividerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#45475a"))
	divider := dividerStyle.Render(strings.Repeat("─", m.width))

	return header + "\n" + sub + "\n" + divider
}

// renderViewport renders the conversation viewport
func (m AgentModel) renderViewport() string {
	// Viewport container style
	containerStyle := m.styles.ViewportBorder.
		Width(m.width - 4).
		Height(m.height - 12) // Reserve space for header, status, input, help

	viewportContent := m.viewport.View()

	// If empty, show welcome message
	if strings.TrimSpace(viewportContent) == "" {
		viewportContent = m.renderWelcome()
	}

	return containerStyle.Render(viewportContent)
}

// renderWelcome renders the welcome message
func (m AgentModel) renderWelcome() string {
	welcomeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6c7086")).
		Italic(true).
		Align(lipgloss.Center)

	welcome := `╷
│  FLOYD Agent Console
│
│  Type your message below to begin.
│  Press ESC for normal mode, 'i' for insert mode.
│  Press Ctrl+C to quit.
╵`

	return welcomeStyle.Render(welcome)
}

// renderStatus renders the status bar
func (m AgentModel) renderStatus() string {
	var status string

	if m.status != "" {
		status = m.status
	} else {
		if m.streaming {
			status = "● Receiving response..."
		} else if m.toolRunning {
			status = "● Running tool..."
		} else {
			status = "Ready"
		}
	}

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#89dceb")).
		Background(lipgloss.Color("#1e1e2e")).
		Padding(0, 1).
		Width(m.width - 4)

	return statusStyle.Render(" " + status)
}

// renderInput renders the input area
func (m AgentModel) renderInput() string {
	inputStyle := m.styles.InputActive.
		Width(m.width - 4).
		Height(7)

	return inputStyle.Render(m.input.View())
}

// renderHelp renders the help footer
func (m AgentModel) renderHelp() string {
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6c7086")).
		Faint(true).
		Padding(0, 1)

	var help string
	if m.input.Mode() == ModeNormal {
		help = " [NORMAL] i:insert j/k:history enter:submit q:quit"
	} else {
		help = " [INSERT] Type your message | esc:normal | enter:submit | ctrl+c:quit"
	}

	return helpStyle.Width(m.width).Render(" " + help)
}

// renderTooSmall renders a message when terminal is too small
func (m AgentModel) renderTooSmall() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#f38ba8")).
		Bold(true).
		Align(lipgloss.Center)

	msg := `Terminal Too Small!

Current: %dx%d
Minimum: 60x15

Please resize your terminal.
Press Ctrl+C to quit.`

	return style.Render(
		sprintf(msg, m.width, m.height),
	)
}

// sprintf is a simple string formatter
func sprintf(format string, args ...int) string {
	result := format
	for i, arg := range args {
		placeholder := "%d"
		if i > 0 {
			placeholder = "%" + itoa(i+1) + "d"
		}
		result = strings.ReplaceAll(result, placeholder, itoa(arg))
	}
	return result
}
