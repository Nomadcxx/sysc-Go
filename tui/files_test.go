package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// saveArt runs saveToAssets from an empty working directory and returns the
// entries that directory ended up with, so a test can prove nothing was
// written into it.
func saveArt(t *testing.T, home, name string) []string {
	t.Helper()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", "")
	cwd := t.TempDir()
	t.Chdir(cwd)
	if err := saveToAssets(name, "HI", false); err != nil {
		t.Fatalf("saveToAssets(%q) error = %v", name, err)
	}
	entries, err := os.ReadDir(cwd)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

func TestSaveToAssetsUsesXDGDataDir(t *testing.T) {
	home := t.TempDir()
	if names := saveArt(t, home, "mine.txt"); len(names) != 0 {
		t.Fatalf("save wrote into the working directory: %v", names)
	}

	want := filepath.Join(home, ".local", "share", "syscgo", "assets", "mine.txt")
	if _, err := os.Stat(want); err != nil {
		t.Fatalf("art not at %s: %v", want, err)
	}
}

func TestSaveToAssetsHonorsXDGDataHome(t *testing.T) {
	home := t.TempDir()
	xdg := filepath.Join(t.TempDir(), "xdg")
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", xdg)
	t.Chdir(t.TempDir())

	if err := saveToAssets("mine.txt", "HI", false); err != nil {
		t.Fatalf("saveToAssets error = %v", err)
	}

	want := filepath.Join(xdg, "syscgo", "assets", "mine.txt")
	if _, err := os.Stat(want); err != nil {
		t.Fatalf("art not at %s: %v", want, err)
	}
}

// A relative XDG_DATA_HOME is not a location: honouring it would write the art
// relative to whatever directory the TUI was started in.
func TestSaveToAssetsIgnoresRelativeXDGDataHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", "relative/xdg")
	cwd := t.TempDir()
	t.Chdir(cwd)

	if err := saveToAssets("mine.txt", "HI", false); err != nil {
		t.Fatalf("saveToAssets error = %v", err)
	}

	want := filepath.Join(home, ".local", "share", "syscgo", "assets", "mine.txt")
	if _, err := os.Stat(want); err != nil {
		t.Fatalf("art not at %s: %v", want, err)
	}
	if _, err := os.Stat(filepath.Join(cwd, "relative")); err == nil {
		t.Fatal("save created a directory under a relative XDG_DATA_HOME in the cwd")
	}
}

// A project that happens to have its own assets/ folder must not collect the
// art, and must not collect the .write_test probe either.
func TestSaveToAssetsIgnoresProjectAssetsDir(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	projectAssets := filepath.Join(project, "assets")
	if err := os.MkdirAll(projectAssets, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", "")
	t.Chdir(project)

	if err := saveToAssets("mine.txt", "HI", false); err != nil {
		t.Fatalf("saveToAssets error = %v", err)
	}

	entries, err := os.ReadDir(projectAssets)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		t.Errorf("save left %q in the project's assets/ directory", e.Name())
	}
	want := filepath.Join(home, ".local", "share", "syscgo", "assets", "mine.txt")
	if _, err := os.Stat(want); err != nil {
		t.Fatalf("art not at %s: %v", want, err)
	}
}

func TestSaveToAssetsLegacyDir(t *testing.T) {
	home := t.TempDir()
	legacy := filepath.Join(home, "sysc-Go", "assets")
	if err := os.MkdirAll(legacy, 0o755); err != nil {
		t.Fatal(err)
	}
	if names := saveArt(t, home, "mine.txt"); len(names) != 0 {
		t.Fatalf("save wrote into the working directory: %v", names)
	}

	if _, err := os.Stat(filepath.Join(legacy, "mine.txt")); err != nil {
		t.Fatalf("art not in the legacy dir %s: %v", legacy, err)
	}
	// The legacy dir still wins the lookup, so read and save agree.
	if got := getAssetPath("mine.txt"); got != filepath.Join(legacy, "mine.txt") {
		t.Fatalf("getAssetPath(mine.txt) = %q, want the legacy dir", got)
	}
}

func TestSavedArtReadableFromOtherDir(t *testing.T) {
	home := t.TempDir()
	saveArt(t, home, "mine.txt")

	// Leave the saving directory entirely; reads happen from somewhere else.
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", "")
	t.Chdir(t.TempDir())

	files := discoverAssetFiles()
	if !containsString(files, "mine.txt") {
		t.Fatalf("discoverAssetFiles() = %v, want it to list mine.txt", files)
	}
	want := filepath.Join(home, ".local", "share", "syscgo", "assets", "mine.txt")
	if got := getAssetPath("mine.txt"); got != want {
		t.Fatalf("getAssetPath(mine.txt) = %q, want %q", got, want)
	}
}

func TestUserAssetsDirIsAbsoluteByDefault(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("HOME", "/home/user")
	if got := userAssetsDir(); got != "/home/user/.local/share/syscgo/assets" {
		t.Fatalf("userAssetsDir() = %q", got)
	}
	t.Setenv("XDG_DATA_HOME", "/xdg")
	if got := userAssetsDir(); got != "/xdg/syscgo/assets" {
		t.Fatalf("userAssetsDir() with XDG_DATA_HOME = %q", got)
	}
}

func containsString(haystack []string, want string) bool {
	for _, s := range haystack {
		if strings.EqualFold(s, want) {
			return true
		}
	}
	return false
}
