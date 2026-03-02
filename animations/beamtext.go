package animations

import (
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss/v2"
)

// BeamTextEffect implements beams that travel across rows and columns, illuminating text
type BeamTextEffect struct {
	width   int
	height  int
	text    string
	auto    bool // Auto-size canvas to fit text
	display bool // Display mode: complete once and hold (true) or loop continuously (false)

	// Configuration
	beamRowSymbols       []rune
	beamColumnSymbols    []rune
	beamDelay            int
	beamRowSpeedRange    [2]int
	beamColumnSpeedRange [2]int
	beamGradientStops    []string
	beamGradientSteps    int
	beamGradientFrames   int
	finalGradientStops   []string
	finalGradientSteps   int
	finalGradientFrames  int
	finalWipeSpeed       int

	// Background beams effect
	backgroundBeams *BeamsEffect

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

	canvas  [][]rune
	colors  [][]string
	builder strings.Builder
}

// BeamTextConfig holds configuration for the beam text effect
type BeamTextConfig struct {
	Width                int
	Height               int
	Text                 string
	Auto                 bool // Auto-size canvas to fit text
	Display              bool // Display mode: complete once and hold (true) or loop continuously (false)
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

// NewBeamTextEffect creates a new beam text effect with given configuration
func NewBeamTextEffect(config BeamTextConfig) *BeamTextEffect {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Set defaults if not provided
	if len(config.BeamRowSymbols) == 0 {
		config.BeamRowSymbols = []rune{'▂', '▁', '_'}
	}
	if len(config.BeamColumnSymbols) == 0 {
		config.BeamColumnSymbols = []rune{'▌', '▍', '▎', '▏'}
	}
	if config.BeamDelay == 0 {
		config.BeamDelay = 2
	}
	if config.BeamRowSpeedRange[0] == 0 {
		config.BeamRowSpeedRange = [2]int{20, 80}
	}
	if config.BeamColumnSpeedRange[0] == 0 {
		config.BeamColumnSpeedRange = [2]int{15, 30}
	}
	if config.BeamGradientSteps == 0 {
		config.BeamGradientSteps = 5
	}
	if config.BeamGradientFrames == 0 {
		config.BeamGradientFrames = 1
	}
	if config.FinalGradientSteps == 0 {
		config.FinalGradientSteps = 8
	}
	if config.FinalGradientFrames == 0 {
		config.FinalGradientFrames = 1
	}
	if config.FinalWipeSpeed == 0 {
		config.FinalWipeSpeed = 3
	}

	// If auto-sizing, calculate dimensions from text
	width := config.Width
	height := config.Height
	if config.Auto && config.Text != "" {
		width, height = calculateTextDimensions(config.Text)
	}

	// Create background beams effect for visual depth
	beamsConfig := BeamsConfig{
		Width:                width,
		Height:               height,
		BeamRowSymbols:       config.BeamRowSymbols,
		BeamColumnSymbols:    config.BeamColumnSymbols,
		BeamDelay:            config.BeamDelay,
		BeamRowSpeedRange:    config.BeamRowSpeedRange,
		BeamColumnSpeedRange: config.BeamColumnSpeedRange,
		BeamGradientStops:    config.BeamGradientStops,
		BeamGradientSteps:    config.BeamGradientSteps,
		BeamGradientFrames:   config.BeamGradientFrames,
		FinalGradientStops:   config.FinalGradientStops,
		FinalGradientSteps:   config.FinalGradientSteps,
		FinalGradientFrames:  config.FinalGradientFrames,
		FinalWipeSpeed:       config.FinalWipeSpeed,
	}

	b := &BeamTextEffect{
		width:                width,
		height:               height,
		text:                 config.Text,
		auto:                 config.Auto,
		display:              config.Display,
		beamRowSymbols:       config.BeamRowSymbols,
		beamColumnSymbols:    config.BeamColumnSymbols,
		beamDelay:            config.BeamDelay,
		beamRowSpeedRange:    config.BeamRowSpeedRange,
		beamColumnSpeedRange: config.BeamColumnSpeedRange,
		beamGradientStops:    config.BeamGradientStops,
		beamGradientSteps:    config.BeamGradientSteps,
		beamGradientFrames:   config.BeamGradientFrames,
		finalGradientStops:   config.FinalGradientStops,
		finalGradientSteps:   config.FinalGradientSteps,
		finalGradientFrames:  config.FinalGradientFrames,
		finalWipeSpeed:       config.FinalWipeSpeed,
		backgroundBeams:      NewBeamsEffect(beamsConfig),
		phase:                "beams",
		frameCount:           0,
		beamDelayCount:       0,
		currentDiag:          0,
		holdFrames:           100,
		holdCounter:          0,
		rng:                  rng,
	}

	b.init()
	return b
}

// calculateTextDimensions returns the width and height needed to fit the text
func calculateTextDimensions(text string) (int, int) {
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

// init initializes characters and beam groups
func (b *BeamTextEffect) init() {
	b.initBuffers()
	b.initTextMode()

	// Create row groups
	b.createRowGroups()

	// Create column groups
	b.createColumnGroups()

	// Shuffle groups for random activation
	b.shuffleGroups()

	// Create diagonal groups for final wipe
	b.createDiagonalGroups()
}

func (b *BeamTextEffect) initBuffers() {
	b.canvas = make([][]rune, b.height)
	b.colors = make([][]string, b.height)
	for i := range b.canvas {
		b.canvas[i] = make([]rune, b.width)
		b.colors[i] = make([]string, b.width)
	}
	b.builder.Grow(b.width * b.height * 4)
}

func (b *BeamTextEffect) clearBuffers() {
	for i := range b.canvas {
		for j := range b.canvas[i] {
			b.canvas[i][j] = ' '
			b.colors[i][j] = ""
		}
	}
}

// initTextMode initializes with centered text (or left-aligned if auto-sized)
func (b *BeamTextEffect) initTextMode() {
	lines := strings.Split(b.text, "\n")

	// If auto-sizing, align text to left-top (no centering needed)
	// Otherwise, center the text in the given canvas
	var startY, blockStartX int
	if b.auto {
		startY = 0
		blockStartX = 0
	} else {
		// Calculate centered position for text block
		startY = (b.height - len(lines)) / 2
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
		blockStartX = (b.width - maxWidth) / 2
		if blockStartX < 0 {
			blockStartX = 0
		}
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

// createRowGroups creates beam groups for each row
func (b *BeamTextEffect) createRowGroups() {
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
func (b *BeamTextEffect) createColumnGroups() {
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
func (b *BeamTextEffect) shuffleGroups() {
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
func (b *BeamTextEffect) createDiagonalGroups() {
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
func (b *BeamTextEffect) createGradient(stops []string, steps int) []string {
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
func (b *BeamTextEffect) createFadeGradient(startColor string, steps int) []string {
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
func (b *BeamTextEffect) Update() {
	b.frameCount++

	// Update background beams for visual depth
	if b.backgroundBeams != nil {
		b.backgroundBeams.Update()
	}

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
func (b *BeamTextEffect) updateBeamsPhase() {
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
func (b *BeamTextEffect) updateGroup(group *BeamGroup) bool {
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
		symbolIndex := 0
		if len(group.symbols) > 0 {
			symbolIndex = 0
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
func (b *BeamTextEffect) updateFinalWipePhase() {
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
func (b *BeamTextEffect) updateHoldPhase() {
	b.holdCounter++

	// In display mode, stay at final state forever
	// In loop mode, reset after hold period
	if !b.display && b.holdCounter >= b.holdFrames {
		b.Reset()
	}
}

// updateCharacterAnimations updates all character animation scenes
func (b *BeamTextEffect) updateCharacterAnimations() {
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
func (b *BeamTextEffect) Render() string {
	b.clearBuffers()

	// First, render background beams directly from their character data (as base layer)
	if b.backgroundBeams != nil {
		// Access background beams' character array directly
		bgChars := getBeamsCharacters(b.backgroundBeams)
		for _, bgChar := range bgChars {
			if !bgChar.visible {
				continue
			}

			if bgChar.y >= 0 && bgChar.y < b.height && bgChar.x >= 0 && bgChar.x < b.width {
				// Only place background if position is empty
				if b.canvas[bgChar.y][bgChar.x] == ' ' {
					b.canvas[bgChar.y][bgChar.x] = bgChar.currentSymbol
					b.colors[bgChar.y][bgChar.x] = bgChar.currentColor
				}
			}
		}
	}

	// Draw text characters (these overlay the background)
	for _, char := range b.chars {
		if !char.visible {
			continue
		}

		if char.y >= 0 && char.y < b.height && char.x >= 0 && char.x < b.width {
			b.canvas[char.y][char.x] = char.currentSymbol
			b.colors[char.y][char.x] = char.currentColor
		}
	}

	// Convert to colored string
	b.builder.Reset()
	for y := 0; y < b.height; y++ {
		if y > 0 {
			b.builder.WriteRune('\n')
		}
		for x := 0; x < b.width; x++ {
			char := b.canvas[y][x]
			if char != ' ' && b.colors[y][x] != "" {
				// Characters with explicit colors
				styled := lipgloss.NewStyle().
					Foreground(lipgloss.Color(b.colors[y][x])).
					Render(string(char))
				b.builder.WriteString(styled)
			} else {
				b.builder.WriteRune(char)
			}
		}
	}

	return b.builder.String()
}

// getBeamsCharacters is a helper to access the background beams' character array
func getBeamsCharacters(beams *BeamsEffect) []BeamCharacter {
	if beams == nil {
		return nil
	}
	// Access the Chars field directly - it's exported in the BeamsEffect struct
	return beams.Chars
}

// Reset restarts the animation from the beginning
func (b *BeamTextEffect) Reset() {
	b.phase = "beams"
	b.frameCount = 0
	b.beamDelayCount = 0
	b.currentDiag = 0
	b.holdCounter = 0

	// Reset background beams
	if b.backgroundBeams != nil {
		b.backgroundBeams.Reset()
	}

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

// SetText updates the displayed text and reinitializes the beam animation.
func (b *BeamTextEffect) SetText(text string) {
	b.text = text
	b.chars = b.chars[:0]
	b.rowGroups = b.rowGroups[:0]
	b.columnGroups = b.columnGroups[:0]
	b.diagonalGroups = b.diagonalGroups[:0]
	b.init()
}

// Resize reinitializes the beam text effect with new dimensions
func (b *BeamTextEffect) Resize(width, height int) {
	b.width = width
	b.height = height
	b.initBuffers()

	// Resize background beams
	if b.backgroundBeams != nil {
		b.backgroundBeams.Resize(width, height)
	}

	b.chars = b.chars[:0]
	b.rowGroups = b.rowGroups[:0]
	b.columnGroups = b.columnGroups[:0]
	b.diagonalGroups = b.diagonalGroups[:0]
	b.init()
}
