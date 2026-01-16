package header

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the header state
type Model struct {
	Width       int
	Frame       int
	ActiveModel string
	LSPStatus   string
	MCPStatus   string
	ContextInfo string

	// Styles
	StyleStripe  lipgloss.Style
	StyleBrand   lipgloss.Style
	StyleMeta    lipgloss.Style
	StyleMetaDim lipgloss.Style

	// Gradient Cache
	GradientColors []lipgloss.Color
}

// New creates a new header model
func New() Model {
	// Initialize with default gradient (Neon Blue -> Slate)
	// We'll generate a 20-step gradient for smooth animation
	colors := make([]lipgloss.Color, 20)
	// Simple simulated gradient for now, can be expanded with true interpolation
	// For now we alternate between Accent and Cyan
	for i := 0; i < 20; i++ {
		if i < 10 {
			colors[i] = lipgloss.Color("#58a6ff") // Blue
		} else {
			colors[i] = lipgloss.Color("#bc8cff") // Purple
		}
	}

	return Model{
		ActiveModel:    "GLM-4.7",
		LSPStatus:      "None",
		MCPStatus:      "None",
		ContextInfo:    "8k",
		GradientColors: colors,
		StyleBrand: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#f0f6fc")). // Bright White
			Padding(0, 1),
		StyleMeta: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#58a6ff")). // Accent
			Padding(0, 1),
		StyleMetaDim: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8b949e")). // Muted
			Padding(0, 0),
	}
}

// Init initializes the component
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg.(type) {
	// We assume a global TickMsg will drive the frame animation
	// tailored in the parent model
	}
	return m, nil
}

// Tick increments the animation frame
func (m *Model) Tick() {
	m.Frame++
}

// View renders the header
func (m Model) View() string {
	if m.Width <= 0 {
		return ""
	}

	// 1. Texture Layer (The Stripes)
	stripes := m.renderStripes()

	// 2. Data Layer (The Grid)
	// Left: Brand
	brand := m.StyleBrand.Render("FLOYD v0.32.0")

	// Right: Indicators
	// [LSPs: Active] [MCPs: None] [Model: GLM-4.7]
	indicators := lipgloss.JoinHorizontal(lipgloss.Top,
		m.renderMetaItem("LSPs", m.LSPStatus),
		m.renderMetaItem("MCPs", m.MCPStatus),
		m.renderMetaItem("Model", m.ActiveModel),
		m.renderMetaItem("Ctx", m.ContextInfo),
	)

	// Spacer
	availableSpace := m.Width - lipgloss.Width(brand) - lipgloss.Width(indicators)
	spacer := strings.Repeat(" ", max(0, availableSpace))

	// Combine
	metaRow := lipgloss.JoinHorizontal(lipgloss.Top,
		brand,
		spacer,
		indicators,
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		stripes,
		metaRow,
	)
}

func (m Model) renderStripes() string {
	var sb strings.Builder
	pattern := "////////////"
	plen := len(pattern)
	glen := len(m.GradientColors)

	// Loop width
	for i := 0; i < m.Width; i++ {
		// Calculate color index based on position + frame (motion)
		// We stretch the gradient across the bar + animate offset
		idx := (i + m.Frame) % (glen * 2)
		if idx >= glen {
			idx = (glen * 2) - idx - 1 // bounce back or cycle? let's just cycle
		}
		if idx >= glen {
			idx = 0
		} // Safety

		style := lipgloss.NewStyle().Foreground(m.GradientColors[idx])

		// Pick char from pattern
		char := string(pattern[i%plen])
		sb.WriteString(style.Render(char))
	}
	return sb.String()
}

func (m Model) renderMetaItem(label, value string) string {
	lbl := m.StyleMetaDim.Render(label)
	val := m.StyleMeta.Render(value)
	return lbl + " " + val + "  "
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
