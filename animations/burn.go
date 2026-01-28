package animations

import (
	"math/rand"
	"strings"
)

// BurnPhase represents the current state of the burn animation
type BurnPhase int

const (
	PhaseIgnition BurnPhase = iota // Sequential ignition
	PhaseBurning                   // Full burning
	PhaseSettle                    // Disappearance
	PhasePause                     // Brief rest before reset
)

// BurnChar represents a character being burned
type BurnChar struct {
	char          rune
	x, y          int
	ignitionFrame int  // When fire reached this char (-1 = not yet)
	burnStage     int  // Current stage in burn sequence
	colorIndex    int  // Current color in gradient (0-8)
	stageFrame    int  // Frames on current stage
	colorFrame    int  // Frames on current color
	active        bool // Still burning vs disappeared
}

// TreeEdge represents an edge in the spanning tree
type TreeEdge struct {
	from, to int     // Character indices
	weight   float64 // For Prim's algorithm
}

// Edge for priority queue (internal use)
type Edge struct {
	from, to int
	weight   float64
}

// SmokeParticle represents rising smoke
type SmokeParticle struct {
	x, y     float64
	char     rune
	velocity float64 // Rise speed (0.5-0.8)
	opacity  float64 // 1.0 to 0.0
	lifetime int     // Frames alive
	maxLife  int     // Frames until full fade
}

// BurnAnimation implements the burn effect
type BurnAnimation struct {
	width, height int
	palette       []string // 12 colors (9 burn + 2 smoke + 1 bg)
	theme         string
	withText      bool
	textContent   string

	// Character management
	chars []BurnChar

	// Spanning tree
	spanningTree []TreeEdge
	burnOrder    []int // Ordered char indices
	nextIgnition int   // Index in burnOrder

	// Smoke
	smokeParticles []SmokeParticle
	smokeSymbols   []rune

	// Phase tracking
	phase      BurnPhase
	frameCount int
	cycleFrame int

	// Sequences
	burnSequence  []rune // For burn-text
	flameSequence []rune // For burn
}

// NewBurnEffect creates a new burn background effect
func NewBurnEffect(width, height int, palette []string, theme string) *BurnAnimation {
	return newBurnAnimation(width, height, palette, theme, false, "")
}

// NewBurnTextEffect creates a burn effect with text destruction
func NewBurnTextEffect(width, height int, palette []string, theme string, text string) *BurnAnimation {
	return newBurnAnimation(width, height, palette, theme, true, text)
}

// newBurnAnimation is the internal constructor
func newBurnAnimation(width, height int, palette []string, theme string, withText bool, text string) *BurnAnimation {
	b := &BurnAnimation{
		width:        width,
		height:       height,
		palette:      palette,
		theme:        theme,
		withText:     withText,
		textContent:  text,
		phase:        PhaseIgnition,
		frameCount:   0,
		cycleFrame:   0,
		nextIgnition: 0,
	}

	// Set sequences
	b.burnSequence = []rune{'\'', '.', '▖', '▙', '█', '▜', '▀', '▝', '.'} // 9 stages
	b.flameSequence = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}      // 8 stages
	b.smokeSymbols = b.getSmokeSymbols()

	// Initialize characters
	b.initializeCharacters()

	// TODO: Generate spanning tree

	return b
}

// getSmokeSymbols returns theme-specific smoke characters
func (b *BurnAnimation) getSmokeSymbols() []rune {
	switch strings.ToLower(b.theme) {
	case "monochrome", "dark":
		return []rune{'▒', '░', '▓'}
	case "dracula", "catppuccin", "rama":
		return []rune{'·', '˙', '°'}
	case "nord", "gruvbox", "tokyo-night", "tokyonight":
		return []rune{'.', '\'', ',', '·'}
	case "material", "solarized":
		return []rune{'░', '·', '.'}
	default:
		return []rune{'·', '.', '░'}
	}
}

// initializeCharacters creates the character array
func (b *BurnAnimation) initializeCharacters() {
	b.chars = []BurnChar{}

	if b.withText {
		b.initializeTextChars()
	} else {
		b.initializeRandomChars()
	}
}

// initializeTextChars parses and positions user text
func (b *BurnAnimation) initializeTextChars() {
	if b.textContent == "" {
		return
	}

	lines := strings.Split(b.textContent, "\n")
	maxWidth := 0
	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}

	// Center text
	textHeight := len(lines)
	offsetY := (b.height - textHeight) / 2
	offsetX := (b.width - maxWidth) / 2

	for y, line := range lines {
		for x, ch := range line {
			if ch != ' ' && ch != '\n' && ch != '\r' {
				b.chars = append(b.chars, BurnChar{
					char:          ch,
					x:             offsetX + x,
					y:             offsetY + y,
					ignitionFrame: -1, // Not yet ignited
					burnStage:     0,
					colorIndex:    0,
					stageFrame:    0,
					colorFrame:    0,
					active:        false,
				})
			}
		}
	}
}

// initializeRandomChars fills canvas with random ASCII
func (b *BurnAnimation) initializeRandomChars() {
	// Printable ASCII characters
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()-=_+[]{}|;:,.<>?/"

	for y := 0; y < b.height; y++ {
		for x := 0; x < b.width; x++ {
			// ~30% fill density for background effect
			if rand.Float64() < 0.3 {
				ch := rune(chars[rand.Intn(len(chars))])
				b.chars = append(b.chars, BurnChar{
					char:          ch,
					x:             x,
					y:             y,
					ignitionFrame: -1,
					burnStage:     0,
					colorIndex:    0,
					stageFrame:    0,
					colorFrame:    0,
					active:        false,
				})
			}
		}
	}
}

// Update advances the animation by one frame
func (b *BurnAnimation) Update() {
	b.frameCount++
	b.cycleFrame++

	// TODO: Phase-based updates
	// TODO: Update smoke
}

// Render returns the current frame as a string
func (b *BurnAnimation) Render() string {
	// TODO: Render burn and smoke
	return ""
}

// Reset restarts the animation
func (b *BurnAnimation) Reset() {
	b.phase = PhaseIgnition
	b.frameCount = 0
	b.cycleFrame = 0
	b.nextIgnition = 0
	// TODO: Regenerate characters and spanning tree
}
