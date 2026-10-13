package artfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadBounded(t *testing.T) {
	p := filepath.Join(t.TempDir(), "art.txt")
	if err := os.WriteFile(p, []byte("RAMA"), 0600); err != nil {
		t.Fatal(err)
	}
	if got, err := Read(p); err != nil || got != "RAMA" {
		t.Fatalf("read = %q, %v", got, err)
	}
	if err := os.WriteFile(p, []byte(strings.Repeat("X", 65537)), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(p); err == nil {
		t.Fatal("accepted oversized artwork")
	}
	if _, err := Read(filepath.Dir(p)); err == nil {
		t.Fatal("accepted directory")
	}
	if _, err := os.Stat("/dev/zero"); err == nil {
		if _, err = Read("/dev/zero"); err == nil {
			t.Fatal("accepted device")
		}
	}
}
