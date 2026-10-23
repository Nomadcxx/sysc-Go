package animations

import (
	_ "embed"
	"math"
	"sync"
	"time"
)

// Model-derived data: "low-poly Australian L2A1 SAR" (https://skfb.ly/oErMB)
// by notcplkerry, licensed under Creative Commons Attribution 4.0.
// See models/README.md for the full attribution and reproduction details.
//
//go:embed models/l2a1-sar.bin
var l2a1Data []byte

var l2a1Points = sync.OnceValue(func() []meshPoint { return decodeMeshPoints(l2a1Data) })

// Use more of the braille grid for the rifle's thin features without adding samples.
const l2a1DisplayScale = 1.12

// L2A1SARAnimation rotates the credited mesh twice, swirls and reforms it.
// Hosts own scheduling; Render does not advance active time.
type L2A1SARAnimation struct{ *meshSpin }

// NewL2A1SAREffect creates the looping Australian L2A1 SAR braille effect.
func NewL2A1SAREffect(width, height int, theme string) *L2A1SARAnimation {
	spin := newMeshSpin(width, height, theme, meshSpinPeriod)
	return &L2A1SARAnimation{meshSpin: spin}
}

func (a *L2A1SARAnimation) rasterize() *LogoSpinAnimation {
	rotation, blend := modelPhase(a.elapsed)
	rotation += math.Pi / 2 // Present the rifle side-on at the start of its cycle.
	swirl := 0.0
	if a.elapsed >= 8*time.Second {
		swirl = float64(a.elapsed-8*time.Second) / float64(4*time.Second) * 2 * math.Pi
	}
	return rasterizeMeshSpinAtPitch(a.meshSpin, l2a1Points(), rotation, blend, swirl, nil, 0, 0, l2a1DisplayScale, .10*math.Sin(rotation))
}

func (a *L2A1SARAnimation) Render() string {
	return meshCredit(a.braille(a.rasterize()), a.canvas.width, "L2A1 SAR / notcplkerry / CC BY 4.0")
}

// RenderFrame returns the current transparent per-dot image and attribution.
// Hosts choose how to present it; Render remains the portable ANSI output.
func (a *L2A1SARAnimation) RenderFrame() MeshFrame {
	return MeshFrame{Pixels: a.rasterize().rasterImage(), Credit: "L2A1 SAR / notcplkerry / CC BY 4.0"}
}

var _ Animation = (*L2A1SARAnimation)(nil)
