package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"testing"

	"github.com/Nomadcxx/sysc-Go/animations/internal/ansiart"
)

func TestBakeSourceAndReproducibility(t *testing.T) {
	source, err := os.ReadFile("../../artwork/ansi-love.ans")
	if err != nil {
		t.Fatal(err)
	}
	b, err := bake(source)
	if err != nil {
		t.Fatal(err)
	}
	r, err := ansiart.UnpackRaster(b)
	if err != nil || r.W != 640 || r.H != 2512 {
		t.Fatal("dimensions", err)
	}

	// Full RGB digest measured from pinned libansilove's canonical 8px output.
	palette := [16][3]byte{{0, 0, 0}, {170, 0, 0}, {0, 170, 0}, {170, 85, 0}, {0, 0, 170}, {170, 0, 170}, {0, 170, 170}, {170, 170, 170}, {85, 85, 85}, {255, 85, 85}, {85, 255, 85}, {255, 255, 85}, {85, 85, 255}, {255, 85, 255}, {85, 255, 255}, {255, 255, 255}}
	rgb := make([]byte, 0, len(r.Idx)*3)
	for _, i := range r.Idx {
		rgb = append(rgb, palette[i][:]...)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(rgb)); got != "0de6e314dcdb1863ad7d49f91a3141c6677a4c655b2182c7dc3277776b61dcac" {
		t.Fatalf("canonical RGB differs: %s", got)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(fontBytes)); got != "ff3640ffffdd774ec56ee7ecacf7bce8324a67071155ade07f4c6380fa879025" {
		t.Fatalf("font digest differs: %s", got)
	}
	for _, p := range []struct {
		x, y int
		want byte
	}{{0, 0, 0}, {9, 51, 7}, {8, 51, 0}} {
		if r.Idx[p.y*r.W+p.x] != p.want {
			t.Fatalf("pixel %d,%d = %d", p.x, p.y, r.Idx[p.y*r.W+p.x])
		}
	}
	again, err := bake(source)
	if err != nil || !bytes.Equal(b, again) {
		t.Fatal("bake is nondeterministic", err)
	}
	checked, err := os.ReadFile("../../artwork/ansi-love.raster.zlib")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b, checked) {
		t.Fatal("committed bake differs; regenerate")
	}
	source[0] ^= 1
	if _, err := bake(source); err == nil {
		t.Fatal("accepted changed source")
	}
}
