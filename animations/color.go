package animations

import (
	"math"
	"strconv"
	"strings"
)

type rgbColor [3]uint8

const ansiReset = "\033[0m"
const ansiFrameEstimateBytesPerCell = 8

func decodeHexRGB(hex string) rgbColor {
	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) != 6 {
		return rgbColor{}
	}

	var color rgbColor
	for i := range color {
		hi := hexDigit(hex[i*2])
		lo := hexDigit(hex[i*2+1])
		if hi < 0 || lo < 0 {
			return rgbColor{}
		}
		color[i] = uint8(hi<<4 | lo)
	}
	return color
}

func hexDigit(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	default:
		return -1
	}
}

func hexToRGB(hex string) (int, int, int) {
	color := decodeHexRGB(hex)
	return int(color[0]), int(color[1]), int(color[2])
}

func parsePalette(palette []string) []rgbColor {
	colors := make([]rgbColor, len(palette))
	for i, hex := range palette {
		colors[i] = decodeHexRGB(hex)
	}
	return colors
}

func parseANSIPalette(palette []string) []string {
	prefixes := make([]string, len(palette))
	for i, hex := range palette {
		prefixes[i] = ansiColorPrefix(decodeHexRGB(hex))
	}
	return prefixes
}

func growANSIFrame(builder *strings.Builder, width, height int) {
	if width <= 0 || height <= 0 {
		return
	}
	maxInt := int(^uint(0) >> 1)
	if height > maxInt/ansiFrameEstimateBytesPerCell {
		return
	}
	rowCapacity := height * ansiFrameEstimateBytesPerCell
	if width > (maxInt-height)/rowCapacity {
		return
	}
	// ponytail: reserve a measured 8 bytes per cell instead of the 28-byte
	// escaped-cell ceiling. Dense frames may grow; raise after profiling them.
	builder.Grow(width*rowCapacity + height)
}

func lerpRGB(from, to rgbColor, t float64) rgbColor {
	if t <= 0 {
		return from
	}
	if t >= 1 {
		return to
	}

	var color rgbColor
	for i := range color {
		color[i] = uint8(int(from[i]) + int(math.Round(t*float64(int(to[i])-int(from[i])))))
	}
	return color
}

func writeANSIColorPrefix(builder *strings.Builder, color rgbColor) {
	var prefix [20]byte
	builder.Write(appendANSIColorPrefix(prefix[:0], color))
}

func ansiColorPrefix(color rgbColor) string {
	var prefix [20]byte
	return string(appendANSIColorPrefix(prefix[:0], color))
}

func appendANSIColorPrefix(prefix []byte, color rgbColor) []byte {
	prefix = append(prefix, "\033[38;2;"...)
	for i, component := range color {
		if i != 0 {
			prefix = append(prefix, ';')
		}
		prefix = strconv.AppendUint(prefix, uint64(component), 10)
	}
	return append(prefix, 'm')
}

func writeANSIColorRune(builder *strings.Builder, color rgbColor, char rune) {
	writeANSIColorPrefix(builder, color)
	builder.WriteRune(char)
	builder.WriteString(ansiReset)
}

func writeANSIColorRuneWithPrefix(builder *strings.Builder, prefix string, char rune) {
	builder.WriteString(prefix)
	builder.WriteRune(char)
	builder.WriteString(ansiReset)
}
