package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type fixedFrameAnimation struct{ frame string }

func (a fixedFrameAnimation) Update()        {}
func (a fixedFrameAnimation) Render() string { return a.frame }
func (a fixedFrameAnimation) Reset()         {}

func TestRunningViewPreservesLayout(t *testing.T) {
	for _, size := range [][2]int{{100, 30}, {200, 55}} {
		m := NewModel()
		m.width, m.height = size[0], size[1]
		m.animationRunning = true
		m.currentAnim = fixedFrameAnimation{frame: "\033[31m" + strings.Repeat("x", size[0]-12) + "界\033[0m\n" + strings.Repeat(" ", size[0]-10)}

		got := ansi.Strip(m.View())
		want := ansi.Strip(legacyRunningView(m))
		if got != want {
			gotLines, wantLines := strings.Split(got, "\n"), strings.Split(want, "\n")
			for i := 0; i < min(len(gotLines), len(wantLines)); i++ {
				if gotLines[i] != wantLines[i] {
					t.Errorf("%dx%d line %d differs: got %q, want %q", size[0], size[1], i, gotLines[i], wantLines[i])
					break
				}
			}
			if len(gotLines) != len(wantLines) {
				t.Errorf("%dx%d has %d lines, want %d", size[0], size[1], len(gotLines), len(wantLines))
			}
		}
	}
}

func TestRunningViewAllocationBudget(t *testing.T) {
	m := NewModel()
	m.width, m.height = 200, 55
	m.animationRunning = true
	row := strings.Repeat("x", m.width-10)
	m.currentAnim = fixedFrameAnimation{frame: strings.Repeat(row+"\n", m.height-mainViewChromeRows-1) + row}

	allocs := testing.AllocsPerRun(10, func() { _ = m.View() })
	if allocs > 100 {
		t.Fatalf("running View allocated %.0f objects, want at most 100", allocs)
	}
}

func TestRunningCanvasTruncatesWideANSIFrames(t *testing.T) {
	m := NewModel()
	m.width, m.height = 100, 30
	m.animationRunning = true
	m.currentAnim = fixedFrameAnimation{frame: "\033[31m" + strings.Repeat("a", 88) + "界tail"}

	lines := strings.Split(m.renderCanvas(), "\n")
	if len(lines) != m.height-mainViewChromeRows+2 {
		t.Fatalf("canvas has %d lines, want %d", len(lines), m.height-mainViewChromeRows+2)
	}
	line := ansi.Strip(lines[1])
	if lipgloss.Width(lines[1]) != m.width-8 {
		t.Fatalf("truncated line has width %d, want %d", lipgloss.Width(lines[1]), m.width-8)
	}
	if !strings.Contains(line, strings.Repeat("a", 88)+"界") || strings.Contains(line, "tail") {
		t.Fatalf("wide frame was not truncated at a character boundary: %q", line)
	}
}

func legacyRunningView(m Model) string {
	content := lipgloss.JoinVertical(lipgloss.Left, legacyCanvas(m), m.renderSelectors(), m.renderGuidance(), m.renderHelp())
	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Top).
		Render(content)
}

func legacyCanvas(m Model) string {
	viewportHeight := max(1, m.height-mainViewChromeRows)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#88C0D0")).
		Width(m.width - 10).
		Height(viewportHeight).
		Render(m.currentAnim.Render())
}

func BenchmarkTUIViewRunning(b *testing.B) {
	for _, name := range []string{"fire", "fire-text", "wasteland", "great-wave", "ak-47", "ak-47-text", "matrix"} {
		b.Run(name+"/View", func(b *testing.B) {
			m := perfModel(b, name, 200, 55)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = m.View()
			}
		})
		b.Run(name+"/Render", func(b *testing.B) {
			m := perfModel(b, name, 200, 55)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = m.currentAnim.Render()
			}
		})
	}
}
