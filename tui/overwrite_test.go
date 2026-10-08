package tui

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestExportRefusesToClobber(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	wallsDir := filepath.Join(tmpHome, ".local", "share", "syscgo", "walls")
	if err := os.MkdirAll(wallsDir, 0o700); err != nil {
		t.Fatalf("mkdir walls: %v", err)
	}
	artPath := filepath.Join(wallsDir, "art.txt")
	if err := os.WriteFile(artPath, []byte("original"), 0o600); err != nil {
		t.Fatalf("seed: %v", err)
	}

	err := ExportToSyscWalls("art.txt", "replacement", false)
	if !errors.Is(err, ErrFileExists) {
		t.Fatalf("err = %v, want ErrFileExists", err)
	}
	var exists *ExistsError
	if !errors.As(err, &exists) {
		t.Fatalf("err = %v, want *ExistsError", err)
	}
	if exists.Path != artPath {
		t.Fatalf("ExistsError.Path = %q, want %q", exists.Path, artPath)
	}
	got, err := os.ReadFile(artPath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "original" {
		t.Fatalf("file was modified: %q", got)
	}

	if err := ExportToSyscWalls("art.txt", "replacement", true); err != nil {
		t.Fatalf("overwrite: %v", err)
	}
	got, err = os.ReadFile(artPath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "replacement" {
		t.Fatalf("overwrite did not replace: %q", got)
	}
}

// An export points the screensaver at a new file; it must not also restate
// the settings the user picked.
func TestUpdateSyscWallsConfigKeepsExistingSettings(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	wallsDir := filepath.Join(tmpHome, ".local", "share", "syscgo", "walls")
	if err := os.MkdirAll(wallsDir, 0o700); err != nil {
		t.Fatalf("mkdir walls: %v", err)
	}
	configPath := filepath.Join(tmpHome, ".config", "sysc-walls", "daemon.conf")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	existing := "[animation]\ntype = matrix\ntheme = solarized\ncycle = true\nfile = /old/art.txt\n"
	if err := os.WriteFile(configPath, []byte(existing), 0o600); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	artPath := filepath.Join(wallsDir, "new.txt")
	if err := updateSyscWallsConfig(configPath, artPath); err != nil {
		t.Fatalf("updateSyscWallsConfig: %v", err)
	}

	got, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	conf := string(got)
	for _, want := range []string{"type = matrix", "theme = solarized", "cycle = true", "file = " + artPath} {
		if !strings.Contains(conf, want) {
			t.Errorf("config lost %q:\n%s", want, conf)
		}
	}
}

// A collision has to reach the user as a question, and answering it has to be
// the thing that replaces the file.
func TestSaveFileConfirmsBeforeReplacing(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	wallsDir := filepath.Join(tmpHome, ".local", "share", "syscgo", "walls")
	if err := os.MkdirAll(wallsDir, 0o700); err != nil {
		t.Fatalf("mkdir walls: %v", err)
	}
	artPath := filepath.Join(wallsDir, "art.txt")
	if err := os.WriteFile(artPath, []byte("original"), 0o600); err != nil {
		t.Fatalf("seed: %v", err)
	}

	m := NewModel()
	m.editorMode = true
	m.showSavePrompt = true
	m.exportTarget = 1
	m.textarea.SetValue("replacement")
	m.filenameInput.SetValue("art.txt")

	asked, _ := m.saveFile()
	if !asked.confirmOverwrite {
		t.Fatal("collision did not raise a confirmation")
	}
	if asked.saveError != "" {
		t.Fatalf("collision reported as an error: %s", asked.saveError)
	}
	if got, _ := os.ReadFile(artPath); string(got) != "original" {
		t.Fatalf("file replaced before confirmation: %q", got)
	}
	if out := asked.renderSavePrompt(); !strings.Contains(out, artPath) {
		t.Fatalf("prompt does not name the destination:\n%s", out)
	}

	keptAny, _ := asked.handleEditorKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	kept := keptAny.(Model)
	if kept.confirmOverwrite || kept.overwrite {
		t.Fatal("declining left the confirmation armed")
	}
	if got, _ := os.ReadFile(artPath); string(got) != "original" {
		t.Fatalf("declining replaced the file: %q", got)
	}

	replacedAny, _ := asked.handleEditorKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	replaced := replacedAny.(Model)
	if got, _ := os.ReadFile(artPath); string(got) != "replacement" {
		t.Fatalf("confirming did not replace the file: %q", got)
	}
	if replaced.confirmOverwrite || replaced.overwrite {
		t.Fatal("confirming left the confirmation armed")
	}
}

// The save prompt names the directory it will actually write to.
func TestSavePromptNamesItsDestination(t *testing.T) {
	m := NewModel()
	m.width = 80
	m.showSavePrompt = true

	m.exportTarget = 0
	if out := m.renderSavePrompt(); !strings.Contains(out, "assets/") {
		t.Errorf("syscgo prompt does not mention assets:\n%s", out)
	}
	m.exportTarget = 1
	if out := m.renderSavePrompt(); !strings.Contains(out, ".local/share/syscgo/walls") {
		t.Errorf("sysc-walls prompt does not name the walls dir:\n%s", out)
	}
}

// A banner taller than the terminal has to be clipped, not scrolled off screen.
func TestBitPreviewIsClippedToFit(t *testing.T) {
	m := NewModel()
	m.width = 80
	m.height = 30
	m.bitPreviewLines = make([]string, 200)
	for i := range m.bitPreviewLines {
		m.bitPreviewLines[i] = "line"
	}

	out := m.renderBitEditorView()
	lines := strings.Split(out, "\n")
	if len(lines) > m.height {
		t.Fatalf("editor rendered %d rows for a %d row terminal", len(lines), m.height)
	}
	if !strings.Contains(out, "earlier rows hidden") {
		t.Error("clipped preview does not say rows were hidden")
	}
	// The notice is the first thing in the frame, and the tail of the banner
	// follows it: the rows being kept are the last ones, not the first.  The
	// frame pads every row, so compare on the text each row starts with.
	frame := []string{}
	for _, row := range strings.Split(out, "\n") {
		if !strings.Contains(row, "│") {
			continue
		}
		frame = append(frame, strings.TrimSpace(strings.Trim(strings.TrimSpace(row), "│")))
	}
	notice := -1
	for i, text := range frame {
		if strings.Contains(text, "earlier rows hidden") {
			notice = i
			break
		}
	}
	if notice < 0 {
		t.Fatalf("no notice row in frame: %q", frame)
	}
	if notice+2 >= len(frame) || frame[notice+1] != "line" || frame[notice+2] != "line" {
		t.Errorf("preview did not keep the banner tail after the notice: %q", frame[notice:])
	}
}
