package animations

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/Nomadcxx/sysc-Go/animations/internal/ansiart"
	"github.com/charmbracelet/x/ansi"
)

func TestANSILoveConvenienceConstructorUsesTrueColorOutput(t *testing.T) {
	rendered := NewANSILoveEffect(80, 24, "nord").Render()
	if !strings.Contains(rendered, "\x1b[38;2;") && !strings.Contains(rendered, "\x1b[48;2;") {
		t.Fatal("default ANSI Love output contains no truecolor SGR")
	}
}

func TestANSILoveGeometry(t *testing.T) {
	for _, backend := range []ANSILoveBackend{ANSILoveExact, ANSILoveHalf, ANSILoveQuad, ANSILoveSextant, ANSILoveOctant, ANSILoveBrailleFG, ANSILoveBrailleTwoColor} {
		for _, size := range [][2]int{{1, 1}, {21, 24}, {40, 12}, {80, 24}, {80, 157}, {120, 36}, {160, 314}, {200, 55}, {320, 90}, {344, 72}, {480, 135}, {600, 300}} {
			a, err := NewANSILoveEffectConfig(ANSILoveConfig{Width: size[0], Height: size[1], Backend: backend})
			if err != nil {
				t.Fatal(err)
			}
			frame := a.Render()
			lines := strings.Split(frame, "\n")
			if len(lines) != min(256, size[1]) {
				t.Fatalf("backend %d size %v: %d rows", backend, size, len(lines))
			}
			for y, line := range lines {
				if !strings.HasSuffix(line, "\x1b[m") {
					t.Fatalf("backend %d size %v row %d: wrong width/reset", backend, size, y)
				}
			}
			_, width, height := parseANSILoveFrame(t, frame)
			if width != min(512, size[0]) || height != min(256, size[1]) {
				t.Fatalf("backend %d size %v: parsed frame is %dx%d", backend, size, width, height)
			}
			if size[0] >= 30 && !strings.HasSuffix(ansi.Strip(lines[len(lines)-1]), "Drawn by Warpus, December 2019") {
				t.Fatal("missing credit")
			}
		}
	}
}

func TestANSILoveClockAndCache(t *testing.T) {
	a := NewANSILoveEffect(120, 36, "dracula")
	b := NewANSILoveEffect(120, 36, "dracula")
	a.SetState(ANSILoveLogin)
	b.SetState(ANSILoveLogin)
	for i := 0; i < 13; i++ {
		a.UpdateElapsed(50 * time.Millisecond)
		b.UpdateElapsed(50 * time.Millisecond)
	}
	frame := a.Render()
	if frame != b.Render() {
		t.Fatal("nondeterministic frame")
	}
	if n := testing.AllocsPerRun(20, func() { a.Render() }); n != 0 {
		t.Fatalf("cached render allocated %g", n)
	}
	a.UpdateElapsed(-time.Second)
	if a.Render() != frame {
		t.Fatal("negative dt advanced")
	}
	a.Resize(200, 55)
	a.Resize(120, 36)
	if a.Render() != frame {
		t.Fatal("resize changed phase")
	}
	a.SetCellAspect(16, 32)
	if a.Render() != frame {
		t.Fatal("absolute cell size affected ratio")
	}
	a.SetCellAspect(0, 0)
	a.SetCellAspect(2000, 10)
	if a.Render() != frame {
		t.Fatal("invalid aspect accepted")
	}
	a.Reset()
	if a.elapsed != 0 || a.camera != 0 {
		t.Fatal("reset retained time or camera")
	}
}

func TestANSILoveAdvancesIn20FPSBuckets(t *testing.T) {
	a := NewANSILoveEffect(120, 36, "")
	a.SetState(ANSILoveLogin)
	frame := a.Render()
	for i := 0; i < 3; i++ {
		a.UpdateElapsed(16 * time.Millisecond)
		if got := a.Render(); got != frame {
			t.Fatalf("frame changed after %d ms, before the 50 ms bucket", (i+1)*16)
		}
	}
	if a.elapsed != 0 {
		t.Fatalf("active time advanced to %s before a full bucket", a.elapsed)
	}
	a.UpdateElapsed(16 * time.Millisecond)
	if a.elapsed != 50*time.Millisecond || a.Render() == frame {
		t.Fatal("completed 50 ms bucket did not advance exactly once")
	}
	b := NewANSILoveEffect(120, 36, "")
	b.SetState(ANSILoveLogin)
	b.UpdateElapsed(64 * time.Millisecond)
	if b.elapsed != a.elapsed || b.Render() != a.Render() {
		t.Fatal("equivalent elapsed-time partitions produced different frames")
	}
}

func TestANSILoveResizePreservesEmittedCameraFraction(t *testing.T) {
	a := NewANSILoveEffect(120, 36, "")
	oldTravel := a.travel()
	a.camera = .5 // desired camera can lead the rate-limited emitted row
	a.cameraRow = 1
	want := int(math.Round(float64(a.cameraRow) / float64(oldTravel) * float64(a.travel())))
	a.Resize(120, 36)
	if a.cameraRow != want {
		t.Fatalf("same-size resize moved emitted row from 1 to %d", a.cameraRow)
	}
	oldRow := a.cameraRow
	oldTravel = a.travel()
	prepared := a.PrepareResize(200, 55)()
	newTravel := max(0, prepared.sheet.layout.SheetRows-prepared.cfg.Height)
	want = int(math.Round(float64(oldRow) / float64(oldTravel) * float64(newTravel)))
	a.Apply(prepared)
	if a.cameraRow != want {
		t.Fatalf("resize camera row=%d, want preserved fraction row %d", a.cameraRow, want)
	}
}

func TestANSILoveThemePaletteUsesSemanticRoles(t *testing.T) {
	dracula := [16][3]uint8{
		{40, 42, 54},    // BgBase
		{255, 85, 85},   // Danger
		{80, 250, 123},  // Accent
		{241, 250, 140}, // Warning
		{139, 233, 253}, // Secondary
		{80, 250, 123},  // Accent
		{139, 233, 253}, // Secondary
		{98, 114, 164},  // FgSecondary
		{68, 71, 90},    // FgMuted
		{255, 85, 85},   // Danger
		{80, 250, 123},  // Accent
		{241, 250, 140}, // Warning
		{189, 147, 249}, // Primary
		{80, 250, 123},  // Accent
		{189, 147, 249}, // Primary
		{248, 248, 242}, // FgPrimary
	}
	lut := ansiLovePalette(ANSILoveConfig{Theme: "dracula", Palette: "theme"}, nil)
	for slot, want := range dracula {
		if got := lut[slot]; got != ansiart.ToLinear(want) {
			t.Errorf("Dracula slot %d=%v, want %v", slot, got, ansiart.ToLinear(want))
		}
	}

	for _, test := range []struct {
		theme string
		slot  int
		want  [3]uint8
	}{
		{"blue", 3, [3]uint8{248, 211, 138}},      // Warning
		{"rose-pine", 4, [3]uint8{156, 207, 216}}, // Secondary
	} {
		lut := ansiLovePalette(ANSILoveConfig{Theme: test.theme, Palette: "theme"}, nil)
		if got, want := lut[test.slot], ansiart.ToLinear(test.want); got != want {
			t.Errorf("theme %s slot %d=%v, want %v", test.theme, test.slot, got, want)
		}
	}
}

func TestANSILoveCombinesChangedForegroundAndBackgroundSGR(t *testing.T) {
	sheet := &ansiLoveSheet{
		layout: ansiart.Layout{Width: 1, SheetRows: 1},
		cells:  []ansiLoveCell{{ch: 'X', fg: [3]uint8{1, 2, 3}, bg: [3]uint8{4, 5, 6}}},
	}
	got := string(sheet.appendRow(nil, ANSILoveConfig{Width: 1, Depth: ANSILoveTrueColor}, 0, false, nil, -1, 0))
	want := "\x1b[48;2;4;5;6;38;2;1;2;3mX\x1b[m"
	if got != want {
		t.Fatalf("combined SGR=%q, want %q", got, want)
	}
}

func TestANSILoveUsesReverseVideoForSwappedColorPairs(t *testing.T) {
	red, blue := [3]uint8{255, 0, 0}, [3]uint8{0, 0, 255}
	sheet := &ansiLoveSheet{
		layout: ansiart.Layout{Width: 3, SheetRows: 1},
		cells: []ansiLoveCell{
			{ch: 'A', fg: red, bg: blue},
			{ch: 'B', fg: blue, bg: red},
			{ch: 'C', fg: blue, bg: red},
		},
	}
	frame := string(sheet.appendRow(nil, ANSILoveConfig{Width: 3, Backend: ANSILoveExact, Depth: ANSILoveTrueColor}, 0, false, nil, -1, 0))
	if !strings.Contains(frame, "\x1b[7m") {
		t.Fatal("swapped color pair did not use reverse video")
	}
	cells, width, height := parseANSILoveFrame(t, frame)
	if width != 3 || height != 1 || cells[0].fg != red || cells[0].bg != blue || cells[1].fg != blue || cells[1].bg != red || cells[2].fg != blue || cells[2].bg != red {
		t.Fatalf("reverse-video colors are not visually equivalent: %+v", cells)
	}
}

func TestANSILoveCompressesLongBlankRunsWithREP(t *testing.T) {
	black := [3]uint8{}
	cells := make([]ansiLoveCell, 32)
	for i := range cells {
		cells[i] = ansiLoveCell{ch: ' ', fg: black, bg: black}
	}
	sheet := &ansiLoveSheet{layout: ansiart.Layout{Width: len(cells), SheetRows: 1}, cells: cells}
	frame := string(sheet.appendRow(nil, ANSILoveConfig{Width: len(cells), Backend: ANSILoveExact, Depth: ANSILoveTrueColor}, 0, false, nil, -1, 0))
	if !strings.Contains(frame, "\x1b[31b") {
		t.Fatalf("32 blank cells were not compressed with REP: %q", frame)
	}
	parsed, width, height := parseANSILoveFrame(t, frame)
	if width != len(cells) || height != 1 || len(parsed) != len(cells) {
		t.Fatalf("REP expanded to %dx%d (%d cells), want %d blank cells", width, height, len(parsed), len(cells))
	}
	for i, cell := range parsed {
		if cell.ch != ' ' || cell.fg != black || cell.bg != black {
			t.Fatalf("cell %d after REP = %+v, want black space", i, cell)
		}
	}
}

func TestANSILoveGlintFadesIntoAndOutOfAVisibleHold(t *testing.T) {
	cfg := ANSILoveConfig{Width: 1, Height: 1, Backend: ANSILoveQuad, Palette: "vga", Motion: ANSILoveMotionSubtle, CellW: 1, CellH: 1}
	weights := &ansiart.Weights{
		Start: []uint32{0, 1, 2, 3, 4},
		Slot:  []uint8{uint8(ansiart.RegionLogo<<4 | 7), uint8(ansiart.RegionLogo<<4 | 7), uint8(ansiart.RegionLogo<<4 | 7), uint8(ansiart.RegionLogo<<4 | 7)},
		W:     []uint16{65535, 65535, 65535, 65535},
	}
	sheet := &ansiLoveSheet{
		layout:  ansiart.Layout{SX: 2, SY: 2, SheetRows: 1, Scale: 1, OX: 1600, CellW: 1, CellH: 1, Width: 1},
		weights: weights,
		cells:   []ansiLoveCell{{ch: '█', fg: ansiart.VGA[7], bg: ansiart.VGA[0], fi: 7, bi: 0}},
		flags:   []uint8{2}, rowFlags: []uint8{2}, rows: []string{""},
	}
	static := sheet.appendRow(nil, cfg, 0, false, nil, -1, 0)
	sheet.rows[0] = string(static)
	a := &ANSILoveAnimation{cfg: cfg, sheet: sheet, state: ANSILoveIdle, bpm: 66, nextBPM: 66}
	a.elapsed = 2 * time.Second
	withoutGlint := a.Render()
	a.elapsed, a.cached = 0, ""
	if withGlint := a.Render(); withGlint != withoutGlint {
		t.Fatal("glint jumped to full strength at the start of its window")
	}
	phase := .75
	tick := int(phase / .05)
	a.sheet.layout.OX = 1600 + float64(tick)/29*(640+352+128)
	a.elapsed, a.stageTime, a.cached = 750*time.Millisecond, phase, ""
	withGlint := a.Render()
	if withGlint == withoutGlint {
		t.Fatal("visible glint did not change the logo")
	}
	a.sheet.layout.OX = 2720
	a.elapsed, a.cached = 1450*time.Millisecond, ""
	if withGlint := a.Render(); withGlint != withoutGlint {
		t.Fatal("glint remained visible at the end of its window")
	}
}

func TestANSILoveAsyncResize(t *testing.T) {
	a := NewANSILoveEffect(80, 24, "")
	old := a.PrepareResize(120, 36)
	newer := a.PrepareResize(200, 55)
	a.UpdateElapsed(100 * time.Millisecond)
	elapsed := a.elapsed
	a.Apply(newer())
	a.Apply(old())
	if a.cfg.Width != 200 || a.cfg.Height != 55 || a.elapsed != elapsed {
		t.Fatal("stale resize won or reset clock")
	}
	pending := a.PrepareResize(480, 135)
	a.SetBackend(ANSILoveBrailleFG)
	a.Apply(pending())
	if a.cfg.Backend != ANSILoveBrailleFG || a.cfg.Width != 200 {
		t.Fatal("stale backend/config result applied")
	}
}

func TestANSILovePreparedThemePalettePreservesAnimation(t *testing.T) {
	first, second := [16][3]uint8{}, [16][3]uint8{}
	for i := range first {
		first[i] = [3]uint8{uint8(i * 13), uint8(i * 7), uint8(i * 3)}
		second[i] = [3]uint8{uint8(255 - i*8), uint8(240 - i*5), uint8(220 - i*3)}
	}
	cfg := ANSILoveConfig{Width: 80, Height: 24, Theme: "dracula", Palette: "theme", Motion: ANSILoveMotionSubtle}
	got, err := NewANSILoveEffectConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	want, err := NewANSILoveEffectConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	got.SetThemePalette(first)
	want.SetThemePalette(first)
	got.SetState(ANSILoveLogin)
	want.SetState(ANSILoveLogin)
	for range 40 {
		got.UpdateElapsed(50 * time.Millisecond)
		want.UpdateElapsed(50 * time.Millisecond)
	}
	before := got.Render()
	elapsed, stage, stageTime, cameraRow := got.elapsed, got.stage, got.stageTime, got.cameraRow
	prepared := got.PrepareThemePalette("nord", second)()
	if got.Render() != before {
		t.Fatal("theme preparation changed the visible frame before Apply")
	}
	got.Apply(prepared)
	want.SetTheme("nord")
	want.SetThemePalette(second)
	if got.Render() != want.Render() || got.elapsed != elapsed || got.stage != stage || got.stageTime != stageTime || got.cameraRow != cameraRow {
		t.Fatal("prepared theme palette did not preserve the animation state")
	}
}

func TestANSILovePreparedThemePaletteLayoutPreservesAnimation(t *testing.T) {
	first, second := [16][3]uint8{}, [16][3]uint8{}
	for i := range first {
		first[i] = [3]uint8{uint8(i * 13), uint8(i * 7), uint8(i * 3)}
		second[i] = [3]uint8{uint8(255 - i*8), uint8(240 - i*5), uint8(220 - i*3)}
	}
	cfg := ANSILoveConfig{Width: 80, Height: 24, Theme: "dracula", Palette: "theme", Motion: ANSILoveMotionSubtle}
	got, err := NewANSILoveEffectConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	wantCfg := cfg
	wantCfg.Width, wantCfg.Height, wantCfg.CellW, wantCfg.CellH, wantCfg.Theme = 120, 36, 9, 18, "nord"
	want, err := NewANSILoveEffectConfig(wantCfg)
	if err != nil {
		t.Fatal(err)
	}
	got.SetThemePalette(first)
	got.SetState(ANSILoveLogin)
	want.SetThemePalette(second)
	want.SetState(ANSILoveLogin)
	for range 9 {
		got.UpdateElapsed(50 * time.Millisecond)
		want.UpdateElapsed(50 * time.Millisecond)
	}
	before := got.Render()
	prepared := got.PrepareThemePaletteLayout("nord", second, 120, 36, 9, 18)()
	if got.Render() != before {
		t.Fatal("theme and layout preparation changed the visible frame before Apply")
	}
	got.Apply(prepared)
	if got.Render() != want.Render() || got.cfg != want.cfg || got.elapsed != want.elapsed || got.stage != want.stage || got.stageTime != want.stageTime {
		t.Fatal("prepared theme and layout did not preserve the animation state")
	}
}

func TestANSILoveValidation(t *testing.T) {
	for _, cfg := range []ANSILoveConfig{
		{Backend: 255}, {Depth: 255}, {Fit: 255}, {Motion: 255},
		{Zoom: math.NaN()}, {Zoom: math.Inf(1)}, {Zoom: -1},
		{Palette: "bogus"}, {Source: "/tmp/untrusted.ans"},
		{CellW: -1}, {CellW: 2048, CellH: 10},
	} {
		if _, err := NewANSILoveEffectConfig(cfg); err == nil {
			t.Fatalf("accepted invalid config %+v", cfg)
		}
	}
}

func TestANSILoveLoginAndMotionOff(t *testing.T) {
	a := NewANSILoveEffect(120, 36, "")
	a.SetState(ANSILoveLogin)
	start := a.camera
	for i := 0; i < 400; i++ {
		a.UpdateElapsed(50 * time.Millisecond)
	}
	if a.camera != start || start != 1 {
		t.Fatal("login camera moved")
	}
	b, err := NewANSILoveEffectConfig(ANSILoveConfig{Width: 120, Height: 36, Motion: ANSILoveMotionOff})
	if err != nil {
		t.Fatal(err)
	}
	frame := b.Render()
	for i := 0; i < 400; i++ {
		b.UpdateElapsed(50 * time.Millisecond)
	}
	if frame != b.Render() {
		t.Fatal("motion-off changed frame")
	}
}

func TestANSILoveRegistry(t *testing.T) {
	if m := GetEffectMetadata("ansi-love"); m == nil || m.Category != "scene" || m.RequiresText {
		t.Fatal("missing standalone ANSI Love scene")
	}
}

func BenchmarkANSILove(b *testing.B) {
	for _, backend := range []ANSILoveBackend{
		ANSILoveHalf,
		ANSILoveQuad,
		ANSILoveSextant,
		ANSILoveOctant,
		ANSILoveBrailleTwoColor,
		ANSILoveBrailleFG,
	} {
		for _, size := range [][2]int{{200, 55}, {480, 135}} {
			name := fmt.Sprintf("%d/%dx%d", backend, size[0], size[1])
			b.Run(name+"/bake", func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					a, _ := NewANSILoveEffectConfig(ANSILoveConfig{Width: size[0], Height: size[1], Backend: backend, Motion: ANSILoveMotionSubtle})
					_ = a
				}
			})
			a, _ := NewANSILoveEffectConfig(ANSILoveConfig{Width: size[0], Height: size[1], Backend: backend, Motion: ANSILoveMotionSubtle})
			a.SetState(ANSILoveLogin)
			a.Render()
			b.Run(name+"/cached", func(b *testing.B) {
				b.ReportAllocs()
				var frame string
				for i := 0; i < b.N; i++ {
					frame = a.Render()
				}
				b.ReportMetric(float64(len(frame)), "render_B")
			})
			b.Run(name+"/animated", func(b *testing.B) {
				b.ReportAllocs()
				var frame string
				for i := 0; i < b.N; i++ {
					a.UpdateElapsed(50 * time.Millisecond)
					frame = a.Render()
				}
				b.ReportMetric(float64(len(frame)), "render_B")
			})
			b.Run(name+"/pan", func(b *testing.B) {
				b.ReportAllocs()
				var frame string
				for i := 0; i < b.N; i++ {
					a.cameraRow = i % (a.travel() + 1)
					frame = a.Render()
				}
				b.ReportMetric(float64(len(frame)), "render_B")
			})
		}
	}
}

func TestANSILoveExactFitDoesNotTour(t *testing.T) {
	a, err := NewANSILoveEffectConfig(ANSILoveConfig{Width: 120, Height: 36, Fit: ANSILoveFitExact, Motion: ANSILoveMotionSubtle})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 400; i++ {
		a.UpdateElapsed(50 * time.Millisecond)
	}
	if a.camera != 0 || a.cameraRow != 0 {
		t.Fatalf("exact fit moved camera: %v %d", a.camera, a.cameraRow)
	}
}
