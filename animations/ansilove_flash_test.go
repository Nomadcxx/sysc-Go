package animations

import (
	"math"
	"os"
	"testing"
	"time"
)

func TestANSILoveFlashSafety(t *testing.T) {
	// ponytail: measure deterministic preview glyph masks; terminal-font coverage needs a host capture.
	font, err := os.ReadFile("cmd/bake-ansilove/ibm-vga8.bin")
	if err != nil {
		t.Fatal(err)
	}
	backends := []struct {
		name    string
		backend ANSILoveBackend
	}{
		{"half", ANSILoveHalf},
		{"quad", ANSILoveQuad},
		{"sextant", ANSILoveSextant},
		{"octant", ANSILoveOctant},
		{"braille-2c", ANSILoveBrailleTwoColor},
		{"braille-fg", ANSILoveBrailleFG},
	}

	for _, backend := range backends {
		t.Run(backend.name, func(t *testing.T) {
			newEffect := func(t *testing.T) *ANSILoveAnimation {
				t.Helper()
				a, err := NewANSILoveEffectConfig(ANSILoveConfig{
					Width: 200, Height: 55, Backend: backend.backend, Motion: ANSILoveMotionSubtle,
				})
				if err != nil {
					t.Fatal(err)
				}
				return a
			}
			t.Run("camera-tour", func(t *testing.T) {
				checkANSILoveFlashSafety(t, newEffect(t), font, 20*20)
			})
			t.Run("heart-hold", func(t *testing.T) {
				hold := newEffect(t)
				hold.stage, hold.camera, hold.cameraRow = 4, 1, hold.travel()
				hold.cached = ""
				checkANSILoveFlashSafety(t, hold, font, 59)
			})
		})
	}
}

func checkANSILoveFlashSafety(t *testing.T, a *ANSILoveAnimation, font []byte, frameCount int) {
	t.Helper()
	coverage := map[rune]float64{}
	luminance := map[[3]uint8]float64{}
	previousCells, width, height := parseANSILoveFrame(t, a.Render())
	previous := make([]float64, len(previousCells))
	for i, cell := range previousCells {
		previous[i] = ansiLoveCellLuminance(t, cell, font, coverage, luminance)
	}

	changes := make([][]int8, 0, frameCount)
	wholeScreen := make([]int8, 0, frameCount)
	for frame := 0; frame < frameCount; frame++ {
		a.UpdateElapsed(50 * time.Millisecond)
		cells, w, h := parseANSILoveFrame(t, a.Render())
		if w != width || h != height || len(cells) != len(previous) {
			t.Fatalf("frame %d changed dimensions to %dx%d", frame, w, h)
		}
		directions := make([]int8, len(cells))
		changedCells := 0
		netChange := 0.0
		for i, cell := range cells {
			current := ansiLoveCellLuminance(t, cell, font, coverage, luminance)
			delta := current - previous[i]
			if math.Abs(delta) > 0.1 {
				if delta > 0 {
					directions[i] = 1
				} else {
					directions[i] = -1
				}
				changedCells++
				netChange += delta
			}
			previous[i] = current
		}
		changes = append(changes, directions)
		if changedCells*4 >= len(cells) {
			wholeScreen = append(wholeScreen, signANSILove(netChange))
		} else {
			wholeScreen = append(wholeScreen, 0)
		}
	}

	if got, at := maxANSILoveWindowReversals(wholeScreen, 20); got > 3 {
		t.Fatalf("%d whole-screen luminance reversals in 1 s window starting at frame %d", got, at)
	}
	for cell := range previous {
		if got, at := maxANSILoveCellWindowReversals(changes, cell, 20); got > 3 {
			t.Fatalf("cell %d had %d luminance reversals in 1 s window starting at frame %d", cell, got, at)
		}
	}
}

func ansiLoveCellLuminance(t *testing.T, cell ansiLovePreviewCell, font []byte, coverage map[rune]float64, cache map[[3]uint8]float64) float64 {
	t.Helper()
	ink, ok := coverage[cell.ch]
	if !ok {
		bits := ansiLovePreviewCoverage(t, cell.ch, font)
		for _, on := range bits {
			if on {
				ink += 1.0 / 128
			}
		}
		coverage[cell.ch] = ink
	}
	colorLuminance := func(rgb [3]uint8) float64 {
		if value, ok := cache[rgb]; ok {
			return value
		}
		linear := func(c uint8) float64 {
			v := float64(c) / 255
			if v <= 0.04045 {
				return v / 12.92
			}
			return math.Pow((v+0.055)/1.055, 2.4)
		}
		value := 0.2126*linear(rgb[0]) + 0.7152*linear(rgb[1]) + 0.0722*linear(rgb[2])
		cache[rgb] = value
		return value
	}
	return ink*colorLuminance(cell.fg) + (1-ink)*colorLuminance(cell.bg)
}

func signANSILove(value float64) int8 {
	if value > 0 {
		return 1
	}
	if value < 0 {
		return -1
	}
	return 0
}

func maxANSILoveWindowReversals(directions []int8, window int) (maxCount, start int) {
	frames := make([]int, 0, len(directions)/4)
	previous := int8(0)
	for frame, direction := range directions {
		if direction == 0 {
			continue
		}
		if previous != 0 && direction != previous {
			frames = append(frames, frame)
		}
		previous = direction
	}
	return maxANSILoveWindowEvents(frames, len(directions), window)
}

func maxANSILoveCellWindowReversals(directions [][]int8, cell, window int) (maxCount, start int) {
	frames := make([]int, 0, len(directions)/4)
	previous := int8(0)
	for frame, row := range directions {
		direction := row[cell]
		if direction == 0 {
			continue
		}
		if previous != 0 && direction != previous {
			frames = append(frames, frame)
		}
		previous = direction
	}
	return maxANSILoveWindowEvents(frames, len(directions), window)
}

func maxANSILoveWindowEvents(events []int, frameCount, window int) (maxCount, start int) {
	left, right := 0, 0
	for at := 0; at+window <= frameCount; at++ {
		for left < len(events) && events[left] < at {
			left++
		}
		for right < len(events) && events[right] < at+window {
			right++
		}
		if count := right - left; count > maxCount {
			maxCount, start = count, at
		}
	}
	return
}
