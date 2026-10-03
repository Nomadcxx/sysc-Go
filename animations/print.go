package animations

import (
	"strings"

	"github.com/charmbracelet/lipgloss/v2"
)

// PrintEffect creates a typewriter/printer effect for text
type PrintEffect struct {
	width           int
	height          int
	text            string
	lines           []string
	currentLine     int
	currentCol      int
	revealed        []string
	frameCounter    int // Frame-based timing instead of time.Duration
	framesPerChar   int // Frames to wait before printing next character
	printSpeed      int
	printHeadSymbol string
	trailSymbols    []string
	gradientStops   []string
	phase           string // "printing", "complete", "holding"
	holdFrameCount  int
	maxLineWidth    int
	auto            bool // Auto-size canvas to fit text
	display         bool // Display mode: complete once and hold
	holdFrames      int  // Frames to hold before looping

	// Pre-allocated buffer for performance
	buffer [][]string
}

// PrintConfig holds configuration for the print effect
type PrintConfig struct {
	Width           int
	Height          int
	Text            string
	FramesPerChar   int // Frames to wait before printing next character (replaces CharDelay)
	PrintSpeed      int // Characters to print per update cycle
	PrintHeadSymbol string
	TrailSymbols    []string
	GradientStops   []string
	Auto            bool // Auto-size canvas to fit text dimensions
	Display         bool // Display mode: complete once and hold (true) or loop (false)
	HoldFrames      int  // Frames to hold completed state before looping (default 100)
}

// calculatePrintTextDimensions calculates the dimensions needed to display text
func calculatePrintTextDimensions(text string) (int, int) {
	lines := strings.Split(text, "\n")
	maxWidth := 0
	for _, line := range lines {
		runes := []rune(line)
		if len(runes) > maxWidth {
			maxWidth = len(runes)
		}
	}
	return maxWidth, len(lines)
}

// NewPrintEffect creates a new print effect with given configuration
func NewPrintEffect(config PrintConfig) *PrintEffect {
	config.Text = normalizeMultilineText(config.Text)
	lines := strings.Split(config.Text, "\n")

	// Don't remove empty lines - they might be part of ASCII art structure!

	// Handle auto-sizing
	width := config.Width
	height := config.Height
	if config.Auto {
		width, height = calculatePrintTextDimensions(config.Text)
	}

	// Set defaults if not provided
	printSpeed := config.PrintSpeed
	if printSpeed <= 0 {
		printSpeed = 1
	}

	framesPerChar := config.FramesPerChar
	if framesPerChar <= 0 {
		framesPerChar = 1 // Print every frame by default
	}

	holdFrames := config.HoldFrames
	if holdFrames <= 0 {
		holdFrames = 100 // Default ~5 seconds at 20fps
	}

	printHeadSymbol := config.PrintHeadSymbol
	if printHeadSymbol == "" {
		printHeadSymbol = "█"
	}

	trailSymbols := config.TrailSymbols
	if len(trailSymbols) == 0 {
		trailSymbols = []string{"░", "▒", "▓"}
	}

	gradientStops := config.GradientStops
	if len(gradientStops) == 0 {
		gradientStops = []string{"#ffffff"}
	}

	// Calculate max line width for proper ASCII art alignment
	maxLineWidth := 0
	for _, line := range lines {
		lineLen := len([]rune(line))
		if lineLen > maxLineWidth {
			maxLineWidth = lineLen
		}
	}

	// Pre-allocate buffer for performance
	buffer := make([][]string, height)
	for i := range buffer {
		buffer[i] = make([]string, width)
	}

	effect := &PrintEffect{
		width:           width,
		height:          height,
		text:            config.Text,
		lines:           lines,
		currentLine:     0,
		currentCol:      0,
		revealed:        []string{},
		frameCounter:    0,
		framesPerChar:   framesPerChar,
		printSpeed:      printSpeed,
		printHeadSymbol: printHeadSymbol,
		trailSymbols:    trailSymbols,
		gradientStops:   gradientStops,
		phase:           "printing",
		holdFrameCount:  0,
		maxLineWidth:    maxLineWidth,
		auto:            config.Auto,
		display:         config.Display,
		holdFrames:      holdFrames,
		buffer:          buffer,
	}

	return effect
}

// Update advances the print effect animation
func (p *PrintEffect) Update() {
	p.frameCounter++

	switch p.phase {
	case "printing":
		p.updatePrintingPhase()
	case "complete":
		p.updateCompletePhase()
	case "holding":
		p.updateHoldingPhase()
	}
}

// updatePrintingPhase handles the main printing animation
func (p *PrintEffect) updatePrintingPhase() {
	// Check if animation is complete
	if p.currentLine >= len(p.lines) {
		p.phase = "complete"
		p.frameCounter = 0
		return
	}

	// Check if enough frames have passed to print next character(s)
	if p.frameCounter >= p.framesPerChar {
		currentLineText := p.lines[p.currentLine]
		runes := []rune(currentLineText)

		// Print multiple characters based on printSpeed
		for i := 0; i < p.printSpeed && p.currentCol < len(runes); i++ {
			p.currentCol++
		}

		// Check if line is complete
		if p.currentCol >= len(runes) {
			p.revealed = append(p.revealed, currentLineText)
			p.currentLine++
			p.currentCol = 0
		}

		p.frameCounter = 0 // Reset frame counter for next character
	}
}

// updateCompletePhase handles transition to holding
func (p *PrintEffect) updateCompletePhase() {
	// Immediately transition to holding phase
	p.phase = "holding"
	p.holdFrameCount = 0
}

// updateHoldingPhase handles the hold state before looping
func (p *PrintEffect) updateHoldingPhase() {
	p.holdFrameCount++

	// In display mode, hold forever
	if p.display {
		return
	}

	// In loop mode, reset after hold period
	if p.holdFrameCount >= p.holdFrames {
		p.Reset()
	}
}

// Render converts the print effect to text output
// Render returns the current state of the print effect with colors
func (p *PrintEffect) Render() string {
	// Clear pre-allocated buffer
	for i := range p.buffer {
		for j := range p.buffer[i] {
			p.buffer[i][j] = " "
		}
	}

	// Calculate centered starting position
	startY := (p.height - len(p.lines)) / 2
	if startY < 0 {
		startY = 0
	}

	// Calculate starting X position based on max line width (centers the entire block)
	baseStartX := (p.width - p.maxLineWidth) / 2
	if baseStartX < 0 {
		baseStartX = 0
	}

	// Render revealed lines and current line being printed
	for lineIdx := 0; lineIdx < len(p.revealed); lineIdx++ {
		y := startY + lineIdx
		if y >= p.height {
			break
		}

		line := p.revealed[lineIdx]
		// All lines start at the same X position for proper ASCII art alignment
		startX := baseStartX

		// Convert to runes to get proper character indices (not byte indices)
		runes := []rune(line)
		for charIdx := 0; charIdx < len(runes); charIdx++ {
			x := startX + charIdx
			if x >= p.width {
				break
			}

			// Calculate gradient color
			color := p.getGradientColor(float64(charIdx) / float64(len(runes)))
			style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
			p.buffer[y][x] = style.Render(string(runes[charIdx]))
		}
	}

	// Render current line being printed
	if p.currentLine < len(p.lines) {
		y := startY + len(p.revealed)
		if y < p.height {
			currentLineText := p.lines[p.currentLine]
			runes := []rune(currentLineText)

			// All lines start at the same X position for proper ASCII art alignment
			startX := baseStartX

			// Render revealed portion of current line
			if p.currentCol > 0 {
				revealedRunes := runes[:min(p.currentCol, len(runes))]
				for charIdx := 0; charIdx < len(revealedRunes); charIdx++ {
					x := startX + charIdx
					if x >= p.width {
						break
					}

					color := p.getGradientColor(float64(charIdx) / float64(len(runes)))
					style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
					p.buffer[y][x] = style.Render(string(revealedRunes[charIdx]))
				}

				// Add trail effect
				trailX := startX + p.currentCol
				for i, trailSymbol := range p.trailSymbols {
					x := trailX + i
					if x >= p.width {
						break
					}
					p.buffer[y][x] = trailSymbol
				}

				// Add print head
				headX := trailX + len(p.trailSymbols)
				if headX < p.width {
					p.buffer[y][headX] = p.printHeadSymbol
				}
			} else {
				// Just starting - show trail and head at beginning
				x := startX
				if x < p.width && len(p.trailSymbols) > 0 {
					p.buffer[y][x] = p.trailSymbols[0]
					if x+1 < p.width {
						p.buffer[y][x+1] = p.printHeadSymbol
					}
				}
			}
		}
	}

	// Convert buffer to string
	var lines []string
	for _, line := range p.buffer {
		lines = append(lines, strings.Join(line, ""))
	}

	return strings.Join(lines, "\n")
}

// Helper to get gradient color for position
func (p *PrintEffect) getGradientColor(progress float64) string {
	if len(p.gradientStops) == 0 {
		return "#ffffff"
	}
	if len(p.gradientStops) == 1 {
		return p.gradientStops[0]
	}

	// Map progress to gradient position
	totalStops := len(p.gradientStops)
	segmentSize := 1.0 / float64(totalStops-1)
	segment := int(progress / segmentSize)

	if segment >= totalStops-1 {
		return p.gradientStops[totalStops-1]
	}

	return p.gradientStops[segment]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Reset restarts the print effect animation
func (p *PrintEffect) Reset() {
	lines := strings.Split(p.text, "\n")
	// Don't remove empty lines - they might be part of ASCII art structure

	p.lines = lines
	p.currentLine = 0
	p.currentCol = 0
	p.revealed = []string{}
	p.frameCounter = 0
	p.phase = "printing"
	p.holdFrameCount = 0
}

// SetText updates the displayed text and restarts the print animation.
func (p *PrintEffect) SetText(text string) {
	p.text = normalizeMultilineText(text)
	p.Reset()
}

// Resize updates the effect dimensions and reinitializes
func (p *PrintEffect) Resize(width, height int) {
	p.width = width
	p.height = height

	// Re-allocate buffer for new dimensions
	p.buffer = make([][]string, height)
	for i := range p.buffer {
		p.buffer[i] = make([]string, width)
	}

	// Recalculate max line width for centering
	maxLineWidth := 0
	for _, line := range p.lines {
		lineLen := len([]rune(line))
		if lineLen > maxLineWidth {
			maxLineWidth = lineLen
		}
	}
	p.maxLineWidth = maxLineWidth
}

// IsComplete returns whether the animation is finished
func (p *PrintEffect) IsComplete() bool {
	return p.phase == "holding"
}
