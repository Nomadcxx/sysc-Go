package termcell

import "testing"

type measuredEffect struct{ w, h int }

func (e *measuredEffect) SetCellAspect(w, h int) { e.w, e.h = w, h }

func TestApplySize(t *testing.T) {
	for _, tc := range []struct {
		name                     string
		cols, rows, px, py, w, h int
	}{
		{"kitty", 88, 72, 792, 1440, 9, 20},
		{"font resize", 100, 30, 1100, 720, 11, 24},
		{"no pixels", 80, 24, 0, 0, 1, 2},
		{"no columns", 0, 24, 800, 480, 1, 2},
		{"no rows", 80, 0, 800, 480, 1, 2},
		{"bad geometry", 80, 24, 1, 1, 1, 2},
		{"oversized cells", 1, 1, 65535, 65535, 1, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := &measuredEffect{1, 2}
			applySize(e, tc.cols, tc.rows, tc.px, tc.py)
			if e.w != tc.w || e.h != tc.h {
				t.Fatalf("got %dx%d want %dx%d", e.w, e.h, tc.w, tc.h)
			}
		})
	}
	applySize(nil, 80, 24, 800, 480)
	applySize(struct{}{}, 80, 24, 800, 480)
}

func TestUnavailableTerminalKeepsAspect(t *testing.T) {
	e := &measuredEffect{9, 20}
	Apply(-1, e)
	if e.w != 9 || e.h != 20 {
		t.Fatalf("invalid terminal replaced known aspect: %dx%d", e.w, e.h)
	}
}
