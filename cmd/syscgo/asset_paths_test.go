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
			if !strings.Contains(warning, share) {
				t.Fatalf("warning = %q, want install share path %s", warning, share)
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
