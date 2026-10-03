package main

import "testing"

func TestAnimationUpdatesPerFrameFitsRevealIntoFirstNinetyPercent(t *testing.T) {
	if got := animationUpdatesPerFrame(100, 20); got != 6 {
		t.Fatalf("animationUpdatesPerFrame(100, 20) = %d, want 6", got)
	}
	if got := animationUpdatesPerFrame(10, 20); got != 1 {
		t.Fatalf("animationUpdatesPerFrame(10, 20) = %d, want 1", got)
	}
	if got := animationUpdatesPerFrame(100, 0); got != 1 {
		t.Fatalf("animationUpdatesPerFrame(100, 0) = %d, want 1 for an unlimited run", got)
	}
}
