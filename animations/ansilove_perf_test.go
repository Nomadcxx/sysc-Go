package animations

import (
	"fmt"
	"math"
	"testing"
	"time"
)

func TestANSILovePerformanceBudgets(t *testing.T) {
	if testing.Short() {
		t.Skip("performance budget check is omitted in short mode")
	}
	backends := []struct {
		name               string
		backend            ANSILoveBackend
		bake200, bake480   time.Duration
		frame200, frame480 time.Duration
	}{
		{"half", ANSILoveHalf, 40 * time.Millisecond, 250 * time.Millisecond, 800 * time.Microsecond, 4 * time.Millisecond},
		{"quad", ANSILoveQuad, 40 * time.Millisecond, 250 * time.Millisecond, 800 * time.Microsecond, 4 * time.Millisecond},
		{"sextant", ANSILoveSextant, 60 * time.Millisecond, 400 * time.Millisecond, time.Millisecond, 5 * time.Millisecond},
		{"octant", ANSILoveOctant, 60 * time.Millisecond, 400 * time.Millisecond, 1200 * time.Microsecond, 6 * time.Millisecond},
		{"braille-2c", ANSILoveBrailleTwoColor, 60 * time.Millisecond, 400 * time.Millisecond, 1200 * time.Microsecond, 6 * time.Millisecond},
		{"braille-fg", ANSILoveBrailleFG, 60 * time.Millisecond, 400 * time.Millisecond, 600 * time.Microsecond, 3 * time.Millisecond},
	}
	for _, backend := range backends {
		for _, size := range []struct {
			width, height int
			bake, frame   time.Duration
			bytes         int
		}{{200, 55, backend.bake200, backend.frame200, 40_000}, {480, 135, backend.bake480, backend.frame480, 150_000}} {
			t.Run(fmt.Sprintf("%s/%dx%d", backend.name, size.width, size.height), func(t *testing.T) {
				start := time.Now()
				a, err := NewANSILoveEffectConfig(ANSILoveConfig{
					Width: size.width, Height: size.height, Backend: backend.backend, Depth: ANSILove256, Motion: ANSILoveMotionSubtle,
				})
				if err != nil {
					t.Fatal(err)
				}
				if elapsed := time.Since(start); elapsed > 2*size.bake {
					t.Fatalf("bake took %s, budget is %s", elapsed, 2*size.bake)
				}

				prepareANSILoveVisibleGlint(t, a)
				a.cached = ""
				start = time.Now()
				glintFrame := a.Render()
				if elapsed := time.Since(start); elapsed > 2*size.frame {
					t.Errorf("visible glint frame took %s, budget is %s", elapsed, 2*size.frame)
				}
				var maxFrame, maxPan time.Duration
				maxBytes := len(glintFrame)
				for i := 0; i < 60; i++ {
					start = time.Now()
					a.UpdateElapsed(50 * time.Millisecond)
					frame := a.Render()
					if elapsed := time.Since(start); elapsed > maxFrame {
						maxFrame = elapsed
					}
					maxBytes = max(maxBytes, len(frame))
				}
				if maxFrame > 2*size.frame {
					t.Errorf("animated frame took %s, budget is %s", maxFrame, 2*size.frame)
				}
				if maxBytes > 2*size.bytes {
					t.Errorf("render string is %d bytes, budget is %d", maxBytes, 2*size.bytes)
				}
				if allocs := testing.AllocsPerRun(5, func() {
					a.UpdateElapsed(50 * time.Millisecond)
					a.Render()
				}); allocs > 16 {
					t.Errorf("animated frame allocated %g objects, budget is 16", allocs)
				}

				travel := a.travel()
				for i := 0; i < 10; i++ {
					a.cameraRow = (a.cameraRow + 1) % (travel + 1)
					a.cached = ""
					start = time.Now()
					a.Render()
					if elapsed := time.Since(start); elapsed > maxPan {
						maxPan = elapsed
					}
				}
				panBudget := 2 * time.Millisecond
				if size.width == 480 {
					panBudget = 10 * time.Millisecond
				}
				if maxPan > panBudget {
					t.Errorf("pan frame took %s, budget is %s", maxPan, panBudget)
				}
			})
		}
	}
}

func prepareANSILoveVisibleGlint(t *testing.T, a *ANSILoveAnimation) {
	t.Helper()
	l := a.sheet.layout
	holds := []struct {
		stage    int
		fraction float64
	}{{0, 0}, {2, .5}, {4, 1}, {6, .5}}
	for _, hold := range holds {
		cameraRow := int(math.Round(hold.fraction * float64(a.travel())))
		for y := range a.sheet.rowFlags {
			if y < cameraRow || y >= cameraRow+a.cfg.Height {
				continue
			}
			for x := 0; x < l.Width; x++ {
				if a.sheet.flags[y*l.Width+x]&2 == 0 {
					continue
				}
				diagonal := l.OX + float64(l.Left+x)*l.CellW/l.Scale + float64(y)*l.CellH/l.Scale
				tick := int(math.Round((diagonal - 1600) * 29 / 1120))
				if tick < 10 || tick > 18 {
					continue
				}
				phase := time.Duration(tick)*50*time.Millisecond + 25*time.Millisecond
				glint := 1600 + float64(tick)/29*1120
				if math.Abs(diagonal-glint) > 32 {
					continue
				}
				a.stage, a.stageTime, a.elapsed = hold.stage, phase.Seconds(), phase
				a.cameraRow = cameraRow
				a.cached = ""
				withGlint := a.Render()
				baseline := *a
				baseline.stage = hold.stage + 1
				baseline.cached = ""
				baseline.scratch = nil
				if withGlint != baseline.Render() {
					return
				}
			}
		}
	}
	t.Fatal("no visible logo glint found in the hold window")
}
