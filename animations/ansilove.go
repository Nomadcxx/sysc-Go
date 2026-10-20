package animations

// Artwork data: "ansi love" by warpus (Legacy Krew), December 2019,
// https://www.ansilove.org/assets/examples/wa-ansilove.ans
// SHA-256: 5c5d75b205680a1098517ef36432ce36161415c2dd2f9552b9ede7cdc2330ce3.

import (
	"crypto/sha256"
	_ "embed"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/Nomadcxx/sysc-Go/animations/internal/ansiart"
)

type ANSILoveBackend uint8

const (
	ANSILoveAuto ANSILoveBackend = iota
	ANSILoveExact
	ANSILoveHalf
	ANSILoveQuad
	ANSILoveSextant
	ANSILoveOctant
	ANSILoveBrailleTwoColor
	ANSILoveBrailleFG
)

type ANSILoveColorDepth uint8

const (
	ANSILoveTrueColor ANSILoveColorDepth = iota
	ANSILove256
	ANSILove16
)

type ANSILoveFit uint8

const (
	ANSILoveFitPan ANSILoveFit = iota
	ANSILoveFitContain
	ANSILoveFitWidth
	ANSILoveFitExact
)

type ANSILoveState uint8

const (
	ANSILoveIdle ANSILoveState = iota
	ANSILoveLogin
)

type ANSILoveMotion uint8

const (
	ANSILoveMotionOff ANSILoveMotion = iota
	ANSILoveMotionSubtle
	ANSILoveMotionFull
)

type ANSILoveConfig struct {
	Width, Height  int
	Theme, Palette string
	Backend        ANSILoveBackend
	Depth          ANSILoveColorDepth
	Fit            ANSILoveFit
	Zoom           float64
	CellW, CellH   int
	Motion         ANSILoveMotion
	// Source accepts empty or "auto" for the verified embedded original.
	Source string
}

const ansiLoveCredit = "Drawn by Warpus, December 2019"

//go:embed artwork/ansi-love.ans
var ansiLoveANSI []byte

//go:embed artwork/ansi-love.raster.zlib
var ansiLoveRaster []byte
var ansiLoveSource struct {
	once   sync.Once
	grid   *ansiart.Grid
	raster *ansiart.Raster
	err    error
}

func loadANSILove() (*ansiart.Grid, *ansiart.Raster, error) {
	ansiLoveSource.once.Do(func() {
		if fmt.Sprintf("%x", sha256.Sum256(ansiLoveANSI)) != "5c5d75b205680a1098517ef36432ce36161415c2dd2f9552b9ede7cdc2330ce3" {
			ansiLoveSource.err = fmt.Errorf("ANSI Love: embedded source digest mismatch")
			return
		}
		body, s := ansiart.ParseSAUCE(ansiLoveANSI)
		if s == nil || s.TInfo1 != 80 || s.TInfo2 != 157 || s.TFlags != 0x12 {
			ansiLoveSource.err = fmt.Errorf("ANSI Love: invalid source metadata")
			return
		}
		ansiLoveSource.grid = ansiart.DecodeANSI(body, 80, false)
		ansiLoveSource.raster, ansiLoveSource.err = ansiart.UnpackRaster(ansiLoveRaster)
	})
	return ansiLoveSource.grid, ansiLoveSource.raster, ansiLoveSource.err
}

type ansiLoveCell struct {
	ch     rune
	fg, bg [3]uint8
	fi, bi uint8
}
type ansiLoveSheet struct {
	layout     ansiart.Layout
	weights    *ansiart.Weights
	cells      []ansiLoveCell
	flags      []uint8 // bit 0 heartbeat, bit 1 logo
	rowFlags   []uint8
	rows       []string
	lut        [112]ansiart.LinRGB
	background ansiLoveCell
}

// ANSILoveAnimation is owned by the host update goroutine. Prepared resize
// closures own immutable snapshots and may run on background goroutines.
type ANSILoveAnimation struct {
	cfg                ANSILoveConfig
	sheet              *ansiLoveSheet
	state              ANSILoveState
	elapsed            time.Duration
	pending            time.Duration
	beat, bpm, nextBPM float64
	camera, stageTime  float64
	stage              int
	cameraRow          int
	stepTime           float64
	generation         uint64
	colors             *[16][3]uint8
	cached             string
	key                [5]int
	scratch            []byte
}

type ANSILovePrepared struct {
	owner      *ANSILoveAnimation
	generation uint64
	cfg        ANSILoveConfig
	sheet      *ansiLoveSheet
}

func normalizeANSILove(c ANSILoveConfig) (ANSILoveConfig, error) {
	c.Width = max(1, min(512, c.Width))
	c.Height = max(1, min(256, c.Height))
	if c.Backend > ANSILoveBrailleFG || c.Depth > ANSILove16 || c.Fit > ANSILoveFitExact || c.Motion > ANSILoveMotionFull {
		return c, fmt.Errorf("ANSI Love: invalid backend, depth, fit or motion")
	}
	if c.Backend == ANSILoveAuto {
		c.Backend = ANSILoveQuad
	}
	if c.Palette == "" {
		c.Palette = "vga"
	}
	if c.Palette != "vga" && c.Palette != "theme" {
		return c, fmt.Errorf("ANSI Love: unknown palette %q", c.Palette)
	}
	if c.Zoom == 0 {
		c.Zoom = 2.5
	}
	if math.IsNaN(c.Zoom) || math.IsInf(c.Zoom, 0) || c.Zoom < 1 || c.Zoom > 3 {
		return c, fmt.Errorf("ANSI Love: zoom must be between 1 and 3")
	}
	if c.CellW == 0 && c.CellH == 0 {
		c.CellW, c.CellH = 8, 16
	}
	if c.CellW < 1 || c.CellW > 1024 || c.CellH < 1 || c.CellH > 1024 {
		return c, fmt.Errorf("ANSI Love: invalid cell size")
	}
	if c.Source != "" && c.Source != "auto" {
		return c, fmt.Errorf("ANSI Love: only the embedded source is available")
	}
	return c, nil
}

func NewANSILoveEffect(w, h int, theme string) *ANSILoveAnimation {
	a, err := NewANSILoveEffectConfig(ANSILoveConfig{Width: w, Height: h, Theme: theme, Motion: ANSILoveMotionSubtle})
	if err != nil {
		panic(err)
	} // Embedded assets are verified by the reproducible bake tests.
	return a
}
func NewANSILoveEffectConfig(cfg ANSILoveConfig) (*ANSILoveAnimation, error) {
	cfg, err := normalizeANSILove(cfg)
	if err != nil {
		return nil, err
	}
	if _, _, err = loadANSILove(); err != nil {
		return nil, err
	}
	a := &ANSILoveAnimation{cfg: cfg, bpm: 66, nextBPM: 66}
	a.Resize(cfg.Width, cfg.Height)
	return a, nil
}

func (a *ANSILoveAnimation) PrepareResize(w, h int) func() *ANSILovePrepared {
	return a.PrepareLayout(w, h, a.cfg.CellW, a.cfg.CellH)
}

// PrepareLayout snapshots a viewport and terminal cell aspect for an async bake.
func (a *ANSILoveAnimation) PrepareLayout(w, h, cellW, cellH int) func() *ANSILovePrepared {
	cfg := a.cfg
	cfg.Width = max(1, min(512, w))
	cfg.Height = max(1, min(256, h))
	if cellW >= 1 && cellW <= 1024 && cellH >= 1 && cellH <= 1024 {
		cfg.CellW, cfg.CellH = cellW, cellH
	}
	a.generation++
	generation, colors := a.generation, a.colors
	return func() *ANSILovePrepared {
		return &ANSILovePrepared{owner: a, generation: generation, cfg: cfg, sheet: bakeANSILove(cfg, colors)}
	}
}
func (a *ANSILoveAnimation) Apply(p *ANSILovePrepared) {
	if p == nil || p.owner != a || p.generation != a.generation {
		return
	}
	oldTravel := a.travel()
	oldDuration := a.stageDuration()
	a.cfg, a.sheet = p.cfg, p.sheet
	if oldDuration > 0 && a.stage%2 == 1 {
		a.stageTime *= a.stageDuration() / oldDuration
	}
	newTravel := a.travel()
	if oldTravel > 0 {
		a.cameraRow = int(math.Round(float64(a.cameraRow) * float64(newTravel) / float64(oldTravel)))
	} else {
		a.cameraRow = int(math.Round(a.camera * float64(newTravel)))
	}
	a.cached = ""
}
func (a *ANSILoveAnimation) Resize(w, h int) { a.Apply(a.PrepareResize(w, h)()) }
func (a *ANSILoveAnimation) SetBackend(b ANSILoveBackend) {
	if b > ANSILoveBrailleFG {
		return
	}
	if b == ANSILoveAuto {
		b = ANSILoveQuad
	}
	if b == a.cfg.Backend {
		return
	}
	a.cfg.Backend = b
	a.Resize(a.cfg.Width, a.cfg.Height)
}
func (a *ANSILoveAnimation) SetColorDepth(d ANSILoveColorDepth) {
	if d > ANSILove16 || d == a.cfg.Depth {
		return
	}
	a.cfg.Depth = d
	a.Resize(a.cfg.Width, a.cfg.Height)
}
func (a *ANSILoveAnimation) SetCellAspect(w, h int) {
	if w < 1 || w > 1024 || h < 1 || h > 1024 || w*a.cfg.CellH == h*a.cfg.CellW {
		return
	}
	a.cfg.CellW, a.cfg.CellH = w, h
	a.Resize(a.cfg.Width, a.cfg.Height)
}
func (a *ANSILoveAnimation) SetTheme(theme string) error {
	if a.cfg.Theme == theme {
		return nil
	}
	a.cfg.Theme = theme
	if a.cfg.Palette == "theme" {
		a.Resize(a.cfg.Width, a.cfg.Height)
	} else {
		a.generation++
	}
	return nil
}

// SetThemePalette supplies the host's semantic VGA mapping, including custom themes.
func (a *ANSILoveAnimation) SetThemePalette(colors [16][3]uint8) {
	if a.colors != nil && *a.colors == colors {
		return
	}
	a.colors = &colors
	if a.cfg.Palette == "theme" {
		a.Resize(a.cfg.Width, a.cfg.Height)
	} else {
		a.generation++
	}
}
func (a *ANSILoveAnimation) SetState(s ANSILoveState) {
	if s > ANSILoveLogin || s == a.state {
		return
	}
	a.state = s
	a.nextBPM = 66
	if s == ANSILoveLogin {
		a.nextBPM = 60
		if a.elapsed == 0 {
			a.camera = 1
			a.cameraRow = a.travel()
			a.bpm = 60
		}
	}
	if s == ANSILoveIdle {
		a.stage, a.stageTime = 0, 0
		a.camera = 0
		a.cameraRow = 0
	}
	a.cached = ""
}
func (a *ANSILoveAnimation) Reset() {
	a.elapsed = 0
	a.pending = 0
	a.beat = 0
	a.camera = 0
	a.stageTime = 0
	a.stage = 0
	a.cameraRow = 0
	a.stepTime = 0
	a.bpm = a.nextBPM
	a.cached = ""
}
func (a *ANSILoveAnimation) Update() { a.UpdateElapsed(50 * time.Millisecond) }
func (a *ANSILoveAnimation) UpdateElapsed(dt time.Duration) {
	if dt <= 0 || a.cfg.Motion == ANSILoveMotionOff {
		return
	}
	if dt > 250*time.Millisecond {
		dt = 250 * time.Millisecond
	}
	a.pending += dt
	for a.pending >= 50*time.Millisecond {
		a.pending -= 50 * time.Millisecond
		a.advanceFrame(50 * time.Millisecond)
	}
}

func (a *ANSILoveAnimation) advanceFrame(dt time.Duration) {
	a.elapsed += dt
	seconds := dt.Seconds()
	left := (1 - a.beat) * 60 / a.bpm
	if seconds >= left {
		a.beat = 0
		a.bpm = a.nextBPM
		seconds -= left
	}
	a.beat = math.Mod(a.beat+seconds*a.bpm/60, 1)
	if a.state == ANSILoveIdle && a.cfg.Fit != ANSILoveFitContain && a.cfg.Fit != ANSILoveFitExact && a.cfg.Backend != ANSILoveExact {
		remaining := dt.Seconds()
		for remaining > 0 {
			duration := a.stageDuration()
			step := math.Min(remaining, math.Max(0, duration-a.stageTime))
			a.stageTime += step
			remaining -= step
			if a.stageTime >= duration {
				a.stage = (a.stage + 1) % 8
				a.stageTime = 0
			} else {
				break
			}
		}
		points := [5]float64{0, .5, 1, .5, 0}
		leg := a.stage / 2
		a.camera = points[leg]
		if a.stage%2 == 1 {
			x := a.stageTime / a.stageDuration()
			smooth := x * x * x * (x*(6*x-15) + 10)
			a.camera += (points[leg+1] - points[leg]) * smooth
		}
		a.stepTime += dt.Seconds()
		target := int(math.Round(a.camera * float64(a.travel())))
		if a.stepTime >= 1/a.panSpeed() {
			if target > a.cameraRow {
				a.cameraRow++
			} else if target < a.cameraRow {
				a.cameraRow--
			}
			a.stepTime = 0
		}
	}
}
func (a *ANSILoveAnimation) travel() int {
	if a.sheet == nil {
		return 0
	}
	return max(0, a.sheet.layout.SheetRows-a.cfg.Height)
}
func (a *ANSILoveAnimation) panSpeed() float64 {
	if a.cfg.Width >= 320 {
		return 1
	}
	return 2
}
func (a *ANSILoveAnimation) stageDuration() float64 {
	if a.stage%2 == 0 {
		return [4]float64{8, 12, 20, 8}[a.stage/2]
	}
	return math.Max(.05, 1.875*.5*float64(a.travel())/a.panSpeed())
}

func ansiLovePalette(c ANSILoveConfig, override *[16][3]uint8) (lut [112]ansiart.LinRGB) {
	colors := ansiart.VGA
	if c.Palette == "theme" {
		if override != nil {
			colors = *override
		} else if roles, ok := ansiLoveShellThemeRoles(c.Theme); ok {
			for i, name := range roles {
				colors[i] = parseANSILoveHex(name)
			}
		} else {
			p := GetLogoPalette(c.Theme)
			f := GetFirePalette(c.Theme)
			parse := parseANSILoveHex
			danger, accent, warning, secondary := f[4], p[5], f[len(f)-2], p[3]
			fgSecondary, fgMuted, primary, fgPrimary := p[2], p[1], p[3], p[4]
			switch strings.ToLower(c.Theme) {
			case "dracula":
				danger, accent, warning, secondary = f[8], f[4], f[5], f[3]
				fgSecondary, fgMuted = f[2], f[1]
			case "tokyo-night", "tokyonight":
				danger = "#f7768e"
			case "gruvbox":
				danger = "#fb4934"
			}
			slots := [16]string{p[0], danger, accent, warning, secondary, accent, secondary, fgSecondary, fgMuted, danger, accent, warning, primary, accent, primary, fgPrimary}
			for i, s := range slots {
				colors[i] = parse(s)
			}
		}
	}
	for i := range lut {
		lut[i] = ansiart.ToLinear(colors[i%16])
	}
	return
}

func parseANSILoveHex(s string) [3]uint8 {
	r, g, b := hexToRGB(s)
	return [3]uint8{uint8(r), uint8(g), uint8(b)}
}

func ansiLoveShellThemeRoles(theme string) ([16]string, bool) {
	switch strings.ToLower(theme) {
	case "rose-pine", "kanagawa", "noctalia", "eldritch-abyss", "void", "red", "cyan", "coral", "pink", "ayu", "amber", "blue", "purple", "green", "orange":
		p := GetBurnPalette(theme)
		return [16]string{p[8], p[6], p[3], p[5], p[2], p[3], p[2], p[1], p[9], p[6], p[3], p[5], p[4], p[3], p[4], p[0]}, true
	default:
		return [16]string{}, false
	}
}

func ansiLoveGlintOpacity(phase, holdElapsed, holdRemaining float64) float64 {
	if phase < 0 || phase >= 1.5 || holdElapsed < 0 || holdRemaining < 0 {
		return 0
	}
	// Fade over five frames; finish one frame before each eligibility boundary.
	return smoothstep01(phase/.25) *
		smoothstep01((1.5-phase-.05)/.2) *
		smoothstep01(holdElapsed/.25) *
		smoothstep01((holdRemaining-.05)/.2)
}

func smoothstep01(x float64) float64 {
	if x < 0 {
		x = 0
	} else if x > 1 {
		x = 1
	}
	return x * x * (3 - 2*x)
}

func bakeANSILove(c ANSILoveConfig, colors *[16][3]uint8) *ansiLoveSheet {
	grid, raster, _ := loadANSILove()
	sx, sy := 2, 2
	switch c.Backend {
	case ANSILoveHalf:
		sx, sy = 1, 2
	case ANSILoveSextant:
		sy = 3
	case ANSILoveOctant, ANSILoveBrailleFG, ANSILoveBrailleTwoColor:
		sy = 4
	}
	fit := [4]ansiart.FitMode{ansiart.FitPan, ansiart.FitContain, ansiart.FitWidth, ansiart.FitExact}[c.Fit]
	exact := c.Backend == ANSILoveExact || c.Fit == ANSILoveFitExact
	if exact {
		fit = ansiart.FitExact
	}
	s := &ansiLoveSheet{layout: ansiart.ComputeLayout(raster.W, raster.H, c.Width, c.Height, c.CellW, c.CellH, sx, sy, fit, c.Zoom, 3), lut: ansiLovePalette(c, colors)}
	l := s.layout
	if exact {
		l.Left = max(0, (c.Width-80)/2)
		l.Width = min(80, c.Width)
		l.SheetRows = max(c.Height, 157)
		s.layout = l
	}
	n := l.Width * l.SheetRows
	s.cells = make([]ansiLoveCell, n)
	s.flags = make([]uint8, n)
	s.rowFlags = make([]uint8, l.SheetRows)
	s.rows = make([]string, l.SheetRows)
	s.background = ansiLoveCell{ch: ' '}
	s.background.bg, s.background.bi = ansiart.Quantize(s.lut[0], int(c.Depth))
	s.background.fg, s.background.fi = s.background.bg, s.background.bi
	if !exact {
		s.weights = ansiart.BuildWeights(raster, l)
	}
	var patch [128]ansiart.LinRGB
	for y := 0; y < l.SheetRows; y++ {
		for x := 0; x < l.Width; x++ {
			p := y*l.Width + x
			if exact {
				cell := s.background
				gx := x + max(0, (80-c.Width)/2)
				if y < grid.H && gx < grid.W {
					src := grid.Cells[y*grid.W+gx]
					cell.ch = src.Ch
					cell.fg, cell.fi = ansiart.Quantize(s.lut[src.Fg], int(c.Depth))
					cell.bg, cell.bi = ansiart.Quantize(s.lut[src.Bg], int(c.Depth))
				}
				s.cells[p] = cell
				continue
			}
			if c.Depth == ANSILove16 && c.Backend != ANSILoveBrailleFG && c.Backend != ANSILoveBrailleTwoColor {
				ansiart.SampleCellPatch(raster, l, x, y, s.lut[:], patch[:])
				ch, fg, bg := ansiart.Fit16(patch[:])
				s.cells[p] = ansiLoveCell{ch: ch, fg: ansiart.VGA[fg], bg: ansiart.VGA[bg], fi: fg, bi: bg}
				continue
			}
			s.cells[p] = s.fitCell(c, x, y, s.lut[:], -1, 0)
			if c.Motion != ANSILoveMotionOff && c.Depth != ANSILove16 {
				for dy := 0; dy < l.SY; dy++ {
					for dx := 0; dx < l.SX; dx++ {
						sample := (y*l.SY+dy)*(l.Width*l.SX) + x*l.SX + dx
						for k := s.weights.Start[sample]; k < s.weights.Start[sample+1]; k++ {
							slot := s.weights.Slot[k]
							if slot>>4 == ansiart.RegionHeart {
								s.flags[p] |= 1
							}
							if slot>>4 == ansiart.RegionLogo && slot%16 == 7 {
								s.flags[p] |= 2
							}
						}
					}
				}
			}
			s.rowFlags[y] |= s.flags[p]
		}
	}
	var row []byte
	for y := range s.rows {
		row = s.appendRow(row[:0], c, y, false, nil, -1, 0)
		s.rows[y] = string(row)
	}
	// Static and 16-colour sheets never sample again.
	if c.Motion == ANSILoveMotionOff || c.Depth == ANSILove16 {
		s.weights = nil
	}
	return s
}

func (s *ansiLoveSheet) fitCell(c ANSILoveConfig, x, y int, lut []ansiart.LinRGB, glint, glintOpacity float64) ansiLoveCell {
	l := s.layout
	var samples [8]ansiart.LinRGB
	for dy := 0; dy < l.SY; dy++ {
		for dx := 0; dx < l.SX; dx++ {
			n := (y*l.SY+dy)*(l.Width*l.SX) + x*l.SX + dx
			samples[dy*l.SX+dx] = ansiart.Sample(s.weights, n, lut)
		}
	}
	if glint >= 0 {
		// Only logo-grey weights brighten; white glyphs and neighbouring heart stay intact.
		u := l.OX + float64(l.Left+x)*l.CellW/l.Scale
		v := float64(y) * l.CellH / l.Scale
		strength := math.Max(0, 1-math.Abs(u+v-glint)/32) * .2 * glintOpacity
		for dy := 0; dy < l.SY; dy++ {
			for dx := 0; dx < l.SX; dx++ {
				n := (y*l.SY+dy)*(l.Width*l.SX) + x*l.SX + dx
				for k := s.weights.Start[n]; k < s.weights.Start[n+1]; k++ {
					slot := s.weights.Slot[k]
					if slot == ansiart.RegionLogo<<4|7 {
						w := float64(s.weights.W[k]) / 65535 * strength
						for j := 0; j < 3; j++ {
							samples[dy*l.SX+dx][j] += (1 - lut[slot][j]) * w
						}
					}
				}
			}
		}
	}
	cell := ansiLoveCell{}
	if c.Backend == ANSILoveBrailleFG {
		var mask uint8
		var fg ansiart.LinRGB
		count := 0
		thresholds := [8]float64{.06, .18, .24, .12, .105, .225, .165, .045}
		for i, p := range samples {
			lum := .2126*p[0] + .7152*p[1] + .0722*p[2]
			if lum > thresholds[i] {
				mask |= 1 << i
				count++
				for j := range fg {
					fg[j] += p[j]
				}
			}
		}
		if count > 0 {
			for j := range fg {
				fg[j] /= float64(count)
			}
		}
		cell.ch = ansiart.BrailleRune(mask)
		if mask == 0 {
			cell.ch = ' '
		}
		cell.fg, cell.fi = ansiart.Quantize(fg, int(c.Depth))
		cell.bg, cell.bi = s.background.bg, s.background.bi
		return cell
	}
	mask, fg, bg := ansiart.Fit2(samples[:l.SX*l.SY], false)
	cell.fg, cell.fi = ansiart.Quantize(fg, int(c.Depth))
	cell.bg, cell.bi = ansiart.Quantize(bg, int(c.Depth))
	switch c.Backend {
	case ANSILoveHalf:
		cell.ch = [4]rune{' ', '▀', '▄', '█'}[mask]
	case ANSILoveSextant:
		cell.ch = ansiart.SextantRune(mask)
	case ANSILoveOctant:
		cell.ch = ansiart.OctantRune(mask)
	case ANSILoveBrailleTwoColor:
		cell.ch = ansiart.BrailleRune(mask)
	default:
		cell.ch = ansiart.QuadRunes[mask]
	}
	if cell.fg == cell.bg {
		cell.ch = ' '
	}
	if cell.ch == ' ' {
		cell.fg, cell.fi = cell.bg, cell.bi
	}
	return cell
}

func ansiLoveSGRParams(out []byte, rgb [3]uint8, index uint8, depth ANSILoveColorDepth, bg bool) []byte {
	if depth == ANSILove16 {
		code := 30 + int(index&7)
		if index >= 8 {
			code += 60
		}
		if bg {
			code += 10
		}
		out = strconv.AppendInt(out, int64(code), 10)
	} else {
		code := 38
		if bg {
			code = 48
		}
		out = strconv.AppendInt(out, int64(code), 10)
		if depth == ANSILove256 {
			out = append(out, ';', '5', ';')
			out = strconv.AppendUint(out, uint64(index), 10)
		} else {
			out = append(out, ';', '2')
			for _, v := range rgb {
				out = append(out, ';')
				out = strconv.AppendUint(out, uint64(v), 10)
			}
		}
	}
	return out
}

func ansiLoveReverseColors(cell ansiLoveCell) ansiLoveCell {
	cell.fg, cell.bg = cell.bg, cell.fg
	cell.fi, cell.bi = cell.bi, cell.fi
	return cell
}

func ansiLoveDigits(n uint8) int {
	switch {
	case n < 10:
		return 1
	case n < 100:
		return 2
	default:
		return 3
	}
}

func ansiLoveIntDigits(n int) int {
	digits := 1
	for n >= 10 {
		n /= 10
		digits++
	}
	return digits
}

func ansiLoveSGRColorSize(rgb [3]uint8, index uint8, depth ANSILoveColorDepth, bg bool) int {
	if depth == ANSILove16 {
		code := 30 + int(index&7)
		if index >= 8 {
			code += 60
		}
		if bg {
			code += 10
		}
		digits := 2
		if code >= 100 {
			digits = 3
		}
		return 3 + digits // ESC [ code m
	}
	if depth == ANSILove256 {
		return 8 + ansiLoveDigits(index) // ESC [ 48;5;index m
	}
	return 10 + ansiLoveDigits(rgb[0]) + ansiLoveDigits(rgb[1]) + ansiLoveDigits(rgb[2])
}

func ansiLoveSGRTransitionSize(previous ansiLoveCell, previousValid, previousReverse bool, next ansiLoveCell, nextReverse bool, depth ANSILoveColorDepth) int {
	bgChanged := !previousValid || previous.bg != next.bg
	fgChanged := !previousValid || previous.fg != next.fg
	colorSize := 0
	switch {
	case bgChanged && fgChanged:
		colorSize = ansiLoveSGRColorSize(next.bg, next.bi, depth, true) + ansiLoveSGRColorSize(next.fg, next.fi, depth, false) - 2
	case bgChanged:
		colorSize = ansiLoveSGRColorSize(next.bg, next.bi, depth, true)
	case fgChanged:
		colorSize = ansiLoveSGRColorSize(next.fg, next.fi, depth, false)
	}
	if previousReverse != nextReverse {
		if colorSize == 0 {
			if nextReverse {
				colorSize = 4 // ESC[7m
			} else {
				colorSize = 5 // ESC[27m
			}
		} else if nextReverse {
			colorSize += 2 // append ;7 to the colour SGR
		} else {
			colorSize += 3 // append ;27 to the colour SGR
		}
	}
	return colorSize
}

func ansiLoveAppendREP(out []byte, count int) []byte {
	out = append(out, '\x1b', '[')
	out = strconv.AppendInt(out, int64(count), 10)
	return append(out, 'b')
}

func ansiLoveAppendSGRTransition(out []byte, previous ansiLoveCell, previousValid bool, next ansiLoveCell, depth ANSILoveColorDepth, reverse, reverseChanged bool) []byte {
	bgChanged := !previousValid || previous.bg != next.bg
	fgChanged := !previousValid || previous.fg != next.fg
	hasColorChange := bgChanged || fgChanged
	if !hasColorChange && !reverseChanged {
		return out
	}
	if hasColorChange || reverseChanged {
		out = append(out, '\x1b', '[')
	}
	switch {
	case bgChanged && fgChanged:
		out = ansiLoveSGRParams(out, next.bg, next.bi, depth, true)
		out = append(out, ';')
		out = ansiLoveSGRParams(out, next.fg, next.fi, depth, false)
	case bgChanged:
		out = ansiLoveSGRParams(out, next.bg, next.bi, depth, true)
	case fgChanged:
		out = ansiLoveSGRParams(out, next.fg, next.fi, depth, false)
	}
	if reverseChanged {
		if hasColorChange {
			out = append(out, ';')
		}
		if reverse {
			out = append(out, '7')
		} else {
			out = append(out, '2', '7')
		}
	}
	return append(out, 'm')
}

func (s *ansiLoveSheet) appendRow(out []byte, c ANSILoveConfig, row int, credit bool, lut []ansiart.LinRGB, glint, glintOpacity float64) []byte {
	var cells [512]ansiLoveCell
	for x := 0; x < c.Width; x++ {
		cell := s.background
		if x >= s.layout.Left && x < s.layout.Left+s.layout.Width && row < s.layout.SheetRows {
			col := x - s.layout.Left
			p := row*s.layout.Width + col
			cell = s.cells[p]
			if lut != nil && s.flags[p] != 0 {
				cell = s.fitCell(c, col, row, lut, glint, glintOpacity)
			}
		}
		if credit && c.Width >= len(ansiLoveCredit) && x >= c.Width-len(ansiLoveCredit) {
			cell.ch = rune(ansiLoveCredit[x-(c.Width-len(ansiLoveCredit))])
			cell.fg = [3]uint8{170, 170, 170}
			cell.bg = [3]uint8{}
			cell.fi = 7
			cell.bi = 0
			if c.Depth == ANSILove256 {
				cell.fi = 248
				cell.bi = 16
			}
		}
		cells[x] = cell
	}

	width := c.Width
	var sizes [512][2]int
	var previousState [512][2]uint8
	maxInt := int(^uint(0) >> 1)
	for x := 0; x < width; x++ {
		for state := 0; state < 2; state++ {
			current := cells[x]
			if state == 1 {
				current = ansiLoveReverseColors(current)
			}
			glyphSize := utf8.RuneLen(cells[x].ch)
			if x == 0 {
				sizes[x][state] = ansiLoveSGRTransitionSize(ansiLoveCell{}, false, false, current, state == 1, c.Depth) + glyphSize
				continue
			}
			best := maxInt
			for previous := 0; previous < 2; previous++ {
				last := cells[x-1]
				if previous == 1 {
					last = ansiLoveReverseColors(last)
				}
				candidate := sizes[x-1][previous] + ansiLoveSGRTransitionSize(last, true, previous == 1, current, state == 1, c.Depth) + glyphSize
				if candidate < best {
					best = candidate
					sizes[x][state] = candidate
					previousState[x][state] = uint8(previous)
				}
			}
		}
	}
	if width > 0 {
		var state [512]uint8
		last := 0
		if sizes[width-1][1] < sizes[width-1][0] {
			last = 1
		}
		for x := width - 1; x >= 0; x-- {
			state[x] = uint8(last)
			last = int(previousState[x][last])
		}
		var previous ansiLoveCell
		previousValid, previousReverse := false, false
		for x := 0; x < width; {
			reverse := state[x] == 1
			cell := cells[x]
			if reverse {
				cell = ansiLoveReverseColors(cell)
			}
			out = ansiLoveAppendSGRTransition(out, previous, previousValid, cell, c.Depth, reverse, reverse != previousReverse)
			out = utf8.AppendRune(out, cells[x].ch)
			runEnd := x + 1
			if cells[x].ch == ' ' {
				for runEnd < width && cells[runEnd].ch == ' ' && cells[runEnd].fg == cells[x].fg && cells[runEnd].bg == cells[x].bg && state[runEnd] == state[x] {
					runEnd++
				}
			}
			runLength := runEnd - x
			repeats := runLength - 1
			repSize := 3 + ansiLoveIntDigits(repeats)
			if runLength > 1 && 1+repSize < runLength {
				out = ansiLoveAppendREP(out, repeats)
			} else {
				for range repeats {
					out = append(out, ' ')
				}
			}
			previous, previousValid, previousReverse = cell, true, reverse
			x = runEnd
		}
	}
	return append(out, "\x1b[m"...)
}

func (a *ANSILoveAnimation) Render() string {
	envelope := math.Min(1, math.Exp(-math.Pow((a.beat-.06)/.045, 2))+.55*math.Exp(-math.Pow((a.beat-.24)/.05, 2)))
	level := int(math.Round(envelope * 23))
	glintTick := -1
	glintLevel := 0
	if a.cfg.Motion == ANSILoveMotionOff || a.cfg.Depth == ANSILove16 {
		level = 0
	} else if a.state == ANSILoveIdle && a.stage%2 == 0 {
		phase := math.Mod(a.elapsed.Seconds(), 9)
		if phase < 1.5 {
			alpha := ansiLoveGlintOpacity(phase, a.stageTime, a.stageDuration()-a.stageTime)
			glintLevel = int(math.Round(alpha * 23))
			if glintLevel > 0 {
				glintTick = int(phase / .05)
			}
		}
	}
	offset := min(a.travel(), max(0, a.cameraRow))
	key := [5]int{level, glintTick, glintLevel, offset, int(a.state)}
	if a.cached != "" && a.key == key {
		return a.cached
	}
	lut := a.sheet.lut
	strength := .35 * float64(level) / 23
	for _, idx := range []int{1, 9} {
		slot := int(ansiart.RegionHeart)*16 + idx
		target := ansiart.ToLinear([3]uint8{255, 128, 128})
		if idx == 9 {
			target = ansiart.ToLinear([3]uint8{217, 217, 217})
		}
		for j := 0; j < 3; j++ {
			lut[slot][j] += (target[j] - lut[slot][j]) * strength
		}
	}
	glint := -1.0
	glintOpacity := float64(glintLevel) / 23
	if glintTick >= 0 {
		glint = 1664 - 64 + float64(glintTick)/29*(640+352+128)
	}
	a.scratch = a.scratch[:0]
	for y := 0; y < a.cfg.Height; y++ {
		row := y + offset
		credit := y == a.cfg.Height-1
		dynamic := a.sheet.rowFlags[row]&1 != 0 && level != 0 || a.sheet.rowFlags[row]&2 != 0 && glint >= 0
		if !credit && !dynamic {
			a.scratch = append(a.scratch, a.sheet.rows[row]...)
		} else {
			var active []ansiart.LinRGB
			if dynamic {
				active = lut[:]
			}
			a.scratch = a.sheet.appendRow(a.scratch, a.cfg, row, credit, active, glint, glintOpacity)
		}
		if y < a.cfg.Height-1 {
			a.scratch = append(a.scratch, '\n')
		}
	}
	a.cached = string(a.scratch)
	a.key = key
	return a.cached
}
