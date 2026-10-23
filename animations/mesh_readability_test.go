package animations

import (
	"math"
	"testing"
)

func TestMeshLightingKeepsShadowVisibleAndSeparatesFaces(t *testing.T) {
	for _, theme := range []string{"kanagawa", "dracula", "nord", "tokyo-night", "void"} {
		t.Run(theme, func(t *testing.T) {
			l := newMeshSpin(40, 12, theme, meshSpinPeriod)
			luminance := func(c uint32) float64 {
				return .2126*float64(c>>16&255) + .7152*float64(c>>8&255) + .0722*float64(c&255)
			}
			points := []meshPoint{{position: [3]float64{-.3, 0, 0}, normal: [3]float64{-.410, -.564, .718}}, {position: [3]float64{.3, 0, 0}, normal: [3]float64{.410, .564, -.718}}}
			renderMeshSpinAtPitch(l, points, 0, 0, 0, nil, 0, 0, 1, 0)
			front := luminance(meshSpinColorAt(l, points[0].position))
			back := luminance(meshSpinColorAt(l, points[1].position))
			bg := .2126*l.canvas.litBg[0] + .7152*l.canvas.litBg[1] + .0722*l.canvas.litBg[2]
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
