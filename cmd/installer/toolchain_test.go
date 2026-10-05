package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// Under sudo's secure_path a GOROOT-only toolchain (#103) is absent from
// PATH. The README path `sudo "$(go env GOROOT)/bin/go" run ./cmd/installer/`
// builds this binary with that toolchain, so runtime.GOROOT() points at it
// and the installer must resolve go from there instead of aborting.
func TestGoToolchainFallsBackToRuntimeGOROOT(t *testing.T) {
	if runtime.GOROOT() == "" {
		t.Skip("test binary records no GOROOT")
	}
	t.Setenv("PATH", t.TempDir()) // sudoesque: no go on PATH

	got, err := goToolchain()
	if err != nil {
		t.Fatalf("goToolchain: %v", err)
	}
	if want := filepath.Join(runtime.GOROOT(), "bin", "go"); got != want {
		t.Fatalf("goToolchain = %q, want %q", got, want)
	}
}

func TestGoToolchainPrefersPATH(t *testing.T) {
	if _, err := os.Stat("/proc/self/exe"); err != nil {
		t.Skip("non-linux sandbox")
	}
	want := t.TempDir()
	bin := filepath.Join(want, "go")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write fake go: %v", err)
	}
	t.Setenv("PATH", want)

	if got, err := goToolchain(); err != nil || got != bin {
		t.Fatalf("goToolchain = %q, %v; want %q", got, err, bin)
	}
}
