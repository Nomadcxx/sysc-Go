package animations

import (
	"regexp"
	"strconv"
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
	ids := []string{"sysc-logo", "cross-logo", "justice-cross", "logo-morph"}
	for _, id := range ids {
		if GetEffectMetadata(id) == nil {
			t.Errorf("registry missing %q", id)
		}
	}
	if got := LibraryVersion; got != "1.0.6" {
		t.Errorf("LibraryVersion = %q, want 1.0.6", got)
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

func TestLogoEdgeGeometry(t *testing.T) {
	justice := logoEdges("justice")
	// 16-vertex loop: 16 front + 16 back + 16 struts, 3 facet lines x2.
	if len(justice) != 54 {
		t.Errorf("justice edge count = %d, want 54", len(justice))
	}
	if logoEdges("no-such-shape") != nil {
		t.Error("unknown shape must return nil edges")
	}
	if len(logoEdges("sysc")) == 0 {
		t.Error("sysc produced no edges")
	}
}

func TestLogoStyleModes(t *testing.T) {
	for _, style := range []string{"", "edge", "ring", "wire", "solid"} {
		l := NewLogoSpinEffect(LogoSpinConfig{Width: 40, Height: 15, Shape: "justice", Style: style})
		l.Update()
		if out := l.Render(); out == "" {
			t.Errorf("style %q: empty render", style)
		}
	}
	m := NewLogoSpinEffect(LogoSpinConfig{Width: 40, Height: 15, Shape: "sysc-cross", Style: "edge"})
	m.Update()
	if out := m.Render(); out == "" {
		t.Error("morph with edge style fell silent")
	}
	if m.mode != logoModeRing {
		t.Error("morph should downgrade edge style to ring")
	}
}

// logoSpinCell is one lit braille dot cell: a true-colour escape, the glyph,
// then a reset. Components are matched as plain decimal with no sign and no
// padding, which is exactly what fmt's %d produced before compose() started
// writing the bytes itself.
var logoSpinCell = regexp.MustCompile("\x1b\\[38;2;(0|[1-9][0-9]*);(0|[1-9][0-9]*);(0|[1-9][0-9]*)m([\u2800-\u28ff])\x1b\\[0m")

// A frame is spaces and newlines with one such escape per lit cell. If the
// hand-rolled formatter drifts — a dropped semicolon, a padded component, an
// unterminated escape — the leftovers stop being layout and this fails.
func TestLogoSpinFrameIsLayoutPlusTruecolourCells(t *testing.T) {
	m := NewLogoSpinEffect(LogoSpinConfig{Width: 80, Height: 30, Shape: "sysc", Theme: "dracula"})
	frame := m.Render()

	if strings.Contains(frame, "%!") {
		t.Fatalf("frame carries a fmt error token: %q", frame)
	}
	cells := logoSpinCell.FindAllStringIndex(frame, -1)
	if len(cells) == 0 {
		t.Fatal("no lit cells, so the escape path went untested")
	}

	var residue strings.Builder
	end := 0
	for _, cell := range cells {
		residue.WriteString(frame[end:cell[0]])
		end = cell[1]
		parts := logoSpinCell.FindStringSubmatch(frame[cell[0]:cell[1]])
		for _, comp := range parts[1:4] {
			n, err := strconv.Atoi(comp)
			if err != nil || n > 255 {
				t.Errorf("colour component %q is not a byte in %q", comp, frame[cell[0]:cell[1]])
			}
		}
		if parts[4] == "" {
			t.Errorf("cell %q carries no glyph", frame[cell[0]:cell[1]])
		}
	}
	residue.WriteString(frame[end:])

	if left := strings.TrimLeft(residue.String(), " \n"); left != "" {
		t.Errorf("bytes between or after the cells: %q", left)
	}
}
