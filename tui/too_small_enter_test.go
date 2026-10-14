package tui

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestEnterWithTooSmallWarningDoesNotPanic sends a too-small window size and
// then Enter for every animation. Enter must be ignored while the warning is
// shown instead of building an engine with a non-positive canvas width.
func TestEnterWithTooSmallWarningDoesNotPanic(t *testing.T) {
	base := NewModel()
	for _, name := range base.animations {
		for _, w := range []int{1, 9, 10} {
			t.Run(fmt.Sprintf("%s-%d", name, w), func(t *testing.T) {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("Enter at width %d panicked for %q: %v", w, name, r)
					}
				}()

				m := base
				m.animations = []string{name}
				m.selectedAnimation = 0

				updated, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: 20})
				m = updated.(Model)
				updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
				m = updated.(Model)

				if m.animationRunning {
					t.Fatalf("%q started at width %d while the too-small warning was shown", name, w)
				}
			})
		}
	}
}
