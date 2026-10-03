package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveFileHonorsExportTarget(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	assetsDir := filepath.Join(tmpHome, "sysc-Go", "assets")
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		t.Fatalf("mkdir assets: %v", err)
	}

	const content = "HELLO\nWALL\n"

	t.Run("syscgo", func(t *testing.T) {
		m := NewModel()
		m.editorMode = true
		m.showSavePrompt = true
		m.exportTarget = 0
		m.textarea.SetValue(content)
		m.filenameInput.SetValue("custom-assets")

		saved, _ := m.saveFile()
		if saved.saveError != "" {
			t.Fatalf("save error: %s", saved.saveError)
		}

		assetsPath := filepath.Join(assetsDir, "custom-assets.txt")
		got, err := os.ReadFile(assetsPath)
		if err != nil {
			t.Fatalf("assets file: %v", err)
		}
		if string(got) != content {
			t.Fatalf("assets content = %q, want %q", got, content)
		}

		wallsPath := filepath.Join(tmpHome, ".local", "share", "syscgo", "walls", "custom-assets.txt")
		if _, err := os.Stat(wallsPath); !os.IsNotExist(err) {
			t.Fatalf("syscgo save wrote walls path: %v", err)
		}
	})

	t.Run("sysc-walls", func(t *testing.T) {
		m := NewModel()
		m.editorMode = true
		m.showSavePrompt = true
		m.exportTarget = 1
		m.textarea.SetValue(content)
		m.filenameInput.SetValue("custom-walls")

		saved, _ := m.saveFile()
		if saved.saveError != "" {
			t.Fatalf("save error: %s", saved.saveError)
		}
		if saved.editorMode || saved.showSavePrompt {
			t.Fatal("successful walls save should leave the editor")
		}

		wallsPath := filepath.Join(tmpHome, ".local", "share", "syscgo", "walls", "custom-walls.txt")
		got, err := os.ReadFile(wallsPath)
		if err != nil {
			t.Fatalf("walls file: %v", err)
		}
		if string(got) != content {
			t.Fatalf("walls content = %q, want %q", got, content)
		}

		assetsPath := filepath.Join(assetsDir, "custom-walls.txt")
		if _, err := os.Stat(assetsPath); !os.IsNotExist(err) {
			t.Fatalf("sysc-walls save wrote assets path: %v", err)
		}

		configPath := filepath.Join(tmpHome, ".config", "sysc-walls", "daemon.conf")
		config, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("daemon.conf: %v", err)
		}
		if !strings.Contains(string(config), wallsPath) {
			t.Fatalf("daemon.conf missing walls path %s:\n%s", wallsPath, config)
		}
	})
}
