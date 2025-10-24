package animations

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
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
	lastUpdate      time.Time
	charDelay       time.Duration
	printSpeed      int
	printHeadSymbol string
	trailSymbols    []string
	gradientStops   []string
	complete        bool
}

// PrintConfig holds configuration for the print effect
type PrintConfig struct {
	Width           int
	Height          int
	Text            string
	CharDelay       time.Duration
	PrintSpeed      int
	PrintHeadSymbol string
	TrailSymbols    []string
	GradientStops   []string
}

// NewPrintEffect creates a new print effect with given configuration
func NewPrintEffect(config PrintConfig) *PrintEffect {
	lines := strings.Split(config.Text, "\n")

	// Remove empty trailing lines
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}

	// Set defaults if not provided
	printSpeed := config.PrintSpeed
	if printSpeed <= 0 {
		printSpeed = 1
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

	return &PrintEffect{
		width:           config.Width,
		height:          config.Height,
		text:            config.Text,
		lines:           lines,
		currentLine:     0,
		currentCol:      0,
		revealed:        []string{},
		lastUpdate:      time.Now(),
		charDelay:       config.CharDelay,
		printSpeed:      printSpeed,
		printHeadSymbol: printHeadSymbol,
		trailSymbols:    trailSymbols,
		gradientStops:   gradientStops,
		complete:        false,
	}
}

// Update advances the print effect animation
func (p *PrintEffect) Update() {
	if p.complete {
		return
	}

	currentTime := time.Now()

	// Check if animation is complete
	if p.currentLine >= len(p.lines) {
		p.complete = true
		return
	}

	// Check if enough time has passed to print next character(s)
	if currentTime.Sub(p.lastUpdate) >= p.charDelay {
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

		p.lastUpdate = currentTime
	}
}

// Render converts the print effect to text output
// Render returns the current state of the print effect with colors
func (p *PrintEffect) Render() string {
	// Create a buffer to hold the output
	buffer := make([][]string, p.height)
	for i := range buffer {
		buffer[i] = make([]string, p.width)
		for j := range buffer[i] {
			buffer[i][j] = " "
		}
	}

	// Calculate centered starting position
	startY := (p.height - len(p.lines)) / 2
	if startY < 0 {
		startY = 0
	}

	// Render revealed lines and current line being printed
	for lineIdx := 0; lineIdx < len(p.revealed); lineIdx++ {
		y := startY + lineIdx
		if y >= p.height {
			break
		}

		line := p.revealed[lineIdx]
		startX := (p.width - len(line)) / 2
		if startX < 0 {
			startX = 0
		}

		for charIdx, char := range line {
			x := startX + charIdx
			if x >= p.width {
				break
			}
			
			// Calculate gradient color
			color := p.getGradientColor(float64(charIdx) / float64(len(line)))
			style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
			buffer[y][x] = style.Render(string(char))
		}
	}

	// Render current line being printed
	if p.currentLine < len(p.lines) {
		y := startY + len(p.revealed)
		if y < p.height {
			currentLineText := p.lines[p.currentLine]
			runes := []rune(currentLineText)
			
			startX := (p.width - len(currentLineText)) / 2
			if startX < 0 {
				startX = 0
			}

			// Render revealed portion of current line
			if p.currentCol > 0 {
				revealedText := string(runes[:min(p.currentCol, len(runes))])
				for charIdx, char := range revealedText {
					x := startX + charIdx
					if x >= p.width {
						break
					}
					
					color := p.getGradientColor(float64(charIdx) / float64(len(currentLineText)))
					style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
					buffer[y][x] = style.Render(string(char))
				}

				// Add trail effect
				trailX := startX + p.currentCol
				for i, trailSymbol := range p.trailSymbols {
					x := trailX + i
					if x >= p.width {
						break
					}
					buffer[y][x] = trailSymbol
				}

				// Add print head
				headX := trailX + len(p.trailSymbols)
				if headX < p.width {
					buffer[y][headX] = p.printHeadSymbol
				}
			} else {
				// Just starting - show trail and head at beginning
				x := startX
				if x < p.width && len(p.trailSymbols) > 0 {
					buffer[y][x] = p.trailSymbols[0]
					if x+1 < p.width {
						buffer[y][x+1] = p.printHeadSymbol
					}
				}
			}
		}
	}

	// Convert buffer to string
	var lines []string
	for _, line := range buffer {
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
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}

	p.lines = lines
	p.currentLine = 0
	p.currentCol = 0
	p.revealed = []string{}
	p.lastUpdate = time.Now()
	p.complete = false
}

// IsComplete returns whether the animation is finished
func (p *PrintEffect) IsComplete() bool {
	return p.complete
}
