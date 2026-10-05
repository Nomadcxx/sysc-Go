package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
)

// goToolchain resolves the go binary. Under sudo's secure_path a toolchain
// that lives only under GOROOT/bin (official tarball, version managers) is
// absent from PATH even when this binary was itself produced by `go run`
// from that toolchain (#103). runtime.GOROOT() records where that was.
func goToolchain() (string, error) {
	if p, err := exec.LookPath("go"); err == nil {
		return p, nil
	}
	if goroot := runtime.GOROOT(); goroot != "" {
		p := filepath.Join(goroot, "bin", "go")
		if _, err := exec.LookPath(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("go binary not found on PATH or under runtime.GOROOT()")
}
