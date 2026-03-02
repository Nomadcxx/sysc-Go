package animations

import (
	"slices"
	"testing"
)

func testSkullPalette() []string {
	return GetSkullPalette("rama")
}

func TestSkullFinalPositionsStayInBounds(t *testing.T) {
	s := NewSkullEffect(90, 35, testSkullPalette(), "rama")
	if len(s.skullChars) == 0 {
		t.Fatal("expected skull characters to be parsed")
	}

	outOfBounds := 0
	for _, ch := range s.skullChars {
		if ch.finalX < 0 || ch.finalX >= s.width || ch.finalY < 0 || ch.finalY >= s.height {
			outOfBounds++
		}
	}

	if outOfBounds != 0 {
		t.Fatalf("expected all skull chars to be in bounds, got %d out-of-bounds chars", outOfBounds)
	}
}

func TestSkullIsHorizontallyCentered(t *testing.T) {
	s := NewSkullEffect(100, 35, testSkullPalette(), "rama")
	if len(s.skullChars) == 0 {
		t.Fatal("expected skull characters to be parsed")
	}

	minX, maxX := s.width, -1
	for _, ch := range s.skullChars {
		if ch.finalX < minX {
			minX = ch.finalX
		}
		if ch.finalX > maxX {
			maxX = ch.finalX
		}
	}

	leftMargin := minX
	rightMargin := (s.width - 1) - maxX
	delta := leftMargin - rightMargin
	if delta < 0 {
		delta = -delta
	}

	if delta > 1 {
		t.Fatalf("expected skull to be centered (left=%d right=%d), delta=%d", leftMargin, rightMargin, delta)
	}
}

func TestSkullIlluminationStartsRelativeToPhaseTransition(t *testing.T) {
	s := NewSkullEffect(100, 35, testSkullPalette(), "rama")

	var accent SkullChar
	found := false
	for _, ch := range s.skullChars {
		if ch.accentType == 1 {
			accent = ch
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected at least one top-accent skull char")
	}

	for i := range s.skullChars {
		s.skullChars[i].locked = true
		s.skullChars[i].currentY = float64(s.skullChars[i].finalY)
	}
	s.phase = PhaseDrip
	s.frameCount = 5

	s.Update()
	if s.phase != PhaseIllumination {
		t.Fatalf("expected transition to illumination, got phase=%v", s.phase)
	}

	if got := s.getSkullCharColor(accent); got != s.palette[1] {
		t.Fatalf("expected top accent color %q at illumination start, got %q", s.palette[1], got)
	}
}

func TestSkullSetTextEmptyDoesNotStallTextEntrance(t *testing.T) {
	s := NewSkullTextEffect(100, 35, testSkullPalette(), "rama", "HELLO")
	s.phase = PhaseTextEntrance
	s.frameCount = 400

	s.SetText("")
	s.Update()

	if s.phase == PhaseTextEntrance {
		t.Fatalf("expected phase to progress after SetText(\"\"), still in text entrance")
	}
}

func TestEffectRegistryIncludesSkullEffects(t *testing.T) {
	names := GetEffectNames()
	if !slices.Contains(names, "skull") {
		t.Fatal("expected effect registry to include skull")
	}
	if !slices.Contains(names, "skull-text") {
		t.Fatal("expected effect registry to include skull-text")
	}

	skullMeta := GetEffectMetadata("skull")
	if skullMeta == nil || skullMeta.RequiresText {
		t.Fatalf("unexpected skull metadata: %#v", skullMeta)
	}

	skullTextMeta := GetEffectMetadata("skull-text")
	if skullTextMeta == nil || !skullTextMeta.RequiresText {
		t.Fatalf("unexpected skull-text metadata: %#v", skullTextMeta)
	}
}
