package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExportToSyscWalls_PathTraversal(t *testing.T) {
	tests := []struct {
		name           string
		filename       string
		wantErr        bool
		expectedOutput string // What filename should be created after sanitization
	}{
		{"Normal filename", "art.txt", false, "art.txt"},
		{"Path traversal sanitized", "../../../etc/passwd", false, "passwd.txt"}, // filepath.Base() extracts "passwd"
		{"Current dir", ".", true, ""},
		{"Parent dir", "..", true, ""},
		{"Shell metachar semicolon", "art;.txt", true, ""},
		{"Shell metachar pipe", "art|.txt", true, ""},
		{"Shell metachar dollar", "art$.txt", true, ""},
		{"Shell metachar ampersand", "art&.txt", true, ""},
		{"Underscore allowed", "art_v2.txt", false, "art_v2.txt"},
		{"Hyphen allowed", "art-v2.txt", false, "art-v2.txt"},
		{"Spaces allowed", "my art.txt", false, "my art.txt"},
		{"Empty filename", "", true, ""},
		{"Backtick", "art`.txt", true, ""},
		{"Forward slash sanitized", "dir/art.txt", false, "art.txt"}, // filepath.Base() extracts "art.txt"
		{"Backslash rejected", "dir\\art.txt", true, ""},             // Backslash is not a separator on Linux, so it's an invalid character
	}

	// Setup temp home dir for tests
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ExportToSyscWalls(tt.filename, "test content")
			if (err != nil) != tt.wantErr {
				t.Errorf("ExportToSyscWalls() error = %v, wantErr %v", err, tt.wantErr)
			}

			// If should succeed, verify file was created with expected name
			if !tt.wantErr && tt.expectedOutput != "" {
				expectedPath := filepath.Join(tmpHome, ".local", "share", "syscgo", "walls", tt.expectedOutput)
				if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
					t.Errorf("Expected file was not created: %s", expectedPath)
				} else {
					// Verify content
					content, err := os.ReadFile(expectedPath)
					if err != nil {
						t.Errorf("Failed to read created file: %v", err)
					} else if string(content) != "test content" {
						t.Errorf("File content mismatch: got %q, want %q", string(content), "test content")
					}
				}
				// Clean up for next test
				os.Remove(expectedPath)
			}
		})
	}
}

func TestExportToSyscWalls_FilePermissions(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	filename := "test-art.txt"
	err := ExportToSyscWalls(filename, "test content")
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Check file permissions
	artPath := filepath.Join(tmpHome, ".local", "share", "syscgo", "walls", filename)
	info, err := os.Stat(artPath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	// Should be 0600 (user read/write only)
	if info.Mode().Perm() != 0600 {
		t.Errorf("File permissions = %o, want 0600", info.Mode().Perm())
	}

	// Check directory permissions
	wallsDir := filepath.Join(tmpHome, ".local", "share", "syscgo", "walls")
	dirInfo, err := os.Stat(wallsDir)
	if err != nil {
		t.Fatalf("Failed to stat directory: %v", err)
	}

	// Should be 0700 (user only)
	if dirInfo.Mode().Perm() != 0700 {
		t.Errorf("Directory permissions = %o, want 0700", dirInfo.Mode().Perm())
	}
}

func TestExportToSyscWalls_ConfigUpdate(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	filename := "custom-art.txt"
	content := "My ASCII Art"

	err := ExportToSyscWalls(filename, content)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Check config file was created/updated
	configPath := filepath.Join(tmpHome, ".config", "sysc-walls", "daemon.conf")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	configStr := string(configData)

	// Should contain file path
	expectedPath := filepath.Join(tmpHome, ".local", "share", "syscgo", "walls", filename)
	if !strings.Contains(configStr, expectedPath) {
		t.Errorf("Config does not contain file path: %s", expectedPath)
	}

	// Should contain animation section
	if !strings.Contains(configStr, "[animation]") {
		t.Error("Config missing [animation] section")
	}

	// Should set effect to beam-text
	if !strings.Contains(configStr, "effect = beam-text") {
		t.Error("Config should set beam-text as effect")
	}

	// Check config file permissions
	configInfo, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Failed to stat config file: %v", err)
	}

	// Should be 0600 (user read/write only)
	if configInfo.Mode().Perm() != 0600 {
		t.Errorf("Config file permissions = %o, want 0600", configInfo.Mode().Perm())
	}
}

func TestExportToSyscWalls_ContentIntegrity(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	filename := "content-test.txt"
	content := "Line 1\nLine 2\nLine 3"

	err := ExportToSyscWalls(filename, content)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Read back the file
	artPath := filepath.Join(tmpHome, ".local", "share", "syscgo", "walls", filename)
	savedContent, err := os.ReadFile(artPath)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	// Verify content matches
	if string(savedContent) != content {
		t.Errorf("Content mismatch:\nGot:  %q\nWant: %q", string(savedContent), content)
	}
}

func TestExportToSyscWalls_AutoTxtExtension(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	tests := []struct {
		name           string
		inputFilename  string
		expectedOutput string
	}{
		{"No extension", "myart", "myart.txt"},
		{"Already has .txt", "myart.txt", "myart.txt"},
		{"Other extension", "myart.dat", "myart.dat.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ExportToSyscWalls(tt.inputFilename, "test")
			if err != nil {
				t.Fatalf("Export failed: %v", err)
			}

			expectedPath := filepath.Join(tmpHome, ".local", "share", "syscgo", "walls", tt.expectedOutput)
			if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
				t.Errorf("Expected file not found: %s", expectedPath)
			}

			// Clean up for next test
			os.Remove(expectedPath)
		})
	}
}

func TestExportToSyscWalls_MultipleExports(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Export multiple files
	files := []string{"art1.txt", "art2.txt", "art3.txt"}
	for i, filename := range files {
		content := "Content " + string(rune('A'+i))
		err := ExportToSyscWalls(filename, content)
		if err != nil {
			t.Fatalf("Export %d failed: %v", i, err)
		}
	}

	// Verify all files exist
	wallsDir := filepath.Join(tmpHome, ".local", "share", "syscgo", "walls")
	entries, err := os.ReadDir(wallsDir)
	if err != nil {
		t.Fatalf("Failed to read walls directory: %v", err)
	}

	if len(entries) != len(files) {
		t.Errorf("Expected %d files, got %d", len(files), len(entries))
	}

	// Verify config points to last exported file
	configPath := filepath.Join(tmpHome, ".config", "sysc-walls", "daemon.conf")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	lastFilePath := filepath.Join(wallsDir, files[len(files)-1])
	if !strings.Contains(string(configData), lastFilePath) {
		t.Errorf("Config should reference last exported file: %s", lastFilePath)
	}
}
