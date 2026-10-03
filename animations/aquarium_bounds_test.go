package animations

import (
	"fmt"
	"testing"
)

func testAquarium(width, height int) *AquariumEffect {
	return NewAquariumEffect(AquariumConfig{
		Width:         width,
		Height:        height,
		FishColors:    []string{"#ff0000"},
		WaterColors:   []string{"#0000ff", "#c2b280"},
		SeaweedColors: []string{"#00ff00"},
		BubbleColor:   "#ffffff",
		DiverColor:    "#ffffff",
		BoatColor:     "#ffffff",
		MermaidColor:  "#ffffff",
	})
}

func exerciseAquarium(a *AquariumEffect) {
	for i := 0; i < 30; i++ {
		a.Update()
	}
	_ = a.Render()
}

func TestAquariumShortTerminalDoesNotPanic(t *testing.T) {
	sizes := [][2]int{
		{80, 23}, // large-fish Intn span is 0
		{80, 20}, // large-fish Intn argument is negative
		{80, 14}, // medium-fish Intn span is 0
		{80, 4},  // tiny/small fish and bubble spans collapse
		{80, 2},  // seaweed Intn(height/3) and ocean row forced to 2
		{80, 1},
		{80, 0},
		{0, 24}, // Intn(width) in init
		{10, 10},
	}

	for _, sz := range sizes {
		w, h := sz[0], sz[1]
		t.Run(fmt.Sprintf("%dx%d", w, h), func(t *testing.T) {
			a := testAquarium(w, h)
			exerciseAquarium(a)
			a.Resize(w, h)
			exerciseAquarium(a)
		})
	}
}

func TestAquariumNormalSizeStillRenders(t *testing.T) {
	a := testAquarium(80, 40)
	exerciseAquarium(a)
	if a.Render() == "" {
		t.Fatal("expected a frame from an 80x40 aquarium")
	}
}

func TestAquariumResizeToShortTerminalDoesNotPanic(t *testing.T) {
	a := testAquarium(80, 40)
	a.Resize(18, 20)
	exerciseAquarium(a)
	_ = a.Render()
}
