package animations

import (
	"math"
	"strings"
	"time"

	"github.com/Nomadcxx/sysc-Go/internal/mesh"
	"github.com/mattn/go-runewidth"
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
	return renderMeshSpinAtPitch(m, points, rotation, blend, swirl, art, artWidth, artHeight, modelScale, .10*math.Sin(rotation))
}

func renderMeshSpinAtPitch(m *meshSpin, points []meshPoint, rotation, blend, swirl float64, art [][2]float64, artWidth, artHeight int, modelScale, pitch float64) string {
	l := m.canvas
	sy, cy := math.Sincos(rotation)
	sx, cx := math.Sincos(pitch)
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
	sampleAt := func(i int, sample meshPoint) ([3]float64, [3]float64) {
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
		return p, n
	}
	// Measure fit while rasterizing, retaining the calibrated scale whenever it fits.
	// ponytail: overflowing frames get one extra pass instead of a point-sized
	// scratch buffer; cache projections if that extra pass becomes expensive.
	centerX, centerY := float64(dotW)/2, float64(dotH-8)/2
	for pass := 0; pass < 2; pass++ {
		for i := range l.depth {
			l.depth[i] = math.Inf(-1)
		}
		fit := 1.0
		for i, sample := range points {
			p, n := sampleAt(i, sample)
			perspective := logoPerspective / (logoPerspective - p[2])
			dx, dy := p[0]*scaleX*perspective, p[1]*scaleY*perspective
			if pass == 0 {
				for axis, offset := range [2]float64{dx, dy} {
					low, high := 2-centerX+1e-6, float64(dotW-2)-centerX-1e-6
					if axis == 1 {
						low, high = 4-centerY+1e-6, float64(dotH-4)-centerY-1e-6
					}
					if offset < low && offset < 0 {
						fit = math.Min(fit, math.Max(0, low/offset))
					}
					if offset > high && offset > 0 {
						fit = math.Min(fit, math.Max(0, high/offset))
					}
				}
			}
			x, y := int(math.Floor(centerX+dx)), int(math.Floor(centerY+dy))
			if x < 0 || y < 0 || x >= dotW || y >= dotH-4 {
				continue
			}
			index := y*dotW + x
			if p[2] <= l.depth[index] {
				continue
			}
			l.depth[index], l.crgb[index] = p[2], meshShade(n, p[2], l)
		}
		if fit == 1 {
			break
		}
		scaleX, scaleY = scaleX*fit, scaleY*fit
	}
	return l.compose()
}

func meshCredit(frame string, width int, credit string) string {
	creditWidth := runewidth.StringWidth(credit)
	if width >= creditWidth {
		end := strings.LastIndexByte(frame, '\n') + 1
		frame = frame[:end] + strings.Repeat(" ", width-creditWidth) + credit
	}
	return frame
}

// meshShade keeps broad surfaces visible after Braille cells average their dots.
// LogoSpin retains its own material lighting.
func meshShade(n [3]float64, depth float64, l *LogoSpinAnimation) uint32 {
	diffuse := clampLogo(n[0]*-.410+n[1]*-.564+n[2]*.718, 0, 1)
	fill := clampLogo(n[0]*.55+n[1]*.20-n[2]*.35, 0, 1) * .15 * (1 - diffuse)
	gloss := math.Pow(clampLogo(n[0]*-.220+n[1]*-.302+n[2]*.928, 0, 1), 18) * .25
	gain := clampLogo(.96+depth*.08, .82, 1)
	var out uint32
	for i := 0; i < 3; i++ {
		shadow := l.litBg[i] + (l.litHigh[i]-l.litBg[i])*.20
		key := l.litKey[i]*.25 + l.litHigh[i]*.75
		value := shadow*(1-diffuse) + key*diffuse
		value = (value*(1-fill)+l.litFill[i]*fill)*(1-gloss) + l.litHigh[i]*gloss
		out = out<<8 | uint32(clampLogo(l.litBg[i]+(value-l.litBg[i])*gain, 0, 255)+.5)
	}
	return out
}
