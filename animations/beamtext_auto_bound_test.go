package animations

import (
	"strings"
	"testing"
)

// Auto-sizing must not let the text size the canvas without end (audit
// sysc-2370); text past the bound is cropped, not a crash.
func TestBeamTextAutoSizeIsBounded(t *testing.T) {
	for _, text := range []string{
		strings.Repeat("#", 200000),
		strings.Repeat("ab\n", 5000),
	} {
		w, h := calculateTextDimensions(text)
		if w > maxAutoWidth || h > maxAutoHeight {
			t.Fatalf("auto canvas %dx%d exceeds %dx%d", w, h, maxAutoWidth, maxAutoHeight)
		}
		e := NewBeamTextEffect(BeamTextConfig{Text: text, Auto: true, BeamGradientStops: []string{"#ffffff"}, FinalGradientStops: []string{"#ffffff"}})
		for i := 0; i < 3; i++ {
			e.Update()
		}
		_ = e.Render()
	}
}
