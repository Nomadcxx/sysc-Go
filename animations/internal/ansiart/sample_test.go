package ansiart

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestLayoutAspectAndCaps(t *testing.T) {
	a := ComputeLayout(640, 2512, 200, 55, 8, 16, 2, 4, FitPan, 2.5, 3)
	b := ComputeLayout(640, 2512, 200, 55, 16, 32, 2, 4, FitPan, 2.5, 3)
	if a != b {
		t.Fatalf("cell-ratio equivalent layouts differ: %+v %+v", a, b)
	}
	if a.SheetRows != 138 || a.Width < 70 || a.Width > 72 {
		t.Fatalf("wrong pan layout: %+v", a)
	}
	c := ComputeLayout(640, 2512, 9000, 9000, 8, 16, 2, 4, FitWidth, 2.5, 3)
	if c.Cols != 512 || c.Rows != 256 || c.Scale != 3 || c.Width != 240 {
		t.Fatalf("unbounded layout: %+v", c)
	}
}

func TestExactAreaWeights(t *testing.T) {
	r := &Raster{W: 2, H: 2, Idx: []uint8{0, 15, 15, 0}, Region: make([]uint8, 4)}
	l := Layout{Cols: 1, Rows: 1, SX: 1, SY: 1, SheetRows: 1, CellW: 2, CellH: 2, Scale: 1, Width: 1}
	w := BuildWeights(r, l)
	lut := make([]LinRGB, 112)
	lut[15] = LinRGB{1, 1, 1}
	c := Sample(w, 0, lut)
	if math.Abs(c[0]-.5) > 1.0/65535 {
		t.Fatalf("checker mean %v", c)
	}
	if ToSRGB(c)[0] != 188 {
		t.Fatalf("linear-light mean encoded as %v", ToSRGB(c))
	}
	for n := 0; n < len(w.Start)-1; n++ {
		s := 0
		for _, v := range w.W[w.Start[n]:w.Start[n+1]] {
			s += int(v)
		}
		if s != 65535 {
			t.Fatalf("weight sum %d", s)
		}
	}
}

func TestAreaPreservesAllSlotsAndFractionalEdges(t *testing.T) {
	r := &Raster{W: 8, H: 1, Idx: []uint8{0, 1, 2, 3, 4, 5, 6, 7}, Region: make([]uint8, 8)}
	l := Layout{Cols: 1, Rows: 1, SX: 1, SY: 1, SheetRows: 1, CellW: 8, CellH: 1, Scale: 1, Width: 1}
	w := BuildWeights(r, l)
	if len(w.Slot) != 8 {
		t.Fatalf("lost palette slots: %v", w.Slot)
	}
	l.CellW = .5
	l.OX = .75
	w = BuildWeights(r, l)
	if len(w.Slot) != 2 || w.Slot[0] != 0 || w.Slot[1] != 1 || math.Abs(float64(w.W[0])-32767.5) > .5 {
		t.Fatalf("fractional overlap: %+v", w)
	}
	l.OX = 1
	l.CellW = .5
	w = BuildWeights(r, l)
	if len(w.Slot) != 1 || w.Slot[0] != 1 || w.W[0] != 65535 {
		t.Fatalf("integer upscale: %+v", w)
	}
}

func TestFit2AndQuantization(t *testing.T) {
	black, white := LinRGB{}, LinRGB{1, 1, 1}
	mask, fg, bg := Fit2([]LinRGB{white, black, white, black}, true)
	for i, c := range []LinRGB{white, black, white, black} {
		got := bg
		if mask&(1<<i) != 0 {
			got = fg
		}
		if got != c {
			t.Fatalf("fit %d %v %v", mask, fg, bg)
		}
	}
	rgb, idx := Quantize(ToLinear(VGA[9]), 2)
	if rgb != VGA[9] || idx != 9 {
		t.Fatalf("red palette %v %d", rgb, idx)
	}
	rgb, idx = Quantize(white, 1)
	if rgb != [3]uint8{255, 255, 255} || idx < 16 {
		t.Fatalf("xterm white %v %d", rgb, idx)
	}
}

func TestGlyphCoverage(t *testing.T) {
	if QuadRunes[5] != '▌' || SextantRune(21) != '▌' || OctantRune(85) != '▌' || BrailleRune(2) != 0x2808 {
		t.Fatal("glyph bit ordering")
	}
	for i := 0; i < 256; i++ {
		if OctantRune(uint8(i)) == 0 {
			t.Fatalf("missing octant %d", i)
		}
	}
	for i := 0; i < 64; i++ {
		if SextantRune(uint8(i)) == 0 {
			t.Fatalf("missing sextant %d", i)
		}
	}
}

func TestFit16MatchesSolidAndShade(t *testing.T) {
	patch := make([]LinRGB, 128)
	for i := range patch {
		patch[i] = ToLinear(VGA[9])
	}
	_, fg, bg := Fit16(patch)
	if fg != 9 && bg != 9 {
		t.Fatalf("solid red lost: %d %d", fg, bg)
	}
	for y := 0; y < 16; y++ {
		for x := 0; x < 8; x++ {
			if (x+y)%2 == 0 {
				patch[y*8+x] = ToLinear(VGA[15])
			} else {
				patch[y*8+x] = LinRGB{}
			}
		}
	}
	glyph, fg, bg := Fit16(patch)
	if glyph != '▒' || !((fg == 15 && bg == 0) || (fg == 0 && bg == 15)) {
		t.Fatalf("shade fit %q %d %d", glyph, fg, bg)
	}
}

func TestSampleCircleAspect(t *testing.T) {
	const size = 96
	r := &Raster{W: size, H: size, Idx: make([]uint8, size*size), Region: make([]uint8, size*size)}
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if math.Hypot(float64(x)-47.5, float64(y)-47.5) <= 40 {
				r.Idx[y*size+x] = 15
			}
		}
	}
	lut := make([]LinRGB, 112)
	lut[15] = LinRGB{1, 1, 1}
	for _, cell := range [][2]int{{8, 16}, {10, 20}, {12, 24}, {8, 18}, {7, 16}, {10, 24}} {
		l := ComputeLayout(size, size, 30, 20, cell[0], cell[1], 2, 4, FitContain, 1, 3)
		w := BuildWeights(r, l)
		minX, minY, maxX, maxY := 1<<30, 1<<30, 0, 0
		for y := 0; y < l.SheetRows*l.SY; y++ {
			for x := 0; x < l.Width*l.SX; x++ {
				if Sample(w, y*l.Width*l.SX+x, lut)[0] > .5 {
					minX = min(minX, x)
					maxX = max(maxX, x)
					minY = min(minY, y)
					maxY = max(maxY, y)
				}
			}
		}
		width := float64(maxX-minX+1) * l.CellW / float64(l.SX)
		height := float64(maxY-minY+1) * l.CellH / float64(l.SY)
		if math.Abs(width/height-1) > .06 {
			t.Fatalf("circle distorted with %v: %.2fx%.2f", cell, width, height)
		}
	}
}

func TestSampleCellPatchUniform(t *testing.T) {
	r := &Raster{W: 8, H: 16, Idx: make([]uint8, 128), Region: make([]uint8, 128)}
	for i := range r.Idx {
		r.Idx[i] = 9
		r.Region[i] = 4
	}
	lut := make([]LinRGB, 112)
	lut[4*16+9] = ToLinear(VGA[9])
	l := ComputeLayout(8, 16, 1, 1, 8, 16, 2, 2, FitContain, 1, 3)
	var out [128]LinRGB
	SampleCellPatch(r, l, 0, 0, lut, out[:])
	for _, p := range out {
		if p != lut[73] {
			t.Fatalf("patch changed uniform color %v", p)
		}
	}
}

func BenchmarkBuildWeights(b *testing.B) {
	r := &Raster{W: 640, H: 2512, Idx: make([]uint8, 640*2512), Region: make([]uint8, 640*2512)}
	for i := range r.Idx {
		r.Idx[i] = uint8((i / 8) % 16)
	}
	for _, size := range [][2]int{{200, 55}, {480, 135}} {
		l := ComputeLayout(r.W, r.H, size[0], size[1], 8, 16, 2, 4, FitPan, 2.5, 3)
		b.Run(fmt.Sprintf("%dx%d", size[0], size[1]), func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				BuildWeights(r, l)
			}
		})
	}
}

func BenchmarkFit16(b *testing.B) {
	var patch [128]LinRGB
	for i := range patch {
		patch[i] = ToLinear(VGA[i%16])
	}
	for n := 0; n < b.N; n++ {
		Fit16(patch[:])
	}
}

func TestFit16MatchesBruteForce(t *testing.T) {
	var patch [128]LinRGB
	for i := range patch {
		patch[i] = LinRGB{float64((i*13)%97) / 96, float64((i*29)%83) / 82, float64((i*7)%71) / 70}
	}
	g, fg, bg := Fit16(patch[:])
	got := 0.0
	for i, p := range patch {
		c := vgaLinear[bg]
		if blockInk(g, i%8, i/8) {
			c = vgaLinear[fg]
		}
		got += distance(p, c)
	}
	want := math.Inf(1)
	for _, glyph := range blockGlyphs {
		for _, front := range vgaLinear {
			for _, back := range vgaLinear {
				e := 0.0
				for i, p := range patch {
					c := back
					if blockInk(glyph, i%8, i/8) {
						c = front
					}
					e += distance(p, c)
				}
				want = math.Min(want, e)
			}
		}
	}
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("optimized search %.12f != full search %.12f", got, want)
	}
}

func TestGlyphNamesMatchMasks(t *testing.T) {
	// Generated comments retain authoritative Unicode names, allowing every
	// non-legacy mask to be checked independently of table enumeration order.
	data, err := os.ReadFile("glyphs_gen.go")
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, line := range strings.Split(string(data), "\n") {
		f := strings.Fields(line)
		if len(f) < 5 || f[3] != "BLOCK" {
			continue
		}
		name := f[4]
		var table []rune
		switch {
		case strings.HasPrefix(name, "OCTANT-"):
			table = octantRunes[:]
			name = strings.TrimPrefix(name, "OCTANT-")
		case strings.HasPrefix(name, "SEXTANT-"):
			table = sextantRunes[:]
			name = strings.TrimPrefix(name, "SEXTANT-")
		default:
			continue
		}
		mask := 0
		for _, d := range name {
			mask |= 1 << (d - '1')
		}
		code, err := strconv.ParseInt(strings.TrimSuffix(f[0], ","), 0, 32)
		if err != nil {
			t.Fatal(err)
		}
		if table[mask] != rune(code) {
			t.Fatalf("Unicode %s mask%x maps to %U instead of %U", f[4], mask, table[mask], rune(code))
		}
		count++
	}
	if count != 290 {
		t.Fatalf("checked %d Unicode names, expected290", count)
	}
	legacy := map[uint8]rune{0: ' ', 1: 0x1cea8, 2: 0x1ceab, 3: 0x1fb82, 5: '▘', 10: '▝', 15: '▀', 20: 0x1fbe6, 40: 0x1fbe7, 63: 0x1fb85, 64: 0x1cea3, 80: '▖', 85: '▌', 90: '▞', 95: '▛', 128: 0x1cea0, 160: '▗', 165: '▚', 170: '▐', 175: '▜', 192: '▂', 240: '▄', 245: '▙', 250: '▟', 252: '▆', 255: '█'}
	for m, r := range legacy {
		if OctantRune(m) != r {
			t.Fatalf("legacy octant%x: %U", m, OctantRune(m))
		}
	}
	for i := 0; i < 256; i++ {
		if BrailleRune(uint8(i)) < 0x2800 || BrailleRune(uint8(i)) > 0x28ff {
			t.Fatalf("invalid braille%d", i)
		}
	}
}

func BenchmarkSourceWeights(b *testing.B) {
	data, err := os.ReadFile("../../artwork/ansi-love.raster.zlib")
	if err != nil {
		b.Fatal(err)
	}
	r, err := UnpackRaster(data)
	if err != nil {
		b.Fatal(err)
	}
	for _, size := range [][2]int{{200, 55}, {480, 135}} {
		l := ComputeLayout(r.W, r.H, size[0], size[1], 8, 16, 2, 4, FitPan, 2.5, 3)
		b.Run(fmt.Sprintf("%dx%d", size[0], size[1]), func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				w := BuildWeights(r, l)
				b.ReportMetric(float64(cap(w.Start)*4+cap(w.W)*2+cap(w.Slot)), "retained-B")
			}
		})
	}
}

func BenchmarkSampleCellPatch(b *testing.B) {
	data, err := os.ReadFile("../../artwork/ansi-love.raster.zlib")
	if err != nil {
		b.Fatal(err)
	}
	r, err := UnpackRaster(data)
	if err != nil {
		b.Fatal(err)
	}
	l := ComputeLayout(r.W, r.H, 200, 55, 8, 16, 2, 2, FitPan, 2.5, 3)
	var lut [112]LinRGB
	for i := range lut {
		lut[i] = ToLinear(VGA[i%16])
	}
	var patch [128]LinRGB
	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		SampleCellPatch(r, l, 20, 60, lut[:], patch[:])
		Fit16(patch[:])
	}
}

func TestPatchAgreesWithAreaWeights(t *testing.T) {
	r := &Raster{W: 9, H: 17, Idx: make([]uint8, 9*17), Region: make([]uint8, 9*17)}
	for i := range r.Idx {
		r.Idx[i] = uint8(i % 16)
		r.Region[i] = uint8(i % 7)
	}
	var lut [112]LinRGB
	for i := range lut {
		lut[i] = ToLinear(VGA[i%16])
	}
	l := ComputeLayout(9, 17, 2, 2, 8, 18, 8, 16, FitPan, 1.5, 3)
	w := BuildWeights(r, l)
	var patch [128]LinRGB
	for row := 0; row < l.SheetRows; row++ {
		for col := 0; col < l.Width; col++ {
			SampleCellPatch(r, l, col, row, lut[:], patch[:])
			for y := 0; y < 16; y++ {
				for x := 0; x < 8; x++ {
					want := Sample(w, (row*16+y)*l.Width*8+col*8+x, lut[:])
					if distance(want, patch[y*8+x]) > 1e-8 {
						t.Fatalf("patch disagrees at%d,%d", x, y)
					}
				}
			}
		}
	}
}

func TestLayoutBoundsExtremeCellRatios(t *testing.T) {
	for _, fit := range []FitMode{FitContain, FitWidth, FitPan, FitExact} {
		for _, cell := range [][2]int{{1024, 1}, {1, 1024}, {5, 1}, {37, 25}, {8, 16}} {
			l := ComputeLayout(640, 2512, 512, 256, cell[0], cell[1], 2, 4, fit, 3, 3)
			if l.SheetRows > 768 {
				t.Fatalf("fit%d cell%v: unbounded sheet %+v", fit, cell, l)
			}
			if l.Scale > 3 || l.Scale <= 0 {
				t.Fatalf("invalid scale %+v", l)
			}
			if float64(l.SheetRows)*l.CellH+1e-9 < 2512*l.Scale {
				t.Fatalf("sheet cropped source %+v", l)
			}
		}
	}
}
