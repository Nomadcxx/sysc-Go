package floydtools

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

func init() {
	Register("ls", func() Tool { return &LSTool{} })
}

// LSTool lists directory entries.
type LSTool struct{}

func (t *LSTool) Name() string { return "ls" }
func (t *LSTool) Description() string {
	return "List directory entries. Provide absolute or relative path."
}
func (t *LSTool) FrameDelay() time.Duration { return 0 }

func (t *LSTool) Validate(input string) error {
	// Empty input defaults to "." in Run, so it's valid
	return nil
}

func (t *LSTool) Run(input string) func(chan<- StreamMsg) {
	return func(ch chan<- StreamMsg) {
		path := ParseInput(input, "path", "directory", "dir")
		if path == "" {
			path = "."
		}

		entries, err := os.ReadDir(path)
		if err != nil {
			ch <- StreamMsg{
				Status:    fmt.Sprintf("LS failed: %v", err),
				LogAction: "ls " + path,
				LogResult: "error: " + err.Error(),
				LogNext:   "verify path exists",
				Done:      true,
			}
			return
		}

		// Sort entries for consistent output
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Name() < entries[j].Name()
		})

		var output strings.Builder
		for _, entry := range entries {
			info, _ := entry.Info()
			suffix := "  "
			if info.IsDir() {
				suffix = "/"
			}
			output.WriteString(entry.Name())
			output.WriteString(suffix)
		}

		ch <- StreamMsg{Chunk: output.String() + "\n"}

		ch <- StreamMsg{
			Status:    fmt.Sprintf("Listed %d entries in %s", len(entries), path),
			LogAction: "ls " + path,
			LogResult: fmt.Sprintf("%d entries", len(entries)),
			LogNext:   "navigate or inspect entries",
			Done:      true,
		}
	}
}
