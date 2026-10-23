package animations

import (
	_ "embed"
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

const ak47Period = meshSpinPeriod

type ak47Point = meshPoint

// ponytail: 24,000 fixed samples bound per-frame work; increase bake density
// only after qualifying the frame budget at the largest supported grid.
var ak47Points = sync.OnceValue(func() []ak47Point {
	return decodeMeshPoints(ak47Data)
})

// AK47Animation rotates the credited mesh twice, swirls and reforms it.
// Hosts own scheduling; Render does not advance active time.
type AK47Animation struct {
	*meshSpin
	art                 [][2]float64
	artWidth, artHeight int
}

func NewAK47Effect(width, height int, theme string) *AK47Animation {
	return &AK47Animation{meshSpin: newMeshSpin(width, height, theme, ak47Period)}
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

func ak47Phase(elapsed time.Duration) (rotation, blend float64) {
	return modelPhase(elapsed)
}
func (a *AK47Animation) rasterize() *LogoSpinAnimation {
	rotation, blend := ak47Phase(a.elapsed)
	if a.art != nil && a.elapsed >= 8*time.Second {
		rotation, blend = 4*math.Pi, 1
		if a.elapsed < 10*time.Second {
			blend = logoSmooth(float64(a.elapsed-8*time.Second) / float64(2*time.Second))
		} else if a.elapsed >= 12*time.Second {
			blend = 1 - logoSmooth(float64(a.elapsed-12*time.Second)/float64(2*time.Second))
		}
	}
	swirl := 0.0
	if a.elapsed >= 8*time.Second {
		swirl = float64(a.elapsed-8*time.Second) / float64(4*time.Second) * 2 * math.Pi
	}
	return rasterizeMeshSpinAtPitch(a.meshSpin, ak47Points(), rotation, blend, swirl, a.art, a.artWidth, a.artHeight, 1, .10*math.Sin(rotation))
}

func (a *AK47Animation) Render() string {
	return meshCredit(a.braille(a.rasterize()), a.canvas.width, "AK-47 / Lokeig / CC BY-NC 4.0")
}

// RenderFrame returns the current transparent per-dot image and attribution.
// Hosts choose how to present it; Render remains the portable ANSI output.
func (a *AK47Animation) RenderFrame() MeshFrame {
	return MeshFrame{Pixels: a.rasterize().rasterImage(), Credit: "AK-47 / Lokeig / CC BY-NC 4.0"}
}
