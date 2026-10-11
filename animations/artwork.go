package animations

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

const (
	artBG = iota
	artInk
	artShade
	artMid
	artSkinShade
	artSkin
	artLight
	artWhite
	artBlue
	artCyan
	artViolet
	artBlood
	artBloodShade
	artWarm
	artWarmShade
	artSun
)

type artPoint struct{ x, y float64 }

// ArtworkAnimation draws authored scene geometry into character cells.
// The canvas owns no clock or goroutine; hosts supply active time in Update.
type ArtworkAnimation struct {
	width, height, cellWidth, cellHeight int
	scene, credit                        string
	elapsed, period                      time.Duration
	pixels                               []uint8
	dotW, dotH, dotsY                    int
	sx, sy, ox, oy                       float64
	colors                               [16][3]uint8
	fg, bg                               [16]string
	cached                               string
}

func NewGreatWaveEffect(w, h int, theme string) (*ArtworkAnimation, error) {
	return newArtwork("great-wave", w, h, theme, "Hokusai / The Met", 8*time.Second), nil
}
func NewCyberpunkGirlEffect(w, h int, theme string) (*ArtworkAnimation, error) {
	return newArtwork("cyberpunk-girl", w, h, theme, "Valenberg.com", 12*time.Second), nil
}
func NewWastelandEffect(w, h int, theme string) (*ArtworkAnimation, error) {
	return newArtwork("wasteland", w, h, theme, "Valenberg.com", 16*time.Second), nil
}
func newArtwork(scene string, w, h int, theme, credit string, period time.Duration) *ArtworkAnimation {
	a := &ArtworkAnimation{scene: scene, credit: credit, period: period, cellWidth: 1, cellHeight: 2, dotsY: 2}
	if scene == "great-wave" {
		a.dotsY = 4
	}
	p := GetLogoPalette(theme)
	convert := func(s string) [3]uint8 { r, g, b := hexToRGB(s); return [3]uint8{uint8(r), uint8(g), uint8(b)} }
	background, white := convert(p[6]), convert(p[4])
	// Use existing theme colours nearest the scene's semantic accents.
	palette := append(GetFireworksPalette(theme), GetParticlePalette(theme)...)
	palette = append(palette, p...)
	palette = append(palette, GetFirePalette(theme)...)
	nearest := func(target [3]float64) [3]uint8 {
		best, out := math.Inf(1), white
		for _, s := range palette {
			c := convert(s)
			peak := float64(max(c[0], max(c[1], c[2])))
			if peak == 0 {
				continue
			}
			d := 0.0
			for i := range c {
				v := float64(c[i])/peak - target[i]
				d += v * v
			}
			if d < best {
				best, out = d, c
			}
		}
		return out
	}
	blue := nearest([3]float64{.25, .5, 1})
	cyan := nearest([3]float64{.3, 1, 1})
	violet := nearest([3]float64{.7, .4, 1})
	red := nearest([3]float64{1, .15, .2})
	warm := nearest([3]float64{1, .4, .1})
	sun := nearest([3]float64{1, .8, .4})
	a.colors = [16][3]uint8{background, artMix(background, [3]uint8{}, .65), artMix(background, blue, .22), artMix(background, blue, .50), artMix(background, violet, .65), artMix(violet, white, .66), artMix(violet, white, .86), white, blue, cyan, violet, red, artMix(background, red, .65), warm, artMix(background, warm, .60), artMix(sun, white, .3)}
	if scene == "wasteland" {
		a.colors[artShade] = artMix(background, warm, .20)
		a.colors[artMid] = artMix(background, warm, .36)
		a.colors[artSkinShade] = artMix(background, warm, .72)
		a.colors[artSkin] = artMix(warm, sun, .20)
		a.colors[artLight] = artMix(warm, white, .28)
		a.colors[artViolet] = artMix(background, warm, .84)
		a.colors[artBloodShade] = artMix(background, warm, .46)
	}
	for i, c := range a.colors {
		a.fg[i] = fmt.Sprintf("\x1b[38;2;%d;%d;%dm", c[0], c[1], c[2])
		a.bg[i] = fmt.Sprintf("\x1b[48;2;%d;%d;%dm", c[0], c[1], c[2])
	}
	a.Resize(w, h)
	return a
}
func artMix(a, b [3]uint8, t float64) (out [3]uint8) {
	for i := range out {
		out[i] = uint8(float64(a[i])*(1-t) + float64(b[i])*t + .5)
	}
	return
}
func (a *ArtworkAnimation) Update() { a.UpdateElapsed(50 * time.Millisecond) }
func (a *ArtworkAnimation) UpdateElapsed(dt time.Duration) {
	if dt <= 0 {
		return
	}
	// ponytail: cap dispatch gaps at 250 ms; resumed hosts reset their baseline.
	if dt > 250*time.Millisecond {
		dt = 250 * time.Millisecond
	}
	a.elapsed = (a.elapsed + dt) % a.period
	a.cached = ""
}
func (a *ArtworkAnimation) Reset() { a.elapsed = 0; a.cached = "" }
func (a *ArtworkAnimation) Resize(w, h int) {
	a.width = max(1, min(512, w))
	a.height = max(1, min(256, h))
	a.dotW, a.dotH = a.width*2, a.height*a.dotsY
	n := a.dotW * a.dotH
	if cap(a.pixels) < n {
		a.pixels = make([]uint8, n)
	} else {
		a.pixels = a.pixels[:n]
	}
	a.cached = ""
}
func (a *ArtworkAnimation) SetCellAspect(w, h int) {
	if w > 0 && w <= 1024 && h > 0 && h <= 1024 {
		a.cellWidth, a.cellHeight = w, h
		a.cached = ""
	}
}
func (a *ArtworkAnimation) view(w, h float64, preserve bool) {
	a.sx, a.sy = float64(a.dotW)/w, float64(a.dotH)/h
	a.ox, a.oy = 0, 0
	if preserve {
		aspect := float64(a.width) * float64(a.cellWidth) / (float64(a.height) * float64(a.cellHeight))
		a.sx = float64(a.dotW) / (h * aspect)
		a.ox = (float64(a.dotW) - w*a.sx) / 2
	}
}
func (a *ArtworkAnimation) Render() string {
	if a.cached != "" {
		return a.cached
	}
	clear(a.pixels)
	t := a.elapsed.Seconds()
	switch a.scene {
	case "great-wave":
		a.drawWave(t)
	case "cyberpunk-girl":
		a.drawGirl(t)
	case "wasteland":
		a.drawWasteland(t)
	}
	a.cached = a.compose()
	return a.cached
}

// Quadrant bits are UL, UR, LL, LR. Colour comes from both SGR planes.
var artQuadrants = [16]rune{' ', '▘', '▝', '▀', '▖', '▌', '▞', '▛', '▗', '▚', '▐', '▜', '▄', '▙', '▟', '█'}

func (a *ArtworkAnimation) compose() string {
	var out strings.Builder
	out.Grow(a.width * a.height * 8)
	lastFG, lastBG := -1, -1
	for y := 0; y < a.height; y++ {
		for x := 0; x < a.width; x++ {
			fg, bg := artBG, artBG
			ch := ' '
			if a.dotsY == 4 {
				var bits uint8
				var counts [16]uint8
				for point, bit := range logoDotBits {
					c := a.pixels[(y*4+point/2)*a.dotW+x*2+point%2]
					if c != artBG {
						bits |= bit
						counts[c]++
					}
				}
				if bits != 0 {
					for i := 1; i < len(counts); i++ {
						if counts[i] > counts[fg] {
							fg = i
						}
					}
					ch = rune(0x2800) + rune(bits)
				}
			} else {
				c := [4]uint8{a.pixels[y*2*a.dotW+x*2], a.pixels[y*2*a.dotW+x*2+1], a.pixels[(y*2+1)*a.dotW+x*2], a.pixels[(y*2+1)*a.dotW+x*2+1]}
				bg, fg = int(c[0]), int(c[0])
				best := -1
				// Two colours per cell: retain the furthest pair and assign each quadrant
				// to its nearest colour. The fixed palette bounds raster tile diversity.
				for i := 0; i < 4; i++ {
					for j := i + 1; j < 4; j++ {
						d := artColorDistance(a.colors[c[i]], a.colors[c[j]])
						if d > best {
							best = d
							bg, fg = int(c[i]), int(c[j])
						}
					}
				}
				bits := 0
				for i, v := range c {
					if artColorDistance(a.colors[v], a.colors[fg]) <= artColorDistance(a.colors[v], a.colors[bg]) {
						bits |= 1 << i
					}
				}
				ch = artQuadrants[bits]
				if bg == fg {
					ch = '█'
					bg = artBG
				}
				if fg == artBG && bg == artBG {
					ch = ' '
				}
			}
			if y == a.height-1 && x >= a.width-len(a.credit) && a.width >= len(a.credit) {
				ch = rune(a.credit[x-(a.width-len(a.credit))])
				fg, bg = artLight, artBG
			}
			if ch == ' ' {
				fg = artBG
			}
			if bg != lastBG {
				out.WriteString(a.bg[bg])
				lastBG = bg
			}
			if fg != lastFG {
				out.WriteString(a.fg[fg])
				lastFG = fg
			}
			out.WriteRune(ch)
		}
		if y < a.height-1 {
			out.WriteByte('\n')
		}
	}
	out.WriteString("\x1b[0m")
	return out.String()
}
func artColorDistance(a, b [3]uint8) int {
	d := 0
	for i := range a {
		v := int(a[i]) - int(b[i])
		d += v * v
	}
	return d
}
func (a *ArtworkAnimation) rect(x, y, w, h float64, c uint8) {
	x0, x1 := max(0, int(math.Floor(a.ox+x*a.sx))), min(a.dotW, int(math.Ceil(a.ox+(x+w)*a.sx)))
	y0, y1 := max(0, int(math.Floor(a.oy+y*a.sy))), min(a.dotH, int(math.Ceil(a.oy+(y+h)*a.sy)))
	for yy := y0; yy < y1; yy++ {
		for xx := x0; xx < x1; xx++ {
			a.pixels[yy*a.dotW+xx] = c
		}
	}
}
func (a *ArtworkAnimation) poly(c uint8, p ...artPoint) {
	if len(p) < 3 || len(p) > 128 {
		return
	}
	minY, maxY := p[0].y, p[0].y
	for _, v := range p {
		minY = math.Min(minY, v.y)
		maxY = math.Max(maxY, v.y)
	}
	y0, y1 := max(0, int(math.Floor(a.oy+minY*a.sy))), min(a.dotH, int(math.Ceil(a.oy+maxY*a.sy)))
	var crossings [128]float64
	for row := y0; row < y1; row++ {
		y := (float64(row) + .5 - a.oy) / a.sy
		n := 0
		for i, v := range p {
			q := p[(i+1)%len(p)]
			if (v.y <= y && q.y > y) || (q.y <= y && v.y > y) {
				crossings[n] = a.ox + (v.x+(y-v.y)*(q.x-v.x)/(q.y-v.y))*a.sx
				n++
			}
		}
		sort.Float64s(crossings[:n])
		for i := 0; i+1 < n; i += 2 {
			x0, x1 := max(0, int(math.Ceil(crossings[i]-.5))), min(a.dotW, int(math.Ceil(crossings[i+1]-.5)))
			for x := x0; x < x1; x++ {
				a.pixels[row*a.dotW+x] = c
			}
		}
	}
}
func (a *ArtworkAnimation) ellipse(x, y, rx, ry float64, c uint8) {
	if rx <= 0 || ry <= 0 {
		return
	}
	y0, y1 := max(0, int(math.Floor(a.oy+(y-ry)*a.sy))), min(a.dotH, int(math.Ceil(a.oy+(y+ry)*a.sy)))
	for row := y0; row < y1; row++ {
		v := ((float64(row)+.5-a.oy)/a.sy - y) / ry
		if math.Abs(v) > 1 {
			continue
		}
		span := rx * math.Sqrt(1-v*v)
		x0, x1 := max(0, int(math.Ceil(a.ox+(x-span)*a.sx-.5))), min(a.dotW, int(math.Ceil(a.ox+(x+span)*a.sx-.5)))
		for col := x0; col < x1; col++ {
			a.pixels[row*a.dotW+col] = c
		}
	}
}
func (a *ArtworkAnimation) line(p, q artPoint, w float64, c uint8) {
	dx, dy := q.x-p.x, q.y-p.y
	length := math.Hypot(dx, dy)
	if length == 0 {
		rx, ry := math.Max(w/2, .5/a.sx), math.Max(w/2, .5/a.sy)
		a.rect(p.x-rx, p.y-ry, rx*2, ry*2, c)
		return
	}
	nx, ny := -dy/length, dx/length
	sampleRadius := .5
	if a.dotsY == 4 {
		// ponytail: one foreground per braille cell needs two-sample strokes;
		// separate background/dot colours if finer water contours are required.
		sampleRadius = 1
	}
	radius := math.Max(w/2, sampleRadius*math.Hypot(nx/a.sx, ny/a.sy))
	capX, capY := dx/length*radius, dy/length*radius
	p.x, p.y = p.x-capX, p.y-capY
	q.x, q.y = q.x+capX, q.y+capY
	// Clip in the bounded polygon scan, independent of off-screen line length.
	a.poly(c, artPoint{p.x + nx*radius, p.y + ny*radius},
		artPoint{q.x + nx*radius, q.y + ny*radius},
		artPoint{q.x - nx*radius, q.y - ny*radius},
		artPoint{p.x - nx*radius, p.y - ny*radius})
}
func artBezier(p [4]artPoint, t float64) artPoint {
	u := 1 - t
	return artPoint{u*u*u*p[0].x + 3*u*u*t*p[1].x + 3*u*t*t*p[2].x + t*t*t*p[3].x, u*u*u*p[0].y + 3*u*u*t*p[1].y + 3*u*t*t*p[2].y + t*t*t*p[3].y}
}
func (a *ArtworkAnimation) curve(p [4]artPoint, w float64, c uint8) {
	previous := p[0]
	for i := 1; i <= 48; i++ {
		next := artBezier(p, float64(i)/48)
		a.line(previous, next, w, c)
		previous = next
	}
}
func artPhase(t, period float64) float64 { return t/period - math.Floor(t/period) }
