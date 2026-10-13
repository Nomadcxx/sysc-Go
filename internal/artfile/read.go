// Package artfile reads bounded literal artwork for the CLI and TUI.
package artfile

import (
	"fmt"
	"io"
	"os"
	"syscall"
)

// Read accepts a regular file of at most 64 KiB. Nonblocking open prevents a
// named pipe from hanging before its type can be checked.
func Read(path string) (string, error) {
	f, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return "", err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() || info.Size() > 64<<10 {
		return "", fmt.Errorf("artwork must be a regular file within 64 KiB")
	}
	b, err := io.ReadAll(io.LimitReader(f, (64<<10)+1))
	if err != nil {
		return "", err
	}
	if len(b) > 64<<10 {
		return "", fmt.Errorf("artwork exceeds 64 KiB")
	}
	return string(b), nil
}
