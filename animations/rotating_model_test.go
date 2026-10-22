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
	"f-14-tomcat",
	"f-15e-strike-eagle",
	"f-22-raptor",
	"fa-18e-super-hornet",
}

func TestFighterModelCreditsMatchSources(t *testing.T) {
	models := []struct{ id, credit string }{
		{"f-14-tomcat", "F-14 Tomcat - Fighter Jet - Free / bohmerang / CC BY-NC-SA 4.0"},
		{"f-15e-strike-eagle", "F-15E Strike Eagle - Fighter Jet - Free / bohmerang / CC BY-NC-SA 4.0"},
		{"f-22-raptor", "F-22 Raptor - Fighter Jet - Free / bohmerang / CC BY-NC-SA 4.0"},
		{"fa-18e-super-hornet", "F/A-18E Super Hornet (USA) / 42manako / CC BY-NC 4.0"},
	}
	for _, model := range models {
		t.Run(model.id, func(t *testing.T) {
			spec := rotatingModelCatalog[model.id]
			if spec == nil {
				t.Fatal("model missing from rotating model catalog")
			}
			if spec.credit != model.credit {
				t.Fatalf("credit = %q, want %q", spec.credit, model.credit)
			}
			a, err := NewRotatingModelEffect(model.id, 100, 30, "kanagawa")
			if err != nil {
				t.Fatal(err)
			}
			if frame := a.Render(); !strings.Contains(frame, model.credit) {
				t.Fatalf("rendered frame omits credit %q", model.credit)
			}
		})
	}
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

func TestMeshSpinLightingSeparatesFrontAndBackFaces(t *testing.T) {
	m := newMeshSpin(80, 24, "kanagawa", meshSpinPeriod)
	points := []meshPoint{
		{position: [3]float64{-.3, 0, 0}, normal: [3]float64{0, 0, 1}},
		{position: [3]float64{.3, 0, 0}, normal: [3]float64{0, 0, -1}},
	}
	renderMeshSpinAtPitch(m, points, 0, 0, 0, nil, 0, 0, 1, 0)
	front := meshSpinColorAt(m, points[0].position)
	back := meshSpinColorAt(m, points[1].position)
	if front == 0 || back == 0 || front == back {
		t.Fatalf("front and back lighting = %#06x and %#06x, want distinct nonzero theme shades", front, back)
	}
}

func meshSpinColorAt(m *meshSpin, p [3]float64) uint32 {
	l := m.canvas
	scale := math.Min(float64(l.width*m.cellWidth)*.40, float64((l.height-2)*m.cellHeight)*.72)
	scaleX, scaleY := scale*2/float64(m.cellWidth), scale*4/float64(m.cellHeight)
	perspective := logoPerspective / (logoPerspective - p[2])
	x := int(math.Floor(float64(l.width*2)/2 + p[0]*scaleX*perspective))
	y := int(math.Floor(float64(l.height*4-8)/2 + p[1]*scaleY*perspective))
	return l.crgb[y*l.width*2+x]
}

func TestRotatingModelBakesUseTheBoundedVariableSampleCodec(t *testing.T) {
	for _, id := range rotatingModelTestIDs {
		spec := rotatingModelCatalog[id]
		if spec == nil {
			t.Errorf("%s missing from rotating model catalog", id)
			continue
		}
		data := spec.data
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
