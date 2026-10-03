package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallAssetFilesKeepsAssetsAndFontsSeparate(t *testing.T) {
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "SYSC.txt"), []byte("sysc"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "fire.gif"), []byte("gif"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(src, "fonts"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "fonts", "banner.bit"), []byte("bit"), 0644); err != nil {
		t.Fatal(err)
	}

	share := t.TempDir()
	if err := installAssetFiles(src, share); err != nil {
		t.Fatal(err)
	}

	assertFile(t, filepath.Join(share, "assets", "SYSC.txt"))
	assertFile(t, filepath.Join(share, "fonts", "banner.bit"))
	if _, err := os.Stat(filepath.Join(share, "SYSC.txt")); !os.IsNotExist(err) {
		t.Fatalf("text asset installed flat in share root, want .../assets/SYSC.txt")
	}
	if _, err := os.Stat(filepath.Join(share, "assets", "fonts", "banner.bit")); !os.IsNotExist(err) {
		t.Fatalf("font installed under assets/fonts, want .../fonts/banner.bit")
	}
}

func assertFile(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("missing %s: %v", path, err)
	}
	if info.IsDir() {
		t.Fatalf("%s is a directory", path)
	}
}
