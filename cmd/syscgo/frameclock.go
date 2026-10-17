package main

import "time"

// cliFramesPerSecond is the rate used to turn -duration seconds into frames.
const cliFramesPerSecond = 30

// cliFrameInterval is the time between animation frame ticks.
// It has to agree with cliFramesPerSecond so -duration is wall-clock seconds.
const cliFrameInterval = time.Second / cliFramesPerSecond

// ponytail: the CLI runs one effect at a time; a single ticker drops missed
// ticks instead of building a delayed frame backlog.
var cliFrameTicker *time.Ticker

func waitForNextCLIFrame() { <-cliFrameTicker.C }

// framesForDuration converts a -duration flag value into a frame budget.
// Zero or negative means run until interrupted.
func framesForDuration(seconds int) int {
	if seconds <= 0 {
		return 0
	}
	return seconds * cliFramesPerSecond
}
