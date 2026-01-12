package agenttui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"
)

// InputMode represents the current input mode
type InputMode int

const (
	ModeInsert InputMode = iota
	ModeNormal
	ModeVisual
)

func (m InputMode) String() string {
	switch m {
	case ModeInsert:
		return "INSERT"
	case ModeNormal:
		return "NORMAL"
	case ModeVisual:
		return "VISUAL"
	default:
		return "UNKNOWN"
	}
}

// InputComponent wraps textarea with vim-style keybindings
type InputComponent struct {
	textarea.Model
	mode         InputMode
	history      []string
	historyIdx   int
	showHelp     bool
	promptText   string
	maxHistory   int
	submitKey    string // "enter" or "ctrl+d"
}

// NewInputComponent creates a new input component
func NewInputComponent() InputComponent {
	ta := textarea.New()
	ta.ShowLineNumbers = true
	ta.Placeholder = "Enter your message..."
	ta.CharLimit = 10000
	ta.SetWidth(80)
	ta.SetHeight(5)
	ta.Focus()

	return InputComponent{
		Model:      ta,
		mode:       ModeInsert,
		history:    make([]string, 0, 100),
		historyIdx: -1,
		maxHistory: 100,
		submitKey:  "enter",
		promptText: "│ You",
	}
}

// SetSize sets the input component size
func (i *InputComponent) SetSize(width, height int) {
	i.SetWidth(width - 4) // Account for borders/padding
	if height > 0 {
		i.SetHeight(height)
	}
}

// SetPromptText sets the prompt text (e.g., "│ You" or "│ FLOYD")
func (i *InputComponent) SetPromptText(text string) {
	i.promptText = text
}

// SetValue sets the input content
func (i *InputComponent) SetValue(val string) {
	i.Model.SetValue(val)
}

// Value returns the current input content
func (i *InputComponent) Value() string {
	return i.Model.Value()
}

// CursorPosition returns the cursor position
func (i *InputComponent) CursorPosition() (int, int) {
	// textarea.Model doesn't expose cursor position directly
	// Return placeholder values
	return 0, 0
}

// Focus sets focus to the input
func (i *InputComponent) Focus() {
	i.Model.Focus()
	if i.mode == ModeNormal {
		i.mode = ModeInsert
	}
}

// Blur removes focus from the input
func (i *InputComponent) Blur() {
	i.Model.Blur()
}

// Focused returns whether the input is focused
func (i *InputComponent) Focused() bool {
	return i.Model.Focused()
}

// AddToHistory adds a line to history
func (i *InputComponent) AddToHistory(line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}

	// Don't add duplicate consecutive entries
	if len(i.history) > 0 && i.history[len(i.history)-1] == line {
		return
	}

	i.history = append(i.history, line)

	// Truncate if at max
	if len(i.history) > i.maxHistory {
		i.history = i.history[1:]
	}

	i.historyIdx = -1
}

// HistoryNavigate navigates through history
func (i *InputComponent) HistoryNavigate(direction int) {
	if len(i.history) == 0 {
		return
	}

	if direction < 0 { // Older
		if i.historyIdx < len(i.history)-1 {
			i.historyIdx++
			i.SetValue(i.history[len(i.history)-1-i.historyIdx])
		}
	} else if direction > 0 { // Newer
		if i.historyIdx > 0 {
			i.historyIdx--
			i.SetValue(i.history[len(i.history)-1-i.historyIdx])
		} else if i.historyIdx == 0 {
			i.historyIdx = -1
			i.SetValue("")
		}
	}

	// Move cursor to end
	i.Model.CursorEnd()
}

// SetMode sets the input mode
func (i *InputComponent) SetMode(mode InputMode) tea.Cmd {
	oldMode := i.mode
	i.mode = mode

	if mode == ModeInsert && oldMode != ModeInsert {
		i.Model.Focus()
	}
	if mode != ModeInsert && oldMode == ModeInsert {
		i.Model.Blur()
	}
	return nil
}

// Mode returns the current mode
func (i *InputComponent) Mode() InputMode {
	return i.mode
}

// Update handles input messages
func (i InputComponent) Update(msg tea.Msg) (InputComponent, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle vim-style keybindings
		switch i.mode {
		case ModeNormal:
			return i.handleNormalMode(msg)
		case ModeInsert:
			return i.handleInsertMode(msg)
		case ModeVisual:
			return i.handleVisualMode(msg)
		}
	}

	// Pass through to textarea
	i.Model, cmd = i.Model.Update(msg)
	return i, cmd
}

// handleNormalMode handles keys in normal mode
func (i InputComponent) handleNormalMode(msg tea.KeyMsg) (InputComponent, tea.Cmd) {
	switch msg.String() {
	case "i", "a":
		// Enter insert mode
		return i, i.SetMode(ModeInsert)

	case ":", "o":
		// Enter insert mode (for command/input mode)
		i.Model.Focus()
		i.mode = ModeInsert
		return i, nil

	case "j", "ctrl+n", "down":
		// Navigate history older
		i.HistoryNavigate(-1)
		return i, nil

	case "k", "ctrl+p", "up":
		// Navigate history newer
		i.HistoryNavigate(1)
		return i, nil

	case "ctrl+j":
		// Next history (older)
		i.HistoryNavigate(-1)
		return i, nil

	case "ctrl+k":
		// Previous history (newer)
		i.HistoryNavigate(1)
		return i, nil

	case "q":
		// Toggle help
		i.showHelp = !i.showHelp
		return i, nil

	case "enter":
		// Submit the current value
		if i.Value() != "" {
			return i, func() tea.Msg {
				return UserSubmitMsg{Content: i.Value()}
			}
		}
		return i, nil

	case "esc":
		// Stay in normal mode
		return i, nil
	}

	return i, nil
}

// handleInsertMode handles keys in insert mode
func (i InputComponent) handleInsertMode(msg tea.KeyMsg) (InputComponent, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Return to normal mode
		i.mode = ModeNormal
		i.Model.Blur()
		return i, nil

	case "ctrl+d":
		// Submit on Ctrl+D (like Git commit message)
		if i.submitKey == "ctrl+d" {
			if i.Value() != "" {
				return i, func() tea.Msg {
					return UserSubmitMsg{Content: i.Value()}
				}
			}
		}
		return i, nil

	case "ctrl+j":
		// Insert newline (default behavior in textarea)

	case "ctrl+k":
		// Delete to end of line or navigate history
		if i.Value() == "" {
			i.HistoryNavigate(1)
			return i, nil
		}

	case "enter":
		if i.submitKey == "enter" && !msg.Alt {
			// Check if we should submit (single line or empty+shift)
			lines := strings.Split(i.Value(), "\n")
			if len(lines) == 1 {
				// Single line - submit
				if i.Value() != "" {
					return i, func() tea.Msg {
						return UserSubmitMsg{Content: i.Value()}
					}
				}
			}
		}
	}

	// Pass through to textarea
	var cmd tea.Cmd
	i.Model, cmd = i.Model.Update(msg)
	return i, cmd
}

// handleVisualMode handles keys in visual mode
func (i InputComponent) handleVisualMode(msg tea.KeyMsg) (InputComponent, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Return to normal mode
		i.mode = ModeNormal
		return i, nil

	case "i", "a":
		// Enter insert mode
		return i, i.SetMode(ModeInsert)
	}

	// Pass through to textarea
	var cmd tea.Cmd
	i.Model, cmd = i.Model.Update(msg)
	return i, cmd
}

// View renders the input component
func (i InputComponent) View() string {
	var sb strings.Builder

	// Mode indicator
	modeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#89dceb")).
		Background(lipgloss.Color("#313244")).
		Bold(true).
		Padding(0, 1)

	sb.WriteString(modeStyle.Render(i.mode.String()))

	// Prompt
	sb.WriteString(" ")
	sb.WriteString(i.promptText)
	sb.WriteString(" > ")

	// Help text
	if i.showHelp {
		helpStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6c7086")).
			Faint(true)
		sb.WriteString(helpStyle.Render(" (esc: normal, i: insert, j/k: history, enter: submit)"))
	}

	sb.WriteString("\n")

	// Textarea
	sb.WriteString(i.Model.View())

	return sb.String()
}
