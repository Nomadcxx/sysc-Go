package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// useTempInstallDirs retargets the install destinations at a throwaway
// directory. The locals must not be called binDir/shareDir: named returns would
// shadow the package variables and leave the install writing to the real paths.
func useTempInstallDirs(t *testing.T) (bin, share string) {
	t.Helper()
	oldBin, oldShare := binDir, shareDir
	binDir, shareDir = t.TempDir(), t.TempDir()
	t.Cleanup(func() { binDir, shareDir = oldBin, oldShare })
	return binDir, shareDir
}

func TestInstallFileAtomicReplacesRunningBinary(t *testing.T) {
	sleepBin, err := exec.LookPath("sleep")
	if err != nil {
		t.Skip("sleep not available")
	}
	dir := t.TempDir()
	dst := filepath.Join(dir, "syscgo")
	data, err := os.ReadFile(sleepBin)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, data, 0755); err != nil {
		t.Fatal(err)
	}

	// Hold the destination open the way a running syscgo would, then wait for
	// the child to actually be exec'd: Start returns after the fork.
	proc := exec.Command(dst, "30")
	if err := proc.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = proc.Process.Kill()
		_, _ = proc.Process.Wait()
	}()

	// The old truncate-then-write path is refused with ETXTBSY once the
	// process is exec'd, which is the precondition this test needs.
	var refused error
	for i := 0; i < 500 && refused == nil; i++ {
		refused = os.WriteFile(dst, []byte("short"), 0755)
		if refused == nil {
			time.Sleep(10 * time.Millisecond)
		}
	}
	if refused == nil {
		t.Skip("destination never became busy; cannot exercise the atomic path")
	}

	if err := installFileAtomic(dst, []byte("new"), binMode); err != nil {
		t.Fatalf("atomic install refused while the binary runs: %v", err)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new" {
		t.Fatalf("destination = %q, want %q", got, "new")
	}
}

func TestInstallFileAtomicRepairsTightMode(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "syscgo")
	// A previous run under a restrictive umask left this unreadable to others.
	if err := os.WriteFile(dst, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := installFileAtomic(dst, []byte("new"), binMode); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != binMode {
		t.Fatalf("mode = %v, want %v", info.Mode().Perm(), os.FileMode(binMode))
	}
}

func TestInstallFileAtomicLeavesNoTempFiles(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "syscgo")
	if err := installFileAtomic(dst, []byte("new"), binMode); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "syscgo" {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Fatalf("directory holds %v, want only syscgo", names)
	}
}

func TestInstallFileAtomicFailureKeepsDestination(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "syscgo")
	// A directory in dst's place makes the rename fail after the write.
	if err := os.Mkdir(dst, 0755); err != nil {
		t.Fatal(err)
	}
	if err := installFileAtomic(dst, []byte("new"), binMode); err == nil {
		t.Fatal("expected the install to fail when the rename cannot land")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.Contains(e.Name(), ".tmp-") {
			t.Fatalf("failed install left %s behind", e.Name())
		}
	}
}

func TestCopyFileIgnoresSourceMode(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "SYSC.txt")
	dst := filepath.Join(dir, "out.txt")
	if err := os.WriteFile(src, []byte("art"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := copyFile(src, dst); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != assetMode {
		t.Fatalf("mode = %v, want %v", info.Mode().Perm(), os.FileMode(assetMode))
	}
}

func TestInstallAssetFilesSetsDirectoryModes(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	if err := os.MkdirAll(filepath.Join(src, "fonts"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "SYSC.txt"), []byte("art"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "fonts", "a.bit"), []byte("f"), 0600); err != nil {
		t.Fatal(err)
	}
	share := t.TempDir()
	// Pre-existing too-tight directories, as a stricter umask would leave them.
	assetsDst := filepath.Join(share, "assets")
	fontsDst := filepath.Join(share, "fonts")
	for _, d := range []string{assetsDst, fontsDst} {
		if err := os.MkdirAll(d, 0700); err != nil {
			t.Fatal(err)
		}
	}

	if err := installAssetFiles(context.Background(), src, share); err != nil {
		t.Fatal(err)
	}
	for _, d := range []string{assetsDst, fontsDst} {
		info, err := os.Stat(d)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != dirMode {
			t.Fatalf("%s mode = %v, want %v", d, info.Mode().Perm(), os.FileMode(dirMode))
		}
	}
	for _, f := range []string{filepath.Join(assetsDst, "SYSC.txt"), filepath.Join(fontsDst, "a.bit")} {
		info, err := os.Stat(f)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != assetMode {
			t.Fatalf("%s mode = %v, want %v", f, info.Mode().Perm(), os.FileMode(assetMode))
		}
	}
}

func TestInstallBinariesUseAtomicInstall(t *testing.T) {
	bin, _ := useTempInstallDirs(t)
	for _, tc := range []struct {
		name string
		fn   func(*model) error
	}{
		{"syscgo", installBinary},
		{"syscgo-tui", installTuiBinary},
	} {
		t.Run(tc.name, func(t *testing.T) {
			old := buildOutputDir
			buildOutputDir = t.TempDir()
			t.Cleanup(func() { buildOutputDir = old })
			// Stage a source that is not executable, so a mode copied from the
			// source would fail the check below.
			src, err := stagedBinaryPath(tc.name)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(src, []byte("bin"), 0644); err != nil {
				t.Fatal(err)
			}
			// A stale, unreadable destination must be repaired by the install.
			dst := filepath.Join(bin, tc.name)
			if err := os.WriteFile(dst, []byte("old"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := tc.fn(&model{}); err != nil {
				t.Fatal(err)
			}
			info, err := os.Stat(dst)
			if err != nil {
				t.Fatal(err)
			}
			if info.Mode().Perm() != binMode {
				t.Fatalf("mode = %v, want %v", info.Mode().Perm(), os.FileMode(binMode))
			}
		})
	}
}

func TestBuildStagedDisablesBuildVCS(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "bin")
	if err := os.Mkdir(bin, 0755); err != nil {
		t.Fatal(err)
	}
	argsPath := filepath.Join(dir, "args")
	writeStub(t, bin, "go", "#!/bin/bash\nprintf '%s\\n' \"$@\" > "+argsPath+"\n")

	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	old := buildOutputDir
	buildOutputDir = t.TempDir()
	t.Cleanup(func() { buildOutputDir = old })

	if err := buildStaged(&model{}, "syscgo", "./cmd/syscgo"); err != nil {
		t.Fatal(err)
	}
	args, err := os.ReadFile(argsPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(args), "-buildvcs=false") {
		t.Fatalf("go build args %q lack -buildvcs=false; root builds over a user checkout fail", args)
	}
}
