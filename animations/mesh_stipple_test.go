package animations

import (
	"math"
	"testing"
)

// stippleCanvas fills a block of dots, all equally lit.
func stippleCanvas(light float64) *meshSpin {
	m := newMeshSpin(16, 8, "dracula", rotatingModelPeriod)
	m.light = make([]float64, len(m.canvas.depth))
	for i := range m.canvas.depth {
		m.canvas.depth[i], m.light[i] = 0, light
	}
	return m
}

func TestMeshStippleThinsShadowedInteriorOnly(t *testing.T) {
	m := stippleCanvas(0)
	m.stipple(m.canvas)
	w := m.canvas.width * 2
	kept, dropped := 0, 0
	for i, d := range m.canvas.depth {
		x, y := i%w, i/w
		silhouette := x == 0 || y == 0 || x == w-1 || y == len(m.canvas.depth)/w-1
		if silhouette && math.IsInf(d, -1) {
			t.Fatalf("silhouette dot (%d,%d) was dropped", x, y)
		}
		if !silhouette {
			if math.IsInf(d, -1) {
				dropped++
			} else {
				kept++
			}
		}
	}
	if dropped == 0 || kept == 0 {
		t.Fatalf("shadowed interior kept %d and dropped %d dots, want a partial stipple", kept, dropped)
	}
}

func TestMeshStippleShadesWithParticleDensity(t *testing.T) {
	previous := 0.0
	for _, light := range []float64{0, .25, .5, 1} {
		m := stippleCanvas(light)
		m.stipple(m.canvas)
		kept, total := 0, 0
		w, h := m.canvas.width*2, m.canvas.height*4
		for y := 1; y < h-1; y++ {
			for x := 1; x < w-1; x++ {
				total++
				if !math.IsInf(m.canvas.depth[y*w+x], -1) {
					kept++
				}
			}
		}
		density := float64(kept) / float64(total)
		if density <= previous || density < .20 || density > .90 {
			t.Fatalf("light %.2f: density %.3f after %.3f; want increasing particle density in [.20,.90]", light, density, previous)
		}
		previous = density
	}
}

func TestMeshRenderFrameIsNotStippled(t *testing.T) {
	a, err := NewRotatingModelEffect("napoleon-bust", 60, 18, "dracula")
	if err != nil {
		t.Fatal(err)
	}
	frame := a.RenderFrame().Pixels
	occupied := 0
	for i := 3; i < len(frame.Pix); i += 4 {
		if frame.Pix[i] != 0 {
			occupied++
		}
	}
	a.Render()
	stippled := 0
	for _, d := range a.canvas.depth {
		if !math.IsInf(d, -1) {
			stippled++
		}
	}
	if stippled >= occupied {
		t.Fatalf("braille render kept %d of %d dots, want shadowed dots thinned", stippled, occupied)
	}
}
