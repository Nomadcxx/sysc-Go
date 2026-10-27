package animations

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss/v2"
)

func TestCellColorCacheMatchesLipgloss(t *testing.T) {
	colors := []string{"#ff79c6", "#000000", "#FF79C6", "#fff", "196", "123456", "ff79c6", ""}
	chars := []rune{'x', '█', '⣿', '界'}
	var cache cellColorCache

	for _, color := range colors {
		for _, char := range chars {
			var got strings.Builder
			cache.writeRune(&got, color, char)
			want := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(string(char))
			if got.String() != want {
				t.Errorf("colored %q with %q = %q, want %q", string(char), color, got.String(), want)
			}
		}
	}

	if len(cache) != len(colors) {
		t.Errorf("cache holds %d colors, want %d", len(cache), len(colors))
	}
}

func TestWriteFrameRowMatchesLipgloss(t *testing.T) {
	chars := []rune{' ', ' ', 'x', ' ', '界', ' ', ' '}
	colors := []string{"", "#ff79c6", "#ff79c6", "", "#fff", "123456", ""}

	for _, styleSpaces := range []bool{false, true} {
		var got strings.Builder
		var cache cellColorCache
		writeFrameRow(&got, chars, colors, &cache, styleSpaces)

		var want strings.Builder
		for i, char := range chars {
			if colors[i] != "" && (styleSpaces || char != ' ') {
				want.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(colors[i])).Render(string(char)))
			} else {
				want.WriteRune(char)
			}
		}
		if got.String() != want.String() {
			t.Errorf("styleSpaces=%t: got %q, want %q", styleSpaces, got.String(), want.String())
		}
	}
}

func TestWritePrintRowMatchesLipgloss(t *testing.T) {
	cells := []printCell{
		{char: ' ', color: "#ff79c6"},
		{char: 'x', color: "196"},
		{text: "界"},
		{char: ' '},
		{char: ' '},
		{char: 'z', color: "#fff"},
		{text: ""},
	}

	var got strings.Builder
	var cache cellColorCache
	writePrintRow(&got, cells, &cache)

	want := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff79c6")).Render(" ") +
		lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("x") +
		"界  " + lipgloss.NewStyle().Foreground(lipgloss.Color("#fff")).Render("z")
	if got.String() != want {
		t.Fatalf("got %q, want %q", got.String(), want)
	}
}

func TestWriteBlankSpanWritesLongSpans(t *testing.T) {
	count := 2*len(blankCellBlock) + 5
	var got strings.Builder
	writeBlankSpan(&got, count)
	if want := strings.Repeat(" ", count); got.String() != want {
		t.Fatalf("got %d spaces, want %d", got.Len(), len(want))
	}
}

func TestDecryptRenderKeepsColoredCellsAndBlankRows(t *testing.T) {
	palette := []string{"#ff79c6"}
	d := NewDecryptEffect(DecryptConfig{
		Width: 5, Height: 2, Text: "x", Palette: palette, CiphertextColors: palette,
	})
	d.chars[0].visible = true
	d.chars[0].current = 'x'
	d.chars[0].color = palette[0]

	want := "  " + lipgloss.NewStyle().Foreground(lipgloss.Color(palette[0])).Render("x") + "  \n     "
	if got := d.Render(); got != want {
		t.Fatalf("colored frame = %q, want %q", got, want)
	}

	d.chars[0].visible = false
	if got, want := d.Render(), "     \n     "; got != want {
		t.Fatalf("blank frame = %q, want %q", got, want)
	}
}

func TestCellColorCacheIsBounded(t *testing.T) {
	var cache cellColorCache
	var got strings.Builder
	for i := 0; i < maxCellColorCacheEntries+64; i++ {
		color := fmt.Sprintf("#%06x", i)
		got.Reset()
		cache.writeRune(&got, color, 'x')
		want := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render("x")
		if got.String() != want {
			t.Fatalf("color %q = %q, want %q", color, got.String(), want)
		}
		if len(cache) > maxCellColorCacheEntries {
			t.Fatalf("cache holds %d colors, limit is %d", len(cache), maxCellColorCacheEntries)
		}
	}
}
