package animations

import (
	"bytes"
	"math"
	"strconv"
)

const ansiReset = "\033[0m"

type rgbColor struct {
	r, g, b int
}

type ansiColor struct {
	rgbColor
	prefix string
}

func hexToRGB(hex string) (int, int, int) {
	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) != 6 {
		return 0, 0, 0
	}

	var rgb rgbColor
	values := [3]*int{&rgb.r, &rgb.g, &rgb.b}
	for i, value := range values {
		high, ok := hexDigit(hex[i*2])
		if !ok {
			return 0, 0, 0
		}
		low, ok := hexDigit(hex[i*2+1])
		if !ok {
			return 0, 0, 0
		}
		*value = high<<4 | low
	}
	return rgb.r, rgb.g, rgb.b
}

func hexDigit(b byte) (int, bool) {
	switch {
	case b >= '0' && b <= '9':
		return int(b - '0'), true
	case b >= 'a' && b <= 'f':
		return int(b-'a') + 10, true
	case b >= 'A' && b <= 'F':
		return int(b-'A') + 10, true
	default:
		return 0, false
	}
}

func cacheANSIColors(palette []string) []ansiColor {
	colors := make([]ansiColor, len(palette))
	for i, hex := range palette {
		r, g, b := hexToRGB(hex)
		colors[i] = newANSIColor(rgbColor{r: r, g: g, b: b})
	}
	return colors
}

func newANSIColor(color rgbColor) ansiColor {
	var prefix [24]byte
	encoded := appendSGR(prefix[:0], color.r, color.g, color.b)
	return ansiColor{rgbColor: color, prefix: string(encoded)}
}

func appendSGR(dst []byte, r, g, b int) []byte {
	dst = append(dst, "\033[38;2;"...)
	dst = strconv.AppendInt(dst, int64(r), 10)
	dst = append(dst, ';')
	dst = strconv.AppendInt(dst, int64(g), 10)
	dst = append(dst, ';')
	dst = strconv.AppendInt(dst, int64(b), 10)
	return append(dst, 'm')
}

func writeANSIRune(dst *bytes.Buffer, color ansiColor, ch rune) {
	dst.WriteString(color.prefix)
	dst.WriteRune(ch)
	dst.WriteString(ansiReset)
}

func writeRGBRune(dst *bytes.Buffer, color rgbColor, ch rune) {
	var prefix [24]byte
	dst.Write(appendSGR(prefix[:0], color.r, color.g, color.b))
	dst.WriteRune(ch)
	dst.WriteString(ansiReset)
}

func lerpRGB(from, to rgbColor, t float64) rgbColor {
	return rgbColor{
		r: from.r + int(math.Round(t*float64(to.r-from.r))),
		g: from.g + int(math.Round(t*float64(to.g-from.g))),
		b: from.b + int(math.Round(t*float64(to.b-from.b))),
	}
}
