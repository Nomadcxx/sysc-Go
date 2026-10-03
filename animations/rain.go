package animations

import (
	"math/rand"
	"strings"

	"github.com/charmbracelet/lipgloss/v2"
)

// RainEffect implements ASCII character rain animation
type RainEffect struct {
	width    int      // Terminal width
	height   int      // Terminal height
	palette  []string // Theme color palette
	chars    []rune   // Raindrop characters
	drops    []RainDrop
	maxDrops int // Maximum number of simultaneous drops

	// Reusable render buffers — allocated once, cleared per frame
	canvas  [][]rune
	colors  [][]string
	builder strings.Builder
}

// RainDrop represents a single falling character
type RainDrop struct {
	X     int    // X position
	Y     int    // Y position
	Speed int    // Falling speed
	Char  rune   // Character to display
	Color string // Color hex code
}

// NewRainEffect creates a new rain effect with given dimensions and theme palette
func NewRainEffect(width, height int, palette []string) *RainEffect {
	r := &RainEffect{
		width:    width,
		height:   height,
		palette:  palette,
		chars:    []rune{'|', '⋮', '║', '¦', '┆', '┊', '╎', '╏', '▏', '▎', '▍', '▌', '▋', '▊', '▉'},
		drops:    make([]RainDrop, 0, 200),
		maxDrops: width * 2, // More drops for wider terminals
	}
	r.init()
	return r
}

// Initialize rain effect with some initial drops
func (r *RainEffect) init() {
	r.initBuffers()

	// Create initial drops scattered across width
	for i := 0; i < r.width/3; i++ {
		drop := RainDrop{
			X:     rand.Intn(r.width),
			Y:     -rand.Intn(r.height), // Start above screen
			Speed: rand.Intn(3) + 1,     // Speed 1-3
			Char:  r.chars[rand.Intn(len(r.chars))],
			Color: r.getRandomColor(),
		}
		r.drops = append(r.drops, drop)
	}
}

func (r *RainEffect) initBuffers() {
	r.canvas = make([][]rune, r.height)
	r.colors = make([][]string, r.height)
	for i := range r.canvas {
		r.canvas[i] = make([]rune, r.width)
		r.colors[i] = make([]string, r.width)
	}
	r.builder.Grow(r.width * r.height * 4)
}

func (r *RainEffect) clearBuffers() {
	for i := range r.canvas {
		for j := range r.canvas[i] {
			r.canvas[i][j] = ' '
			r.colors[i][j] = ""
		}
	}
}

// UpdatePalette changes the rain color palette (for theme switching)
func (r *RainEffect) UpdatePalette(palette []string) {
	r.palette = palette
}

// Resize reinitializes the rain effect with new dimensions
func (r *RainEffect) Resize(width, height int) {
	r.width = width
	r.height = height
	r.maxDrops = width * 2
	r.initBuffers()
	r.init()
}

// getRandomColor returns a random color from the theme palette
func (r *RainEffect) getRandomColor() string {
	if len(r.palette) == 0 {
		return "#00aaff" // Default blue if no palette
	}
	return r.palette[rand.Intn(len(r.palette))]
}

// Update advances the rain simulation by one frame
func (r *RainEffect) Update() {
	// Update existing drops
	activeDrops := r.drops[:0] // Reuse slice for efficiency
	for _, drop := range r.drops {
		// Move drop downward
		drop.Y += drop.Speed

		// Reset drop when it reaches bottom
		if drop.Y >= r.height {
			drop.Y = -rand.Intn(10) // Start above screen
			drop.X = rand.Intn(r.width)
			drop.Speed = rand.Intn(3) + 1 // Speed 1-3
			drop.Char = r.chars[rand.Intn(len(r.chars))]
			drop.Color = r.getRandomColor()
		}

		activeDrops = append(activeDrops, drop)
	}
	r.drops = activeDrops

	// Add new drops randomly
	for len(r.drops) < r.maxDrops && rand.Float64() < 0.3 {
		drop := RainDrop{
			X:     rand.Intn(r.width),
			Y:     -rand.Intn(10),   // Start above screen
			Speed: rand.Intn(3) + 1, // Speed 1-3
			Char:  r.chars[rand.Intn(len(r.chars))],
			Color: r.getRandomColor(),
		}
		r.drops = append(r.drops, drop)
	}
}

// Render converts the rain drops to colored text output
func (r *RainEffect) Render() string {
	r.clearBuffers()

	// Place active drops on canvas
	for _, drop := range r.drops {
		if drop.Y >= 0 && drop.Y < r.height && drop.X >= 0 && drop.X < r.width {
			r.canvas[drop.Y][drop.X] = drop.Char
			r.colors[drop.Y][drop.X] = drop.Color
		}
	}

	// Convert to colored string
	var lines []string
	for y := 0; y < r.height; y++ {
		r.builder.Reset()
		for x := 0; x < r.width; x++ {
			char := r.canvas[y][x]
			if char != ' ' && r.colors[y][x] != "" {
				// Render colored character
				styled := lipgloss.NewStyle().
					Foreground(lipgloss.Color(r.colors[y][x])).
					Render(string(char))
				r.builder.WriteString(styled)
			} else {
				r.builder.WriteRune(char)
			}
		}
		lines = append(lines, r.builder.String())
	}

	return strings.Join(lines, "\n")
}

// Reset restarts the animation from the beginning
func (r *RainEffect) Reset() {
	r.drops = r.drops[:0]
	r.init()
}
