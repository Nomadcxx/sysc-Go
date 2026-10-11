// artwork-bake converts local reference artwork into bounded ASCII cell loops.
package main

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	_ "image/jpeg"
	"io"
	"math"
	"os"
)

const ramp = " .:-=+*#%@"

func main() {
	input := flag.String("input", "", "local GIF or JPEG reference")
	output := flag.String("output", "", "output .cells.gz")
	wave := flag.Bool("wave", false, "make restrained Great Wave glyph motion")
	flag.Parse()
	if err := bake(*input, *output, *wave); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func bake(input, output string, wave bool) error {
	f, err := os.Open(input)
	if err != nil {
		return err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	if info.Size() > 16<<20 {
		return fmt.Errorf("source exceeds 16 MiB")
	}
	cfg, kind, err := image.DecodeConfig(f)
	if err != nil {
		return err
	}
	if cfg.Width < 1 || cfg.Height < 1 || cfg.Width > 65535 || cfg.Height > 65535 || cfg.Width*cfg.Height > 32<<20 {
		return fmt.Errorf("invalid source dimensions")
	}
	if _, err = f.Seek(0, 0); err != nil {
		return err
	}
	cols := 240
	rows := max(1, min(128, int(math.Round(float64(cols*cfg.Height)/float64(cfg.Width)*0.5))))
	var frames [][]byte
	var delays []uint16
	if kind == "gif" {
		// ponytail: DecodeAll is offline only, bounded to 128 MiB of indexed
		// pixels at 160 frames; stream frames if larger sources are commissioned.
		if cfg.Width*cfg.Height*160 > 128<<20 {
			return fmt.Errorf("GIF decoded pixel budget exceeded")
		}
		if err = checkGIFFrames(io.LimitReader(f, (16<<20)+1)); err != nil {
			return err
		}
		if _, err = f.Seek(0, 0); err != nil {
			return err
		}
		g, err := gif.DecodeAll(f)
		if err != nil {
			return err
		}
		if len(g.Image) < 1 || len(g.Image) > 160 {
			return fmt.Errorf("GIF frame count exceeds 160")
		}
		err = walkGIF(g, func(i int, im image.Image) error {
			d := g.Delay[i] * 10
			if d < 50 || d > 1000 {
				return fmt.Errorf("unsupported GIF delay %d ms", d)
			}
			frames = append(frames, asciiCells(im, cols, rows))
			delays = append(delays, uint16(d))
			return nil
		})
		if err != nil {
			return err
		}
	} else {
		im, _, err := image.Decode(f)
		if err != nil {
			return err
		}
		base := asciiCells(im, cols, rows)
		n := 1
		if wave {
			n = 24
		}
		for k := 0; k < n; k++ {
			frame := append([]byte(nil), base...)
			if wave {
				for y := 0; y < rows; y++ {
					for x := 0; x < cols; x++ {
						i := 2 * (y*cols + x)
						shade := bytes.IndexByte([]byte(ramp), base[i])
						water := base[i+1] == 1 && y > rows/2
						foam := base[i+1] == 0 && shade >= 8 && x < cols*3/5 && y < rows*3/5
						if (water || foam) && (x*7+y*11)%9 == 0 {
							v := int(math.Round(math.Sin(float64(k)*2*math.Pi/24 + float64(x+y)*0.2)))
							frame[i] = ramp[max(0, min(9, shade+v))]
						}
					}
				}
			}
			frames = append(frames, frame)
			delays = append(delays, 250)
		}
	}
	var raw bytes.Buffer
	raw.WriteString("SYSCART1")
	for _, v := range []int{cols, rows, len(frames), cfg.Width, cfg.Height} {
		_ = binary.Write(&raw, binary.LittleEndian, uint16(v))
	}
	for i, frame := range frames {
		_ = binary.Write(&raw, binary.LittleEndian, delays[i])
		raw.Write(frame)
	}
	if raw.Len() > 8<<20 {
		return fmt.Errorf("inflated bake exceeds 8 MiB")
	}
	var compressed bytes.Buffer
	z, err := gzip.NewWriterLevel(&compressed, gzip.BestCompression)
	if err != nil {
		return err
	}
	if _, err = z.Write(raw.Bytes()); err != nil {
		return err
	}
	if err = z.Close(); err != nil {
		return err
	}
	if compressed.Len() > 1<<20 {
		return fmt.Errorf("compressed bake exceeds 1 MiB")
	}
	if err = os.WriteFile(output, compressed.Bytes(), 0644); err != nil {
		return err
	}
	fmt.Printf("%s: %dx%d cells, %d frames, %d raw bytes, %d gzip bytes\n", output, cols, rows, len(frames), raw.Len(), compressed.Len())
	return nil
}

// checkGIFFrames scans block lengths without decoding pixels. DecodeAll alone
// checks frame count too late to enforce the indexed-pixel memory budget.
func checkGIFFrames(r io.Reader) error {
	var header [13]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return err
	}
	if string(header[:6]) != "GIF87a" && string(header[:6]) != "GIF89a" {
		return fmt.Errorf("invalid GIF header")
	}
	skip := func(n int64) error { _, err := io.CopyN(io.Discard, r, n); return err }
	if header[10]&0x80 != 0 {
		if err := skip(3 << ((header[10] & 7) + 1)); err != nil {
			return err
		}
	}
	var bytebuf [1]byte
	readByte := func() (byte, error) { _, err := io.ReadFull(r, bytebuf[:]); return bytebuf[0], err }
	frames := 0
	for {
		kind, err := readByte()
		if err != nil {
			return err
		}
		switch kind {
		case 0x3b:
			if frames == 0 {
				return fmt.Errorf("empty GIF")
			}
			return nil
		case 0x21:
			if err = skip(1); err != nil {
				return err
			}
		case 0x2c:
			frames++
			if frames > 160 {
				return fmt.Errorf("GIF frame count exceeds 160")
			}
			var descriptor [9]byte
			if _, err = io.ReadFull(r, descriptor[:]); err != nil {
				return err
			}
			if descriptor[8]&0x80 != 0 {
				if err = skip(3 << ((descriptor[8] & 7) + 1)); err != nil {
					return err
				}
			}
			if err = skip(1); err != nil {
				return err
			} // LZW minimum code size
		default:
			return fmt.Errorf("invalid GIF block %x", kind)
		}
		for {
			n, err := readByte()
			if err != nil {
				return err
			}
			if n == 0 {
				break
			}
			if err = skip(int64(n)); err != nil {
				return err
			}
		}
	}
}

func walkGIF(g *gif.GIF, visit func(int, image.Image) error) error {
	bounds := image.Rect(0, 0, g.Config.Width, g.Config.Height)
	canvas := image.NewRGBA(bounds)
	background := color.Color(color.RGBA{})
	if p, ok := g.Config.ColorModel.(color.Palette); ok && int(g.BackgroundIndex) < len(p) {
		background = p[g.BackgroundIndex]
	}
	draw.Draw(canvas, bounds, image.NewUniform(background), image.Point{}, draw.Src)
	for i, frame := range g.Image {
		disposal := byte(0)
		if i < len(g.Disposal) {
			disposal = g.Disposal[i]
		}
		var saved *image.RGBA
		if disposal == gif.DisposalPrevious {
			saved = image.NewRGBA(bounds)
			copy(saved.Pix, canvas.Pix)
		}
		draw.Draw(canvas, frame.Bounds(), frame, frame.Bounds().Min, draw.Over)
		if err := visit(i, canvas); err != nil {
			return err
		}
		switch disposal {
		case gif.DisposalBackground:
			draw.Draw(canvas, frame.Bounds(), image.NewUniform(background), image.Point{}, draw.Src)
		case gif.DisposalPrevious:
			canvas = saved
		}
	}
	return nil
}

func asciiCells(im image.Image, cols, rows int) []byte {
	b := im.Bounds()
	out := make([]byte, cols*rows*2)
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			x1, y1 := b.Min.X+x*b.Dx()/cols, b.Min.Y+y*b.Dy()/rows
			x2, y2 := max(x1+1, b.Min.X+(x+1)*b.Dx()/cols), max(y1+1, b.Min.Y+(y+1)*b.Dy()/rows)
			r, g, bl, n := 0.0, 0.0, 0.0, 0.0
			for yy := y1; yy < y2; yy++ {
				for xx := x1; xx < x2; xx++ {
					rr, gg, bb, _ := im.At(xx, yy).RGBA()
					r += float64(rr) / 257
					g += float64(gg) / 257
					bl += float64(bb) / 257
					n++
				}
			}
			r /= n
			g /= n
			bl /= n
			lum := (0.2126*r + 0.7152*g + 0.0722*bl) / 255
			shade := max(0, min(9, int(math.Round(math.Pow(lum, 0.65)*9))))
			role := byte(0)
			chroma := math.Max(r, math.Max(g, bl)) - math.Min(r, math.Min(g, bl))
			if chroma > 20 && !(lum > 0.65 && chroma < 70) {
				switch {
				case bl > r*1.1:
					role = 1
				case r > bl*1.25 && r >= g:
					role = 2
				case r > g*1.1 && bl > g*1.1:
					role = 3
				default:
					role = 1
				}
			}
			i := 2 * (y*cols + x)
			out[i] = ramp[shade]
			out[i+1] = role
		}
	}
	return out
}
