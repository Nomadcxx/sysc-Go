package ansiart

import "math"

type Weights struct {
	Start []uint32
	Slot  []uint8
	W     []uint16
}

// BuildWeights integrates every intersecting source pixel, retaining all region
// palette slots. Samples run left-to-right through each subpixel scanline.
func BuildWeights(r *Raster, l Layout) *Weights {
	n := l.Width * l.SheetRows * l.SX * l.SY
	w := &Weights{Start: make([]uint32, 0, n+1), Slot: make([]uint8, 0, n), W: make([]uint16, 0, n)}
	if l.Scale <= 0 || n == 0 {
		w.Start = append(w.Start, 0)
		return w
	}
	dx, dy := l.CellW/float64(l.SX)/l.Scale, l.CellH/float64(l.SY)/l.Scale
	xstart := l.OX + float64(l.Left)*l.CellW/l.Scale
	for y := 0; y < l.SheetRows*l.SY; y++ {
		for x := 0; x < l.Width*l.SX; x++ {
			w.Start = append(w.Start, uint32(len(w.Slot)))
			areas := rectAreas(r, xstart+float64(x)*dx, l.OY+float64(y)*dy, dx, dy)
			total := dx * dy
			// Cumulative rounding preserves the exact uint16 unity sum deterministically.
			cumulative, previous := 0.0, 0
			for _, slot := range areas.slots[:areas.n] {
				area := areas.weights[slot]
				cumulative += area
				next := min(65535, int(math.Round(cumulative/total*65535)))
				weight := next - previous
				previous = next
				if weight > 0 {
					w.Slot = append(w.Slot, uint8(slot))
					w.W = append(w.W, uint16(weight))
				}
			}
		}
	}
	w.Start = append(w.Start, uint32(len(w.Slot)))
	return w
}

type areaWeights struct {
	weights [256]float64
	slots   [256]uint8
	n       int
}

func (a *areaWeights) add(slot uint8, value float64) {
	if value <= 0 {
		return
	}
	if a.weights[slot] == 0 {
		a.slots[a.n] = slot
		a.n++
	}
	a.weights[slot] += value
}

func rectAreas(r *Raster, x0, y0, dx, dy float64) (areas areaWeights) {
	x1, y1 := x0+dx, y0+dy
	total := dx * dy
	if r == nil || r.W <= 0 || r.H <= 0 || len(r.Idx) != r.W*r.H || len(r.Region) != len(r.Idx) {
		areas.add(0, total)
		return
	}
	left, right := math.Max(0, x0), math.Min(float64(r.W), x1)
	top, bottom := math.Max(0, y0), math.Min(float64(r.H), y1)
	covered := math.Max(0, right-left) * math.Max(0, bottom-top)
	areas.add(0, math.Max(0, total-covered))
	for y := max(0, int(math.Floor(top))); y < min(r.H, int(math.Ceil(bottom))); y++ {
		ay := math.Max(0, math.Min(bottom, float64(y+1))-math.Max(top, float64(y)))
		for x := max(0, int(math.Floor(left))); x < min(r.W, int(math.Ceil(right))); x++ {
			ax := math.Max(0, math.Min(right, float64(x+1))-math.Max(left, float64(x)))
			p := y*r.W + x
			slot := (r.Region[p] << 4) | (r.Idx[p] & 15)
			areas.add(slot, ax*ay)
		}
	}
	return
}

func Sample(w *Weights, n int, lut []LinRGB) (c LinRGB) {
	for k := w.Start[n]; k < w.Start[n+1]; k++ {
		slot := int(w.Slot[k])
		if slot >= len(lut) {
			continue
		}
		weight := float64(w.W[k]) / 65535
		for j := range c {
			c[j] += lut[slot][j] * weight
		}
	}
	return
}

// SampleCellPatch produces an 8x16 patch without retaining large 16-colour sheets.
func SampleCellPatch(r *Raster, l Layout, col, row int, lut []LinRGB, out []LinRGB) {
	if len(out) < 128 || l.Scale <= 0 || r == nil || r.W <= 0 || r.H <= 0 || len(r.Idx) != r.W*r.H || len(r.Region) != len(r.Idx) {
		return
	}
	dx, dy := l.CellW/8/l.Scale, l.CellH/16/l.Scale
	x0 := l.OX + float64(l.Left+col)*l.CellW/l.Scale
	y0 := l.OY + float64(row)*l.CellH/l.Scale
	for y := 0; y < 16; y++ {
		for x := 0; x < 8; x++ {
			left, top := x0+float64(x)*dx, y0+float64(y)*dy
			right, bottom := left+dx, top+dy
			left = math.Max(0, left)
			right = math.Min(float64(r.W), right)
			top = math.Max(0, top)
			bottom = math.Min(float64(r.H), bottom)
			c := LinRGB{}
			area := dx * dy
			outside := 1 - math.Max(0, right-left)*math.Max(0, bottom-top)/area
			if len(lut) > 0 {
				for j := range c {
					c[j] = lut[0][j] * outside
				}
			}
			for py := max(0, int(math.Floor(top))); py < min(r.H, int(math.Ceil(bottom))); py++ {
				ay := math.Max(0, math.Min(bottom, float64(py+1))-math.Max(top, float64(py)))
				for px := max(0, int(math.Floor(left))); px < min(r.W, int(math.Ceil(right))); px++ {
					ax := math.Max(0, math.Min(right, float64(px+1))-math.Max(left, float64(px)))
					p := py*r.W + px
					slot := int((r.Region[p] << 4) | (r.Idx[p] & 15))
					if slot < len(lut) {
						weight := ax * ay / area
						for j := range c {
							c[j] += lut[slot][j] * weight
						}
					}
				}
			}
			out[y*8+x] = c
		}
	}
}
