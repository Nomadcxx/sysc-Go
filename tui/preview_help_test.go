package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestRunningPreviewHelpMatchesHandledKeys covers issue #89: the running
// preview help advertised arrow navigation that handleKeyPress discards.
func TestRunningPreviewHelpMatchesHandledKeys(t *testing.T) {
	m := NewModel()
	m.width = 120
	m.height = 40
	m.canvasHeight = 24

	idle := m.renderHelp()
	if !strings.Contains(idle, "↑/↓ Navigate options") || !strings.Contains(idle, "←/→ Change selector") {
		t.Fatalf("idle help = %q, want arrow navigation", idle)
	}

	started, startCmd := m.Update(key(tea.KeyEnter))
	running := started.(Model)
	if !running.animationRunning {
		t.Fatal("enter should start the preview")
	}
	if startCmd == nil {
		t.Fatal("starting the preview should schedule ticks")
	}

	help := running.renderHelp()
	if strings.Contains(help, "Navigate options") || strings.Contains(help, "Change selector") ||
		strings.Contains(help, "↑") || strings.Contains(help, "←") {
		t.Fatalf("running help promises keys that are ignored: %q", help)
	}
	if !strings.Contains(help, "ESC Stop animation") {
		t.Fatalf("running help = %q, want ESC Stop animation", help)
	}

	// The repro size is a terminal tall enough to show the help line.
	running.height = 36
	view := running.View()
	if !strings.Contains(view, "ESC Stop animation") {
		t.Fatalf("running view hid the stop hint: %q", view)
	}
	if strings.Contains(view, "Navigate options") || strings.Contains(view, "Change selector") {
		t.Fatalf("running view still advertises arrow navigation: %q", view)
	}

	startAnim := running.selectedAnimation
	startTheme := running.selectedTheme
	startFocus := running.focusedSelector
	for _, k := range []tea.KeyType{tea.KeyUp, tea.KeyDown, tea.KeyLeft, tea.KeyRight} {
		next, cmd := running.Update(key(k))
		if isQuit(cmd) {
			t.Fatalf("%s quit the running preview", k)
		}
		got := next.(Model)
		if !got.animationRunning || got.currentAnim == nil {
			t.Fatalf("%s stopped the preview", k)
		}
		if got.selectedAnimation != startAnim || got.selectedTheme != startTheme || got.focusedSelector != startFocus {
			t.Fatalf("%s changed selectors while the preview is running", k)
		}
	}

	// Q stays ignored on a running preview. Esc stops it; Ctrl+C quits.
	afterQ, cmd := running.Update(key(tea.KeyRunes, 'q'))
	if isQuit(cmd) || !afterQ.(Model).animationRunning {
		t.Fatal("q should stay ignored on a running preview")
	}

	stopped, cmd := running.Update(key(tea.KeyEsc))
	if isQuit(cmd) {
		t.Fatal("esc should stop the preview, not quit")
	}
	if stopped.(Model).animationRunning || stopped.(Model).currentAnim != nil {
		t.Fatal("esc should stop the running preview")
	}

	_, cmd = running.Update(key(tea.KeyCtrlC))
	if !isQuit(cmd) {
		t.Fatal("ctrl+c should quit a running preview")
	}
}
