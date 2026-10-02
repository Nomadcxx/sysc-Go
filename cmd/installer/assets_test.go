package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyRuntimeAssetsSkipsDemoMedia(t *testing.T) {
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

	dst := t.TempDir()
	if err := copyRuntimeAssets(src, dst); err != nil {
		t.Fatal(err)
	}

	assertPresent(t, filepath.Join(dst, "SYSC.txt"))
	assertPresent(t, filepath.Join(dst, "fonts", "banner.bit"))
	for _, missing := range []string{
		filepath.Join(dst, "fire.gif"),
		filepath.Join(dst, "preview.png"),
		filepath.Join(dst, "README.md"),
		filepath.Join(dst, "fonts", "preview.gif"),
	} {
		if _, err := os.Stat(missing); !os.IsNotExist(err) {
			t.Fatalf("installed non-runtime file %s", missing)
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
