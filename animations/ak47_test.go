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

func TestAK47ArtworkCycle(t *testing.T) {
	for _, input := range []string{"", " \n\t", "\x1b[31mX", "X\x00", string([]byte{255}), strings.Repeat("X", 65537)} {
		if _, err := NewAK47TextEffect(80, 24, "kanagawa", input); err == nil {
			t.Fatalf("accepted invalid artwork %q", input[:min(20, len(input))])
		}
	}
	a, err := NewAK47RAMAEffect(80, 24, "kanagawa")
	if err != nil {
		t.Fatal(err)
	}
	first := a.Render()
	b := NewAK47Effect(80, 24, "kanagawa")
	for i := 0; i < 160; i++ {
		if a.Render() != b.Render() {
			t.Fatal("artwork changed the two rotations")
		}
		a.Update()
		b.Update()
	}
	a.UpdateElapsed(2 * time.Second)
	held := a.Render()
	if held == first {
		t.Fatal("artwork did not form")
	}
	a.UpdateElapsed(time.Second)
	if a.Render() != held {
		t.Fatal("artwork moves during hold")
	}
	a.Resize(120, 40)
	a.SetCellAspect(7, 16)
	if a.Render() == held {
		t.Fatal("resize/aspect ignored")
	}
	a.SetCellAspect(1, 2)
	a.Resize(80, 24)
	if a.Render() != held {
		t.Fatal("resize changed phase or target")
	}
	a.UpdateElapsed(3 * time.Second)
	if a.Render() != first {
		t.Fatal("14-second loop does not reconstruct")
	}
	c, err := NewAK47TextEffect(80, 24, "kanagawa", "X   X\n XXX")
	if err != nil {
		t.Fatal(err)
	}
	c.UpdateElapsed(10 * time.Second)
	if c.Render() == held {
		t.Fatal("selected artwork ignored")
	}
	c.Reset()
	if c.Render() != first {
		t.Fatal("reset did not restore rifle")
	}
	for _, id := range []string{"ak-47-rama", "ak-47-text"} {
		m := GetEffectMetadata(id)
		if m == nil || m.RequiresText != (id == "ak-47-text") {
			t.Fatalf("wrong metadata for %s", id)
		}
	}
	for _, id := range []string{"justice-cross", "cyberpunk-girl"} {
		if GetEffectMetadata(id) != nil {
			t.Fatalf("unfinished effect %s is released", id)
		}
	}
}
