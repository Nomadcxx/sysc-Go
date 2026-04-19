package animations

import (
	"math"
	"slices"
	"strings"
	"testing"
)

func sonarTestPalette() []string {
	return GetSkullPalette("rama")
}

func TestSonarEffect_ImplementsAnimation(t *testing.T) {
	var _ Animation = NewSonarEffect(80, 24, sonarTestPalette(), "rama")
}

func TestSonarEffect_SpawnsMultipleConcurrentRings(t *testing.T) {
	s := NewSonarEffect(120, 40, sonarTestPalette(), "rama")

	// Scan (60) + grid (50) = 110 frames before hold; give hold plenty of
	// time to spawn several rings (interval=22, lifespan=110).
	for i := 0; i < 300; i++ {
		s.Update()
	}

	active := 0
	for _, r := range s.rings {
		if radius, _ := s.ringRadius(r); radius > 0 {
			active++
		}
	}

	if active < 2 {
		t.Fatalf("expected multiple concurrent sonar rings, got %d active", active)
	}
}

func TestSonarEffect_RunsScanThenGridThenHold(t *testing.T) {
	s := NewSonarEffect(120, 40, sonarTestPalette(), "rama")

	s.Update()
	if s.phase != PhaseDrip {
		t.Fatalf("expected PhaseDrip at frame 1, got %v", s.phase)
	}
	if s.scanRow() < 0 {
		t.Fatal("expected active scan row during scan phase")
	}

	for s.phase == PhaseDrip {
		s.Update()
	}
	if s.phase != PhaseGrid {
		t.Fatalf("expected PhaseGrid after scan, got %v", s.phase)
	}
	if lit := s.gridLitFraction(); lit <= 0 || lit >= 1 {
		// First frame of grid — lit fraction may be 0; advance a few frames.
		for i := 0; i < 10; i++ {
			s.Update()
		}
		if lit := s.gridLitFraction(); lit <= 0 {
			t.Fatalf("expected grid ignition to advance, got lit=%f", lit)
		}
	}

	for s.phase != PhaseHold {
		s.Update()
	}
	for i := 0; i < 80; i++ {
		s.Update()
	}
	if len(s.rings) == 0 {
		t.Fatal("expected rings to spawn during hold")
	}
}

func TestSonarEffect_RingsExpandToViewportBorders(t *testing.T) {
	w, h := 120, 40
	s := NewSonarEffect(w, h, sonarTestPalette(), "rama")

	cornerDX := float64(w) / 2.0
	cornerDY := float64(h) / 2.0 * 2.0 // 2:1 aspect correction
	cornerDist := math.Sqrt(cornerDX*cornerDX + cornerDY*cornerDY)

	if s.maxRadius < cornerDist {
		t.Fatalf("sonar maxRadius %f must reach viewport corner (dist=%f)",
			s.maxRadius, cornerDist)
	}
}

func TestSonarEffect_NoSkullGlyphsRendered(t *testing.T) {
	s := NewSonarEffect(120, 40, sonarTestPalette(), "rama")

	for i := 0; i < 60; i++ {
		s.Update()
	}
	out := s.Render()

	skullGlyphs := []rune{'▄', '█', '▟', '▛', '▜', '▙', '▀'}
	for _, g := range skullGlyphs {
		if strings.ContainsRune(out, g) {
			t.Fatalf("sonar render should not contain skull glyph %q", g)
		}
	}
}

func TestSonarEffect_NoBackgroundFill(t *testing.T) {
	s := NewSonarEffect(80, 24, sonarTestPalette(), "rama")
	for i := 0; i < 30; i++ {
		s.Update()
	}
	if strings.Contains(s.Render(), "\033[48;2;") {
		t.Fatal("sonar render must not emit background-fill escapes")
	}
}

func TestSonarEffect_RingsAgeAndExpire(t *testing.T) {
	s := NewSonarEffect(80, 24, sonarTestPalette(), "rama")

	for i := 0; i < 500; i++ {
		s.Update()
	}

	for _, r := range s.rings {
		age := s.frameCount - r.birthFrame
		if age < 0 || age >= s.lifespan {
			t.Fatalf("expired ring retained: age=%d lifespan=%d", age, s.lifespan)
		}
	}
}

func TestSonarEffect_RegisteredInRegistry(t *testing.T) {
	names := GetEffectNames()
	if !slices.Contains(names, "sonar") {
		t.Fatal("expected effect registry to include sonar")
	}
	meta := GetEffectMetadata("sonar")
	if meta == nil {
		t.Fatal("expected sonar metadata")
	}
	if meta.RequiresText {
		t.Fatal("sonar must not require text input")
	}
}

func TestSonarEffect_RendersNonEmpty(t *testing.T) {
	s := NewSonarEffect(80, 24, sonarTestPalette(), "rama")
	for i := 0; i < 30; i++ {
		s.Update()
	}
	out := s.Render()
	if !strings.Contains(out, "\033[38;2;") {
		t.Fatal("expected sonar render to contain at least one colour escape")
	}
}

func TestSonarEffect_ResetClearsState(t *testing.T) {
	s := NewSonarEffect(80, 24, sonarTestPalette(), "rama")
	for i := 0; i < 100; i++ {
		s.Update()
	}
	s.Reset()
	if s.frameCount != 0 {
		t.Fatalf("expected frameCount=0 after Reset, got %d", s.frameCount)
	}
}
