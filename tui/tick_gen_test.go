package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TestStaleTickIsIgnored covers Esc then Enter inside one tick: the first
// run's TickMsg is still in flight and must not re-arm a second tick loop.
func TestStaleTickIsIgnored(t *testing.T) {
	m := NewModel()
	m.width, m.height = 120, 40
	m.animations = []string{"fire"}
	m.selectedAnimation = 0

	started, firstCmd := m.startAnimation()
	if !started.animationRunning || firstCmd == nil {
		t.Fatal("animation did not start")
	}
	staleMsg := firstCmd() // the tick armed by the first run

	stopped, _ := started.Update(tea.KeyMsg{Type: tea.KeyEsc})
	restarted, secondCmd := stopped.(Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := restarted.(Model)

	if m2.tickGen == started.tickGen {
		t.Fatal("startAnimation did not bump tickGen")
	}
	if _, cmd := m2.Update(staleMsg); cmd != nil {
		t.Fatal("stale TickMsg re-armed the tick loop")
	}

	freshMsg := secondCmd()
	if _, cmd := m2.Update(freshMsg); cmd == nil {
		t.Fatal("fresh TickMsg did not re-arm the tick loop")
	}
}

type elapsedProbeAnimation struct {
	lastElapsed time.Duration
}

func (a *elapsedProbeAnimation) Update()                        {}
func (a *elapsedProbeAnimation) UpdateElapsed(dt time.Duration) { a.lastElapsed = dt }
func (a *elapsedProbeAnimation) Render() string                 { return "" }
func (a *elapsedProbeAnimation) Reset()                         {}

func TestTickUsesElapsedTimeSincePreviousUpdate(t *testing.T) {
	m := NewModel()
	m.width, m.height = 120, 40
	m.animations = []string{"fire"}
	m.selectedAnimation = 0
	m.durations = []string{"infinite"}
	m.selectedDuration = 0

	started, _ := m.startAnimation()
	probe := &elapsedProbeAnimation{}
	started.currentAnim = probe
	tickAt := time.Now().Add(100 * time.Millisecond)
	started.Update(TickMsg{gen: started.tickGen, t: tickAt})

	if probe.lastElapsed < 90*time.Millisecond || probe.lastElapsed > 110*time.Millisecond {
		t.Fatalf("UpdateElapsed received %s; want about 100ms from the tick timestamp", probe.lastElapsed)
	}
}

func TestTickStopsAnimationAtWallClockDuration(t *testing.T) {
	m := NewModel()
	m.width, m.height = 120, 40
	m.animations = []string{"fire"}
	m.selectedAnimation = 0
	m.durations = []string{"1ns"}
	m.selectedDuration = 0

	started, _ := m.startAnimation()
	updated, cmd := started.Update(TickMsg{gen: started.tickGen, t: time.Now()})
	got := updated.(Model)
	if got.animationRunning || got.currentAnim != nil || cmd != nil {
		t.Fatal("animation continued after its wall-clock duration expired")
	}
}
