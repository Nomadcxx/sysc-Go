package main

import (
	"os"
	"strings"
	"testing"
	"time"
)

// TestPrintFrameIntervalMatchesDuration keeps the print frame budget aligned
// with the CLI ticker so -duration stays wall-clock time.
func TestPrintFrameIntervalMatchesDuration(t *testing.T) {
	const seconds = 10
	frames := framesForDuration(seconds)
	elapsed := time.Duration(frames) * cliFrameInterval
	want := time.Duration(seconds) * time.Second
	if delta := want - elapsed; delta < 0 || delta >= time.Duration(frames) {
		t.Fatalf("print -duration %d schedules %s (%d frames × %s); want %s",
			seconds, elapsed, frames, cliFrameInterval, want)
	}

	if framesForDuration(0) != 0 {
		t.Fatal("duration 0 must stay infinite (zero frame budget)")
	}

	body := funcSource(t, "main.go", "func runPrint(")
	if !strings.Contains(body, "waitForNextCLIFrame()") {
		t.Fatal("runPrint must wait for the CLI frame ticker")
	}

	mainBody := funcSource(t, "main.go", "func main(")
	if !strings.Contains(mainBody, "framesForDuration(") {
		t.Fatal("main must budget frames with framesForDuration")
	}
}

func funcSource(t *testing.T, file, sig string) string {
	t.Helper()
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read %s: %v", file, err)
	}
	src := string(data)
	start := strings.Index(src, sig)
	if start < 0 {
		t.Fatalf("%s not found in %s", sig, file)
	}
	rest := src[start+len(sig):]
	next := strings.Index(rest, "\nfunc ")
	if next < 0 {
		return src[start:]
	}
	return src[start : start+len(sig)+next]
}
