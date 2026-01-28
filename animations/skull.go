package animations

import (
	"math"
	"math/rand"
	"strings"
)

// skullArt is the ASCII skull template
const skullArt = `
                ..::::::::::::::...
           ..:-----------------------:.
         :------------------------------::
       :----:------------------------------:
     .-----------------------------------:---
     ----:-------------------------------:----
    :----::-----------------------------::----.
    .-:-:.:-----------------------------:.-::-.
     :-  .--.      .:-::---::-:.      :--.  -:
      : .--            :---.           .--. :
       .:--.         .:-:.:-:.         :--:.
      .------:....::---.   .---:.....:------.
       ..::::--------:       :--------::::..
         . :.   :-----:::-:::-----.   ..
         -: -:   ::::::::.:::.::::   :: :
         :: :--. ..:::::-.-::::... .--: :
          :.  .--:::-...:::.:.::::--.  :.
          ..    :::-------------::.    :
          .      ...::-:::::-:...      .
          ..     :-.     :.    --      .
          ..      -.     .:    ::      .
                  ..            .
                  .      .      .
                  ..     .      .
                  ..     .      .
                                .
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

	// TODO: Initialize text characters if withText

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

	// Find max line width for centering
	maxWidth := 0
	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}
	offsetX := (s.width - maxWidth) / 2

	// Parse each character
	s.skullChars = []SkullChar{}
	for y, line := range lines {
		for x, ch := range line {
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
					accentType: 0, // TODO: Set based on position
				}
				s.skullChars = append(s.skullChars, char)
			}
		}
	}
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
		x:         float64(rand.Intn(s.width)),
		y:         0,
		char:      char,
		velocityY: velocityY,
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
	// TODO: Implement illumination sequence

	// Temporary: transition after 80 frames
	if s.frameCount >= 190 {
		if s.withText {
			s.phase = PhaseTextEntrance
		} else {
			s.phase = PhaseHold
		}
	}
}

// updateTextEntrance handles Phase 3: text sliding in
func (s *SkullAnimation) updateTextEntrance() {
	// TODO: Implement text sliding

	// Temporary: transition after 40 frames
	if s.frameCount >= 230 {
		s.phase = PhaseHold
	}
}

// updateHold handles Phase 4: hold state before reset
func (s *SkullAnimation) updateHold() {
	// Hold for 200 frames (10 seconds)
	if s.frameCount >= 430 { // 110 + 80 + 40 + 200
		s.Reset()
	}
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
	// TODO: Render ash background
	// TODO: Render skull with current colors
	// TODO: Render text if in TextEntrance or Hold phase
	return ""
}

// Reset restarts the animation
func (s *SkullAnimation) Reset() {
	s.phase = PhaseDrip
	s.frameCount = 0
	// TODO: Reset skull character positions
	// Note: ash particles continue (don't reset)
	// TODO: Reset text character positions if withText
}
