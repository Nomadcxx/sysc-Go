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
	if !strings.Contains(body, "clock.wait()") {
		t.Fatal("runPrint must wait for the CLI frame clock")
	}

	clockBody := funcSource(t, "frameclock.go", "func newCLIFrameClock(")
	if !strings.Contains(clockBody, "framesForDuration(seconds)") {
		t.Fatal("the CLI clock must retain a nominal frame budget for staged effects")
	}
}

func TestCLIEffectsUseWallClockAndElapsedUpdates(t *testing.T) {
	printBody := funcSource(t, "main.go", "func runPrint(")
	if !strings.Contains(printBody, "for clock.running()") {
		t.Fatal("runPrint must stop from the wall-clock deadline")
	}
	if !strings.Contains(printBody, "clock.wait()") {
		t.Fatal("runPrint must wait through the shared frame clock")
	}

	animationBody := funcSource(t, "main.go", "func runAnimation(")
	if !strings.Contains(animationBody, "UpdateElapsed(clock.elapsed())") {
		t.Fatal("runAnimation must advance elapsed-time effects by actual time")
	}

	mainBody := funcSource(t, "main.go", "func main(")
	if !strings.Contains(mainBody, "newCLIFrameClock(*duration)") {
		t.Fatal("main must create a wall-clock deadline from -duration")
	}
}

func TestCLIFrameClockMeasuresElapsedTimeAndDeadline(t *testing.T) {
	start := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	clock := &cliFrameClock{
		lastFrameAt: start,
		deadline:    start.Add(time.Second),
	}

	if got := clock.elapsedAt(start.Add(125 * time.Millisecond)); got != 125*time.Millisecond {
		t.Fatalf("first elapsed interval = %s; want 125ms", got)
	}
	if got := clock.elapsedAt(start.Add(475 * time.Millisecond)); got != 350*time.Millisecond {
		t.Fatalf("delayed elapsed interval = %s; want 350ms", got)
	}
	if got := clock.elapsedAt(start.Add(400 * time.Millisecond)); got != 0 {
		t.Fatalf("out-of-order elapsed interval = %s; want 0", got)
	}
	if got := clock.elapsedAt(start.Add(500 * time.Millisecond)); got != 25*time.Millisecond {
		t.Fatalf("elapsed interval after an out-of-order tick = %s; want 25ms", got)
	}
	if !clock.runningAt(start.Add(999 * time.Millisecond)) {
		t.Fatal("clock stopped before its deadline")
	}
	if clock.runningAt(start.Add(time.Second)) {
		t.Fatal("clock remained active at its deadline")
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
