package animations

import (
	"math"
	"strings"
	"testing"
	"time"
)

func TestAK47Cycle(t *testing.T) {
	if len(ak47Data) != 24000*12 {
		t.Fatal("unexpected baked model size")
	}
	last := 0.0
	for ms := 0; ms <= 8000; ms += 25 {
		rotation, blend := ak47Phase(time.Duration(ms) * time.Millisecond)
		if blend != 0 || rotation < last || rotation > 4*math.Pi {
			t.Fatalf("early morph/reversed turn at %dms: %g %g", ms, rotation, blend)
		}
		last = rotation
	}
	if math.Abs(last-4*math.Pi) > 1e-10 {
		t.Fatal("did not finish exactly two rotations")
	}
	_, peak := ak47Phase(10 * time.Second)
	if peak != 1 {
		t.Fatal("swirl does not reach full extent")
	}
	a := NewAK47Effect(80, 24, "kanagawa")
	first := a.Render()
	if first == "" || first != a.Render() {
		t.Fatal("empty/impure render")
	}
	if strings.Count(first, "\n") != 23 {
		t.Fatal("wrong row count")
	}
	dots := 0
	for _, r := range first {
		if r > 0x2800 && r <= 0x28ff {
			dots++
		}
	}
	if dots < 80 {
		t.Fatalf("only %d braille cells", dots)
	}
	a.UpdateElapsed(9 * time.Second)
	if first == a.Render() {
		t.Fatal("swirl did not change frame")
	}
	a.UpdateElapsed(3 * time.Second)
	if first != a.Render() {
		t.Fatal("loop does not reconstruct exactly")
	}
	b := NewAK47Effect(80, 24, "kanagawa")
	for i := 0; i < 120; i++ {
		a.Update()
	}
	for i := 0; i < 50; i++ {
		b.UpdateElapsed(120 * time.Millisecond)
	}
	if a.Render() != b.Render() {
		t.Fatal("frame cadence changes phase")
	}
	frozen := a.Render()
	a.UpdateElapsed(-time.Second)
	if a.Render() != frozen {
		t.Fatal("negative elapsed changed phase")
	}
	a.Resize(491, 90)
	a.Resize(80, 24)
	if a.Render() != frozen {
		t.Fatal("resize reset phase")
	}
	a.Reset()
	if a.Render() != first {
		t.Fatal("reset did not restore initial frame")
	}
	a.SetCellAspect(7, 16)
	a.Resize(1<<30, 1<<30)
	if a.canvas.width != 512 || a.canvas.height != 256 {
		t.Fatal("grid not bounded")
	}
	meta := GetEffectMetadata("ak-47")
	if meta == nil || meta.RequiresText {
		t.Fatal("missing standalone registry entry")
	}
}
