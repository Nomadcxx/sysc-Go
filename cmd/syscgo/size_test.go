package main

import (
	"errors"
	"testing"
)

func TestFallbackTerminalSizeTreatsZeroWinsizeAsMissing(t *testing.T) {
	cases := []struct {
		name          string
		width, height int
		err           error
		wantW, wantH  int
	}{
		{name: "error", width: 0, height: 0, err: errors.New("not a tty"), wantW: 80, wantH: 24},
		{name: "zero height", width: 80, height: 0, err: nil, wantW: 80, wantH: 24},
		{name: "zero width", width: 0, height: 24, err: nil, wantW: 80, wantH: 24},
		{name: "both zero", width: 0, height: 0, err: nil, wantW: 80, wantH: 24},
		{name: "real size", width: 100, height: 40, err: nil, wantW: 100, wantH: 40},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotW, gotH := fallbackTerminalSize(tc.width, tc.height, tc.err)
			if gotW != tc.wantW || gotH != tc.wantH {
				t.Fatalf("fallbackTerminalSize(%d, %d, %v) = %d, %d; want %d, %d",
					tc.width, tc.height, tc.err, gotW, gotH, tc.wantW, tc.wantH)
			}
		})
	}
}
