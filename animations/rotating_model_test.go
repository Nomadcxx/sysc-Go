package animations

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/Nomadcxx/sysc-Go/internal/mesh"
)

var rotatingModelTestIDs = []string{
	"jesus-christ-statue",
	"jason",
	"nymphe-der-flora",
	"lowe",
	"roman-gladiator",
	"grabfigur",
}

func TestRotatingModelConstructorAndLoop(t *testing.T) {
	if _, err := NewRotatingModelEffect("missing", 80, 24, "kanagawa"); err == nil {
		t.Fatal("accepted unknown model ID")
	}
	for _, id := range rotatingModelTestIDs {
		t.Run(id, func(t *testing.T) {
			a, err := NewRotatingModelEffect(id, 80, 24, "kanagawa")
			if err != nil {
				t.Fatal(err)
			}
			first := a.Render()
			if first == "" || first != a.Render() {
				t.Fatal("render is empty or nondeterministic")
			}
			if !strings.Contains(first, rotatingModelCatalog[id].credit) {
				t.Fatalf("missing model attribution: %s", rotatingModelCatalog[id].credit)
			}
			if meta := GetEffectMetadata(id); meta == nil || meta.RequiresText {
				t.Fatal("missing standalone effect registry entry")
			}

			a.UpdateElapsed(rotatingModelPeriod / 4)
			quarterTurn := a.Render()
			if first == quarterTurn {
				t.Fatal("horizontal rotation did not advance")
			}
			quarterTurnRotation := 2 * math.Pi * float64(a.elapsed) / float64(a.period)
			expected := renderMeshSpinAtPitch(a.meshSpin, a.spec.surface(), quarterTurnRotation, 0, 0, nil, 0, 0, a.spec.displayScale, 0)
			expected = meshCredit(expected, a.canvas.width, a.spec.credit)
			if quarterTurn != expected {
				t.Fatal("statue rotation changed camera pitch, blend, or swirl")
			}
			a.UpdateElapsed(rotatingModelPeriod * 3 / 4)
			if first != a.Render() {
				t.Fatal("full yaw loop did not return to its initial frame")
			}
		})
	}
}

func TestRotatingModelsStayInsideScreenThroughFullYaw(t *testing.T) {
	for _, id := range rotatingModelTestIDs {
		t.Run(id, func(t *testing.T) {
			for _, size := range []struct{ width, height int }{{80, 24}, {491, 90}} {
				a, err := NewRotatingModelEffect(id, size.width, size.height, "kanagawa")
				if err != nil {
					t.Fatal(err)
				}
				for step := 0; step < 24; step++ {
					a.Reset()
					a.UpdateElapsed(rotatingModelPeriod * time.Duration(step) / 24)
					minX, maxX, minY, maxY, ok := brailleBounds(a.Render())
					if !ok || minX < 1 || maxX >= size.width-1 || minY < 1 || maxY >= size.height-1 {
						t.Fatalf("model touches frame at %d° on %dx%d: x[%d,%d], y[%d,%d]", step*15, size.width, size.height, minX, maxX, minY, maxY)
					}
				}
			}
		})
	}
}

func TestRotatingModelsShadeTheirSurfaces(t *testing.T) {
	for _, id := range rotatingModelTestIDs {
		t.Run(id, func(t *testing.T) {
			a, err := NewRotatingModelEffect(id, 80, 24, "kanagawa")
			if err != nil {
				t.Fatal(err)
			}
			a.Render()
			colors := make(map[uint32]struct{})
			for _, color := range a.canvas.crgb {
				if color != 0 {
					colors[color] = struct{}{}
				}
			}
			if len(colors) < 2 {
				t.Fatalf("surface normals produced %d shades, want directional separation", len(colors))
			}
		})
	}
}

func TestRotatingModelBakesUseTheBoundedVariableSampleCodec(t *testing.T) {
	for _, id := range rotatingModelTestIDs {
		data := rotatingModelCatalog[id].data
		if got, want := len(data), 96000*mesh.SampleBytes; got != want {
			t.Errorf("%s data has %d bytes, want %d", id, got, want)
		}
	}
}

func TestMeshCreditUsesDisplayCellWidth(t *testing.T) {
	if got, want := meshCredit("frame\n", 10, "Löwe"), "frame\n      Löwe"; got != want {
		t.Fatalf("meshCredit() = %q, want %q", got, want)
	}
}

func TestRotatingModelPerformanceBudget(t *testing.T) {
	if testing.Short() {
		t.Skip("render budget check runs outside race instrumentation")
	}
	const frameBudget = 30 * time.Millisecond
	for _, id := range rotatingModelTestIDs {
		t.Run(id, func(t *testing.T) {
			a, err := NewRotatingModelEffect(id, 491, 90, "kanagawa")
			if err != nil {
				t.Fatal(err)
			}
			var maxFrame time.Duration
			for frame := 0; frame < 30; frame++ {
				a.Update()
				start := time.Now()
				a.Render()
				if elapsed := time.Since(start); elapsed > maxFrame {
					maxFrame = elapsed
				}
			}
			if maxFrame > frameBudget {
				t.Fatalf("max frame render time %s exceeds %s", maxFrame, frameBudget)
			}
		})
	}
}
