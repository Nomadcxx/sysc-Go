package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAssetSearchPathsUseSyscgoShare(t *testing.T) {
	paths := assetSearchPaths("SYSC.txt", "/home/user", "/usr/local/bin")

	want := []string{
		"/usr/local/share/syscgo/assets/SYSC.txt",
		"/usr/share/syscgo/assets/SYSC.txt",
	}
	for _, path := range want {
		if !containsPath(paths, path) {
			t.Errorf("search paths missing %s\n got: %v", path, paths)
		}
	}
	for _, path := range paths {
		if strings.Contains(path, "sysc-Go") && strings.Contains(path, "share") {
			t.Errorf("system share path still uses sysc-Go: %s", path)
		}
	}
}

// withInstallShare points the asset search at a temp directory that holds SYSC.txt
// and nowhere else, matching an install where the file is only under the share path.
func withInstallShare(t *testing.T, contents string) string {
	t.Helper()
	share := t.TempDir()
	restore := assetShareDirs
	assetShareDirs = []string{share}
	t.Cleanup(func() { assetShareDirs = restore })

	if err := os.WriteFile(filepath.Join(share, "SYSC.txt"), []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", t.TempDir())
	t.Chdir(t.TempDir())
	return share
}

// A bare name that exists in the install share resolves to it directly, with no
// fallback warning.  It used to warn, because the requested name was only ever
// read relative to the cwd and always went through the SYSC.txt fallback.
func TestPourPrintBlackholeResolveInstallShareFile(t *testing.T) {
	const installed = "installed SYSC art\n"
	share := withInstallShare(t, installed)

	for _, effect := range []string{"pour", "print", "blackhole"} {
		t.Run(effect, func(t *testing.T) {
			text, warning, err := effectFileText(effect, "SYSC.txt")
			if err != nil {
				t.Fatalf("effectFileText(%q, SYSC.txt) error = %v", effect, err)
			}
			if text != installed {
				t.Fatalf("effectFileText(%q, SYSC.txt) = %q, want installed asset", effect, text)
			}
			if warning != "" {
				t.Fatalf("warning = %q, want none: the name was found in an asset dir", warning)
			}
			if got := findAssetFile("SYSC.txt"); got != filepath.Join(share, "SYSC.txt") {
				t.Fatalf("findAssetFile(SYSC.txt) = %q, want %q", got, filepath.Join(share, "SYSC.txt"))
			}
		})
	}
}

func TestEffectFileTextFallsBackToInstalledSYSC(t *testing.T) {
	const installed = "fallback art\n"
	share := withInstallShare(t, installed)

	text, warning, err := effectFileText("blackhole", "missing.txt")
	if err != nil {
		t.Fatalf("blackhole missing file error = %v, want installed SYSC.txt", err)
	}
	if text != installed {
		t.Fatalf("text = %q, want installed SYSC.txt", text)
	}
	if !strings.Contains(warning, "missing.txt") || !strings.Contains(warning, share) {
		t.Fatalf("warning = %q, want missing name and share path", warning)
	}
}

func TestEffectFileTextPrefersExplicitPath(t *testing.T) {
	withInstallShare(t, "share")
	if err := os.WriteFile("banner.txt", []byte("local"), 0o644); err != nil {
		t.Fatal(err)
	}

	text, warning, err := effectFileText("print", "banner.txt")
	if err != nil {
		t.Fatal(err)
	}
	if text != "local" || warning != "" {
		t.Fatalf("text=%q warning=%q, want the explicit file and no fallback warning", text, warning)
	}
}

func TestBlackholeEmptyFileStaysParticleMode(t *testing.T) {
	withInstallShare(t, "share")

	text, warning, err := effectFileText("blackhole", "")
	if err != nil {
		t.Fatal(err)
	}
	if text != "" || warning != "" {
		t.Fatalf("text=%q warning=%q, want empty particle mode", text, warning)
	}
}

func TestPourWithoutFileUsesInstalledSYSC(t *testing.T) {
	const installed = "default art\n"
	withInstallShare(t, installed)

	text, warning, err := effectFileText("pour", "")
	if err != nil {
		t.Fatal(err)
	}
	if text != installed || warning != "" {
		t.Fatalf("text=%q warning=%q, want installed SYSC.txt with no warning", text, warning)
	}
}

func containsPath(paths []string, want string) bool {
	want = filepath.Clean(want)
	for _, path := range paths {
		if filepath.Clean(path) == want {
			return true
		}
	}
	return false
}

func TestResolveTextFileFindsNamedInstalledAsset(t *testing.T) {
	share := withInstallShare(t, "SYSC1 art\n")
	if err := os.WriteFile(filepath.Join(share, "SYSC2.txt"), []byte("SYSC2 art\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	text, warning, err := resolveTextFile("SYSC2.txt")
	if err != nil {
		t.Fatalf("resolveTextFile(SYSC2.txt) error = %v", err)
	}
	if text != "SYSC2 art\n" {
		t.Fatalf("text = %q, want the installed SYSC2 art, not the SYSC.txt fallback", text)
	}
	if warning != "" {
		t.Fatalf("warning = %q, want none: the named file was found in an asset dir", warning)
	}
}

func TestResolveTextFileFindsUserAssetDir(t *testing.T) {
	home := t.TempDir()
	share := withInstallShare(t, "SYSC.txt art\n")
	t.Setenv("HOME", home)

	for _, dir := range []string{
		filepath.Join(home, "sysc-Go", "assets"),
		filepath.Join(home, ".local", "share", "syscgo", "assets"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "hyprland.txt"), []byte("MY ART\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	t.Chdir(t.TempDir()) // cwd holds neither name

	text, warning, err := resolveTextFile("hyprland.txt")
	if err != nil {
		t.Fatalf("resolveTextFile(hyprland.txt) error = %v", err)
	}
	if text != "MY ART\n" {
		t.Fatalf("text = %q, want the user-dir art; share=%s", text, share)
	}
	if warning != "" {
		t.Fatalf("warning = %q, want none", warning)
	}
}

func TestResolveTextFileDoesNotSearchForPaths(t *testing.T) {
	const installed = "SYSC art\n"
	share := withInstallShare(t, installed)
	// The file really is installed, so a searched path would find it.
	if err := os.WriteFile(filepath.Join(share, "SYSC2.txt"), []byte("SYSC2 art\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	text, warning, err := resolveTextFile("sub/SYSC2.txt")
	if err != nil {
		t.Fatalf("resolveTextFile(sub/SYSC2.txt) error = %v", err)
	}
	if text != installed {
		t.Fatalf("text = %q, want the SYSC.txt fallback for a name with a separator", text)
	}
	if !strings.Contains(warning, "sub/SYSC2.txt") || !strings.Contains(warning, share) {
		t.Fatalf("warning = %q, want the missing name and the share path", warning)
	}
}

func TestResolveTextFileSkipsUnreadableCandidate(t *testing.T) {
	home := t.TempDir()
	share := withInstallShare(t, "SYSC.txt art\n")
	t.Setenv("HOME", home)

	// A directory named like the request sits first in the search order.  It
	// stats fine and opens fine, so it would shadow a real file behind it.
	legacy := filepath.Join(home, "sysc-Go", "assets")
	if err := os.MkdirAll(legacy, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(legacy, "clash.txt"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(share, "clash.txt"), []byte("REAL ART\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(t.TempDir())

	text, warning, err := resolveTextFile("clash.txt")
	if err != nil {
		t.Fatalf("resolveTextFile(clash.txt) error = %v", err)
	}
	if text != "REAL ART\n" {
		t.Fatalf("text = %q, want the readable later candidate", text)
	}
	if warning != "" {
		t.Fatalf("warning = %q, want none", warning)
	}
}

func TestAssetSearchPathsSkipsEmptyHome(t *testing.T) {
	paths := assetSearchPaths("SYSC.txt", "", "/usr/local/bin")
	for _, path := range paths {
		if strings.HasPrefix(path, "sysc-Go/") || strings.HasPrefix(path, ".local/share") {
			t.Errorf("home-derived candidate %q leaks into the search when HOME is unset", path)
		}
	}
}

func TestAssetSearchPathsIncludeUserDataDir(t *testing.T) {
	paths := assetSearchPaths("SYSC.txt", "/home/user", "/usr/local/bin")

	want := "/home/user/.local/share/syscgo/assets/SYSC.txt"
	if !containsPath(paths, want) {
		t.Errorf("search paths missing %s\n got: %v", want, paths)
	}
	// The user data dir must outrank the system share, per the README order.
	var dataIdx, shareIdx = -1, -1
	for i, path := range paths {
		if path == want {
			dataIdx = i
		}
		if path == "/usr/local/share/syscgo/assets/SYSC.txt" {
			shareIdx = i
		}
	}
	if dataIdx > shareIdx {
		t.Errorf("user data dir at %d comes after the system share at %d: %v", dataIdx, shareIdx, paths)
	}
}
