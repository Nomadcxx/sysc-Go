package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		// Enforce minimum terminal dimensions (at least reasonable full screen)
		minWidth := 100
		minHeight := 30

		m.width = msg.Width
		m.height = msg.Height

		// Check if terminal is too small
		if m.width < minWidth || m.height < minHeight {
			// Terminal too small - show warning instead
			m.width = msg.Width
			m.height = msg.Height
			m.canvasHeight = 5
			return m, nil
		}

		// Canvas takes up maximum available screen space
		// Total overhead: Border (2) + Selectors (4) + Guidance (1) + Help (1) = 8 lines
		m.canvasHeight = m.height - 8 // Account for canvas border + UI elements
		if m.canvasHeight < 20 {
			m.canvasHeight = 20 // Minimum viewport height
		}
		// Update textarea size if in editor mode
		if m.editorMode {
			m.textarea.SetWidth(m.width - 10)
			m.textarea.SetHeight(m.height - 10)
		}
		//  Resize animation if running
		if m.animationRunning {
			// Recreate animation with new dimensions (or create first time)
			m.currentAnim = m.createAnimation()
		}
		m = m.ensureFloydViewportSize()
		return m, nil

	case TickMsg:
		// Handle animation tick
		if m.animationRunning && m.currentAnim != nil {
			m.currentAnim.Update()
			m.animFrames++

			// Check duration limit
			duration := m.durations[m.selectedDuration]
			if duration != "infinite" {
				// Parse duration and check if we should stop
				// For now, simplified: stop after reasonable frame count
				// TODO: Parse actual duration string and calculate frames
				maxFrames := 200 // ~10 seconds at 50ms per frame
				switch duration {
				case "5s":
					maxFrames = 100
				case "10s":
					maxFrames = 200
				case "30s":
					maxFrames = 600
				case "60s":
					maxFrames = 1200
				}

				if m.animFrames >= maxFrames {
					m.animationRunning = false
					m.currentAnim = nil
					m.animFrames = 0
					return m, nil
				}
			}

			// Continue ticking
			return m, tickCmd()
		}
		return m, nil

	case floydStreamMsg:
		m = m.handleFloydStream(msg)
		if m.floydStream != nil && !msg.done {
			return m, waitForFloydChunk(m.floydStream)
		}
		return m, nil
	}

	return m, nil
}

// handleKeyPress processes keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle FLOYD console mode first
	if m.floydMode {
		return m.handleFloydKeyPress(msg)
	}

	// Handle BIT editor mode separately
	if m.bitEditorMode {
		return m.handleBitEditorKeyPress(msg)
	}

	// Handle editor mode separately
	if m.editorMode {
		return m.handleEditorKeyPress(msg)
	}

	// Global quit
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	// If animation is running, only allow ESC to stop it
	if m.animationRunning {
		if msg.String() == "esc" {
			m.animationRunning = false
			m.currentAnim = nil
			m.animFrames = 0
			return m, nil
		}
		// Ignore other keys while animation is running
		return m, nil
	}

	// Normal navigation when not running animation
	switch msg.String() {
	case "q", "esc":
		return m, tea.Quit

	case "ctrl+k":
		m = m.prepareFloydMode()
		m.floydMode = true
		return m, nil

	case "ctrl+b":
		// Launch BIT editor
		m.bitEditorMode = true
		m.bitTextInput.Focus()
		m.bitFocusedControl = 0
		return m, nil

	case "up":
		return m.navigateUp(), nil

	case "down":
		return m.navigateDown(), nil

	case "left":
		return m.navigateLeft(), nil

	case "right":
		return m.navigateRight(), nil

	case "enter":
		return m.startAnimation()
	}

	return m, nil
}

// navigateUp moves the selection up within the current selector
func (m Model) navigateUp() Model {
	// Don't allow navigation while animation is running
	if m.animationRunning {
		return m
	}

	switch m.focusedSelector {
	case 0: // Animation selector
		if m.selectedAnimation > 0 {
			m.selectedAnimation--
		}
	case 1: // Theme selector
		if m.selectedTheme > 0 {
			m.selectedTheme--
		}
	case 2: // File selector
		if m.selectedFile > 0 {
			m.selectedFile--
		}
	case 3: // Duration selector
		if m.selectedDuration > 0 {
			m.selectedDuration--
		}
	}
	return m
}

// navigateDown moves the selection down within the current selector
func (m Model) navigateDown() Model {
	// Don't allow navigation while animation is running
	if m.animationRunning {
		return m
	}

	switch m.focusedSelector {
	case 0: // Animation selector
		if m.selectedAnimation < len(m.animations)-1 {
			m.selectedAnimation++
		}
	case 1: // Theme selector
		if m.selectedTheme < len(m.themes)-1 {
			m.selectedTheme++
		}
	case 2: // File selector
		if m.selectedFile < len(m.files)-1 {
			m.selectedFile++
		}
	case 3: // Duration selector
		if m.selectedDuration < len(m.durations)-1 {
			m.selectedDuration++
		}
	}
	return m
}

// navigateLeft moves focus to the previous selector
func (m Model) navigateLeft() Model {
	// Don't allow navigation while animation is running
	if m.animationRunning {
		return m
	}

	if m.focusedSelector > 0 {
		m.focusedSelector--
	}
	return m
}

// navigateRight moves focus to the next selector
func (m Model) navigateRight() Model {
	// Don't allow navigation while animation is running
	if m.animationRunning {
		return m
	}

	if m.focusedSelector < 3 {
		m.focusedSelector++
	}
	return m
}

// startAnimation creates and starts the animation in viewport
func (m Model) startAnimation() (Model, tea.Cmd) {
	// Create animation instance (this may set editor mode instead)
	anim := m.createAnimation()

	// If createAnimation set editor mode, return early
	if m.editorMode || m.bitEditorMode {
		return m, nil
	}

	// If animation was created, start it
	if anim != nil {
		m.currentAnim = anim
		m.animationRunning = true
		m.animFrames = 0
		return m, tickCmd() // Start the tick loop
	}

	// No animation created (shouldn't happen)
	return m, nil
}

// handleEditorKeyPress handles keyboard input in editor mode
func (m Model) handleEditorKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle export prompt separately
	if m.showExportPrompt {
		switch msg.String() {
		case "esc":
			// Cancel export prompt
			m.showExportPrompt = false
			return m, nil

		case "up":
			// Navigate up in export options
			if m.exportTarget > 0 {
				m.exportTarget--
			}
			return m, nil

		case "down":
			// Navigate down in export options
			if m.exportTarget < 1 { // 0=syscgo, 1=sysc-walls
				m.exportTarget++
			}
			return m, nil

		case "enter":
			// Confirm export target - both syscgo and sysc-walls are supported
			m.showExportPrompt = false
			m.showSavePrompt = true
			m.filenameInput.Focus()
			return m, nil
		}
		return m, nil
	}

	// Handle save prompt separately
	if m.showSavePrompt {
		switch msg.String() {
		case "esc":
			// Cancel save prompt
			m.showSavePrompt = false
			m.filenameInput.SetValue("")
			m.filenameInput.Blur()
			return m, nil

		case "enter":
			// Save file
			return m.saveFile()

		default:
			// Update filename input
			m.filenameInput, cmd = m.filenameInput.Update(msg)
			return m, cmd
		}
	}

	// Handle editor input
	switch msg.String() {
	case "esc":
		// Exit editor mode and return to main UI
		m.editorMode = false
		m.textarea.Blur()
		m.textarea.Reset()
		// Refresh file list to include any newly saved files
		files := discoverAssetFiles()
		if len(files) == 0 {
			files = []string{"SYSC.txt"}
		}
		files = append([]string{"BIT Text Editor", "Custom text"}, files...)
		m.files = files
		return m, nil

	case "ctrl+s":
		// Show export/save prompt to choose target
		m.showExportPrompt = true
		m.exportTarget = 0 // Default to syscgo
		return m, nil

	default:
		// Update textarea
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd
	}
}

// saveFile saves the text area content to assets folder
func (m Model) saveFile() (Model, tea.Cmd) {
	// Clear previous error
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

	// Validate content is not empty
	if len(m.textarea.Value()) == 0 {
		m.saveError = "Content cannot be empty"
		return m, nil
	}

	// Save to assets folder
	err := saveToAssets(filename, m.textarea.Value())
	if err != nil {
		m.saveError = err.Error()
		return m, nil
	}

	// Exit editor mode and return to main UI
	m.editorMode = false
	m.showSavePrompt = false
	m.saveError = ""
	m.textarea.Blur()
	m.textarea.Reset()
	m.filenameInput.SetValue("")
	m.filenameInput.Blur()

	// Refresh file list to include the newly saved file
	files := discoverAssetFiles()
	if len(files) == 0 {
		files = []string{"SYSC.txt"}
	}
	files = append([]string{"BIT Text Editor", "Custom text"}, files...)
	m.files = files

	// Set the selected file to the newly created file
	for i, file := range m.files {
		if file == filename {
			m.selectedFile = i
			break
		}
	}

	return m, nil
}
