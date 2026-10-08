package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// selectAnimation points the model at an animation by name.
func selectAnimation(t *testing.T, m Model, name string) Model {
	t.Helper()
	for i, a := range m.animations {
		if a == name {
			m.selectedAnimation = i
			return m
		}
	}
	t.Fatalf("animation %q not in selector", name)
	return m
}

func newTestModel(t *testing.T) Model {
	t.Helper()
	m := NewModel()
	m.width, m.height = 120, 40
	return m
}

// TestEnterStartsNonFileAnimationDespiteEditorFileSelection reproduces issue
// #115: "BIT Text Editor"/"Custom text" are File entries and sat at index 0-1, so
// createAnimation saw an editor before it ever looked at the animation type and
// Enter reopened the editor forever instead of starting the preview.
func TestEnterStartsNonFileAnimationDespiteEditorFileSelection(t *testing.T) {
	for _, fileName := range []string{"BIT Text Editor", "Custom text"} {
		t.Run(fileName, func(t *testing.T) {
			m := newTestModel(t)

			// Open the editor from a text animation, then leave it with Esc.
			m = selectAnimation(t, m, "fire-text")
			m = selectFile(t, m, fileName)
			entered, _ := m.Update(key(tea.KeyEnter))
			m = entered.(Model)
			if !m.bitEditorMode && !m.editorMode {
				t.Fatalf("%s should open from a text animation", fileName)
			}
			left, _ := m.Update(key(tea.KeyEsc))
			m = left.(Model)
			if m.bitEditorMode || m.editorMode {
				t.Fatalf("esc should leave %s", fileName)
			}

			// The File selector still points at the editor. Enter must start fire.
			m = selectAnimation(t, m, "fire")
			started, _ := m.Update(key(tea.KeyEnter))
			m = started.(Model)
			if m.bitEditorMode || m.editorMode {
				t.Fatalf("enter on non-file animation reopened %s", fileName)
			}
			if !m.animationRunning || m.currentAnim == nil {
				t.Fatal("enter on non-file animation should start the preview")
			}
		})
	}
}

// TestCtrlBStillOpensEditorFromNonFileAnimation: the editors are File entries,
// so gating them on a text animation must not remove the way in. Ctrl+B is
// animation-independent and is the documented shortcut.
func TestCtrlBStillOpensEditorFromNonFileAnimation(t *testing.T) {
	m := newTestModel(t)
	m = selectAnimation(t, m, "fire")
	opened, _ := m.Update(key(tea.KeyCtrlB))
	m = opened.(Model)
	if !m.bitEditorMode {
		t.Fatal("ctrl+b should open the BIT editor regardless of the animation")
	}
	if m.bitTextInput.Focused() == false {
		t.Fatal("ctrl+b should focus the BIT text input")
	}
}

// TestTextAnimationsStillOpenEditors guards the other direction: the gate must not
// break the editors for the animations that actually read a file.
func TestTextAnimationsStillOpenEditors(t *testing.T) {
	for _, anim := range []string{"fire-text", "matrix-art", "rain-art", "pour", "print", "beam-text", "ring-text", "blackhole-text"} {
		for _, fileName := range []string{"BIT Text Editor", "Custom text"} {
			t.Run(anim+"/"+fileName, func(t *testing.T) {
				m := newTestModel(t)
				m = selectAnimation(t, m, anim)
				m = selectFile(t, m, fileName)
				opened, _ := m.Update(key(tea.KeyEnter))
				m = opened.(Model)
				if !m.bitEditorMode && !m.editorMode {
					t.Fatalf("%s should still open %s", anim, fileName)
				}
				if m.animationRunning {
					t.Fatalf("%s should not start a preview while an editor opens", anim)
				}
			})
		}
	}
}

// TestGuidanceOmitsEditorWhenFileDisabled reproduces #115's second symptom: the
// guidance box advertised the editor while the File selector read (disabled).
func TestGuidanceOmitsEditorWhenFileDisabled(t *testing.T) {
	for _, anim := range []string{"fire", "matrix", "rain", "fireworks", "aquarium", "beams", "sonar", "logo-morph"} {
		for _, fileName := range []string{"BIT Text Editor", "Custom text"} {
			m := newTestModel(t)
			m = selectAnimation(t, m, anim)
			m = selectFile(t, m, fileName)
			g := m.renderGuidance()
			if strings.Contains(g, "BIT Editor") || strings.Contains(g, "Custom text editor") {
				t.Errorf("%s: guidance advertises %q while the File selector is (disabled): %q", anim, fileName, g)
			}
		}
	}
}

// TestGuidanceKeepsFileInfoForTextAnimations is the other direction.
func TestGuidanceKeepsFileInfoForTextAnimations(t *testing.T) {
	m := newTestModel(t)
	m = selectAnimation(t, m, "fire-text")
	m = selectFile(t, m, "BIT Text Editor")
	if g := m.renderGuidance(); !strings.Contains(g, "BIT Editor") {
		t.Errorf("text animation guidance should mention the BIT editor: %q", g)
	}
}

// TestAnimationNeedsFile pins the single source of truth. fireworks must NOT be
// here: animfactory.go builds it with no file, and passing -file to the CLI
// silenced the effect.
func TestAnimationNeedsFile(t *testing.T) {
	want := map[string]bool{
		"fire-text": true, "matrix-art": true, "rain-art": true,
		"pour": true, "print": true, "beam-text": true,
		"ring-text": true, "blackhole-text": true,
		"fire": false, "matrix": false, "rain": false, "beams": false,
		"fireworks": false, "aquarium": false, "sonar": false,
		"sysc-logo": false, "cross-logo": false, "justice-cross": false,
		"logo-morph": false, "": false,
	}
	for anim, expected := range want {
		if got := animationNeedsFile(anim); got != expected {
			t.Errorf("animationNeedsFile(%q) = %v, want %v", anim, got, expected)
		}
	}

	// Every animation in the selector must be covered by the table, so a new
	// effect cannot silently land on the wrong side of the gate.
	for _, a := range NewModel().animations {
		if _, ok := want[a]; !ok {
			t.Errorf("animation %q missing from the animationNeedsFile table", a)
		}
	}
}

// TestRenderSelectorShowsDisabledFileKeepsItsPlace: the File selector must still
// display (disabled) for non-file animations, and the real value for text ones.
func TestRenderSelectorShowsDisabledFileForNonFileAnimations(t *testing.T) {
	m := newTestModel(t)
	m = selectAnimation(t, m, "fire")
	if !strings.Contains(m.renderSelector(2, "File", m.files[m.selectedFile]), "(disabled)") {
		t.Error("non-file animation should render the File selector as (disabled)")
	}

	m = selectAnimation(t, m, "fire-text")
	if strings.Contains(m.renderSelector(2, "File", m.files[m.selectedFile]), "(disabled)") {
		t.Error("text animation should render a live File selector")
	}
}

// TestWindowResizeDoesNotOpenEditor: a resize must never flip into an editor
// mode, which is what made #115 feel random.
func TestWindowResizeDoesNotOpenEditor(t *testing.T) {
	for _, anim := range []string{"fire", "fireworks", "matrix", "aquarium", "logo-morph"} {
		m := newTestModel(t)
		m = selectAnimation(t, m, anim)
		resized, _ := m.Update(tea.WindowSizeMsg{Width: 200, Height: 60})
		m = resized.(Model)
		if m.bitEditorMode || m.editorMode {
			t.Errorf("resize on %s opened an editor", anim)
		}
	}
}
