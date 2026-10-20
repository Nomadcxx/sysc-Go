package ansiart

import "math"

func partition(c []LinRGB, mask uint8) (fg, bg LinRGB, err float64) {
	n := 0
	for i, p := range c {
		if mask&(1<<i) != 0 {
			n++
			for j := range p {
				fg[j] += p[j]
			}
		} else {
			for j := range p {
				bg[j] += p[j]
			}
		}
	}
	for j := range fg {
		if n > 0 {
			fg[j] /= float64(n)
		}
		if len(c) > n {
			bg[j] /= float64(len(c) - n)
		}
	}
	for i, p := range c {
		if mask&(1<<i) != 0 {
			err += distance(p, fg)
		} else {
			err += distance(p, bg)
		}
	}
	return
}

// Fit2 minimizes linear-light squared error, exactly for up to six samples.
func Fit2(c []LinRGB, exact bool) (mask uint8, fg, bg LinRGB) {
	if len(c) == 0 {
		return
	}
	if len(c) > 8 {
		c = c[:8]
	}
	if exact || len(c) <= 6 {
		best := math.Inf(1)
		for m := 0; m < 1<<(len(c)-1); m++ {
			a, b, e := partition(c, uint8(m))
			if e < best {
				best, mask, fg, bg = e, uint8(m), a, b
			}
		}
		return
	}
	// A deterministic farthest-pair seed avoids dependence on input brightness.
	far := -1.0
	for i, a := range c {
		for _, b := range c[i+1:] {
			if d := distance(a, b); d > far {
				far, fg, bg = d, a, b
			}
		}
	}
	for iteration := 0; iteration < 8; iteration++ {
		next := uint8(0)
		for i, p := range c {
			if distance(p, fg) < distance(p, bg) {
				next |= 1 << i
			}
		}
		a, b, _ := partition(c, next)
		stable := iteration > 0 && next == mask
		mask, fg, bg = next, a, b
		if stable {
			break
		}
	}
	return
}

var blockGlyphs = []rune{' ', '░', '▒', '▓', '█', '▀', '▄', '▌', '▐', '▘', '▝', '▖', '▗', '▚', '▞', '▛', '▜', '▙', '▟'}

func blockInk(glyph rune, x, y int) bool {
	switch glyph {
	case ' ':
		return false
	case '█':
		return true
	case '░':
		return y%2 == 0 && x%2 == 0
	case '▒':
		return (x+y)%2 == 0
	case '▓':
		return !(y%2 == 0 && x%2 == 0)
	}
	mask := 0
	for i, g := range QuadRunes {
		if g == glyph {
			mask = i
			break
		}
	}
	bit := x/4 + 2*(y/8)
	return mask&(1<<bit) != 0
}

var vgaLinear = func() (p [16]LinRGB) {
	for i, c := range VGA {
		p[i] = ToLinear(c)
	}
	return
}()

// Fit16 searches all foreground/background VGA pairs for the block and shade
// repertoire. Palette error separates by partition; the patch's squared sum is
// constant, so partition sums alone determine the exact minimum.
func Fit16(patch []LinRGB) (glyph rune, fg, bg uint8) {
	if len(patch) != 128 {
		return ' ', 0, 0
	}
	var quadrants [4]LinRGB
	var shades [2]LinRGB
	var total LinRGB
	for i, p := range patch {
		q := (i%8)/4 + 2*(i/64)
		for j, v := range p {
			quadrants[q][j] += v
			total[j] += v
			if (i/8)%2 == 0 && i%2 == 0 {
				shades[0][j] += v
			}
			if (i/8+i%8)%2 == 0 {
				shades[1][j] += v
			}
		}
	}
	best := math.Inf(1)
	for _, g := range blockGlyphs {
		sum := LinRGB{}
		count := 0
		switch g {
		case '░':
			sum, count = shades[0], 32
		case '▒':
			sum, count = shades[1], 64
		case '▓':
			count = 96
			for j := range sum {
				sum[j] = total[j] - shades[0][j]
			}
		default:
			mask := 0
			for m, r := range QuadRunes {
				if r == g {
					mask = m
					break
				}
			}
			for q, p := range quadrants {
				if mask&(1<<q) != 0 {
					count += 32
					for j := range sum {
						sum[j] += p[j]
					}
				}
			}
		}
		var indices [2]uint8
		error := 0.0
		for n := 0; n < 2; n++ {
			part, nPixels := sum, count
			if n == 0 {
				nPixels = 128 - count
				for j := range part {
					part[j] = total[j] - sum[j]
				}
			}
			score := math.Inf(1)
			for i, c := range vgaLinear {
				e := float64(nPixels)*(c[0]*c[0]+c[1]*c[1]+c[2]*c[2]) - 2*(c[0]*part[0]+c[1]*part[1]+c[2]*part[2])
				if e < score {
					score, indices[n] = e, uint8(i)
				}
			}
			error += score
		}
		if error < best {
			best, glyph, fg, bg = error, g, indices[1], indices[0]
		}
	}
	return
}
