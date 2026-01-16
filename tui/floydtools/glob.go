package floydtools

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func init() {
	Register("glob", func() Tool { return &GlobTool{} })
}

// GlobTool matches files with Go-style patterns.
type GlobTool struct{}

func (t *GlobTool) Name() string { return "glob" }
func (t *GlobTool) Description() string {
	return "Match files with Go-style patterns. Format: pattern [base path]."
}
func (t *GlobTool) FrameDelay() time.Duration { return 0 }

func (t *GlobTool) Validate(input string) error {
	if strings.TrimSpace(input) == "" {
		return fmt.Errorf("pattern cannot be empty")
	}
	return nil
}

func (t *GlobTool) Run(input string) func(chan<- StreamMsg) {
	return func(ch chan<- StreamMsg) {
		// New robust JSON parsing
		pattern := ParseInput(input, "pattern", "query", "glob")
		basePath := ParseInput(input, "path", "directory", "dir", "basePath")

		// If it wasn't JSON or didn't match, fallback to legacy space format
		if pattern == strings.TrimSpace(input) {
			parts := strings.Fields(input)
			if len(parts) == 0 {
				ch <- StreamMsg{
					Status:    "Glob failed: no pattern provided",
					LogAction: "glob",
					Done:      true,
				}
				return
			}
			pattern = parts[0]
			basePath = "."
			if len(parts) > 1 {
				basePath = parts[1]
			}
		}

		if basePath == "" {
			basePath = "."
		}

		matches, err := filepath.Glob(filepath.Join(basePath, pattern))
		if err != nil {
			ch <- StreamMsg{
				Status:    fmt.Sprintf("Glob failed: %v", err),
				LogAction: "glob " + pattern,
				LogResult: "error: " + err.Error(),
				LogNext:   "check pattern syntax",
				Done:      true,
			}
			return
		}

		sort.Strings(matches)

		if len(matches) == 0 {
			ch <- StreamMsg{Chunk: "(no matches)\n"}
		} else {
			for _, match := range matches {
				// Convert to relative path if base is "."
				display := match
				if basePath == "." {
					display = filepath.Base(match)
				}
				ch <- StreamMsg{Chunk: display + "\n"}
			}
		}

		ch <- StreamMsg{
			Status:    fmt.Sprintf("Found %d matches for %s", len(matches), pattern),
			LogAction: "glob " + pattern,
			LogResult: fmt.Sprintf("%d matches", len(matches)),
			LogNext:   "inspect matched files",
			Done:      true,
		}
	}
}
