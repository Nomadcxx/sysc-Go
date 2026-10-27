package tui

import (
	"fmt"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

const (
	minTerminalWidth  = 100
	minTerminalHeight = 30

	// mainViewChromeRows is the vertical budget outside the canvas content
	// box: rounded border (2) + selectors (4) + guidance (1) + help (1).
	// Lipgloss Height sets the content box and draws the border outside it,
	// so the canvas content height is terminal height minus this budget.
	mainViewChromeRows = 8
)

type viewRenderCache struct {
	mu sync.Mutex

	borderWidth  int
	borderTheme  int
	topBorder    string
	bottomBorder string
	leftBorder   string
	rightBorder  string

	chromeKey   runningChromeKey
	chromeValid bool
	selectors   string
	guidance    string
	help        string
}

type runningChromeKey struct {
	width, height int
	animation     int
	theme         int
	file          int
	duration      int
	focus         int
	bitFontCount  int
}

type canvasBorder struct {
	top, bottom string
	left, right string
}

var viewSpaceBlock = strings.Repeat(" ", 256)

// terminalTooSmall reports whether a sized terminal is below the TUI minimum.
// Width 0 means the first window-size message has not arrived; View shows "Loading...".
func (m Model) terminalTooSmall() bool {
	if m.width == 0 {
		return false
	}
	return m.width < minTerminalWidth || m.height < minTerminalHeight
}

// View renders the TUI
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Check if terminal is too small
	if m.terminalTooSmall() {
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

	// If in editor mode, render editor view
	if m.editorMode {
		return m.renderEditorView()
	}
	if m.animationRunning && m.currentAnim != nil {
		return m.renderRunningView()
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
		return m.renderAnimationCanvas(m.currentAnim.Render())
	}

	var content string
	if m.animationError != "" {
		content = "Cannot start animation:\n" + ansi.Strip(m.animationError)
	} else {
		// Show welcome/instructions
		content = m.renderWelcome()
	}

	// Content box leaves room for the border and the rows stacked under the
	// canvas. Height pads short frames and does not clip, so an 80% box plus
	// that chrome is taller than a 30-row terminal and pushes help off screen.
	viewportHeight := m.height - mainViewChromeRows
	if viewportHeight < 1 {
		viewportHeight = 1
	}
	viewportWidth := m.width - 10 // Leave some margin

	// Wrap raw content in a styled box WITHOUT transforming the content itself
	// Pattern from installer/sysc-greet: only border, NO padding/align on ASCII
	// NO Align() here - that distorts ASCII. Centering happens at outer container.
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#88C0D0")).
		Width(viewportWidth).
		Height(viewportHeight).
		Render(content)
}

func (m Model) renderAnimationCanvas(frame string) string {
	var builder strings.Builder
	viewportWidth := max(1, m.width-10)
	viewportHeight := max(1, m.height-mainViewChromeRows)
	builder.Grow(len(frame) + (viewportWidth+2)*(viewportHeight+2))
	rows := 0
	m.writeAnimationCanvas(&builder, frame, 0, 0, &rows)
	return builder.String()
}

func (m Model) renderRunningView() string {
	frame := m.currentAnim.Render()
	selectors, guidance, help := m.cachedRunningChrome()

	contentWidth := max(m.width-8, lipgloss.Width(selectors), lipgloss.Width(guidance), lipgloss.Width(help))
	contentLeft := max(0, (m.width-contentWidth)/2)
	contentRight := max(0, m.width-contentWidth-contentLeft)
	var builder strings.Builder
	builder.Grow(len(frame) + m.width*m.height)
	rows := 0
	m.writeAnimationCanvas(&builder, frame, contentLeft, contentRight+contentWidth-(m.width-8), &rows)
	appendRunningSection(&builder, selectors, contentWidth, contentLeft, contentRight, m.height, &rows)
	appendRunningSection(&builder, guidance, contentWidth, contentLeft, contentRight, m.height, &rows)
	appendRunningSection(&builder, help, contentWidth, contentLeft, contentRight, m.height, &rows)
	for rows < m.height {
		if rows > 0 {
			builder.WriteByte('\n')
		}
		writeViewSpaces(&builder, m.width)
		rows++
	}
	return builder.String()
}

func (m Model) writeAnimationCanvas(dst *strings.Builder, frame string, leftPadding, rightPadding int, rows *int) {
	viewportHeight := max(1, m.height-mainViewChromeRows)
	viewportWidth := max(1, m.width-10)
	border := m.cachedCanvasBorder(viewportWidth)
	lines := strings.Split(frame, "\n")

	if *rows > 0 {
		dst.WriteByte('\n')
	}
	writeViewSpaces(dst, leftPadding)
	dst.WriteString(border.top)
	writeViewSpaces(dst, rightPadding)
	(*rows)++
	for y := 0; y < viewportHeight; y++ {
		dst.WriteByte('\n')
		writeViewSpaces(dst, leftPadding)
		dst.WriteString(border.left)
		line := ""
		if y < len(lines) {
			line = lines[y]
		}
		lineWidth := ansi.StringWidth(line)
		if lineWidth > viewportWidth {
			line = ansi.Truncate(line, viewportWidth, "")
			lineWidth = viewportWidth
		}
		dst.WriteString(line)
		writeViewSpaces(dst, viewportWidth-lineWidth)
		dst.WriteString("\033[0m")
		dst.WriteString(border.right)
		writeViewSpaces(dst, rightPadding)
		(*rows)++
	}
	dst.WriteByte('\n')
	writeViewSpaces(dst, leftPadding)
	dst.WriteString(border.bottom)
	writeViewSpaces(dst, rightPadding)
	(*rows)++
}

func (m Model) cachedCanvasBorder(width int) canvasBorder {
	if m.viewCache == nil {
		return newCanvasBorder(width)
	}

	cache := m.viewCache
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if cache.borderWidth != width || cache.borderTheme != m.selectedTheme {
		border := newCanvasBorder(width)
		cache.borderWidth = width
		cache.borderTheme = m.selectedTheme
		cache.topBorder = border.top
		cache.bottomBorder = border.bottom
		cache.leftBorder = border.left
		cache.rightBorder = border.right
	}
	return canvasBorder{top: cache.topBorder, bottom: cache.bottomBorder, left: cache.leftBorder, right: cache.rightBorder}
}

func newCanvasBorder(width int) canvasBorder {
	border := lipgloss.RoundedBorder()
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#88C0D0"))
	return canvasBorder{
		top:    style.Render(border.TopLeft + strings.Repeat(border.Top, width) + border.TopRight),
		bottom: style.Render(border.BottomLeft + strings.Repeat(border.Bottom, width) + border.BottomRight),
		left:   style.Render(border.Left),
		right:  style.Render(border.Right),
	}
}

func (m Model) cachedRunningChrome() (selectors, guidance, help string) {
	if m.viewCache == nil {
		return m.renderSelectors(), m.renderGuidance(), m.renderHelp()
	}

	key := runningChromeKey{
		width: m.width, height: m.height,
		animation: m.selectedAnimation, theme: m.selectedTheme,
		file: m.selectedFile, duration: m.selectedDuration,
		focus: m.focusedSelector, bitFontCount: len(m.bitFonts),
	}
	cache := m.viewCache
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if !cache.chromeValid || cache.chromeKey != key {
		cache.chromeKey = key
		cache.chromeValid = true
		cache.selectors = m.renderSelectors()
		cache.guidance = m.renderGuidance()
		cache.help = m.renderHelp()
	}
	return cache.selectors, cache.guidance, cache.help
}

func appendRunningSection(dst *strings.Builder, section string, contentWidth, contentLeft, contentRight, maxRows int, rows *int) {
	for start := 0; start <= len(section); {
		if *rows >= maxRows {
			return
		}
		end := strings.IndexByte(section[start:], '\n')
		line := section[start:]
		if end >= 0 {
			line = section[start : start+end]
		}
		lineWidth := ansi.StringWidth(line)
		if lineWidth > contentWidth {
			line = ansi.Truncate(line, contentWidth, "")
			lineWidth = contentWidth
		}
		if *rows > 0 {
			dst.WriteByte('\n')
		}
		writeViewSpaces(dst, contentLeft)
		dst.WriteString(line)
		writeViewSpaces(dst, contentWidth-lineWidth)
		writeViewSpaces(dst, contentRight)
		(*rows)++
		if end < 0 {
			return
		}
		start += end + 1
	}
}

func writeViewSpaces(dst *strings.Builder, count int) {
	for count >= len(viewSpaceBlock) {
		dst.WriteString(viewSpaceBlock)
		count -= len(viewSpaceBlock)
	}
	dst.WriteString(viewSpaceBlock[:count])
}

// renderWelcome renders the welcome screen
func (m Model) renderWelcome() string {
	// Render ASCII art as raw string to preserve exact spacing
	// Pattern from sysc-greet: keep ASCII in raw backticks to prevent distortion
	welcome := `▄▀▀▀▀ █   █ ▄▀▀▀▀ ▄▀▀▀▀       ▄▀▀▀▀ ▄▀▀▀▄    ▄▀    ▄▀
 ▀▀▀▄ ▀▀▀▀█  ▀▀▀▄ █     ▀▀▀▀▀ █ ▀▀█ █   █  ▄▀    ▄▀
▀▀▀▀  ▀▀▀▀▀ ▀▀▀▀   ▀▀▀▀        ▀▀▀   ▀▀▀  ▀     ▀
             /// SEE YOU SPACE COWBOY//

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
	needsFile := animationNeedsFile(animName)

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
		// Arrow keys are ignored until Esc stops the preview.
		helpText = "ESC Stop animation"
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
	case "blackhole":
		guidance = "Blackhole vortex"
	case "decrypt":
		guidance = "Text decryption"
	case "cracktro":
		guidance = "Demoscene cracktro"
	case "aquarium":
		guidance = "Aquarium"
	default:
		guidance = animName
	}

	// Add file info inline, but only while the File selector is live. Otherwise
	// the box would advertise an editor the selector shows as (disabled).
	if animationNeedsFile(animName) {
		if fileName == "BIT Text Editor" {
			guidance += fmt.Sprintf(" • BIT Editor (%d fonts)", len(m.bitFonts))
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
		sections = append(sections, instructionsStyle.Render("Overwrite it with the file on screen?"))
		sections = append(sections, helpStyle.Render("y Overwrite • any other key Keep existing"))
		return lipgloss.JoinVertical(lipgloss.Left, sections...)
	}

	// Instructions.  The destination depends on the selected target, so name
	// it rather than claiming everything lands in assets/.
	sections = append(sections, instructionsStyle.Render(fmt.Sprintf("Enter filename (will be saved to %s):", exportDirLabel(m.exportTarget))))

	// Filename input
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#88C0D0")).
		Padding(1, 2).
		Width(m.width - 6).
		Background(lipgloss.Color("#2E3440"))
	sections = append(sections, inputStyle.Render(m.filenameInput.View()))

	// Help text
	sections = append(sections, helpStyle.Render("Enter Confirm • Esc Cancel"))

	// No background wrapping to prevent bleeding
	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return content
}
