package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	syscGoAnimations "github.com/Nomadcxx/sysc-Go/animations"
	tea "github.com/charmbracelet/bubbletea"
)

func TestAK47ArtworkSelection(t *testing.T) {
	m := selectAnimation(t, newTestModel(t), "ak-47")
	m.canvasHeight = m.height - mainViewChromeRows
	m.currentAnim = m.createAnimation()
	m.animationRunning = true
	for i := 0; i < 210; i++ {
		m.currentAnim.Update()
	}
	held := m.currentAnim.Render()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: m.width + 10, Height: m.height + 5})
	m2 := updated.(Model)
	updated, _ = m2.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
	if updated.(Model).currentAnim.Render() != held {
		t.Fatal("resize reset phase")
	}
	p := filepath.Join(t.TempDir(), "art.txt")
	if err := os.WriteFile(p, []byte("\x1b[31mX"), 0600); err != nil {
		t.Fatal(err)
	}
	m = selectAnimation(t, newTestModel(t), "ak-47-text")
	m.files, m.selectedFile = []string{p}, 0
	m, _ = m.startAnimation()
	if m.animationError == "" || !strings.Contains(m.renderCanvas(), "control characters") {
		t.Fatal("missing visible rejection")
	}
	if err := os.WriteFile(p, []byte("XX  XX\n XXXX"), 0600); err != nil {
		t.Fatal(err)
	}
	m, _ = m.startAnimation()
	if m.currentAnim == nil || m.animationError != "" {
		t.Fatal("valid selected file rejected")
	}
}

func TestL2A1SARIsSelectableAndMoves(t *testing.T) {
	m := selectAnimation(t, newTestModel(t), "l2a1-sar")
	m.canvasHeight = 24
	fx := m.createAnimation()
	if fx == nil {
		t.Fatalf("L2A1 effect was not constructed: %s", m.animationError)
	}
	frame := fx.Render()
	for i := 0; i < 20; i++ {
		fx.Update()
	}
	if fx.Render() == frame {
		t.Fatal("L2A1 effect did not move in the TUI")
	}
}

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
	for _, anim := range []string{"fire-text", "matrix-art", "rain-art", "pour", "print", "beam-text", "ring-text", "blackhole"} {
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

func TestAnimationNeedsFileMatchesLibraryMetadata(t *testing.T) {
	for _, id := range syscGoAnimations.GetEffectNames() {
		if got, want := animationNeedsFile(id), syscGoAnimations.IsTextBasedEffect(id); got != want {
			t.Errorf("animationNeedsFile(%q) = %v, library metadata says %v", id, got, want)
		}
	}
	for _, id := range []string{"", "blackhole-text", "justice-cross", "ak-47-rama"} {
		if animationNeedsFile(id) {
			t.Errorf("unknown or unfinished effect %q requires artwork", id)
		}
	}
}

func TestAnimationsMatchLibraryRegistry(t *testing.T) {
	want := map[string]bool{}
	for _, id := range syscGoAnimations.GetEffectNames() {
		want[id] = true
	}
	model := NewModel()
	got := map[string]bool{}
	for _, id := range model.animations {
		got[id] = true
	}
	if len(got) != len(model.animations) {
		t.Fatal("picker contains duplicate effects")
	}
	if len(got) != len(want) {
		t.Fatalf("picker has %d effects, sysc-Go registry has %d", len(got), len(want))
	}
	for id := range want {
		if !got[id] {
			t.Errorf("picker is missing sysc-Go effect %q", id)
		}
	}
	for id := range got {
		if !want[id] {
			t.Errorf("picker exposes non-registry effect %q", id)
		}
	}
	for _, id := range []string{"ak-47-rama", "justice-cross"} {
		if got[id] {
			t.Errorf("picker exposes unfinished effect %q", id)
		}
	}
	if selected := model.animations[model.selectedAnimation]; selected != "fire" {
		t.Errorf("default picker effect = %q, want fire", selected)
	}
}

func TestEveryLibraryEffectConstructsInPicker(t *testing.T) {
	art := filepath.Join(t.TempDir(), "SYSC.txt")
	if err := os.WriteFile(art, []byte("SYSC\n"), 0600); err != nil {
		t.Fatal(err)
	}

	for _, id := range syscGoAnimations.GetEffectNames() {
		t.Run(id, func(t *testing.T) {
			m := selectAnimation(t, newTestModel(t), id)
			m.canvasHeight = 24
			m.files, m.selectedFile = []string{art}, 0
			if got := m.createAnimation(); got == nil {
				t.Fatalf("registry effect has no TUI constructor (error: %s)", m.animationError)
			}
		})
	}
}

func TestArtworkSelector(t *testing.T) {
	for _, id := range []string{"great-wave", "wasteland", "ak-47"} {
		m := selectAnimation(t, newTestModel(t), id)
		m.canvasHeight = 24
		m = selectFile(t, m, "BIT Text Editor")
		fx := m.createAnimation()
		if fx == nil || m.bitEditorMode || m.editorMode {
			t.Fatalf("%s did not construct without a file", id)
		}
		before := fx.Render()
		for i := 0; i < 8; i++ {
			fx.Update()
		}
		if fx.Render() == before {
			t.Fatalf("%s has no motion", id)
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
