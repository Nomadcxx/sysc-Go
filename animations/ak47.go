package animations

import (
	_ "embed"
	"encoding/binary"
	"math"
	"strings"
	"sync"
	"time"
)

// Model-derived data: "AK-47" by Lokeig, https://skfb.ly/oyAsI, CC BY-NC 4.0.
// See models/README.md for provenance, changes and reproduction.
//
//go:embed models/ak47.bin
var ak47Data []byte

const ak47Period = 12 * time.Second

type ak47Point struct{ position, normal, cloud [3]float64 }

// ponytail: 24,000 fixed samples bound per-frame work; increase bake density
// only after qualifying the frame budget at the largest supported grid.
var ak47Points = sync.OnceValue(func() []ak47Point {
	points := make([]ak47Point, len(ak47Data)/12)
	for i := range points {
		p := &points[i]
		for k := 0; k < 3; k++ {
			p.position[k] = float64(int16(binary.LittleEndian.Uint16(ak47Data[i*12+k*2:]))) / 32767
			p.normal[k] = float64(int16(binary.LittleEndian.Uint16(ak47Data[i*12+6+k*2:]))) / 32767
		}
		angle := float64(i) * 2.399963229728653
		radius := math.Sqrt((float64(i) + .5) / float64(len(points)))
		s, c := math.Sincos(angle)
		p.cloud = [3]float64{radius * c, .42 * radius * s, .35 * math.Sin(angle*.37)}
	}
	return points
})

// AK47Animation rotates the credited mesh twice, swirls and reforms it.
// Hosts own scheduling; Render does not advance active time.
type AK47Animation struct {
	canvas                *LogoSpinAnimation
	elapsed               time.Duration
	cellWidth, cellHeight int
	colors                [16]uint32
}

func NewAK47Effect(width, height int, theme string) *AK47Animation {
	a := &AK47Animation{canvas: NewLogoSpinEffect(LogoSpinConfig{Width: width, Height: height, Theme: theme}), cellWidth: 1, cellHeight: 2}
	// A fixed shade ramp keeps the model legible in each existing logo palette.
	for i := range a.colors {
		t := float64(i) / 15
		for k := 0; k < 3; k++ {
			low := a.canvas.litFill[k]*.55 + a.canvas.litKey[k]*.45
			c := low*(1-t) + a.canvas.litHigh[k]*t
			a.colors[i] = a.colors[i]<<8 | uint32(c+.5)
		}
	}
	return a
}
func (a *AK47Animation) Update() { a.UpdateElapsed(50 * time.Millisecond) }
func (a *AK47Animation) UpdateElapsed(d time.Duration) {
	if d > 0 {
		a.elapsed = (a.elapsed + d%ak47Period) % ak47Period
	}
}
func (a *AK47Animation) Reset()          { a.elapsed = 0 }
func (a *AK47Animation) Resize(w, h int) { a.canvas.Resize(w, h) }
func (a *AK47Animation) SetCellAspect(w, h int) {
	if w > 0 && h > 0 && w <= 1024 && h <= 1024 {
		a.cellWidth, a.cellHeight = w, h
	}
}
func ak47Phase(elapsed time.Duration) (rotation, blend float64) {
	seconds := float64(elapsed%ak47Period) / float64(time.Second)
	if seconds < 8 {
		turn := math.Floor(seconds / 4)
		return (turn + logoSmooth((seconds/4-turn)/.86)) * 2 * math.Pi, 0
	}
	return 4 * math.Pi, logoSmooth(1 - math.Abs(seconds-10)/2)
}
func (a *AK47Animation) Render() string {
	l := a.canvas
	for i := range l.depth {
		l.depth[i] = math.Inf(-1)
	}
	rotation, blend := ak47Phase(a.elapsed)
	sy, cy := math.Sincos(rotation)
	sx, cx := math.Sincos(.10 * math.Sin(rotation))
	swirl := 0.0
	if a.elapsed >= 8*time.Second {
		swirl = float64(a.elapsed-8*time.Second) / float64(4*time.Second) * 2 * math.Pi
	}
	ss, cs := math.Sincos(swirl)
	dotW, dotH := l.width*2, l.height*4
	// Fit a unit-radius cloud plus perspective margin in physical pixels.
	// The rifle and cloud retain the same scale through the whole cycle.
	scale := math.Min(float64(l.width*a.cellWidth)*.40, float64((l.height-2)*a.cellHeight)*.72)
	scaleX, scaleY := scale*2/float64(a.cellWidth), scale*4/float64(a.cellHeight)
	rot := func(p [3]float64) [3]float64 {
		x, z := p[0]*cy+p[2]*sy, -p[0]*sy+p[2]*cy
		return [3]float64{x, p[1]*cx - z*sx, p[1]*sx + z*cx}
	}
	for _, sample := range ak47Points() {
		p, n := rot(sample.position), rot(sample.normal)
		if blend > 0 {
			// Rotate in the cloud's unit circle, then restore its vertical scale.
			cloud := [3]float64{sample.cloud[0]*cs - sample.cloud[1]/.42*ss, sample.cloud[0]*ss*.42 + sample.cloud[1]*cs, sample.cloud[2]}
			for k := range p {
				p[k] = p[k]*(1-blend) + cloud[k]*blend
			}
		}
		perspective := 4.4 / (4.4 - p[2])
		x, y := int(math.Floor(float64(dotW)/2+p[0]*scaleX*perspective)), int(math.Floor(float64(dotH-8)/2+p[1]*scaleY*perspective))
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
		l.depth[index], l.crgb[index] = p[2], a.colors[shade]
	}
	frame := l.compose()
	credit := "AK-47 / Lokeig / CC BY-NC 4.0"
	if l.width >= len(credit) {
		end := strings.LastIndexByte(frame, '\n') + 1
		frame = frame[:end] + strings.Repeat(" ", l.width-len(credit)) + credit
	}
	return frame
}
