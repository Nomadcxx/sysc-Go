package animations

import (
	"fmt"
	"testing"
)

func testFireworks(width, height int) *FireworksEffect {
	return NewFireworksEffect(width, height, []string{"#ffffff", "#ff0000"})
}

func exerciseFireworks(fw *FireworksEffect) {
	for i := 0; i < 50; i++ {
		fw.Update()
		_ = fw.Render()
	}
}

func TestFireworksNarrowTerminalDoesNotPanic(t *testing.T) {
	sizes := [][2]int{
		{18, 24}, // Intn(width-20) is negative
		{20, 24}, // Intn(0)
		{21, 24}, // smallest width the old formula accepted
		{40, 2},  // Intn(height/3) is 0
		{40, 0},
		{1, 1},
		{0, 24},
	}

	for _, sz := range sizes {
		w, h := sz[0], sz[1]
		t.Run(fmt.Sprintf("%dx%d", w, h), func(t *testing.T) {
			fw := testFireworks(w, h)
			exerciseFireworks(fw)
		})
	}
}

func TestFireworksResizeToNarrowTerminalDoesNotPanic(t *testing.T) {
	fw := testFireworks(80, 24)
	fw.Resize(10, 8)
	exerciseFireworks(fw)
}
