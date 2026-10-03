package animations

import (
	"testing"
)

var testPalette = []string{"#00ff00", "#00cc00", "#009900", "#006600", "#003300"}
var testText = "HELLO WORLD"

type renderableEffect interface {
	Render() string
}

func newTestBeams() renderableEffect {
	return NewBeamsEffect(BeamsConfig{
		Width:                80,
		Height:               24,
		BeamRowSymbols:       []rune{'▀', '▄'},
		BeamColumnSymbols:    []rune{'▌', '▐'},
		BeamDelay:            2,
		BeamRowSpeedRange:    [2]int{5, 15},
		BeamColumnSpeedRange: [2]int{3, 8},
		BeamGradientStops:    testPalette,
		BeamGradientSteps:    5,
		BeamGradientFrames:   10,
		FinalGradientStops:   testPalette,
		FinalGradientSteps:   5,
		FinalGradientFrames:  10,
		FinalWipeSpeed:       1,
	})
}

func newTestBeamText() renderableEffect {
	return NewBeamTextEffect(BeamTextConfig{
		Width:                80,
		Height:               24,
		Text:                 testText,
		BeamRowSymbols:       []rune{'▀', '▄'},
		BeamColumnSymbols:    []rune{'▌', '▐'},
		BeamDelay:            2,
		BeamRowSpeedRange:    [2]int{5, 15},
		BeamColumnSpeedRange: [2]int{3, 8},
		BeamGradientStops:    testPalette,
		BeamGradientSteps:    5,
		BeamGradientFrames:   10,
		FinalGradientStops:   testPalette,
		FinalGradientSteps:   5,
		FinalGradientFrames:  10,
		FinalWipeSpeed:       1,
	})
}

func newTestAquarium() renderableEffect {
	return NewAquariumEffect(AquariumConfig{
		Width:         80,
		Height:        24,
		FishColors:    testPalette,
		WaterColors:   testPalette,
		SeaweedColors: testPalette,
		BubbleColor:   "#00ccff",
		DiverColor:    "#ffcc00",
		BoatColor:     "#996633",
		MermaidColor:  "#ff66cc",
	})
}

func newTestBlackhole() renderableEffect {
	return NewBlackholeEffect(BlackholeConfig{
		Width:  80,
		Height: 24,
		Text:   testText,
	})
}

func newTestRingText() renderableEffect {
	return NewRingTextEffect(RingTextConfig{
		Width:  80,
		Height: 24,
		Text:   testText,
	})
}

func newTestDecrypt() renderableEffect {
	return NewDecryptEffect(DecryptConfig{
		Width:              80,
		Height:             24,
		Text:               testText,
		Palette:            testPalette,
		CiphertextColors:   testPalette,
		FinalGradientStops: testPalette,
		FinalGradientSteps: 5,
		TypingSpeed:        1,
	})
}

func newTestPour() renderableEffect {
	return NewPourEffect(PourConfig{
		Width:  80,
		Height: 24,
		Text:   testText,
	})
}

func newTestPrint() renderableEffect {
	return NewPrintEffect(PrintConfig{
		Width:  80,
		Height: 24,
		Text:   testText,
	})
}

type effectFactory struct {
	name    string
	factory func() renderableEffect
}

func allEffects() []effectFactory {
	return []effectFactory{
		{"matrix", func() renderableEffect { return NewMatrixEffect(80, 24, testPalette) }},
		{"matrixart", func() renderableEffect { return NewMatrixArtEffect(80, 24, testPalette, testText) }},
		{"rain", func() renderableEffect { return NewRainEffect(80, 24, testPalette) }},
		{"rainart", func() renderableEffect { return NewRainArtEffect(80, 24, testPalette, testText) }},
		{"fireworks", func() renderableEffect { return NewFireworksEffect(80, 24, testPalette) }},
		{"skull", func() renderableEffect {
			return NewSkullEffect(80, 24, []string{"#00ff00", "#00cc00", "#009900", "#006600", "#003300", "#444444", "#222222"}, "rama")
		}},
		{"beams", newTestBeams},
		{"beamtext", newTestBeamText},
		{"aquarium", newTestAquarium},
		{"blackhole", newTestBlackhole},
		{"ringtext", newTestRingText},
		{"decrypt", newTestDecrypt},
		{"pour", newTestPour},
		{"print", newTestPrint},
	}
}

func TestRenderDoesNotPanic(t *testing.T) {
	for _, ef := range allEffects() {
		t.Run(ef.name, func(t *testing.T) {
			e := ef.factory()
			result := e.Render()
			if len(result) == 0 {
				t.Logf("%s: Render() returned empty string (may be normal for initial frame)", ef.name)
			}
		})
	}
}

func TestMultipleRendersStable(t *testing.T) {
	for _, ef := range allEffects() {
		t.Run(ef.name, func(t *testing.T) {
			e := ef.factory()
			for i := 0; i < 100; i++ {
				e.Render()
			}
		})
	}
}

func BenchmarkMatrixRender(b *testing.B) {
	e := NewMatrixEffect(80, 24, testPalette)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Render()
	}
}

func BenchmarkMatrixArtRender(b *testing.B) {
	e := NewMatrixArtEffect(80, 24, testPalette, testText)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Render()
	}
}

func BenchmarkRainRender(b *testing.B) {
	e := NewRainEffect(80, 24, testPalette)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Render()
	}
}

func BenchmarkRainArtRender(b *testing.B) {
	e := NewRainArtEffect(80, 24, testPalette, testText)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Render()
	}
}

func BenchmarkBeamsRender(b *testing.B) {
	e := newTestBeams()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Render()
	}
}

func BenchmarkBeamTextRender(b *testing.B) {
	e := newTestBeamText()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Render()
	}
}

func BenchmarkAquariumRender(b *testing.B) {
	e := newTestAquarium()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Render()
	}
}

func BenchmarkBlackholeRender(b *testing.B) {
	e := newTestBlackhole()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Render()
	}
}

func BenchmarkRingTextRender(b *testing.B) {
	e := newTestRingText()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Render()
	}
}

func BenchmarkFireworksRender(b *testing.B) {
	e := NewFireworksEffect(80, 24, testPalette)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Render()
	}
}

func BenchmarkDecryptRender(b *testing.B) {
	e := newTestDecrypt()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Render()
	}
}

func BenchmarkPourRender(b *testing.B) {
	e := newTestPour()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Render()
	}
}

func BenchmarkPrintRender(b *testing.B) {
	e := newTestPrint()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Render()
	}
}

func BenchmarkSkullRender(b *testing.B) {
	e := NewSkullEffect(80, 24, []string{"#00ff00", "#00cc00", "#009900", "#006600", "#003300", "#444444", "#222222"}, "rama")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Render()
	}
}
