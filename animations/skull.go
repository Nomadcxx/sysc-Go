package animations

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
)

// hexToAnsiRGB converts hex color (#RRGGBB) to ANSI RGB format "r;g;b"
func hexToAnsiRGB(hex string) string {
	if len(hex) < 7 || hex[0] != '#' {
		return "255;255;255"
	}

	// Parse r, g, b from hex
	var r, g, b int
	fmt.Sscanf(hex[1:3], "%x", &r)
	fmt.Sscanf(hex[3:5], "%x", &g)
	fmt.Sscanf(hex[5:7], "%x", &b)

	return fmt.Sprintf("%d;%d;%d", r, g, b)
}

// skullArt is the ASCII skull template
const skullArt = `
                ▄▄▟██████████▄▄▄▖
            ▗▄███████████▄▄▄▄▄▛▀▀▜▙▄▖
          ▄▟█████████████████████▙▄█▀▙▄
        ▄████████████████████████████▙██▄
      ▗▟██████████████████████████████████▖
      ▟████████████████████████████████████▄
      ██████████████████████████████████████
      ██████████████████████████████████████
      ███▛██████████████████████████████▀███
      ███▌  ▀▜██████████████████████▛▀   ██▛
      ▜██▌     ▀▀████████████████▀▘     ▐██▌
      ▐██▙        ▝▀▜████████▛▀▘        ▟██
       ▜██▙▄▄▖        ▝▜██▌▘         ▄▄▟██▛
        ▀██████████████▛▜▛▜███████████████▄
       ▗▄█████████████▛ ▐▌ ▜██████████████▛
       ▐██████████████▌▗██▖▗███████████████
       ▝██████▀▘  ▐███████▙██████    ▝████▛
        ▝▜▀▀▜▄▖   ▐██████████████    ▝█▛▀▘
                  ▝██▛▜██ ██ ▟██▛
                   ███▐██ ██ ███
                   ▜██▐██ ██ ██▛
                   ▐██▐██ ██ ██▌
                   ▐██▐██ ██ ██▌
                   ▐██▐██ █▌▐██▌
                    ██ ██ █▌▐██▌
                    ██ █▌▐█▘▐██
                    ▀▛ ▜▌▝▛ ▐▛▀
`

// AnimationPhase represents the current state of the skull animation
type AnimationPhase int

const (
	PhaseDrip AnimationPhase = iota
	PhaseIllumination
	PhaseTextEntrance // skull-text only
	PhaseHold
)

// SkullChar represents a single character in the skull with animation state
type SkullChar struct {
	char       rune
	finalX     int
	finalY     int
	currentX   float64
	currentY   float64
	velocity   float64 // Drip speed (0.3-0.8 cells/frame)
	locked     bool    // Has reached final position
	accentType int     // 0=none, 1=top, 2=eyes, 3=cheeks, 4=teeth
}

// AshParticle represents falling ash with drift
type AshParticle struct {
	x          float64
	y          float64
	char       rune
	velocityY  float64 // Fall speed
	driftPhase float64 // For sine wave drift
	driftAmp   float64 // Drift amplitude
	driftFreq  float64 // Drift frequency
	layer      int     // 0=light, 1=dense
	opacity    float64 // For fading effect
}

// TextChar represents sliding text character
type TextChar struct {
	char     rune
	finalX   int
	finalY   int
	startX   float64 // Off-screen starting position
	currentX float64
	progress float64 // 0.0-1.0 animation progress
}

// SkullAnimation implements the skull drip effect
type SkullAnimation struct {
	width       int
	height      int
	palette     []string
	theme       string
	withText    bool
	textContent string

	// Skull state
	skullChars   []SkullChar
	skullArt     string // The ASCII skull art
	skullOffsetY int    // Vertical position

	// Ash particles
	ashParticles []AshParticle

	// Text state (skull-text only)
	textChars    []TextChar
	textGradient []string

	// Animation state
	phase      AnimationPhase
	frameCount int
}

// NewSkullEffect creates a new skull background effect
func NewSkullEffect(width, height int, palette []string, theme string) *SkullAnimation {
	return newSkullAnimation(width, height, palette, theme, false, "")
}

// NewSkullTextEffect creates a skull effect with text integration
func NewSkullTextEffect(width, height int, palette []string, theme string, text string) *SkullAnimation {
	return newSkullAnimation(width, height, palette, theme, true, text)
}

// newSkullAnimation is the internal constructor
func newSkullAnimation(width, height int, palette []string, theme string, withText bool, text string) *SkullAnimation {
	s := &SkullAnimation{
		width:       width,
		height:      height,
		palette:     palette,
		theme:       theme,
		withText:    withText,
		textContent: text,
		phase:       PhaseDrip,
		frameCount:  0,
	}

	// Initialize skull art
	s.skullArt = skullArt
	s.parseSkullArt()

	// Initialize ash particles (start with ~50 particles)
	s.ashParticles = []AshParticle{}
	for i := 0; i < 50; i++ {
		s.spawnAshParticle()
	}

	// Initialize text if withText mode
	if withText {
		s.parseText()
	}

	return s
}

// parseSkullArt converts ASCII art to SkullChar array with positions
func (s *SkullAnimation) parseSkullArt() {
	lines := strings.Split(strings.TrimSpace(s.skullArt), "\n")
	skullHeight := len(lines)

	// Position skull in bottom half of terminal
	s.skullOffsetY = s.height - skullHeight - 2 // 2 rows padding from bottom
	if s.skullOffsetY < 0 {
		s.skullOffsetY = 0
	}

	// Find minimum leading whitespace to trim consistently
	minLeading := -1
	for _, line := range lines {
		// Count leading spaces
		leading := 0
		for _, ch := range line {
			if ch == ' ' {
				leading++
			} else {
				break
			}
		}
		// Skip empty lines
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}
		if minLeading == -1 || leading < minLeading {
			minLeading = leading
		}
	}
	if minLeading == -1 {
		minLeading = 0
	}

	// Find max line width after trimming leading spaces
	maxWidth := 0
	for _, line := range lines {
		trimmed := line
		if len(line) >= minLeading {
			trimmed = line[minLeading:]
		}
		contentWidth := 0
		for _, ch := range trimmed {
			if ch != ' ' {
				contentWidth++
			}
		}
		if contentWidth > maxWidth {
			maxWidth = contentWidth
		}
	}
	offsetX := (s.width - maxWidth) / 2

	// Parse each character
	s.skullChars = []SkullChar{}
	for y, line := range lines {
		// Trim the minimum leading spaces from this line
		trimmedLine := line
		if len(line) >= minLeading {
			trimmedLine = line[minLeading:]
		}
		for x, ch := range trimmedLine {
			if ch != ' ' && ch != '\n' && ch != '\r' {
				finalY := s.skullOffsetY + y
				finalX := offsetX + x

				// Random starting position above screen
				startY := float64(finalY) - float64(5+rand.Intn(10))

				char := SkullChar{
					char:       ch,
					finalX:     finalX,
					finalY:     finalY,
					currentX:   float64(finalX),
					currentY:   startY,
					velocity:   0.3 + rand.Float64()*0.5, // 0.3-0.8
					locked:     false,
					accentType: s.identifyAccentType(finalX, finalY),
				}
				s.skullChars = append(s.skullChars, char)
			}
		}
	}
}

// identifyAccentType determines if a character is in an accent region
// Returns: 0=none, 1=top, 2=eyes, 3=cheeks, 4=teeth
func (s *SkullAnimation) identifyAccentType(finalX, finalY int) int {
	// Calculate relative position within skull
	relY := finalY - s.skullOffsetY

	// Top details: upper portion (crown/forehead)
	if relY >= 0 && relY <= 8 {
		return 1
	}

	// Eyes: middle-upper region, specific columns (approximate based on skull art)
	if relY >= 9 && relY <= 12 {
		// Estimate eye positions based on ~50 char skull width
		// Left eye around cols 8-15, right eye around cols 36-43
		relX := finalX - ((s.width - 50) / 2)
		if (relX >= 8 && relX <= 18) || (relX >= 32 && relX <= 42) {
			return 2
		}
	}

	// Cheekbones: middle region, outer columns
	if relY >= 13 && relY <= 18 {
		relX := finalX - ((s.width - 50) / 2)
		if (relX >= 5 && relX <= 12) || (relX >= 38 && relX <= 45) {
			return 3
		}
	}

	// Teeth: lower region (bottom jaw)
	if relY >= 19 && relY <= 25 {
		return 4
	}

	return 0 // No accent
}

// parseText converts text into TextChar array with sliding positions
func (s *SkullAnimation) parseText() {
	if !s.withText || s.textContent == "" {
		return
	}

	lines := strings.Split(s.textContent, "\n")
	maxWidth := 0
	for _, line := range lines {
		lineLen := len([]rune(line))
		if lineLen > maxWidth {
			maxWidth = lineLen
		}
	}

	// Center text in terminal
	textHeight := len(lines)
	offsetY := (s.height - textHeight) / 2
	offsetX := (s.width - maxWidth) / 2
	centerX := s.width / 2

	s.textChars = []TextChar{}
	for y, line := range lines {
		col := 0
		for _, ch := range line {
			if ch != ' ' && ch != '\n' && ch != '\r' {
				finalX := offsetX + col
				finalY := offsetY + y

				// Determine starting position (left or right of center)
				var startX float64
				if finalX < centerX {
					startX = -5 // Off-screen left
				} else {
					startX = float64(s.width + 5) // Off-screen right
				}

				textChar := TextChar{
					char:     ch,
					finalX:   finalX,
					finalY:   finalY,
					startX:   startX,
					currentX: startX,
					progress: 0.0,
				}
				s.textChars = append(s.textChars, textChar)
			}
			col++
		}
	}
}

// SetTextGradient sets the gradient colors for text rendering
func (s *SkullAnimation) SetTextGradient(gradient []string) {
	s.textGradient = gradient
}

// getTextCharColor returns gradient color for text character
func (s *SkullAnimation) getTextCharColor(char TextChar) string {
	if len(s.textGradient) == 0 {
		return s.palette[4] // Fallback to brightest
	}

	// Calculate gradient position based on X coordinate
	// Find min and max X for normalization
	minX := s.width
	maxX := 0
	for _, tc := range s.textChars {
		if tc.finalX < minX {
			minX = tc.finalX
		}
		if tc.finalX > maxX {
			maxX = tc.finalX
		}
	}

	if maxX <= minX {
		return s.textGradient[0]
	}

	// Normalize position: 0.0 (left) to 1.0 (right)
	t := float64(char.finalX-minX) / float64(maxX-minX)

	// Map to gradient index
	gradLen := len(s.textGradient)
	idx := int(t * float64(gradLen-1))
	if idx >= gradLen {
		idx = gradLen - 1
	}

	return s.textGradient[idx]
}

// spawnAshParticle creates a new ash particle at the top
func (s *SkullAnimation) spawnAshParticle() {
	// Density breathing effect (10% variation)
	densityMod := 0.9 + 0.1*math.Sin(float64(s.frameCount)/100.0)

	// Skip spawn randomly based on density
	if rand.Float64() > densityMod {
		return
	}

	// Random layer: 70% light, 30% dense
	layer := 0
	if rand.Float64() < 0.3 {
		layer = 1
	}

	var char rune
	var velocityY, opacity float64

	if layer == 0 {
		// Light ash
		chars := []rune{'.', ',', '\'', '·'}
		char = chars[rand.Intn(len(chars))]
		velocityY = 0.5 + rand.Float64()   // 0.5-1.5
		opacity = 0.4 + rand.Float64()*0.3 // 0.4-0.7
	} else {
		// Dense ash
		chars := []rune{'▒', '░', '·', '*'}
		char = chars[rand.Intn(len(chars))]
		velocityY = 0.3 + rand.Float64()*0.5 // 0.3-0.8
		opacity = 0.6 + rand.Float64()*0.3   // 0.6-0.9
	}

	particle := AshParticle{
		x:          float64(rand.Intn(s.width)),
		y:          0,
		char:       char,
		velocityY:  velocityY,
		driftPhase: rand.Float64() * 2 * math.Pi, // Random starting phase
		driftAmp:   0.1 + rand.Float64()*0.4,     // 0.1-0.5 for variety
		driftFreq:  0.05 + rand.Float64()*0.05,   // 0.05-0.1
		layer:      layer,
		opacity:    opacity,
	}

	s.ashParticles = append(s.ashParticles, particle)
}

// Update advances the animation by one frame
func (s *SkullAnimation) Update() {
	s.frameCount++

	// Update ash continuously (all phases)
	s.updateAsh()

	// Phase-specific updates
	switch s.phase {
	case PhaseDrip:
		s.updateDrip()
	case PhaseIllumination:
		s.updateIllumination()
	case PhaseTextEntrance:
		if s.withText {
			s.updateTextEntrance()
		}
	case PhaseHold:
		s.updateHold()
	}
}

// updateDrip handles Phase 1: skull characters dripping into place
func (s *SkullAnimation) updateDrip() {
	allLocked := true

	for i := range s.skullChars {
		char := &s.skullChars[i]

		if !char.locked {
			// Apply gravity (simple acceleration)
			char.velocity += 0.02 // Small acceleration
			char.currentY += char.velocity

			// Check if reached final position
			if char.currentY >= float64(char.finalY) {
				char.currentY = float64(char.finalY)
				char.locked = true
			} else {
				allLocked = false
			}
		}
	}

	// Transition to illumination phase when all locked (around frame 110)
	if allLocked || s.frameCount >= 120 {
		s.phase = PhaseIllumination
	}
}

// updateIllumination handles Phase 2: accent point illumination
func (s *SkullAnimation) updateIllumination() {
	// Phase starts at frame 111, runs for 80 frames (until frame 190)
	relFrame := s.frameCount - 111

	// No illumination updates needed - color is determined in render
	// based on frame count and accent type

	// Transition to next phase after 80 frames
	if relFrame >= 80 {
		if s.withText {
			s.phase = PhaseTextEntrance
		} else {
			s.phase = PhaseHold
		}
	}
}

// updateTextEntrance handles Phase 3: text sliding in
func (s *SkullAnimation) updateTextEntrance() {
	// Phase starts at frame 191, runs for 40 frames (until frame 230)
	relFrame := s.frameCount - 191

	// Update text character positions
	for i := range s.textChars {
		char := &s.textChars[i]

		// Progress: 0.0 to 1.0 over 40 frames
		char.progress = float64(relFrame) / 40.0
		if char.progress > 1.0 {
			char.progress = 1.0
		}

		// Apply easing
		easedProgress := easeOutBack(char.progress)

		// Interpolate position
		finalX := float64(char.finalX)
		char.currentX = char.startX + (finalX-char.startX)*easedProgress
	}

	// Transition to hold phase after 40 frames
	if relFrame >= 40 {
		s.phase = PhaseHold
	}
}

// updateHold handles Phase 4: hold state before reset
func (s *SkullAnimation) updateHold() {
	// Determine when hold phase started
	holdStart := 190 // After drip (110) + illumination (80)
	if s.withText {
		holdStart = 230 // After drip (110) + illumination (80) + text (40)
	}

	// Hold for 200 frames (10 seconds)
	if s.frameCount >= holdStart+200 {
		s.Reset()
	}
}

// easeOutBack implements out-back easing (overshoot and settle)
func easeOutBack(t float64) float64 {
	c1 := 1.70158
	c3 := c1 + 1
	return 1 + c3*math.Pow(t-1, 3) + c1*math.Pow(t-1, 2)
}

// updateAsh updates all ash particles (wind drift, falling, lifecycle)
func (s *SkullAnimation) updateAsh() {
	// Update existing particles
	toKeep := []AshParticle{}
	for _, p := range s.ashParticles {
		// Update vertical position
		p.y += p.velocityY

		// Update horizontal drift (sine wave)
		p.driftPhase += p.driftFreq
		p.x += math.Sin(p.driftPhase) * p.driftAmp

		// Fade as it falls (for light ash)
		if p.layer == 0 && p.y > float64(s.height)*0.7 {
			fadeStart := float64(s.height) * 0.7
			fadeRange := float64(s.height) * 0.3
			fadeProgress := (p.y - fadeStart) / fadeRange
			p.opacity *= (1.0 - fadeProgress*0.5) // Fade to 50% opacity
		}

		// Keep if still on screen (with accumulation zone at bottom)
		accumZone := 3
		if p.layer == 1 {
			// Dense ash accumulates at bottom
			if p.y < float64(s.height+accumZone) {
				toKeep = append(toKeep, p)
			}
		} else {
			// Light ash disappears at bottom
			if p.y < float64(s.height) {
				toKeep = append(toKeep, p)
			}
		}
	}
	s.ashParticles = toKeep

	// Spawn new particles (2-4 per frame)
	spawnCount := 2 + rand.Intn(3)
	for i := 0; i < spawnCount; i++ {
		s.spawnAshParticle()
	}
}

// Render returns the current frame as a string
func (s *SkullAnimation) Render() string {
	// Create buffer
	buffer := make([][]rune, s.height)
	colors := make([][]string, s.height)
	for i := range buffer {
		buffer[i] = make([]rune, s.width)
		colors[i] = make([]string, s.width)
		for j := range buffer[i] {
			buffer[i][j] = ' '
			colors[i][j] = ""
		}
	}

	// Render ash particles (background layer)
	ashColor1 := s.palette[5] // ash1_dark
	ashColor2 := s.palette[6] // ash2_darker

	for _, p := range s.ashParticles {
		x := int(p.x)
		y := int(p.y)
		if x >= 0 && x < s.width && y >= 0 && y < s.height {
			buffer[y][x] = p.char
			if p.layer == 0 {
				colors[y][x] = ashColor1
			} else {
				colors[y][x] = ashColor2
			}
		}
	}

	// Render skull (middle layer)
	for _, char := range s.skullChars {
		x := int(char.currentX)
		y := int(char.currentY)
		if x >= 0 && x < s.width && y >= 0 && y < s.height {
			buffer[y][x] = char.char

			// Color based on phase and accent type
			color := s.getSkullCharColor(char)
			colors[y][x] = color
		}
	}

	// Render text (foreground layer, skull-text only)
	if s.withText && (s.phase == PhaseTextEntrance || s.phase == PhaseHold) {
		// First, clear a rectangle area where text will be to prevent skull showing through
		// Calculate text bounds
		if len(s.textChars) > 0 {
			minX, maxX := s.width, 0
			minY, maxY := s.height, 0
			for _, tchar := range s.textChars {
				if tchar.finalX < minX {
					minX = tchar.finalX
				}
				if tchar.finalX > maxX {
					maxX = tchar.finalX
				}
				if tchar.finalY < minY {
					minY = tchar.finalY
				}
				if tchar.finalY > maxY {
					maxY = tchar.finalY
				}
			}
			// Clear the text area (fill with spaces)
			for y := minY; y <= maxY && y >= 0 && y < s.height; y++ {
				for x := minX; x <= maxX && x >= 0 && x < s.width; x++ {
					buffer[y][x] = ' '
					colors[y][x] = ""
				}
			}
		}
		// Now render the actual text characters
		for _, tchar := range s.textChars {
			x := int(tchar.currentX)
			y := tchar.finalY
			if x >= 0 && x < s.width && y >= 0 && y < s.height {
				buffer[y][x] = tchar.char
				colors[y][x] = s.getTextCharColor(tchar)
			}
		}
	}

	// Build output with ANSI color codes (more efficient than lipgloss per-character)
	var sb strings.Builder
	currentColor := ""

	for y := 0; y < s.height; y++ {
		for x := 0; x < s.width; x++ {
			char := buffer[y][x]
			color := colors[y][x]

			if color != currentColor {
				if color != "" {
					// Set new color using ANSI escape code
					sb.WriteString("\033[38;2;")
					sb.WriteString(hexToAnsiRGB(color))
					sb.WriteString("m")
				} else {
					// Reset color
					sb.WriteString("\033[0m")
				}
				currentColor = color
			}
			sb.WriteRune(char)
		}
		if y < s.height-1 {
			sb.WriteString("\n")
		}
	}

	// Reset color at end
	if currentColor != "" {
		sb.WriteString("\033[0m")
	}

	return sb.String()
}

// getSkullCharColor returns the color for a skull character based on phase
func (s *SkullAnimation) getSkullCharColor(char SkullChar) string {
	// During drip, use base muted color
	if s.phase == PhaseDrip {
		return s.palette[0]
	}

	// During illumination, text entrance, and hold: accent colors
	if char.accentType == 0 {
		return s.palette[0] // Base color for non-accent
	}

	// Check if this accent region is illuminated yet
	relFrame := s.frameCount - 111 // Illumination starts at frame 111

	// Top details: frames 0-19 (111-130)
	if char.accentType == 1 && relFrame >= 0 {
		return s.palette[1] // accent1_dark
	}

	// Eyes: frames 20-44 (131-155)
	if char.accentType == 2 && relFrame >= 20 {
		return s.palette[2] // accent2_mid
	}

	// Cheekbones: frames 45-64 (156-175)
	if char.accentType == 3 && relFrame >= 45 {
		return s.palette[3] // accent3_bright
	}

	// Teeth: frames 65-79 (176-190)
	if char.accentType == 4 && relFrame >= 65 {
		return s.palette[4] // accent4_brightest
	}

	// Not yet illuminated
	return s.palette[0]
}

// Reset restarts the animation
func (s *SkullAnimation) Reset() {
	s.phase = PhaseDrip
	s.frameCount = 0

	// Reset skull characters to starting drip positions
	for i := range s.skullChars {
		char := &s.skullChars[i]
		char.currentY = float64(char.finalY) - float64(5+rand.Intn(10))
		char.currentX = float64(char.finalX)
		char.velocity = 0.3 + rand.Float64()*0.5
		char.locked = false
	}

	// Reset text characters if withText
	if s.withText {
		for i := range s.textChars {
			char := &s.textChars[i]
			char.currentX = char.startX
			char.progress = 0.0
		}
	}

	// DO NOT reset ash particles - they continue seamlessly
}
