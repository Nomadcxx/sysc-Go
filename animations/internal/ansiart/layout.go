package ansiart

import "math"

type FitMode uint8

const (
	FitContain FitMode = iota
	FitWidth
	FitPan
	FitExact
)

type Layout struct {
	Cols, Rows, SX, SY, SheetRows int
	Scale, OX, OY                 float64
	CellW, CellH                  float64
	Left, Width                   int
}

// ComputeLayout measures physical pixels in units of an eight-pixel-wide cell.
// Only the reported cell aspect ratio affects the source scale.
func ComputeLayout(srcW, srcH, cols, rows, cw, ch, sx, sy int, fit FitMode, zoom, sMax float64) Layout {
	cols = min(512, max(1, cols))
	rows = min(256, max(1, rows))
	if cw <= 0 || ch <= 0 {
		cw, ch = 8, 16
	}
	l := Layout{Cols: cols, Rows: rows, SX: min(8, max(1, sx)), SY: min(16, max(1, sy)), CellW: 8, CellH: 8 * float64(ch) / float64(cw)}
	if srcW <= 0 || srcH <= 0 {
		return l
	}
	if !finite(zoom) || zoom < 1 {
		zoom = 2.5
	}
	if !finite(sMax) || sMax <= 0 {
		sMax = 3
	}
	pw, ph := float64(cols)*l.CellW, float64(rows)*l.CellH
	sw, sh := float64(srcW), float64(srcH)
	l.Scale = math.Min(pw/sw, ph/sh)
	switch fit {
	case FitWidth:
		l.Scale = pw / sw
	case FitPan:
		l.Scale = math.Min(pw/sw, zoom*ph/sh)
	case FitExact:
		l.Scale = 1
	}
	l.Scale = math.Min(l.Scale, sMax)
	// Bound work even for extreme cell aspect ratios; uniformly shrink the
	// complete artwork rather than truncating its tall sheet.
	l.Scale = math.Min(l.Scale, 768*l.CellH/sh)
	offset := (pw - sw*l.Scale) / 2
	l.OX = -offset / l.Scale
	l.Left = max(0, int(math.Floor(offset/l.CellW)))
	right := min(cols, int(math.Ceil((offset+sw*l.Scale)/l.CellW)))
	l.Width = max(0, right-l.Left)
	// Clamp ceil roundoff at the already applied physical-height limit.
	l.SheetRows = min(768, max(rows, int(math.Ceil(sh*l.Scale/l.CellH))))
	return l
}

func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
