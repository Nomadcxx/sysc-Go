package footer

import (
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Mode constants to match parent (will act as localized integers)
const (
	ModeChat = iota
	ModeCommand
	ModeModelSelect
)

type Model struct {
	Width       int
	CurrentMode int
	Spinner     spinner.Model
	IsThinking  bool

	// Styles
	StyleTextDim    lipgloss.Style
	StyleTextActive lipgloss.Style
	StyleSpinner    lipgloss.Style
}

func New() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#58a6ff")) // Neon Blue

	return Model{
		Spinner: s,
		StyleTextDim: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8b949e")), // Dimmed
		StyleTextActive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f0f6fc")), // Bright
		StyleSpinner: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#58a6ff")), // Neon
	}
}

func (m Model) Init() tea.Cmd {
	return m.Spinner.Tick
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	if m.IsThinking {
		m.Spinner, cmd = m.Spinner.Update(msg)
	}
	return m, cmd
}

func (m Model) View() string {
	// Left: Status
	status := m.renderStatus()

	// Right: Key Hints
	hints := m.renderHints()

	// Spacer
	space := m.Width - lipgloss.Width(status) - lipgloss.Width(hints)
	spacer := strings.Repeat(" ", max(0, space))

	return lipgloss.JoinHorizontal(lipgloss.Top,
		status,
		spacer,
		hints,
	)
}

func (m Model) renderStatus() string {
	if m.IsThinking {
		return m.StyleSpinner.Render(m.Spinner.View()) + m.StyleTextActive.Render(" Thinking...")
	}
	return m.StyleTextDim.Render("> Ready...")
}

func (m Model) renderHints() string {
	var keys []string

	// Define hints based on Mode
	switch m.CurrentMode {
	case ModeChat:
		keys = []string{"^P commands", "^L models", "Enter send"}
	case ModeCommand:
		keys = []string{"↑/↓ navigate", "Enter select", "Esc cancel"}
	case ModeModelSelect:
		keys = []string{"/ filter", "Enter confirm", "Esc back"}
	default:
		keys = []string{"Esc back"}
	}

	// Render pills
	var rendered []string
	for _, k := range keys {
		parts := strings.Split(k, " ")
		if len(parts) >= 2 {
			// Key in bright, Desc in dim
			keyStr := m.StyleTextActive.Render(parts[0])
			descStr := m.StyleTextDim.Render(strings.Join(parts[1:], " "))
			rendered = append(rendered, keyStr+" "+descStr)
		} else {
			rendered = append(rendered, m.StyleTextDim.Render(k))
		}
	}

	return strings.Join(rendered, m.StyleTextDim.Render("  •  "))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
