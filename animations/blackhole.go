package animations

import (
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss/v2"
)

// BlackholeConfig holds the configuration for the Blackhole effect
type BlackholeConfig struct {
	Width               int
	Height              int
	Text                string
	BlackholeColor      string   // Border color for singularity
	StarColors          []string // Colors for post-explosion stars
	FinalGradientStops  []string // Gradient for final text state
	FinalGradientSteps  int      // Number of gradient steps
	FinalGradientDir    GradientDirection
	StaticGradientStops []string // Gradient for static ASCII
	StaticGradientDir   GradientDirection
	FormingFrames       int // Frames for border formation
	ConsumingFrames     int // Frames for consumption
	CollapsingFrames    int // Frames for border collapse
	ExplodingFrames     int // Frames for explosion scatter
	ReturningFrames     int // Frames for return to text
	StaticFrames        int // Frames to display static text initially
}

// BlackholeEffect represents the multi-phase blackhole animation
type BlackholeEffect struct {
	width  int
	height int
	text   string

	// Blackhole configuration
	blackholeColor      string
	starColors          []string
	finalGradientStops  []string
	finalGradientSteps  int
	finalGradientDir    GradientDirection
	staticGradientStops []string
	staticGradientDir   GradientDirection
	formingFrames       int
	consumingFrames     int
	collapsingFrames    int
	explodingFrames     int
	returningFrames     int
	staticFrames        int

	// Gradients
	finalGradient  []string
	staticGradient []string
	starGradient   []string

	// Character data
	chars           []BlackholeCharacter
	borderChars     []BorderCharacter
	centerX         float64
	centerY         float64
	blackholeRadius float64
	rng             *rand.Rand
	frameCount      int

	// Animation state
	phase              string // "static", "forming", "consuming", "collapsing", "exploding", "returning", "hold"
	consumeCounter     int    // Track consumption progress
	nextConsumeDelay   int    // Random delay before next character consumption
	currentConsumeWait int    // Current wait counter for consumption
	particleMode       bool   // True for particle mode (no text), false for text mode
}

// BlackholeCharacter represents a single character in the animation
type BlackholeCharacter struct {
	original     rune
	x            int     // Original position
	y            int     // Original position
	currentX     float64 // Current animated position
	currentY     float64 // Current animated position
	scatterX     float64 // Scatter position for explosion
	scatterY     float64 // Scatter position for explosion
	visible      bool
	currentColor string
	consumed     bool    // Has been consumed by blackhole
	consumeOrder int     // Order in which character is consumed
	scatterAngle float64 // Direction for explosion scatter
	scatterDist  float64 // Distance for explosion scatter
}

// BorderCharacter represents a character on the blackhole border
type BorderCharacter struct {
	angle          float64
	currentX       float64
	currentY       float64
	symbol         rune
	currentColor   string
	visible        bool
	formationDelay int // Delay before this char becomes visible
}

var unstableSymbols = []rune{'◦', '◎', '◉', '●'}

// NewBlackholeEffect creates a new Blackhole effect
func NewBlackholeEffect(config BlackholeConfig) *BlackholeEffect {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Set defaults
	if config.BlackholeColor == "" {
		config.BlackholeColor = "#ffffff"
	}
	if len(config.StarColors) == 0 {
		config.StarColors = []string{"#ffffff", "#ffd700", "#ff6b6b", "#4ecdc4", "#95e1d3", "#f38181"}
	}
	if len(config.FinalGradientStops) == 0 {
		config.FinalGradientStops = []string{"#6272a4", "#bd93f9", "#f8f8f2"}
	}
	if config.FinalGradientSteps == 0 {
		config.FinalGradientSteps = 12
	}
	if len(config.StaticGradientStops) == 0 {
		config.StaticGradientStops = config.StarColors
	}
	if config.FormingFrames == 0 {
		config.FormingFrames = 100
	}
	if config.ConsumingFrames == 0 {
		config.ConsumingFrames = 150
	}
	if config.CollapsingFrames == 0 {
		config.CollapsingFrames = 50
	}
	if config.ExplodingFrames == 0 {
		config.ExplodingFrames = 100
	}
	if config.ReturningFrames == 0 {
		config.ReturningFrames = 120
	}
	if config.StaticFrames == 0 {
		config.StaticFrames = 100
	}

	effect := &BlackholeEffect{
		width:               config.Width,
		height:              config.Height,
		text:                config.Text,
		blackholeColor:      config.BlackholeColor,
		starColors:          config.StarColors,
		finalGradientStops:  config.FinalGradientStops,
		finalGradientSteps:  config.FinalGradientSteps,
		finalGradientDir:    config.FinalGradientDir,
		staticGradientStops: config.StaticGradientStops,
		staticGradientDir:   config.StaticGradientDir,
		formingFrames:       config.FormingFrames,
		consumingFrames:     config.ConsumingFrames,
		collapsingFrames:    config.CollapsingFrames,
		explodingFrames:     config.ExplodingFrames,
		returningFrames:     config.ReturningFrames,
		staticFrames:        config.StaticFrames,
		rng:                 rng,
		phase:               "static",
		frameCount:          0,
		consumeCounter:      0,
	}

	effect.init()
	return effect
}

// init initializes the effect
func (e *BlackholeEffect) init() {
	e.centerX = float64(e.width) / 2
	e.centerY = float64(e.height) / 2

	// Determine if we're in particle mode (no text)
	e.particleMode = (e.text == "")

	// Calculate blackhole radius (30% for text mode, 60% for particle mode)
	smallestDim := float64(e.width)
	if float64(e.height) < smallestDim {
		smallestDim = float64(e.height)
	}
	radiusPercent := 0.3
	if e.particleMode {
		radiusPercent = 0.6
	}
	e.blackholeRadius = math.Max(smallestDim*radiusPercent, 3)

	// Create gradients
	e.finalGradient = e.createGradient(e.finalGradientStops, e.finalGradientSteps)
	e.staticGradient = e.createGradient(e.staticGradientStops, 100)
	e.starGradient = e.createGradient(e.starColors, 100)

	// Parse text and create characters (or generate random particles if no text)
	if e.particleMode {
		e.generateRandomParticles()
	} else {
		e.parseText()
	}

	// Create border characters
	e.createBorder()

	// Apply initial static gradient
	e.applyStaticGradient()

	// Generate scatter positions for explosion
	e.generateScatterPositions()
}

// parseText converts the text into positioned characters
func (e *BlackholeEffect) parseText() {
	lines := strings.Split(e.text, "\n")
	totalLines := len(lines)

	startY := (e.height - totalLines) / 2

	e.chars = make([]BlackholeCharacter, 0)

	for lineIdx, line := range lines {
		lineRunes := []rune(line)
		lineLen := len(lineRunes)
		startX := (e.width - lineLen) / 2

		for charIdx, char := range lineRunes {
			if char == ' ' || char == '\n' {
				continue
			}

			x := startX + charIdx
			y := startY + lineIdx

			character := BlackholeCharacter{
				original:     char,
				x:            x,
				y:            y,
				currentX:     float64(x),
				currentY:     float64(y),
				visible:      true,
				currentColor: e.staticGradient[0],
				consumed:     false,
				consumeOrder: -1,
			}

			e.chars = append(e.chars, character)
		}
	}

	// Assign random consumption order
	indices := make([]int, len(e.chars))
	for i := range indices {
		indices[i] = i
	}
	// Fisher-Yates shuffle
	for i := len(indices) - 1; i > 0; i-- {
		j := e.rng.Intn(i + 1)
		indices[i], indices[j] = indices[j], indices[i]
	}
	for order, idx := range indices {
		e.chars[idx].consumeOrder = order
	}
}

// generateRandomParticles creates random star particles across the screen for non-text mode
func (e *BlackholeEffect) generateRandomParticles() {
	// Generate 200-400 random star particles scattered across the screen
	numParticles := 200 + e.rng.Intn(200)

	// Star symbols to use for particles
	starSymbols := []rune{'*', '·', '•', '∗', '⋆', '✦', '✧', '✨', '✶', '✷', '✸', '✹'}

	e.chars = make([]BlackholeCharacter, 0, numParticles)

	for i := 0; i < numParticles; i++ {
		// Random position across entire screen
		x := e.rng.Intn(e.width)
		y := e.rng.Intn(e.height)

		// Random star symbol
		symbol := starSymbols[e.rng.Intn(len(starSymbols))]

		// Random color from star gradient
		color := e.staticGradient[e.rng.Intn(len(e.staticGradient))]

		character := BlackholeCharacter{
			original:     symbol,
			x:            x,
			y:            y,
			currentX:     float64(x),
			currentY:     float64(y),
			visible:      true,
			currentColor: color,
			consumed:     false,
			consumeOrder: i, // Sequential order for smooth consumption
		}

		e.chars = append(e.chars, character)
	}
}

// createBorder creates the circular border around the blackhole
func (e *BlackholeEffect) createBorder() {
	// Create approximately 3× radius border characters (matching TTE)
	numBorderChars := int(e.blackholeRadius * 3)
	if numBorderChars < 20 {
		numBorderChars = 20
	}
	e.borderChars = make([]BorderCharacter, numBorderChars)

	// Calculate staggered formation delay (matching TTE)
	formationDelayIncrement := 100 / numBorderChars
	if formationDelayIncrement < 10 {
		formationDelayIncrement = 10
	}

	for i := range e.borderChars {
		angle := (float64(i) / float64(numBorderChars)) * 2 * math.Pi
		e.borderChars[i] = BorderCharacter{
			angle:          angle,
			currentX:       e.centerX + e.blackholeRadius*math.Cos(angle),
			currentY:       e.centerY + e.blackholeRadius*math.Sin(angle),
			symbol:         '●',
			currentColor:   e.blackholeColor,
			visible:        false,
			formationDelay: i * formationDelayIncrement,
		}
	}
}

// generateScatterPositions creates random scatter positions for explosion
func (e *BlackholeEffect) generateScatterPositions() {
	for i := range e.chars {
		// Random angle for scatter direction
		e.chars[i].scatterAngle = e.rng.Float64() * 2 * math.Pi

		// Random distance (50-150% of blackhole radius)
		e.chars[i].scatterDist = e.blackholeRadius * (0.5 + e.rng.Float64())

		// Calculate scatter position
		e.chars[i].scatterX = e.centerX + math.Cos(e.chars[i].scatterAngle)*e.chars[i].scatterDist
		e.chars[i].scatterY = e.centerY + math.Sin(e.chars[i].scatterAngle)*e.chars[i].scatterDist

		// Clamp to canvas
		if e.chars[i].scatterX < 0 {
			e.chars[i].scatterX = 0
		}
		if e.chars[i].scatterX >= float64(e.width) {
			e.chars[i].scatterX = float64(e.width - 1)
		}
		if e.chars[i].scatterY < 0 {
			e.chars[i].scatterY = 0
		}
		if e.chars[i].scatterY >= float64(e.height) {
			e.chars[i].scatterY = float64(e.height - 1)
		}
	}
}

// applyStaticGradient applies gradient to static ASCII (same as ringtext)
func (e *BlackholeEffect) applyStaticGradient() {
	if len(e.chars) == 0 || len(e.staticGradient) == 0 {
		return
	}

	minX, maxX := e.width, 0
	minY, maxY := e.height, 0

	for i := range e.chars {
		if e.chars[i].x < minX {
			minX = e.chars[i].x
		}
		if e.chars[i].x > maxX {
			maxX = e.chars[i].x
		}
		if e.chars[i].y < minY {
			minY = e.chars[i].y
		}
		if e.chars[i].y > maxY {
			maxY = e.chars[i].y
		}
	}

	textWidth := float64(maxX - minX)
	textHeight := float64(maxY - minY)
	if textWidth == 0 {
		textWidth = 1
	}
	if textHeight == 0 {
		textHeight = 1
	}

	for i := range e.chars {
		var gradientPos float64

		switch e.staticGradientDir {
		case GradientHorizontal:
			gradientPos = float64(e.chars[i].x-minX) / textWidth
		case GradientVertical:
			gradientPos = float64(e.chars[i].y-minY) / textHeight
		case GradientDiagonal:
			xPos := float64(e.chars[i].x-minX) / textWidth
			yPos := float64(e.chars[i].y-minY) / textHeight
			gradientPos = (xPos + yPos) / 2.0
		case GradientRadial:
			dx := float64(e.chars[i].x) - e.centerX
			dy := float64(e.chars[i].y) - e.centerY
			maxDist := math.Sqrt(textWidth*textWidth+textHeight*textHeight) / 2.0
			dist := math.Sqrt(dx*dx + dy*dy)
			gradientPos = math.Min(dist/maxDist, 1.0)
		default:
			gradientPos = 0
		}

		gradientPos = math.Max(0, math.Min(1, gradientPos))
		gradientIndex := int(gradientPos * float64(len(e.staticGradient)-1))
		if gradientIndex >= len(e.staticGradient) {
			gradientIndex = len(e.staticGradient) - 1
		}

		e.chars[i].currentColor = e.staticGradient[gradientIndex]
	}
}

// Update advances the animation by one frame
func (e *BlackholeEffect) Update() {
	e.frameCount++

	// Rotate border continuously for swirling effect (matching TTE speed of 0.2)
	rotationSpeed := 0.2 // radians per frame
	for i := range e.borderChars {
		e.borderChars[i].angle += rotationSpeed
		// Keep angle in 0-2π range
		if e.borderChars[i].angle > 2*math.Pi {
			e.borderChars[i].angle -= 2 * math.Pi
		}
	}

	switch e.phase {
	case "static":
		if e.frameCount >= e.staticFrames {
			e.phase = "forming"
			e.frameCount = 0
		}

	case "forming":
		// Update border positions based on current angles
		for i := range e.borderChars {
			e.borderChars[i].currentX = e.centerX + e.blackholeRadius*math.Cos(e.borderChars[i].angle)
			e.borderChars[i].currentY = e.centerY + e.blackholeRadius*math.Sin(e.borderChars[i].angle)
		}

		// Staggered formation - characters appear based on individual delays
		for i := range e.borderChars {
			if e.frameCount >= e.borderChars[i].formationDelay {
				e.borderChars[i].visible = true
			}
		}

		if e.frameCount >= e.formingFrames {
			e.phase = "consuming"
			e.frameCount = 0
			e.consumeCounter = 0
			e.nextConsumeDelay = e.rng.Intn(10) // Random delay before first consumption
			e.currentConsumeWait = 0
		}

	case "consuming":
		progress := float64(e.frameCount) / float64(e.consumingFrames)
		if progress > 1.0 {
			progress = 1.0
		}

		// Update border positions based on current angles (swirling)
		for i := range e.borderChars {
			e.borderChars[i].currentX = e.centerX + e.blackholeRadius*math.Cos(e.borderChars[i].angle)
			e.borderChars[i].currentY = e.centerY + e.blackholeRadius*math.Sin(e.borderChars[i].angle)
		}

		// Consume multiple characters per frame for dramatic dissolution
		// Start slow, accelerate dramatically as progress increases
		charsPerFrame := 1 + int(progress*progress*12) // 1-13 characters per frame (quadratic acceleration)
		for i := 0; i < charsPerFrame && e.consumeCounter < len(e.chars); i++ {
			// Find next character to consume
			for j := range e.chars {
				if e.chars[j].consumeOrder == e.consumeCounter && !e.chars[j].consumed {
					e.chars[j].consumed = true
					break
				}
			}
			e.consumeCounter++
		}

		// Move consumed characters toward center
		if e.particleMode {
			// Particle mode: Use enhanced spiral/sucking motion
			for i := range e.chars {
				if e.chars[i].consumed {
					// Find when this character was consumed
					consumeProgress := float64(e.chars[i].consumeOrder) / float64(len(e.chars))
					timeConsumed := progress - consumeProgress
					if timeConsumed < 0 {
						timeConsumed = 0
					}
					if timeConsumed > 1 {
						timeConsumed = 1
					}

					// Super-exponential ease toward center (stronger gravity effect)
					easedProgress := timeConsumed * timeConsumed * timeConsumed // Cubic for stronger pull

					// Calculate spiral path toward center
					startX := float64(e.chars[i].x)
					startY := float64(e.chars[i].y)

					// Add spiral/swirl motion as particle approaches center
					angle := math.Atan2(e.centerY-startY, e.centerX-startX)
					spiralAngle := angle + easedProgress*math.Pi*4 // 2 full rotations on the way in

					// Distance decreases with eased progress
					currentDist := math.Sqrt(math.Pow(startX-e.centerX, 2)+math.Pow(startY-e.centerY, 2)) * (1 - easedProgress)

					// Calculate position along spiral
					e.chars[i].currentX = e.centerX + currentDist*math.Cos(spiralAngle)
					e.chars[i].currentY = e.centerY + currentDist*math.Sin(spiralAngle)

					// Fade to black as approaching center (more dramatic)
					dist := math.Sqrt(math.Pow(e.chars[i].currentX-e.centerX, 2) + math.Pow(e.chars[i].currentY-e.centerY, 2))
					brightness := dist / e.blackholeRadius
					if brightness < 0.5 {
						e.chars[i].visible = false
					}
				}
			}
		} else {
			// Text mode: Use original Bézier curve logic
			for i := range e.chars {
				if e.chars[i].consumed {
					// Exponential ease toward center (gravity effect)
					easedProgress := e.easeInExpo(progress)

					// Bézier curve toward center
					startX := float64(e.chars[i].x)
					startY := float64(e.chars[i].y)

					// Control point for curve (offset perpendicular to direction)
					dx := e.centerX - startX
					dy := e.centerY - startY
					controlX := startX + dx*0.5 + dy*0.3
					controlY := startY + dy*0.5 - dx*0.3

					// Quadratic Bézier curve
					t := easedProgress
					e.chars[i].currentX = (1-t)*(1-t)*startX + 2*(1-t)*t*controlX + t*t*e.centerX
					e.chars[i].currentY = (1-t)*(1-t)*startY + 2*(1-t)*t*controlY + t*t*e.centerY

					// Fade to black as approaching center
					dist := math.Sqrt(math.Pow(e.chars[i].currentX-e.centerX, 2) + math.Pow(e.chars[i].currentY-e.centerY, 2))
					brightness := dist / e.blackholeRadius
					if brightness < 0.3 {
						e.chars[i].visible = false
					}
				}
			}
		}

		if e.consumeCounter >= len(e.chars) {
			// All characters consumed, move to collapsing
			e.phase = "collapsing"
			e.frameCount = 0
		}

	case "collapsing":
		progress := float64(e.frameCount) / float64(e.collapsingFrames)
		if progress > 1.0 {
			progress = 1.0
		}

		// Contract border toward center
		currentRadius := e.blackholeRadius * (1.0 - progress)

		for i := range e.borderChars {
			e.borderChars[i].currentX = e.centerX + currentRadius*math.Cos(e.borderChars[i].angle)
			e.borderChars[i].currentY = e.centerY + currentRadius*math.Sin(e.borderChars[i].angle)

			// Random unstable symbols
			if e.rng.Float64() < 0.1 {
				e.borderChars[i].symbol = unstableSymbols[e.rng.Intn(len(unstableSymbols))]
			}

			// Random colors from star palette
			if e.rng.Float64() < 0.05 {
				e.borderChars[i].currentColor = e.starColors[e.rng.Intn(len(e.starColors))]
			}
		}

		if e.frameCount >= e.collapsingFrames {
			e.phase = "exploding"
			e.frameCount = 0
			// Hide border
			for i := range e.borderChars {
				e.borderChars[i].visible = false
			}
			// Reset character visibility
			for i := range e.chars {
				e.chars[i].visible = true
			}
		}

	case "exploding":
		progress := float64(e.frameCount) / float64(e.explodingFrames)
		if progress > 1.0 {
			progress = 1.0
		}

		if e.particleMode {
			// Particle mode: Use enhanced explosion (easeOutQuart, 1.5x scatter)
			easedProgress := e.easeOutQuart(progress)

			for i := range e.chars {
				// Calculate scatter distance (particles fly further out - 150% of original scatter distance)
				scatterMultiplier := 1.5
				finalScatterX := e.centerX + (e.chars[i].scatterX-e.centerX)*scatterMultiplier
				finalScatterY := e.centerY + (e.chars[i].scatterY-e.centerY)*scatterMultiplier

				// Clamp to screen bounds
				if finalScatterX < 0 {
					finalScatterX = 0
				}
				if finalScatterX >= float64(e.width) {
					finalScatterX = float64(e.width - 1)
				}
				if finalScatterY < 0 {
					finalScatterY = 0
				}
				if finalScatterY >= float64(e.height) {
					finalScatterY = float64(e.height - 1)
				}

				// Scatter from center to extended scatter position with explosive motion
				e.chars[i].currentX = e.centerX + (finalScatterX-e.centerX)*easedProgress
				e.chars[i].currentY = e.centerY + (finalScatterY-e.centerY)*easedProgress

				// Rapid color cycling through star colors for explosive effect
				colorIndex := int((progress*3 + float64(i)*0.1) * float64(len(e.starGradient)))
				colorIndex = colorIndex % len(e.starGradient)
				e.chars[i].currentColor = e.starGradient[colorIndex]
			}
		} else {
			// Text mode: Use original explosion (easeOutExpo, 1.0x scatter)
			easedProgress := e.easeOutExpo(progress)

			for i := range e.chars {
				// Scatter from center to scatter position
				e.chars[i].currentX = e.centerX + (e.chars[i].scatterX-e.centerX)*easedProgress
				e.chars[i].currentY = e.centerY + (e.chars[i].scatterY-e.centerY)*easedProgress

				// Cycle through star colors
				colorIndex := int((progress + float64(i)*0.1) * float64(len(e.starGradient)))
				colorIndex = colorIndex % len(e.starGradient)
				e.chars[i].currentColor = e.starGradient[colorIndex]
			}
		}

		if e.frameCount >= e.explodingFrames {
			e.phase = "returning"
			e.frameCount = 0
		}

	case "returning":
		progress := float64(e.frameCount) / float64(e.returningFrames)
		if progress > 1.0 {
			progress = 1.0
		}

		easedProgress := e.easeInOutCubic(progress)

		for i := range e.chars {
			// Return from scatter position to original
			e.chars[i].currentX = e.chars[i].scatterX + (float64(e.chars[i].x)-e.chars[i].scatterX)*easedProgress
			e.chars[i].currentY = e.chars[i].scatterY + (float64(e.chars[i].y)-e.chars[i].scatterY)*easedProgress

			// Transition to final gradient color
			gradientIndex := int(easedProgress * float64(len(e.finalGradient)-1))
			if gradientIndex >= len(e.finalGradient) {
				gradientIndex = len(e.finalGradient) - 1
			}
			e.chars[i].currentColor = e.finalGradient[gradientIndex]
		}

		if e.frameCount >= e.returningFrames {
			e.phase = "hold"
			e.frameCount = 0
		}

	case "hold":
		if e.frameCount >= 60 {
			e.Reset()
		}
	}
}

// Render returns the current frame as a colored string
func (e *BlackholeEffect) Render() string {
	buffer := make([][]rune, e.height)
	colors := make([][]string, e.height)
	for i := range buffer {
		buffer[i] = make([]rune, e.width)
		colors[i] = make([]string, e.width)
		for j := range buffer[i] {
			buffer[i][j] = ' '
			colors[i][j] = ""
		}
	}

	// Draw characters
	for _, char := range e.chars {
		if !char.visible {
			continue
		}

		x := int(math.Round(char.currentX))
		y := int(math.Round(char.currentY))

		if x >= 0 && x < e.width && y >= 0 && y < e.height {
			buffer[y][x] = char.original
			colors[y][x] = char.currentColor
		}
	}

	// Border animation removed - doesn't work well with ASCII characters
	/*
		// Draw border
		for _, borderChar := range e.borderChars {
			if !borderChar.visible {
				continue
			}

			x := int(math.Round(borderChar.currentX))
			y := int(math.Round(borderChar.currentY))

			if x >= 0 && x < e.width && y >= 0 && y < e.height {
				buffer[y][x] = borderChar.symbol
				colors[y][x] = borderChar.currentColor
			}
		}
	*/

	// Build output (line-by-line like other effects)
	var lines []string
	for y := 0; y < e.height; y++ {
		var line strings.Builder
		for x := 0; x < e.width; x++ {
			char := buffer[y][x]
			color := colors[y][x]

			if color != "" && char != ' ' {
				styled := lipgloss.NewStyle().
					Foreground(lipgloss.Color(color)).
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

// Reset restarts the animation
func (e *BlackholeEffect) Reset() {
	e.phase = "static"
	e.frameCount = 0
	e.consumeCounter = 0
	e.nextConsumeDelay = 0
	e.currentConsumeWait = 0

	// Reset characters
	for i := range e.chars {
		e.chars[i].currentX = float64(e.chars[i].x)
		e.chars[i].currentY = float64(e.chars[i].y)
		e.chars[i].visible = true
		e.chars[i].consumed = false
	}

	// Reset border
	for i := range e.borderChars {
		e.borderChars[i].visible = false
		e.borderChars[i].symbol = '●'
		e.borderChars[i].currentColor = e.blackholeColor
		e.borderChars[i].currentX = e.centerX + e.blackholeRadius*math.Cos(e.borderChars[i].angle)
		e.borderChars[i].currentY = e.centerY + e.blackholeRadius*math.Sin(e.borderChars[i].angle)
	}

	// Reapply static gradient
	e.applyStaticGradient()

	// Regenerate scatter positions
	e.generateScatterPositions()
}

// SetText updates the displayed text and reinitializes the blackhole animation.
func (e *BlackholeEffect) SetText(text string) {
	e.text = text
	e.particleMode = text == ""
	e.init()
}

// createGradient creates a gradient between color stops
func (e *BlackholeEffect) createGradient(stops []string, steps int) []string {
	if len(stops) == 0 {
		return []string{"#ffffff"}
	}
	if len(stops) == 1 {
		return []string{stops[0]}
	}

	gradient := make([]string, 0)
	stepsPerSegment := steps / (len(stops) - 1)

	for i := 0; i < len(stops)-1; i++ {
		startColor := parseHexColor(stops[i])
		endColor := parseHexColor(stops[i+1])

		for j := 0; j < stepsPerSegment; j++ {
			t := float64(j) / float64(stepsPerSegment)
			r := uint8(float64(startColor[0]) + (float64(endColor[0])-float64(startColor[0]))*t)
			g := uint8(float64(startColor[1]) + (float64(endColor[1])-float64(startColor[1]))*t)
			b := uint8(float64(startColor[2]) + (float64(endColor[2])-float64(startColor[2]))*t)
			gradient = append(gradient, formatHexColor([3]uint8{r, g, b}))
		}
	}

	gradient = append(gradient, stops[len(stops)-1])
	return gradient
}

// Easing functions
func (e *BlackholeEffect) easeInExpo(t float64) float64 {
	if t == 0 {
		return 0
	}
	return math.Pow(2, 10*(t-1))
}

func (e *BlackholeEffect) easeOutExpo(t float64) float64 {
	if t == 1 {
		return 1
	}
	return 1 - math.Pow(2, -10*t)
}

func (e *BlackholeEffect) easeInOutCubic(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	return 1 - math.Pow(-2*t+2, 3)/2
}

func (e *BlackholeEffect) easeOutQuart(t float64) float64 {
	return 1 - math.Pow(1-t, 4)
}
