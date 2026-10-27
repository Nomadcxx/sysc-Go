package tui

import (
	"fmt"
	"testing"
)

func perfModel(b *testing.B, name string, width, height int) Model {
	m := NewModel()
	m.width, m.height = width, height
	found := false
	for i, animation := range m.animations {
		if animation == name {
			m.selectedAnimation = i
			found = true
			break
		}
	}
	if !found {
		b.Fatalf("animation %q is not in the picker", name)
	}
	if animationNeedsFile(name) {
		for i, file := range m.files {
			if file != "BIT Text Editor" && file != "Custom text" {
				m.selectedFile = i
				break
			}
		}
	}
	m.canvasHeight = m.height - mainViewChromeRows
	m.currentAnim = m.createAnimation()
	if m.currentAnim == nil {
		b.Fatalf("animation %q did not initialize", name)
	}
	m.animationRunning = true
	for range 60 {
		m.currentAnim.Update()
	}
	return m
}

func BenchmarkPerfStyleEffects(b *testing.B) {
	for _, size := range [][2]int{{200, 55}, {380, 100}} {
		for _, name := range []string{
			"matrix", "matrix-art", "rain", "rain-art", "beams", "beam-text",
			"ring-text", "blackhole", "aquarium", "pour", "print", "decrypt", "fireworks",
		} {
			b.Run(fmt.Sprintf("%s/%dx%d", name, size[0], size[1]), func(b *testing.B) {
				m := perfModel(b, name, size[0]+10, size[1]+mainViewChromeRows)
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					m.currentAnim.Update()
					_ = m.currentAnim.Render()
				}
			})
		}
	}
}
