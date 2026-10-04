package main

import "time"

// cliFramesPerSecond is the rate used to turn -duration seconds into frames.
const cliFramesPerSecond = 20

// cliFrameInterval is the sleep between animation frames.
// It has to agree with cliFramesPerSecond so -duration is wall-clock seconds.
// 20 frames × 50ms = 1s. runPrint previously slept 30ms and exited early.
const cliFrameInterval = 50 * time.Millisecond

// framesForDuration converts a -duration flag value into a frame budget.
// Zero or negative means run until interrupted.
func framesForDuration(seconds int) int {
	if seconds <= 0 {
		return 0
	}
	return seconds * cliFramesPerSecond
}
