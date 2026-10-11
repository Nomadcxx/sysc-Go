package main

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"testing"
)

func TestArtworkGIFBounds(t *testing.T) {
	frame := image.NewPaletted(image.Rect(0, 0, 1, 1), color.Palette{color.Black, color.White})
	g := &gif.GIF{}
	for i := 0; i < 161; i++ {
		g.Image = append(g.Image, frame)
		g.Delay = append(g.Delay, 12)
	}
	var b bytes.Buffer
	if err := gif.EncodeAll(&b, g); err != nil {
		t.Fatal(err)
	}
	if err := checkGIFFrames(bytes.NewReader(b.Bytes())); err == nil {
		t.Fatal("oversized GIF accepted before allocation")
	}
	g.Image = g.Image[:160]
	g.Delay = g.Delay[:160]
	b.Reset()
	if err := gif.EncodeAll(&b, g); err != nil {
		t.Fatal(err)
	}
	if err := checkGIFFrames(bytes.NewReader(b.Bytes())); err != nil {
		t.Fatal(err)
	}
}

func TestArtworkGIFDisposal(t *testing.T) {
	p := color.Palette{color.RGBA{}, color.RGBA{255, 0, 0, 255}, color.RGBA{0, 255, 0, 255}}
	frame := func(x int, c uint8) *image.Paletted {
		f := image.NewPaletted(image.Rect(x, 0, x+1, 1), p)
		f.Pix[0] = c
		return f
	}
	g := &gif.GIF{Config: image.Config{ColorModel: p, Width: 3, Height: 1}, Image: []*image.Paletted{frame(0, 1), frame(1, 2), frame(2, 1), frame(1, 2), frame(2, 0)}, Delay: []int{12, 12, 12, 12, 12}, Disposal: []byte{1, 3, 1, 2, 1}}
	want := [][3]uint8{{1, 0, 0}, {1, 2, 0}, {1, 0, 1}, {1, 2, 1}, {1, 0, 1}}
	err := walkGIF(g, func(i int, f image.Image) error {
		for x, c := range want[i] {
			if f.At(x, 0) != p[c] {
				t.Fatalf("frame %d cell %d: %v, want %v", i, x, f.At(x, 0), p[c])
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
