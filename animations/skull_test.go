package animations

import (
	"fmt"
	"slices"
	"strings"
	"testing"
)

func testSkullPalette() []string {
	return GetSkullPalette("rama")
}

func TestSkullFinalPositionsStayInBounds(t *testing.T) {
	s := NewSkullEffect(90, 35, testSkullPalette(), "rama")
	if len(s.skullChars) == 0 {
		t.Fatal("expected skull characters to be parsed")
	}

	outOfBounds := 0
	for _, ch := range s.skullChars {
		if ch.finalX < 0 || ch.finalX >= s.width || ch.finalY < 0 || ch.finalY >= s.height {
			outOfBounds++
		}
	}

	if outOfBounds != 0 {
		t.Fatalf("expected all skull chars to be in bounds, got %d out-of-bounds chars", outOfBounds)
	}
}

func TestSkullIsHorizontallyCentered(t *testing.T) {
	s := NewSkullEffect(100, 35, testSkullPalette(), "rama")
	if len(s.skullChars) == 0 {
		t.Fatal("expected skull characters to be parsed")
	}

	minX, maxX := s.width, -1
	for _, ch := range s.skullChars {
		if ch.finalX < minX {
			minX = ch.finalX
		}
		if ch.finalX > maxX {
			maxX = ch.finalX
		}
	}

	leftMargin := minX
	rightMargin := (s.width - 1) - maxX
	delta := leftMargin - rightMargin
	if delta < 0 {
		delta = -delta
	}

	if delta > 1 {
		t.Fatalf("expected skull to be centered (left=%d right=%d), delta=%d", leftMargin, rightMargin, delta)
	}
}

// TestSkullIsVerticallyCentered verifies the skull sits in the middle of the
// viewport, not biased toward the bottom (which was required when text
// appeared above the skull; text is now removed).
func TestSkullIsVerticallyCentered(t *testing.T) {
	s := NewSkullEffect(90, 40, testSkullPalette(), "rama")

	minY, maxY := s.height, -1
	for _, ch := range s.skullChars {
		if ch.finalY < minY {
			minY = ch.finalY
		}
		if ch.finalY > maxY {
			maxY = ch.finalY
		}
	}
	topMargin := minY
	bottomMargin := (s.height - 1) - maxY
	delta := topMargin - bottomMargin
	if delta < 0 {
		delta = -delta
	}
	if delta > 2 {
		t.Fatalf("expected skull to be vertically centered (top=%d bottom=%d), delta=%d",
			topMargin, bottomMargin, delta)
	}
}

func TestSkullDripCharsStartAtTopOfScreen(t *testing.T) {
	s := NewSkullEffect(90, 35, testSkullPalette(), "rama")
	for _, ch := range s.skullChars {
		if ch.currentY > 0 {
			t.Fatalf("expected all skull chars to start at currentY<=0, got currentY=%.1f for char finalY=%d",
				ch.currentY, ch.finalY)
		}
	}
}

func TestSkullTopCharsLockBeforeBottomChars(t *testing.T) {
	s := NewSkullEffect(90, 35, testSkullPalette(), "rama")

	topFinalY, bottomFinalY := s.height, 0
	for _, ch := range s.skullChars {
		if ch.finalY < topFinalY {
			topFinalY = ch.finalY
		}
		if ch.finalY > bottomFinalY {
			bottomFinalY = ch.finalY
		}
	}

	topLockFrame, bottomLockFrame := -1, -1
	topDone, bottomDone := false, false

	for i := 0; i < 300; i++ {
		s.Update()

		if !topDone {
			allTopLocked := true
			for _, ch := range s.skullChars {
				if ch.finalY == topFinalY && !ch.locked {
					allTopLocked = false
					break
				}
			}
			if allTopLocked {
				topLockFrame = s.frameCount
				topDone = true
			}
		}
		if !bottomDone {
			allBottomLocked := true
			for _, ch := range s.skullChars {
				if ch.finalY == bottomFinalY && !ch.locked {
					allBottomLocked = false
					break
				}
			}
			if allBottomLocked {
				bottomLockFrame = s.frameCount
				bottomDone = true
			}
		}

		if topDone && bottomDone {
			break
		}
	}

	if topLockFrame < 0 || bottomLockFrame < 0 {
		t.Fatalf("chars did not lock within 300 frames (topLock=%d, bottomLock=%d)", topLockFrame, bottomLockFrame)
	}
	if topLockFrame >= bottomLockFrame {
		t.Fatalf("expected top row (finalY=%d) to lock before bottom row (finalY=%d): topFrame=%d, bottomFrame=%d",
			topFinalY, bottomFinalY, topLockFrame, bottomLockFrame)
	}
}

// TestSkullAlwaysWhite verifies that regardless of phase, skull characters
// render with the brightest palette slot (palette[4]). There must be no
// coloured illumination ramp on the skull itself.
func TestSkullAlwaysWhite(t *testing.T) {
	s := NewSkullEffect(90, 35, testSkullPalette(), "rama")

	for i := range s.skullChars {
		s.skullChars[i].locked = true
		s.skullChars[i].currentY = float64(s.skullChars[i].finalY)
	}

	r, g, b := hexToRGB(s.palette[4])
	whiteEscape := fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)

	phases := []AnimationPhase{PhaseDrip, PhaseGrid, PhaseHold}
	for _, p := range phases {
		s.phase = p
		if p == PhaseHold {
			s.holdStartFrame = s.frameCount
		}
		if !strings.Contains(s.Render(), whiteEscape) {
			t.Errorf("phase %v: expected skull colour escape %q in render", p, whiteEscape)
		}
	}
}

// TestSkullInteriorNegativeSpacePreserved verifies the renderer never emits a
// background-fill escape. Interior gaps in the ASCII art must show through.
func TestSkullInteriorNegativeSpacePreserved(t *testing.T) {
	s := NewSkullEffect(90, 35, testSkullPalette(), "rama")
	for i := range s.skullChars {
		s.skullChars[i].locked = true
		s.skullChars[i].currentY = float64(s.skullChars[i].finalY)
	}

	if strings.Contains(s.Render(), "\033[48;2;") {
		t.Error("expected no background-fill escape; negative space must show through")
	}
}

// TestSkullTopLineCentredWithBody asserts the narrow crown is horizontally
// aligned above the wider body, not flush-left.
func TestSkullTopLineCentredWithBody(t *testing.T) {
	s := NewSkullEffect(90, 35, testSkullPalette(), "rama")

	rows := make(map[int][2]int)
	for _, ch := range s.skullChars {
		if b, ok := rows[ch.finalY]; ok {
			if ch.finalX < b[0] {
				b[0] = ch.finalX
			}
			if ch.finalX > b[1] {
				b[1] = ch.finalX
			}
			rows[ch.finalY] = b
		} else {
			rows[ch.finalY] = [2]int{ch.finalX, ch.finalX}
		}
	}

	topY := s.height
	bodyMid := -1
	for y, b := range rows {
		if y < topY {
			topY = y
		}
		mid := (b[0] + b[1]) / 2
		if bodyMid < 0 || y > topY {
			bodyMid = mid
		}
	}

	topMid := (rows[topY][0] + rows[topY][1]) / 2
	if diff := topMid - bodyMid; diff < -4 || diff > 4 {
		t.Fatalf("expected top crown mid (%d) close to body mid (%d), delta=%d", topMid, bodyMid, diff)
	}
}

// TestGridBuiltBehindSkull verifies grid cells are produced and sit at or
// below the skull's vertical centre (they form the floor, not a ceiling).
func TestGridBuiltBehindSkull(t *testing.T) {
	s := NewSkullEffect(90, 40, testSkullPalette(), "rama")

	if len(s.gridCells) == 0 {
		t.Fatal("expected background grid cells to be built")
	}
	for _, g := range s.gridCells {
		if g.y < s.horizonY {
			t.Fatalf("grid cell above horizon (y=%d, horizon=%d)", g.y, s.horizonY)
		}
	}
}

// TestGridIgnitesLeftToRight verifies that during PhaseGrid the lit fraction
// advances monotonically from 0 to 1.
func TestGridIgnitesLeftToRight(t *testing.T) {
	s := NewSkullEffect(90, 40, testSkullPalette(), "rama")
	s.phase = PhaseGrid
	s.gridStartFrame = 0

	s.frameCount = 0
	if f := s.gridLitFraction(); f != 0 {
		t.Errorf("expected lit=0 at gridStart, got %f", f)
	}
	s.frameCount = 25
	mid := s.gridLitFraction()
	if mid <= 0 || mid >= 1 {
		t.Errorf("expected intermediate lit fraction mid-ignition, got %f", mid)
	}
	s.frameCount = 60
	if f := s.gridLitFraction(); f != 1 {
		t.Errorf("expected lit=1 after ignition, got %f", f)
	}
}

// TestScanSweepOnlyDuringDrip verifies the Act 1 scan line is active during
// PhaseDrip only and sweeps downward with frame count.
func TestScanSweepOnlyDuringDrip(t *testing.T) {
	s := NewSkullEffect(90, 40, testSkullPalette(), "rama")

	s.frameCount = 5
	early := s.scanRow()

	s.frameCount = 30
	later := s.scanRow()

	if early < 0 || later < 0 {
		t.Fatalf("expected scan row active during drip (early=%d, later=%d)", early, later)
	}
	if later <= early {
		t.Fatalf("expected scan row to advance downward (early=%d, later=%d)", early, later)
	}

	s.phase = PhaseHold
	if got := s.scanRow(); got != -1 {
		t.Fatalf("expected no scan row during hold, got %d", got)
	}
}

// TestSonarActiveOnlyDuringHold verifies sonar pulses don't run during drip
// or grid phases but pulse cyclically during hold.
func TestSonarActiveOnlyDuringHold(t *testing.T) {
	s := NewSkullEffect(90, 40, testSkullPalette(), "rama")

	s.phase = PhaseDrip
	if r, _ := s.sonarRadius(); r != 0 {
		t.Errorf("expected no sonar during drip, got radius=%f", r)
	}
	s.phase = PhaseGrid
	if r, _ := s.sonarRadius(); r != 0 {
		t.Errorf("expected no sonar during grid, got radius=%f", r)
	}

	s.phase = PhaseHold
	s.holdStartFrame = 0
	s.frameCount = 10
	r, intensity := s.sonarRadius()
	if r <= 0 || intensity <= 0 || intensity > 1 {
		t.Errorf("expected active sonar during hold early in cycle: radius=%f intensity=%f", r, intensity)
	}
	s.frameCount = 50 // past expandFrames=40, before cycleFrames=80
	if r, _ := s.sonarRadius(); r != 0 {
		t.Errorf("expected sonar dormant between pings, got radius=%f", r)
	}
}

func TestEffectRegistryIncludesSkullAndExcludesSkullText(t *testing.T) {
	names := GetEffectNames()
	if !slices.Contains(names, "skull") {
		t.Fatal("expected effect registry to include skull")
	}
	if slices.Contains(names, "skull-text") {
		t.Fatal("expected effect registry to NOT include skull-text (removed)")
	}

	skullMeta := GetEffectMetadata("skull")
	if skullMeta == nil || skullMeta.RequiresText {
		t.Fatalf("unexpected skull metadata: %#v", skullMeta)
	}
}
