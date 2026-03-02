package animations

import (
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss/v2"
)

// GradientDirection specifies the direction of gradient application
type GradientDirection int

const (
	GradientHorizontal GradientDirection = iota // Left to right
	GradientVertical                            // Top to bottom
	GradientDiagonal                            // Top-left to bottom-right
	GradientRadial                              // Center outward
)

// RingTextConfig holds the configuration for the RingText effect
type RingTextConfig struct {
	Width               int
	Height              int
	Text                string
	RingColors          []string          // Colors for each ring
	RingGap             float64           // Distance between rings as a percent of smallest dimension
	SpinSpeedRange      [2]float64        // Speed range for rotation (min, max radians per frame)
	SpinDuration        int               // Frames to spin on rings
	DisperseDuration    int               // Frames to stay in dispersed state
	SpinDisperseCycles  int               // Number of spin/disperse cycles before returning
	TransitionFrames    int               // Frames for transitions between states
	StaticFrames        int               // Frames to display static text initially
	FinalGradientStops  []string          // Gradient for final text state
	FinalGradientSteps  int               // Number of gradient steps
	StaticGradientStops []string          // Gradient for static ASCII presentation
	StaticGradientDir   GradientDirection // Direction of static gradient
}

// RingTextEffect represents the multi-phase ring text animation
type RingTextEffect struct {
	width  int
	height int
	text   string

	// Ring configuration
	ringColors         []string
	ringGap            float64
	spinSpeedRange     [2]float64
	spinDuration       int
	disperseDuration   int
	spinDisperseCycles int
	transitionFrames   int
	staticFrames       int

	// Gradient configuration
	finalGradientStops  []string
	finalGradientSteps  int
	finalGradient       []string
	staticGradientStops []string
	staticGradientDir   GradientDirection
	staticGradient      []string         // Pre-computed static gradient
	ringGradients       map[int][]string // 8-step gradients for each ring

	// Character data
	chars      []RingTextCharacter
	rings      []Ring
	centerX    float64
	centerY    float64
	rng        *rand.Rand
	frameCount int

	// Animation state
	phase        string // "static", "transition_to_disperse", "disperse", "transition_to_spin", "spin", "return_to_text", "hold"
	currentCycle int    // Current spin/disperse cycle
}

// RingTextCharacter represents a single character in the animation
type RingTextCharacter struct {
	original       rune
	x              int     // Original position
	y              int     // Original position
	currentX       float64 // Current animated position
	currentY       float64 // Current animated position
	disperseRadius float64 // Radius for circular disperse position (larger circles)
	disperseAngle  float64 // Angle for circular disperse position
	visible        bool
	currentColor   string
	ringIndex      int     // Which ring this character belongs to
	angleOnRing    float64 // Current angle on the ring (in radians)
	rotationSpeed  float64 // Individual rotation speed
	clockwise      bool    // Rotation direction
}

// Ring represents a circular ring of positions
type Ring struct {
	radius           float64
	color            string
	rotationSpeed    float64
	clockwise        bool
	characterIndices []int // Indices of characters on this ring
}

// NewRingTextEffect creates a new RingText effect
func NewRingTextEffect(config RingTextConfig) *RingTextEffect {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Set defaults
	if config.RingGap == 0 {
		config.RingGap = 0.1
	}
	if config.SpinSpeedRange[0] == 0 && config.SpinSpeedRange[1] == 0 {
		config.SpinSpeedRange = [2]float64{0.025, 0.075} // Min-max range like TTE
	}
	if config.SpinDuration == 0 {
		config.SpinDuration = 200
	}
	if config.DisperseDuration == 0 {
		config.DisperseDuration = 200
	}
	if config.SpinDisperseCycles == 0 {
		config.SpinDisperseCycles = 3 // Like TTE default
	}
	if config.TransitionFrames == 0 {
		config.TransitionFrames = 100
	}
	if config.StaticFrames == 0 {
		config.StaticFrames = 100
	}
	if config.FinalGradientSteps == 0 {
		config.FinalGradientSteps = 12
	}
	if len(config.RingColors) == 0 {
		config.RingColors = []string{"#bd93f9", "#ff79c6", "#f1fa8c"} // Default: Dracula purple, pink, yellow
	}
	if len(config.FinalGradientStops) == 0 {
		config.FinalGradientStops = []string{"#6272a4", "#bd93f9", "#f8f8f2"}
	}
	if len(config.StaticGradientStops) == 0 {
		// Default: use ring colors for static gradient
		config.StaticGradientStops = config.RingColors
	}

	effect := &RingTextEffect{
		width:               config.Width,
		height:              config.Height,
		text:                config.Text,
		ringColors:          config.RingColors,
		ringGap:             config.RingGap,
		spinSpeedRange:      config.SpinSpeedRange,
		spinDuration:        config.SpinDuration,
		disperseDuration:    config.DisperseDuration,
		spinDisperseCycles:  config.SpinDisperseCycles,
		transitionFrames:    config.TransitionFrames,
		staticFrames:        config.StaticFrames,
		finalGradientStops:  config.FinalGradientStops,
		finalGradientSteps:  config.FinalGradientSteps,
		staticGradientStops: config.StaticGradientStops,
		staticGradientDir:   config.StaticGradientDir,
		rng:                 rng,
		phase:               "static",
		frameCount:          0,
		currentCycle:        0,
		ringGradients:       make(map[int][]string),
	}

	effect.init()
	return effect
}

// init initializes the effect
func (e *RingTextEffect) init() {
	e.centerX = float64(e.width) / 2
	e.centerY = float64(e.height) / 2

	// Create gradient for final state
	e.finalGradient = e.createGradient(e.finalGradientStops, e.finalGradientSteps)

	// Create gradient for static ASCII presentation (higher resolution for smooth transitions)
	e.staticGradient = e.createGradient(e.staticGradientStops, 100)

	// Parse text and create characters
	e.parseText()

	// Create rings and assign characters
	e.createRings()

	// Create 8-step gradients for each ring (for transitions)
	for i := range e.rings {
		// Gradient from final color to ring color
		e.ringGradients[i] = e.createGradient([]string{e.finalGradient[0], e.rings[i].color}, 8)
	}

	// Generate random disperse positions for all characters
	e.generateDispersePositions()

	// Apply initial static gradient colors to all characters
	e.applyStaticGradient()
}

// parseText converts the text into positioned characters
func (e *RingTextEffect) parseText() {
	lines := strings.Split(e.text, "\n")
	totalLines := len(lines)

	// Calculate starting Y position to center text vertically
	startY := (e.height - totalLines) / 2

	e.chars = make([]RingTextCharacter, 0)

	for lineIdx, line := range lines {
		lineRunes := []rune(line)
		lineLen := len(lineRunes)

		// Calculate starting X position to center line horizontally
		startX := (e.width - lineLen) / 2

		for charIdx, char := range lineRunes {
			if char == ' ' || char == '\n' {
				continue // Skip spaces and newlines
			}

			x := startX + charIdx
			y := startY + lineIdx

			character := RingTextCharacter{
				original:     char,
				x:            x,
				y:            y,
				currentX:     float64(x),
				currentY:     float64(y),
				visible:      true,
				currentColor: e.finalGradient[0], // Start with first gradient color
			}

			e.chars = append(e.chars, character)
		}
	}
}

// createRings creates concentric rings and assigns characters to them
func (e *RingTextEffect) createRings() {
	if len(e.chars) == 0 {
		return
	}

	// Calculate maximum radius based on smallest dimension
	smallestDim := float64(e.width)
	if float64(e.height) < smallestDim {
		smallestDim = float64(e.height)
	}

	ringGapPixels := smallestDim * e.ringGap
	maxRadius := smallestDim / 2

	// Create rings
	e.rings = make([]Ring, 0)
	for radius := ringGapPixels; radius < maxRadius; radius += ringGapPixels {
		colorIndex := len(e.rings) % len(e.ringColors)
		clockwise := len(e.rings)%2 == 0

		// Random speed from range (like TTE)
		speed := e.spinSpeedRange[0] + e.rng.Float64()*(e.spinSpeedRange[1]-e.spinSpeedRange[0])

		ring := Ring{
			radius:           radius,
			color:            e.ringColors[colorIndex],
			rotationSpeed:    speed,
			clockwise:        clockwise,
			characterIndices: make([]int, 0),
		}

		e.rings = append(e.rings, ring)
	}

	// Assign characters to rings evenly
	if len(e.rings) > 0 {
		for i := range e.chars {
			ringIndex := i % len(e.rings)
			e.chars[i].ringIndex = ringIndex
			e.chars[i].clockwise = e.rings[ringIndex].clockwise
			e.chars[i].rotationSpeed = e.rings[ringIndex].rotationSpeed

			// Calculate initial angle on ring based on character's original position
			dx := float64(e.chars[i].x) - e.centerX
			dy := float64(e.chars[i].y) - e.centerY
			e.chars[i].angleOnRing = math.Atan2(dy, dx)

			e.rings[ringIndex].characterIndices = append(e.rings[ringIndex].characterIndices, i)
		}
	}
}

// generateDispersePositions creates circular scatter positions (larger circles)
func (e *RingTextEffect) generateDispersePositions() {
	if len(e.rings) == 0 {
		return
	}

	// Instead of rectangular scatter, place characters on larger circles
	// This creates a circular vortex effect instead of box explosion
	for i := range e.chars {
		ring := &e.rings[e.chars[i].ringIndex]

		// Scatter radius: 2-3x the final ring radius (creates expanding/contracting vortex)
		scatterRadiusMultiplier := 2.0 + e.rng.Float64() // 2.0x to 3.0x final radius
		e.chars[i].disperseRadius = ring.radius * scatterRadiusMultiplier

		// Use the character's ring angle, but add some randomness
		// This spreads characters around the circle while maintaining circular shape
		angleVariation := (e.rng.Float64() - 0.5) * math.Pi / 4 // ±45 degrees
		e.chars[i].disperseAngle = e.chars[i].angleOnRing + angleVariation
	}
}

// Update advances the animation by one frame
func (e *RingTextEffect) Update() {
	e.frameCount++

	switch e.phase {
	case "static":
		if e.frameCount >= e.staticFrames {
			e.phase = "swirl_to_rings"
			e.frameCount = 0
		}

	case "swirl_to_rings":
		// CIRCULAR VORTEX MOTION - no linear interpolation, only orbital motion
		// Characters spiral from ASCII → outer circles → inner ring positions
		// All motion is calculated using polar coordinates (radius + angle)

		totalDuration := float64(e.disperseDuration + e.transitionFrames*2)
		progress := float64(e.frameCount) / totalDuration
		if progress > 1.0 {
			progress = 1.0
		}

		for i := range e.chars {
			ring := &e.rings[e.chars[i].ringIndex]
			ringGradient := e.ringGradients[e.chars[i].ringIndex]

			// Start position (ASCII)
			startX := float64(e.chars[i].x)
			startY := float64(e.chars[i].y)

			// Calculate radius from ASCII position to center
			startDeltaX := startX - e.centerX
			startDeltaY := startY - e.centerY
			startRadius := math.Sqrt(startDeltaX*startDeltaX + startDeltaY*startDeltaY)
			if startRadius < 0.1 {
				startRadius = 0.1 // Avoid division by zero
			}
			startAngle := math.Atan2(startDeltaY, startDeltaX)

			// Target is the final ring position
			targetRadius := ring.radius
			targetAngle := e.chars[i].angleOnRing

			// Disperse position is on a larger circle
			disperseRadius := e.chars[i].disperseRadius
			disperseAngle := e.chars[i].disperseAngle

			var currentRadius float64
			var currentAngle float64
			var colorIdx int

			// PHASE 1 (0-0.25): Spiral outward from ASCII to outer circles
			// PHASE 2 (0.25-0.75): Swirl on outer circles while contracting
			// PHASE 3 (0.75-1.0): Final spiral inward to exact ring positions

			if progress < 0.25 {
				// Expanding vortex: ASCII → outer circles
				expandProgress := progress / 0.25
				easedExpand := e.easeInOutCubic(expandProgress)

				// Radius expands from start to disperse
				currentRadius = startRadius + (disperseRadius-startRadius)*easedExpand

				// Angle spirals from start to disperse (creates vortex motion)
				currentAngle = startAngle + (disperseAngle-startAngle)*easedExpand

				colorIdx = int(easedExpand * float64(len(ringGradient)-1))

			} else if progress < 0.75 {
				// Swirling vortex: orbit on large circles while contracting toward rings
				swirlProgress := (progress - 0.25) / 0.5
				easedSwirl := 1 - math.Pow(1-swirlProgress, 2) // quadratic ease-out

				// Radius contracts from disperse to target
				currentRadius = disperseRadius + (targetRadius-disperseRadius)*easedSwirl

				// Angle continues rotating (continuous swirl)
				// Use disperseAngle as base and add continuous rotation
				rotationAmount := swirlProgress * math.Pi * 2 // Full rotation during swirl
				if e.chars[i].clockwise {
					currentAngle = disperseAngle + rotationAmount
				} else {
					currentAngle = disperseAngle - rotationAmount
				}

				colorIdx = len(ringGradient) - 1

			} else {
				// Contracting vortex: final spiral to exact ring positions
				tightenProgress := (progress - 0.75) / 0.25
				easedTighten := e.easeInOutCubic(tightenProgress)

				// Calculate where we were at 75% mark
				radius75 := disperseRadius + (targetRadius-disperseRadius)*0.99 // Almost at target
				angle75Progress := 0.5
				rotationAmount75 := angle75Progress * math.Pi * 2
				var angle75 float64
				if e.chars[i].clockwise {
					angle75 = disperseAngle + rotationAmount75
				} else {
					angle75 = disperseAngle - rotationAmount75
				}

				// Final tighten to exact positions
				currentRadius = radius75 + (targetRadius-radius75)*easedTighten
				currentAngle = angle75 + (targetAngle-angle75)*easedTighten

				colorIdx = len(ringGradient) - 1
			}

			// Convert polar coordinates (radius, angle) to Cartesian (x, y)
			e.chars[i].currentX = e.centerX + currentRadius*math.Cos(currentAngle)
			e.chars[i].currentY = e.centerY + currentRadius*math.Sin(currentAngle)

			// Apply color gradient
			if colorIdx >= len(ringGradient) {
				colorIdx = len(ringGradient) - 1
			}
			if colorIdx < 0 {
				colorIdx = 0
			}
			e.chars[i].currentColor = ringGradient[colorIdx]

			// Update the angle for next iteration
			e.chars[i].angleOnRing = currentAngle
		}

		if progress >= 1.0 {
			e.phase = "spin"
			e.frameCount = 0
		}

	case "spin":
		// Rotate characters around their rings
		for i := range e.chars {
			if e.chars[i].clockwise {
				e.chars[i].angleOnRing += e.chars[i].rotationSpeed
			} else {
				e.chars[i].angleOnRing -= e.chars[i].rotationSpeed
			}

			// Normalize angle to [0, 2π]
			for e.chars[i].angleOnRing > 2*math.Pi {
				e.chars[i].angleOnRing -= 2 * math.Pi
			}
			for e.chars[i].angleOnRing < 0 {
				e.chars[i].angleOnRing += 2 * math.Pi
			}

			ring := &e.rings[e.chars[i].ringIndex]
			e.chars[i].currentX = e.centerX + ring.radius*math.Cos(e.chars[i].angleOnRing)
			e.chars[i].currentY = e.centerY + ring.radius*math.Sin(e.chars[i].angleOnRing)
		}

		if e.frameCount >= e.spinDuration {
			e.currentCycle++

			// Check if we should cycle back to disperse or move to final phase
			if e.currentCycle < e.spinDisperseCycles {
				e.phase = "swirl_to_rings"
				e.frameCount = 0
			} else {
				e.phase = "return_to_text"
				e.frameCount = 0
			}
		}

	case "return_to_text":
		progress := float64(e.frameCount) / float64(e.transitionFrames)
		if progress > 1.0 {
			progress = 1.0
		}

		// Ease-in-out function for smooth transition
		easedProgress := e.easeInOutCubic(progress)

		for i := range e.chars {
			ring := &e.rings[e.chars[i].ringIndex]
			ringGradient := e.ringGradients[e.chars[i].ringIndex]

			// Calculate current ring position
			ringX := e.centerX + ring.radius*math.Cos(e.chars[i].angleOnRing)
			ringY := e.centerY + ring.radius*math.Sin(e.chars[i].angleOnRing)

			// Interpolate back to original position
			e.chars[i].currentX = ringX + (float64(e.chars[i].x)-ringX)*easedProgress
			e.chars[i].currentY = ringY + (float64(e.chars[i].y)-ringY)*easedProgress

			// Reverse 8-step gradient back to final color
			gradientIndex := len(ringGradient) - 1 - int(easedProgress*float64(len(ringGradient)-1))
			if gradientIndex < 0 {
				gradientIndex = 0
			}
			if gradientIndex >= len(ringGradient) {
				gradientIndex = len(ringGradient) - 1
			}
			e.chars[i].currentColor = ringGradient[gradientIndex]
		}

		if e.frameCount >= e.transitionFrames {
			e.phase = "hold"
			e.frameCount = 0
		}

	case "hold":
		// Hold the final state for a bit before looping
		if e.frameCount >= 60 {
			e.Reset()
		}
	}
}

// Render returns the current frame as a colored string
func (e *RingTextEffect) Render() string {
	// Create a 2D buffer for the screen
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

		// Bounds check
		if x >= 0 && x < e.width && y >= 0 && y < e.height {
			buffer[y][x] = char.original
			colors[y][x] = char.currentColor
		}
	}

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
func (e *RingTextEffect) Reset() {
	e.phase = "static"
	e.frameCount = 0
	e.currentCycle = 0

	// Reset character positions
	for i := range e.chars {
		e.chars[i].currentX = float64(e.chars[i].x)
		e.chars[i].currentY = float64(e.chars[i].y)
		e.chars[i].currentColor = e.finalGradient[0]

		// Reset angle
		dx := float64(e.chars[i].x) - e.centerX
		dy := float64(e.chars[i].y) - e.centerY
		e.chars[i].angleOnRing = math.Atan2(dy, dx)
	}

	// Regenerate random disperse positions
	e.generateDispersePositions()
}

// SetText updates the displayed text and reinitializes the ring animation.
func (e *RingTextEffect) SetText(text string) {
	e.text = text
	e.init()
}

// createGradient creates a gradient between color stops
func (e *RingTextEffect) createGradient(stops []string, steps int) []string {
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

	// Add final color
	gradient = append(gradient, stops[len(stops)-1])
	return gradient
}

// applyStaticGradient applies theme-sensitive gradient to static ASCII presentation
func (e *RingTextEffect) applyStaticGradient() {
	if len(e.chars) == 0 || len(e.staticGradient) == 0 {
		return
	}

	// Find text bounds for gradient calculation
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

	// Apply gradient based on direction
	for i := range e.chars {
		var gradientPos float64

		switch e.staticGradientDir {
		case GradientHorizontal:
			// Left to right
			gradientPos = float64(e.chars[i].x-minX) / textWidth

		case GradientVertical:
			// Top to bottom
			gradientPos = float64(e.chars[i].y-minY) / textHeight

		case GradientDiagonal:
			// Top-left to bottom-right
			xPos := float64(e.chars[i].x-minX) / textWidth
			yPos := float64(e.chars[i].y-minY) / textHeight
			gradientPos = (xPos + yPos) / 2.0

		case GradientRadial:
			// Center outward
			dx := float64(e.chars[i].x) - e.centerX
			dy := float64(e.chars[i].y) - e.centerY
			maxDist := math.Sqrt(textWidth*textWidth+textHeight*textHeight) / 2.0
			dist := math.Sqrt(dx*dx + dy*dy)
			gradientPos = math.Min(dist/maxDist, 1.0)

		default:
			gradientPos = 0
		}

		// Clamp to [0, 1]
		if gradientPos < 0 {
			gradientPos = 0
		}
		if gradientPos > 1 {
			gradientPos = 1
		}

		// Map to gradient index
		gradientIndex := int(gradientPos * float64(len(e.staticGradient)-1))
		if gradientIndex >= len(e.staticGradient) {
			gradientIndex = len(e.staticGradient) - 1
		}
		if gradientIndex < 0 {
			gradientIndex = 0
		}

		e.chars[i].currentColor = e.staticGradient[gradientIndex]
	}
}

// easeInOutCubic applies an ease-in-out cubic easing function
func (e *RingTextEffect) easeInOutCubic(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	return 1 - math.Pow(-2*t+2, 3)/2
}
