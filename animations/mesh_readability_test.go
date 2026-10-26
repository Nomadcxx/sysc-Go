package animations

import (
	"math"
	"math/bits"
	"testing"
	"time"
)

// Check the colours after Braille composition: per-dot contrast alone missed
// the washed-out shader regression.
func TestMeshBrailleShadowParticlesRetainThemeChroma(t *testing.T) {
	for _, theme := range []string{"dracula", "cyan", "orange"} {
		for _, id := range []string{"napoleon-bust", "f-14-tomcat"} {
			t.Run(theme+"/"+id, func(t *testing.T) {
				a, err := NewRotatingModelEffect(id, 100, 30, theme)
				if err != nil {
					t.Fatal(err)
				}
				a.UpdateElapsed(1400 * time.Millisecond)
				cells, _, _ := parseBrailleTruecolorFrame(t, a.Render())
				var total float64
				count := 0
				for _, cell := range cells {
					if cell.ch <= 0x2800 || cell.ch > 0x28ff || bits.OnesCount8(uint8(cell.ch-0x2800)) > 4 {
						continue
					}
					r, g, b := float64(cell.fg[0]), float64(cell.fg[1]), float64(cell.fg[2])
					hi, lo := math.Max(r, math.Max(g, b)), math.Min(r, math.Min(g, b))
					if hi > 0 {
						total += (hi - lo) / hi
					}
					count++
				}
				if count == 0 || total/float64(count) < .15 {
					t.Fatalf("mean sparse Braille saturation %.3f across %d cells, want at least .15", total/float64(count), count)
				}
			})
		}
	}
}

func TestMeshLightingShadesBrailleWithVisibleParticles(t *testing.T) {
	for _, theme := range []string{"kanagawa", "dracula", "nord", "tokyo-night", "void"} {
		t.Run(theme, func(t *testing.T) {
			l := newMeshSpin(40, 12, theme, meshSpinPeriod)
			l.light = make([]float64, len(l.canvas.depth))
			for i := range l.canvas.depth {
				light := 0.0
				if i%(l.canvas.width*2) < l.canvas.width {
					light = 1
				}
				l.canvas.depth[i], l.canvas.crgb[i], l.light[i] = 0, meshShade(light, l.canvas), light
			}
			bg := .2126*l.canvas.litBg[0] + .7152*l.canvas.litBg[1] + .0722*l.canvas.litBg[2]
			cells, w, h := parseBrailleTruecolorFrame(t, l.braille(l.canvas))
			var front, back float64
			for i, cell := range cells {
				coverage := 0.0
				if cell.ch >= 0x2800 && cell.ch <= 0x28ff {
					coverage = float64(bits.OnesCount8(uint8(cell.ch-0x2800))) / 8
				}
				value := .2126*float64(cell.fg[0]) + .7152*float64(cell.fg[1]) + .0722*float64(cell.fg[2])
				value = bg + (value-bg)*coverage
				if i%w < w/2 {
					front += value
				} else {
					back += value
				}
			}
			front, back = front/float64(w*h/2), back/float64(w*h/2)
			t.Logf("front %.1f shadow %.1f background %.1f", front, back, bg)
			if back-bg < 25 || front-back < 100 {
				t.Fatalf("front %.1f shadow %.1f background %.1f: need visible shadows and separated faces", front, back, bg)
			}
		})
	}
}

func TestMeshSmallViewportBounds(t *testing.T) {
	for _, aspect := range [][2]int{{1, 2}, {9, 20}, {7, 16}} {
		for _, size := range [][2]int{{40, 12}, {60, 18}, {80, 24}} {
			for _, id := range []string{"l2a1-sar", "f-14-tomcat", "lowe"} {
				m := newMeshSpin(size[0], size[1], "kanagawa", meshSpinPeriod)
				m.SetCellAspect(aspect[0], aspect[1])
				points := l2a1Points()
				scale := l2a1DisplayScale
				if id != "l2a1-sar" {
					s := rotatingModelCatalog[id]
					points = s.surface()
					scale = s.displayScale
				}
				for step := 0; step < 24; step++ {
					frame := renderMeshSpinAtPitch(m, points, float64(step)*math.Pi/12, 0, 0, nil, 0, 0, scale, 0)
					x0, x1, y0, y1, ok := brailleBounds(frame)
					if !ok || x0 < 1 || x1 >= size[0]-1 || y0 < 1 || y1 >= size[1]-1 {
						t.Fatalf("%s %dx%d aspect %v yaw %d bounds %d,%d,%d,%d", id, size[0], size[1], aspect, step*15, x0, x1, y0, y1)
					}
				}
			}
		}
	}
}
