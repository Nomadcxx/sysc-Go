package animations

import (
	"regexp"
	"strings"
	"testing"
)

func stripDecryptANSI(s string) string {
	return regexp.MustCompile(`\x1b\[[0-9;]*m`).ReplaceAllString(s, "")
}

func issuePalette() []string {
	return []string{"#ff0000", "#00ff00"}
}

// TestDecryptSetTextReplacesText: SetText must replace the character slice and
// restart the phase, not append to it.
func TestDecryptSetTextReplacesText(t *testing.T) {
	d := NewDecryptEffect(DecryptConfig{
		Width: 20, Height: 3, Text: "OLDOLDOLDOLD",
		Palette: issuePalette(), CiphertextColors: issuePalette(), TypingSpeed: 1,
	})
	for i := 0; i < 400; i++ {
		d.Update()
	}

	d.SetText("NEW")
	if got, want := len(d.chars), len([]rune("NEW")); got != want {
		t.Fatalf("len(chars)=%d after SetText, want %d", got, want)
	}

	for i := 0; i < 5000 && d.phase != "complete"; i++ {
		d.Update()
	}
	if d.phase != "complete" {
		t.Fatalf("never reached the complete phase, phase=%q", d.phase)
	}
	out := stripDecryptANSI(d.Render())
	if !strings.Contains(out, "NEW") {
		t.Fatalf("render does not contain NEW:\n%s", out)
	}
	if strings.Contains(out, "OLD") {
		t.Fatalf("render still contains OLD:\n%s", out)
	}
}

// TestDecryptParseHexColor: a #rrggbb string is 7 bytes, not 8.
func TestDecryptParseHexColor(t *testing.T) {
	d := NewDecryptEffect(DecryptConfig{
		Width: 10, Height: 3, Text: "A",
		Palette: issuePalette(), CiphertextColors: issuePalette(),
	})
	if r, g, b := d.parseHexColor("#eda000"); r != 0xed || g != 0xa0 || b != 0x00 {
		t.Fatalf("parseHexColor(#eda000) = (%#x, %#x, %#x), want (0xed, 0xa0, 0x00)", r, g, b)
	}

	gradient := d.createSimpleGradient("#ffffff", "#eda000", 15)
	if got := gradient[len(gradient)-1]; got != "#eda000" {
		t.Fatalf("last gradient entry = %q, want #eda000", got)
	}
}

// TestDecryptTypingSpeedZeroStillTypes: the zero value must behave like 1.
func TestDecryptTypingSpeedZeroStillTypes(t *testing.T) {
	d := NewDecryptEffect(DecryptConfig{
		Width: 20, Height: 3, Text: "AB",
		Palette: issuePalette(), CiphertextColors: issuePalette(),
	})
	for i := 0; i < 200; i++ {
		d.Update()
	}
	if len(d.getVisibleChars()) == 0 {
		t.Fatal("no characters became visible with an unset TypingSpeed")
	}
}

// TestDecryptSetTextEmptyProducesNoChars: an empty SetText must not panic.
func TestDecryptSetTextEmptyProducesNoChars(t *testing.T) {
	d := NewDecryptEffect(DecryptConfig{
		Width: 20, Height: 3, Text: "AB",
		Palette: issuePalette(), CiphertextColors: issuePalette(), TypingSpeed: 1,
	})
	d.SetText("")
	if len(d.chars) != 0 {
		t.Fatalf("len(chars)=%d after SetText(\"\"), want 0", len(d.chars))
	}
	d.Update()
	d.Render()
}
