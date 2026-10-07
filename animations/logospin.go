package animations

import (
	"fmt"
	"math"
	"strings"
	"sync"
)

// logoSpinFrames is one spin loop in frames (~7.2 s at the 20 fps both hosts
// drive effects at): two eased full turns, each covering 82% of a half-loop
// followed by a facing-forward hold. logoMorphPeriodFrames is a full morph
// cycle.
const (
	logoGrid            = 160
	logoExtent          = 1.08
	logoStep            = logoExtent * 2.0 / float64(logoGrid-1)
	logoHalfDepth       = 0.1642880353482906
	logoWobble          = 0.11242630996974184
	logoGrainPhase      = 3.8990964158506114
	logoSpinFrames      = 144
	logoMorphFrames     = 4 * logoSpinFrames
	logoPerspective     = 4.4
	logoRingSpacing     = logoStep * 4
	logoRingWidth       = logoStep * 0.7
	logoTentRef         = 0.10
	logoCanvasScaleX    = 0.2575 // 206/800: reference surface half-width ratio
	logoCanvasScaleY    = 0.3745 // 206/550: reference surface half-height ratio
	logoBrailleRuneBase = 0x2800
	logoMaxGridCols     = 512
	logoMaxGridRows     = 256
)

// logoDotBits maps a dot position (dx%2) + (dy%4)*2 inside a braille cell to
// its bitmask. Standard braille dot order: 1,4 / 2,5 / 3,6 / 7,8.
var logoDotBits = [8]uint8{1, 8, 2, 16, 4, 32, 64, 128}

type logoStroke struct {
	x1, y1, x2, y2, t float64
}

// logoRun is one filled horizontal span of a pixel-font glyph row.
type logoRun struct {
	r, c0, c1 int
}

// Faceted "justice cross" geometry in the 486x486 viewBox space: one closed
// outline and three inner crease lines. Shared by the stroke (surface) and
// edge (wireframe) builders.
var (
	logoJusticeOutline = [][2]float64{
		{225, 71}, {262, 71}, {272, 113}, {336, 113}, {355, 144}, {337, 208},
		{293, 208}, {326, 337}, {307, 407}, {180, 407}, {161, 337}, {194, 208},
		{149, 208}, {132, 144}, {150, 113}, {214, 113},
	}
	logoJusticeFacets = [][4]float64{
		{132, 144, 212, 144},
		{278, 144, 355, 144},
		{161, 337, 326, 337},
	}
)

// logoShapeStrokes returns the stroke list for a named shape, or nil.
func logoShapeStrokes(name string) []logoStroke {
	switch name {
	case "cross":
		return []logoStroke{
			{0.5, 0.06, 0.5, 0.94, 0.20},
			{0.16, 0.34, 0.84, 0.34, 0.20},
		}
	case "justice":
		// Faceted "justice cross" wireframe: 16-vertex outline loop plus
		// three inner facet lines, in the 486x486 viewBox coordinate space.
		var out []logoStroke
		for i, p := range logoJusticeOutline {
			q := logoJusticeOutline[(i+1)%len(logoJusticeOutline)]
			out = append(out, logoStroke{p[0], p[1], q[0], q[1], 3})
		}
		for _, f := range logoJusticeFacets {
			out = append(out, logoStroke{f[0], f[1], f[2], f[3], 3})
		}
		return out
	case "sysc":
		// 5x7 pixel-font glyphs as {row, colStart, colEnd} runs, one unit per pixel.
		sGlyph := []logoRun{
			{0, 1, 4}, {1, 0, 0}, {2, 0, 0}, {3, 1, 3}, {4, 4, 4}, {5, 4, 4}, {6, 0, 3},
		}
		yGlyph := []logoRun{
			{0, 0, 0}, {0, 4, 4}, {1, 0, 0}, {1, 4, 4}, {2, 1, 1}, {2, 3, 3},
			{3, 2, 2}, {4, 2, 2}, {5, 2, 2}, {6, 2, 2},
		}
		cGlyph := []logoRun{
			{0, 1, 4}, {1, 0, 0}, {2, 0, 0}, {3, 0, 0}, {4, 0, 0}, {5, 0, 0}, {6, 1, 4},
		}
		var out []logoStroke
		for i, g := range [][]logoRun{sGlyph, yGlyph, sGlyph, cGlyph} {
			dx := float64(i) * 6.0
			for _, r := range g {
				out = append(out, logoStroke{
					dx + float64(r.c0), float64(r.r),
					dx + float64(r.c1) + 1, float64(r.r), 1,
				})
			}
		}
		return out
	default:
		return nil
	}
}

// logoShapeName resolves registry effect ids to baked field names.
func logoShapeName(shape string) (a, b string) {
	switch shape {
	case "cross":
		return "cross", ""
	case "justice":
		return "justice", ""
	case "sysc", "":
		return "sysc", ""
	case "sysc-cross", "cross-sysc", "morph":
		a, b = "sysc", "cross"
		if shape == "cross-sysc" {
			a, b = b, a
		}
		return a, b
	}
	return "", ""
}

type logoSeg struct {
	ax, ay, bx, by float64
}

// logoSegments expands strokes into oriented rectangle loops and normalizes
// the union bounding box to [-1,1] on its longest axis.
func logoSegments(strokes []logoStroke) []logoSeg {
	var segs []logoSeg
	minx, miny := math.Inf(1), math.Inf(1)
	maxx, maxy := math.Inf(-1), math.Inf(-1)
	for _, s := range strokes {
		dx, dy := s.x2-s.x1, s.y2-s.y1
		length := math.Hypot(dx, dy)
		if length == 0 {
			continue
		}
		nx, ny := -dy/length, dx/length
		h := s.t / 2
		c1x, c1y := s.x1+nx*h, s.y1+ny*h
		c2x, c2y := s.x2+nx*h, s.y2+ny*h
		c3x, c3y := s.x2-nx*h, s.y2-ny*h
		c4x, c4y := s.x1-nx*h, s.y1-ny*h
		corners := [4][2]float64{{c1x, c1y}, {c2x, c2y}, {c3x, c3y}, {c4x, c4y}}
		for _, c := range corners {
			minx = math.Min(minx, c[0])
			maxx = math.Max(maxx, c[0])
			miny = math.Min(miny, c[1])
			maxy = math.Max(maxy, c[1])
		}
		for k := 0; k < 4; k++ {
			a, b := corners[k], corners[(k+1)%4]
			segs = append(segs, logoSeg{a[0], a[1], b[0], b[1]})
		}
	}
	scale := math.Max(maxx-minx, maxy-miny) / 2
	if scale <= 0 {
		return nil
	}
	cx, cy := (minx+maxx)/2, (miny+maxy)/2
	for i := range segs {
		s := &segs[i]
		s.ax, s.bx = (s.ax-cx)/scale, (s.bx-cx)/scale
		s.ay, s.by = (s.ay-cy)/scale, (s.by-cy)/scale
	}
	return segs
}

// logoBake computes the signed distance field of a shape in normalized units.
func logoBake(strokes []logoStroke) []float64 {
	segs := logoSegments(strokes)
	if segs == nil {
		return nil
	}
	inside := make([]bool, logoGrid*logoGrid)
	crossings := make([]struct {
		x   float64
		dir int
	}, 0, len(segs))
	for row := 0; row < logoGrid; row++ {
		y := float64(row)*logoStep - logoExtent
		crossings = crossings[:0]
		for _, s := range segs {
			if (s.ay <= y && s.by > y) || (s.by <= y && s.ay > y) {
				x := s.ax + (y-s.ay)/(s.by-s.ay)*(s.bx-s.ax)
				dir := 1
				if s.by <= s.ay {
					dir = -1
				}
				crossings = append(crossings, struct {
					x   float64
					dir int
				}{x, dir})
			}
		}
		sortLogoCrossings(crossings)
		cursor, winding := 0, 0
		base := row * logoGrid
		for column := 0; column < logoGrid; column++ {
			x := float64(column)*logoStep - logoExtent
			for cursor < len(crossings) && crossings[cursor].x <= x {
				winding += crossings[cursor].dir
				cursor++
			}
			inside[base+column] = winding != 0
		}
	}

	dist := make([]float64, logoGrid*logoGrid)
	for i := range dist {
		dist[i] = logoGrid
	}
	for row := 1; row < logoGrid-1; row++ {
		for column := 1; column < logoGrid-1; column++ {
			i := row*logoGrid + column
			if inside[i] != inside[i-1] || inside[i] != inside[i+1] ||
				inside[i] != inside[i-logoGrid] || inside[i] != inside[i+logoGrid] {
				dist[i] = 0.5
			}
		}
	}
	sqrt2 := math.Sqrt2
	for reverse := 0; reverse < 2; reverse++ {
		for row := 1; row < logoGrid-1; row++ {
			for column := 1; column < logoGrid-1; column++ {
				i := row*logoGrid + column
				if reverse == 1 {
					i = (logoGrid-1-row)*logoGrid + logoGrid - 1 - column
				}
				var neighbors [4]int
				if reverse == 1 {
					neighbors = [4]int{i + 1, i + logoGrid, i + logoGrid - 1, i + logoGrid + 1}
				} else {
					neighbors = [4]int{i - 1, i - logoGrid, i - logoGrid - 1, i - logoGrid + 1}
				}
				best := dist[i]
				for k, j := range neighbors {
					w := 1.0
					if k >= 2 {
						w = sqrt2
					}
					if d := dist[j] + w; d < best {
						best = d
					}
				}
				dist[i] = best
			}
		}
	}
	for i := range dist {
		sign := -1.0
		if inside[i] {
			sign = 1
		}
		dist[i] = dist[i] * logoStep * sign
	}
	return dist
}

func sortLogoCrossings(c []struct {
	x   float64
	dir int
}) {
	for i := 1; i < len(c); i++ {
		for j := i; j > 0 && c[j].x < c[j-1].x; j-- {
			c[j], c[j-1] = c[j-1], c[j]
		}
	}
}

var (
	logoSDFMu     sync.Mutex
	logoSDFCache  = map[string][]float64{}
	logoSDFShared = func(name string) []float64 { return logoSDF(name) }
)

// logoSDF returns the shared baked field for a shape (nil if unknown).
func logoSDF(name string) []float64 {
	logoSDFMu.Lock()
	defer logoSDFMu.Unlock()
	if f, ok := logoSDFCache[name]; ok {
		return f
	}
	strokes := logoShapeStrokes(name)
	if strokes == nil {
		logoSDFCache[name] = nil
		return nil
	}
	f := logoBake(strokes)
	logoSDFCache[name] = f
	return f
}

func logoGridSize() int        { return logoGrid }
func logoStepLength() float64  { return logoStep }
func logoExtentValue() float64 { return logoExtent }

// logoEdge is one 3D wire segment with per-endpoint object normals.
type logoEdge struct {
	ax, ay, az    float64
	bx, by, bz    float64
	nax, nay, naz float64
	nbx, nby, nbz float64
}

// logoLoop is a polyline in raw shape space. Closed loops are extruded with
// front and back copies joined by vertex struts; struts use radial normals.
type logoLoop struct {
	pts    [][2]float64
	closed bool
	struts bool
}

// logoShapeLoops returns the wireframe loop set for a shape.
func logoShapeLoops(name string) []logoLoop {
	switch name {
	case "justice":
		loops := []logoLoop{{pts: logoJusticeOutline, closed: true, struts: true}}
		for _, f := range logoJusticeFacets {
			loops = append(loops, logoLoop{
				pts: [][2]float64{{f[0], f[1]}, {f[2], f[3]}},
			})
		}
		return loops
	default:
		strokes := logoShapeStrokes(name)
		if strokes == nil {
			return nil
		}
		var loops []logoLoop
		for _, s := range strokes {
			dx, dy := s.x2-s.x1, s.y2-s.y1
			length := math.Hypot(dx, dy)
			if length == 0 {
				continue
			}
			nx, ny := -dy/length, dx/length
			h := s.t / 2
			loops = append(loops, logoLoop{
				pts: [][2]float64{
					{s.x1 + nx*h, s.y1 + ny*h}, {s.x2 + nx*h, s.y2 + ny*h},
					{s.x2 - nx*h, s.y2 - ny*h}, {s.x1 - nx*h, s.y1 - ny*h},
				},
				closed: true,
				struts: true,
			})
		}
		return loops
	}
}

// logoBuildEdges normalizes loops to the shared [-1,1] space and extrudes
// them into the front/back/strut edge set of a solid wireframe plate.
func logoBuildEdges(name string) []logoEdge {
	loops := logoShapeLoops(name)
	if len(loops) == 0 {
		return nil
	}
	minx, miny := math.Inf(1), math.Inf(1)
	maxx, maxy := math.Inf(-1), math.Inf(-1)
	for _, lp := range loops {
		for _, p := range lp.pts {
			minx = math.Min(minx, p[0])
			maxx = math.Max(maxx, p[0])
			miny = math.Min(miny, p[1])
			maxy = math.Max(maxy, p[1])
		}
	}
	scale := math.Max(maxx-minx, maxy-miny) / 2
	if scale <= 0 {
		return nil
	}
	cx, cy := (minx+maxx)/2, (miny+maxy)/2
	var out []logoEdge
	for _, lp := range loops {
		norm := make([][2]float64, len(lp.pts))
		for i, p := range lp.pts {
			norm[i] = [2]float64{(p[0] - cx) / scale, (p[1] - cy) / scale}
		}
		segments := len(norm) - 1
		if lp.closed {
			segments = len(norm)
		}
		for i := 0; i < segments; i++ {
			a := norm[i]
			b := norm[(i+1)%len(norm)]
			out = append(out,
				logoEdge{a[0], a[1], logoHalfDepth, b[0], b[1], logoHalfDepth, 0, 0, 1, 0, 0, 1},
				logoEdge{a[0], a[1], -logoHalfDepth, b[0], b[1], -logoHalfDepth, 0, 0, -1, 0, 0, -1},
			)
		}
		if lp.struts {
			for _, p := range norm {
				ox, oy := p[0], p[1]
				m := math.Hypot(ox, oy)
				if m == 0 {
					continue
				}
				ox, oy = ox/m, oy/m
				out = append(out, logoEdge{
					p[0], p[1], logoHalfDepth, p[0], p[1], -logoHalfDepth,
					ox, oy, 0, ox, oy, 0,
				})
			}
		}
	}
	return out
}

var (
	logoEdgeMu    sync.Mutex
	logoEdgeCache = map[string][]logoEdge{}
)

// logoEdges returns the cached edge set for a shape (nil if unknown).
func logoEdges(name string) []logoEdge {
	logoEdgeMu.Lock()
	defer logoEdgeMu.Unlock()
	if e, ok := logoEdgeCache[name]; ok {
		return e
	}
	e := logoBuildEdges(name)
	logoEdgeCache[name] = e
	return e
}

// Style modes for LogoSpinConfig.Style.
const (
	logoModeEdge = iota
	logoModeRing
	logoModeSolid
)

// LogoSpinConfig configures a LogoSpinAnimation.
type LogoSpinConfig struct {
	Width   int
	Height  int
	Shape   string   // "sysc" | "cross" | "justice" | "sysc-cross" | "cross-sysc"
	Palette []string // GetLogoPalette(theme)
	Theme   string
	Style   string // "edge" (default: crisp 3D wireframe), "ring" (contour bands; old "wire"), "solid" (beveled plate)
}

// LogoSpinAnimation renders a vector shape as a spinning 3D braille plate.
type LogoSpinAnimation struct {
	width, height int
	shapeA        string
	shapeB        string
	morph         bool
	mode          int
	palette       []string
	theme         string

	frame   int
	builder strings.Builder

	shape []float64
	depth []float64
	crgb  []uint32
	dots  []uint8
	count []uint8

	litShadow, litKey, litFill, litRim, litHigh, litBg [3]float64
}

// NewLogoSpinEffect creates a spinning logo animation.
func NewLogoSpinEffect(c LogoSpinConfig) *LogoSpinAnimation {
	if c.Width < 4 {
		c.Width = 4
	}
	if c.Height < 4 {
		c.Height = 4
	}
	if c.Width > logoMaxGridCols {
		c.Width = logoMaxGridCols
	}
	if c.Height > logoMaxGridRows {
		c.Height = logoMaxGridRows
	}
	a, b := logoShapeName(c.Shape)
	if a == "" {
		a, b = "sysc", ""
	}
	palette := c.Palette
	if len(palette) < 7 {
		palette = GetLogoPalette(c.Theme)
	}
	mode := logoModeEdge
	switch c.Style {
	case "ring", "wire":
		mode = logoModeRing
	case "solid":
		mode = logoModeSolid
	}
	if b != "" && mode == logoModeEdge {
		// Morphing blends two distance fields, so it needs a surface mode.
		mode = logoModeRing
	}
	l := &LogoSpinAnimation{
		width:   c.Width,
		height:  c.Height,
		shapeA:  a,
		shapeB:  b,
		morph:   b != "",
		mode:    mode,
		palette: palette,
		theme:   c.Theme,
	}
	l.initBuffers()
	l.initColors()
	return l
}

func (l *LogoSpinAnimation) initColors() {
	conv := func(i int) [3]float64 {
		r, g, b := hexToRGB(l.palette[i])
		return [3]float64{float64(r), float64(g), float64(b)}
	}
	l.litBg = conv(6)
	l.litShadow = conv(1)
	l.litFill = conv(2)
	l.litKey = conv(3)
	l.litHigh = conv(4)
	l.litRim = conv(5)
}

func (l *LogoSpinAnimation) initBuffers() {
	samples := l.width * 2 * l.height * 4
	if cap(l.depth) < samples {
		l.depth = make([]float64, samples)
		l.crgb = make([]uint32, samples)
	}
	l.depth = l.depth[:samples]
	l.crgb = l.crgb[:samples]
	cells := l.width * l.height
	if cap(l.dots) < cells {
		l.dots = make([]uint8, cells)
		l.count = make([]uint8, cells)
	}
	l.dots = l.dots[:cells]
	l.count = l.count[:cells]
	if cap(l.shape) < logoGrid*logoGrid {
		l.shape = make([]float64, logoGrid*logoGrid)
	}
	l.shape = l.shape[:logoGrid*logoGrid]
}

// Update advances the animation by one frame.
func (l *LogoSpinAnimation) Update() { l.frame++ }

// Reset restarts the animation from the beginning.
func (l *LogoSpinAnimation) Reset() { l.frame = 0 }

// Resize updates the render grid.
func (l *LogoSpinAnimation) Resize(width, height int) {
	if width < 4 {
		width = 4
	}
	if height < 4 {
		height = 4
	}
	if width > logoMaxGridCols {
		width = logoMaxGridCols
	}
	if height > logoMaxGridRows {
		height = logoMaxGridRows
	}
	l.width, l.height = width, height
	l.initBuffers()
}

func logoSmooth(v float64) float64 {
	t := v
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	return t * t * t * (t*(t*6-15) + 10)
}

func logoMorphPeriodFrames() int { return logoMorphFrames }

// logoSpinRotation reproduces the reference bloom spin driver: phase loops
// every logoSpinFrames; each half-loop eases one full turn over 82% of its
// time then holds facing forward. X tilt oscillates once per turn, Z sways
// once per loop.
func logoSpinRotation(frame int) (rotation, tiltX, tiltZ float64) {
	phase := float64(frame) / float64(logoSpinFrames)
	phase -= math.Floor(phase)
	second := 0.0
	if phase >= 0.5 {
		second = 1.0
	}
	spin := logoSmooth((phase*2 - second) / 0.82)
	rotation = (second + spin) * 2 * math.Pi
	tiltX = math.Sin(rotation) * logoWobble
	tiltZ = math.Sin(phase*2*math.Pi) * 0.045
	return rotation, tiltX, tiltZ
}

// Render draws the current frame.
func (l *LogoSpinAnimation) Render() string {
	if l.mode == logoModeEdge {
		return l.renderEdges()
	}
	a := logoSDFShared(l.shapeA)
	if a == nil {
		return ""
	}
	var b []float64
	if l.morph {
		b = logoSDFShared(l.shapeB)
	}

	rotation, tiltX, tiltZ := logoSpinRotation(l.frame)
	blend := 0.0
	if l.morph {
		u := float64(l.frame) / float64(logoMorphFrames)
		u -= math.Floor(u)
		tri := 1 - math.Abs(2*u-1)
		blend = logoSmooth(tri)
	}
	bevel := 0.075 + blend*0.012

	for i := range l.shape {
		if l.morph {
			l.shape[i] = a[i]*(1-blend) + b[i]*blend
		} else {
			l.shape[i] = a[i]
		}
	}

	dotColumns := l.width * 2
	dotRows := l.height * 4
	for i := range l.depth {
		l.depth[i] = math.Inf(-1)
	}
	for i := range l.dots {
		l.dots[i] = 0
		l.count[i] = 0
	}

	// The plate swings ±6.4° about X once per turn and breathes on Z once per
	// loop, exactly like the reference bloom driver; the turn itself eases.
	sx, cx := math.Sincos(tiltX)
	sy, cy := math.Sincos(rotation)
	sz, cz := math.Sincos(tiltZ)

	coefX := logoCanvasScaleX * float64(dotColumns)
	coefY := logoCanvasScaleY * float64(dotRows)

	project := func(x, y, z, nx, ny, nz float64) {
		uu := x*2.2 + y*0.7
		vv := y*2.8 - x*0.4
		z += math.Sin(uu)*0.075 + math.Sin(vv)*0.04
		nx -= nz * (math.Cos(uu)*0.165 - math.Cos(vv)*0.016)
		ny -= nz * (math.Cos(uu)*0.0525 + math.Cos(vv)*0.112)
		length := math.Sqrt(nx*nx + ny*ny + nz*nz)
		if length == 0 {
			length = 1
		}
		nx, ny, nz = nx/length, ny/length, nz/length

		x1 := x*cy + z*sy
		z1 := -x*sy + z*cy
		y2 := y*cx - z1*sx
		z2 := y*sx + z1*cx
		x3 := x1*cz - y2*sz
		y3 := x1*sz + y2*cz

		persp := logoPerspective / (logoPerspective - z2)
		column := math.Floor(float64(dotColumns)/2 + x3*coefX*persp)
		row := math.Floor(float64(dotRows)/2 + y3*coefY*persp)
		if column < 0 || column >= float64(dotColumns) || row < 0 || row >= float64(dotRows) {
			return
		}
		i := int(row)*dotColumns + int(column)
		if z2 <= l.depth[i] {
			return
		}
		nx1 := nx*cy + nz*sy
		nz1 := -nx*sy + nz*cy
		ny2 := ny*cx - nz1*sx
		nz2 := ny*sx + nz1*cx
		l.depth[i] = z2
		l.crgb[i] = logoShade(
			[3]float64{nx1*cz - ny2*sz, nx1*sz + ny2*cz, nz2},
			z2,
			math.Sin(x*23+y*19+logoGrainPhase)*0.01,
			l,
		)
	}

	for row := 1; row < logoGrid-1; row++ {
		base := row * logoGrid
		y := float64(row)*logoStep - logoExtent
		for column := 1; column < logoGrid-1; column++ {
			i := base + column
			d := l.shape[i]
			if d < -logoStep {
				continue
			}
			x := float64(column)*logoStep - logoExtent
			dx := l.shape[i+1] - l.shape[i-1]
			dy := l.shape[i+logoGrid] - l.shape[i-logoGrid]
			length := math.Hypot(dx, dy)
			if length == 0 {
				length = 1
			}
			gx, gy := dx/length, dy/length

			if d > 0 {
				if l.mode == logoModeRing {
					frac := math.Mod(d, logoRingSpacing)
					if frac < logoRingWidth || frac > logoRingSpacing-logoRingWidth {
						h := clampLogo(d/logoTentRef, 0, 1)
						k := 0.0
						if h < 1 {
							k = logoHalfDepth / logoTentRef
						}
						nl := math.Sqrt(1 + k*k)
						z := logoHalfDepth * h
						project(x, y, z, -gx*k/nl, -gy*k/nl, 1/nl)
						project(x, y, -z, -gx*k/nl, -gy*k/nl, -1/nl)
					}
				} else {
					edge := clampLogo(1-d/bevel, 0, 1)
					nz := math.Sqrt(1 - edge*edge)
					z := logoHalfDepth - bevel + bevel*nz
					project(x, y, z, -gx*edge, -gy*edge, nz)
					project(x, y, -z, -gx*edge, -gy*edge, -nz)
				}
			}
			if math.Abs(d) < logoStep*0.8 {
				sideDepth := logoHalfDepth - bevel
				layers := int(math.Ceil(sideDepth * 2 / logoStep))
				for layer := 0; layer <= layers; layer++ {
					project(
						x-gx*d, y-gy*d,
						-sideDepth+float64(layer)/float64(layers)*sideDepth*2,
						-gx, -gy, 0,
					)
				}
			}
		}
	}

	return l.compose()
}

// compose turns the shaded dot buffers into the ANSI braille frame.
func (l *LogoSpinAnimation) compose() string {
	dotColumns := l.width * 2
	l.builder.Reset()
	for row := 0; row < l.height; row++ {
		for column := 0; column < l.width; column++ {
			var dots uint8
			var count uint8
			for point, bit := range logoDotBits {
				i := (row*4+point/2)*dotColumns + column*2 + point%2
				if !math.IsInf(l.depth[i], -1) {
					dots |= bit
					count++
				}
			}
			if dots == 0 {
				l.builder.WriteByte(' ')
				continue
			}
			// Average the lit dot colours.
			var rs, gs, bs uint32
			for point := 0; point < 8; point++ {
				if dots&logoDotBits[point] == 0 {
					continue
				}
				i := (row*4+point/2)*dotColumns + column*2 + point%2
				c := l.crgb[i]
				rs += uint32(c>>16) & 0xff
				gs += uint32(c>>8) & 0xff
				bs += uint32(c) & 0xff
			}
			n := uint32(count)
			r := (rs + n/2) / n
			g := (gs + n/2) / n
			bv := (bs + n/2) / n
			fmt.Fprintf(&l.builder, "\033[38;2;%d;%d;%dm%c\033[0m", r, g, bv, rune(logoBrailleRuneBase)+rune(dots))
		}
		if row < l.height-1 {
			l.builder.WriteByte('\n')
		}
	}
	return l.builder.String()
}

// renderEdges draws the crisp 3D wireframe mode: rigid front and back
// outline loops joined by vertex struts, projected through the same camera
// as the surface modes and z-buffered per braille dot. Because every sample
// comes from a real 3D segment, the spin reads as one solid object turning
// instead of contour bands crawling over a surface.
func (l *LogoSpinAnimation) renderEdges() string {
	dotColumns := l.width * 2
	dotRows := l.height * 4
	for i := range l.depth {
		l.depth[i] = math.Inf(-1)
	}
	edges := logoEdges(l.shapeA)
	if len(edges) == 0 {
		return ""
	}

	rotation, tiltX, tiltZ := logoSpinRotation(l.frame)
	sx, cx := math.Sincos(tiltX)
	sy, cy := math.Sincos(rotation)
	sz, cz := math.Sincos(tiltZ)

	coefX := logoCanvasScaleX * float64(dotColumns)
	coefY := logoCanvasScaleY * float64(dotRows)
	halfColumns := float64(dotColumns) / 2
	halfRows := float64(dotRows) / 2

	rot := func(x, y, z float64) (float64, float64, float64) {
		x1 := x*cy + z*sy
		z1 := -x*sy + z*cy
		y2 := y*cx - z1*sx
		z2 := y*sx + z1*cx
		return x1*cz - y2*sz, x1*sz + y2*cz, z2
	}

	for _, e := range edges {
		aX, aY, aZ := rot(e.ax, e.ay, e.az)
		bX, bY, bZ := rot(e.bx, e.by, e.bz)
		pa := logoPerspective / (logoPerspective - aZ)
		pb := logoPerspective / (logoPerspective - bZ)
		x0 := halfColumns + aX*coefX*pa
		y0 := halfRows + aY*coefY*pa
		x1 := halfColumns + bX*coefX*pb
		y1 := halfRows + bY*coefY*pb
		nax, nay, naz := rot(e.nax, e.nay, e.naz)
		nbx, nby, nbz := rot(e.nbx, e.nby, e.nbz)

		steps := int(math.Ceil(math.Hypot(x1-x0, y1-y0) / 0.55))
		if steps < 1 {
			steps = 1
		}
		for k := 0; k <= steps; k++ {
			t := float64(k) / float64(steps)
			col := int(math.Floor(x0 + (x1-x0)*t))
			row := int(math.Floor(y0 + (y1-y0)*t))
			if col < 0 || col >= dotColumns || row < 0 || row >= dotRows {
				continue
			}
			i := row*dotColumns + col
			z2 := aZ + (bZ-aZ)*t
			if z2 <= l.depth[i] {
				continue
			}
			nx := nax + (nbx-nax)*t
			ny := nay + (nby-nay)*t
			nz := naz + (nbz-naz)*t
			n := math.Sqrt(nx*nx + ny*ny + nz*nz)
			if n == 0 {
				continue
			}
			l.depth[i] = z2
			l.crgb[i] = logoShade([3]float64{nx / n, ny / n, nz / n}, z2, 0, l)
		}
	}
	return l.compose()
}

func logoShade(n [3]float64, depth, grain float64, l *LogoSpinAnimation) uint32 {
	nx, ny, nz := n[0], n[1], n[2]
	diffuse := clampLogo(nx*-0.410+ny*-0.564+nz*0.718+grain, 0, 1)
	bounce := clampLogo(nx*0.55+ny*0.20-nz*0.35, 0, 1) * (1 - diffuse) * 0.38
	edge := math.Pow(1-math.Abs(nz), 2.4) * clampLogo(nx*0.85-ny*0.38, 0, 1) * 0.82
	gloss := math.Pow(clampLogo(nx*-0.220+ny*-0.302+nz*0.928, 0, 1), 18) * 0.96
	base := (1 - bounce) * (1 - edge) * (1 - gloss)
	gain := clampLogo(0.90+depth*0.18, 0.68, 1)

	var out uint32
	for i := 0; i < 3; i++ {
		value := l.litShadow[i]*(1-diffuse)*base +
			l.litKey[i]*diffuse*base +
			l.litFill[i]*bounce*(1-edge)*(1-gloss) +
			l.litRim[i]*edge*(1-gloss) +
			l.litHigh[i]*gloss
		channel := clampLogo(l.litBg[i]+(value-l.litBg[i])*gain, 0, 255)
		out = out<<8 | uint32(channel+0.5)
	}
	return out
}

func clampLogo(x, lo, hi float64) float64 {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}
