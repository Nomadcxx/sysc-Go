package animations

import (
	"math"
	"testing"
)

func TestMatrixSpawnsAcrossWideGridAtCapacity(t *testing.T) {
	const width = 10000
	m := MatrixEffect{width: width, height: 24, streaks: make([]MatrixStreak, 149, 150)}
	for i := range m.streaks {
		m.streaks[i] = MatrixStreak{X: 0, Speed: math.MaxInt, Active: true}
	}
	var seen [4]bool
	// With uniform placement, missing a quarter in 256 samples has probability
	// below 4*(3/4)^256. The old left-to-right scan fills its last slot near zero.
	for sample := 0; sample < 256; sample++ {
		m.streaks = m.streaks[:149]
		m.Update()
		if len(m.streaks) != 150 {
			t.Fatalf("stream limit changed: got %d", len(m.streaks))
		}
		x := m.streaks[149].X
		if x < 0 || x >= width {
			t.Fatalf("spawn outside grid: %d", x)
		}
		seen[x/(width/4)] = true
	}
	for quarter, visited := range seen {
		if !visited {
			t.Errorf("no streams spawned in quarter %d at capacity", quarter)
		}
	}
}
