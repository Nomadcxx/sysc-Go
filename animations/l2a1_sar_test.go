package animations

import (
	"math"
	"math/bits"
	"strings"
	"testing"
	"time"
)

func TestL2A1SARRenderCycleAndCredit(t *testing.T) {
	if len(l2a1Data) != 24000*12 {
		t.Fatalf("unexpected baked model size: %d", len(l2a1Data))
	}
	rotation, blend := modelPhase(8 * time.Second)
	if math.Abs(rotation-4*math.Pi) > 1e-10 || blend != 0 {
		t.Fatalf("rotation phase ended at %g with blend %g", rotation, blend)
	}
	_, blend = modelPhase(10 * time.Second)
	if blend != 1 {
		t.Fatalf("particle cloud reached blend %g", blend)
	}

	a := NewL2A1SAREffect(80, 24, "kanagawa")
	first := a.Render()
	if first == "" || first != a.Render() || strings.Count(first, "\n") != 23 {
		t.Fatal("empty, impure, or incorrectly sized render")
	}
	dots := 0
	for _, r := range first {
		if r > 0x2800 && r <= 0x28ff {
			dots++
		}
	}
	if dots < 80 {
		t.Fatalf("only %d braille cells", dots)
	}
	if !strings.Contains(first, "L2A1 SAR / notcplkerry / CC BY 4.0") {
		t.Fatal("missing on-screen attribution")
	}
	if meta := GetEffectMetadata("l2a1-sar"); meta == nil || meta.RequiresText {
		t.Fatal("missing standalone effect registry entry")
	}
	a.UpdateElapsed(9 * time.Second)
	if first == a.Render() {
		t.Fatal("particle phase did not change the frame")
	}
	a.UpdateElapsed(3 * time.Second)
	if first != a.Render() {
		t.Fatal("12-second loop did not reconstruct exactly")
	}
	a.Reset()
	if first != a.Render() {
		t.Fatal("reset did not restore the initial frame")
	}
}

func TestL2A1SARResizeRetainsPhaseAndAspect(t *testing.T) {
	a := NewL2A1SAREffect(80, 24, "kanagawa")
	a.UpdateElapsed(5 * time.Second)
	frame := a.Render()
	a.Resize(120, 40)
	a.SetCellAspect(7, 16)
	a.Resize(80, 24)
	if a.Render() == frame {
		t.Fatal("resize and aspect did not affect the rendered mesh")
	}
	a.SetCellAspect(1, 2)
	if a.Render() != frame {
		t.Fatal("restoring the grid and cell aspect changed the phase")
	}
	a.Resize(1<<30, 1<<30)
	if a.canvas.width != logoMaxGridCols || a.canvas.height != logoMaxGridRows {
		t.Fatal("render grid exceeded the logo canvas limits")
	}
}

func TestL2A1DisplayScaleUsesMoreBrailleDots(t *testing.T) {
	a := NewL2A1SAREffect(80, 24, "kanagawa")
	rotation, blend := modelPhase(a.elapsed)
	rotation += math.Pi / 2
	points := l2a1Points()
	baseline := renderMeshSpin(a.meshSpin, points, rotation, blend, 0, nil, 0, 0, 1)
	scaled := renderMeshSpin(a.meshSpin, points, rotation, blend, 0, nil, 0, 0, l2a1DisplayScale)
	if countBrailleDots(scaled) <= countBrailleDots(baseline) {
		t.Fatal("L2A1 display scale did not expose more braille detail")
	}
	if a.Render() != meshCredit(scaled, a.canvas.width, "L2A1 SAR / notcplkerry / CC BY 4.0") {
		t.Fatal("L2A1 render did not apply its detail scale")
	}
}

func TestL2A1DisplayScaleKeepsRotationInsideScreenEdges(t *testing.T) {
	a := NewL2A1SAREffect(80, 24, "kanagawa")
	for elapsed := time.Duration(0); elapsed < meshSpinPeriod; elapsed += 250 * time.Millisecond {
		a.Reset()
		a.UpdateElapsed(elapsed)
		minX, maxX, minY, maxY, ok := brailleBounds(a.Render())
		if !ok || minX < 1 || maxX >= 79 || minY < 1 || maxY >= 21 {
			t.Fatalf("scaled model touches screen edge at %s: x[%d,%d], y[%d,%d]", elapsed, minX, maxX, minY, maxY)
		}
	}
}

func countBrailleDots(frame string) int {
	dots := 0
	for _, r := range frame {
		if r > 0x2800 && r <= 0x28ff {
			dots += bits.OnesCount(uint(r - 0x2800))
		}
	}
	return dots
}

func brailleBounds(frame string) (minX, maxX, minY, maxY int, ok bool) {
	minX, minY = int(^uint(0)>>1), int(^uint(0)>>1)
	x, y := 0, 0
	escape := false
	for _, r := range frame {
		if r == 0x1b {
			escape = true
			continue
		}
		if escape {
			if r == 'm' {
				escape = false
			}
			continue
		}
		if r == '\n' {
			x, y = 0, y+1
			continue
		}
		if r > 0x2800 && r <= 0x28ff && bits.OnesCount(uint(r-0x2800)) > 0 {
			minX, maxX = min(minX, x), max(maxX, x)
			minY, maxY = min(minY, y), max(maxY, y)
			ok = true
		}
		x++
	}
	return
}
