// Command bake-ansilove regenerates the embedded palette and semantic planes.
package main

import (
	"crypto/sha256"
	_ "embed"
	"flag"
	"fmt"
	"os"

	"github.com/Nomadcxx/sysc-Go/animations/internal/ansiart"
)

// IBM VGA 8x16 bytes from libansilove e22973d569c97c1a521a5246e057d300e865bd9f.
// See FONT-LICENSE and ../../artwork/README.md for provenance.
//
//go:embed ibm-vga8.bin
var fontBytes []byte

const sourceSHA256 = "5c5d75b205680a1098517ef36432ce36161415c2dd2f9552b9ede7cdc2330ce3"

func bake(source []byte) ([]byte, error) {
	if fmt.Sprintf("%x", sha256.Sum256(source)) != sourceSHA256 {
		return nil, fmt.Errorf("source SHA-256 differs from reviewed original")
	}
	body, s := ansiart.ParseSAUCE(source)
	if s == nil || s.TInfo1 != 80 || s.TInfo2 != 157 || s.DataType != 1 || s.FileType != 1 {
		return nil, fmt.Errorf("unexpected SAUCE metadata")
	}
	if len(fontBytes) != 4096 {
		return nil, fmt.Errorf("invalid VGA font length")
	}
	var font [256][16]uint8
	for i := range font {
		copy(font[i][:], fontBytes[i*16:(i+1)*16])
	}
	grid := ansiart.DecodeANSI(body, int(s.TInfo1), s.TFlags&1 != 0)
	if grid.H != int(s.TInfo2) {
		return nil, fmt.Errorf("decoded height %d differs from SAUCE", grid.H)
	}
	return ansiart.PackRaster(ansiart.Rasterize(grid, &font))
}

func run() error {
	in := flag.String("in", "animations/artwork/ansi-love.ans", "reviewed original ANSI source")
	out := flag.String("out", "animations/artwork/ansi-love.raster.zlib", "output compressed raster")
	flag.Parse()
	source, err := os.ReadFile(*in)
	if err != nil {
		return err
	}
	b, err := bake(source)
	if err != nil {
		return err
	}
	if err := os.WriteFile(*out, b, 0644); err != nil {
		return err
	}
	fmt.Printf("wrote %s (%d bytes)\n", *out, len(b))
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
