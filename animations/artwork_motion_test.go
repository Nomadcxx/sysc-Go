package animations

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
)

func TestArtworkPhysicalAspect(t *testing.T) {
	for _, id := range []string{"great-wave", "wasteland"} {
		for _, size := range [][2]int{{491, 90}, {365, 90}, {80, 24}} {
			var a *ArtworkAnimation
			if id == "great-wave" {
				a, _ = NewGreatWaveEffect(size[0], size[1], "kanagawa")
			} else {
				a, _ = NewWastelandEffect(size[0], size[1], "kanagawa")
			}
			a.SetCellAspect(7, 16)
			a.Render()
			xRate, yRate := a.sx*7/2, a.sy*16/float64(a.dotsY)
			if math.Abs(xRate-yRate) > .001 {
				t.Fatalf("%s at %v stretches physical shapes: %f/%f", id, size, xRate, yRate)
			}
			for _, col := range []int{0, a.dotW - 1} {
				if a.pixels[(a.dotH-2)*a.dotW+col] == artBG {
					t.Fatalf("%s at %v leaves extended scene empty", id, size)
				}
			}
		}
	}
}

func TestArtworkMotion(t *testing.T) {
	for _, id := range []string{"great-wave", "cyberpunk-girl", "wasteland"} {
		t.Run(id, func(t *testing.T) {
			var a *ArtworkAnimation
			var err error
			switch id {
			case "great-wave":
				a, err = NewGreatWaveEffect(120, 40, "kanagawa")
			case "cyberpunk-girl":
				a, err = NewCyberpunkGirlEffect(120, 40, "dracula")
			case "wasteland":
				a, err = NewWastelandEffect(120, 40, "kanagawa")
			}
			if err != nil {
				t.Fatal(err)
			}
			a.SetCellAspect(7, 16)
			first := a.Render()
			if a.Render() != first {
				t.Fatal("Render advanced time")
			}
			plain := ansi.Strip(first)
			rows := strings.Split(plain, "\n")
			if len(rows) != 40 {
				t.Fatalf("height = %d", len(rows))
			}
			ink := 0
			for _, row := range rows {
				if len([]rune(row)) != 120 {
					t.Fatal("wrong cell width")
				}
				for _, ch := range row {
					if id == "great-wave" && ch >= 0x2801 && ch <= 0x28ff {
						ink++
					}
					if id != "great-wave" && ch >= 0x2580 && ch <= 0x259f {
						ink++
					}
				}
			}
			if ink < 200 {
				t.Fatalf("only %d authored character cells", ink)
			}
			seen := map[string]bool{}
			for i := 0; i < 240; i++ {
				a.Update()
				seen[a.Render()] = true
			}
			if len(seen) < 150 {
				t.Fatalf("only %d distinct motion steps", len(seen))
			}
			a.Reset()
			if a.Render() != first {
				t.Fatal("reset changed initial frame")
			}
			a.Resize(491, 90)
			if len(a.Render()) > 4<<20 {
				t.Fatal("frame exceeds bound")
			}
			a.Resize(int(^uint(0)>>1), int(^uint(0)>>1))
			if a.width != 512 || a.height != 256 || len(a.Render()) > 4<<20 {
				t.Fatal("size cap failed")
			}
			a.Resize(-1, 0)
			if a.width != 1 || a.height != 1 {
				t.Fatal("size floor failed")
			}
		})
	}
}

func TestArtworkDetails(t *testing.T) {
	girl, _ := NewCyberpunkGirlEffect(491, 90, "dracula")
	girl.SetCellAspect(7, 16)
	girl.Render()
	countBlood := func() int {
		n := 0
		for y := 0; y < girl.dotH; y++ {
			for x := 0; x < girl.dotW; x++ {
				worldY := (float64(y) - girl.oy) / girl.sy
				if worldY > 266 && girl.pixels[y*girl.dotW+x] == artBlood {
					n++
				}
			}
		}
		return n
	}
	if countBlood() != 0 {
		t.Fatal("blood drop already detached in first frame")
	}
	for i := 0; i < 10; i++ {
		girl.UpdateElapsed(250 * time.Millisecond)
	}
	girl.Render()
	if countBlood() == 0 {
		t.Fatal("nose blood never detaches and falls below the mouth")
	}

	for _, id := range []string{"great-wave", "wasteland"} {
		var a *ArtworkAnimation
		if id == "great-wave" {
			a, _ = NewGreatWaveEffect(240, 60, "kanagawa")
		} else {
			a, _ = NewWastelandEffect(240, 60, "kanagawa")
		}
		a.Render()
		before := append([]uint8(nil), a.pixels...)
		for i := 0; i < 16; i++ {
			a.UpdateElapsed(250 * time.Millisecond)
		}
		a.Render()
		changed := 0
		for y := 0; y < a.dotH; y++ {
			for x := 0; x < a.dotW; x++ {
				wx, wy := (float64(x)-a.ox)/a.sx, (float64(y)-a.oy)/a.sy
				inside := wx > 250 && wx < 560 && wy > 50 && wy < 390
				if id == "wasteland" {
					inside = wx > 330 && wx < 650 && wy > 307 && wy < 323
				}
				if inside && before[y*a.dotW+x] != a.pixels[y*a.dotW+x] {
					changed++
				}
			}
		}
		if changed < 5 {
			t.Fatalf("%s critical region has only %d changed samples", id, changed)
		}
	}
}

func TestArtworkElapsedLoop(t *testing.T) {
	for _, theme := range []string{"kanagawa", "dracula", "monochrome"} {
		for _, id := range []string{"great-wave", "cyberpunk-girl", "wasteland"} {
			var a *ArtworkAnimation
			switch id {
			case "great-wave":
				a, _ = NewGreatWaveEffect(80, 24, theme)
			case "cyberpunk-girl":
				a, _ = NewCyberpunkGirlEffect(80, 24, theme)
			case "wasteland":
				a, _ = NewWastelandEffect(80, 24, theme)
			}
			first := a.Render()
			for elapsed := time.Duration(0); elapsed < a.period; elapsed += 50 * time.Millisecond {
				a.Update()
			}
			if a.Render() != first {
				t.Fatalf("%s/%s loop does not close", id, theme)
			}
			before := a.elapsed
			a.UpdateElapsed(-time.Second)
			if a.elapsed != before {
				t.Fatal("negative time advanced phase")
			}
			a.SetCellAspect(7, 16)
			a.Render()
			phase := a.elapsed
			a.Resize(365, 90)
			a.Render()
			if a.elapsed != phase {
				t.Fatal("resize advanced phase")
			}
		}
	}
}

func TestArtworkVehicleLoop(t *testing.T) {
	a, _ := NewWastelandEffect(491, 90, "kanagawa")
	a.SetCellAspect(7, 16)
	centroid := func() float64 {
		a.Render()
		row := int(a.oy + 314*a.sy)
		sum, count := 0.0, 0
		for col := 0; col < a.dotW; col++ {
			x := (float64(col) + .5 - a.ox) / a.sx
			if x > 345 && x < 645 && a.pixels[row*a.dotW+col] == artInk {
				sum += x
				count++
			}
		}
		if count == 0 {
			t.Fatal("vehicle vanished")
		}
		return sum / float64(count)
	}
	first := centroid()
	a.elapsed = 16*time.Second - 50*time.Millisecond
	a.cached = ""
	if delta := math.Abs(first - centroid()); delta > 5 {
		t.Fatalf("vehicle jumps %f world units at loop boundary", delta)
	}
}

func TestArtworkStrokeBounded(t *testing.T) {
	a, _ := NewWastelandEffect(16, 16, "kanagawa")
	a.view(32, 32, false)
	a.line(artPoint{-1e12, 16.5}, artPoint{1e12, 16.5}, .1, artLight)
	for row := 0; row < a.dotH; row++ {
		for col := 0; col < a.dotW; col++ {
			want := uint8(artBG)
			if row == 16 {
				want = artLight
			}
			if a.pixels[row*a.dotW+col] != want {
				t.Fatalf("thin clipped stroke at %d,%d = %d; want %d", col, row, a.pixels[row*a.dotW+col], want)
			}
		}
	}
}

func TestArtworkExtendedForeground(t *testing.T) {
	a, _ := NewWastelandEffect(491, 90, "kanagawa")
	a.SetCellAspect(7, 16)
	a.Render()
	details := 0
	for row := 0; row < a.dotH; row++ {
		for col := 0; col < a.dotW; col++ {
			x := (float64(col) + .5 - a.ox) / a.sx
			if x > 980 && a.pixels[row*a.dotW+col] != artInk && a.pixels[row*a.dotW+col] != artBG {
				details++
			}
		}
	}
	if details < 200 {
		t.Fatalf("ultrawide wreckage extension has only %d detail samples", details)
	}
}
