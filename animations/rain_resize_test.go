package animations

import (
	"strings"
	"testing"
)

// TestRainResizeKeepsDropsBounded guards against Resize stacking a fresh batch
// of drops on top of the existing ones on every resize event.
func TestRainResizeKeepsDropsBounded(t *testing.T) {
	r := NewRainEffect(120, 40, GetRainPalette("dracula"))
	for i := 0; i < 50; i++ {
		r.Update()
	}
	for i := 0; i < 100; i++ {
		r.Resize(120+i%2, 40)
		if len(r.drops) > r.maxDrops {
			t.Fatalf("after resize %d: len(drops)=%d exceeds maxDrops=%d", i, len(r.drops), r.maxDrops)
		}
		r.Update()
	}
	for i := 0; i < 200; i++ {
		r.Update()
	}
	if len(r.drops) > r.maxDrops {
		t.Errorf("after settling: len(drops)=%d exceeds maxDrops=%d", len(r.drops), r.maxDrops)
	}
}

// TestRainResizeShrinkKeepsDropsOnCanvas checks that resizing down re-seeds
// instead of leaving drops outside the new width.
func TestRainResizeShrinkKeepsDropsOnCanvas(t *testing.T) {
	r := NewRainEffect(200, 50, GetRainPalette("dracula"))
	for i := 0; i < 100; i++ {
		r.Update()
	}
	r.Resize(40, 10)
	if len(r.drops) > 80 {
		t.Errorf("len(drops)=%d exceeds maxDrops=%d", len(r.drops), r.maxDrops)
	}
	for i, d := range r.drops {
		if d.X < 0 || d.X >= 40 {
			t.Fatalf("drop %d has X=%d outside [0,40)", i, d.X)
		}
	}
}

// TestRainResizeRendersExactSize checks the canvas is rebuilt at the new size.
func TestRainResizeRendersExactSize(t *testing.T) {
	r := NewRainEffect(120, 40, GetRainPalette("dracula"))
	r.Resize(30, 7)
	if got := len(strings.Split(r.Render(), "\n")); got != 7 {
		t.Errorf("Render() has %d lines, want 7", got)
	}
}
