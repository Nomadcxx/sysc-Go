package ansiart

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestDecodeDeferredWrap(t *testing.T) {
	for _, in := range []string{"ABCD\r\nE", "ABCD\x1b[31mE"} {
		g := DecodeANSI([]byte(in), 4, false)
		if g.W != 4 || g.H != 2 || g.Cells[4].Ch != 'E' {
			t.Fatalf("wrap %q: %+v", in, g)
		}
	}
	g := DecodeANSI([]byte("\x1b[1;31;44m\xdb\x1b[0m \x1b[2CZ"), 8, false)
	if g.Cells[0] != (Cell{'█', 9, 4}) || g.Cells[1] != (Cell{' ', 7, 0}) || g.Cells[4].Ch != 'Z' {
		t.Fatalf("decode: %+v", g.Cells)
	}
	if DecodeANSI([]byte("\x1b[5;41mX"), 1, true).Cells[0].Bg != 9 {
		t.Fatal("iCE background")
	}
}

func TestOriginalSAUCEAndDimensions(t *testing.T) {
	b, err := os.ReadFile("../../artwork/ansi-love.ans")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(b)) != "5c5d75b205680a1098517ef36432ce36161415c2dd2f9552b9ede7cdc2330ce3" {
		t.Fatal("source changed")
	}
	body, s := ParseSAUCE(b)
	if s == nil || s.Title != "ansi love" || s.Author != "warpus" || s.Group != "Legacy Krew" || s.Date != "20191228" || s.TInfo1 != 80 || s.TInfo2 != 157 || s.TFlags != 18 || s.TInfoS != "IBM VGA" || len(s.Comments) != 2 {
		t.Fatalf("SAUCE: %+v", s)
	}
	g := DecodeANSI(body, 80, false)
	if g.W != 80 || g.H != 157 {
		t.Fatalf("dimensions %dx%d", g.W, g.H)
	}
	var header strings.Builder
	for _, c := range g.Cells[3*80 : 4*80] {
		header.WriteRune(c.Ch)
	}
	if !strings.Contains(header.String(), "#include <ansilove.h>") {
		t.Fatal(header.String())
	}
}

func TestDecodeBoundsAndControlStrings(t *testing.T) {
	for _, w := range []int{-1, 0, 1000000000} {
		g := DecodeANSI([]byte("\x1b[999999999999999999C X\x1b]0;SECRET\aY\x1bPSECRET\x1b\\Z"), w, false)
		if g.W < 1 || g.W > 1024 || g.H > 4096 || len(g.Cells) != g.W*g.H {
			t.Fatal("unbounded grid")
		}
		for _, c := range g.Cells {
			if c.Ch == 'S' || c.Ch == 27 {
				t.Fatal("control string leaked")
			}
		}
	}
	g := DecodeANSI([]byte(strings.Repeat("\n", 5000)+"X"), 80, false)
	if g.H > 4096 {
		t.Fatal("height unbounded")
	}
	for _, b := range [][]byte{nil, []byte("SAUCE00"), append([]byte("abc"), make([]byte, 128)...)} {
		body, s := ParseSAUCE(b)
		if s != nil || len(body) != len(b) {
			t.Fatal("false SAUCE")
		}
	}
}

func TestCursorForwardReachesDeferredWrapColumn(t *testing.T) {
	g := DecodeANSI([]byte("A\x1b[3CB"), 4, false)
	if g.H != 2 || g.Cells[4].Ch != 'B' || g.Cells[3].Ch != ' ' {
		t.Fatalf("cursor forward at right edge: %+v", g)
	}
}

func TestCursorForwardAfterFullRow(t *testing.T) {
	g := DecodeANSI([]byte("ABCD\x1b[31m\x1b[2CE"), 4, false)
	if g.H != 2 || g.Cells[6].Ch != 'E' {
		t.Fatalf("pending wrap before cursor movement: %+v", g)
	}
}
