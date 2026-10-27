package animations

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
)

type colorPerfAnimation interface {
	Update()
	Render() string
}

func colorPerfFrames() map[string]string {
	palette := []string{"#102030", "#ff0000", "#00ff00", "#0000ff", "#e0e0e0", "#ffaa00", "#224466"}
	fire := NewFireEffect(80, 24, palette)
	for i := range fire.buffer {
		fire.buffer[i] = i * 19 % 66
	}

	fireText := NewFireTextEffect(80, 24, palette, "SYS C\nGO")
	for i := range fireText.buffer {
		fireText.buffer[i] = i * 23 % 66
	}

	cracktro := NewCracktroEffect(80, 24, palette, "dracula")
	cracktro.phase = CracktroRun
	cracktro.frameCount = 12
	cracktro.rotY, cracktro.rotX = .4, -.2
	cracktro.innerRotY, cracktro.innerRotX = -.3, .1
	// ponytail: omit cube geometry because wireMap iteration makes glow-depth
	// ties nondeterministic; restore this when cracktro render ordering is stable.
	cracktro.cubeVerts, cracktro.cubeEdges = nil, nil
	cracktro.stars = []Star{
		{x: 5.5, y: 6, speed: .4, layer: 0},
		{x: 18.5, y: 8, speed: .8, layer: 1},
		{x: 72.5, y: 18, speed: 1.4, layer: 2},
	}
	cracktro.trails = []trailPoint{
		{x: 2, y: 2, age: 1, ch: '*'},
		{x: 3, y: 2, age: 2, ch: '*'},
		{x: 4, y: 2, age: 3, ch: '*'},
	}

	sonar := NewSonarEffect(80, 24, palette, "dracula")
	sonar.phase, sonar.frameCount, sonar.holdStartFrame = PhaseHold, 40, 0
	sonar.rings = []SonarRing{{birthFrame: 10}}
	sonar.ashParticles = []AshParticle{
		{x: 3, y: 2, char: '.', layer: 0},
		{x: 72, y: 7, char: '▒', layer: 1},
	}

	skull := NewSkullEffect(80, 24, palette, "dracula")
	skull.phase, skull.frameCount, skull.holdStartFrame = PhaseHold, 20, 0
	for i := range skull.skullChars {
		skull.skullChars[i].locked = true
		skull.skullChars[i].currentY = float64(skull.skullChars[i].finalY)
	}
	skull.ashParticles = []AshParticle{
		{x: 4, y: 2, char: '.', layer: 0},
		{x: 73, y: 9, char: '▒', layer: 1},
	}

	return map[string]string{
		"fire":      fire.Render(),
		"fire-text": fireText.Render(),
		"cracktro":  cracktro.Render(),
		"sonar":     sonar.Render(),
		"skull":     skull.Render(),
	}
}

func TestColorRenderGolden(t *testing.T) {
	want := map[string]string{
		"fire":      "0558d929d8332ed5c64d5cd331df369bbe22b427ff21f9a1a68f21c95c913109",
		"fire-text": "1cf04ca473e8a63d49d812dd127d75696e2c2e048dd6ee0da40bbebf80c4d2d6",
		"cracktro":  "a3272cd844150f5dd2c94590c993bf44bc2b7740f7c2496a9ab2e417b90572c8",
		"sonar":     "391cf45ecfcc978205e98c1cfc3f67ba6169fed30ba2e5915b206bef8bae49c2",
		"skull":     "a7df882ffa4fd59da224dd513dee292118cc724360856e93522ed35a23bed1bf",
	}
	frames := colorPerfFrames()
	if len(frames) != len(want) {
		t.Fatalf("got %d render fixtures, want %d", len(frames), len(want))
	}
	for name, frame := range frames {
		sum := sha256.Sum256([]byte(frame))
		got := hex.EncodeToString(sum[:])
		if got != want[name] {
			t.Errorf("%s render changed: got %s, want %s", name, got, want[name])
		}
	}
}

func TestHexToRGB(t *testing.T) {
	tests := []struct {
		input               string
		wantR, wantG, wantB int
	}{
		{input: "#ff79c6", wantR: 255, wantG: 121, wantB: 198},
		{input: "FF79C6", wantR: 255, wantG: 121, wantB: 198},
		{input: "#000000"},
		{input: ""},
		{input: "#fff"},
		{input: "zzzzzz"},
		{input: "ffzz00"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			r, g, b := hexToRGB(tt.input)
			if r != tt.wantR || g != tt.wantG || b != tt.wantB {
				t.Fatalf("hexToRGB(%q) = (%d, %d, %d), want (%d, %d, %d)", tt.input, r, g, b, tt.wantR, tt.wantG, tt.wantB)
			}
		})
	}
}

func TestHexToRGBNoAllocations(t *testing.T) {
	allocs := testing.AllocsPerRun(100, func() {
		r, g, b := hexToRGB("#ff79c6")
		if r != 255 || g != 121 || b != 198 {
			t.Fatalf("unexpected color: (%d, %d, %d)", r, g, b)
		}
	})
	if allocs != 0 {
		t.Fatalf("hexToRGB allocated %g times per call, want 0", allocs)
	}
}

func BenchmarkPerfHexFrame(b *testing.B) {
	for _, size := range [][2]int{{200, 55}, {380, 100}} {
		width, height := size[0], size[1]
		constructors := map[string]func() colorPerfAnimation{
			"fire": func() colorPerfAnimation {
				return NewFireEffect(width, height, GetFirePalette("dracula"))
			},
			"fire-text": func() colorPerfAnimation {
				return NewFireTextEffect(width, height, GetFirePalette("dracula"), "SYSC\nSYSC")
			},
			"cracktro": func() colorPerfAnimation {
				return NewCracktroEffect(width, height, GetCracktroPalette("dracula"), "dracula")
			},
			"sonar": func() colorPerfAnimation {
				return NewSonarEffect(width, height, GetSkullPalette("dracula"), "dracula")
			},
			"skull": func() colorPerfAnimation {
				return NewSkullEffect(width, height, GetSkullPalette("dracula"), "dracula")
			},
		}
		for name, newEffect := range constructors {
			b.Run(fmt.Sprintf("%s/%dx%d", name, width, height), func(b *testing.B) {
				effect := newEffect()
				for i := 0; i < 60; i++ {
					effect.Update()
				}
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					effect.Update()
					_ = effect.Render()
				}
			})
		}
	}
}

func BenchmarkPerfHexToRGB(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		hexToRGB("#ff79c6")
	}
}
