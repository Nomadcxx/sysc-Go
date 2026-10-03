package animations

import (
	"math/rand"
	"strings"
	"testing"
)

func TestMatrixArtCompletionFramesFreezeEveryArtCell(t *testing.T) {
	const width, height = 12, 10
	matrix := NewMatrixArtEffect(width, height, []string{"#00ff00"}, "HI\nGO")
	if got, want := matrix.CompletionFrames(), 48; got != want {
		t.Fatalf("CompletionFrames() = %d, want %d", got, want)
	}

	for frame := 0; frame < matrix.CompletionFrames(); frame++ {
		matrix.Update()
	}
	for y, row := range matrix.artPositions {
		for x := range row {
			if matrix.frozenChars[y] == nil || matrix.frozenChars[y][x] == nil {
				t.Fatalf("art cell (%d,%d) was not frozen by its completion bound", x, y)
			}
		}
	}
	if !matrix.IsComplete() {
		t.Fatal("matrix art did not report completion after every art cell froze")
	}
}

func TestBeamTextCompletionFramesReachHold(t *testing.T) {
	lines := make([]string, 32)
	for i := range lines {
		lines[i] = "X"
	}

	for seed := int64(0); seed < 12; seed++ {
		beam := NewBeamTextEffect(BeamTextConfig{
			Width:                1,
			Height:               len(lines),
			Text:                 strings.Join(lines, "\n"),
			Auto:                 true,
			Display:              true,
			BeamDelay:            2,
			BeamRowSpeedRange:    [2]int{20, 21},
			BeamColumnSpeedRange: [2]int{15, 16},
			FinalWipeSpeed:       3,
		})
		beam.rng = rand.New(rand.NewSource(seed))

		for frame := 0; frame < beam.CompletionFrames() && beam.phase != "hold"; frame++ {
			beam.Update()
		}
		if beam.phase != "hold" {
			t.Fatalf("seed %d did not reach hold within its completion bound", seed)
		}
	}
}
