package animations

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss/v2"
)

// BeamsEffect implements beams that travel across rows and columns, illuminating text
type BeamsEffect struct {
	width  int
	height int
	text   string

	// Configuration
	beamRowSymbols      []rune
	beamColumnSymbols   []rune
	beamDelay           int
	beamRowSpeedRange   [2]int
	beamColumnSpeedRange [2]int
	beamGradientStops   []string
	beamGradientSteps   int
	beamGradientFrames  int
	finalGradientStops  []string
	finalGradientSteps  int
	finalGradientFrames int
	finalWipeSpeed      int

	// Character data
	chars []BeamCharacter

	// Beam groups
	rowGroups    []BeamGroup
	columnGroups []BeamGroup

	// Final wipe diagonal groups
	diagonalGroups [][]int

	// Animation state
	phase          string // "beams", "final_wipe", or "hold"
	frameCount     int
	beamDelayCount int
	currentDiag    int
	holdFrames     int // Frames to hold after completion
	holdCounter    int // Current hold frame count

	rng *rand.Rand
}

// BeamCharacter represents a single character in the beams animation
type BeamCharacter struct {
	original rune
	x        int
	y        int

	// Animation state
	visible         bool
	currentSymbol   rune
	currentColor    string
	sceneActive     string // "beam_row", "beam_column", or "brighten"
	sceneFrame      int
	beamGradient    []string
	fadeGradient    []string
	brightenGradient []string
}

// BeamGroup represents a group of characters for beam animation
type BeamGroup struct {
	charIndices        []int
	direction          string  // "row" or "column"
	speed              float64
	nextCharCounter    float64
	currentCharIndex   int
	symbols            []rune
	beamGradientStops  []string
	beamGradientSteps  int
	beamGradientFrames int
	beamLength         int // Length of visible beam trail
}

// BeamsConfig holds configuration for the beams effect
type BeamsConfig struct {
	Width                int
	Height               int
	Text                 string
	BeamRowSymbols       []rune
	BeamColumnSymbols    []rune
	BeamDelay            int
	BeamRowSpeedRange    [2]int
	BeamColumnSpeedRange [2]int
	BeamGradientStops    []string
	BeamGradientSteps    int
	BeamGradientFrames   int
	FinalGradientStops   []string
	FinalGradientSteps   int
	FinalGradientFrames  int
	FinalWipeSpeed       int
}

// NewBeamsEffect creates a new beams effect with given configuration
func NewBeamsEffect(config BeamsConfig) *BeamsEffect {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Set defaults if not provided
	if len(config.BeamRowSymbols) == 0 {
		config.BeamRowSymbols = []rune{'▂', '▁', '_'}
	}
	if len(config.BeamColumnSymbols) == 0 {
		config.BeamColumnSymbols = []rune{'▌', '▍', '▎', '▏'}
	}
	if config.BeamDelay == 0 {
		config.BeamDelay = 2 // Faster group activation
	}
	if config.BeamRowSpeedRange[0] == 0 {
		config.BeamRowSpeedRange = [2]int{20, 80} // Much faster speeds
	}
	if config.BeamColumnSpeedRange[0] == 0 {
		config.BeamColumnSpeedRange = [2]int{15, 30} // Much faster speeds
	}
	if config.BeamGradientSteps == 0 {
		config.BeamGradientSteps = 5 // Shorter gradient
	}
	if config.BeamGradientFrames == 0 {
		config.BeamGradientFrames = 1
	}
	if config.FinalGradientSteps == 0 {
		config.FinalGradientSteps = 8 // Shorter gradient
	}
	if config.FinalGradientFrames == 0 {
		config.FinalGradientFrames = 1
	}
	if config.FinalWipeSpeed == 0 {
		config.FinalWipeSpeed = 3 // Activate multiple diagonal groups per frame
	}

	// Adjust speeds for background mode
	rowSpeedRange := config.BeamRowSpeedRange
	colSpeedRange := config.BeamColumnSpeedRange
	beamDelay := config.BeamDelay
	gradientSteps := config.BeamGradientSteps

	if config.Text == "" {
		// Background mode: much faster, denser beams
		rowSpeedRange = [2]int{40, 120}
		colSpeedRange = [2]int{30, 60}
		beamDelay = 1
		gradientSteps = 3
	}

	b := &BeamsEffect{
		width:                config.Width,
		height:               config.Height,
		text:                 config.Text,
		beamRowSymbols:       config.BeamRowSymbols,
		beamColumnSymbols:    config.BeamColumnSymbols,
		beamDelay:            beamDelay,
		beamRowSpeedRange:    rowSpeedRange,
		beamColumnSpeedRange: colSpeedRange,
		beamGradientStops:    config.BeamGradientStops,
		beamGradientSteps:    gradientSteps,
		beamGradientFrames:   config.BeamGradientFrames,
		finalGradientStops:   config.FinalGradientStops,
		finalGradientSteps:   config.FinalGradientSteps,
		finalGradientFrames:  config.FinalGradientFrames,
		finalWipeSpeed:       config.FinalWipeSpeed,
		phase:                "beams",
		frameCount:           0,
		beamDelayCount:       0,
		currentDiag:          0,
		holdFrames:           100, // 5 seconds at 20fps
		holdCounter:          0,
		rng:                  rng,
	}

	b.init()
	return b
}

// init initializes characters and beam groups
func (b *BeamsEffect) init() {
	if b.text == "" {
		// Background mode: fill entire terminal with invisible positions
		b.initBackgroundMode()
	} else {
		// Text mode: use provided text
		b.initTextMode()
	}

	// Create row groups
	b.createRowGroups()

	// Create column groups
	b.createColumnGroups()

	// Shuffle groups for random activation
	b.shuffleGroups()

	// Create diagonal groups for final wipe
	b.createDiagonalGroups()
}

// initTextMode initializes with centered text
func (b *BeamsEffect) initTextMode() {
	lines := strings.Split(b.text, "\n")

	// Calculate centered position for text block
	startY := (b.height - len(lines)) / 2
	if startY < 0 {
		startY = 0
	}

	// Find the longest line for centering the entire block
	maxWidth := 0
	for _, line := range lines {
		runes := []rune(line)
		if len(runes) > maxWidth {
			maxWidth = len(runes)
		}
	}

	// Center based on longest line
	blockStartX := (b.width - maxWidth) / 2
	if blockStartX < 0 {
		blockStartX = 0
	}

	// Create characters from text
	for lineIdx, line := range lines {
		runes := []rune(line)

		for charIdx, char := range runes {
			if char == ' ' || char == '\t' {
				continue
			}

			x := blockStartX + charIdx
			y := startY + lineIdx

			if x >= b.width || y >= b.height {
				continue
			}

			// Create beam gradients for this character
			beamGradient := b.createGradient(b.beamGradientStops, b.beamGradientSteps)
			fadeGradient := b.createFadeGradient(beamGradient[len(beamGradient)-1], 5)
			brightenGradient := b.createGradient(b.finalGradientStops, b.finalGradientSteps)

			b.chars = append(b.chars, BeamCharacter{
				original:         char,
				x:                x,
				y:                y,
				visible:          false,
				currentSymbol:    char,
				currentColor:     "",
				sceneActive:      "",
				sceneFrame:       0,
				beamGradient:     beamGradient,
				fadeGradient:     fadeGradient,
				brightenGradient: brightenGradient,
			})
		}
	}
}

// initBackgroundMode initializes full-screen background mode
func (b *BeamsEffect) initBackgroundMode() {
	// Create beam gradients
	beamGradient := b.createGradient(b.beamGradientStops, b.beamGradientSteps)
	fadeGradient := b.createFadeGradient(beamGradient[len(beamGradient)-1], 3)

	// Fill terminal with dense distribution for glowing effect
	// Every position for maximum density
	for y := 0; y < b.height; y++ {
		for x := 0; x < b.width; x++ {
			b.chars = append(b.chars, BeamCharacter{
				original:         ' ',
				x:                x,
				y:                y,
				visible:          false,
				currentSymbol:    ' ',
				currentColor:     "",
				sceneActive:      "",
				sceneFrame:       0,
				beamGradient:     beamGradient,
				fadeGradient:     fadeGradient,
				brightenGradient: nil, // No final brighten in background mode
			})
		}
	}
}

// createRowGroups creates beam groups for each row
func (b *BeamsEffect) createRowGroups() {
	// Group characters by row
	rowMap := make(map[int][]int)
	for i, char := range b.chars {
		rowMap[char.y] = append(rowMap[char.y], i)
	}

	// Create groups
	for _, indices := range rowMap {
		// Sort by x coordinate
		sort.Slice(indices, func(i, j int) bool {
			return b.chars[indices[i]].x < b.chars[indices[j]].x
		})

		// Randomly reverse
		if b.rng.Float64() < 0.5 {
			for i := 0; i < len(indices)/2; i++ {
				j := len(indices) - 1 - i
				indices[i], indices[j] = indices[j], indices[i]
			}
		}

		speed := float64(b.rng.Intn(b.beamRowSpeedRange[1]-b.beamRowSpeedRange[0])+b.beamRowSpeedRange[0]) * 0.1

		b.rowGroups = append(b.rowGroups, BeamGroup{
			charIndices:        indices,
			direction:          "row",
			speed:              speed,
			nextCharCounter:    0,
			currentCharIndex:   0,
			symbols:            b.beamRowSymbols,
			beamGradientStops:  b.beamGradientStops,
			beamGradientSteps:  b.beamGradientSteps,
			beamGradientFrames: b.beamGradientFrames,
			beamLength:         len(b.beamRowSymbols),
		})
	}
}

// createColumnGroups creates beam groups for each column
func (b *BeamsEffect) createColumnGroups() {
	// Group characters by column
	colMap := make(map[int][]int)
	for i, char := range b.chars {
		colMap[char.x] = append(colMap[char.x], i)
	}

	// Create groups
	for _, indices := range colMap {
		// Sort by y coordinate
		sort.Slice(indices, func(i, j int) bool {
			return b.chars[indices[i]].y < b.chars[indices[j]].y
		})

		// Randomly reverse
		if b.rng.Float64() < 0.5 {
			for i := 0; i < len(indices)/2; i++ {
				j := len(indices) - 1 - i
				indices[i], indices[j] = indices[j], indices[i]
			}
		}

		speed := float64(b.rng.Intn(b.beamColumnSpeedRange[1]-b.beamColumnSpeedRange[0])+b.beamColumnSpeedRange[0]) * 0.1

		b.columnGroups = append(b.columnGroups, BeamGroup{
			charIndices:        indices,
			direction:          "column",
			speed:              speed,
			nextCharCounter:    0,
			currentCharIndex:   0,
			symbols:            b.beamColumnSymbols,
			beamGradientStops:  b.beamGradientStops,
			beamGradientSteps:  b.beamGradientSteps,
			beamGradientFrames: b.beamGradientFrames,
			beamLength:         len(b.beamColumnSymbols),
		})
	}
}

// shuffleGroups shuffles row and column groups together
func (b *BeamsEffect) shuffleGroups() {
	// Combine both types of groups
	allGroups := append(b.rowGroups, b.columnGroups...)

	// Fisher-Yates shuffle
	for i := len(allGroups) - 1; i > 0; i-- {
		j := b.rng.Intn(i + 1)
		allGroups[i], allGroups[j] = allGroups[j], allGroups[i]
	}

	// Split back
	b.rowGroups = b.rowGroups[:0]
	b.columnGroups = b.columnGroups[:0]

	for _, group := range allGroups {
		if group.direction == "row" {
			b.rowGroups = append(b.rowGroups, group)
		} else {
			b.columnGroups = append(b.columnGroups, group)
		}
	}
}

// createDiagonalGroups creates diagonal groups for final wipe
func (b *BeamsEffect) createDiagonalGroups() {
	// Group by diagonal (top-left to bottom-right)
	diagMap := make(map[int][]int)
	for i, char := range b.chars {
		diag := char.x + char.y
		diagMap[diag] = append(diagMap[diag], i)
	}

	// Sort by diagonal index and create groups
	keys := make([]int, 0, len(diagMap))
	for k := range diagMap {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	for _, k := range keys {
		b.diagonalGroups = append(b.diagonalGroups, diagMap[k])
	}
}

// createGradient creates a color gradient from stops
func (b *BeamsEffect) createGradient(stops []string, steps int) []string {
	if len(stops) == 0 {
		return []string{"#ffffff"}
	}
	if len(stops) == 1 {
		return []string{stops[0]}
	}

	var gradient []string
	stepsPerSegment := steps / (len(stops) - 1)

	for i := 0; i < len(stops)-1; i++ {
		c1 := parseHexColor(stops[i])
		c2 := parseHexColor(stops[i+1])

		for j := 0; j < stepsPerSegment; j++ {
			t := float64(j) / float64(stepsPerSegment)
			r := uint8(float64(c1[0])*(1-t) + float64(c2[0])*t)
			g := uint8(float64(c1[1])*(1-t) + float64(c2[1])*t)
			b := uint8(float64(c1[2])*(1-t) + float64(c2[2])*t)
			gradient = append(gradient, formatHexColor([3]uint8{r, g, b}))
		}
	}

	// Add final color
	gradient = append(gradient, stops[len(stops)-1])

	return gradient
}

// createFadeGradient creates a fade to dark gradient
func (b *BeamsEffect) createFadeGradient(startColor string, steps int) []string {
	rgb := parseHexColor(startColor)
	targetRGB := [3]uint8{
		uint8(float64(rgb[0]) * 0.3),
		uint8(float64(rgb[1]) * 0.3),
		uint8(float64(rgb[2]) * 0.3),
	}

	var gradient []string
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		r := uint8(float64(rgb[0])*(1-t) + float64(targetRGB[0])*t)
		g := uint8(float64(rgb[1])*(1-t) + float64(targetRGB[1])*t)
		b := uint8(float64(rgb[2])*(1-t) + float64(targetRGB[2])*t)
		gradient = append(gradient, formatHexColor([3]uint8{r, g, b}))
	}

	return gradient
}

// Update advances the beams animation by one frame
func (b *BeamsEffect) Update() {
	b.frameCount++

	if b.phase == "beams" {
		b.updateBeamsPhase()
	} else if b.phase == "final_wipe" {
		b.updateFinalWipePhase()
	} else if b.phase == "hold" {
		b.updateHoldPhase()
	}

	// Update character animations
	b.updateCharacterAnimations()
}

// updateBeamsPhase handles the beam movement phase
func (b *BeamsEffect) updateBeamsPhase() {
	// Decrement delay counter
	if b.beamDelayCount > 0 {
		b.beamDelayCount--
		return
	}

	// Activate next group(s)
	groupsToActivate := b.rng.Intn(5) + 1
	activated := false

	for i := 0; i < groupsToActivate; i++ {
		// Try to activate a row group
		for j := range b.rowGroups {
			if b.rowGroups[j].currentCharIndex == 0 && b.rowGroups[j].nextCharCounter == 0 {
				b.rowGroups[j].nextCharCounter = 0.01 // Start the group
				activated = true
				break
			}
		}

		// Try to activate a column group
		for j := range b.columnGroups {
			if b.columnGroups[j].currentCharIndex == 0 && b.columnGroups[j].nextCharCounter == 0 {
				b.columnGroups[j].nextCharCounter = 0.01
				activated = true
				break
			}
		}
	}

	if activated {
		b.beamDelayCount = b.beamDelay
	}

	// Update all active groups
	allGroupsComplete := true

	for i := range b.rowGroups {
		if b.updateGroup(&b.rowGroups[i]) {
			allGroupsComplete = false
		}
	}

	for i := range b.columnGroups {
		if b.updateGroup(&b.columnGroups[i]) {
			allGroupsComplete = false
		}
	}

	// Check if all groups are complete
	if allGroupsComplete {
		b.phase = "final_wipe"
	}
}

// updateGroup updates a single beam group and returns true if still active
func (b *BeamsEffect) updateGroup(group *BeamGroup) bool {
	if group.nextCharCounter == 0 {
		return false // Group not started
	}

	if group.currentCharIndex >= len(group.charIndices) {
		return false // Group complete
	}

	// Increment counter
	group.nextCharCounter += group.speed

	// Activate characters
	charsToActivate := int(group.nextCharCounter)
	group.nextCharCounter -= float64(charsToActivate)

	for i := 0; i < charsToActivate && group.currentCharIndex < len(group.charIndices); i++ {
		charIdx := group.charIndices[group.currentCharIndex]
		char := &b.chars[charIdx]

		// Activate beam scene
		if group.direction == "row" {
			char.sceneActive = "beam_row"
		} else {
			char.sceneActive = "beam_column"
		}
		char.sceneFrame = 0
		char.visible = true

		// Use symbol based on position in beam for gradient effect
		// Most recent chars get thickest symbol, trailing chars get thinner
		symbolIndex := 0 // Default to thickest
		if len(group.symbols) > 0 {
			symbolIndex = 0 // Head of beam is always thickest
		}
		char.currentSymbol = group.symbols[symbolIndex]

		// Update trailing characters to use progressively thinner symbols
		for j := 1; j < group.beamLength && group.currentCharIndex-j >= 0; j++ {
			trailCharIdx := group.charIndices[group.currentCharIndex-j]
			trailChar := &b.chars[trailCharIdx]

			if trailChar.sceneActive == "beam_row" || trailChar.sceneActive == "beam_column" {
				symbolIdx := j
				if symbolIdx >= len(group.symbols) {
					symbolIdx = len(group.symbols) - 1
				}
				trailChar.currentSymbol = group.symbols[symbolIdx]
			}
		}

		group.currentCharIndex++
	}

	return true
}

// updateFinalWipePhase handles the final diagonal wipe
func (b *BeamsEffect) updateFinalWipePhase() {
	// In background mode, skip final wipe and go straight to hold
	if b.text == "" {
		b.phase = "hold"
		b.holdCounter = 0
		return
	}

	// Activate diagonal groups at specified speed
	for i := 0; i < b.finalWipeSpeed && b.currentDiag < len(b.diagonalGroups); i++ {
		for _, charIdx := range b.diagonalGroups[b.currentDiag] {
			char := &b.chars[charIdx]
			char.sceneActive = "brighten"
			char.sceneFrame = 0
			char.visible = true
			char.currentSymbol = char.original
		}
		b.currentDiag++
	}

	// Check if final wipe is complete
	if b.currentDiag >= len(b.diagonalGroups) {
		// Check if all characters have finished their brighten animation
		allComplete := true
		for i := range b.chars {
			char := &b.chars[i]
			if char.sceneActive == "brighten" {
				gradientLen := len(char.brightenGradient)
				framesPerStep := b.finalGradientFrames
				totalFrames := gradientLen * framesPerStep
				if char.sceneFrame < totalFrames {
					allComplete = false
					break
				}
			}
		}

		if allComplete {
			b.phase = "hold"
			b.holdCounter = 0
		}
	}
}

// updateHoldPhase handles the hold period after completion
func (b *BeamsEffect) updateHoldPhase() {
	b.holdCounter++

	// In background mode, loop immediately without hold
	holdDuration := b.holdFrames
	if b.text == "" {
		holdDuration = 0
	}

	// After hold period, reset the animation
	if b.holdCounter >= holdDuration {
		b.Reset()
	}
}

// updateCharacterAnimations updates all character animation scenes
func (b *BeamsEffect) updateCharacterAnimations() {
	for i := range b.chars {
		char := &b.chars[i]

		if !char.visible {
			continue
		}

		switch char.sceneActive {
		case "beam_row", "beam_column":
			// Beam gradient phase
			gradientLen := len(char.beamGradient)
			if gradientLen == 0 {
				break
			}

			framesPerStep := b.beamGradientFrames
			totalFrames := gradientLen * framesPerStep

			if char.sceneFrame < totalFrames {
				step := char.sceneFrame / framesPerStep
				if step >= gradientLen {
					step = gradientLen - 1
				}
				char.currentColor = char.beamGradient[step]
				char.sceneFrame++
			} else {
				// Move to fade phase
				char.sceneActive = "fade"
				char.sceneFrame = 0
			}

		case "fade":
			// Fade to dark
			fadeLen := len(char.fadeGradient)
			if fadeLen == 0 {
				char.sceneActive = ""
				char.currentSymbol = char.original
				break
			}

			if char.sceneFrame < fadeLen {
				char.currentColor = char.fadeGradient[char.sceneFrame]
				char.sceneFrame++
			} else {
				// Done fading, show original character dimly
				char.sceneActive = ""
				char.currentSymbol = char.original
			}

		case "brighten":
			// Brighten to final color
			gradientLen := len(char.brightenGradient)
			if gradientLen == 0 {
				break
			}

			framesPerStep := b.finalGradientFrames
			totalFrames := gradientLen * framesPerStep

			if char.sceneFrame < totalFrames {
				step := char.sceneFrame / framesPerStep
				if step >= gradientLen {
					step = gradientLen - 1
				}
				char.currentColor = char.brightenGradient[step]
				char.sceneFrame++
			}
		}
	}
}

// Render converts the beams effect to colored text output
func (b *BeamsEffect) Render() string {
	// Create empty canvas
	canvas := make([][]rune, b.height)
	colors := make([][]string, b.height)
	for i := range canvas {
		canvas[i] = make([]rune, b.width)
		colors[i] = make([]string, b.width)
		for j := range canvas[i] {
			canvas[i][j] = ' '
			colors[i][j] = ""
		}
	}

	// Draw characters
	for _, char := range b.chars {
		if !char.visible {
			continue
		}

		if char.y >= 0 && char.y < b.height && char.x >= 0 && char.x < b.width {
			canvas[char.y][char.x] = char.currentSymbol
			colors[char.y][char.x] = char.currentColor
		}
	}

	// Convert to colored string
	var lines []string
	for y := 0; y < b.height; y++ {
		var line strings.Builder
		for x := 0; x < b.width; x++ {
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

// Reset restarts the animation from the beginning
func (b *BeamsEffect) Reset() {
	b.phase = "beams"
	b.frameCount = 0
	b.beamDelayCount = 0
	b.currentDiag = 0
	b.holdCounter = 0

	// Reset all characters
	for i := range b.chars {
		b.chars[i].visible = false
		b.chars[i].sceneActive = ""
		b.chars[i].sceneFrame = 0
		b.chars[i].currentSymbol = b.chars[i].original
		b.chars[i].currentColor = ""
	}

	// Reset all groups
	for i := range b.rowGroups {
		b.rowGroups[i].nextCharCounter = 0
		b.rowGroups[i].currentCharIndex = 0
	}
	for i := range b.columnGroups {
		b.columnGroups[i].nextCharCounter = 0
		b.columnGroups[i].currentCharIndex = 0
	}
}

// parseHexColor converts hex color to RGB
func parseHexColor(hex string) [3]uint8 {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return [3]uint8{255, 255, 255}
	}

	var r, g, b uint8
	_, _ = fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return [3]uint8{r, g, b}
}

// formatHexColor converts RGB to hex color
func formatHexColor(rgb [3]uint8) string {
	return fmt.Sprintf("#%02x%02x%02x", rgb[0], rgb[1], rgb[2])
}

// Resize reinitializes the beams effect with new dimensions
func (b *BeamsEffect) Resize(width, height int) {
	b.width = width
	b.height = height
	b.chars = b.chars[:0]
	b.rowGroups = b.rowGroups[:0]
	b.columnGroups = b.columnGroups[:0]
	b.diagonalGroups = b.diagonalGroups[:0]
	b.init()
}

// Helper function to adjust brightness
func adjustColorBrightness(color string, factor float64) string {
	rgb := parseHexColor(color)
	r := uint8(math.Min(255, float64(rgb[0])*factor))
	g := uint8(math.Min(255, float64(rgb[1])*factor))
	b := uint8(math.Min(255, float64(rgb[2])*factor))
	return formatHexColor([3]uint8{r, g, b})
}
