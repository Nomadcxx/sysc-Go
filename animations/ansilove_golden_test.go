package animations

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/Nomadcxx/sysc-Go/animations/internal/ansiart"
)

var updateANSILoveGoldens = flag.Bool("update", false, "regenerate ANSI Love SGR golden files and optional PNG previews")

var ansiLoveGoldenBackends = []struct {
	name    string
	backend ANSILoveBackend
}{
	{"quad", ANSILoveQuad}, {"sextant", ANSILoveSextant}, {"braille-fg", ANSILoveBrailleFG}, {"braille-2c", ANSILoveBrailleTwoColor},
}

func TestANSILoveGoldens(t *testing.T) {
	for _, backend := range ansiLoveGoldenBackends {
		t.Run(backend.name, func(t *testing.T) {
			actual := map[string]string{}
			for _, size := range [][2]int{{80, 24}, {120, 36}, {200, 55}, {480, 135}} {
				for depth := ANSILoveTrueColor; depth <= ANSILove16; depth++ {
					a, err := NewANSILoveEffectConfig(ANSILoveConfig{Width: size[0], Height: size[1], Backend: backend.backend, Depth: depth, Motion: ANSILoveMotionSubtle})
					if err != nil {
						t.Fatal(err)
					}
					a.SetState(ANSILoveLogin)
					previous := time.Duration(0)
					for _, at := range []time.Duration{0, 50 * time.Millisecond, 3300 * time.Millisecond} {
						for dt := at - previous; dt > 0; {
							step := dt
							if step > 50*time.Millisecond {
								step = 50 * time.Millisecond
							}
							a.UpdateElapsed(step)
							dt -= step
						}
						previous = at
						name := ansiLoveGoldenName(backend.name, size[0], size[1], depth, at)
						actual[name] = a.Render()
					}
					if size == [2]int{200, 55} && depth == ANSILoveTrueColor {
						for _, fraction := range []float64{0, .5, 1} {
							a.camera = fraction
							a.cameraRow = int(math.Round(fraction * float64(a.travel())))
							a.cached = ""
							name := filepath.Join("testdata", "ansilove", fmt.Sprintf("%s_200x55_truecolor_camera-%.1f.golden", backend.name, fraction))
							actual[name] = a.Render()
						}
					}
				}
			}
			if *updateANSILoveGoldens {
				for name, frame := range actual {
					if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(name, []byte(frame), 0644); err != nil {
						t.Fatal(err)
					}
				}
			} else {
				for name, got := range actual {
					want, err := os.ReadFile(name)
					if err != nil {
						t.Fatalf("read golden %s: %v", name, err)
					}
					if string(want) != got {
						t.Errorf("%s differs; inspect preview then regenerate with -update", name)
					}
				}
			}
			if *updateANSILoveGoldens {
				if dir := os.Getenv("ANSILOVE_PREVIEWS"); dir != "" {
					font, err := os.ReadFile("cmd/bake-ansilove/ibm-vga8.bin")
					if err != nil {
						t.Fatal(err)
					}
					for name, frame := range actual {
						cells, w, h := parseANSILoveFrame(t, frame)
						writeANSILovePNG(t, filepath.Join(dir, backend.name, strings.TrimSuffix(filepath.Base(name), ".golden")+".png"), cells, w, h, font)
					}
				}
			}
		})
	}
}

func ansiLoveGoldenName(backend string, width, height int, depth ANSILoveColorDepth, at time.Duration) string {
	depthName := [...]string{"truecolor", "256", "16"}[depth]
	return filepath.Join("testdata", "ansilove", fmt.Sprintf("%s_%dx%d_%s_t-%04dms.golden", backend, width, height, depthName, at.Milliseconds()))
}

type ansiLovePreviewCell struct {
	ch     rune
	fg, bg [3]uint8
}

func parseANSILoveFrame(t *testing.T, frame string) ([]ansiLovePreviewCell, int, int) {
	t.Helper()
	fg, bg := ansiart.VGA[7], ansiart.VGA[0]
	cells := []ansiLovePreviewCell{}
	width, lineWidth, height := 0, 0, 1
	for len(frame) > 0 {
		if frame[0] == '\n' {
			if width == 0 {
				width = lineWidth
			}
			if lineWidth != width {
				t.Fatal("unequal frame rows")
			}
			lineWidth = 0
			height++
			frame = frame[1:]
			continue
		}
		if strings.HasPrefix(frame, "\x1b[") {
			end := strings.IndexByte(frame, 'm')
			if end < 0 {
				t.Fatal("unterminated SGR")
			}
			var p []int
			for _, s := range strings.Split(frame[2:end], ";") {
				v, err := strconv.Atoi(s)
				if err != nil {
					t.Fatal(err)
				}
				p = append(p, v)
			}
			frame = frame[end+1:]
			switch {
			case len(p) == 1 && p[0] == 0:
				fg, bg = ansiart.VGA[7], ansiart.VGA[0]
			case len(p) == 1 && p[0] >= 30 && p[0] <= 37:
				fg = ansiart.VGA[p[0]-30]
			case len(p) == 1 && p[0] >= 40 && p[0] <= 47:
				bg = ansiart.VGA[p[0]-40]
			case len(p) == 1 && p[0] >= 90 && p[0] <= 97:
				fg = ansiart.VGA[p[0]-90+8]
			case len(p) == 1 && p[0] >= 100 && p[0] <= 107:
				bg = ansiart.VGA[p[0]-100+8]
			case len(p) == 5 && (p[0] == 38 || p[0] == 48) && p[1] == 2:
				c := [3]uint8{uint8(p[2]), uint8(p[3]), uint8(p[4])}
				if p[0] == 38 {
					fg = c
				} else {
					bg = c
				}
			case len(p) == 3 && (p[0] == 38 || p[0] == 48) && p[1] == 5:
				c := ansiLoveXtermColor(p[2])
				if p[0] == 38 {
					fg = c
				} else {
					bg = c
				}
			default:
				t.Fatalf("unexpected SGR %v", p)
			}
			continue
		}
		r, n := utf8.DecodeRuneInString(frame)
		if r == utf8.RuneError || r < 32 {
			t.Fatalf("invalid rendered rune %U", r)
		}
		frame = frame[n:]
		cells = append(cells, ansiLovePreviewCell{r, fg, bg})
		lineWidth++
	}
	if width == 0 {
		width = lineWidth
	}
	if lineWidth != width {
		t.Fatal("unequal final row")
	}
	return cells, width, height
}
func ansiLoveXtermColor(i int) [3]uint8 {
	if i < 16 {
		return ansiart.VGA[i]
	}
	if i >= 232 {
		v := uint8(8 + 10*(i-232))
		return [3]uint8{v, v, v}
	}
	levels := [6]uint8{0, 95, 135, 175, 215, 255}
	i -= 16
	return [3]uint8{levels[i/36], levels[(i/6)%6], levels[i%6]}
}

// Preview geometry is deterministic, not a promise about an installed terminal
// font. Blocks fill their subcells; braille dots occupy 2x2 pixels in each 4x4
// tile. ASCII uses the pinned VGA font, including the unscaled credit.
func ansiLovePreviewCoverage(t *testing.T, ch rune, font []byte) [128]bool {
	t.Helper()
	var bits [128]bool
	if ch >= ' ' && ch < 127 {
		for y := 0; y < 16; y++ {
			for x := 0; x < 8; x++ {
				bits[y*8+x] = font[int(ch)*16+y]&(0x80>>x) != 0
			}
		}
		return bits
	}
	for mask, r := range ansiart.QuadRunes {
		if ch == r {
			for y := 0; y < 16; y++ {
				for x := 0; x < 8; x++ {
					bits[y*8+x] = mask&(1<<(x/4+2*(y/8))) != 0
				}
			}
			return bits
		}
	}
	if ch == '░' || ch == '▒' || ch == '▓' {
		for y := 0; y < 16; y++ {
			for x := 0; x < 8; x++ {
				on := y%2 == 0 && x%2 == 0
				if ch == '▒' {
					on = (x+y)%2 == 0
				}
				if ch == '▓' {
					on = !on
				}
				bits[y*8+x] = on
			}
		}
		return bits
	}
	if ch >= 0x2800 && ch <= 0x28ff {
		for i, b := range [8]int{0, 3, 1, 4, 2, 5, 6, 7} {
			if (ch-0x2800)&(1<<b) != 0 {
				x, y := (i%2)*4+1, (i/2)*4+1
				for dy := 0; dy < 2; dy++ {
					for dx := 0; dx < 2; dx++ {
						bits[(y+dy)*8+x+dx] = true
					}
				}
			}
		}
		return bits
	}
	for rows := 3; rows <= 4; rows++ {
		for mask := 0; mask < (1 << (2 * rows)); mask++ {
			r := ansiart.SextantRune(uint8(mask))
			if rows == 4 {
				r = ansiart.OctantRune(uint8(mask))
			}
			if r != ch {
				continue
			}
			for y := 0; y < 16; y++ {
				for x := 0; x < 8; x++ {
					bits[y*8+x] = mask&(1<<(x/4+2*(y*rows/16))) != 0
				}
			}
			return bits
		}
	}
	// Exact mode's CP437 cells are not part of the visual golden matrix.
	t.Fatalf("unsupported preview glyph %U", ch)
	return bits
}
func writeANSILovePNG(t *testing.T, path string, cells []ansiLovePreviewCell, w, h int, font []byte) {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, w*8, h*16))
	coverage := map[rune][128]bool{}
	for i, c := range cells {
		bits, ok := coverage[c.ch]
		if !ok {
			bits = ansiLovePreviewCoverage(t, c.ch, font)
			coverage[c.ch] = bits
		}
		for j, ink := range bits {
			rgb := c.bg
			if ink {
				rgb = c.fg
			}
			img.SetNRGBA((i%w)*8+j%8, (i/w)*16+j/8, color.NRGBA{rgb[0], rgb[1], rgb[2], 255})
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}
