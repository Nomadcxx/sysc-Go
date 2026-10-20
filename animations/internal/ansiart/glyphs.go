package ansiart

//go:generate go run ./genglyphs

func SextantRune(mask uint8) rune { return sextantRunes[mask&63] }
func OctantRune(mask uint8) rune  { return octantRunes[mask] }

// BrailleRune maps row-major 2x4 sample bits to Unicode's column-major dots.
func BrailleRune(mask uint8) rune {
	bits := [8]uint8{0, 3, 1, 4, 2, 5, 6, 7}
	r := rune(0x2800)
	for i, b := range bits {
		if mask&(1<<i) != 0 {
			r |= 1 << b
		}
	}
	return r
}
