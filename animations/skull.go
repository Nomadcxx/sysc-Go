package animations

import (
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
	// TODO: Initialize ash particles
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

// Update advances the animation by one frame
func (s *SkullAnimation) Update() {
	s.frameCount++

	// TODO: Update based on phase
	// TODO: Update ash particles continuously
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
