package palette

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Item represents a command in the palette
type Item struct {
	Title       string
	Description string
	Cmd         tea.Cmd
}

func (i Item) FilterValue() string { return i.Title }

// ItemDelegate handles the rendering of list items with "Glass" styling
type ItemDelegate struct {
	Styles itemStyles
}

type itemStyles struct {
	NormalTitle   lipgloss.Style
	NormalDesc    lipgloss.Style
	SelectedTitle lipgloss.Style
	SelectedDesc  lipgloss.Style
}

func newItemStyles() itemStyles {
	return itemStyles{
		NormalTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#c9d1d9")). // Gray
			PaddingLeft(2),
		NormalDesc: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8b949e")). // Dimmed
			PaddingLeft(2),
		SelectedTitle: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true). // Left Border
			BorderForeground(lipgloss.Color("#58a6ff")).                // Neon Blue
			Foreground(lipgloss.Color("#f0f6fc")).                      // Bright White
			Bold(true).
			PaddingLeft(1),
		SelectedDesc: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8b949e")). // Muted
			PaddingLeft(2),
	}
}

func (d ItemDelegate) Height() int                             { return 2 }
func (d ItemDelegate) Spacing() int                            { return 0 }
func (d ItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Item)
	if !ok {
		return
	}

	s := newItemStyles()

	if index == m.Index() {
		fmt.Fprint(w, s.SelectedTitle.Render("> "+i.Title))
		fmt.Fprint(w, "\n")
		fmt.Fprint(w, s.SelectedDesc.Render(i.Description))
	} else {
		fmt.Fprint(w, s.NormalTitle.Render(i.Title))
		fmt.Fprint(w, "\n")
		fmt.Fprint(w, s.NormalDesc.Render(i.Description))
	}
}

// Model is the palette component
type Model struct {
	List   list.Model
	Active bool
	Width  int
	Height int
}

func New() Model {
	items := []list.Item{
		Item{Title: "New Session", Description: "Start a fresh context (Ctrl+N)"},
		Item{Title: "Switch Model", Description: "Change active LLM (Ctrl+L)"},
		Item{Title: "Toggle Help", Description: "View keyboard shortcuts"},
		Item{Title: "Quit", Description: "Exit application"},
	}

	// Setup delegate
	delegate := ItemDelegate{}

	// Setup List
	l := list.New(items, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.Styles.PaginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	l.Styles.HelpStyle = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)

	return Model{
		List:   l,
		Active: false,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	// Wrap in Glass Box
	glassStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#58a6ff")). // Neon Blue
		Background(lipgloss.Color("#161b22")).       // Glass Slate
		Padding(1, 1).
		Width(50).
		Height(14)

	return glassStyle.Render(m.List.View())
}

func (m *Model) SetSize(w, h int) {
	m.Width = w
	m.Height = h
	// The palette itself is fixed size usually,
	// but we might want to center it or adjust list height
	m.List.SetSize(46, 12) // Slightly smaller than box
}

// SetItems updates the items in the list
func (m *Model) SetItems(items []list.Item) {
	m.List.SetItems(items)
}
