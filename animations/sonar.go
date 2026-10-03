package animations

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
)

// SonarRing is a single expanding pulse. Birth frame stores when it was
// emitted so radius + intensity can be derived at render time.
type SonarRing struct {
	birthFrame int
}

// SonarAnimation is the skull scene minus the skull ASCII: a three-act
// cyberpunk sequence (scan sweep → tron grid ignition → multi-ring sonar
// pulses) that radiates from the viewport centre to the edges.
type SonarAnimation struct {
	width, height int
	palette       []string
	theme         string

	centerX int
	centerY int

	// Background grid (tron perspective).
	gridCells []gridCell
	horizonY  int

	// Ash particles drifting down — same visual texture as the skull scene.
	ashParticles []AshParticle

	// Concurrent expanding sonar pulses.
	rings         []SonarRing
	spawnInterval int
	lifespan      int
	maxRadius     float64

	// Phase state.
	phase          AnimationPhase
	frameCount     int
	scanFrames     int
	gridStartFrame int
	gridDuration   int
	holdStartFrame int

	builder strings.Builder
}

// NewSonarEffect builds the sonar scene. It reuses the skull palette layout
// (slot 0 = dim grid, slot 1 = bright pulse, slot 2 = scanline, slots 5/6 =
// ash) so theme switching stays consistent with the skull effect.
func NewSonarEffect(width, height int, palette []string, theme string) *SonarAnimation {
	s := &SonarAnimation{
		width:          width,
		height:         height,
		palette:        palette,
		theme:          theme,
		centerX:        width / 2,
		centerY:        height / 2,
		phase:          PhaseDrip,
		scanFrames:     60,
		gridDuration:   50,
		gridStartFrame: -1,
		holdStartFrame: -1,
		spawnInterval:  22,
		lifespan:       110,
	}

	dx := float64(width) / 2.0
	dy := float64(height) / 2.0 * 2.0
	s.maxRadius = math.Sqrt(dx*dx+dy*dy) + 1.0

	s.buildGrid()

	for i := 0; i < 50; i++ {
		s.spawnAshParticle()
	}

	return s
}

// buildGrid precomputes the tron-style floor grid: horizontal dashed lines
// at perspective-spaced intervals below the horizon plus four converging
// vertical lane markers.
func (s *SonarAnimation) buildGrid() {
	s.horizonY = s.centerY
	s.gridCells = s.gridCells[:0]

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

	lanes := []int{-6, -3, 3, 6}
	for _, lane := range lanes {
		for y := s.horizonY + 1; y < s.height; y++ {
			dy := y - s.horizonY
			x := s.centerX + lane*dy/3
			if x < 0 || x >= s.width {
				continue
			}
			s.gridCells = append(s.gridCells, gridCell{x: x, y: y, ch: '│'})
		}
	}
}

// spawnAshParticle creates a new ash particle at the top of the screen.
func (s *SonarAnimation) spawnAshParticle() {
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

// updateAsh moves particles, applies drift/fade, recycles off-screen ones,
// then spawns a handful of fresh particles per frame to keep density steady.
func (s *SonarAnimation) updateAsh() {
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

// Update advances the animation by one frame: drives phase transitions,
// spawns and ages sonar rings, and updates ash.
func (s *SonarAnimation) Update() {
	s.frameCount++
	s.updateAsh()

	switch s.phase {
	case PhaseDrip:
		if s.frameCount >= s.scanFrames {
			s.phase = PhaseGrid
			s.gridStartFrame = s.frameCount
		}
	case PhaseGrid:
		if s.frameCount-s.gridStartFrame >= s.gridDuration {
			s.phase = PhaseHold
			s.holdStartFrame = s.frameCount
		}
	case PhaseHold:
		if (s.frameCount-s.holdStartFrame)%s.spawnInterval == 0 {
			s.rings = append(s.rings, SonarRing{birthFrame: s.frameCount})
		}
		kept := s.rings[:0]
		for _, r := range s.rings {
			if s.frameCount-r.birthFrame < s.lifespan {
				kept = append(kept, r)
			}
		}
		s.rings = kept
	}
}

// scanRow returns the row occupied by the Act 1 scanline sweep, or -1 once
// the sweep is past.
func (s *SonarAnimation) scanRow() int {
	if s.phase != PhaseDrip {
		return -1
	}
	if s.frameCount >= s.scanFrames {
		return -1
	}
	return s.frameCount * s.height / s.scanFrames
}

// gridLitFraction returns the 0..1 fraction of the grid currently ignited
// (left-to-right sweep during Act 2, fully lit during Act 3).
func (s *SonarAnimation) gridLitFraction() float64 {
	switch s.phase {
	case PhaseDrip:
		return 0
	case PhaseGrid:
		t := float64(s.frameCount-s.gridStartFrame) / float64(s.gridDuration)
		if t > 1 {
			t = 1
		}
		return t
	default:
		return 1
	}
}

// ringRadius returns the current radius and 0..1 intensity of a ring. Fresh
// rings are small and bright; expiring rings are large and dim.
func (s *SonarAnimation) ringRadius(r SonarRing) (float64, float64) {
	age := s.frameCount - r.birthFrame
	if age < 0 || age >= s.lifespan {
		return 0, 0
	}
	t := float64(age) / float64(s.lifespan)
	return t * s.maxRadius, 1.0 - t
}

// onRing reports whether (x, y) sits on the ring at the given radius, with
// 2:1 aspect correction for terminal cells.
func (s *SonarAnimation) onRing(x, y int, radius float64) bool {
	dx := float64(x - s.centerX)
	dy := float64(y-s.centerY) * 2.0
	d := math.Sqrt(dx*dx + dy*dy)
	return math.Abs(d-radius) < 0.9
}

// Render produces the current frame. Priority per cell: ash > scan > sonar
// ring > grid > void. Brighter sonar rings win when radii overlap.
func (s *SonarAnimation) Render() string {
	s.builder.Reset()

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
			continue
		}
		if _, exists := gridMap[[2]int{g.x, g.y}]; !exists {
			gridMap[[2]int{g.x, g.y}] = g.ch
		}
	}

	type ringHit struct {
		intensity float64
	}
	ringMap := make(map[[2]int]ringHit)
	for _, r := range s.rings {
		radius, intensity := s.ringRadius(r)
		if radius <= 0 {
			continue
		}
		for y := 0; y < s.height; y++ {
			dy := float64(y-s.centerY) * 2.0
			if math.Abs(dy) > radius+1 {
				continue
			}
			for x := 0; x < s.width; x++ {
				if s.onRing(x, y, radius) {
					if prev, ok := ringMap[[2]int{x, y}]; !ok || intensity > prev.intensity {
						ringMap[[2]int{x, y}] = ringHit{intensity: intensity}
					}
				}
			}
		}
	}

	scanY := s.scanRow()
	gridColor := s.palette[0]
	scanColor := s.palette[2]

	for y := 0; y < s.height; y++ {
		for x := 0; x < s.width; x++ {
			pos := [2]int{x, y}

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

			if h, ok := ringMap[pos]; ok {
				col := hexLerp(s.palette[0], s.palette[1], h.intensity)
				r, g, b := hexToRGB(col)
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

// Reset returns the animation to frame 0.
func (s *SonarAnimation) Reset() {
	s.frameCount = 0
	s.phase = PhaseDrip
	s.gridStartFrame = -1
	s.holdStartFrame = -1
	s.rings = s.rings[:0]
	s.ashParticles = s.ashParticles[:0]
	for i := 0; i < 50; i++ {
		s.spawnAshParticle()
	}
}

// Resize handles terminal size changes by rebuilding the grid and
// recomputing the max sonar radius.
func (s *SonarAnimation) Resize(width, height int) {
	s.width = width
	s.height = height
	s.centerX = width / 2
	s.centerY = height / 2
	dx := float64(width) / 2.0
	dy := float64(height) / 2.0 * 2.0
	s.maxRadius = math.Sqrt(dx*dx+dy*dy) + 1.0
	s.buildGrid()
}
