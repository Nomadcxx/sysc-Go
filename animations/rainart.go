package animations

import (
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss/v2"
)

// RainArtEffect implements rain animation that gradually forms ASCII art
type RainArtEffect struct {
	width    int
	height   int
	palette  []string
	chars    []rune // Rain characters
	drops    []RainDrop
	maxDrops int

	// ASCII art formation
	text         string
	artPositions map[int]map[int]rune // [y][x] = character
	frozenChars  map[int]map[int]*FrozenChar
	centerX      int
	centerY      int
	artWidth     int
	artHeight    int
	rng          *rand.Rand
	freezeChance float64 // Probability a drop freezes when passing art position
}

// FrozenChar represents a rain character that has frozen to form the art
type FrozenChar struct {
	char  rune
	color string
}

// NewRainArtEffect creates a new rain-art effect
func NewRainArtEffect(width, height int, palette []string, text string) *RainArtEffect {
	r := &RainArtEffect{
		width:        width,
		height:       height,
		palette:      palette,
		chars:        []rune{'|', '⋮', '║', '¦', '┆', '┊', '╎', '╏', '▏', '▎', '▍', '▌', '▋', '▊', '▉'},
		drops:        make([]RainDrop, 0, 200),
		maxDrops:     width * 4, // 4x width for very dense rain
		text:         text,
		artPositions: make(map[int]map[int]rune),
		frozenChars:  make(map[int]map[int]*FrozenChar),
		rng:          rand.New(rand.NewSource(time.Now().UnixNano())),
		freezeChance: 0.90, // 90% chance to freeze when passing through art position (very fast crystallization)
	}

	r.parseArt()
	r.init()
	return r
}

// parseArt extracts ASCII art character positions
func (r *RainArtEffect) parseArt() {
	lines := strings.Split(r.text, "\n")
	r.artHeight = len(lines)

	// Find max line width
	r.artWidth = 0
	for _, line := range lines {
		if len([]rune(line)) > r.artWidth {
			r.artWidth = len([]rune(line))
		}
	}

	// Center the art
	r.centerX = (r.width - r.artWidth) / 2
	r.centerY = (r.height - r.artHeight) / 2

	// Parse character positions
	for lineIdx, line := range lines {
		lineRunes := []rune(line)
		for charIdx, char := range lineRunes {
			if char != ' ' && char != '\n' {
				x := r.centerX + charIdx
				y := r.centerY + lineIdx

				// Only store if within bounds
				if x >= 0 && x < r.width && y >= 0 && y < r.height {
					if r.artPositions[y] == nil {
						r.artPositions[y] = make(map[int]rune)
					}
					r.artPositions[y][x] = char
				}
			}
		}
	}
}

// init initializes rain drops
func (r *RainArtEffect) init() {
	// Create many initial drops for dense rain effect
	for i := 0; i < r.width*2; i++ {
		drop := RainDrop{
			X:     r.rng.Intn(r.width),
			Y:     -r.rng.Intn(r.height),
			Speed: r.rng.Intn(3) + 1,
			Char:  r.chars[r.rng.Intn(len(r.chars))],
			Color: r.getRandomColor(),
		}
		r.drops = append(r.drops, drop)
	}
}

// getRandomColor returns a random color from palette
func (r *RainArtEffect) getRandomColor() string {
	if len(r.palette) == 0 {
		return "#00aaff"
	}
	return r.palette[r.rng.Intn(len(r.palette))]
}

// Update advances the simulation by one frame
func (r *RainArtEffect) Update() {
	// Update existing drops
	activeDrops := r.drops[:0]
	for _, drop := range r.drops {
		// Check if drop should freeze at this position
		if _, yExists := r.artPositions[drop.Y]; yExists {
			if artChar, xExists := r.artPositions[drop.Y][drop.X]; xExists {
				// This position is part of the art
				if r.frozenChars[drop.Y] == nil || r.frozenChars[drop.Y][drop.X] == nil {
					// Position not yet frozen, maybe freeze it
					if r.rng.Float64() < r.freezeChance {
						// Freeze this character
						if r.frozenChars[drop.Y] == nil {
							r.frozenChars[drop.Y] = make(map[int]*FrozenChar)
						}
						r.frozenChars[drop.Y][drop.X] = &FrozenChar{
							char:  artChar,
							color: drop.Color,
						}
						// Don't add this drop back (it's frozen)
						continue
					}
				}
			}
		}

		// Move drop downward
		drop.Y += drop.Speed

		// Reset drop when it reaches bottom
		if drop.Y >= r.height {
			drop.Y = -r.rng.Intn(10)
			drop.X = r.rng.Intn(r.width)
			drop.Speed = r.rng.Intn(3) + 1
			drop.Char = r.chars[r.rng.Intn(len(r.chars))]
			drop.Color = r.getRandomColor()
		}

		activeDrops = append(activeDrops, drop)
	}
	r.drops = activeDrops

	// Aggressively spawn new drops to maintain high density
	for len(r.drops) < r.maxDrops && r.rng.Float64() < 0.5 {
		drop := RainDrop{
			X:     r.rng.Intn(r.width),
			Y:     -r.rng.Intn(10),
			Speed: r.rng.Intn(3) + 1,
			Char:  r.chars[r.rng.Intn(len(r.chars))],
			Color: r.getRandomColor(),
		}
		r.drops = append(r.drops, drop)
	}
}

// Render converts the rain and frozen art to colored output
func (r *RainArtEffect) Render() string {
	// Create empty canvas
	canvas := make([][]rune, r.height)
	colors := make([][]string, r.height)
	for i := range canvas {
		canvas[i] = make([]rune, r.width)
		colors[i] = make([]string, r.width)
		for j := range canvas[i] {
			canvas[i][j] = ' '
			colors[i][j] = ""
		}
	}

	// Place active rain drops on canvas
	for _, drop := range r.drops {
		if drop.Y >= 0 && drop.Y < r.height && drop.X >= 0 && drop.X < r.width {
			canvas[drop.Y][drop.X] = drop.Char
			colors[drop.Y][drop.X] = drop.Color
		}
	}

	// Place frozen characters on top (they override rain)
	for y, row := range r.frozenChars {
		for x, frozen := range row {
			if y >= 0 && y < r.height && x >= 0 && x < r.width {
				canvas[y][x] = frozen.char
				colors[y][x] = frozen.color
			}
		}
	}

	// Convert to colored string
	var lines []string
	for y := 0; y < r.height; y++ {
		var line strings.Builder
		for x := 0; x < r.width; x++ {
			char := canvas[y][x]
			if char != ' ' && colors[y][x] != "" {
				styled := lipgloss.NewStyle().
					Foreground(lipgloss.Color(colors[y][x])).
					Render(string(char))
				line.WriteString(styled)
			} else {
				line.WriteRune(char)
			}
		}
		lines = append(lines, line.String())
	}

	return strings.Join(lines, "\n")
}

// Reset clears frozen characters to restart the formation
func (r *RainArtEffect) Reset() {
	r.frozenChars = make(map[int]map[int]*FrozenChar)
}

// SetText updates the displayed text and restarts the crystallization.
func (r *RainArtEffect) SetText(text string) {
	r.text = text
	r.artPositions = make(map[int]map[int]rune)
	r.frozenChars = make(map[int]map[int]*FrozenChar)
	r.parseArt()
}
