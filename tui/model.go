package tui

import (
	"strings"
	"time"

	"github.com/Nomadcxx/sysc-Go/animations"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the TUI state
type Model struct {
	width  int
	height int

	// Canvas area for animation preview
	canvasHeight int

	// Available options
	animations []string
	themes     []string
	files      []string
	durations  []string

	// Current selections
	selectedAnimation int
	selectedTheme     int
	selectedFile      int
	selectedDuration  int

	// Which selector is focused (0=animation, 1=theme, 2=file, 3=duration)
	focusedSelector int

	// Animation preview state
	animationRunning bool
	currentAnim      animations.Animation
	animFrames       int // Frame counter

	// Editor mode for custom text creation
	editorMode       bool
	textarea         textarea.Model
	filenameInput    textinput.Model
	showSavePrompt   bool
	showExportPrompt bool
	exportTarget     int    // 0=syscgo, 1=sysc-walls
	saveError        string // Error message from save operation
	savingInProgress bool

	// BIT Editor mode for banner text creation
	bitEditorMode     bool
	bitTextInput      textinput.Model
	bitFonts          []string // Available font names
	bitSelectedFont   int      // Currently selected font index
	bitCurrentFont    *BitFont // Loaded font
	bitAlignment      int      // 0=left, 1=center, 2=right
	bitColor          string   // Hex color
	bitScale          float64  // 0.5, 1.0, 2.0, 3.0, 4.0
	bitShadow         bool     // Shadow enabled
	bitShadowOffsetX  int      // Shadow horizontal offset
	bitShadowOffsetY  int      // Shadow vertical offset
	bitShadowStyle    int      // 0=light, 1=medium, 2=dark
	bitCharSpacing    int      // Character spacing (0-10)
	bitWordSpacing    int      // Word spacing (0-20)
	bitLineSpacing    int      // Line spacing (0-10)
	bitUseGradient    bool     // Gradient enabled
	bitGradientColor  string   // Gradient end color (hex)
	bitGradientDir    int      // 0=up-down, 1=down-up, 2=left-right, 3=right-left
	bitPreviewLines   []string // Rendered preview output
	bitFocusedControl int      // Which control has focus
	bitColorPicker    bool     // Color picker open
	bitShowFontList   bool     // Font browser open

	// FLOYD agent console
	floydMode         bool
	floydTools        []FloydTool
	floydSelectedTool int
	floydPromptActive bool
	floydPromptLabel  string
	floydPromptKind   FloydToolKind
	floydInput        textinput.Model
	floydContent      textarea.Model
	floydAwaitContent bool
	floydWritePath    string
	floydStatus       string
	floydDir          string
	floydMasterPlan   string
	floydStack        string
	floydRecentLog    string
	floydViewport     viewport.Model
	floydPinned       bool
	floydStreaming    bool
	floydStream       chan floydStreamMsg
	floydBuilder      strings.Builder

	// Styles
	styles Styles
}

// Styles holds lipgloss styles for the TUI
type Styles struct {
	Canvas          lipgloss.Style
	Selector        lipgloss.Style
	SelectorFocused lipgloss.Style
	SelectorLabel   lipgloss.Style
	SelectorValue   lipgloss.Style
	GuidanceBox     lipgloss.Style
	Help            lipgloss.Style
	Background      lipgloss.Style
}

// NewModel creates a new TUI model with default values
func NewModel() Model {
	// Discover .txt files in assets folder
	files := discoverAssetFiles()
	if len(files) == 0 {
		files = []string{"SYSC.txt"} // fallback
	}
	// Prepend editor options to files list
	files = append([]string{"BIT Text Editor", "Custom text"}, files...)

	// Initialize textarea for editor mode
	ta := textarea.New()
	ta.Placeholder = "Enter your ASCII art here..."
	ta.Focus()
	ta.CharLimit = 10000
	ta.SetWidth(100)
	ta.SetHeight(20)
	ta.ShowLineNumbers = true

	// Initialize filename input
	fi := textinput.New()
	fi.Placeholder = "filename.txt"
	fi.CharLimit = 256
	fi.Width = 40

	// Initialize BIT text input
	bitInput := textinput.New()
	bitInput.Placeholder = "Enter text here..."
	bitInput.CharLimit = 100
	bitInput.Width = 40
	bitInput.Focus()

	// Initialize FLOYD console inputs
	floydInput := textinput.New()
	floydInput.Placeholder = "Enter command or path..."
	floydInput.CharLimit = 2048
	floydInput.Width = 60

	floydContent := textarea.New()
	floydContent.Placeholder = "Enter file content here. Press Ctrl+D when finished."
	floydContent.CharLimit = 20000
	floydContent.SetWidth(80)
	floydContent.SetHeight(10)
	floydContent.ShowLineNumbers = true

	floydViewport := viewport.New(80, 12)
	floydViewport.SetContent("(no output yet)")
	floydViewport.HighPerformanceRendering = true

	// Discover available .bit fonts
	bitFonts := ListAvailableFonts()
	if len(bitFonts) == 0 {
		bitFonts = []string{"block"} // fallback
	}

	// Load default font
	var defaultFont *BitFont
	if len(bitFonts) > 0 {
		fontPath, err := FindFontPath(bitFonts[0])
		if err == nil {
			defaultFont, _ = LoadBitFont(fontPath)
		}
	}

	return Model{
		animations: []string{
			"fire",
			"fire-text",
			"matrix",
			"matrix-art",
			"rain",
			"rain-art",
			"fireworks",
			"pour",
			"print",
			"beams",
			"beam-text",
			"ring-text",
			"blackhole-text",
			"aquarium",
		},
		themes: []string{
			"dracula",
			"gruvbox",
			"nord",
			"tokyo-night",
			"catppuccin",
			"material",
			"solarized",
			"monochrome",
			"transishardjob",
			"rama",
			"eldritch",
			"dark",
			"default",
			"white-bricks",
			"warm-luxury",
		},
		files: files,
		durations: []string{
			"5s",
			"10s",
			"30s",
			"60s",
			"infinite",
		},
		selectedAnimation: 3, // matrix-art
		selectedTheme:     4, // catppuccin
		selectedFile:      2, // FLOYD.txt (likely)
		selectedDuration:  4, // infinite by default
		focusedSelector:   0,
		animationRunning:  true, // Start immediately!
		currentAnim:       nil,
		animFrames:        0,
		editorMode:        false,
		textarea:          ta,
		filenameInput:     fi,
		showSavePrompt:    false,
		showExportPrompt:  false,
		exportTarget:      0,
		saveError:         "",
		savingInProgress:  false,
		// BIT Editor defaults
		bitEditorMode:     false,
		bitTextInput:      bitInput,
		bitFonts:          bitFonts,
		bitSelectedFont:   0,
		bitCurrentFont:    defaultFont,
		bitAlignment:      1, // center
		bitColor:          "#88C0D0",
		bitScale:          1.0,
		bitShadow:         false,
		bitShadowOffsetX:  1,
		bitShadowOffsetY:  1,
		bitShadowStyle:    0,
		bitCharSpacing:    1,
		bitWordSpacing:    2,
		bitLineSpacing:    1,
		bitUseGradient:    false,
		bitGradientColor:  "#FFFFFF",
		bitGradientDir:    0,
		bitPreviewLines:   []string{},
		bitFocusedControl: 0,
		bitColorPicker:    false,
		bitShowFontList:   false,
		floydMode:         false,
		floydTools:        defaultFloydTools(),
		floydSelectedTool: 0,
		floydPromptActive: false,
		floydPromptLabel:  "",
		floydPromptKind:   ToolBash,
		floydInput:        floydInput,
		floydContent:      floydContent,
		floydAwaitContent: false,
		floydWritePath:    "",
		floydStatus:       "Select a tool and press Enter to run it.",
		floydViewport:     floydViewport,
		floydPinned:       true,
		floydStreaming:    false,
		floydStream:       nil,
		styles:            NewStyles(),
	}
}

// NewStyles creates the dark theme styles
func NewStyles() Styles {
	return Styles{
		Canvas: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#88C0D0")).
			Padding(1),

		Selector: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#3B4252")).
			Padding(0, 1).
			MarginRight(2).
			Width(18).
			Align(lipgloss.Center, lipgloss.Top),

		SelectorFocused: lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color("#88C0D0")).
			Padding(0, 1).
			MarginRight(2).
			Width(18).
			Align(lipgloss.Center, lipgloss.Top),

		SelectorLabel: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#88C0D0")).
			Bold(true).
			Align(lipgloss.Center).
			MarginBottom(1),

		SelectorValue: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ECEFF4")).
			Bold(false).
			Align(lipgloss.Center).
			MarginTop(0).
			MarginBottom(1),

		GuidanceBox: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D8DEE9")).
			Align(lipgloss.Center),

		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4C566A")).
			Align(lipgloss.Center),

		Background: lipgloss.NewStyle(),
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// TickMsg is sent when animation should update
type TickMsg time.Time

// tickCmd returns a command that sends a tick message for animation updates
func tickCmd() tea.Cmd {
	return tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
