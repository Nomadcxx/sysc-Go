package ansiart

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"testing"
)

func TestRasterGlyphAndRegions(t *testing.T) {
	var font [256][16]uint8
	font[219][0] = 0x81
	g := &Grid{W: 1, H: 150, Cells: make([]Cell, 150)}
	for i := range g.Cells {
		g.Cells[i] = Cell{' ', 7, 0}
	}
	g.Cells[126] = Cell{'█', 1, 0}
	g.Cells[135] = Cell{'█', 9, 0}
	r := Rasterize(g, &font)
	if r.W != 8 || r.H != 2400 || r.Idx[126*16*8] != 1 || r.Idx[126*16*8+7] != 1 || r.Idx[126*16*8+1] != 0 {
		t.Fatal("glyph bit order")
	}
	if r.Region[126*16*8] != RegionHeart || r.Region[126*16*8+1] != RegionBackground {
		t.Fatal("mask recolours negative space")
	}
	if got := r.Region[135*16*8]; got != RegionEKG {
		t.Fatalf("unreviewed EKG band region=%d, want static EKG region %d", got, RegionEKG)
	}
	data, err := PackRaster(r)
	if err != nil {
		t.Fatal(err)
	}
	r2, err := UnpackRaster(data)
	if err != nil || r2.W != r.W || r2.H != r.H || !bytes.Equal(r2.Idx, r.Idx) || !bytes.Equal(r2.Region, r.Region) {
		t.Fatal("roundtrip", err)
	}
	again, _ := PackRaster(r)
	if !bytes.Equal(data, again) {
		t.Fatal("nondeterministic compression")
	}
}

func TestRasterRejectMalformed(t *testing.T) {
	for _, r := range []*Raster{nil, {}, {W: 1, H: 1}, {W: 1, H: 1, Idx: []byte{16}, Region: []byte{0}}, {W: 1, H: 1, Idx: []byte{0}, Region: []byte{7}}} {
		if _, err := PackRaster(r); err == nil {
			t.Fatal("accepted malformed raster")
		}
	}
	valid, _ := PackRaster(&Raster{W: 1, H: 1, Idx: []byte{1}, Region: []byte{0}})
	for _, b := range [][]byte{nil, []byte("bogus"), valid[:len(valid)-1], append(append([]byte{}, valid...), 1)} {
		if _, err := UnpackRaster(b); err == nil {
			t.Fatal("accepted malformed stream")
		}
	}
	var out bytes.Buffer
	z := zlib.NewWriter(&out)
	raw := make([]byte, 14)
	copy(raw, "ALR1")
	binary.LittleEndian.PutUint32(raw[4:], 0xffffffff)
	binary.LittleEndian.PutUint32(raw[8:], 0xffffffff)
	z.Write(raw)
	z.Close()
	if _, err := UnpackRaster(out.Bytes()); err == nil {
		t.Fatal("accepted oversized raster")
	}
}
