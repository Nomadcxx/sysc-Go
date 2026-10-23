package animations

import (
	"image"
	"image/color"
	"math"
	"testing"
)

func TestMeshRasterImageKeepsDotColorsAndTransparency(t *testing.T) {
	depth := make([]float64, 8)
	for i := range depth {
		depth[i] = math.Inf(-1)
	}
	depth[0], depth[1] = 1, 1
	canvas := &LogoSpinAnimation{
		width:  1,
		height: 1,
		depth:  depth,
		crgb:   []uint32{0x123456, 0xabcdef, 0, 0, 0, 0, 0, 0},
	}

	got := canvas.rasterImage()
	if got.Bounds() != image.Rect(0, 0, 2, 4) {
		t.Fatalf("image bounds = %v, want (0,0)-(2,4)", got.Bounds())
	}
	if want := (color.NRGBA{R: 0x12, G: 0x34, B: 0x56, A: 0xff}); got.NRGBAAt(0, 0) != want {
		t.Fatalf("first dot = %#v, want %#v", got.NRGBAAt(0, 0), want)
	}
	if want := (color.NRGBA{R: 0xab, G: 0xcd, B: 0xef, A: 0xff}); got.NRGBAAt(1, 0) != want {
		t.Fatalf("second dot = %#v, want %#v", got.NRGBAAt(1, 0), want)
	}
	if got := got.NRGBAAt(0, 1); got.A != 0 {
		t.Fatalf("unoccupied dot alpha = %d, want 0", got.A)
	}
}

func TestL2A1RenderFrameHasPixelsAndCredit(t *testing.T) {
	frame := NewL2A1SAREffect(4, 4, "Dracula").RenderFrame()
	if frame.Pixels.Bounds() != image.Rect(0, 0, 8, 16) {
		t.Fatalf("image bounds = %v, want (0,0)-(8,16)", frame.Pixels.Bounds())
	}
	if frame.Credit != "L2A1 SAR / notcplkerry / CC BY 4.0" {
		t.Fatalf("credit = %q", frame.Credit)
	}
	for y := 0; y < frame.Pixels.Bounds().Dy(); y++ {
		for x := 0; x < frame.Pixels.Bounds().Dx(); x++ {
			if frame.Pixels.NRGBAAt(x, y).A != 0 {
				return
			}
		}
	}
	t.Fatal("rendered mesh frame has no occupied dots")
}
