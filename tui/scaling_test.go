package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A later row shorter than row 0 used to panic in scaleBitmap: downscale
// indexed with the width of row 0. dosrebel has many such glyphs.
func TestScaleHalfShortRowDoesNotPanic(t *testing.T) {
	bitmap := []string{"█ ", " "}

	got := scaleCharacter(bitmap, 0.5)
	if len(got) != 1 || got[0] != "▀" {
		t.Fatalf("scale 0.5 = %#v, want [\"▀\"]", got)
	}
}

// Empty ANSI rows become nil pixel rows. Downscale must treat them as blank.
func TestScaleHalfNilRowDoesNotPanic(t *testing.T) {
	bitmap := []string{"██", "", "██"}

	got := scaleCharacter(bitmap, 0.5)
	want := []string{"▀", "▀"}
	if len(got) != len(want) {
		t.Fatalf("scale 0.5 = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("scale 0.5 = %#v, want %#v", got, want)
		}
	}
}

func TestScaleThreeLeavesGlyphUnchanged(t *testing.T) {
	bitmap := []string{"█ ", " "}
	got := scaleCharacter(bitmap, 3)
	if len(got) != len(bitmap) {
		t.Fatalf("scale 3.0 = %#v, want %#v", got, bitmap)
	}
	for i := range bitmap {
		if got[i] != bitmap[i] {
			t.Fatalf("scale 3.0 row %d = %q, want %q", i, got[i], bitmap[i])
		}
	}
}

func TestHalfScaleShortRowPreviewAndSave(t *testing.T) {
	font := &BitFont{
		Name: "ragged",
		Characters: map[string][]string{
			"A": {"█ ", " "},
		},
	}

	m, assetsDir := halfScaleModel(t, font, "A", "ragged-half")
	m = m.updateBitPreview()
	if len(m.bitPreviewLines) == 0 {
		t.Fatal("expected 0.5x preview lines")
	}
	plain := stripANSI(strings.Join(m.bitPreviewLines, "\n"))
	if !strings.Contains(plain, "▀") {
		t.Fatalf("preview = %q, want half-scale ink", plain)
	}
	if m.View() == "" {
		t.Fatal("expected BIT editor view")
	}

	saved, _ := m.saveBitArt()
	if saved.saveError != "" {
		t.Fatalf("save error: %s", saved.saveError)
	}
	if _, err := os.Stat(filepath.Join(assetsDir, "ragged-half.txt")); err != nil {
		t.Fatalf("saved banner: %v", err)
	}
}

func TestDosrebelHalfScalePreviewAndSave(t *testing.T) {
	font, err := loadEmbeddedBitFont("dosrebel")
	if err != nil {
		t.Fatalf("load dosrebel: %v", err)
	}

	m, _ := halfScaleModel(t, font, "A", "dosrebel-half")
	m = m.updateBitPreview()
	if m.View() == "" {
		t.Fatal("expected BIT editor view")
	}

	// dosrebel ink is not █/▀/▄, so half-scale can be blank. Save must still return.
	saved, _ := m.saveBitArt()
	if len(m.bitPreviewLines) > 0 && saved.saveError != "" {
		t.Fatalf("save error: %s", saved.saveError)
	}
}

func TestHalfScaleEveryEmbeddedFont(t *testing.T) {
	names := embeddedFontNames()
	if len(names) == 0 {
		t.Fatal("no embedded fonts")
	}

	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			font, err := loadEmbeddedBitFont(name)
			if err != nil {
				t.Fatalf("load %s: %v", name, err)
			}
			sample := "A"
			if _, ok := font.Characters[sample]; !ok {
				for ch := range font.Characters {
					if ch != "" {
						sample = ch
						break
					}
				}
			}
			for _, bitmap := range font.Characters {
				_ = scaleCharacter(bitmap, 0.5)
			}
			_ = RenderBitText(TUIRenderOptions{
				Font:  font,
				Text:  sample,
				Color: "#FFFFFF",
				Scale: 0.5,
			})
		})
	}
}

func halfScaleModel(t *testing.T, font *BitFont, text, filename string) (Model, string) {
	t.Helper()

	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	assetsDir := filepath.Join(tmpHome, "sysc-Go", "assets")
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		t.Fatalf("mkdir assets: %v", err)
	}

	m := NewModel()
	m.width = 120
	m.height = 40
	m.bitEditorMode = true
	m.bitCurrentFont = font
	m.bitTextInput.SetValue(text)
	m.bitScale = 0.5
	m.exportTarget = 0
	m.filenameInput.SetValue(filename)
	return m, assetsDir
}
