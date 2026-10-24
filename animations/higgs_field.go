package animations

import (
	"math"
	"time"
)

const higgsFieldPeriod = 20 * time.Second

// HiggsFieldAnimation excites a standalone particle field with travelling waves
// and vortex motion. It is field-inspired rather than a physics simulation.
type HiggsFieldAnimation struct {
	*meshSpin
	particles []meshPoint
}

// NewHiggsFieldEffect creates a looping particle field that spans the viewport.
func NewHiggsFieldEffect(width, height int, theme string) *HiggsFieldAnimation {
	return &HiggsFieldAnimation{
		meshSpin:  newMeshSpin(width, height, theme, higgsFieldPeriod),
		particles: make([]meshPoint, 18000),
	}
}

// ponytail: analytic flow closes the loop without integration drift. Use a
// bounded time-step simulation if particles later need interactive forces.
func higgsParticle(i, count int, phase, homeHeight float64) meshPoint {
	angle := float64(i)*2.399963229728653 + phase*3 + homeHeight*5
	s, c := math.Sincos(angle)
	radius := .30 + .85*math.Sqrt((float64(i)+.5)/float64(count))
	height := .22*math.Sin(3*angle-2*phase) + .20*math.Sin(radius*9-phase*4) + homeHeight*.20
	return meshPoint{
		position: [3]float64{radius * c, height, radius * s},
		normal:   [3]float64{c * .6, -.5, s * .6},
	}
}

func (a *HiggsFieldAnimation) rasterize() *LogoSpinAnimation {
	phase := 2 * math.Pi * float64(a.elapsed) / float64(higgsFieldPeriod)
	w, h := float64(a.canvas.width*a.cellWidth), float64((a.canvas.height-2)*a.cellHeight)
	scale := math.Min(w*.40, h*.72)
	spreadX, spreadY := w*.36/scale, h*.43/(scale*.45)
	for i := range a.particles {
		p := higgsParticle(i, len(a.particles), phase, 0)
		p.position[0] *= spreadX
		p.position[1] *= spreadY
		a.particles[i] = p
	}
	return rasterizeMeshSpinAtPose(a.meshSpin, a.particles, meshPose{}, 0, 0, nil, 0, 0, 1)
}

// Render returns the current particle field as ANSI Braille.
func (a *HiggsFieldAnimation) Render() string {
	return a.braille(a.rasterize())
}

// RenderFrame returns the current per-dot field image.
func (a *HiggsFieldAnimation) RenderFrame() MeshFrame {
	return MeshFrame{Pixels: a.rasterize().rasterImage()}
}

var _ Animation = (*HiggsFieldAnimation)(nil)
