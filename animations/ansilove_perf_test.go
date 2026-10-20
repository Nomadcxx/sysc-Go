package animations

import (
	"fmt"
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
					Width: size.width, Height: size.height, Backend: backend.backend, Motion: ANSILoveMotionSubtle,
				})
				if err != nil {
					t.Fatal(err)
				}
				if elapsed := time.Since(start); elapsed > 2*size.bake {
					t.Fatalf("bake took %s, budget is %s", elapsed, 2*size.bake)
				}

				// Warm the scene through the logo glint so frame timing includes it.
				for i := 0; i < 180; i++ {
					a.UpdateElapsed(50 * time.Millisecond)
					a.Render()
				}
				var maxFrame, maxPan time.Duration
				maxBytes := 0
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
