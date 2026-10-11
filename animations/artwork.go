package animations

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"sync"
)

const artworkRamp = " .:-=+*#%@"

//go:embed artwork/great-wave.cells.gz
var greatWaveBake []byte

//go:embed artwork/cyberpunk-girl.cells.gz
var cyberpunkGirlBake []byte

//go:embed artwork/wasteland.cells.gz
var wastelandBake []byte

type artworkScene struct {
	cols, rows, sourceWidth, sourceHeight int
	delays                                []uint16
	frames                                [][]byte
}

var (
	greatWaveScene     = sync.OnceValues(func() (*artworkScene, error) { return decodeArtwork(greatWaveBake) })
	cyberpunkGirlScene = sync.OnceValues(func() (*artworkScene, error) { return decodeArtwork(cyberpunkGirlBake) })
	wastelandScene     = sync.OnceValues(func() (*artworkScene, error) { return decodeArtwork(wastelandBake) })
)

func decodeArtwork(bake []byte) (*artworkScene, error) {
	if len(bake) > 1<<20 {
		return nil, fmt.Errorf("artwork: compressed size exceeds 1 MiB")
	}
	z, err := gzip.NewReader(bytes.NewReader(bake))
	if err != nil {
		return nil, fmt.Errorf("artwork: %w", err)
	}
	defer z.Close()
	raw, err := io.ReadAll(io.LimitReader(z, (8<<20)+1))
	if err != nil {
		return nil, fmt.Errorf("artwork: %w", err)
	}
	if len(raw) > 8<<20 || len(raw) < 18 || string(raw[:8]) != "SYSCART1" {
		return nil, fmt.Errorf("artwork: invalid header or inflated size")
	}
	u := func(i int) int { return int(binary.LittleEndian.Uint16(raw[i : i+2])) }
	s := &artworkScene{cols: u(8), rows: u(10), sourceWidth: u(14), sourceHeight: u(16)}
	n := u(12)
	if s.cols < 1 || s.cols > 256 || s.rows < 1 || s.rows > 128 || n < 1 || n > 160 || s.sourceWidth < 1 || s.sourceHeight < 1 {
		return nil, fmt.Errorf("artwork: invalid dimensions")
	}
	size := s.cols * s.rows * 2
	if len(raw) != 18+n*(2+size) {
		return nil, fmt.Errorf("artwork: invalid frame length")
	}
	for pos := 18; pos < len(raw); pos += 2 + size {
		d := binary.LittleEndian.Uint16(raw[pos : pos+2])
		if d < 50 || d > 1000 {
			return nil, fmt.Errorf("artwork: invalid delay")
		}
		frame := raw[pos+2 : pos+2+size]
		for i := 0; i < len(frame); i += 2 {
			if !strings.ContainsRune(artworkRamp, rune(frame[i])) || frame[i+1] > 3 {
				return nil, fmt.Errorf("artwork: invalid cell")
			}
		}
		s.delays = append(s.delays, d)
		s.frames = append(s.frames, frame)
	}
	return s, nil
}

// ArtworkAnimation plays immutable baked ASCII cells using the host's Update calls.
// ponytail: nominal 50 ms ticks follow existing effects; blur slows playback.
// A future elapsed-time host API can preserve real-time speed at every cadence.
type ArtworkAnimation struct {
	scene                                *artworkScene
	width, height, cellWidth, cellHeight int
	frame, elapsed                       int
	credit                               string
	styles                               [4][10]string
	background                           string
	cached                               string
}

func NewGreatWaveEffect(width, height int, theme string) (*ArtworkAnimation, error) {
	s, err := greatWaveScene()
	return newArtwork(s, err, width, height, theme, "Hokusai / The Met")
}
func NewCyberpunkGirlEffect(width, height int, theme string) (*ArtworkAnimation, error) {
	s, err := cyberpunkGirlScene()
	return newArtwork(s, err, width, height, theme, "Valenberg.com")
}
func NewWastelandEffect(width, height int, theme string) (*ArtworkAnimation, error) {
	s, err := wastelandScene()
	return newArtwork(s, err, width, height, theme, "Valenberg.com")
}

func newArtwork(s *artworkScene, err error, width, height int, theme, credit string) (*ArtworkAnimation, error) {
	if err != nil {
		return nil, err
	}
	a := &ArtworkAnimation{scene: s, credit: credit, cellWidth: 1, cellHeight: 2}
	p := GetLogoPalette(theme)
	particle := GetParticlePalette(theme)
	// Existing palette roles: neutral highlight, cool rim, warm/primary accent,
	// secondary particle accent. Density carries luminance independently of hue.
	// Pick the nearest warm hue already present in the theme. In Kanagawa
	// the primary accent is green, which would turn the sunset into grass.
	warm := p[3]
	best := 10.0
	for _, c := range GetFireworksPalette(theme) {
		r, g, b := hexToRGB(c)
		peak := max(r, g, b)
		if peak-min(r, min(g, b)) < 40 {
			continue
		}
		red, green, blue := float64(r)/float64(peak), float64(g)/float64(peak), float64(b)/float64(peak)
		d := (red-1)*(red-1) + (green-0.5)*(green-0.5) + (blue-0.25)*(blue-0.25)
		if d < best {
			best = d
			warm = c
		}
	}
	colors := []string{p[4], p[5], warm, particle[0]}
	r, g, b := hexToRGB(p[6])
	a.background = fmt.Sprintf("\x1b[48;2;%d;%d;%dm", r, g, b)
	for role, c := range colors {
		r, g, b := hexToRGB(c)
		for shade := 0; shade < len(artworkRamp); shade++ {
			// Keep sparse glyphs visible on dark themes without lifting empty cells.
			v := 0.55 + 0.45*float64(shade)/9
			a.styles[role][shade] = fmt.Sprintf("\x1b[38;2;%d;%d;%dm", int(float64(r)*v), int(float64(g)*v), int(float64(b)*v))
		}
	}
	a.Resize(width, height)
	return a, nil
}

func (a *ArtworkAnimation) Update() {
	a.elapsed += 50
	for a.elapsed >= int(a.scene.delays[a.frame]) {
		a.elapsed -= int(a.scene.delays[a.frame])
		a.frame = (a.frame + 1) % len(a.scene.frames)
		a.cached = ""
	}
}
func (a *ArtworkAnimation) Reset() { a.frame = 0; a.elapsed = 0; a.cached = "" }
func (a *ArtworkAnimation) Resize(width, height int) {
	a.width = max(1, min(512, width))
	a.height = max(1, min(256, height))
	a.cached = ""
}

// SetCellAspect fits the scene using measured font cells; invalid input keeps
// the current ratio. Ordinary terminals default to one column per two pixels.
func (a *ArtworkAnimation) SetCellAspect(width, height int) {
	if width > 0 && width <= 1024 && height > 0 && height <= 1024 {
		a.cellWidth = width
		a.cellHeight = height
		a.cached = ""
	}
}

func (a *ArtworkAnimation) Render() string {
	if a.cached != "" {
		return a.cached
	}
	available := a.height
	if a.height >= 3 {
		available--
	}
	w := a.width
	aspect := float64(a.scene.sourceWidth) * float64(a.cellHeight) / (float64(a.scene.sourceHeight) * float64(a.cellWidth))
	h := available
	if float64(w)/aspect > float64(available) {
		h = available
		w = max(1, int(float64(h)*aspect))
		w = min(w, a.width)
	} else {
		h = max(1, int(float64(w)/aspect))
	}
	x0, y0 := (a.width-w)/2, (available-h)/2
	frame := a.scene.frames[a.frame]
	var out strings.Builder
	out.Grow(a.width * a.height * 4)
	out.WriteString(a.background)
	last := ""
	for y := 0; y < a.height; y++ {
		for x := 0; x < a.width; x++ {
			glyph := byte(' ')
			style := ""
			if x >= x0 && x < x0+w && y >= y0 && y < y0+h {
				// Area-average when shrinking; nearest cell when enlarging. This
				// retains thin subjects without inventing a runtime image renderer.
				sx, sy := (x-x0)*a.scene.cols/w, (y-y0)*a.scene.rows/h
				ex, ey := max(sx+1, (x-x0+1)*a.scene.cols/w), max(sy+1, (y-y0+1)*a.scene.rows/h)
				shade, n := 0, 0
				var weights [4]int
				for yy := sy; yy < ey; yy++ {
					for xx := sx; xx < ex; xx++ {
						i := 2 * (yy*a.scene.cols + xx)
						v := strings.IndexByte(artworkRamp, frame[i])
						shade += v
						n++
						weights[frame[i+1]] += v + 1
					}
				}
				shade = (shade + n/2) / n
				role := 0
				for r := 1; r < 4; r++ {
					if weights[r] > weights[role] {
						role = r
					}
				}
				glyph = artworkRamp[shade]
				style = a.styles[role][shade]
			} else if a.height >= 3 && y == a.height-1 && x >= a.width-min(a.width, len(a.credit)) {
				glyph = a.credit[x-(a.width-min(a.width, len(a.credit)))]
				style = a.styles[0][9]
			}
			// Blank cells need a stable foreground too: otherwise the previous
			// ink colour invalidates whole letterbox regions during a glitch.
			if glyph == ' ' {
				style = a.styles[0][0]
			}
			if style != last {
				out.WriteString(style)
				last = style
			}
			out.WriteByte(glyph)
		}
		if y < a.height-1 {
			out.WriteByte('\n')
		}
	}
	out.WriteString("\x1b[0m")
	a.cached = out.String()
	return a.cached
}
