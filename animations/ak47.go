package animations

import (
	_ "embed"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

// Model-derived data: "AK-47" by Lokeig, https://skfb.ly/oyAsI, CC BY-NC 4.0.
// See models/README.md for provenance, changes and reproduction.
//
//go:embed models/ak47.bin
var ak47Data []byte

//go:embed models/RAMA_SYSC.txt
var ak47RAMA string

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
	period                time.Duration
	art                   [][2]float64
	artWidth, artHeight   int
}

func NewAK47Effect(width, height int, theme string) *AK47Animation {
	a := &AK47Animation{canvas: NewLogoSpinEffect(LogoSpinConfig{Width: width, Height: height, Theme: theme}), cellWidth: 1, cellHeight: 2, period: ak47Period}
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

// NewAK47RAMAEffect forms the embedded RAMA artwork after two rotations.
func NewAK47RAMAEffect(width, height int, theme string) (*AK47Animation, error) {
	return NewAK47TextEffect(width, height, theme, ak47RAMA)
}

// NewAK47TextEffect forms the silhouette of occupied artwork cells in braille.
// It accepts up to 64 KiB of literal UTF-8 artwork, with tabs and CRLF lines.
func NewAK47TextEffect(width, height int, theme, text string) (*AK47Animation, error) {
	if len(text) > 64<<10 || !utf8.ValidString(text) {
		return nil, fmt.Errorf("AK-47 artwork must be valid UTF-8 within 64 KiB")
	}
	text = strings.ReplaceAll(text, "\r\n", "\n")
	type inkCell struct{ x, y, width int }
	var cells []inkCell
	x, y, left, top, right, bottom := 0, 0, 1024, 512, 0, 0
	for _, r := range text {
		switch r {
		case '\n':
			x, y = 0, y+1
		case '\t':
			x += 8 - x%8
		default:
			if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
				return nil, fmt.Errorf("AK-47 artwork contains control characters")
			}
			w := runewidth.RuneWidth(r)
			if w > 0 && !unicode.IsSpace(r) && r != '\u2800' {
				cells = append(cells, inkCell{x, y, w})
				left, top, right, bottom = min(left, x), min(top, y), max(right, x+w), max(bottom, y+1)
			}
			x += w
		}
		if x > 1024 || y >= 512 {
			return nil, fmt.Errorf("AK-47 artwork exceeds 1024 columns or 512 rows")
		}
	}
	if len(cells) == 0 {
		return nil, fmt.Errorf("AK-47 artwork has no visible cells")
	}
	a := NewAK47Effect(width, height, theme)
	a.period, a.artWidth, a.artHeight = 14*time.Second, right-left, bottom-top
	a.art = make([][2]float64, len(ak47Points()))
	// ponytail: fixed model density also bounds artwork work. Above 24,000 ink
	// cells, uniformly subsample; denser art needs a separately qualified budget.
	for i := range a.art {
		index := i % len(cells)
		if len(cells) > len(a.art) {
			index = i * len(cells) / len(a.art)
		}
		c := cells[index]
		u, v := math.Mod((float64(i)+.5)*.754877666, 1), math.Mod((float64(i)+.5)*.569840291, 1)
		a.art[i] = [2]float64{float64(c.x-left) + u*float64(c.width) - float64(a.artWidth)/2, float64(c.y-top) + v - float64(a.artHeight)/2}
	}
	return a, nil
}

func (a *AK47Animation) Update() { a.UpdateElapsed(50 * time.Millisecond) }
func (a *AK47Animation) UpdateElapsed(d time.Duration) {
	if d > 0 {
		a.elapsed = (a.elapsed + d%a.period) % a.period
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
	if a.art != nil && a.elapsed >= 8*time.Second {
		rotation, blend = 4*math.Pi, 1
		if a.elapsed < 10*time.Second {
			blend = logoSmooth(float64(a.elapsed-8*time.Second) / float64(2*time.Second))
		} else if a.elapsed >= 12*time.Second {
			blend = 1 - logoSmooth(float64(a.elapsed-12*time.Second)/float64(2*time.Second))
		}
	}
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
	artScale := 0.0
	if a.art != nil {
		artScale = math.Min(float64(l.width)*.8/float64(a.artWidth), float64(l.height-2)*.8/float64(a.artHeight))
	}
	rot := func(p [3]float64) [3]float64 {
		x, z := p[0]*cy+p[2]*sy, -p[0]*sy+p[2]*cy
		return [3]float64{x, p[1]*cx - z*sx, p[1]*sx + z*cx}
	}
	for i, sample := range ak47Points() {
		p, n := rot(sample.position), rot(sample.normal)
		if blend > 0 {
			// Rotate in the cloud's unit circle, then restore its vertical scale.
			cloud := [3]float64{sample.cloud[0]*cs - sample.cloud[1]/.42*ss, sample.cloud[0]*ss*.42 + sample.cloud[1]*cs, sample.cloud[2]}
			if a.art != nil {
				cloud = [3]float64{a.art[i][0] * float64(a.cellWidth) * artScale / scale, a.art[i][1] * float64(a.cellHeight) * artScale / scale, 0}
				for k := range n {
					n[k] *= 1 - blend
				}
				n[2] += blend
			}
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
