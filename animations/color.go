package animations

import (
	"bytes"
	"math"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss/v2"
)

const ansiReset = "\033[0m"
const lipglossCellReset = "\033[m"

// ponytail: Cache up to 256 colors per effect; overflow colors render uncached.
// Add bounded eviction if a real palette needs reuse beyond this cap.
const maxCellColorCacheEntries = 256

type rgbColor struct {
	r, g, b int
}

type ansiColor struct {
	rgbColor
	prefix string
}

func hexToRGB(hex string) (int, int, int) {
	color, ok := decodeHexColor(hex)
	if !ok {
		return 0, 0, 0
	}
	return color.r, color.g, color.b
}

func decodeHexColor(hex string) (rgbColor, bool) {
	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) != 6 {
		return rgbColor{}, false
	}

	var rgb rgbColor
	values := [3]*int{&rgb.r, &rgb.g, &rgb.b}
	for i, value := range values {
		high, ok := hexDigit(hex[i*2])
		if !ok {
			return rgbColor{}, false
		}
		low, ok := hexDigit(hex[i*2+1])
		if !ok {
			return rgbColor{}, false
		}
		*value = high<<4 | low
	}
	return rgb, true
}

type cachedCellColor struct {
	prefix string
	suffix string
}

type cellColorCache map[string]cachedCellColor

var blankCellBlock = strings.Repeat(" ", 1024)

func resetFrameBuilder(builder *strings.Builder) {
	// String keeps the old buffer immutable; use its length to size the next frame after Reset.
	lastFrameSize := builder.Len()
	builder.Reset()
	if lastFrameSize > 0 {
		builder.Grow(lastFrameSize)
	}
}

func writeBlankSpan(dst *strings.Builder, count int) {
	for count >= len(blankCellBlock) {
		dst.WriteString(blankCellBlock)
		count -= len(blankCellBlock)
	}
	dst.WriteString(blankCellBlock[:count])
}

func writeFrameRow(dst *strings.Builder, chars []rune, colors []string, cache *cellColorCache, styleSpaces bool) {
	for x := 0; x < len(chars); {
		char := chars[x]
		color := ""
		if x < len(colors) {
			color = colors[x]
		}
		if char == ' ' && (!styleSpaces || color == "") {
			start := x
			for x < len(chars) && chars[x] == ' ' {
				if styleSpaces && x < len(colors) && colors[x] != "" {
					break
				}
				x++
			}
			writeBlankSpan(dst, x-start)
			continue
		}
		if color != "" && (styleSpaces || char != ' ') {
			cache.writeRune(dst, color, char)
		} else {
			dst.WriteRune(char)
		}
		x++
	}
}

func (c *cellColorCache) writeRune(dst *strings.Builder, color string, ch rune) {
	if *c == nil {
		*c = make(cellColorCache)
	}

	entry, ok := (*c)[color]
	if !ok {
		if rgb, valid := decodeHexColor(color); valid && strings.HasPrefix(color, "#") {
			var prefix [24]byte
			entry.prefix = string(appendSGR(prefix[:0], rgb.r, rgb.g, rgb.b))
			entry.suffix = lipglossCellReset
		} else {
			styled := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render("x")
			if i := strings.IndexByte(styled, 'x'); i >= 0 {
				entry.prefix, entry.suffix = styled[:i], styled[i+1:]
			}
		}
		if len(*c) < maxCellColorCacheEntries {
			(*c)[color] = entry
		}
	}

	dst.WriteString(entry.prefix)
	dst.WriteRune(ch)
	dst.WriteString(entry.suffix)
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
