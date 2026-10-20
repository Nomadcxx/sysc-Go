package ansiart

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
)

const (
	RegionBackground uint8 = iota
	RegionHeader
	RegionUpper
	RegionLogo
	RegionHeart
	RegionEKG
	RegionCredits
)

const maxRasterPixels = 16 << 20

type Raster struct {
	W, H        int
	Idx, Region []uint8
}

// Rasterize uses the eight VGA glyph columns; labels only coloured pixels.
func Rasterize(g *Grid, font *[256][16]uint8) *Raster {
	if g == nil || font == nil || g.W < 1 || g.H < 1 || g.W > 1024 || g.H > 4096 || g.W*g.H > maxRasterPixels/128 || len(g.Cells) != g.W*g.H {
		return nil
	}
	r := &Raster{W: g.W * 8, H: g.H * 16, Idx: make([]uint8, g.W*g.H*128), Region: make([]uint8, g.W*g.H*128)}
	glyphs := make(map[rune]uint8, 256)
	for i, c := range cp437 {
		glyphs[c] = uint8(i)
	}
	for cy := 0; cy < g.H; cy++ {
		for cx := 0; cx < g.W; cx++ {
			c := g.Cells[cy*g.W+cx]
			if c.Fg < 0 || c.Fg > 15 || c.Bg < 0 || c.Bg > 15 {
				return nil
			}
			glyph := font[glyphs[c.Ch]]
			for gy, bits := range glyph {
				for gx := 0; gx < 8; gx++ {
					index := (cy*16+gy)*r.W + cx*8 + gx
					colour := uint8(c.Bg)
					if bits&(0x80>>gx) != 0 {
						colour = uint8(c.Fg)
					}
					r.Idx[index] = colour
					if colour == 0 {
						continue
					}
					switch {
					case cy >= 3 && cy <= 13:
						r.Region[index] = RegionHeader
					case cy >= 104 && cy <= 126 && (colour == 7 || colour == 8 || colour == 15):
						r.Region[index] = RegionLogo
					case cy >= 132 && cy <= 141 && (colour == 1 || colour == 9):
						// ponytail: Keep the provisional EKG band static until its exact mask is reviewed.
						r.Region[index] = RegionEKG
					case cy >= 126 && cy <= 149 && (colour == 1 || colour == 9):
						r.Region[index] = RegionHeart
					case cy >= 15 && cy <= 106:
						r.Region[index] = RegionUpper
					case cy >= 151 && cy <= 155:
						r.Region[index] = RegionCredits
					}
				}
			}
		}
	}
	return r
}

func validRaster(r *Raster) error {
	if r == nil || r.W < 1 || r.H < 1 || r.W > 65536 || r.H > 65536 || r.W > maxRasterPixels/r.H {
		return fmt.Errorf("invalid raster dimensions")
	}
	n := r.W * r.H
	if len(r.Idx) != n || len(r.Region) != n {
		return fmt.Errorf("invalid raster plane lengths")
	}
	for i, c := range r.Idx {
		if c > 15 || r.Region[i] > RegionCredits {
			return fmt.Errorf("invalid raster palette or region at %d", i)
		}
	}
	return nil
}

// PackRaster writes ALR1, little-endian dimensions, palette and region planes
// inside one deterministic zlib stream.
func PackRaster(r *Raster) ([]byte, error) {
	if err := validRaster(r); err != nil {
		return nil, err
	}
	var out bytes.Buffer
	z := zlib.NewWriter(&out)
	header := make([]byte, 12)
	copy(header, "ALR1")
	binary.LittleEndian.PutUint32(header[4:], uint32(r.W))
	binary.LittleEndian.PutUint32(header[8:], uint32(r.H))
	for _, b := range [][]byte{header, r.Idx, r.Region} {
		if _, err := z.Write(b); err != nil {
			return nil, err
		}
	}
	if err := z.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// UnpackRaster validates dimensions before allocating and rejects truncated,
// oversized or trailing streams, invalid palette values and unknown regions.
func UnpackRaster(data []byte) (*Raster, error) {
	src := bytes.NewReader(data)
	z, err := zlib.NewReader(src)
	if err != nil {
		return nil, fmt.Errorf("raster zlib: %w", err)
	}
	defer z.Close()
	var header [12]byte
	if _, err := io.ReadFull(z, header[:]); err != nil {
		return nil, fmt.Errorf("raster header: %w", err)
	}
	w, h := uint64(binary.LittleEndian.Uint32(header[4:])), uint64(binary.LittleEndian.Uint32(header[8:]))
	if string(header[:4]) != "ALR1" || w < 1 || h < 1 || w > 65536 || h > 65536 || w*h > maxRasterPixels {
		return nil, fmt.Errorf("invalid raster header")
	}
	n := int(w * h)
	r := &Raster{W: int(w), H: int(h), Idx: make([]byte, n), Region: make([]byte, n)}
	for _, b := range [][]byte{r.Idx, r.Region} {
		if _, err := io.ReadFull(z, b); err != nil {
			return nil, fmt.Errorf("raster plane: %w", err)
		}
	}
	extra, err := io.ReadAll(io.LimitReader(z, 1))
	if err != nil {
		return nil, fmt.Errorf("raster checksum: %w", err)
	}
	if len(extra) != 0 || src.Len() != 0 {
		return nil, fmt.Errorf("trailing raster data")
	}
	if err := validRaster(r); err != nil {
		return nil, err
	}
	return r, nil
}
