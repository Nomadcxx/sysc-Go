package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindModuleRootFromInstallScriptLayout(t *testing.T) {
	root := t.TempDir()
	writeGoMod(t, root)

	// install.sh builds -o install-syscgo in the repo root, then runs that binary.
	execPath := filepath.Join(root, "install-syscgo")
	got, ok := findModuleRoot(filepath.Dir(execPath))
	if !ok || got != root {
		t.Fatalf("findModuleRoot(%s) = %q, %v; want %s", filepath.Dir(execPath), got, ok, root)
	}

	// Three filepath.Dir calls from that binary leave the module.
	old := filepath.Dir(filepath.Dir(filepath.Dir(execPath)))
	if _, err := os.Stat(filepath.Join(old, "go.mod")); err == nil {
		t.Fatalf("old fixed-depth root %s unexpectedly contains go.mod", old)
	}
}

func TestFindModuleRootFromCmdInstaller(t *testing.T) {
	root := t.TempDir()
	writeGoMod(t, root)
	start := filepath.Join(root, "cmd", "installer")
	if err := os.MkdirAll(start, 0755); err != nil {
		t.Fatal(err)
	}

	got, ok := findModuleRoot(start)
	if !ok || got != root {
		t.Fatalf("findModuleRoot(%s) = %q, %v; want %s", start, got, ok, root)
	}
}

func TestFindModuleRootMissing(t *testing.T) {
	if _, ok := findModuleRoot(t.TempDir()); ok {
		t.Fatal("expected no module root")
	}
}

func TestGetProjectRootFallsBackToCwd(t *testing.T) {
	root := t.TempDir()
	writeGoMod(t, root)
	nested := filepath.Join(root, "cmd", "installer")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatal(err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := os.Chdir(nested); err != nil {
		t.Fatal(err)
	}

	// The test binary lives outside this temp module, so resolution uses cwd.
	got := getProjectRoot()
	if got != root {
		t.Fatalf("getProjectRoot() = %q, want %q", got, root)
	}
}

func TestStagedBinaryIsOutsideProject(t *testing.T) {
	t.Cleanup(cleanupBuildOutput)

	project := t.TempDir()
	writeGoMod(t, project)
	path, err := stagedBinaryPath("syscgo")
	if err != nil {
		t.Fatal(err)
	}
	if inside(project, path) {
		t.Fatalf("staged binary %s is inside project %s", path, project)
	}
	if !inside(os.TempDir(), path) {
		t.Fatalf("staged binary %s is not under %s", path, os.TempDir())
	}
	if filepath.Base(path) != "syscgo" {
		t.Fatalf("base = %s", filepath.Base(path))
	}

	again, err := stagedBinaryPath("syscgo-tui")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(again) != filepath.Dir(path) {
		t.Fatalf("staging dir changed: %s vs %s", again, path)
	}
}

func writeGoMod(t *testing.T, dir string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/sysc\n"), 0644); err != nil {
		t.Fatal(err)
	}
}

func inside(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
