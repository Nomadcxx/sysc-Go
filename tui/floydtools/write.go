package floydtools

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func init() {
	Register("write", func() Tool { return &WriteTool{} })
}

// WriteTool creates or overwrites files.
type WriteTool struct{}

func (t *WriteTool) Name() string { return "write" }
func (t *WriteTool) Description() string {
	return "Create or overwrite a file. Enter path, then content (Ctrl+D to save)."
}
func (t *WriteTool) FrameDelay() time.Duration { return 0 }

func (t *WriteTool) Validate(input string) error {
	if strings.TrimSpace(input) == "" {
		return fmt.Errorf("file path cannot be empty")
	}
	return nil
}

// Run executes a write operation. The input format is: "path:content"
// For multi-line content, the tool reads from stdin after the path.
func (t *WriteTool) Run(input string) func(chan<- StreamMsg) {
	return func(ch chan<- StreamMsg) {
		// Try parsing as JSON first (standard agent behavior)
		path := ParseInput(input, "path", "file_path", "filename", "file")
		content := ParseInput(input, "content", "code", "text", "body")

		// If path is still the same as input, it wasn't JSON or didn't match.
		// Fallback to legacy path:content format
		if path == strings.TrimSpace(input) {
			parts := strings.SplitN(input, ":", 2)
			path = strings.TrimSpace(parts[0])
			if len(parts) > 1 {
				content = parts[1]
			} else {
				content = ""
			}
		}

		// Write the file
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			ch <- StreamMsg{
				Status:    fmt.Sprintf("Write failed: %v", err),
				LogAction: "write " + path,
				LogResult: "error: " + err.Error(),
				LogNext:   "check permissions and path",
				Done:      true,
			}
			return
		}

		bytesWritten := len(content)

		ch <- StreamMsg{
			Chunk: fmt.Sprintf("Wrote %d bytes to %s\n", bytesWritten, path),
		}

		ch <- StreamMsg{
			Status:    fmt.Sprintf("Successfully wrote %d bytes to %s", bytesWritten, path),
			LogAction: "write " + path,
			LogResult: fmt.Sprintf("%d bytes written", bytesWritten),
			LogNext:   "verify file contents",
			Done:      true,
		}
	}
}
