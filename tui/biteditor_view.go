package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderBitEditorView renders the BIT text editor interface
func (m Model) renderBitEditorView() string {
	if m.bitShowFontList {
		return m.renderFontBrowser()
	}

	if m.bitColorPicker {
		return m.renderColorPicker()
	}

	if m.showExportPrompt {
		return m.renderExportPrompt() // Reuse existing export prompt
	}

	if m.showSavePrompt {
		return m.renderBitSavePrompt()
	}

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#88C0D0")).
		Padding(1, 0).
		Render("BIT Text Editor - Banner Text Generator")

	// Text input
	input := m.renderBitTextInput()

	// Controls
	controls := m.renderBitControls()

	// Help text
	help := m.renderBitHelp()

	// The banner can be far taller than the terminal, so give the preview
	// whatever the rest leaves over and clip the rest away.  Measuring the
	// chrome keeps this honest when any of these blocks changes height.
	chrome := lipgloss.Height(lipgloss.JoinVertical(lipgloss.Left, title, input, controls, help))
	sections := []string{title, m.renderBitPreview(m.bitPreviewHeight(chrome)), input, controls, help}

	// No background wrapping to prevent bleeding
	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return content
}

// bitPreviewRowsFrame is what the preview's own rounded border and padding
// cost, beyond the preview lines themselves.
const bitPreviewRowsFrame = 4

// bitPreviewHeight is how many preview rows fit below chrome rows of editor
// furniture, once the preview frame's own border and padding are paid for.
func (m Model) bitPreviewHeight(chrome int) int {
	h := m.height - chrome - bitPreviewRowsFrame
	if h < 1 {
		return 1
	}
	return h
}

// renderBitPreview renders the live preview canvas, showing at most maxRows
// rows of a banner that does not fit.
func (m Model) renderBitPreview(maxRows int) string {
	var preview string
	if len(m.bitPreviewLines) > 0 {
		lines := m.bitPreviewLines
		truncated := false
		// The notice is a preview row too, so a clipped banner gives one up.
		keep := len(lines)
		if keep > maxRows {
			keep = maxRows
			if maxRows > 1 {
				keep = maxRows - 1
				truncated = true
			}
		}
		preview = strings.Join(lines[len(lines)-keep:], "\n")
		if truncated {
			preview = "… " + strconv.Itoa(len(m.bitPreviewLines)-keep) + " earlier rows hidden\n" + preview
		}
	} else {
		// Show placeholder
		preview = "Preview will appear here... Type text below to see it rendered."
	}

	// Wrap raw preview in a styled box WITHOUT transforming the content itself
	// Pattern from sysc-greet: border provides structure, content stays raw
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#88C0D0")).
		Padding(1).
		Render(preview)
}

// renderBitTextInput renders the text input field
func (m Model) renderBitTextInput() string {
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#88C0D0")).
		Padding(0, 1).
		Width(m.width - 8)

	if m.bitFocusedControl == 0 {
		inputStyle = inputStyle.
			Background(lipgloss.Color("#88C0D0")).
			Foreground(lipgloss.Color("#2E3440")).
			Bold(true)
	}

	label := "Text: "

	return inputStyle.Render(label + m.bitTextInput.View())
}

// renderBitControls renders all control panels
func (m Model) renderBitControls() string {
	controlsStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3B4252")).
		Padding(1, 2).
		Width(m.width - 8)

	var controls []string

	// Row 1: Font, Alignment, Color
	row1 := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.renderFontControl(),
		m.renderAlignmentControl(),
		m.renderColorControl(),
	)
	controls = append(controls, row1)

	// Row 2: Scale, Shadow, Spacing
	row2 := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.renderScaleControl(),
		m.renderShadowControl(),
		m.renderSpacingControl(),
	)
	controls = append(controls, row2)

	return controlsStyle.Render(lipgloss.JoinVertical(lipgloss.Left, controls...))
}

// renderFontControl renders the font selector
func (m Model) renderFontControl() string {
	focused := m.bitFocusedControl == 1
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#88C0D0")).
		Padding(0, 1).
		Width(20)

	if focused {
		style = style.
			Background(lipgloss.Color("#88C0D0")).
			Foreground(lipgloss.Color("#2E3440")).
			Bold(true)
	}

	fontName := "none"
	if m.bitCurrentFont != nil {
		fontName = m.bitCurrentFont.Name
	}

	label := "Font: "
	value := fmt.Sprintf("%s (%d/%d)", fontName, m.bitSelectedFont+1, len(m.bitFonts))

	return style.Render(label + "\n" + value)
}

// renderAlignmentControl renders alignment buttons
func (m Model) renderAlignmentControl() string {
	focused := m.bitFocusedControl == 2
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#A3BE8C")).
		Padding(0, 1).
		Width(18)

	if focused {
		style = style.
			Background(lipgloss.Color("#A3BE8C")).
			Foreground(lipgloss.Color("#2E3440")).
			Bold(true)
	}

	label := "Align: "

	buttons := []string{}
	alignments := []string{"[L]", "[C]", "[R]"}
	for i, text := range alignments {
		if i == m.bitAlignment {
			buttons = append(buttons, text)
		} else {
			buttons = append(buttons, text)
		}
	}

	return style.Render(label + "\n" + strings.Join(buttons, " "))
}

// renderColorControl renders color selector
func (m Model) renderColorControl() string {
	focused := m.bitFocusedControl == 3
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#D08770")).
		Padding(0, 1).
		Width(20)

	if focused {
		style = style.
			Background(lipgloss.Color("#D08770")).
			Foreground(lipgloss.Color("#2E3440")).
			Bold(true)
	}

	label := "Color: "
	value := "███ " + m.bitColor

	return style.Render(label + "\n" + value)
}

// renderScaleControl renders scale selector
func (m Model) renderScaleControl() string {
	focused := m.bitFocusedControl == 4
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#EBCB8B")).
		Padding(0, 1).
		Width(18)

	if focused {
		style = style.
			Background(lipgloss.Color("#EBCB8B")).
			Foreground(lipgloss.Color("#2E3440")).
			Bold(true)
	}

	label := "Scale: "
	value := fmt.Sprintf("%.1fx", m.bitScale)

	return style.Render(label + "\n" + value)
}

// renderShadowControl renders shadow toggle
func (m Model) renderShadowControl() string {
	focused := m.bitFocusedControl == 5
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#B48EAD")).
		Padding(0, 1).
		Width(20)

	if focused {
		style = style.
			Background(lipgloss.Color("#B48EAD")).
			Foreground(lipgloss.Color("#2E3440")).
			Bold(true)
	}

	label := "Shadow: "

	status := "Off"
	if m.bitShadow {
		status = fmt.Sprintf("On (%d,%d)", m.bitShadowOffsetX, m.bitShadowOffsetY)
	}

	return style.Render(label + "\n" + status)
}

// renderSpacingControl renders spacing controls
func (m Model) renderSpacingControl() string {
	focused := m.bitFocusedControl == 6
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#5E81AC")).
		Padding(0, 1).
		Width(20)

	if focused {
		style = style.
			Background(lipgloss.Color("#5E81AC")).
			Foreground(lipgloss.Color("#ECEFF4")).
			Bold(true)
	}

	label := "Spacing: "
	value := fmt.Sprintf("C:%d W:%d L:%d", m.bitCharSpacing, m.bitWordSpacing, m.bitLineSpacing)

	return style.Render(label + "\n" + value)
}

// renderBitHelp renders help text for BIT editor
func (m Model) renderBitHelp() string {
	helpText := "Tab/Shift+Tab Controls • ←/→ Adjust • Enter Select • Ctrl+F Font • Ctrl+C Color • Ctrl+S Save • Esc Back"
	return m.styles.Help.Render(helpText)
}

// renderFontBrowser renders the font selection browser
func (m Model) renderFontBrowser() string {
	var sections []string

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#88C0D0")).
		Padding(1, 0).
		Render("Select Font")
	sections = append(sections, title)

	// Font list
	listStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#88C0D0")).
		Padding(1, 2).
		Width(m.width - 8).
		Height(m.height - 10)

	var fontItems []string
	startIdx := 0
	endIdx := len(m.bitFonts)

	// Show window of fonts around selection
	windowSize := m.height - 12
	if windowSize < 10 {
		windowSize = 10
	}

	if len(m.bitFonts) > windowSize {
		startIdx = m.bitSelectedFont - windowSize/2
		if startIdx < 0 {
			startIdx = 0
		}
		endIdx = startIdx + windowSize
		if endIdx > len(m.bitFonts) {
			endIdx = len(m.bitFonts)
			startIdx = endIdx - windowSize
			if startIdx < 0 {
				startIdx = 0
			}
		}
	}

	for i := startIdx; i < endIdx; i++ {
		fontName := m.bitFonts[i]
		if i == m.bitSelectedFont {
			fontItems = append(fontItems, lipgloss.NewStyle().
				Foreground(lipgloss.Color("#A3BE8C")).
				Bold(true).
				Render(fmt.Sprintf("▸ %s", fontName)))
		} else {
			fontItems = append(fontItems, lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ECEFF4")).
				Render(fmt.Sprintf("  %s", fontName)))
		}
	}

	sections = append(sections, listStyle.Render(strings.Join(fontItems, "\n")))

	helpText := "↑/↓ Navigate • Enter Select • Esc Cancel"
	sections = append(sections, m.styles.Help.Render(helpText))

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return content
}

// renderColorPicker renders the color picker
func (m Model) renderColorPicker() string {
	var sections []string

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#88C0D0")).
		Padding(1, 0).
		Render("Select Color")
	sections = append(sections, title)

	// Theme colors
	themeColors := []struct {
		name  string
		color string
	}{
		{"Nord Blue", "#88C0D0"},
		{"Nord Green", "#A3BE8C"},
		{"Nord Purple", "#B48EAD"},
		{"Nord Orange", "#D08770"},
		{"Nord Red", "#BF616A"},
		{"Nord Yellow", "#EBCB8B"},
		{"Dracula Purple", "#BD93F9"},
		{"Dracula Pink", "#FF79C6"},
		{"Dracula Cyan", "#8BE9FD"},
		{"Dracula Green", "#50FA7B"},
		{"White", "#FFFFFF"},
		{"Gray", "#808080"},
	}

	listStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#88C0D0")).
		Padding(1, 2).
		Width(m.width - 8)

	var colorItems []string
	for _, c := range themeColors {
		swatch := lipgloss.NewStyle().
			Foreground(lipgloss.Color(c.color)).
			Render("███ ")

		item := swatch + c.name + " " + c.color

		if c.color == m.bitColor {
			colorItems = append(colorItems, lipgloss.NewStyle().
				Foreground(lipgloss.Color("#A3BE8C")).
				Bold(true).
				Render("▸ "+item))
		} else {
			colorItems = append(colorItems, lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ECEFF4")).
				Render("  "+item))
		}
	}

	sections = append(sections, listStyle.Render(strings.Join(colorItems, "\n")))

	helpText := "↑/↓ Navigate • Enter Select • Esc Cancel"
	sections = append(sections, m.styles.Help.Render(helpText))

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return content
}

// renderBitSavePrompt renders the save dialog for BIT editor
func (m Model) renderBitSavePrompt() string {
	var sections []string

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#88C0D0")).
		Padding(1, 0).
		Render("Save Banner Text")
	sections = append(sections, title)

	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#BF616A")).
		Bold(true).
		Padding(1, 0)
	if m.saveError != "" {
		sections = append(sections, warningStyle.Render("⚠ "+m.saveError))
	}

	instructionsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ECEFF4")).
		Padding(1, 0)
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4C566A")).
		Padding(1, 0)

	// A pending collision replaces the filename form with the question.
	if m.confirmOverwrite {
		sections = append(sections, warningStyle.Render("⚠ "+m.overwritePath+" already exists"))
		sections = append(sections, instructionsStyle.Render("Overwrite it with the banner on screen?"))
		sections = append(sections, helpStyle.Render("y Overwrite • any other key Keep existing"))
		return lipgloss.JoinVertical(lipgloss.Left, sections...)
	}

	// The destination follows the selected target, so name it.
	sections = append(sections, instructionsStyle.Render(fmt.Sprintf("Enter filename (will be saved to %s):", exportDirLabel(m.exportTarget))))

	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#88C0D0")).
		Padding(1, 2).
		Width(m.width - 6)
	sections = append(sections, inputStyle.Render(m.filenameInput.View()))

	sections = append(sections, m.styles.Help.Render("Enter Confirm • Esc Cancel"))

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return content
}
