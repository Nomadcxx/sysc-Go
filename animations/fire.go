package animations

import (
	"bytes"
	"math/rand"
)

// FireEffect implements PSX DOOM-style fire algorithm with enhanced character gradient
type FireEffect struct {
	width   int      // Terminal width
	height  int      // Terminal height
	buffer  []int    // Heat values (0-65), size = width * height
	palette []string // Hex color codes from theme
	colors  []ansiColor
	chars   []rune // Fire characters for density (8-level gradient)
	output  bytes.Buffer
}

// NewFireEffect creates a new fire effect with given dimensions and theme palette
func NewFireEffect(width, height int, palette []string) *FireEffect {
	f := &FireEffect{
		width:   width,
		height:  height,
		palette: palette,
		colors:  cacheANSIColors(palette),
		// Enhanced 8-character gradient for smoother fire rendering
		chars: []rune{' ', '░', '░', '▒', '▒', '▓', '▓', '█'},
	}
	f.init()
	return f
}

// Initialize fire buffer with bottom row as heat source.
// A zero or negative dimension has no bottom row; indexing height-1 panics.
func (f *FireEffect) init() {
	if f.width <= 0 || f.height <= 0 {
		f.buffer = make([]int, 0)
		return
	}

	f.buffer = make([]int, f.width*f.height)

	// Set bottom row to maximum heat (fire source)
	for i := 0; i < f.width; i++ {
		f.buffer[(f.height-1)*f.width+i] = 65
	}
}

// UpdatePalette changes the fire color palette (for theme switching)
func (f *FireEffect) UpdatePalette(palette []string) {
	f.palette = palette
	f.colors = cacheANSIColors(palette)
}

// Resize reinitializes the fire effect with new dimensions
func (f *FireEffect) Resize(width, height int) {
	f.width = width
	f.height = height
	f.colors = cacheANSIColors(f.palette)
	f.init()
}

// spreadFire propagates heat upward with random decay (DOOM algorithm)
func (f *FireEffect) spreadFire(from int) {
	// Random horizontal offset (0-3) for flickering effect
	offset := rand.Intn(4)
	to := from - f.width - offset + 1

	// Bounds check
	if to < 0 || to >= len(f.buffer) {
		return
	}

	// Random decay (0-3) for natural fade
	decay := rand.Intn(4)

	newHeat := f.buffer[from] - decay
	if newHeat < 0 {
		newHeat = 0
	}

	f.buffer[to] = newHeat
}

// Update advances the fire simulation by one frame
func (f *FireEffect) Update() {
	// Process all pixels from bottom to top
	// (Fire spreads upward, must process bottom row first)
	for y := f.height - 1; y > 0; y-- {
		for x := 0; x < f.width; x++ {
			index := y*f.width + x
			f.spreadFire(index)
		}
	}
}

// Render converts fire to colored block output with batched raw ANSI codes
func (f *FireEffect) Render() string {
	f.output.Reset()
	output := &f.output

	// Always render full viewport height to anchor fire at bottom
	// This prevents jumping as fire spreads upward
	for y := 0; y < f.height; y++ {
		var currentColor string
		colored := false
		for x := 0; x < f.width; x++ {
			heat := f.buffer[y*f.width+x]

			// Skip very low heat (natural fade to background)
			if heat < 5 {
				// Flush any pending batch
				if colored {
					output.WriteString(ansiReset)
					colored = false
				}
				output.WriteString(" ")
				currentColor = ""
				continue
			}

			// Map heat to character (0-65 → 8 chars)
			charIndex := (heat * (len(f.chars) - 1)) / 65
			if charIndex >= len(f.chars) {
				charIndex = len(f.chars) - 1
			}
			char := f.chars[charIndex]

			// Map heat to color from palette
			colorIndex := (heat * (len(f.palette) - 1)) / 65
			if colorIndex >= len(f.palette) {
				colorIndex = len(f.palette) - 1
			}
			colorHex := f.palette[colorIndex]

			// If color changed, flush previous batch and start new one
			if !colored || colorHex != currentColor {
				if colored {
					output.WriteString(ansiReset)
				}
				output.WriteString(f.colors[colorIndex].prefix)
				currentColor = colorHex
				colored = true
			}

			output.WriteRune(char)
		}

		// Flush any remaining colored run at end of line.
		if colored {
			output.WriteString(ansiReset)
			colored = false
		}

		if y < f.height-1 {
			output.WriteByte('\n')
		}
	}

	return output.String()
}
