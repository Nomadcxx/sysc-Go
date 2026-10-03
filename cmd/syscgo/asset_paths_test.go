package main

import (
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

func containsPath(paths []string, want string) bool {
	want = filepath.Clean(want)
	for _, path := range paths {
		if filepath.Clean(path) == want {
			return true
		}
	}
	return false
}
