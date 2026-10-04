package animations

import (
	"fmt"
	"strings"
	"testing"
)

func TestHexToRGBPreservesScannerOutput(t *testing.T) {
	for _, value := range []string{"#000000", "#ffffff", "ABCDEF", "ff0000", "123456", "#12345g", "12zz56", " 12345", "+12345", "#bad", ""} {
		var r, g, b int
		raw := strings.TrimPrefix(value, "#")
		if len(raw) == 6 {
			fmt.Sscanf(raw, "%02x%02x%02x", &r, &g, &b)
		}
		ar, ag, ab := hexToRGB(value)
		if ar != r || ag != g || ab != b {
			t.Fatalf("%q: got %d,%d,%d want %d,%d,%d", value, ar, ag, ab, r, g, b)
		}
	}
}

func TestHexToRGBValidColourDoesNotAllocate(t *testing.T) {
	if allocs := testing.AllocsPerRun(100, func() { hexToRGB("#aBcDeF") }); allocs != 0 {
		t.Fatalf("valid palette colour allocated %.0f objects", allocs)
	}
}
