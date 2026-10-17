package main

import "time"

// cliFramesPerSecond is the target refresh rate and the basis for nominal
// completion pacing in staged text effects.
const cliFramesPerSecond = 30

const cliFrameInterval = time.Second / cliFramesPerSecond

// ponytail: render one frame at a time and drop missed ticks; under slow output,
// frame-only effects advance more slowly and need UpdateElapsed support to track time.
type cliFrameClock struct {
	duration    time.Duration
	frameBudget int
	startedAt   time.Time
	lastFrameAt time.Time
	deadline    time.Time
	ticker      *time.Ticker
}

func newCLIFrameClock(seconds int) *cliFrameClock {
	return &cliFrameClock{
		duration:    durationForSeconds(seconds),
		frameBudget: framesForDuration(seconds),
	}
}

func durationForSeconds(seconds int) time.Duration {
	if seconds <= 0 {
		return 0
	}
	const maxDuration = time.Duration(1<<63 - 1)
	if int64(seconds) > int64(maxDuration/time.Second) {
		return maxDuration
	}
	return time.Duration(seconds) * time.Second
}

func framesForDuration(seconds int) int {
	if seconds <= 0 {
		return 0
	}
	maxInt := int(^uint(0) >> 1)
	if seconds > maxInt/cliFramesPerSecond {
		return maxInt
	}
	return seconds * cliFramesPerSecond
}

func (c *cliFrameClock) start() {
	if !c.startedAt.IsZero() {
		return
	}
	c.startedAt = time.Now()
	c.lastFrameAt = c.startedAt
	if c.duration > 0 {
		c.deadline = c.startedAt.Add(c.duration)
	}
	c.ticker = time.NewTicker(cliFrameInterval)
}

func (c *cliFrameClock) running() bool {
	c.start()
	return c.runningAt(time.Now())
}

func (c *cliFrameClock) runningAt(now time.Time) bool {
	return c.deadline.IsZero() || now.Before(c.deadline)
}

func (c *cliFrameClock) nominalFrames() int { return c.frameBudget }

func (c *cliFrameClock) elapsed() time.Duration {
	c.start()
	return c.elapsedAt(time.Now())
}

func (c *cliFrameClock) elapsedAt(now time.Time) time.Duration {
	if !now.After(c.lastFrameAt) {
		return 0
	}
	elapsed := now.Sub(c.lastFrameAt)
	c.lastFrameAt = now
	return elapsed
}

func (c *cliFrameClock) wait() {
	c.start()
	if c.deadline.IsZero() {
		<-c.ticker.C
		return
	}

	remaining := time.Until(c.deadline)
	if remaining <= 0 {
		return
	}
	timer := time.NewTimer(remaining)
	defer timer.Stop()
	select {
	case <-c.ticker.C:
	case <-timer.C:
	}
}

func (c *cliFrameClock) stop() {
	if c.ticker != nil {
		c.ticker.Stop()
	}
}
