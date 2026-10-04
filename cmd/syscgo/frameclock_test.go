package main

import (
	"os"
	"strings"
	"testing"
	"time"
)

// TestPrintFrameIntervalMatchesDuration fails on the old print clock:
// frames were counted at 20fps while runPrint slept 30ms, so -duration 10
// exited after 6s.
func TestPrintFrameIntervalMatchesDuration(t *testing.T) {
	const seconds = 10
	frames := framesForDuration(seconds)
	elapsed := time.Duration(frames) * cliFrameInterval
	want := time.Duration(seconds) * time.Second
	if elapsed != want {
		t.Fatalf("print -duration %d schedules %s (%d frames × %s); want %s",
			seconds, elapsed, frames, cliFrameInterval, want)
	}

	if framesForDuration(0) != 0 {
		t.Fatal("duration 0 must stay infinite (zero frame budget)")
	}

	body := funcSource(t, "main.go", "func runPrint(")
	if strings.Contains(body, "30 * time.Millisecond") {
		t.Fatal("runPrint still sleeps 30ms while -duration is counted at 20fps")
	}
	if !strings.Contains(body, "time.Sleep(cliFrameInterval)") {
		t.Fatal("runPrint must sleep cliFrameInterval so -duration matches wall time")
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
