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
			err := ExportToSyscWalls(tt.filename, "test content", false)
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
	err := ExportToSyscWalls(filename, "test content", false)
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

	err := ExportToSyscWalls(filename, content, false)
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

	// sysc-walls reads [animation] effect.  It has no `type` key and silently
	// ignores it, which is why the export used to leave the screensaver running
	// whatever effect the user already had.
	if !strings.Contains(configStr, "effect = beam-text") {
		t.Errorf("Config should set effect = beam-text, got:\n%s", configStr)
	}
	if strings.Contains(configStr, "type =") {
		t.Errorf("Config should not contain a type key, got:\n%s", configStr)
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

	err := ExportToSyscWalls(filename, content, false)
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
			err := ExportToSyscWalls(tt.inputFilename, "test", false)
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
		err := ExportToSyscWalls(filename, content, false)
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

// The config from issue #117's repro: a normal sysc-walls setup on a
// non-text effect.  Everything but the four owned keys must survive verbatim.
const reproConfig = `# sysc-walls daemon configuration
[idle]
timeout = 5m
min_duration = 30s
[daemon]
debug = false
[animation]
effect = fire
# Available effects: fire, matrix, ...
theme = nord
datetime = false
cycle = true
cycle_interval = 5m
[datetime]
position = bottom
interval = 1s
[terminal]
kitty = true
fullscreen = true
`

// writeConfig drops a config into a temp home and returns its path.
func writeConfig(t *testing.T, content string, mode os.FileMode) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".config", "sysc-walls")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "daemon.conf")
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatal(err)
	}
	return path
}

func readConfig(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestUpdateSyscWallsConfigPreservesUserConfig(t *testing.T) {
	const art = "/home/user/.local/share/syscgo/walls/wallhi.txt"
	path := writeConfig(t, reproConfig, 0o600)

	if err := updateSyscWallsConfig(path, art); err != nil {
		t.Fatalf("updateSyscWallsConfig error = %v", err)
	}

	// theme and cycle belong to the user: the export changes what plays, not how
	// it looks (#114).
	want := `# sysc-walls daemon configuration
[idle]
timeout = 5m
min_duration = 30s
[daemon]
debug = false
[animation]
effect = beam-text
# Available effects: fire, matrix, ...
theme = nord
datetime = false
cycle = true
cycle_interval = 5m
file = ` + art + `
[datetime]
position = bottom
interval = 1s
[terminal]
kitty = true
fullscreen = true
`
	if got := readConfig(t, path); got != want {
		t.Fatalf("config =\n%s\nwant\n%s", got, want)
	}
}

func TestUpdateSyscWallsConfigRemovesLegacyType(t *testing.T) {
	const art = "/walls/legacy.txt"
	in := strings.Replace(reproConfig, "effect = fire\n", "effect = fire\ntype = beam-text\n", 1)
	path := writeConfig(t, in, 0o600)

	if err := updateSyscWallsConfig(path, art); err != nil {
		t.Fatalf("updateSyscWallsConfig error = %v", err)
	}

	got := readConfig(t, path)
	if strings.Contains(got, "type =") {
		t.Errorf("legacy type key survived:\n%s", got)
	}
	if !strings.Contains(got, "effect = beam-text") {
		t.Errorf("effect not set:\n%s", got)
	}
}

func TestUpdateSyscWallsConfigCollapsesDuplicateOwnedKeys(t *testing.T) {
	const art = "/walls/dup.txt"
	in := `[animation]
effect = fire
theme = nord
effect = matrix
file = /walls/old.txt
file = /walls/stale.txt
`
	path := writeConfig(t, in, 0o600)

	if err := updateSyscWallsConfig(path, art); err != nil {
		t.Fatalf("updateSyscWallsConfig error = %v", err)
	}

	got := readConfig(t, path)
	if n := strings.Count(got, "effect ="); n != 1 {
		t.Errorf("effect appears %d times, want 1:\n%s", n, got)
	}
	if n := strings.Count(got, "file ="); n != 1 {
		t.Errorf("file appears %d times, want 1:\n%s", n, got)
	}
	if strings.Contains(got, "/walls/stale.txt") {
		t.Errorf("the superseded file value survived:\n%s", got)
	}
	if !strings.Contains(got, "file = "+art) {
		t.Errorf("file not pointed at the export:\n%s", got)
	}
}

func TestUpdateSyscWallsConfigCreatesDefault(t *testing.T) {
	const art = "/walls/fresh.txt"
	home := t.TempDir()
	t.Setenv("HOME", home)
	path := filepath.Join(home, ".config", "sysc-walls", "daemon.conf")

	if err := updateSyscWallsConfig(path, art); err != nil {
		t.Fatalf("updateSyscWallsConfig error = %v", err)
	}

	got := readConfig(t, path)
	if !strings.Contains(got, "[animation]") {
		t.Errorf("default config has no [animation] section:\n%s", got)
	}
	if !strings.Contains(got, "effect = beam-text") {
		t.Errorf("default config has no effect = beam-text:\n%s", got)
	}
	if !strings.Contains(got, "file = "+art) {
		t.Errorf("default config does not point at the export:\n%s", got)
	}
	if strings.Contains(got, "%") {
		t.Errorf("default config has an unsubstituted placeholder:\n%s", got)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("new config mode = %o, want 0600", info.Mode().Perm())
	}
}

func TestUpdateSyscWallsConfigPreservesExistingMode(t *testing.T) {
	path := writeConfig(t, reproConfig, 0o644)

	if err := updateSyscWallsConfig(path, "/walls/mode.txt"); err != nil {
		t.Fatalf("updateSyscWallsConfig error = %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o644 {
		t.Errorf("config mode = %o, want the pre-existing 0644", info.Mode().Perm())
	}
}

func TestUpdateSyscWallsConfigKeyOrderStable(t *testing.T) {
	const first, second = "/walls/one.txt", "/walls/two.txt"
	path := writeConfig(t, reproConfig, 0o600)

	if err := updateSyscWallsConfig(path, first); err != nil {
		t.Fatal(err)
	}
	if err := updateSyscWallsConfig(path, second); err != nil {
		t.Fatal(err)
	}

	got := readConfig(t, path)
	if !strings.Contains(got, "file = "+second) {
		t.Fatalf("second export did not win:\n%s", got)
	}
	// Running the update again must be a no-op apart from the file value.
	want := strings.Replace(got, "file = "+second, "file = "+second, 1)
	if err := updateSyscWallsConfig(path, second); err != nil {
		t.Fatal(err)
	}
	if again := readConfig(t, path); again != want {
		t.Fatalf("config churned on a repeat export:\nfirst\n%s\nsecond\n%s", want, again)
	}
}

func TestUpdateSyscWallsConfigAppendsMissingSection(t *testing.T) {
	const art = "/walls/append.txt"
	in := "[idle]\ntimeout = 5m\n"
	path := writeConfig(t, in, 0o600)

	if err := updateSyscWallsConfig(path, art); err != nil {
		t.Fatalf("updateSyscWallsConfig error = %v", err)
	}

	got := readConfig(t, path)
	if !strings.Contains(got, "[idle]\ntimeout = 5m\n") {
		t.Errorf("existing section lost:\n%s", got)
	}
	if !strings.Contains(got, "[animation]\neffect = beam-text") {
		t.Errorf("[animation] not appended with its keys:\n%s", got)
	}
	if !strings.Contains(got, "file = "+art) {
		t.Errorf("appended section does not point at the export:\n%s", got)
	}
}

func TestUpdateSyscWallsConfigSpacedSectionHeader(t *testing.T) {
	const art = "/walls/spaced.txt"
	in := "[ animation ]\neffect = fire\n[terminal]\nkitty = true\n"
	path := writeConfig(t, in, 0o600)

	if err := updateSyscWallsConfig(path, art); err != nil {
		t.Fatalf("updateSyscWallsConfig error = %v", err)
	}

	got := readConfig(t, path)
	if !strings.Contains(got, "[ animation ]\neffect = beam-text") {
		t.Errorf("spaced section header not matched:\n%s", got)
	}
	if !strings.Contains(got, "[terminal]\nkitty = true") {
		t.Errorf("following section disturbed:\n%s", got)
	}
}

func TestUpdateSyscWallsConfigCRLFAndSpacing(t *testing.T) {
	const art = "/walls/crlf.txt"
	in := strings.ReplaceAll(`[animation]
  effect=fire
	theme=nord
cycle = true
[terminal]
kitty = true
`, "\n", "\r\n")
	path := writeConfig(t, in, 0o600)

	if err := updateSyscWallsConfig(path, art); err != nil {
		t.Fatalf("updateSyscWallsConfig error = %v", err)
	}

	got := readConfig(t, path)
	// The user's own spelling of each key is kept, including indentation and the
	// absence of spaces around "=".
	for _, want := range []string{
		"\r\n  effect=beam-text\r\n", // indentation and "=" spacing preserved
		"\r\n\ttheme=nord\r\n",       // not owned: left as the user wrote it
		"\r\ncycle = true\r\n",       // not owned: left as the user wrote it
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%q", want, got)
		}
	}
	if strings.Contains(got, "\n\n") {
		t.Errorf("CRLF config gained a stray blank line:\n%q", got)
	}
	if !strings.Contains(got, "\r\nfile = "+art+"\r\n") {
		t.Errorf("inserted file line should use CRLF:\n%q", got)
	}
}

func TestUpdateSyscWallsConfigAnimationLastSectionNoTrailingNewline(t *testing.T) {
	const art = "/walls/notrail.txt"
	in := "[terminal]\nkitty = true\n[animation]\neffect = fire"
	path := writeConfig(t, in, 0o600)

	if err := updateSyscWallsConfig(path, art); err != nil {
		t.Fatalf("updateSyscWallsConfig error = %v", err)
	}

	got := readConfig(t, path)
	if !strings.Contains(got, "effect = beam-text") {
		t.Errorf("unterminated line not updated:\n%q", got)
	}
	if !strings.Contains(got, "file = "+art+"\n") {
		t.Errorf("missing key not appended with a newline:\n%q", got)
	}
}

func TestUpdateSyscWallsConfigInsertedBeforeTrailingBlanks(t *testing.T) {
	const art = "/walls/blank.txt"
	in := "[animation]\neffect = fire\n\n\n[terminal]\nkitty = true\n"
	path := writeConfig(t, in, 0o600)

	if err := updateSyscWallsConfig(path, art); err != nil {
		t.Fatalf("updateSyscWallsConfig error = %v", err)
	}

	got := readConfig(t, path)
	if !strings.Contains(got, "effect = beam-text\nfile = "+art+"\n\n\n[terminal]") {
		t.Errorf("inserted keys did not land before the trailing blanks:\n%q", got)
	}
}

// A symlinked config is a legitimate setup (dotfile managed).  Renaming onto
// the link would silently detach it, so the resolved file is replaced.
func TestUpdateSyscWallsConfigKeepsSymlink(t *testing.T) {
	const art = "/walls/linked.txt"
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".config", "sysc-walls")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(home, "real-daemon.conf")
	if err := os.WriteFile(target, []byte(reproConfig), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "daemon.conf")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	if err := updateSyscWallsConfig(link, art); err != nil {
		t.Fatalf("updateSyscWallsConfig error = %v", err)
	}

	info, err := os.Lstat(link)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("the symlink was replaced by a regular file")
	}
	if !strings.Contains(readConfig(t, target), "effect = beam-text") {
		t.Errorf("symlink target not updated:\n%s", readConfig(t, target))
	}
}

func TestUpdateSyscWallsConfigLeavesNoTempFiles(t *testing.T) {
	path := writeConfig(t, reproConfig, 0o600)
	if err := updateSyscWallsConfig(path, "/walls/clean.txt"); err != nil {
		t.Fatalf("updateSyscWallsConfig error = %v", err)
	}

	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.Contains(e.Name(), ".tmp-") {
			t.Errorf("temp file left behind: %s", e.Name())
		}
	}
}

// A truncate-then-write config can be read half-updated by the daemon.  The
// write goes through a temp file and a rename, so the destination is swapped
// for a new inode rather than rewritten in place.
func TestUpdateSyscWallsConfigReplacesAtomically(t *testing.T) {
	path := writeConfig(t, reproConfig, 0o600)

	before, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := updateSyscWallsConfig(path, "/walls/atomic.txt"); err != nil {
		t.Fatalf("updateSyscWallsConfig error = %v", err)
	}
	after, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	if os.SameFile(before, after) {
		t.Error("config was rewritten in place instead of being replaced; a reader could see it half-written")
	}
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("config dir should hold only daemon.conf, got %v", names)
	}
}

// When the rename cannot happen, the old config must survive and the temp file
// must not be left behind.
func TestUpdateSyscWallsConfigFailureKeepsConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".config", "sysc-walls")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	// A directory cannot be renamed onto, so the update fails at the last step.
	bad := filepath.Join(dir, "daemon.conf")
	if err := os.MkdirAll(bad, 0o700); err != nil {
		t.Fatal(err)
	}

	if err := updateSyscWallsConfig(bad, "/walls/bad.txt"); err == nil {
		t.Fatal("updateSyscWallsConfig succeeded writing over a directory")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.Contains(e.Name(), ".tmp-") {
			t.Errorf("failed update left %s behind", e.Name())
		}
	}
	if info, err := os.Stat(bad); err != nil || !info.IsDir() {
		t.Errorf("destination no longer a directory: %v", err)
	}
}
