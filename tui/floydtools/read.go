package floydtools

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"
)

func init() {
	Register("read", func() Tool { return &ReadTool{} })
}

// ReadTool reads files from disk.
type ReadTool struct{}

func (t *ReadTool) Name() string              { return "read" }
func (t *ReadTool) Description() string       { return "Read a file from disk and display its contents." }
func (t *ReadTool) FrameDelay() time.Duration { return 0 } // Instant output

func (t *ReadTool) Validate(input string) error {
	// The provided snippet for Validate seems to be for a different tool that parses JSON for an "operation".
	// Assuming the intent is to keep the ReadTool's validation focused on its input (path).
	// If the user intended to change ReadTool's validation to parse JSON, this would be a breaking change
	// to its current behavior. I will keep the original validation logic for ReadTool.
	if strings.TrimSpace(input) == "" {
		return fmt.Errorf("file path cannot be empty")
	}
	return nil
}

func (t *ReadTool) Run(input string) func(chan<- StreamMsg) {
	return func(ch chan<- StreamMsg) {
		path := ParseInput(input, "path", "file_path", "filename", "file")

		data, err := os.ReadFile(path)

		if err != nil {
			ch <- StreamMsg{
				Status:    fmt.Sprintf("Read failed: %v", err),
				LogAction: "read " + input,
				LogResult: "error: " + err.Error(),
				LogNext:   "verify path or permissions",
				Done:      true,
			}
			return
		}

		if len(data) == 0 {
			ch <- StreamMsg{Chunk: "(file is empty)\n"}
		} else {
			scanner := bufio.NewScanner(bytes.NewReader(data))
			for scanner.Scan() {
				ch <- StreamMsg{Chunk: scanner.Text() + "\n"}
			}
		}

		ch <- StreamMsg{
			Status:    fmt.Sprintf("Read %d bytes from %s", len(data), input),
			LogAction: "read " + input,
			LogResult: fmt.Sprintf("%d bytes", len(data)),
			LogNext:   "inspect content",
			Done:      true,
		}
	}
}
