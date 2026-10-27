package animations

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"
)

func TestHexToRGB(t *testing.T) {
	tests := []struct {
		name string
		hex  string
		want [3]int
	}{
		{name: "with hash", hex: "#ff79c6", want: [3]int{255, 121, 198}},
		{name: "without hash", hex: "FF79C6", want: [3]int{255, 121, 198}},
		{name: "black", hex: "#000000"},
		{name: "empty", hex: ""},
		{name: "short", hex: "#fff"},
		{name: "invalid", hex: "zzzzzz"},
		{name: "partially invalid", hex: "ffzzzz"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, g, b := hexToRGB(tt.hex)
			if got := [3]int{r, g, b}; got != tt.want {
				t.Fatalf("hexToRGB(%q) = %v, want %v", tt.hex, got, tt.want)
			}
		})
	}
}

func TestHexToRGBDoesNotAllocate(t *testing.T) {
	var got [3]int
	allocs := testing.AllocsPerRun(100, func() {
		got[0], got[1], got[2] = hexToRGB("#ff79c6")
	})
	if got != [3]int{255, 121, 198} {
		t.Fatalf("hexToRGB() = %v, want [255 121 198]", got)
	}
	if allocs != 0 {
		t.Fatalf("hexToRGB() allocated %v times per call, want 0", allocs)
	}
}

func TestANSIColorPrefix(t *testing.T) {
	tests := []struct {
		color rgbColor
		want  string
	}{
		{color: rgbColor{0, 10, 255}, want: "\033[38;2;0;10;255m"},
		{color: rgbColor{102, 85, 12}, want: "\033[38;2;102;85;12m"},
		{color: rgbColor{255, 121, 198}, want: "\033[38;2;255;121;198m"},
	}
	for _, tt := range tests {
		if got := ansiColorPrefix(tt.color); got != tt.want {
			t.Errorf("ansiColorPrefix(%v) = %q, want %q", tt.color, got, tt.want)
		}
	}
}

func TestGrowANSIFrameHandlesOverflowingDimensions(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	var builder strings.Builder
	growANSIFrame(&builder, maxInt, 1)
	if builder.Cap() != 0 {
		t.Fatalf("capacity = %d, want 0 for overflowing dimensions", builder.Cap())
	}
}

func TestColorRendererOutputMatchesGolden(t *testing.T) {
	palette := []string{"#101820", "#ff79c6", "#00ffff", "#ff0000", "#ffffff", "#8899aa", "#445566"}
	tests := []struct {
		name   string
		render func() string
		want   string
	}{
		{
			name: "fire",
			render: func() string {
				fire := NewFireEffect(80, 24, palette)
				for i := range fire.buffer {
					fire.buffer[i] = (i*17 + i/80*3) % 66
				}
				return fire.Render()
			},
			want: "cd2e340a2ee077daafdc4b7ccdb3b3e55d6c80d6a6a15315feb00d404bb2f8b3",
		},
		{
			name: "fire-text",
			render: func() string {
				fire := NewFireTextEffect(80, 24, palette, "SYSC")
				for i := range fire.buffer {
					fire.buffer[i] = (i*17 + i/80*3) % 66
				}
				for y := range fire.textMask {
					for x := range fire.textMask[y] {
						fire.textMask[y][x] = y >= 10 && y <= 12 && x%20 < 3
					}
				}
				return fire.Render()
			},
			want: "8d51aa20bd6d2658ed3e83b3cb95bed83dc5cf5f98c0ddbf3d4d12f3fba552e2",
		},
		{
			name: "cracktro",
			render: func() string {
				cracktro := NewCracktroEffect(80, 24, palette, "rama")
				cracktro.phase = CracktroRun
				cracktro.frameCount = 46
				cracktro.rotY, cracktro.rotX = 0, 0
				cracktro.innerRotY, cracktro.innerRotX = 0, 0
				cracktro.cubeVerts = []vec3{{-1, 0, 0}, {1, 0, 0}}
				cracktro.cubeEdges = []wireEdge{{0, 1}}
				cracktro.stars = []Star{{x: 4, y: 4, layer: 1}, {x: 75, y: 18, layer: 2}}
				cracktro.bars = []RasterBar{{phase: 0.2, colorIdx: 3}, {phase: 2.1, colorIdx: 4}}
				cracktro.trails = []trailPoint{{x: 2, y: 20, age: 1, ch: '+'}}
				cracktro.maxTrails = 0
				return cracktro.Render()
			},
			want: "a1ce2968a618d2a2b4838470b5940a63c6037afa8f814f0e756861a5fe2f7847",
		},
		{
			name: "sonar",
			render: func() string {
				sonar := NewSonarEffect(80, 24, palette, "rama")
				sonar.phase = PhaseHold
				sonar.frameCount = 35
				sonar.rings = []SonarRing{{birthFrame: 5}, {birthFrame: 23}}
				sonar.ashParticles = []AshParticle{{x: 3, y: 4, char: '*', layer: 0}, {x: 9, y: 12, char: 'o', layer: 1}}
				return sonar.Render()
			},
			want: "02f92f43e55ca32db8cb86db12b9a687a938c1a8ec02c0b6d33b4b3ca5aa785f",
		},
		{
			name: "skull",
			render: func() string {
				skull := NewSkullEffect(80, 24, palette, "rama")
				skull.phase = PhaseHold
				skull.frameCount = 20
				skull.holdStartFrame = 0
				skull.ashParticles = []AshParticle{{x: 3, y: 4, char: '*', layer: 0}, {x: 9, y: 12, char: 'o', layer: 1}}
				return skull.Render()
			},
			want: "c5939faf4d05764addb55afe3fd1a493aa837f10541d5c17ece64a773ecaf12b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rendered := tt.render()
			got := fmt.Sprintf("%x", sha256.Sum256([]byte(rendered)))
			if got != tt.want {
				t.Fatalf("render digest = %s, want %s", got, tt.want)
			}
		})
	}
}

type colorPerfEffect interface {
	Update()
	Render() string
}

func BenchmarkPerfHexFrame(b *testing.B) {
	for _, size := range [][2]int{{200, 55}, {380, 100}} {
		width, height := size[0], size[1]
		effects := []struct {
			name string
			new  func() colorPerfEffect
		}{
			{name: "fire", new: func() colorPerfEffect { return NewFireEffect(width, height, GetFirePalette("dracula")) }},
			{name: "fire-text", new: func() colorPerfEffect {
				return NewFireTextEffect(width, height, GetFirePalette("dracula"), "SYSC\nSYSC")
			}},
			{name: "cracktro", new: func() colorPerfEffect {
				return NewCracktroEffect(width, height, GetCracktroPalette("dracula"), "dracula")
			}},
			{name: "sonar", new: func() colorPerfEffect { return NewSonarEffect(width, height, GetSkullPalette("dracula"), "dracula") }},
			{name: "skull", new: func() colorPerfEffect { return NewSkullEffect(width, height, GetSkullPalette("dracula"), "dracula") }},
		}
		for _, effect := range effects {
			name, newEffect := effect.name, effect.new
			b.Run(fmt.Sprintf("%s/%dx%d", name, width, height), func(b *testing.B) {
				animation := newEffect()
				for i := 0; i < 60; i++ {
					animation.Update()
				}
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					animation.Update()
					_ = animation.Render()
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
