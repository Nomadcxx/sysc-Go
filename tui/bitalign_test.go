package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

func bannerFont() *BitFont {
	return &BitFont{
		Name: "banner-test",
		Characters: map[string][]string{
			"H": {"# #", "###", "# #"},
			"I": {"#", "#", "#"},
			"W": {"####", "####", "####"},
		},
	}
}

func bannerOpts(text string, align, maxWidth int) TUIRenderOptions {
	return TUIRenderOptions{
		Font:      bannerFont(),
		Text:      text,
		Alignment: align,
		Color:     "#FFFFFF",
		Scale:     1,
		MaxWidth:  maxWidth,
	}
}

// inkIndent is the leftmost ink column across rendered rows.
func inkIndent(lines []string) int {
	min := -1
	for _, line := range lines {
		plain := stripANSI(line)
		if strings.TrimSpace(plain) == "" {
			continue
		}
		n := 0
		for _, r := range plain {
			if r != ' ' {
				break
			}
			n++
		}
		if min < 0 || n < min {
			min = n
		}
	}
	return min
}

func TestBitAlignmentPadsSingleLineToCanvas(t *testing.T) {
	const canvas = 40
	left := RenderBitText(bannerOpts("HI", 0, canvas))
	center := RenderBitText(bannerOpts("HI", 1, canvas))
	right := RenderBitText(bannerOpts("HI", 2, canvas))

	if len(left) == 0 || len(center) == 0 || len(right) == 0 {
		t.Fatal("expected rendered banner lines")
	}

	li, ci, ri := inkIndent(left), inkIndent(center), inkIndent(right)
	if !(li < ci && ci < ri) {
		t.Fatalf("canvas alignment indents left=%d center=%d right=%d; want left < center < right", li, ci, ri)
	}

	// Without a canvas wider than the glyphs, a single line cannot move.
	plainLeft := inkIndent(RenderBitText(bannerOpts("HI", 0, 0)))
	plainCenter := inkIndent(RenderBitText(bannerOpts("HI", 1, 0)))
	plainRight := inkIndent(RenderBitText(bannerOpts("HI", 2, 0)))
	if plainLeft != plainCenter || plainCenter != plainRight {
		t.Fatalf("MaxWidth 0 should not pad a single line, got left=%d center=%d right=%d", plainLeft, plainCenter, plainRight)
	}
}

func TestBitPreviewAlignmentFollowsControl(t *testing.T) {
	m := NewModel()
	m.width = 80
	m.bitCurrentFont = bannerFont()
	m.bitTextInput.SetValue("HI")
	m.bitScale = 1

	m.bitAlignment = 0
	m = m.updateBitPreview()
	left := inkIndent(m.bitPreviewLines)

	m.bitAlignment = 1
	m = m.updateBitPreview()
	center := inkIndent(m.bitPreviewLines)

	m.bitAlignment = 2
	m = m.updateBitPreview()
	right := inkIndent(m.bitPreviewLines)

	if !(left < center && center < right) {
		t.Fatalf("preview indents left=%d center=%d right=%d; want left < center < right", left, center, right)
	}
}

func TestBitAlignmentStillPadsNarrowerSiblingLine(t *testing.T) {
	// MaxWidth 0 keeps the old rule: pad a short line toward the widest sibling.
	center := RenderBitText(bannerOpts("W\nI", 1, 0))
	if inkIndent(center) != 0 {
		t.Fatalf("wider line indent = %d, want 0", inkIndent(center))
	}

	// The narrow second glyph is the only row whose every line is a single column
	// of ink after the wider block. Find a row whose ink starts after column 0.
	shifted := false
	for _, line := range center {
		plain := strings.TrimRight(stripANSI(line), " ")
		if strings.TrimSpace(plain) == "" {
			continue
		}
		if strings.TrimLeft(plain, " ") == "#" && inkOf(plain) > 0 {
			shifted = true
			break
		}
	}
	if !shifted {
		t.Fatalf("narrower line was not centered against its sibling:\n%s", strings.Join(center, "\n"))
	}
}

func inkOf(plain string) int {
	n := 0
	for _, r := range plain {
		if r != ' ' {
			break
		}
		n++
	}
	return n
}

// --- Issue #116: alignment padding must not reach disk. ---

// trailPad is the number of trailing space columns on a rendered row.
func trailPad(plain string) int {
	return utf8.RuneCountInString(plain) - utf8.RuneCountInString(strings.TrimRight(plain, " "))
}

func alignModel(t *testing.T, width, align int) Model {
	t.Helper()
	m := NewModel()
	m.width, m.height = width, 40
	m.bitCurrentFont = bannerFont()
	m.bitTextInput.SetValue("HI")
	m.bitScale = 1
	m.bitAlignment = align
	return m.updateBitPreview()
}

// TestBitPreviewCenterPaddingIsSymmetric: the preview kept the left half of the
// canvas indent and dropped the right half, so the banner looked off-centre even
// in the editor.
func TestBitPreviewCenterPaddingIsSymmetric(t *testing.T) {
	const width = 200
	m := alignModel(t, width, 1)

	canvas := width - 10
	seen := 0
	for _, line := range m.bitPreviewLines {
		plain := stripANSI(line)
		if strings.TrimSpace(plain) == "" {
			continue
		}
		seen++
		if got := utf8.RuneCountInString(plain); got != canvas {
			t.Errorf("preview row width = %d, want the full canvas %d: %q", got, canvas, plain)
		}
		left, right := inkOf(plain), trailPad(plain)
		if left == 0 {
			t.Fatalf("centre row has no indent: %q", plain)
		}
		// Shadow is off here, so the two halves must be equal or differ by the
		// single column an odd pad cannot split.
		if d := left - right; d < -1 || d > 1 {
			t.Errorf("centre padding left=%d right=%d, want symmetric: %q", left, right, plain)
		}
	}
	if seen == 0 {
		t.Fatal("no rendered rows")
	}
}

// TestBitPreviewEveryAlignmentFillsTheCanvas pins the same invariant for all
// three alignments so a future trim cannot come back on any of them.
func TestBitPreviewEveryAlignmentFillsTheCanvas(t *testing.T) {
	const width = 120
	for _, align := range []int{0, 1, 2} {
		m := alignModel(t, width, align)
		for _, line := range m.bitPreviewLines {
			plain := stripANSI(line)
			if got := utf8.RuneCountInString(plain); got != width-10 {
				t.Errorf("align=%d row width = %d, want %d: %q", align, got, width-10, plain)
			}
		}
	}
}

// TestBitExportLinesAreCanvasIndependent is the core of #116: the exported art
// must depend on the banner and the font only, never on the terminal the editor
// happened to be sized to.
func TestBitExportLinesAreCanvasIndependent(t *testing.T) {
	narrow := alignModel(t, 100, 1).bitExportLines()
	wide := alignModel(t, 200, 2).bitExportLines()

	if len(narrow) != len(wide) {
		t.Fatalf("row count differs with terminal width: %d vs %d", len(narrow), len(wide))
	}
	for i := range narrow {
		if narrow[i] != wide[i] {
			t.Fatalf("row %d depends on terminal width:\n  w=100 align=centre %q\n  w=200 align=right  %q", i, narrow[i], wide[i])
		}
	}
}

// TestBitExportLinesCarryNoAlignmentPadding: no leading or trailing spaces, and
// the widest row sits exactly at the banner ink width.
func TestBitExportLinesCarryNoAlignmentPadding(t *testing.T) {
	for _, align := range []int{0, 1, 2} {
		m := alignModel(t, 200, align)
		lines := m.bitExportLines()
		if len(lines) == 0 {
			t.Fatalf("align=%d produced no lines", align)
		}

		widest := 0
		for i, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}
			if strings.HasPrefix(line, " ") {
				t.Errorf("align=%d row %d has leading spaces: %q", align, i, line)
			}
			if strings.HasSuffix(line, " ") {
				t.Errorf("align=%d row %d has trailing spaces: %q", align, i, line)
			}
			widest = max(widest, utf8.RuneCountInString(line))
		}
		if want := 5; widest != want {
			t.Errorf("align=%d widest exported row = %d cols, want the %d-col banner", align, widest, want)
		}
	}
}

// TestBitExportLinesKeepShadowGlyphs: trailing padding is trimmed, ink is not.
func TestBitExportLinesKeepShadowGlyphs(t *testing.T) {
	m := alignModel(t, 200, 1)
	m.bitShadow = true
	m.bitShadowOffsetX = 2
	m.bitShadowOffsetY = 1
	m.bitShadowStyle = 0
	m = m.updateBitPreview()

	lines := m.bitExportLines()
	shadows := 0
	for _, line := range lines {
		if strings.HasSuffix(line, " ") {
			t.Errorf("shadowed export row has trailing spaces: %q", line)
		}
		for _, r := range line {
			if r != ' ' && r != '#' {
				shadows++
				break
			}
		}
	}
	if shadows == 0 {
		t.Errorf("expected shadow glyphs in the export, got:\n%s", strings.Join(lines, "\n"))
	}
}

// TestSaveBitArtWritesInkWidthNotCanvas: end-to-end through saveBitArt at a wide
// terminal with Right alignment, which used to write the full canvas indent and
// push the banner off screen in every text effect.
func TestSaveBitArtWritesInkWidthNotCanvas(t *testing.T) {
	home := t.TempDir()
	assets := filepath.Join(home, "sysc-Go", "assets")
	if err := os.MkdirAll(assets, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)

	m := alignModel(t, 200, 2)
	m.filenameInput.SetValue("banner.txt")
	m.exportTarget = 0

	saved, _ := m.saveBitArt()
	if saved.saveError != "" {
		t.Fatalf("save failed: %s", saved.saveError)
	}

	data, err := os.ReadFile(filepath.Join(assets, "banner.txt"))
	if err != nil {
		t.Fatal(err)
	}
	rows := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	for i, row := range rows {
		if len(row) != 5 {
			t.Errorf("saved row %d is %d cols, want the 5-col banner: %q", i, len(row), row)
		}
	}
	if len(rows) != 3 {
		t.Errorf("saved %d rows, want 3", len(rows))
	}
}

// TestExportBitArtTrimsTrailingPadding: ExportBitArt is also reached from the
// generic export prompt, so it must trim on its own.
func TestExportBitArtTrimsTrailingPadding(t *testing.T) {
	home := t.TempDir()
	assets := filepath.Join(home, "sysc-Go", "assets")
	if err := os.MkdirAll(assets, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)

	padded := []string{"\x1b[31m#\x1b[0m    ", "   \x1b[31m##\x1b[0m   ", "\x1b[31m#\x1b[0m    "}
	if err := ExportBitArt("padded.txt", padded, 0, false); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(assets, "padded.txt"))
	if err != nil {
		t.Fatal(err)
	}
	want := "#\n   ##\n#\n"
	if string(data) != want {
		t.Errorf("exported content = %q, want %q", string(data), want)
	}
}
