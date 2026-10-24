package animations

import (
	"math"
	"strings"
	"testing"
	"time"
)

func TestNapoleonDissolvesAndReforms(t *testing.T) {
	a, err := NewRotatingModelEffect("napoleon-bust", 100, 30, "dracula")
	if err != nil {
		t.Fatal(err)
	}
	source := a.spec.surface()
	first := a.Render()
	if first == "" || first != a.Render() || !strings.Contains(first, rotatingModelCatalog["napoleon-bust"].credit) {
		t.Fatal("initial sculpture is empty, nondeterministic or missing its attribution")
	}
	a.UpdateElapsed(10 * time.Second)
	if a.Render() == first {
		t.Fatal("sculpture did not dissolve into the field")
	}
	displaced := 0
	for i, p := range a.particles {
		for _, value := range p.position {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				t.Fatal("field produced a non-finite particle")
			}
		}
		if p.position != source[i].position {
			displaced++
		}
	}
	if displaced < len(source)*3/4 {
		t.Fatalf("only %d of %d particles left the sculpture", displaced, len(source))
	}
	a.UpdateElapsed(8 * time.Second)
	a.Render()
	for i, p := range a.particles {
		if p.position != source[i].position {
			t.Fatalf("particle %d did not return exactly to the sculpture", i)
		}
	}
	a.UpdateElapsed(2 * time.Second)
	if first != a.Render() {
		t.Fatal("twenty-second loop did not return to its first frame")
	}
	a.UpdateElapsed(-time.Second)
	if first != a.Render() {
		t.Fatal("negative elapsed time advanced the effect")
	}
	a.Update()
	if first == a.Render() {
		t.Fatal("ordinary Update did not advance the sculpture")
	}
	a.Reset()
	if first != a.Render() {
		t.Fatal("reset did not restore the sculpture")
	}
}

func TestHiggsFieldIsStandaloneAndFillsViewport(t *testing.T) {
	a := NewHiggsFieldEffect(100, 30, "dracula")
	first := a.Render()
	if first == "" || first != a.Render() || strings.Contains(first, "Napoleon") {
		t.Fatal("standalone field is empty, nondeterministic or starts with Napoleon")
	}
	for _, sec := range []int{0, 4, 8, 12, 16} {
		a.Reset()
		a.UpdateElapsed(time.Duration(sec) * time.Second)
		x0, x1, y0, y1, ok := brailleBounds(a.Render())
		if !ok || x1-x0 < 65 || y1-y0 < 10 {
			t.Fatalf("field at %ds covers (%d,%d)-(%d,%d), want at least 65 columns and 10 rows", sec, x0, y0, x1, y1)
		}
	}
	a.Reset()
	a.UpdateElapsed(8 * time.Second)
	if a.Render() == first {
		t.Fatal("standalone field did not evolve")
	}
	a.Reset()
	a.UpdateElapsed(higgsFieldPeriod)
	if a.Render() != first {
		t.Fatal("standalone field loop did not close")
	}
}

func TestHiggsFieldViewportAndResize(t *testing.T) {
	a := NewHiggsFieldEffect(100, 30, "kanagawa")
	a.SetCellAspect(9, 20)
	for _, size := range [][2]int{{40, 12}, {60, 18}, {160, 48}} {
		a.Resize(size[0], size[1])
		for step := 0; step < 20; step++ {
			a.Reset()
			a.UpdateElapsed(time.Duration(step) * time.Second)
			x0, x1, y0, y1, ok := brailleBounds(a.Render())
			if !ok || x0 < 1 || x1 >= size[0]-1 || y0 < 1 || y1 >= size[1]-1 {
				t.Fatalf("field step %d at %v: bounds (%d,%d)-(%d,%d)", step, size, x0, y0, x1, y1)
			}
		}
	}
}

func TestHiggsFieldPerformanceBudget(t *testing.T) {
	if testing.Short() {
		t.Skip("render budget check runs outside race instrumentation")
	}
	napoleon, err := NewRotatingModelEffect("napoleon-bust", 491, 90, "dracula")
	if err != nil {
		t.Fatal(err)
	}
	for name, a := range map[string]interface {
		Animation
		UpdateElapsed(time.Duration)
	}{"standalone": NewHiggsFieldEffect(491, 90, "dracula"), "napoleon": napoleon} {
		t.Run(name, func(t *testing.T) {
			a.UpdateElapsed(7 * time.Second)
			a.Render()
			for frame := 0; frame < 30; frame++ {
				a.Update()
				start := time.Now()
				a.Render()
				if elapsed := time.Since(start); elapsed > 30*time.Millisecond {
					t.Fatalf("field frame took %s, budget 30ms", elapsed)
				}
			}
		})
	}
}
