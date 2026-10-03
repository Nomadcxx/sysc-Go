package animations

import "testing"

func TestFireEffectZeroHeightDoesNotPanic(t *testing.T) {
	f := NewFireEffect(40, 0, []string{"#ff0000", "#ffff00"})
	f.Update()
	if got := f.Render(); got != "" {
		t.Fatalf("expected empty render at height 0, got %q", got)
	}

	f = NewFireEffect(40, 12, []string{"#ff0000", "#ffff00"})
	f.Resize(40, 0)
	f.Update()
	if got := f.Render(); got != "" {
		t.Fatalf("expected empty render after resize to height 0, got %q", got)
	}
}

func TestFireTextEffectZeroHeightDoesNotPanic(t *testing.T) {
	const text = "HI"
	palette := []string{"#ff0000", "#ffff00"}

	f := NewFireTextEffect(40, 0, palette, text)
	f.Update()
	if got := f.Render(); got != "" {
		t.Fatalf("expected empty render at height 0, got %q", got)
	}
	f.SetText("OK")

	f = NewFireTextEffect(40, 12, palette, text)
	f.Resize(0, 0)
	f.Update()
	if got := f.Render(); got != "" {
		t.Fatalf("expected empty render after resize to 0x0, got %q", got)
	}
}
