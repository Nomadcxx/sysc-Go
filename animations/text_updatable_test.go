package animations

import (
	"strings"
	"testing"
)

func textUpdatableTestPalette() []string {
	return []string{
		"#000000", "#1a0000", "#330000", "#4d0000",
		"#660000", "#800000", "#990000", "#cc0000",
		"#ff0000", "#ff3300",
	}
}

func TestTextBasedEffectsSatisfyTextUpdatable(t *testing.T) {
	w, h := 40, 20
	p := textUpdatableTestPalette()
	text := "TEST"

	effects := map[string]interface{}{
		"fire-text":  NewFireTextEffect(w, h, p, text),
		"matrix-art": NewMatrixArtEffect(w, h, p, text),
		"rain-art":   NewRainArtEffect(w, h, p, text),
		"ring-text":  NewRingTextEffect(RingTextConfig{Width: w, Height: h, Text: text}),
		"blackhole":  NewBlackholeEffect(BlackholeConfig{Width: w, Height: h, Text: text}),
		"beam-text":  NewBeamTextEffect(BeamTextConfig{Width: w, Height: h, Text: text}),
		"pour":       NewPourEffect(PourConfig{Width: w, Height: h, Text: text}),
		"print":      NewPrintEffect(PrintConfig{Width: w, Height: h, Text: text}),
		"decrypt":    NewDecryptEffect(DecryptConfig{Width: w, Height: h, Text: text, Palette: p, CiphertextColors: p}),
		"burn":       NewBurnTextEffect(w, h, p, "default", text),
		"skull":      NewSkullTextEffect(w, h, p, "default", text),
	}

	for name, anim := range effects {
		t.Run(name, func(t *testing.T) {
			tu, ok := anim.(TextUpdatable)
			if !ok {
				t.Errorf("%s does not implement TextUpdatable", name)
				return
			}
			tu.SetText("CHANGED")
		})
	}
}

type updatableRenderer interface {
	Update()
	Render() string
	SetText(text string)
}

func TestSetText_ChangesRenderedOutput(t *testing.T) {
	w, h := 60, 20
	p := textUpdatableTestPalette()
	textA := "AAAA"
	textB := "ZZZZ"

	type effectFactory struct {
		name   string
		frames int
		create func(text string) updatableRenderer
	}

	factories := []effectFactory{
		{"fire-text", 5, func(text string) updatableRenderer { return NewFireTextEffect(w, h, p, text) }},
		{"matrix-art", 5, func(text string) updatableRenderer { return NewMatrixArtEffect(w, h, p, text) }},
		{"rain-art", 5, func(text string) updatableRenderer { return NewRainArtEffect(w, h, p, text) }},
		{"ring-text", 5, func(text string) updatableRenderer {
			return NewRingTextEffect(RingTextConfig{Width: w, Height: h, Text: text})
		}},
		{"blackhole", 5, func(text string) updatableRenderer {
			return NewBlackholeEffect(BlackholeConfig{Width: w, Height: h, Text: text})
		}},
		{"beam-text", 5, func(text string) updatableRenderer {
			return NewBeamTextEffect(BeamTextConfig{Width: w, Height: h, Text: text})
		}},
		{"print", 5, func(text string) updatableRenderer {
			return NewPrintEffect(PrintConfig{Width: w, Height: h, Text: text})
		}},
		{"skull", 5, func(text string) updatableRenderer { return NewSkullTextEffect(w, h, p, "default", text) }},
	}

	// Sequential effects where initial frames may look identical regardless of text.
	// These are tested for panic safety and interface compliance instead.
	t.Run("sequential-effects-skipped", func(t *testing.T) {
		t.Log("pour, decrypt, burn: sequential effects whose initial render is text-independent (tested via panic safety)")
	})

	for _, ef := range factories {
		t.Run(ef.name, func(t *testing.T) {
			anim := ef.create(textA)

			for i := 0; i < ef.frames; i++ {
				anim.Update()
			}
			outputA := anim.Render()

			anim.SetText(textB)

			for i := 0; i < ef.frames; i++ {
				anim.Update()
			}
			outputB := anim.Render()

			if outputA == outputB {
				t.Errorf("SetText(%q) did not change rendered output", textB)
			}
		})
	}
}

func TestSetText_DoesNotPanic(t *testing.T) {
	w, h := 40, 20
	p := textUpdatableTestPalette()

	effects := map[string]interface{}{
		"fire-text":  NewFireTextEffect(w, h, p, "HELLO"),
		"matrix-art": NewMatrixArtEffect(w, h, p, "HELLO"),
		"rain-art":   NewRainArtEffect(w, h, p, "HELLO"),
		"ring-text":  NewRingTextEffect(RingTextConfig{Width: w, Height: h, Text: "HELLO"}),
		"blackhole":  NewBlackholeEffect(BlackholeConfig{Width: w, Height: h, Text: "HELLO"}),
		"beam-text":  NewBeamTextEffect(BeamTextConfig{Width: w, Height: h, Text: "HELLO"}),
		"pour":       NewPourEffect(PourConfig{Width: w, Height: h, Text: "HELLO"}),
		"print":      NewPrintEffect(PrintConfig{Width: w, Height: h, Text: "HELLO"}),
		"decrypt":    NewDecryptEffect(DecryptConfig{Width: w, Height: h, Text: "HELLO", Palette: p, CiphertextColors: p}),
		"burn":       NewBurnTextEffect(w, h, p, "default", "HELLO"),
		"skull":      NewSkullTextEffect(w, h, p, "default", "HELLO"),
	}

	testTexts := []string{"", "A", "LONG TEXT WITH SPACES", "12:34:56 PM\nMONDAY, MARCH 2, 2026"}

	for name, anim := range effects {
		t.Run(name, func(t *testing.T) {
			tu, ok := anim.(TextUpdatable)
			if !ok {
				t.Skipf("%s does not implement TextUpdatable yet", name)
				return
			}

			type updater interface {
				Update()
				Render() string
			}
			ua := anim.(updater)

			for _, text := range testTexts {
				func() {
					defer func() {
						if r := recover(); r != nil {
							t.Errorf("SetText(%q) panicked: %v", text, r)
						}
					}()
					tu.SetText(text)
					ua.Update()
					ua.Render()
				}()
			}
		})
	}
}

func TestSetText_StabilityAfterMultipleCalls(t *testing.T) {
	w, h := 40, 20
	p := textUpdatableTestPalette()

	anim := NewFireTextEffect(w, h, p, "HELLO")

	for i := 0; i < 100; i++ {
		if i%20 == 0 {
			anim.SetText("WORLD")
		}
		if i%20 == 10 {
			anim.SetText("HELLO")
		}
		anim.Update()
		output := anim.Render()
		if output == "" {
			t.Fatalf("empty render at frame %d", i)
		}
	}
}

func TestBurnSetText_ReplacesCharacters(t *testing.T) {
	anim := NewBurnTextEffect(40, 20, textUpdatableTestPalette(), "default", "AB")

	if got := len(anim.chars); got != 2 {
		t.Fatalf("expected 2 chars after init, got %d", got)
	}

	anim.SetText("CD")
	if got := len(anim.chars); got != 2 {
		t.Fatalf("expected 2 chars after first SetText, got %d", got)
	}

	anim.SetText("EF")
	if got := len(anim.chars); got != 2 {
		t.Fatalf("expected 2 chars after second SetText, got %d", got)
	}
}

func TestTextEffects_NormalizeCRLFInput(t *testing.T) {
	w, h := 80, 24
	p := textUpdatableTestPalette()
	skullPalette := GetSkullPalette("default")
	crlfText := "A\r\nB"

	type effectFactory struct {
		name   string
		frames int
		create func(text string) updatableRenderer
	}

	factories := []effectFactory{
		{"fire-text", 40, func(text string) updatableRenderer { return NewFireTextEffect(w, h, p, text) }},
		{"matrix-art", 40, func(text string) updatableRenderer { return NewMatrixArtEffect(w, h, p, text) }},
		{"rain-art", 40, func(text string) updatableRenderer { return NewRainArtEffect(w, h, p, text) }},
		{"ring-text", 40, func(text string) updatableRenderer {
			return NewRingTextEffect(RingTextConfig{Width: w, Height: h, Text: text})
		}},
		{"blackhole", 40, func(text string) updatableRenderer {
			return NewBlackholeEffect(BlackholeConfig{Width: w, Height: h, Text: text})
		}},
		{"beam-text", 40, func(text string) updatableRenderer {
			return NewBeamTextEffect(BeamTextConfig{Width: w, Height: h, Text: text})
		}},
		{"print", 80, func(text string) updatableRenderer {
			return NewPrintEffect(PrintConfig{Width: w, Height: h, Text: text})
		}},
		{"skull-text", 40, func(text string) updatableRenderer {
			return NewSkullTextEffect(w, h, skullPalette, "default", text)
		}},
	}

	for _, ef := range factories {
		t.Run(ef.name, func(t *testing.T) {
			anim := ef.create(crlfText)
			for i := 0; i < ef.frames; i++ {
				anim.Update()
			}

			output := anim.Render()
			if strings.ContainsRune(output, '\r') {
				t.Fatalf("render output contains carriage return for CRLF input")
			}

			anim.SetText(crlfText)
			for i := 0; i < ef.frames; i++ {
				anim.Update()
			}
			output = anim.Render()
			if strings.ContainsRune(output, '\r') {
				t.Fatalf("render output contains carriage return after SetText with CRLF input")
			}
		})
	}
}

func TestFireAndSkullText_NormalizeStoredText(t *testing.T) {
	w, h := 80, 24
	p := textUpdatableTestPalette()
	skullPalette := GetSkullPalette("default")
	crlfText := "A\r\nB"
	lfText := "A\nB"

	fire := NewFireTextEffect(w, h, p, crlfText)
	if fire.text != lfText {
		t.Fatalf("fire-text constructor did not normalize CRLF: got %q want %q", fire.text, lfText)
	}
	fire.SetText(crlfText)
	if fire.text != lfText {
		t.Fatalf("fire-text SetText did not normalize CRLF: got %q want %q", fire.text, lfText)
	}

	skull := NewSkullTextEffect(w, h, skullPalette, "default", crlfText)
	if skull.textContent != lfText {
		t.Fatalf("skull-text constructor did not normalize CRLF: got %q want %q", skull.textContent, lfText)
	}
	skull.SetText(crlfText)
	if skull.textContent != lfText {
		t.Fatalf("skull-text SetText did not normalize CRLF: got %q want %q", skull.textContent, lfText)
	}
}

func TestIsTextUpdatable(t *testing.T) {
	textEffect := NewFireTextEffect(40, 20, textUpdatableTestPalette(), "TEST")
	nonTextEffect := NewFireEffect(40, 20, textUpdatableTestPalette())

	if _, ok := interface{}(textEffect).(TextUpdatable); !ok {
		t.Error("FireTextEffect should be TextUpdatable")
	}
	if _, ok := interface{}(nonTextEffect).(TextUpdatable); ok {
		t.Error("FireEffect should NOT be TextUpdatable")
	}
}
