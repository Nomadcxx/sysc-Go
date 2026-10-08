package animations

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func newLayoutDecrypt(text string, w, h int) *DecryptEffect {
	p := []string{"#ff79c6", "#bd93f9", "#8be9fd"}
	return NewDecryptEffect(DecryptConfig{
		Width: w, Height: h, Text: text,
		Palette: p, CiphertextColors: p, TypingSpeed: 1,
	})
}

// TestDecryptLayoutASCIIUnchanged pins the pure-ASCII layout so examples do not move.
func TestDecryptLayoutASCIIUnchanged(t *testing.T) {
	d := newLayoutDecrypt("HELLO\nAB", 20, 5)

	rows := map[int][]rune{}
	xs := map[int][]int{}
	for _, c := range d.chars {
		rows[c.y] = append(rows[c.y], c.original)
		xs[c.y] = append(xs[c.y], c.x)
	}
	startY := (5 - 2) / 2 // two lines, centred in 5 rows
	want := map[int]string{startY: "HELLO", startY + 1: "AB"}
	for y, text := range want {
		if got := string(rows[y]); got != text {
			t.Errorf("row %d is %q, want %q", y, got, text)
		}
		// Each line is centred on its own: startX = (20-len)/2.
		startX := (20 - len([]rune(text))) / 2
		for i, x := range xs[y] {
			if x != startX+i {
				t.Errorf("row %d char %d at x=%d, want %d", y, i, x, startX+i)
			}
		}
	}
}

// TestDecryptLayoutMultiByteRunes checks every rune occupies one column.
func TestDecryptLayoutMultiByteRunes(t *testing.T) {
	d := newLayoutDecrypt("█▀▄é", 20, 3)
	if len(d.chars) != 4 {
		t.Fatalf("len(chars)=%d, want 4", len(d.chars))
	}
	for i, c := range d.chars {
		if want := 8 + i; c.x != want {
			t.Errorf("char %d at x=%d, want %d", i, c.x, want)
		}
		if c.y != 1 {
			t.Errorf("char %d at y=%d, want 1", i, c.y)
		}
	}
}

// TestDecryptLayoutSYSCBanner checks the default non-ASCII asset decrypts back
// into the source art, row for row.
func TestDecryptLayoutSYSCBanner(t *testing.T) {
	data, err := os.ReadFile("../assets/SYSC.txt")
	if err != nil {
		t.Skipf("SYSC.txt unavailable: %v", err)
	}
	text := strings.TrimSpace(string(data))
	wantRunes := len([]rune(strings.ReplaceAll(text, "\n", "")))

	d := newLayoutDecrypt(text, 120, 30)
	if len(d.chars) != wantRunes {
		t.Fatalf("len(chars)=%d, want %d non-newline runes", len(d.chars), wantRunes)
	}

	for i := range d.chars {
		d.chars[i].visible = true
		d.chars[i].current = d.chars[i].original
	}
	out := regexp.MustCompile(`\x1b\[[0-9;]*m`).ReplaceAllString(d.Render(), "")
	lines := strings.Split(out, "\n")

	startY := (30 - len(strings.Split(text, "\n"))) / 2
	for i, in := range strings.Split(text, "\n") {
		row := i + startY
		if row >= len(lines) {
			t.Fatalf("rendered row %d out of range", row)
		}
		got := strings.TrimSpace(lines[row])
		if got != strings.TrimSpace(in) {
			t.Errorf("row %d:\n got %q\nwant %q", row, got, strings.TrimSpace(in))
		}
	}
}

// TestDecryptCRLF checks line endings are normalised, so no \r glyph is laid out.
func TestDecryptCRLF(t *testing.T) {
	d := newLayoutDecrypt("AB\r\nCD\r\n", 20, 5)
	for i, c := range d.chars {
		if c.original == '\r' {
			t.Fatalf("char %d is a carriage return", i)
		}
	}
}
