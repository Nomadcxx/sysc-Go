package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// handleBitEditorKeyPress handles keyboard input in BIT editor mode
func (m Model) handleBitEditorKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle font browser
	if m.bitShowFontList {
		return m.handleFontBrowserKeyPress(msg)
	}

	// Handle color picker
	if m.bitColorPicker {
		return m.handleColorPickerKeyPress(msg)
	}

	// Handle export prompt
	if m.showExportPrompt {
		return m.handleBitExportPromptKeyPress(msg)
	}

	// Handle save prompt
	if m.showSavePrompt {
		return m.handleBitSavePromptKeyPress(msg)
	}

	// Main BIT editor keys
	switch msg.String() {
	case "esc":
		// Exit BIT editor mode
		m.bitEditorMode = false
		m.bitTextInput.Blur()
		return m, nil

	case "ctrl+s":
		// Show export prompt first
		m.showExportPrompt = true
		m.exportTarget = 0 // Default to syscgo
		return m, nil

	case "ctrl+f":
		// Open font browser
		m.bitShowFontList = true
		return m, nil

	case "ctrl+c":
		// Open color picker
		m.bitColorPicker = true
		return m, nil

	case "tab":
		// Next control
		m.bitFocusedControl++
		if m.bitFocusedControl > 6 {
			m.bitFocusedControl = 0
		}
		// Update input focus
		if m.bitFocusedControl == 0 {
			m.bitTextInput.Focus()
		} else {
			m.bitTextInput.Blur()
		}
		return m, nil

	case "shift+tab":
		// Previous control
		m.bitFocusedControl--
		if m.bitFocusedControl < 0 {
			m.bitFocusedControl = 6
		}
		// Update input focus
		if m.bitFocusedControl == 0 {
			m.bitTextInput.Focus()
		} else {
			m.bitTextInput.Blur()
		}
		return m, nil

	case "enter":
		// Handle control-specific actions
		switch m.bitFocusedControl {
		case 1: // Font
			m.bitShowFontList = true
		case 3: // Color
			m.bitColorPicker = true
		case 5: // Shadow - toggle; arrows adjust the offset
			m.bitShadow = !m.bitShadow
			m = m.updateBitPreview()
		}
		return m, nil

	case "left":
		return m.handleBitControlLeft(), nil

	case "right":
		return m.handleBitControlRight(), nil

	case "up":
		return m.handleBitControlUp(), nil

	case "down":
		return m.handleBitControlDown(), nil

	default:
		// h/j/k/l are navigation aliases on every control except the text field.
		// While Text is focused they must be inserted; otherwise banner words
		// such as "Hello" lose those letters.
		key := msg.String()
		if m.bitFocusedControl != 0 {
			switch key {
			case "h":
				return m.handleBitControlLeft(), nil
			case "l":
				return m.handleBitControlRight(), nil
			case "k":
				return m.handleBitControlUp(), nil
			case "j":
				return m.handleBitControlDown(), nil
			}
		}

		// Auto-focus text input when typing (excluding single-char special keys)
		// This provides better UX - user can just start typing without focusing first
		isTyping := len(key) == 1 || key == "space" || key == "backspace" || key == "delete"

		if isTyping {
			// Auto-focus text input
			m.bitFocusedControl = 0
			m.bitTextInput.Focus()
			m.bitTextInput, cmd = m.bitTextInput.Update(msg)
			m = m.updateBitPreview()
			return m, cmd
		}

		// Update text input if already focused
		if m.bitFocusedControl == 0 {
			m.bitTextInput, cmd = m.bitTextInput.Update(msg)
			m = m.updateBitPreview()
			return m, cmd
		}
	}

	return m, nil
}

// handleFontBrowserKeyPress handles font browser navigation
func (m Model) handleFontBrowserKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.bitShowFontList = false
		return m, nil

	case "up", "k":
		if m.bitSelectedFont > 0 {
			m.bitSelectedFont--
		}
		return m, nil

	case "down", "j":
		if m.bitSelectedFont < len(m.bitFonts)-1 {
			m.bitSelectedFont++
		}
		return m, nil

	case "enter":
		// Load selected font
		if m.bitSelectedFont < len(m.bitFonts) {
			fontPath, err := FindFontPath(m.bitFonts[m.bitSelectedFont])
			if err == nil {
				font, err := LoadBitFont(fontPath)
				if err == nil {
					m.bitCurrentFont = font
					m = m.updateBitPreview()
				}
			}
		}
		m.bitShowFontList = false
		return m, nil
	}

	return m, nil
}

// handleColorPickerKeyPress handles color picker navigation
func (m Model) handleColorPickerKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	themeColors := []string{
		"#88C0D0", "#A3BE8C", "#B48EAD", "#D08770", "#BF616A", "#EBCB8B",
		"#BD93F9", "#FF79C6", "#8BE9FD", "#50FA7B", "#FFFFFF", "#808080",
	}

	// Find current color index
	currentIdx := 0
	for i, c := range themeColors {
		if c == m.bitColor {
			currentIdx = i
			break
		}
	}

	switch msg.String() {
	case "esc":
		m.bitColorPicker = false
		return m, nil

	case "up", "k":
		if currentIdx > 0 {
			m.bitColor = themeColors[currentIdx-1]
			m = m.updateBitPreview()
		}
		return m, nil

	case "down", "j":
		if currentIdx < len(themeColors)-1 {
			m.bitColor = themeColors[currentIdx+1]
			m = m.updateBitPreview()
		}
		return m, nil

	case "enter":
		m.bitColorPicker = false
		return m, nil
	}

	return m, nil
}

// handleBitExportPromptKeyPress handles export target selection
func (m Model) handleBitExportPromptKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.showExportPrompt = false
		return m, nil

	case "up", "k":
		if m.exportTarget > 0 {
			m.exportTarget--
		}
		return m, nil

	case "down", "j":
		if m.exportTarget < 1 {
			m.exportTarget++
		}
		return m, nil

	case "enter":
		// Move to filename prompt
		m.showExportPrompt = false
		m.showSavePrompt = true
		m.filenameInput.Focus()
		m.filenameInput.SetValue("")
		return m, nil
	}

	return m, nil
}

// handleBitSavePromptKeyPress handles save prompt input
func (m Model) handleBitSavePromptKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc":
		m.showSavePrompt = false
		m.saveError = ""
		m.filenameInput.SetValue("")
		m.filenameInput.Blur()
		return m, nil

	case "enter":
		return m.saveBitArt()

	default:
		m.filenameInput, cmd = m.filenameInput.Update(msg)
		return m, cmd
	}
}

// handleBitControlLeft handles left arrow on focused control
func (m Model) handleBitControlLeft() Model {
	switch m.bitFocusedControl {
	case 1: // Font
		if m.bitSelectedFont > 0 {
			m.bitSelectedFont--
			// Load font
			fontPath, err := FindFontPath(m.bitFonts[m.bitSelectedFont])
			if err == nil {
				font, err := LoadBitFont(fontPath)
				if err == nil {
					m.bitCurrentFont = font
					m = m.updateBitPreview()
				}
			}
		}

	case 2: // Alignment
		if m.bitAlignment > 0 {
			m.bitAlignment--
			m = m.updateBitPreview()
		}

	case 4: // Scale
		scales := []float64{0.5, 1.0, 2.0, 3.0, 4.0}
		for i, s := range scales {
			if s == m.bitScale && i > 0 {
				m.bitScale = scales[i-1]
				m = m.updateBitPreview()
				break
			}
		}

	case 5: // Shadow
		if m.bitShadow {
			if m.bitShadowOffsetX > -5 {
				m.bitShadowOffsetX--
				m = m.updateBitPreview()
			}
		}

	case 6: // Spacing - decrease char spacing
		if m.bitCharSpacing > 0 {
			m.bitCharSpacing--
			m = m.updateBitPreview()
		}
	}

	return m
}

// handleBitControlRight handles right arrow on focused control
func (m Model) handleBitControlRight() Model {
	switch m.bitFocusedControl {
	case 1: // Font
		if m.bitSelectedFont < len(m.bitFonts)-1 {
			m.bitSelectedFont++
			// Load font
			fontPath, err := FindFontPath(m.bitFonts[m.bitSelectedFont])
			if err == nil {
				font, err := LoadBitFont(fontPath)
				if err == nil {
					m.bitCurrentFont = font
					m = m.updateBitPreview()
				}
			}
		}

	case 2: // Alignment
		if m.bitAlignment < 2 {
			m.bitAlignment++
			m = m.updateBitPreview()
		}

	case 4: // Scale
		scales := []float64{0.5, 1.0, 2.0, 3.0, 4.0}
		for i, s := range scales {
			if s == m.bitScale && i < len(scales)-1 {
				m.bitScale = scales[i+1]
				m = m.updateBitPreview()
				break
			}
		}

	case 5: // Shadow
		if m.bitShadow {
			if m.bitShadowOffsetX < 5 {
				m.bitShadowOffsetX++
				m = m.updateBitPreview()
			}
		}

	case 6: // Spacing - increase char spacing
		if m.bitCharSpacing < 10 {
			m.bitCharSpacing++
			m = m.updateBitPreview()
		}
	}

	return m
}

// handleBitControlUp handles up arrow on focused control
func (m Model) handleBitControlUp() Model {
	switch m.bitFocusedControl {
	case 5: // Shadow - decrease Y offset
		if m.bitShadow && m.bitShadowOffsetY > -5 {
			m.bitShadowOffsetY--
			m = m.updateBitPreview()
		}

	case 6: // Spacing - increase word spacing
		if m.bitWordSpacing < 20 {
			m.bitWordSpacing++
			m = m.updateBitPreview()
		}
	}

	return m
}

// handleBitControlDown handles down arrow on focused control
func (m Model) handleBitControlDown() Model {
	switch m.bitFocusedControl {
	case 5: // Shadow - increase Y offset
		if m.bitShadow && m.bitShadowOffsetY < 5 {
			m.bitShadowOffsetY++
			m = m.updateBitPreview()
		}

	case 6: // Spacing - decrease word spacing
		if m.bitWordSpacing > 0 {
			m.bitWordSpacing--
			m = m.updateBitPreview()
		}
	}

	return m
}

// updateBitPreview regenerates the preview with current settings
func (m Model) updateBitPreview() Model {
	text := m.bitTextInput.Value()
	if text == "" || m.bitCurrentFont == nil {
		m.bitPreviewLines = []string{}
		return m
	}

	opts := TUIRenderOptions{
		Font:          m.bitCurrentFont,
		Text:          text,
		Alignment:     m.bitAlignment,
		Color:         m.bitColor,
		Scale:         m.bitScale,
		Shadow:        m.bitShadow,
		ShadowOffsetX: m.bitShadowOffsetX,
		ShadowOffsetY: m.bitShadowOffsetY,
		ShadowStyle:   m.bitShadowStyle,
		CharSpacing:   m.bitCharSpacing,
		WordSpacing:   m.bitWordSpacing,
		LineSpacing:   m.bitLineSpacing,
		UseGradient:   m.bitUseGradient,
		GradientColor: m.bitGradientColor,
		GradientDir:   m.bitGradientDir,
		MaxWidth:      m.width - 10,
	}

	m.bitPreviewLines = RenderBitText(opts)
	return m
}

// saveBitArt saves the rendered text to a file
func (m Model) saveBitArt() (Model, tea.Cmd) {
	m.saveError = ""

	filename := m.filenameInput.Value()
	if filename == "" {
		m.saveError = "Filename cannot be empty"
		return m, nil
	}

	// Add .txt extension if not present
	if len(filename) < 4 || filename[len(filename)-4:] != ".txt" {
		filename += ".txt"
	}

	// Validate content
	if len(m.bitPreviewLines) == 0 {
		m.saveError = "Nothing to save - enter text to generate banner"
		return m, nil
	}

	// Export using selected target
	err := ExportBitArt(filename, m.bitPreviewLines, m.exportTarget)
	if err != nil {
		m.saveError = err.Error()
		return m, nil
	}

	// Exit BIT editor mode
	m.bitEditorMode = false
	m.showSavePrompt = false
	m.saveError = ""
	m.filenameInput.SetValue("")
	m.filenameInput.Blur()
	m.bitTextInput.SetValue("")
	m.bitTextInput.Blur()
	m.bitPreviewLines = []string{}

	// Refresh file list
	files := discoverAssetFiles()
	if len(files) == 0 {
		files = []string{"SYSC.txt"}
	}
	files = append([]string{"BIT Text Editor", "Custom text"}, files...)
	m.files = files

	// Select the newly created file
	for i, file := range m.files {
		if file == filename {
			m.selectedFile = i
			break
		}
	}

	return m, nil
}
