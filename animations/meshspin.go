package animations

import (
	"image"
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
	light                 []float64 // diffuse term per dot, read by stipple
}

// MeshFrame contains one rendered mesh at two pixels per terminal column and
// four pixels per terminal row. Unoccupied dots have zero alpha.
type MeshFrame struct {
	Pixels *image.NRGBA
	Credit string
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
	return m.braille(rasterizeMeshSpinAtPitch(m, points, rotation, blend, swirl, art, artWidth, artHeight, modelScale, pitch))
}

func rasterizeMeshSpinAtPitch(m *meshSpin, points []meshPoint, rotation, blend, swirl float64, art [][2]float64, artWidth, artHeight int, modelScale, pitch float64) *LogoSpinAnimation {
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
	if len(m.light) != len(l.depth) {
		m.light = make([]float64, len(l.depth))
	}
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
			light := meshDiffuse(n)
			l.depth[index], l.crgb[index], m.light[index] = p[2], meshShade(light, l), light
		}
		if fit == 1 {
			break
		}
		scaleX, scaleY = scaleX*fit, scaleY*fit
	}
	return l
}

func meshCredit(frame string, width int, credit string) string {
	creditWidth := runewidth.StringWidth(credit)
	if width >= creditWidth {
		end := strings.LastIndexByte(frame, '\n') + 1
		frame = frame[:end] + strings.Repeat(" ", width-creditWidth) + credit
	}
	return frame
}

// ponytail: a fixed screen-space tile avoids random flicker but repeats every
// four dots; use a larger blue-noise tile if that pattern becomes distracting.
var meshStippleThresholds = [4][4]float64{
	{0, 8, 2, 10}, {12, 4, 14, 6}, {3, 11, 1, 9}, {15, 7, 13, 5},
}

// braille composes the mesh as Braille text after stippling its shadows.
func (m *meshSpin) braille(l *LogoSpinAnimation) string {
	m.stipple(l)
	return l.compose()
}

// stipple drops shadowed interior dots in an ordered pattern. A Braille cell
// can show only one colour, so averaging leaves curved surfaces as flat
// patches; dot density is the cell's remaining channel for shading. Silhouette
// dots stay, keeping outlines intact at small sizes. Even lit interiors retain
// gaps so the surface reads as particles. Image output keeps every dot's colour.
func (m *meshSpin) stipple(l *LogoSpinAnimation) {
	w := l.width * 2
	occupied := func(i int) bool { return i >= 0 && i < len(l.depth) && !math.IsInf(l.depth[i], -1) }
	// Mark first and clear afterwards, so dropped dots never read as silhouette.
	for i := range l.depth {
		x := i % w
		if !occupied(i) || x == 0 || x == w-1 || !occupied(i-1) || !occupied(i+1) || !occupied(i-w) || !occupied(i+w) {
			continue
		}
		coverage := .25 + .60*m.light[i]
		if coverage < (meshStippleThresholds[(i/w)%4][x%4]+.5)/16 {
			m.light[i] = -1
		}
	}
	for i := range l.depth {
		if occupied(i) && m.light[i] < 0 {
			l.depth[i] = math.Inf(-1)
		}
	}
}

func meshDiffuse(n [3]float64) float64 {
	return clampLogo(n[0]*-.410+n[1]*-.564+n[2]*.718, 0, 1)
}

// meshShade restores the theme-coloured ramp used before the mesh shader.
// Density carries shadow contrast without dimming particles into the background.
func meshShade(diffuse float64, l *LogoSpinAnimation) uint32 {
	light := .14 + .86*diffuse
	var out uint32
	for i := 0; i < 3; i++ {
		low := l.litFill[i]*.55 + l.litKey[i]*.45
		out = out<<8 | uint32(low*(1-light)+l.litHigh[i]*light+.5)
	}
	return out
}
