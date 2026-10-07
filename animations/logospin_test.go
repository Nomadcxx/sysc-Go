package animations

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func sampleLogoPalette() []string {
	return GetLogoPalette("nord")
}

func TestLogoSpinSDFSigns(t *testing.T) {
	cross := logoSDF("cross")
	if cross == nil {
		t.Fatal("missing cross field")
	}
	grid := logoGridSize()

	// Center of the cross vertical bar must be inside, far corner outside.
	i := (grid/2)*grid + grid/2
	if cross[i] <= 0 {
		t.Errorf("cross centre distance = %v, want > 0", cross[i])
	}
	corner := 2*grid + 2
	if cross[corner] >= 0 {
		t.Errorf("cross corner distance = %v, want < 0", cross[corner])
	}

	sysc := logoSDF("sysc")
	if sysc == nil {
		t.Fatal("missing sysc field")
	}
	pos := 0
	for _, v := range sysc {
		if v > 0 {
			pos++
		}
	}
	// The wordmark must fill a believable plate area.
	if pos < 800 {
		t.Errorf("sysc field has only %d inside cells, want >= 800", pos)
	}
	if sysc[corner] >= 0 {
		t.Errorf("sysc corner distance = %v, want < 0", sysc[corner])
	}

	missing := logoSDF("no-such-shape")
	if missing != nil {
		t.Error("unknown shape must return nil field")
	}
}

func TestLogoSpinRenderBraille(t *testing.T) {
	l := NewLogoSpinEffect(LogoSpinConfig{Width: 40, Height: 15, Shape: "sysc", Palette: sampleLogoPalette()})
	seenBraille := 0
	seenColor := 0
	for i := 0; i < 60; i++ {
		l.Update()
		out := l.Render()
		if lines := strings.Split(strings.TrimRight(out, "\n"), "\n"); len(lines) != 15 {
			t.Fatalf("frame %d: got %d lines, want 15", i, len(lines))
		}
		if strings.Contains(out, "\033[38;2;") {
			seenColor++
		}
		for _, r := range out {
			if r >= 0x2800 && r <= 0x28FF {
				seenBraille++
			}
		}
	}
	if seenBraille == 0 {
		t.Error("no braille glyphs emitted across 60 frames")
	}
	if seenColor != 60 {
		t.Errorf("truecolor SGR present in %d/60 frames", seenColor)
	}
}

func TestLogoSpinDeterministic(t *testing.T) {
	mk := func() *LogoSpinAnimation {
		return NewLogoSpinEffect(LogoSpinConfig{Width: 30, Height: 12, Shape: "cross", Palette: sampleLogoPalette()})
	}
	a, b := mk(), mk()
	for i := 0; i < 25; i++ {
		a.Update()
		b.Update()
		if a.Render() != b.Render() {
			t.Fatalf("frames diverge at %d", i)
		}
	}
}

func TestLogoSpinResetReturnsToFrameZero(t *testing.T) {
	a := NewLogoSpinEffect(LogoSpinConfig{Width: 30, Height: 12, Shape: "cross", Palette: sampleLogoPalette()})
	first := a.Render()
	for i := 0; i < 40; i++ {
		a.Update()
		a.Render()
	}
	a.Reset()
	if got := a.Render(); got != first {
		t.Error("Reset did not restore the initial frame")
	}
}

func TestLogoSpinMinimumGridNoPanic(t *testing.T) {
	for _, shape := range []string{"sysc", "cross", "sysc-cross"} {
		l := NewLogoSpinEffect(LogoSpinConfig{Width: 21, Height: 24, Shape: shape, Palette: sampleLogoPalette()})
		for i := 0; i < 200; i++ {
			l.Update()
			out := l.Render()
			if utf8.RuneCountInString(out) == 0 {
				t.Fatalf("shape %s frame %d: empty render", shape, i)
			}
		}
	}
}

func TestLogoMorphMidDiffers(t *testing.T) {
	l := NewLogoSpinEffect(LogoSpinConfig{Width: 40, Height: 15, Shape: "sysc-cross", Palette: sampleLogoPalette()})
	single := NewLogoSpinEffect(LogoSpinConfig{Width: 40, Height: 15, Shape: "sysc", Palette: sampleLogoPalette()})

	mid := ""
	singleMid := ""
	for i := 0; i < logoMorphPeriodFrames()/2; i++ {
		l.Update()
		single.Update()
		mid = l.Render()
		singleMid = single.Render()
	}
	if mid == singleMid {
		t.Error("morph midpoint matches static shape render")
	}
	if !strings.Contains(mid, "\033[38;2;") {
		t.Error("morph midpoint lacks color")
	}
}

func TestLogoPaletteCoverage(t *testing.T) {
	for _, theme := range GetThemeNames() {
		p := GetLogoPalette(theme)
		if len(p) < 7 {
			t.Errorf("theme %q: logo palette has %d slots, want >= 7", theme, len(p))
		}
		for _, c := range p {
			if len(c) != 7 || c[0] != '#' {
				t.Errorf("theme %q: bad hex %q", theme, c)
			}
		}
	}
}

func TestLogoSpinRegistry(t *testing.T) {
	for _, id := range []string{"sysc-logo", "cross-logo", "logo-morph"} {
		if GetEffectMetadata(id) == nil {
			t.Errorf("registry missing %q", id)
		}
	}
}

func TestLogoJusticeShapeBakes(t *testing.T) {
	field := logoSDF("justice")
	if field == nil {
		t.Fatal("missing justice field")
	}
	pos := 0
	for _, v := range field {
		if v > 0 {
			pos++
		}
	}
	if pos < 300 {
		t.Errorf("justice field has only %d inside cells, want >= 300", pos)
	}
	if field[logoGrid*logoGrid-1] >= 0 {
		t.Errorf("justice corner distance = %v, want < 0", field[logoGrid*logoGrid-1])
	}
	l := NewLogoSpinEffect(LogoSpinConfig{Width: 40, Height: 15, Shape: "justice"})
	l.Update()
	if out := l.Render(); !strings.Contains(out, string(rune(logoBrailleRuneBase))) && out == "" {
		t.Fatal("empty justice render")
	}
}
