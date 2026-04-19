package animations

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
)

type CracktroPhase int

const (
	CracktroBoot CracktroPhase = iota
	CracktroRun
)

type Star struct {
	x     float64
	y     int
	speed float64
	layer int
}

type RasterBar struct {
	phase    float64
	speed    float64
	colorIdx int
}

type vec3 struct {
	x, y, z float64
}

type wireEdge struct {
	a, b int
}

type wirePixel struct {
	ch    rune
	depth float64
}

type trailPoint struct {
	x, y int
	age  int
	ch   rune
}

type CracktroAnimation struct {
	width, height int
	palette       []string
	theme         string

	stars      []Star
	bars       []RasterBar
	phase      CracktroPhase
	frameCount int
	bootFrames int

	rotY      float64
	rotX      float64
	innerRotY float64
	innerRotX float64

	trails    []trailPoint
	maxTrails int

	cubeVerts []vec3
	cubeEdges []wireEdge

	builder strings.Builder
}

func NewCracktroEffect(width, height int, palette []string, theme string) *CracktroAnimation {
	c := &CracktroAnimation{
		width:      width,
		height:     height,
		palette:    palette,
		theme:      theme,
		phase:      CracktroBoot,
		bootFrames: 45,
		maxTrails:  800,
	}

	c.cubeVerts = []vec3{
		{-1, -1, -1}, {1, -1, -1}, {1, 1, -1}, {-1, 1, -1},
		{-1, -1, 1}, {1, -1, 1}, {1, 1, 1}, {-1, 1, 1},
	}
	c.cubeEdges = []wireEdge{
		{0, 1}, {1, 2}, {2, 3}, {3, 0},
		{4, 5}, {5, 6}, {6, 7}, {7, 4},
		{0, 4}, {1, 5}, {2, 6}, {3, 7},
	}

	c.initCracktroStars()
	c.initCracktroBars()
	return c
}

func (c *CracktroAnimation) initCracktroStars() {
	c.stars = nil
	totalStars := (c.width * c.height) / 25
	if totalStars < 40 {
		totalStars = 40
	}
	if totalStars > 200 {
		totalStars = 200
	}
	for i := 0; i < totalStars; i++ {
		layer := rand.Intn(3)
		var speed float64
		switch layer {
		case 0:
			speed = 0.3 + rand.Float64()*0.2
		case 1:
			speed = 0.7 + rand.Float64()*0.3
		case 2:
			speed = 1.2 + rand.Float64()*0.5
		}
		c.stars = append(c.stars, Star{
			x:     rand.Float64() * float64(c.width),
			y:     rand.Intn(c.height),
			speed: speed,
			layer: layer,
		})
	}
}

func (c *CracktroAnimation) initCracktroBars() {
	c.bars = []RasterBar{
		{phase: 0, speed: 0.06, colorIdx: 3},
		{phase: math.Pi * 0.5, speed: 0.05, colorIdx: 4},
		{phase: math.Pi, speed: 0.07, colorIdx: 3},
		{phase: math.Pi * 1.5, speed: 0.055, colorIdx: 4},
	}
}

func (c *CracktroAnimation) UpdatePalette(palette []string) {
	c.palette = palette
}

func (c *CracktroAnimation) Update() {
	c.frameCount++

	if c.phase == CracktroBoot && c.frameCount >= c.bootFrames {
		c.phase = CracktroRun
	}

	for i := range c.stars {
		c.stars[i].x -= c.stars[i].speed
		if c.stars[i].x < 0 {
			c.stars[i].x = float64(c.width) + rand.Float64()*10
			c.stars[i].y = rand.Intn(c.height)
		}
	}

	for i := range c.bars {
		c.bars[i].phase += c.bars[i].speed
	}

	c.rotY += 0.03
	c.rotX += 0.012
	c.innerRotY -= 0.045
	c.innerRotX += 0.02

	kept := c.trails[:0]
	for i := range c.trails {
		c.trails[i].age++
		if c.trails[i].age < 4 {
			kept = append(kept, c.trails[i])
		}
	}
	c.trails = kept
}

func (c *CracktroAnimation) Resize(width, height int) {
	c.width = width
	c.height = height
	c.initCracktroStars()
	c.trails = nil
}

func (c *CracktroAnimation) Reset() {
	c.frameCount = 0
	c.phase = CracktroBoot
	c.rotY = 0
	c.rotX = 0
	c.innerRotY = 0
	c.innerRotX = 0
	c.trails = nil
	c.initCracktroStars()
	c.initCracktroBars()
}

func cracktroRotate(v vec3, rotY, rotX float64) vec3 {
	cosY := math.Cos(rotY)
	sinY := math.Sin(rotY)
	x1 := v.x*cosY + v.z*sinY
	z1 := -v.x*sinY + v.z*cosY

	cosX := math.Cos(rotX)
	sinX := math.Sin(rotX)
	y1 := v.y*cosX - z1*sinX
	z2 := v.y*sinX + z1*cosX

	return vec3{x1, y1, z2}
}

func (c *CracktroAnimation) project(v vec3, scale float64) (int, int, float64) {
	camDist := 4.0
	z := v.z + camDist
	if z < 0.1 {
		z = 0.1
	}
	factor := camDist / z
	sx := float64(c.width)/2.0 + v.x*scale*factor
	sy := float64(c.height)/2.0 + v.y*scale*factor*0.5
	return int(math.Round(sx)), int(math.Round(sy)), v.z
}

func (c *CracktroAnimation) Render() string {
	c.builder.Reset()

	bootFade := 1.0
	if c.phase == CracktroBoot {
		bootFade = float64(c.frameCount) / float64(c.bootFrames)
		if bootFade > 1.0 {
			bootFade = 1.0
		}
	}

	// Stars
	starMap := make(map[[2]int]int, len(c.stars))
	for _, s := range c.stars {
		x := int(s.x)
		if x >= 0 && x < c.width && s.y >= 0 && s.y < c.height {
			if c.phase == CracktroBoot && s.speed/1.7 > bootFade*0.8 {
				continue
			}
			pos := [2]int{x, s.y}
			if existing, ok := starMap[pos]; !ok || s.layer > existing {
				starMap[pos] = s.layer
			}
		}
	}

	// Raster bars
	type barHit struct {
		colorIdx  int
		intensity float64
	}
	barMap := make(map[int]barHit)
	for _, bar := range c.bars {
		centerY := float64(c.height)/2.0 + math.Sin(bar.phase)*float64(c.height)*0.38
		for dy := -2; dy <= 2; dy++ {
			y := int(centerY) + dy
			if y < 0 || y >= c.height {
				continue
			}
			dist := math.Abs(float64(dy)) / 2.5
			intensity := (1.0 - dist*0.4) * bootFade
			if intensity < 0.2 {
				intensity = 0.2
			}
			if existing, ok := barMap[y]; !ok || intensity > existing.intensity {
				barMap[y] = barHit{colorIdx: bar.colorIdx, intensity: intensity}
			}
		}
	}

	// Wireframe
	wireMap := make(map[[2]int]wirePixel)
	if bootFade > 0.5 {
		outerScale := math.Min(float64(c.width)*0.18, float64(c.height)*0.6)
		innerScale := outerScale * 0.5
		c.renderCube(wireMap, c.rotY, c.rotX, outerScale)
		c.renderCube(wireMap, c.innerRotY, c.innerRotX, innerScale)
	}

	// Build glow map around wireframe edges
	type glowHit struct {
		depth float64
		dist  int // 1 = adjacent, 2 = further
	}
	glowMap := make(map[[2]int]glowHit)
	for pos, wp := range wireMap {
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				if dx == 0 && dy == 0 {
					continue
				}
				gx, gy := pos[0]+dx, pos[1]+dy
				if gx < 0 || gx >= c.width || gy < 0 || gy >= c.height {
					continue
				}
				gpos := [2]int{gx, gy}
				if _, isWire := wireMap[gpos]; isWire {
					continue
				}
				dist := 1
				if dx != 0 && dy != 0 {
					dist = 2 // diagonal = further
				}
				if existing, ok := glowMap[gpos]; !ok || dist < existing.dist {
					glowMap[gpos] = glowHit{depth: wp.depth, dist: dist}
				}
			}
		}
	}

	// Store trails from current wireframe
	for pos, wp := range wireMap {
		if len(c.trails) < c.maxTrails {
			c.trails = append(c.trails, trailPoint{x: pos[0], y: pos[1], age: 0, ch: wp.ch})
		}
	}

	// Build trail map (exclude live wire and glow positions)
	type trailHit struct {
		age int
	}
	trailMap := make(map[[2]int]trailHit)
	for _, t := range c.trails {
		pos := [2]int{t.x, t.y}
		if _, isWire := wireMap[pos]; isWire {
			continue
		}
		if _, isGlow := glowMap[pos]; isGlow {
			continue
		}
		if existing, ok := trailMap[pos]; !ok || t.age < existing.age {
			trailMap[pos] = trailHit{age: t.age}
		}
	}

	// Compose frame
	for y := 0; y < c.height; y++ {
		for x := 0; x < c.width; x++ {
			pos := [2]int{x, y}

			if wp, ok := wireMap[pos]; ok {
				col := c.cracktroSlot(5)
				r, g, b := hexToRGB(col)
				depthFactor := 0.5 + 0.5*(1.0-((wp.depth+1.0)/2.0))
				pulse := 0.85 + 0.15*math.Sin(float64(c.frameCount)*0.12)
				brightness := depthFactor * pulse * bootFade
				r = cracktroClamp(int(float64(r)*brightness), 0, 255)
				g = cracktroClamp(int(float64(g)*brightness), 0, 255)
				b = cracktroClamp(int(float64(b)*brightness), 0, 255)
				fmt.Fprintf(&c.builder, "\033[38;2;%d;%d;%dm%c\033[0m", r, g, b, wp.ch)
				continue
			}

			if gh, ok := glowMap[pos]; ok {
				col := c.cracktroSlot(5)
				r, g, b := hexToRGB(col)
				depthFactor := 0.5 + 0.5*(1.0-((gh.depth+1.0)/2.0))
				var glowIntensity float64
				var ch rune
				if gh.dist == 1 {
					glowIntensity = 0.4
					ch = '\u2593' // ▓
				} else {
					glowIntensity = 0.2
					ch = '\u2592' // ▒
				}
				brightness := depthFactor * glowIntensity * bootFade
				r = cracktroClamp(int(float64(r)*brightness), 0, 255)
				g = cracktroClamp(int(float64(g)*brightness), 0, 255)
				b = cracktroClamp(int(float64(b)*brightness), 0, 255)
				fmt.Fprintf(&c.builder, "\033[38;2;%d;%d;%dm%c\033[0m", r, g, b, ch)
				continue
			}

			if th, ok := trailMap[pos]; ok {
				col := c.cracktroSlot(5)
				r, g, b := hexToRGB(col)
				var decay float64
				var ch rune
				switch th.age {
				case 1:
					decay = 0.45
					ch = '\u2593' // ▓
				case 2:
					decay = 0.2
					ch = '\u2592' // ▒
				default:
					decay = 0.08
					ch = '\u2591' // ░
				}
				r = cracktroClamp(int(float64(r)*decay), 0, 255)
				g = cracktroClamp(int(float64(g)*decay), 0, 255)
				b = cracktroClamp(int(float64(b)*decay), 0, 255)
				fmt.Fprintf(&c.builder, "\033[38;2;%d;%d;%dm%c\033[0m", r, g, b, ch)
				continue
			}

			if bh, ok := barMap[y]; ok {
				col := c.cracktroSlot(bh.colorIdx)
				r, g, b := hexToRGB(col)
				r = cracktroClamp(int(float64(r)*bh.intensity), 0, 255)
				g = cracktroClamp(int(float64(g)*bh.intensity), 0, 255)
				b = cracktroClamp(int(float64(b)*bh.intensity), 0, 255)
				var ch rune
				if bh.intensity > 0.8 {
					ch = '\u2588' // █
				} else if bh.intensity > 0.6 {
					ch = '\u2593' // ▓
				} else if bh.intensity > 0.4 {
					ch = '\u2592' // ▒
				} else {
					ch = '\u2591' // ░
				}
				fmt.Fprintf(&c.builder, "\033[38;2;%d;%d;%dm%c\033[0m", r, g, b, ch)
				continue
			}

			if layer, ok := starMap[pos]; ok {
				col := c.cracktroSlot(layer)
				r, g, b := hexToRGB(col)
				var ch rune
				switch layer {
				case 0:
					ch = '\u00b7' // ·
				case 1:
					ch = '\u2219' // ∙
				case 2:
					ch = '*'
				}
				fmt.Fprintf(&c.builder, "\033[38;2;%d;%d;%dm%c\033[0m", r, g, b, ch)
				continue
			}

			c.builder.WriteByte(' ')
		}
		if y < c.height-1 {
			c.builder.WriteByte('\n')
		}
	}

	return c.builder.String()
}

func (c *CracktroAnimation) renderCube(wireMap map[[2]int]wirePixel, rotY, rotX, scale float64) {
	type projVert struct {
		sx, sy int
		z      float64
	}
	var proj [8]projVert
	for i, v := range c.cubeVerts {
		rv := cracktroRotate(v, rotY, rotX)
		sx, sy, z := c.project(rv, scale)
		proj[i] = projVert{sx, sy, z}
	}

	// Vertex dots
	for _, p := range proj {
		if p.sx >= 0 && p.sx < c.width && p.sy >= 0 && p.sy < c.height {
			pos := [2]int{p.sx, p.sy}
			if existing, ok := wireMap[pos]; !ok || p.z < existing.depth {
				wireMap[pos] = wirePixel{ch: '\u25cf', depth: p.z} // ●
			}
		}
	}

	// Edges
	for _, e := range c.cubeEdges {
		c.drawEdge(wireMap, proj[e.a].sx, proj[e.a].sy, proj[e.a].z,
			proj[e.b].sx, proj[e.b].sy, proj[e.b].z)
	}
}

func (c *CracktroAnimation) drawEdge(wireMap map[[2]int]wirePixel, x0, y0 int, z0 float64, x1, y1 int, z1 float64) {
	dx := float64(x1 - x0)
	dy := float64(y1 - y0)
	steps := math.Max(math.Abs(dx), math.Abs(dy))
	if steps < 1 {
		return
	}

	xInc := dx / steps
	yInc := dy / steps
	zInc := (z1 - z0) / steps

	x := float64(x0)
	y := float64(y0)
	z := z0

	for i := 0; i <= int(steps); i++ {
		ix := int(math.Round(x))
		iy := int(math.Round(y))

		if ix >= 0 && ix < c.width && iy >= 0 && iy < c.height {
			pos := [2]int{ix, iy}
			if existing, ok := wireMap[pos]; !ok || z < existing.depth {
				wireMap[pos] = wirePixel{ch: '\u2588', depth: z} // █ solid block
			}
		}

		x += xInc
		y += yInc
		z += zInc
	}
}

func (c *CracktroAnimation) cracktroSlot(i int) string {
	if len(c.palette) == 0 {
		return "#ffffff"
	}
	if i >= len(c.palette) {
		return c.palette[len(c.palette)-1]
	}
	return c.palette[i]
}

func cracktroClamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
