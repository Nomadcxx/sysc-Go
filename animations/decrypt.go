package animations

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss/v2"
)

// DecryptEffect implements a movie-style text decryption animation
type DecryptEffect struct {
	width                  int
	height                 int
	text                   string
	chars                  []DecryptCharacter
	palette                []string
	typingSpeed            int
	ciphertextColors       []string
	finalGradientStops     []string
	finalGradientSteps     int
	finalGradientDirection string
	phase                  string
	frameCount             int
	rng                    *rand.Rand
}

// DecryptCharacter represents a single character in the decryption effect
type DecryptCharacter struct {
	original   rune
	current    rune
	x          int
	y          int
	visible    bool
	animation  []DecryptAnimationFrame
	frameIndex int
	duration   int
	color      string
}

// DecryptAnimationFrame represents a single frame in a character's animation
type DecryptAnimationFrame struct {
	symbol rune
	color  string
}

// DecryptConfig holds configuration for the decrypt effect
type DecryptConfig struct {
	Width                  int
	Height                 int
	Text                   string
	Palette                []string
	TypingSpeed            int
	CiphertextColors       []string
	FinalGradientStops     []string
	FinalGradientSteps     int
	FinalGradientDirection string
}

// NewDecryptEffect creates a new decrypt effect with given configuration
func NewDecryptEffect(config DecryptConfig) *DecryptEffect {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	effect := &DecryptEffect{
		width:                  config.Width,
		height:                 config.Height,
		text:                   config.Text,
		palette:                config.Palette,
		typingSpeed:            config.TypingSpeed,
		ciphertextColors:       config.CiphertextColors,
		finalGradientStops:     config.FinalGradientStops,
		finalGradientSteps:     config.FinalGradientSteps,
		finalGradientDirection: config.FinalGradientDirection,
		phase:                  "typing",
		rng:                    rng,
	}

	effect.init()
	return effect
}

// Initialize the decrypt effect with characters and their animations
func (d *DecryptEffect) init() {
	lines := strings.Split(d.text, "\n")

	// Calculate centered position for multi-line text
	startY := (d.height - len(lines)) / 2
	if startY < 0 {
		startY = 0
	}

	// Create characters from all lines
	for lineIdx, line := range lines {
		startX := (d.width - len(line)) / 2
		if startX < 0 {
			startX = 0
		}

		for charIdx, char := range line {
			finalX := startX + charIdx
			finalY := startY + lineIdx
			
			// Skip characters that would be off-screen
			if finalX >= d.width || finalY >= d.height {
				continue
			}
			
			d.chars = append(d.chars, DecryptCharacter{
				original: char,
				current:  char,
				x:        finalX,
				y:        finalY,
				visible:  false,
			})
		}
	}

	// Prepare animations for each character
	d.prepareAnimations()
}

// Prepare the animations for each character
func (d *DecryptEffect) prepareAnimations() {
	encryptedSymbols := d.makeEncryptedSymbols()

	// Calculate final colors with proper gradient
	finalColors := d.calculateGradientColors()

	for i := range d.chars {
		char := &d.chars[i]

		// Get a random color for this character's ciphertext
		ciphertextColor := d.ciphertextColors[d.rng.Intn(len(d.ciphertextColors))]

		// Prepare typing animation (block characters)
		typingAnimation := make([]DecryptAnimationFrame, 0)

		// Add block characters with same color
		blockChars := []rune{'▉', '▓', '▒', '░'}
		for _, blockChar := range blockChars {
			typingAnimation = append(typingAnimation, DecryptAnimationFrame{
				symbol: blockChar,
				color:  ciphertextColor,
			})
		}

		// Add one random encrypted symbol
		symbol := encryptedSymbols[d.rng.Intn(len(encryptedSymbols))]
		typingAnimation = append(typingAnimation, DecryptAnimationFrame{
			symbol: symbol,
			color:  ciphertextColor,
		})

		// Prepare decrypting animations
		decryptAnimation := make([]DecryptAnimationFrame, 0)

		// Fast decrypt phase (80 frames with short duration = 3)
		for j := 0; j < 80; j++ {
			symbol := encryptedSymbols[d.rng.Intn(len(encryptedSymbols))]
			decryptAnimation = append(decryptAnimation, DecryptAnimationFrame{
				symbol: symbol,
				color:  ciphertextColor,
			})
		}

		// Slow decrypt phase (1-15 frames with variable durations)
		slowFrames := d.rng.Intn(15) + 1
		for j := 0; j < slowFrames; j++ {
			symbol := encryptedSymbols[d.rng.Intn(len(encryptedSymbols))]
			decryptAnimation = append(decryptAnimation, DecryptAnimationFrame{
				symbol: symbol,
				color:  ciphertextColor,
			})
		}

		// Discovered phase - create gradient transition from white to final color
		discoveredGradient := d.createSimpleGradient("#ffffff", finalColors[i], 15)
		for _, color := range discoveredGradient {
			decryptAnimation = append(decryptAnimation, DecryptAnimationFrame{
				symbol: char.original,
				color:  color,
			})
		}
		
		// Hold on final decrypted text for remaining duration
		for j := 0; j < 50; j++ {
			decryptAnimation = append(decryptAnimation, DecryptAnimationFrame{
				symbol: char.original,
				color:  finalColors[i],
			})
		}

		char.animation = append(typingAnimation, decryptAnimation...)
	}
}

// Create a list of encrypted symbols
func (d *DecryptEffect) makeEncryptedSymbols() []rune {
	var symbols []rune

	// Keyboard characters (33-126)
	for i := 33; i <= 126; i++ {
		symbols = append(symbols, rune(i))
	}

	// Block characters (9608-9631)
	for i := 9608; i <= 9631; i++ {
		symbols = append(symbols, rune(i))
	}

	// Box drawing characters (9472-9599)
	for i := 9472; i <= 9599; i++ {
		symbols = append(symbols, rune(i))
	}

	// Misc characters (174-451)
	for i := 174; i <= 451; i++ {
		symbols = append(symbols, rune(i))
	}

	return symbols
}

// Calculate gradient colors for all characters based on coordinates
func (d *DecryptEffect) calculateGradientColors() []string {
	colors := make([]string, len(d.chars))

	if len(d.finalGradientStops) == 0 {
		defaultColor := "#eda000"
		for i := range colors {
			colors[i] = defaultColor
		}
		return colors
	}

	if len(d.finalGradientStops) == 1 {
		for i := range colors {
			colors[i] = d.finalGradientStops[0]
		}
		return colors
	}

	// Find min/max coordinates for normalization
	minX, maxX := d.width, 0
	minY, maxY := d.height, 0
	for _, char := range d.chars {
		if char.x < minX {
			minX = char.x
		}
		if char.x > maxX {
			maxX = char.x
		}
		if char.y < minY {
			minY = char.y
		}
		if char.y > maxY {
			maxY = char.y
		}
	}

	// Calculate gradient for each character based on position
	for i := range d.chars {
		char := d.chars[i]
		var ratio float64

		if d.finalGradientDirection == "vertical" {
			// Vertical gradient (top to bottom)
			if maxY > minY {
				ratio = float64(char.y-minY) / float64(maxY-minY)
			}
		} else {
			// Horizontal gradient (left to right)
			if maxX > minX {
				ratio = float64(char.x-minX) / float64(maxX-minX)
			}
		}

		// Map ratio to gradient stops
		step := int(ratio * float64(len(d.finalGradientStops)-1))
		if step >= len(d.finalGradientStops) {
			step = len(d.finalGradientStops) - 1
		}
		if step < 0 {
			step = 0
		}

		colors[i] = d.finalGradientStops[step]
	}

	return colors
}

// Create a simple gradient between two colors with specified steps
func (d *DecryptEffect) createSimpleGradient(startColor, endColor string, steps int) []string {
	if steps <= 0 {
		return []string{endColor}
	}

	gradient := make([]string, steps)

	// Parse start color
	startR, startG, startB := d.parseHexColor(startColor)

	// Parse end color
	endR, endG, endB := d.parseHexColor(endColor)

	// Calculate step increments
	rStep := float64(endR-startR) / float64(steps-1)
	gStep := float64(endG-startG) / float64(steps-1)
	bStep := float64(endB-startB) / float64(steps-1)

	// Generate gradient colors
	for i := 0; i < steps; i++ {
		r := int(float64(startR) + float64(i)*rStep)
		g := int(float64(startG) + float64(i)*gStep)
		b := int(float64(startB) + float64(i)*bStep)

		// Clamp values to 0-255
		r = d.clamp(r, 0, 255)
		g = d.clamp(g, 0, 255)
		b = d.clamp(b, 0, 255)

		gradient[i] = fmt.Sprintf("#%02x%02x%02x", r, g, b)
	}

	return gradient
}

// Parse hex color string to RGB values
func (d *DecryptEffect) parseHexColor(hex string) (int, int, int) {
	if len(hex) < 8 || hex[0] != '#' {
		// Default to white if invalid
		return 255, 255, 255
	}

	r, _ := strconv.ParseInt(hex[1:3], 16, 64)
	g, _ := strconv.ParseInt(hex[3:5], 16, 64)
	b, _ := strconv.ParseInt(hex[5:7], 16, 64)

	return int(r), int(g), int(b)
}

// Clamp value between min and max
func (d *DecryptEffect) clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// Update advances the decrypt animation by one frame
func (d *DecryptEffect) Update() {
	d.frameCount++

	switch d.phase {
	case "typing":
		d.updateTypingPhase()
	case "decrypting":
		d.updateDecryptingPhase()
	case "complete":
		return
	}
}

// Update the typing phase of the animation
func (d *DecryptEffect) updateTypingPhase() {
	// Randomly decide whether to type new characters (75% chance)
	if len(d.getVisibleChars()) < len(d.chars) && d.rng.Intn(100) <= 75 {
		// Make a few characters visible based on typing speed
		for i := 0; i < d.typingSpeed; i++ {
			visibleCount := len(d.getVisibleChars())
			if visibleCount < len(d.chars) {
				// Find the next invisible character
				for j := 0; j < len(d.chars); j++ {
					if !d.chars[j].visible {
						d.chars[j].visible = true
						d.chars[j].frameIndex = 0
						d.chars[j].duration = 0
						break
					}
				}
			}
		}
	}

	// Update visible characters
	for i := range d.chars {
		if d.chars[i].visible {
			d.updateCharacter(&d.chars[i])
		}
	}

	// Transition to decrypting phase when typing is complete
	if len(d.getVisibleChars()) == len(d.chars) && d.allCharsStill() {
		d.phase = "decrypting"
		// Reset frame indices for decrypting phase
		for i := range d.chars {
			// Set frame index to start of decrypting animation (after typing frames)
			typingFrames := 5 // 4 block chars + 1 encrypted symbol
			if d.chars[i].frameIndex < typingFrames {
				d.chars[i].frameIndex = typingFrames
			}
			d.chars[i].duration = 0
		}
	}
}

// Check if all visible characters are still (not animating)
func (d *DecryptEffect) allCharsStill() bool {
	for _, char := range d.chars {
		if char.visible {
			// Check if character is still in typing phase
			if char.frameIndex < 5 { // 5 typing frames (4 blocks + 1 encrypted)
				return false
			}
		}
	}
	return true
}

// Update the decrypting phase of the animation
func (d *DecryptEffect) updateDecryptingPhase() {
	allDone := true

	for i := range d.chars {
		char := &d.chars[i]
		// Continue updating until we've shown all frames including the last one
		if char.frameIndex < len(char.animation) {
			allDone = false
			d.updateCharacter(char)
		} else {
			// Ensure last frame is set
			if len(char.animation) > 0 {
				lastFrame := char.animation[len(char.animation)-1]
				char.current = lastFrame.symbol
				char.color = lastFrame.color
			}
		}
	}

	// Move to complete phase when all done
	if allDone {
		d.phase = "complete"
	}
}

// Update a single character's animation
func (d *DecryptEffect) updateCharacter(char *DecryptCharacter) {
	if len(char.animation) == 0 {
		return
	}

	// Update duration counter
	char.duration++

	// Determine frame duration based on current animation phase
	frameDuration := 3 // Default for typing phase (slowed down)

	// Check if we're in the decrypting phase (past the typing frames)
	typingFrames := 5 // 4 block chars + 1 encrypted symbol
	if char.frameIndex >= typingFrames {
		// Decrypting phase - much slower variable durations
		if d.rng.Intn(100) <= 40 {
			frameDuration = d.rng.Intn(100) + 80 // Longer duration (80-180)
		} else {
			frameDuration = d.rng.Intn(10) + 10 // Shorter duration (10-19)
		}
	}

	// Advance frame if duration has elapsed
	if char.duration >= frameDuration {
		char.frameIndex++
		char.duration = 0

		// Update current symbol and color (but allow frameIndex to exceed length)
		if char.frameIndex < len(char.animation) {
			frame := char.animation[char.frameIndex]
			char.current = frame.symbol
			char.color = frame.color
		}
		// Don't cap frameIndex - let it go beyond array length to signal completion
	}
}

// Get visible characters
func (d *DecryptEffect) getVisibleChars() []DecryptCharacter {
	var visible []DecryptCharacter
	for _, char := range d.chars {
		if char.visible {
			visible = append(visible, char)
		}
	}
	return visible
}

// Render converts the decrypt effect to colored text output
func (d *DecryptEffect) Render() string {
	// Create a buffer to hold the output
	buffer := make([][]string, d.height)
	for i := range buffer {
		buffer[i] = make([]string, d.width)
		for j := range buffer[i] {
			buffer[i][j] = " "
		}
	}

	// Render visible characters
	for _, char := range d.chars {
		if char.visible && char.y >= 0 && char.y < d.height && char.x >= 0 && char.x < d.width {
			style := lipgloss.NewStyle().Foreground(lipgloss.Color(char.color))
			buffer[char.y][char.x] = style.Render(string(char.current))
		}
	}

	// Convert buffer to string
	var lines []string
	for _, line := range buffer {
		lines = append(lines, strings.Join(line, ""))
	}

	return strings.Join(lines, "\n")
}

// Reset restarts the animation from the beginning
func (d *DecryptEffect) Reset() {
	d.phase = "typing"
	d.frameCount = 0

	// Reset character states
	for i := range d.chars {
		d.chars[i].visible = false
		d.chars[i].frameIndex = 0
		d.chars[i].duration = 0
		d.chars[i].current = d.chars[i].original
	}

	// Reprepare animations
	d.prepareAnimations()
}
