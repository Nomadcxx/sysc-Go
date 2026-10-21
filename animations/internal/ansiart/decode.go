// Package ansiart decodes the bounded, offline ANSI Love source into pixel planes.
package ansiart

import (
	"encoding/binary"
	"strings"
)

type Cell struct {
	Ch     rune
	Fg, Bg int8
}
type Grid struct {
	W, H  int
	Cells []Cell
}
type Sauce struct {
	Title, Author, Group, Date string
	DataType, FileType         uint8
	TInfo1, TInfo2             uint16
	TFlags                     uint8
	TInfoS                     string
	Comments                   []string
}

// ParseSAUCE removes the trailing metadata, optional comments and DOS EOF.
func ParseSAUCE(data []byte) ([]byte, *Sauce) {
	if len(data) < 128 || string(data[len(data)-128:len(data)-121]) != "SAUCE00" {
		return data, nil
	}
	b := data[len(data)-128:]
	text := func(b []byte) string { return strings.TrimRight(string(b), " \x00") }
	s := &Sauce{Title: text(b[7:42]), Author: text(b[42:62]), Group: text(b[62:82]), Date: text(b[82:90]), DataType: b[94], FileType: b[95], TInfo1: binary.LittleEndian.Uint16(b[96:98]), TInfo2: binary.LittleEndian.Uint16(b[98:100]), TFlags: b[105], TInfoS: text(b[106:128])}
	end := len(data) - 128
	n := int(b[104])
	start := end - 5 - 64*n
	if n > 0 && start >= 0 && string(data[start:start+5]) == "COMNT" {
		for i := 0; i < n; i++ {
			s.Comments = append(s.Comments, text(data[start+5+i*64:start+5+(i+1)*64]))
		}
		end = start
	}
	if end > 0 && data[end-1] == 0x1a {
		end--
	}
	return data[:end], s
}

// DecodeANSI implements the ANSI.SYS subset used by the artwork. Cursor motion
// is clamped, controls are consumed, and malformed input cannot grow indefinitely.
func DecodeANSI(body []byte, width int, ice bool) *Grid {
	const maxHeight = 4096
	width = max(1, min(width, 1024))
	g := &Grid{W: width, H: 1}
	blank := Cell{' ', 7, 0}
	grow := func(y int) {
		for len(g.Cells) < (y+1)*width {
			g.Cells = append(g.Cells, blank)
		}
	}
	grow(0)
	x, y := 0, 0
	fg, bg := int8(7), int8(0)
	bold, blink := false, false
	for i := 0; i < len(body); i++ {
		b := body[i]
		if b == 0x1b {
			if i+1 >= len(body) {
				break
			}
			i++
			switch body[i] {
			case '[':
				start := i + 1
				for i++; i < len(body) && (body[i] < 0x40 || body[i] > 0x7e); i++ {
				}
				if i >= len(body) {
					break
				}
				// Saturation avoids overflow even for maliciously long numeric arguments.
				args := []int{0}
				valid := true
				for _, c := range body[start:i] {
					if c == ';' {
						if len(args) >= 32 {
							valid = false
							break
						}
						args = append(args, 0)
					} else if c >= '0' && c <= '9' {
						j := len(args) - 1
						args[j] = min(100000, args[j]*10+int(c-'0'))
					} else {
						valid = false
						break
					}
				}
				if !valid {
					continue
				}
				switch body[i] {
				case 'C':
					if x >= width {
						x = 0
						y++
						if y >= maxHeight {
							return g
						}
					}
					x = min(width, x+max(1, args[0]))
				case 'm':
					for _, v := range args {
						switch {
						case v == 0:
							fg, bg, bold, blink = 7, 0, false, false
						case v == 1:
							bold = true
						case v == 22:
							bold = false
						case v == 5:
							blink = true
						case v == 25:
							blink = false
						case v >= 30 && v <= 37:
							fg = int8(v - 30)
						case v == 39:
							fg = 7
						case v >= 40 && v <= 47:
							bg = int8(v - 40)
						case v == 49:
							bg = 0
						}
					}
				}
			case ']', 'P', '^', '_':
				for i++; i < len(body); i++ {
					if body[i] == 7 {
						break
					}
					if body[i] == 27 && i+1 < len(body) && body[i+1] == '\\' {
						i++
						break
					}
				}
			}
			continue
		}
		switch b {
		case '\r':
			x = 0
			continue
		case '\n':
			y++
			if y >= maxHeight {
				return g
			}
			continue
		case 0x1a:
			return g
		case 0, 7, 8, 9:
			continue
		}
		if x >= width {
			x = 0
			y++
		}
		if y >= maxHeight {
			return g
		}
		grow(y)
		g.H = y + 1
		f, bgidx := fg, bg
		if bold {
			f += 8
		}
		if ice && blink {
			bgidx += 8
		}
		g.Cells[y*width+x] = Cell{cp437[b], f, bgidx}
		x++
	}
	return g
}

// The IBM PC character repertoire includes visible glyphs below ASCII space.
var cp437 = []rune("\u0000☺☻♥♦♣♠•◘○◙♂♀♪♫☼►◄↕‼¶§▬↨↑↓→←∟↔▲▼ !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~⌂ÇüéâäàåçêëèïîìÄÅÉæÆôöòûùÿÖÜ¢£¥₧ƒáíóúñÑªº¿⌐¬½¼¡«»░▒▓│┤╡╢╖╕╣║╗╝╜╛┐└┴┬├─┼╞╟╚╔╩╦╠═╬╧╨╤╥╙╘╒╓╫╪┘┌█▄▌▐▀αßΓπΣσµτΦΘΩδ∞φε∩≡±≥≤⌠⌡÷≈°∙·√ⁿ²■ ")
