package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallAssetFilesCopiesOnlyRuntimeAssets(t *testing.T) {
	src := t.TempDir()
	mustWrite(t, filepath.Join(src, "SYSC.txt"), "sysc")
	mustWrite(t, filepath.Join(src, "fire.gif"), "gif-bytes")
	mustWrite(t, filepath.Join(src, "preview.png"), "png-bytes")
	mustWrite(t, filepath.Join(src, "README.md"), "notes")
	if err := os.Mkdir(filepath.Join(src, "fonts"), 0755); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(src, "fonts", "banner.bit"), "bit")
	mustWrite(t, filepath.Join(src, "fonts", "preview.gif"), "font-gif")

	share := t.TempDir()
	if err := installAssetFiles(src, share); err != nil {
		t.Fatal(err)
	}

	assertPresent(t, filepath.Join(share, "assets", "SYSC.txt"))
	assertPresent(t, filepath.Join(share, "fonts", "banner.bit"))
	for _, missing := range []string{
		filepath.Join(share, "SYSC.txt"),
		filepath.Join(share, "assets", "fire.gif"),
		filepath.Join(share, "assets", "preview.png"),
		filepath.Join(share, "assets", "README.md"),
		filepath.Join(share, "assets", "fonts", "banner.bit"),
		filepath.Join(share, "fonts", "preview.gif"),
	} {
		if _, err := os.Stat(missing); !os.IsNotExist(err) {
			t.Fatalf("unexpected installed file %s", missing)
		}
	}
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
}

func assertPresent(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("missing %s: %v", path, err)
	}
	if info.IsDir() {
		t.Fatalf("%s is a directory", path)
	}
}
