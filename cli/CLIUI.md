# 

To achieve a CLI experience like `gemini-cli` or `grok-cli` (where you have a chat history that scrolls) while keeping that massive ASCII header **persistently locked at the top**, we need to introduce a **Viewport**.

A Viewport allows the chat history to scroll up and down *independently* of the screen. This ensures the header never moves, even if you type 100 pages of text.

Here is the upgraded "FLOYD Agent" code using Bubble Tea.

### Prerequisites
You will need the viewport and textinput packages now:
```bash
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/charmbracelet/bubbles/viewport
go get github.com/charmbracelet/bubbles/textinput
```

### The Code (`main.go`)

```go
package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

// 1. CONFIG & ASSETS
// ------------------
const (
	pinkColor  = "#FF007F" // Hot Pink
	whiteColor = "#FFFFFF"
	blackColor = "#000000"
	darkPink   = "#C7006E" // Slightly darker for borders/dividers
)

const prismArt = `
    __/\\\\\\\\\\\\\\\__/\\\___________________/\\\\\_______/\\\________/\\\__/\\\\\\\\\\\\_____________________/\\\\\\\\\_______/\\\\\_______/\\\\\\\\\\\\_____/\\\\\\\\\\\\\\\__________________/\\\\\\\\\__/\\\______________/\\\\\\\\\\\_        
     _\/\\\///////////__\/\\\_________________/\\\///\\\____\///\\\____/\\\/__\/\\\////////\\\________________/\\\////////______/\\\///\\\____\/\\\////////\\\__\/\\\///////////________________/\\\////////__\/\\\_____________\/////\\\///__       
      _\/\\\_____________\/\\\_______________/\\\/__\///\\\____\///\\\/\\\/____\/\\\______\//\\\_____________/\\\/_____________/\\\/__\///\\\__\/\\\______\//\\\_\/\\\_________________________/\\\/___________\/\\\_________________\/\\\_____      
       _\/\\\\\\\\\\\_____\/\\\______________/\\\______\//\\\_____\///\\\/______\/\\\_______\/\\\____________/\\\______________/\\\______\//\\\_\/\\\_______\/\\\_\/\\\\\\\\\\\________________/\\\_____________\/\\\_________________\/\\\_____     
        _\/\\\///////______\/\\\_____________\/\\\_______\/\\\_______\/\\\_______\/\\\_______\/\\\___________\/\\\_____________\/\\\_______\/\\\_\/\\\_______\/\\\_\/\\\///////________________\/\\\_____________\/\\\_________________\/\\\_____    
         _\/\\\_____________\/\\\_____________\//\\\______/\\\________\/\\\_______\/\\\_______\/\\\___________\//\\\____________\//\\\______/\\\__\/\\\_______\/\\\_\/\\\_______________________\//\\\____________\/\\\_________________\/\\\_____   
          _\/\\\_____________\/\\\______________\///\\\__/\\\__________\/\\\_______\/\\\_______/\\\_____________\///\\\___________\///\\\__/\\\____\/\\\_______/\\\__\/\\\________________________\///\\\__________\/\\\_________________\/\\\_____  
           _\/\\\_____________\/\\\\\\\\\\\\\\\____\///\\\\\/___________\/\\\_______\/\\\\\\\\\\\\/________________\////\\\\\\\\\____\///\\\\\/_____\/\\\\\\\\\\\\/___\/\\\\\\\\\\\\\\\______________\////\\\\\\\\\_\/\\\\\\\\\\\\\\\__/\\\\\\\\\\\_ 
            _\///______________\///////////////_______\/////_____________\///________\////////////_____________________\/////////_______\/////_______\////////////_____\///////////////__________________\/////////__\///////////////__\///////////__ 
`

// Styles
var (
	// The main pink background container
	baseStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(pinkColor)).
			Foreground(lipgloss.Color(whiteColor)).
			Padding(0)

	// Header Style: Black text for the prism art
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(blackColor)).
			Background(lipgloss.Color(pinkColor)).
			Bold(true).
			Align(lipgloss.Center)

	// User Message Style
	userMsgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(whiteColor)).
			Background(lipgloss.Color(darkPink)).
			Padding(0, 1).
			MarginBottom(1)

	// Agent/Floyd Message Style
	floydMsgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(blackColor)). // Black text on Pink bg
			Background(lipgloss.Color(pinkColor)).
			Padding(0, 1).
			MarginBottom(1)

	// Input Box Style
	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(blackColor)).
			Background(lipgloss.Color(whiteColor)). // White input box to stand out
			Padding(0, 1)
)

// 2. MODEL
// --------
type model struct {
	viewport  viewport.Model
	input     textinput.Model
	messages  []string // Chat history
	ready     bool     // Tracks if layout is calculated
}

func initialModel() model {
	// Setup Input
	ti := textinput.New()
	ti.Placeholder = "Ask Floyd..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50

	// Setup Viewport (Scrolling area)
	vp := viewport.New(0, 0)

	return model{
		viewport: vp,
		input:    ti,
		messages: []string{
			"FLOYD: Tuning in to the cosmic transmissions...",
			"FLOYD: The system is online. Pink side initialized.",
		},
		ready:    false,
	}
}

// 3. INIT
// -------
func (m model) Init() tea.Cmd {
	return textinput.Blink
}

// 4. UPDATE
// ---------
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	// Handle Window Resizing
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(prismArt)
		inputHeight := 3 // Approx height for input area
		
		// Calculate how much space is left for the chat
		verticalMargin := headerHeight + inputHeight
		
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMargin)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMargin
		}
		m.input.Width = msg.Width - 4 // Leave room for padding
	}

	// Handle Key Presses
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			// 1. Add User Message
			userMsg := fmt.Sprintf("USER: %s", m.input.Value())
			m.messages = append(m.messages, userMsg)

			// 2. Generate Agent Response
			agentMsg := fmt.Sprintf("FLOYD: I heard you say [%s].", m.input.Value())
			m.messages = append(m.messages, agentMsg)

			// 3. Reset Input
			m.input.Reset()
			
			// 4. Scroll to bottom of viewport
			m.viewport.GotoBottom()

		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	}

	// Update Input and Viewport
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// 5. VIEW
// -------
func (m model) View() string {
	if !m.ready {
		return "\n  Initializing Floyd System..."
	}

	// -- RENDER THE HEADER --
	// We use the exact ASCII art provided
	header := headerStyle.Render(prismArt)

	// -- RENDER THE CHAT HISTORY --
	// We iterate through messages and style them
	var chatHistory strings.Builder
	for _, msg := range m.messages {
		// Simple logic to style User vs Floyd
		if strings.HasPrefix(msg, "USER:") {
			chatHistory.WriteString(userMsgStyle.Render(msg) + "\n")
		} else {
			chatHistory.WriteString(floydMsgStyle.Render(msg) + "\n")
		}
	}
	
	// Put the styled text into the viewport
	m.viewport.SetContent(chatHistory.String())

	// -- RENDER THE INPUT --
	// A simple prompt line at the bottom
	inputLine := fmt.Sprintf("> %s", m.input.View())
	styledInput := inputStyle.Width(m.viewport.Width).Render(inputLine)

	// -- ASSEMBLE LAYOUT --
	// Stack them: Header -> Chat (Viewport) -> Input
	ui := lipgloss.JoinVertical(lipgloss.Left, 
		header, 
		m.viewport.View(), 
		"", // spacer
		styledInput,
	)

	// Wrap in full pink background
	return baseStyle.Render(ui)
}

// 6. MAIN
// -------
func main() {
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(), // Use alternate screen buffer for TUI
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
	}
}
```

### Key Changes from a basic CLI:

1. **`viewport` Package**: This creates a "window" inside the terminal. The text inside this window scrolls, but the frame of the window (and the Header above it) stays perfectly still.
2. **Layout Calculation**:
   * The code calculates `Header Height` (size of your ASCII art) and `Input Height` (space for typing).
   * It subtracts these from the total screen size to determine exactly how big the chat area should be.
3. **Auto-Scrolling**:
   * `m.viewport.GotoBottom()` is called automatically when you hit Enter. This ensures you always see the newest message, just like `gemini-cli` or `ChatGPT`.
4. **Styling**:
   * **User Messages**: Dark Pink background box.
   * **Floyd Messages**: Transparent (shows the main pink background) but with Black text for high readability.
   * **Input**: A white bar at the bottom to clearly distinguish where you type.

### How to run it:
1. Save the code as `main.go`.
2. Run `go run main.go`.
3. Type a message and hit Enter. The header will remain locked at the top while your conversation flows below it.