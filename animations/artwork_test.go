package animations

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestArtworkWarmPalette(t *testing.T) {
	a, err := NewWastelandEffect(80, 24, "kanagawa")
	if err != nil {
		t.Fatal(err)
	}
	var r, g, b int
	if _, err = fmt.Sscanf(a.styles[2][9], "\x1b[38;2;%d;%d;%dm", &r, &g, &b); err != nil {
		t.Fatal(err)
	}
	if r <= g || r <= b {
		t.Fatal("Kanagawa sunset lost its warm hue")
	}
	other, err := NewWastelandEffect(80, 24, "dracula")
	if err != nil {
		t.Fatal(err)
	}
	if a.Render() == other.Render() {
		t.Fatal("theme ignored")
	}
}

func TestArtwork(t *testing.T) {
	for _, id := range []string{"great-wave", "cyberpunk-girl", "wasteland"} {
		t.Run(id, func(t *testing.T) {
			meta := GetEffectMetadata(id)
			if meta == nil || meta.RequiresText {
				t.Fatalf("missing standalone metadata: %+v", meta)
			}
			for _, theme := range []string{"kanagawa", "dracula", "monochrome"} {
				var a *ArtworkAnimation
				var err error
				switch id {
				case "great-wave":
					a, err = NewGreatWaveEffect(80, 24, theme)
				case "cyberpunk-girl":
					a, err = NewCyberpunkGirlEffect(80, 24, theme)
				case "wasteland":
					a, err = NewWastelandEffect(80, 24, theme)
				}
				if err != nil {
					t.Fatal(err)
				}
				count, delay := 24, uint16(250)
				if id == "cyberpunk-girl" {
					count, delay = 12, 120
				}
				if id == "wasteland" {
					count, delay = 160, 90
				}
				if len(a.scene.frames) != count {
					t.Fatalf("lost source frames: %d, want %d", len(a.scene.frames), count)
				}
				for _, d := range a.scene.delays {
					if d != delay {
						t.Fatalf("source delay %d, want %d", d, delay)
					}
				}
				first := a.Render()
				if a.Render() != first {
					t.Fatal("Render advances animation")
				}
				rows := strings.Split(ansi.Strip(first), "\n")
				if len(rows) != 24 {
					t.Fatalf("rows = %d", len(rows))
				}
				ink := 0
				for _, row := range rows {
					if len(row) != 80 {
						t.Fatalf("columns = %d", len(row))
					}
					for _, c := range row {
						if c < 32 || c > 126 {
							t.Fatalf("non-ASCII glyph %q", c)
						}
						if c != ' ' {
							ink++
						}
					}
				}
				if ink < 200 {
					t.Fatalf("only %d ink cells", ink)
				}
				credit := "Valenberg.com"
				if id == "great-wave" {
					credit = "Hokusai"
				}
				if !strings.Contains(rows[23], credit) {
					t.Fatal("source credit missing")
				}
				period := 0
				for _, d := range a.scene.delays {
					period += int(d)
				}
				// 1.44 s is not divisible by 50 ms: five loops meet a tick
				// boundary. A truncated 100 ms delay must fail this check.
				cycle := period
				for cycle%50 != 0 {
					cycle += period
				}
				seen := map[int]bool{a.frame: true}
				changed := false
				for i := 0; i < cycle/50; i++ {
					a.Update()
					seen[a.frame] = true
					if a.Render() != first {
						changed = true
					}
				}
				if !changed || len(seen) != len(a.scene.delays) || a.Render() != first {
					t.Fatalf("incomplete loop: %d/%d frames, changed %v", len(seen), len(a.scene.delays), changed)
				}
				a.Update()
				a.Reset()
				if a.Render() != first {
					t.Fatal("reset did not restore first frame")
				}
				a.SetCellAspect(7, 16)
				a.Resize(491, 90)
				if len(a.Render()) > 4<<20 {
					t.Fatal("wallpaper frame too large")
				}
				a.Resize(int(^uint(0)>>1), int(^uint(0)>>1))
				if a.width != 512 || a.height != 256 || len(a.Render()) > 4<<20 {
					t.Fatal("size cap failed")
				}
				a.Resize(-1, 0)
				if a.width != 1 || a.height != 1 {
					t.Fatal("size floor failed")
				}
			}
		})
	}
}

func TestArtworkRejectsInvalidBake(t *testing.T) {
	for _, raw := range [][]byte{nil, []byte("SYSCART1"), append([]byte("SYSCART1"), make([]byte, 10)...), bytes.Repeat([]byte{'x'}, (8<<20)+1)} {
		var b bytes.Buffer
		z := gzip.NewWriter(&b)
		_, _ = z.Write(raw)
		_ = z.Close()
		if _, err := decodeArtwork(b.Bytes()); err == nil {
			t.Fatal("invalid bake accepted")
		}
	}
	z, err := gzip.NewReader(bytes.NewReader(cyberpunkGirlBake))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(z)
	_ = z.Close()
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range []struct {
		offset int
		value  byte
	}{{8, 0}, {12, 161}, {14, 0}, {18, 0}, {20, 0x1b}, {21, 4}} {
		bad := append([]byte(nil), raw...)
		bad[change.offset] = change.value
		if change.offset == 14 {
			bad[15] = 0
		}
		var b bytes.Buffer
		w := gzip.NewWriter(&b)
		_, _ = w.Write(bad)
		_ = w.Close()
		if _, err := decodeArtwork(b.Bytes()); err == nil {
			t.Fatalf("invalid byte at offset %d accepted", change.offset)
		}
	}
}
