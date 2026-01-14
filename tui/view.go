package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// View renders the TUI
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// FLOYD agent console takes over entire view
	if m.floydMode {
		return m.renderFloydMode()
	}

	// Check if terminal is too small
	if m.width < 100 || m.height < 30 {
		warning := fmt.Sprintf(
			"Terminal too small!\n\n"+
				"Current: %dx%d\n"+
				"Minimum: 100x30\n\n"+
				"Please resize your terminal to at least full screen.\n"+
				"Press Q to quit.",
			m.width, m.height,
		)
		return warning
	}

	// If in BIT editor mode, render BIT editor view
	if m.bitEditorMode {
		return m.renderBitEditorView()
	}

	// If animation is running, show full screen animation
	if m.animationRunning && m.currentAnim != nil {
		return m.renderCanvas()
	}

	// If in editor mode, render editor view
	if m.editorMode {
		return m.renderEditorView()
	}

	var sections []string

	// Canvas area (viewport for animations)
	sections = append(sections, m.renderCanvas())

	// Selector area
	sections = append(sections, m.renderSelectors())

	// Guidance box (explains current selection)
	sections = append(sections, m.renderGuidance())

	// Help text (no j/k hints)
	sections = append(sections, m.renderHelp())

	// Join all sections
	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Wrap in outer container with centering - pattern from installer
	// This centers the entire layout without distorting ASCII art
	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Top).
		Render(content)
}

// renderCanvas renders the animation preview viewport
func (m Model) renderCanvas() string {
	if m.animationRunning && m.currentAnim != nil {
		// Render actual animation frame (raw content) full screen
		return m.currentAnim.Render()
	} else {
		// Show welcome/instructions
		return m.renderWelcome()
	}
}

// renderWelcome renders the welcome screen
func (m Model) renderWelcome() string {
	// Render ASCII art as raw string to preserve exact spacing
	// Pattern from sysc-greet: keep ASCII in raw backticks to prevent distortion
	welcome := floydHeaderCommand + `

Terminal Animation Library - Interactive TUI

Select animation settings below
Press ENTER to preview animation in viewport
Press ESC to stop preview`

	// Return raw ASCII without any manipulation
	return welcome
}

// renderSelectors renders the selector controls
func (m Model) renderSelectors() string {
	selectors := []string{
		m.renderSelector(0, "Animation", m.animations[m.selectedAnimation]),
		m.renderSelector(1, "Theme", m.themes[m.selectedTheme]),
		m.renderSelector(2, "File", m.files[m.selectedFile]),
		m.renderSelector(3, "Duration", m.durations[m.selectedDuration]),
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, selectors...)
}

// renderSelector renders a single selector
func (m Model) renderSelector(index int, label, value string) string {
	// Check if this is the File selector and current animation doesn't need a file
	isFileSelector := (index == 2)
	animName := m.animations[m.selectedAnimation]
	needsFile := animName == "fire-text" || animName == "matrix-art" || animName == "rain-art" || animName == "pour" ||
		animName == "print" || animName == "beam-text" || animName == "ring-text" || animName == "blackhole-text"

	// Disable file selector for non-text animations
	if isFileSelector && !needsFile {
		value = "(disabled)"
	}

	// Truncate long values
	maxValueLen := 14
	if len(value) > maxValueLen {
		value = value[:maxValueLen-2] + ".."
	}

	// Title style - with border, bold, and focus background
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1).
		Align(lipgloss.Center)

	// When focused: solid border + filled background
	focused := index == m.focusedSelector
	if focused {
		titleStyle = titleStyle.
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#88C0D0")).
			Background(lipgloss.Color("#88C0D0")).
			Foreground(lipgloss.Color("#2E3440"))
	} else {
		// Not focused: just border, no fill
		titleStyle = titleStyle.
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#4C566A")).
			Foreground(lipgloss.Color("#88C0D0"))
	}

	// Value style - simple text
	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ECEFF4")).
		Align(lipgloss.Center)

	if isFileSelector && !needsFile {
		valueStyle = valueStyle.Faint(true)
	}

	// Render title and value separately, then join vertically
	title := titleStyle.Render(label)
	val := valueStyle.Render(value)

	// Outer container - minimal styling, just width constraint
	container := lipgloss.NewStyle().
		Width(20).
		Align(lipgloss.Center, lipgloss.Top)

	content := lipgloss.JoinVertical(lipgloss.Center, title, val)
	return container.Render(content)
}

// renderHelp renders the help text
func (m Model) renderHelp() string {
	var helpText string
	if m.animationRunning {
		helpText = "ESC Stop animation • ↑/↓ Navigate options • ←/→ Change selector"
	} else {
		helpText = "↑/↓ Navigate options • ←/→ Change selector • ENTER Start animation • Ctrl+B BIT Editor • Q Quit"
	}
	return m.styles.Help.Render(helpText)
}

// renderGuidance renders guidance/explainer box for current selection
func (m Model) renderGuidance() string {
	animName := m.animations[m.selectedAnimation]
	fileName := m.files[m.selectedFile]

	// Short one-line descriptions
	var guidance string
	switch animName {
	case "fire":
		guidance = "Fire effect"
	case "fire-text":
		guidance = "Fire with ASCII art"
	case "matrix", "matrix-art":
		guidance = "Matrix rain"
	case "rain", "rain-art":
		guidance = "ASCII rain"
	case "fireworks":
		guidance = "Fireworks"
	case "pour":
		guidance = "Pour effect"
	case "print":
		guidance = "Typewriter"
	case "beams", "beam-text":
		guidance = "Light beams"
	case "ring-text":
		guidance = "3D ring text"
	case "blackhole-text":
		guidance = "Blackhole vortex"
	case "aquarium":
		guidance = "Aquarium"
	default:
		guidance = animName
	}

	// Add file info inline if relevant
	if fileName == "BIT Text Editor" {
		guidance += " • BIT Editor (130 fonts)"
	} else if fileName == "Custom text" {
		guidance += " • Custom text editor"
	} else if fileName != "(disabled)" && fileName != "" {
		// Truncate long filenames
		displayName := fileName
		if len(displayName) > 20 {
			displayName = displayName[:17] + "..."
		}
		guidance += " • " + displayName
	}

	return m.styles.GuidanceBox.Render(guidance)
}

// renderEditorView renders the ASCII text editor
func (m Model) renderEditorView() string {
	if m.showExportPrompt {
		return m.renderExportPrompt()
	}

	if m.showSavePrompt {
		return m.renderSavePrompt()
	}

	var sections []string

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#88C0D0")).
		Padding(1, 0).
		Render("ASCII Text Editor")
	sections = append(sections, title)

	// Textarea
	textareaStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#88C0D0")).
		Padding(1, 2).
		Width(m.width - 6).
		Background(lipgloss.Color("#1E1E2E"))

	sections = append(sections, textareaStyle.Render(m.textarea.View()))

	// Help text
	helpText := "Type your ASCII art • Ctrl+S Save/Export • Esc Cancel"
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4C566A")).
		Padding(1, 0)
	sections = append(sections, helpStyle.Render(helpText))

	// No background wrapping to prevent bleeding
	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return content
}

// renderExportPrompt renders the export target selection dialog
func (m Model) renderExportPrompt() string {
	var sections []string

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#88C0D0")).
		Padding(1, 0).
		Render("Save/Export ASCII Art")
	sections = append(sections, title)

	// Instructions
	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ECEFF4")).
		Padding(1, 0).
		Render("Select where to save:")
	sections = append(sections, instructions)

	// Export options
	exportOptions := []string{
		"syscgo - Save to assets/ folder for animations",
		"sysc-walls - Save as wallpaper for sysc-walls",
	}

	optionStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(lipgloss.Color("#ECEFF4"))

	selectedStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Bold(true).
		Foreground(lipgloss.Color("#88C0D0"))

	var optionsRendered []string
	for i, option := range exportOptions {
		if i == m.exportTarget {
			optionsRendered = append(optionsRendered, selectedStyle.Render("▸ "+option))
		} else {
			optionsRendered = append(optionsRendered, optionStyle.Render("  "+option))
		}
	}

	optionsBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#88C0D0")).
		Padding(1, 2).
		Width(m.width - 6)

	sections = append(sections, optionsBox.Render(lipgloss.JoinVertical(lipgloss.Left, optionsRendered...)))

	// Help text
	helpText := "↑/↓ Select • Enter Confirm • Esc Cancel"
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4C566A")).
		Padding(1, 0)
	sections = append(sections, helpStyle.Render(helpText))

	// No background wrapping to prevent bleeding
	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return content
}

// renderSavePrompt renders the save dialog
func (m Model) renderSavePrompt() string {
	var sections []string

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#88C0D0")).
		Padding(1, 0).
		Render("Save ASCII Art")
	sections = append(sections, title)

	// Error message if any
	if m.saveError != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#BF616A")).
			Bold(true).
			Padding(1, 0)
		sections = append(sections, errorStyle.Render("⚠ "+m.saveError))
	}

	// Instructions
	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ECEFF4")).
		Padding(1, 0).
		Render("Enter filename (will be saved to assets/ folder):")
	sections = append(sections, instructions)

	// Filename input
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#88C0D0")).
		Padding(1, 2).
		Width(m.width - 6).
		Background(lipgloss.Color("#2E3440"))
	sections = append(sections, inputStyle.Render(m.filenameInput.View()))

	// Help text
	helpText := "Enter Confirm • Esc Cancel"
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4C566A")).
		Padding(1, 0)
	sections = append(sections, helpStyle.Render(helpText))

	// No background wrapping to prevent bleeding
	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return content
}
