package ansiart

import (
	"math"
	"sync"
)

type LinRGB [3]float64

// VGA is the original IBM palette in ANSI SGR order.
var VGA = [16][3]uint8{
	{0, 0, 0}, {170, 0, 0}, {0, 170, 0}, {170, 85, 0}, {0, 0, 170}, {170, 0, 170}, {0, 170, 170}, {170, 170, 170},
	{85, 85, 85}, {255, 85, 85}, {85, 255, 85}, {255, 255, 85}, {85, 85, 255}, {255, 85, 255}, {85, 255, 255}, {255, 255, 255},
}

func ToLinear(c [3]uint8) (out LinRGB) {
	for i, v := range c {
		s := float64(v) / 255
		if s <= .04045 {
			out[i] = s / 12.92
		} else {
			out[i] = math.Pow((s+.055)/1.055, 2.4)
		}
	}
	return
}
func ToSRGB(c LinRGB) (out [3]uint8) {
	for i, v := range c {
		v = math.Max(0, math.Min(1, v))
		if v <= .0031308 {
			v *= 12.92
		} else {
			v = 1.055*math.Pow(v, 1/2.4) - .055
		}
		out[i] = uint8(math.Round(v * 255))
	}
	return
}

var quantOnce sync.Once
var quantLUT [32 * 32 * 32]uint8
var xterm [256][3]uint8
var vgaLab [16]LinRGB

func oklab(c LinRGB) LinRGB {
	l := math.Cbrt(.4122214708*c[0] + .5363325363*c[1] + .0514459929*c[2])
	m := math.Cbrt(.2119034982*c[0] + .6806995451*c[1] + .1073969566*c[2])
	s := math.Cbrt(.0883024619*c[0] + .2817188376*c[1] + .6299787005*c[2])
	return LinRGB{.2104542553*l + .793617785*m - .0040720468*s, 1.9779984951*l - 2.428592205*m + .4505937099*s, .0259040371*l + .7827717662*m - .808675766*s}
}
func distance(a, b LinRGB) float64 {
	d := 0.0
	for i := range a {
		v := a[i] - b[i]
		d += v * v
	}
	return d
}

func initQuant() {
	var lab [256]LinRGB
	for i, c := range VGA {
		xterm[i] = c
		vgaLab[i] = oklab(ToLinear(c))
	}
	levels := [6]uint8{0, 95, 135, 175, 215, 255}
	for r := 0; r < 6; r++ {
		for g := 0; g < 6; g++ {
			for b := 0; b < 6; b++ {
				xterm[16+36*r+6*g+b] = [3]uint8{levels[r], levels[g], levels[b]}
			}
		}
	}
	for i := 0; i < 24; i++ {
		v := uint8(8 + i*10)
		xterm[232+i] = [3]uint8{v, v, v}
	}
	for i := 16; i < 256; i++ {
		lab[i] = oklab(ToLinear(xterm[i]))
	}
	for r := 0; r < 32; r++ {
		for g := 0; g < 32; g++ {
			for b := 0; b < 32; b++ {
				c := oklab(ToLinear([3]uint8{uint8((r*255 + 15) / 31), uint8((g*255 + 15) / 31), uint8((b*255 + 15) / 31)}))
				best, idx := math.Inf(1), 16
				for i := 16; i < 256; i++ {
					if d := distance(c, lab[i]); d < best {
						best, idx = d, i
					}
				}
				quantLUT[(r*32+g)*32+b] = uint8(idx)
			}
		}
	}
}

// Quantize uses truecolor (0), xterm entries 16..255 (1), or VGA (2).
func Quantize(c LinRGB, depth int) (rgb [3]uint8, index uint8) {
	rgb = ToSRGB(c)
	if depth == 0 {
		return rgb, 0
	}
	quantOnce.Do(initQuant)
	if depth == 1 {
		r, g, b := (int(rgb[0])*31+127)/255, (int(rgb[1])*31+127)/255, (int(rgb[2])*31+127)/255
		index = quantLUT[(r*32+g)*32+b]
		return xterm[index], index
	}
	lab := oklab(c)
	best := math.Inf(1)
	for i, p := range vgaLab {
		if d := distance(lab, p); d < best {
			best, index = d, uint8(i)
		}
	}
	return VGA[index], index
}
