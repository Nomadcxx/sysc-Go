package floydtools

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func init() {
	Register("grep", func() Tool { return &GrepTool{} })
}

// GrepTool searches files using ripgrep.
type GrepTool struct{}

func (t *GrepTool) Name() string { return "grep" }
func (t *GrepTool) Description() string {
	return "Fast regex search in files using ripgrep. Format: pattern [path] [glob]"
}
func (t *GrepTool) FrameDelay() time.Duration { return 0 }

func (t *GrepTool) Validate(input string) error {
	if strings.TrimSpace(input) == "" {
		return fmt.Errorf("pattern cannot be empty")
	}
	return nil
}

func (t *GrepTool) Run(input string) func(chan<- StreamMsg) {
	return func(ch chan<- StreamMsg) {
		// New robust JSON parsing
		pattern := ParseInput(input, "query", "pattern", "regex")
		path := ParseInput(input, "path", "directory", "dir")
		glob := ParseInput(input, "glob", "include", "pattern")

		// If it wasn't JSON or didn't match, fallback to legacy space format
		if pattern == strings.TrimSpace(input) {
			parts := strings.Fields(input)
			if len(parts) == 0 {
				ch <- StreamMsg{
					Status:    "Grep failed: no pattern provided",
					LogAction: "grep",
					Done:      true,
				}
				return
			}
			pattern = parts[0]
			path = "."
			glob = ""
			if len(parts) > 1 {
				if strings.Contains(parts[1], "*") {
					glob = parts[1]
				} else {
					path = parts[1]
				}
			}
			if len(parts) > 2 {
				glob = parts[2]
			}
		}

		if path == "" {
			path = "."
		}

		// Build rg command
		args := []string{pattern, path}
		if glob != "" {
			args = append(args, "--glob", glob)
		}

		cmd := exec.Command("rg", args...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		output := stdout.String()
		errStr := stderr.String()

		// rg returns exit 1 when no matches found - that's OK
		if err != nil && errStr != "" && !strings.Contains(output, "\n") {
			ch <- StreamMsg{
				Status:    fmt.Sprintf("Grep failed: %v", err),
				LogAction: "grep " + pattern,
				LogResult: "error: " + errStr,
				LogNext:   "check pattern and path",
				Done:      true,
			}
			return
		}

		if strings.TrimSpace(output) == "" {
			ch <- StreamMsg{
				Chunk: "(no matches found)\n",
			}
			ch <- StreamMsg{
				Status:    "No matches found",
				LogAction: "grep " + pattern,
				LogResult: "0 matches",
				LogNext:   "try different pattern or path",
				Done:      true,
			}
			return
		}

		// Stream results
		scanner := bufio.NewScanner(strings.NewReader(output))
		lineCount := 0
		for scanner.Scan() {
			ch <- StreamMsg{Chunk: scanner.Text() + "\n"}
			lineCount++
		}

		ch <- StreamMsg{
			Status:    fmt.Sprintf("Found %d matches", lineCount),
			LogAction: "grep " + pattern,
			LogResult: fmt.Sprintf("%d matches", lineCount),
			LogNext:   "inspect results",
			Done:      true,
		}
	}
}
