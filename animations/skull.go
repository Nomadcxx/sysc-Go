package animations

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
)

// skullArt is the ASCII skull template. Leading/trailing newlines only — the
// per-line indentation is significant and defines the skull silhouette.
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

// AnimationPhase represents the current act of the skull animation.
type AnimationPhase int

const (
	// PhaseDrip — Act 1: skull drips into place, cyan scanline sweeps once
	// down the screen behind it.
	PhaseDrip AnimationPhase = iota

	// PhaseGrid — Act 2: tron-style horizon grid ignites left-to-right
	// behind the settled skull.
	PhaseGrid

	// PhaseHold — Act 3: grid holds steady, sonar pulses radiate from the
	// skull's centre on a repeating cycle.
	PhaseHold
)

// SkullChar represents a single character in the skull with drip state.
type SkullChar struct {
	char       rune
	finalX     int
	finalY     int
	currentY   float64
	velocity   float64
	startFrame int  // Row-based stagger so the skull cascades top-to-bottom
	locked     bool // Has reached final position
}

// AshParticle represents falling ash with horizontal drift.
type AshParticle struct {
	x          float64
	y          float64
	char       rune
	velocityY  float64
	driftPhase float64
	driftAmp   float64
	driftFreq  float64
	layer      int // 0=light, 1=dense
	opacity    float64
}

// gridCell marks a background grid glyph; `dist` is used for left-to-right
// ignition during Act 2.
type gridCell struct {
	x, y int
	ch   rune
}

// SkullAnimation implements the white-skull cyberpunk scene effect.
type SkullAnimation struct {
	width       int
	height      int
	palette     []string
	theme       string
	skullColor  string // Always bright — skull stays white through every phase.

	// Skull state
	skullChars   []SkullChar
	skullOffsetX int
	skullOffsetY int
	skullHeight  int
	skullCenterX int
	skullCenterY int

	// Background grid
	gridCells []gridCell
	horizonY  int

	// Ash particles
	ashParticles []AshParticle

	// Animation state
	phase          AnimationPhase
	frameCount     int
	gridStartFrame int
	holdStartFrame int

	builder strings.Builder
}

// NewSkullEffect creates a new skull scene effect.
func NewSkullEffect(width, height int, palette []string, theme string) *SkullAnimation {
	s := &SkullAnimation{
		width:          width,
		height:         height,
		palette:        palette,
		theme:          theme,
		skullColor:     palette[4], // Brightest slot — white in every theme
		phase:          PhaseDrip,
		frameCount:     0,
		gridStartFrame: -1,
		holdStartFrame: -1,
	}

	s.parseSkullArt()
	s.buildGrid()

	for i := 0; i < 50; i++ {
		s.spawnAshParticle()
	}

	return s
}

// parseSkullArt converts the ASCII skull template into an array of SkullChars
// positioned at their final coordinates, centred in the screen.
func (s *SkullAnimation) parseSkullArt() {
	lines := strings.Split(strings.Trim(skullArt, "\n"), "\n")
	skullHeight := len(lines)
	s.skullHeight = skullHeight

	// Centre vertically. Leave a 1-row top margin so the crown isn't flush.
	s.skullOffsetY = (s.height - skullHeight) / 2
	if s.skullOffsetY < 0 {
		s.skullOffsetY = 0
	}

	// Compute the minimum leading-space count across non-empty lines; this
	// becomes the "left margin" we strip so the widest row lands flush with
	// the start of the computed skull box.
	minLeading := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		leading := 0
		for _, ch := range line {
			if ch != ' ' {
				break
			}
			leading++
		}
		if minLeading == -1 || leading < minLeading {
			minLeading = leading
		}
	}
	if minLeading < 0 {
		minLeading = 0
	}

	// Measure the widest row (to the rightmost non-space rune) after trim.
	maxWidth := 0
	for _, line := range lines {
		runes := []rune(line)
		if minLeading < len(runes) {
			runes = runes[minLeading:]
		} else {
			runes = nil
		}
		rightMost := -1
		for i, ch := range runes {
			if ch != ' ' {
				rightMost = i
			}
		}
		if rightMost+1 > maxWidth {
			maxWidth = rightMost + 1
		}
	}

	s.skullOffsetX = (s.width - maxWidth) / 2
	if s.skullOffsetX < 0 {
		s.skullOffsetX = 0
	}
	s.skullCenterX = s.skullOffsetX + maxWidth/2
	s.skullCenterY = s.skullOffsetY + skullHeight/2

	s.skullChars = nil
	for y, line := range lines {
		runes := []rune(line)
		if minLeading < len(runes) {
			runes = runes[minLeading:]
		} else {
			continue
		}
		for col, ch := range runes {
			if ch == ' ' || ch == '\n' || ch == '\r' {
				continue
			}
			finalX := s.skullOffsetX + col
			finalY := s.skullOffsetY + y
			s.skullChars = append(s.skullChars, SkullChar{
				char:       ch,
				finalX:     finalX,
				finalY:     finalY,
				currentY:   -1,
				velocity:   0.3 + rand.Float64()*0.5,
				startFrame: y,
				locked:     false,
			})
		}
	}
}

// buildGrid precomputes the tron-style floor grid cells — horizontal dashed
// lines at perspective-spaced intervals below the horizon, plus four vertical
// "lane" markers converging toward the vanishing point.
func (s *SkullAnimation) buildGrid() {
	// Horizon sits at the skull's vertical centre so the grid feels like it
	// extends out from behind the skull.
	s.horizonY = s.skullCenterY

	s.gridCells = nil

	// Horizontal floor lines — gaps grow as we approach the viewer.
	rowOffsets := []int{2, 4, 7, 11, 16}
	for _, off := range rowOffsets {
		y := s.horizonY + off
		if y < 0 || y >= s.height {
			continue
		}
		for x := 0; x < s.width; x += 2 {
			s.gridCells = append(s.gridCells, gridCell{x: x, y: y, ch: '─'})
		}
	}

	// Vertical lane markers radiating from the vanishing point (skullCenterX,
	// horizonY). Four lanes — subtle but enough to read as grid perspective.
	lanes := []int{-6, -3, 3, 6}
	for _, lane := range lanes {
		for y := s.horizonY + 1; y < s.height; y++ {
			dy := y - s.horizonY
			x := s.skullCenterX + lane*dy/3
			if x < 0 || x >= s.width {
				continue
			}
			s.gridCells = append(s.gridCells, gridCell{x: x, y: y, ch: '│'})
		}
	}
}

// Update advances the animation by one frame.
func (s *SkullAnimation) Update() {
	s.frameCount++

	s.updateAsh()

	switch s.phase {
	case PhaseDrip:
		s.updateDrip()
	case PhaseGrid:
		if s.frameCount-s.gridStartFrame >= 50 {
			s.phase = PhaseHold
			s.holdStartFrame = s.frameCount
		}
	case PhaseHold:
		// Indefinite — caller stops via duration. Sonar loops.
	}
}

// updateDrip handles Act 1: the skull falls into place with a row-based
// stagger so it cascades top-to-bottom.
func (s *SkullAnimation) updateDrip() {
	allLocked := true
	for i := range s.skullChars {
		c := &s.skullChars[i]
		if c.locked {
			continue
		}
		allLocked = false
		if s.frameCount < c.startFrame {
			continue
		}
		c.velocity += 0.02
		c.currentY += c.velocity
		if c.currentY >= float64(c.finalY) {
			c.currentY = float64(c.finalY)
			c.locked = true
		}
	}

	if allLocked || s.frameCount >= 200 {
		s.phase = PhaseGrid
		s.gridStartFrame = s.frameCount
	}
}

// spawnAshParticle creates a new ash particle at the top of the screen.
func (s *SkullAnimation) spawnAshParticle() {
	densityMod := 0.9 + 0.1*math.Sin(float64(s.frameCount)/100.0)
	if rand.Float64() > densityMod {
		return
	}

	layer := 0
	if rand.Float64() < 0.3 {
		layer = 1
	}

	var char rune
	var velocityY, opacity float64

	if layer == 0 {
		chars := []rune{'.', ',', '\'', '·'}
		char = chars[rand.Intn(len(chars))]
		velocityY = 0.5 + rand.Float64()
		opacity = 0.4 + rand.Float64()*0.3
	} else {
		chars := []rune{'▒', '░', '·', '*'}
		char = chars[rand.Intn(len(chars))]
		velocityY = 0.3 + rand.Float64()*0.5
		opacity = 0.6 + rand.Float64()*0.3
	}

	s.ashParticles = append(s.ashParticles, AshParticle{
		x:          float64(rand.Intn(s.width)),
		y:          0,
		char:       char,
		velocityY:  velocityY,
		driftPhase: rand.Float64() * 2 * math.Pi,
		driftAmp:   0.1 + rand.Float64()*0.4,
		driftFreq:  0.05 + rand.Float64()*0.05,
		layer:      layer,
		opacity:    opacity,
	})
}

// updateAsh moves particles, applies drift/fade, recycles off-screen ones.
func (s *SkullAnimation) updateAsh() {
	toKeep := s.ashParticles[:0]
	for _, p := range s.ashParticles {
		p.y += p.velocityY
		p.driftPhase += p.driftFreq
		p.x += math.Sin(p.driftPhase) * p.driftAmp

		if p.layer == 0 && p.y > float64(s.height)*0.7 {
			fadeStart := float64(s.height) * 0.7
			fadeRange := float64(s.height) * 0.3
			fadeProgress := (p.y - fadeStart) / fadeRange
			p.opacity *= 1.0 - fadeProgress*0.5
		}

		accumZone := 3
		if p.layer == 1 {
			if p.y < float64(s.height+accumZone) {
				toKeep = append(toKeep, p)
			}
		} else if p.y < float64(s.height) {
			toKeep = append(toKeep, p)
		}
	}
	s.ashParticles = toKeep

	spawnCount := 2 + rand.Intn(3)
	for i := 0; i < spawnCount; i++ {
		s.spawnAshParticle()
	}
}

// scanRow returns the row the Act 1 scanline occupies this frame, or -1 if
// the sweep has completed or the animation is past Act 1.
func (s *SkullAnimation) scanRow() int {
	if s.phase != PhaseDrip {
		return -1
	}
	const scanDuration = 60
	if s.frameCount >= scanDuration {
		return -1
	}
	return s.frameCount * s.height / scanDuration
}

// gridLitFraction returns how far left-to-right the grid has ignited, 0..1.
func (s *SkullAnimation) gridLitFraction() float64 {
	switch s.phase {
	case PhaseDrip:
		return 0
	case PhaseGrid:
		const ignitionFrames = 50
		t := float64(s.frameCount-s.gridStartFrame) / ignitionFrames
		if t > 1 {
			t = 1
		}
		return t
	default:
		return 1
	}
}

// sonarRadius returns the current radius of the sonar ring plus its intensity
// (0..1 where 1 is freshly emitted). Returns (0, 0) when no ping is active.
func (s *SkullAnimation) sonarRadius() (float64, float64) {
	if s.phase != PhaseHold {
		return 0, 0
	}
	const cycleFrames = 80
	const expandFrames = 40
	rel := (s.frameCount - s.holdStartFrame) % cycleFrames
	if rel >= expandFrames {
		return 0, 0
	}
	t := float64(rel) / float64(expandFrames)
	radius := t * 28.0
	intensity := 1.0 - t
	return radius, intensity
}

// hexLerp linearly interpolates between two hex colour strings.
func hexLerp(from, to string, t float64) string {
	if t <= 0 {
		return from
	}
	if t >= 1 {
		return to
	}
	r1, g1, b1 := hexToRGB(from)
	r2, g2, b2 := hexToRGB(to)
	r := r1 + int(math.Round(t*float64(r2-r1)))
	g := g1 + int(math.Round(t*float64(g2-g1)))
	b := b1 + int(math.Round(t*float64(b2-b1)))
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// Render returns the current frame as a single string with ANSI escape codes.
func (s *SkullAnimation) Render() string {
	s.builder.Reset()

	// Build maps for fast per-cell lookup.
	skullMap := make(map[[2]int]SkullChar, len(s.skullChars))
	for _, c := range s.skullChars {
		x := int(c.currentX())
		y := int(c.currentY)
		if x >= 0 && x < s.width && y >= 0 && y < s.height {
			skullMap[[2]int{x, y}] = c
		}
	}

	ashMap := make(map[[2]int]AshParticle, len(s.ashParticles))
	for _, p := range s.ashParticles {
		x := int(p.x)
		y := int(p.y)
		if x >= 0 && x < s.width && y >= 0 && y < s.height {
			ashMap[[2]int{x, y}] = p
		}
	}

	gridMap := make(map[[2]int]rune, len(s.gridCells))
	litFrac := s.gridLitFraction()
	litX := int(float64(s.width) * litFrac)
	for _, g := range s.gridCells {
		if g.x >= litX && s.phase != PhaseHold {
			continue // Unlit during ignition sweep
		}
		if _, exists := gridMap[[2]int{g.x, g.y}]; !exists {
			gridMap[[2]int{g.x, g.y}] = g.ch
		}
	}

	scanY := s.scanRow()
	sonarR, sonarI := s.sonarRadius()

	// Pre-resolved colours.
	gridColor := s.palette[0]
	scanColor := s.palette[2]
	sonarColor := hexLerp(s.palette[0], s.palette[1], sonarI)

	// Render line-by-line. Skull always paints on top; priority below.
	for y := 0; y < s.height; y++ {
		for x := 0; x < s.width; x++ {
			pos := [2]int{x, y}

			if sc, ok := skullMap[pos]; ok {
				// Skull — always pure bright.
				r, g, b := hexToRGB(s.skullColor)
				fmt.Fprintf(&s.builder, "\033[38;2;%d;%d;%dm%c\033[0m", r, g, b, sc.char)
				continue
			}

			if ap, ok := ashMap[pos]; ok {
				var col string
				if ap.layer == 0 {
					col = s.palette[5]
				} else {
					col = s.palette[6]
				}
				r, g, b := hexToRGB(col)
				fmt.Fprintf(&s.builder, "\033[38;2;%d;%d;%dm%c\033[0m", r, g, b, ap.char)
				continue
			}

			if y == scanY {
				r, g, b := hexToRGB(scanColor)
				fmt.Fprintf(&s.builder, "\033[38;2;%d;%d;%dm─\033[0m", r, g, b)
				continue
			}

			if sonarR > 0 && s.onSonarRing(x, y, sonarR) {
				r, g, b := hexToRGB(sonarColor)
				fmt.Fprintf(&s.builder, "\033[38;2;%d;%d;%dm·\033[0m", r, g, b)
				continue
			}

			if ch, ok := gridMap[pos]; ok {
				r, g, b := hexToRGB(gridColor)
				fmt.Fprintf(&s.builder, "\033[38;2;%d;%d;%dm%c\033[0m", r, g, b, ch)
				continue
			}

			s.builder.WriteByte(' ')
		}
		if y < s.height-1 {
			s.builder.WriteByte('\n')
		}
	}

	return s.builder.String()
}

// onSonarRing reports whether (x, y) sits on the sonar ring at the given
// radius, accounting for the roughly 2:1 aspect ratio of terminal cells.
func (s *SkullAnimation) onSonarRing(x, y int, radius float64) bool {
	dx := float64(x - s.skullCenterX)
	dy := float64(y-s.skullCenterY) * 2.0
	d := math.Sqrt(dx*dx + dy*dy)
	return math.Abs(d-radius) < 0.9
}

// currentX returns the rendered X of a skull char (unchanged after placement).
func (c SkullChar) currentX() float64 {
	return float64(c.finalX)
}

// Reset restarts the animation from Act 1.
func (s *SkullAnimation) Reset() {
	s.phase = PhaseDrip
	s.frameCount = 0
	s.gridStartFrame = -1
	s.holdStartFrame = -1

	for i := range s.skullChars {
		c := &s.skullChars[i]
		c.currentY = -1
		c.velocity = 0.3 + rand.Float64()*0.5
		c.locked = false
	}
	// Ash particles continue seamlessly across resets.
}
