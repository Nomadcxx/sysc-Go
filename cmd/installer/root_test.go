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

	// Three filepath.Dir calls from that binary overshoot the module, which is
	// why the lookup walks instead of assuming a depth. Asserted by arithmetic
	// rather than by stat: a go.mod anywhere above t.TempDir() — a stray
	// /tmp/go.mod, say — belongs to the machine, not to this fixture.
	old := filepath.Dir(filepath.Dir(filepath.Dir(execPath)))
	if inside(root, old) {
		t.Fatalf("fixed-depth root %s did not overshoot the module %s", old, root)
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
	useTempDirWithoutModule(t)
	start := t.TempDir()
	if _, err := os.Stat(filepath.Join(start, "go.mod")); err == nil {
		t.Fatalf("%s already contains a go.mod", start)
	}

	if got, ok := findModuleRoot(start); ok {
		t.Fatalf("findModuleRoot(%s) = %q, %v; want no module root", start, got, ok)
	}
}

// getProjectRoot resolves from the executable before the cwd, because
// install.sh builds the binary into the repo root and runs it from elsewhere.
// projectRootFrom takes that path as an argument so both halves of the order
// can be exercised; a real test binary cannot be moved out of its build dir.
func TestProjectRootFromPrefersExecutableModule(t *testing.T) {
	planted := t.TempDir()
	writeGoMod(t, planted)
	exeDir := filepath.Join(planted, "go-build", "b001", "exe")
	if err := os.MkdirAll(exeDir, 0755); err != nil {
		t.Fatal(err)
	}
	exe := filepath.Join(exeDir, "installer.test")

	cwd := t.TempDir()
	writeGoMod(t, cwd)
	chdir(t, filepath.Join(cwd, "cmd", "installer"))

	if got, want := projectRootFrom(exe), planted; got != want {
		t.Fatalf("projectRootFrom(%s) = %q, want the executable's module %q", exe, got, want)
	}
}

// A binary built outside any module — `go run ./cmd/installer`, which lands in
// the build cache — still resolves, from the cwd it was invoked in.
func TestProjectRootFromFallsBackToCwd(t *testing.T) {
	useTempDirWithoutModule(t)
	exe := filepath.Join(t.TempDir(), "go-build", "b001", "exe", "installer.test")

	cwd := t.TempDir()
	writeGoMod(t, cwd)
	chdir(t, filepath.Join(cwd, "cmd", "installer"))

	if got, want := projectRootFrom(exe), cwd; got != want {
		t.Fatalf("projectRootFrom(%s) = %q, want the cwd's module %q", exe, got, want)
	}
}

func TestProjectRootFromWithNothingToFind(t *testing.T) {
	useTempDirWithoutModule(t)
	exe := filepath.Join(t.TempDir(), "installer.test")

	chdir(t, t.TempDir())

	if got, want := projectRootFrom(exe), "."; got != want {
		t.Fatalf("projectRootFrom(%s) = %q, want %q", exe, got, want)
	}
}

// t.TempDir lives under os.TempDir, and a go.mod anywhere above that — a stray
// /tmp/go.mod is enough to poison every walk on the machine — makes "no module
// root" untestable. These tests need a directory whose whole ancestry is
// go.mod-free, so TMPDIR is moved to one that is. Candidates are probed rather
// than hardcoded because none of them exist on every machine.
func useTempDirWithoutModule(t *testing.T) {
	t.Helper()
	for _, candidate := range []string{os.Getenv("XDG_RUNTIME_DIR"), "/var/tmp", "/dev/shm"} {
		if candidate == "" {
			continue
		}
		probe, err := os.MkdirTemp(candidate, "syscgo-clean-*")
		if err != nil {
			continue
		}
		t.Cleanup(func() { _ = os.RemoveAll(probe) })
		if _, found := findModuleRoot(probe); found {
			continue
		}
		t.Setenv("TMPDIR", probe)
		return
	}
	t.Skip("no directory with a go.mod-free ancestry is available on this machine")
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
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
