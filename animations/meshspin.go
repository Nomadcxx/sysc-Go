package animations

import (
	"math"
	"strings"
	"time"

	"github.com/Nomadcxx/sysc-Go/internal/mesh"
)

const meshSpinPeriod = 12 * time.Second

type meshPoint struct {
	position [3]float64
	normal   [3]float64
	cloud    [3]float64
}

type meshSpin struct {
	canvas                *LogoSpinAnimation
	elapsed               time.Duration
	cellWidth, cellHeight int
	colors                [16]uint32
	period                time.Duration
}

func decodeMeshPoints(data []byte) []meshPoint {
	samples, err := mesh.Decode(data)
	if err != nil {
		panic("invalid embedded surface mesh: " + err.Error())
	}
	points := make([]meshPoint, len(samples))
	for i, sample := range samples {
		point := &points[i]
		point.position = sample.Position
		point.normal = sample.Normal
		angle := float64(i) * 2.399963229728653
		radius := math.Sqrt((float64(i) + .5) / float64(len(points)))
		s, c := math.Sincos(angle)
		point.cloud = [3]float64{radius * c, .42 * radius * s, .35 * math.Sin(angle*.37)}
	}
	return points
}

func newMeshSpin(width, height int, theme string, period time.Duration) *meshSpin {
	m := &meshSpin{
		canvas:     NewLogoSpinEffect(LogoSpinConfig{Width: width, Height: height, Theme: theme}),
		cellWidth:  1,
		cellHeight: 2,
		period:     period,
	}
	for i := range m.colors {
		t := float64(i) / 15
		for k := 0; k < 3; k++ {
			low := m.canvas.litFill[k]*.55 + m.canvas.litKey[k]*.45
			c := low*(1-t) + m.canvas.litHigh[k]*t
			m.colors[i] = m.colors[i]<<8 | uint32(c+.5)
		}
	}
	return m
}

func (m *meshSpin) Update() { m.UpdateElapsed(50 * time.Millisecond) }

func (m *meshSpin) UpdateElapsed(d time.Duration) {
	if d > 0 {
		m.elapsed = (m.elapsed + d%m.period) % m.period
	}
}

func (m *meshSpin) Reset()          { m.elapsed = 0 }
func (m *meshSpin) Resize(w, h int) { m.canvas.Resize(w, h) }

func (m *meshSpin) SetCellAspect(w, h int) {
	if w > 0 && h > 0 && w <= 1024 && h <= 1024 {
		m.cellWidth, m.cellHeight = w, h
	}
}

func modelPhase(elapsed time.Duration) (rotation, blend float64) {
	seconds := float64(elapsed%meshSpinPeriod) / float64(time.Second)
	if seconds < 8 {
		turn := math.Floor(seconds / 4)
		return (turn + logoSmooth((seconds/4-turn)/.86)) * 2 * math.Pi, 0
	}
	return 4 * math.Pi, logoSmooth(1 - math.Abs(seconds-10)/2)
}

func renderMeshSpin(m *meshSpin, points []meshPoint, rotation, blend, swirl float64, art [][2]float64, artWidth, artHeight int, modelScale float64) string {
	l := m.canvas
	for i := range l.depth {
		l.depth[i] = math.Inf(-1)
	}
	sy, cy := math.Sincos(rotation)
	sx, cx := math.Sincos(.10 * math.Sin(rotation))
	ss, cs := math.Sincos(swirl)
	dotW, dotH := l.width*2, l.height*4
	// Fit a unit-radius cloud plus perspective margin in physical pixels.
	// Apply the same model-specific scale to mesh and cloud through the cycle.
	scale := math.Min(float64(l.width*m.cellWidth)*.40, float64((l.height-2)*m.cellHeight)*.72) * modelScale
	scaleX, scaleY := scale*2/float64(m.cellWidth), scale*4/float64(m.cellHeight)
	artScale := 0.0
	if art != nil {
		artScale = math.Min(float64(l.width)*.8/float64(artWidth), float64(l.height-2)*.8/float64(artHeight))
	}
	rot := func(p [3]float64) [3]float64 {
		x, z := p[0]*cy+p[2]*sy, -p[0]*sy+p[2]*cy
		return [3]float64{x, p[1]*cx - z*sx, p[1]*sx + z*cx}
	}
	for i, sample := range points {
		p, n := rot(sample.position), rot(sample.normal)
		if blend > 0 {
			cloud := [3]float64{sample.cloud[0]*cs - sample.cloud[1]/.42*ss, sample.cloud[0]*ss*.42 + sample.cloud[1]*cs, sample.cloud[2]}
			if art != nil {
				cloud = [3]float64{art[i][0] * float64(m.cellWidth) * artScale / scale, art[i][1] * float64(m.cellHeight) * artScale / scale, 0}
				for k := range n {
					n[k] *= 1 - blend
				}
				n[2] += blend
			}
			for k := range p {
				p[k] = p[k]*(1-blend) + cloud[k]*blend
			}
		}
		perspective := logoPerspective / (logoPerspective - p[2])
		x := int(math.Floor(float64(dotW)/2 + p[0]*scaleX*perspective))
		y := int(math.Floor(float64(dotH-8)/2 + p[1]*scaleY*perspective))
		if x < 0 || y < 0 || x >= dotW || y >= dotH-4 {
			continue
		}
		index := y*dotW + x
		if p[2] <= l.depth[index] {
			continue
		}
		// Both faces stay readable as the thin mesh turns through an edge view.
		light := clampLogo(math.Abs(n[0]*-.35+n[1]*-.45+n[2]*.82), 0, 1)
		shade := int(light*15 + .5)
		l.depth[index], l.crgb[index] = p[2], m.colors[shade]
	}
	return l.compose()
}

func meshCredit(frame string, width int, credit string) string {
	if width >= len(credit) {
		end := strings.LastIndexByte(frame, '\n') + 1
		frame = frame[:end] + strings.Repeat(" ", width-len(credit)) + credit
	}
	return frame
}
