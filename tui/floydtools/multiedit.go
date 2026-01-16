package floydtools

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

func init() {
	Register("multiedit", func() Tool { return &MultiEditTool{} })
}

// MultiEditTool performs multiple string replacements atomically.
type MultiEditTool struct{}

func (t *MultiEditTool) Name() string { return "multiedit" }
func (t *MultiEditTool) Description() string {
	return "Perform multiple string replacements in a single file atomically."
}
func (t *MultiEditTool) FrameDelay() time.Duration { return 0 }

func (t *MultiEditTool) Validate(input string) error {
	if strings.TrimSpace(input) == "" {
		return fmt.Errorf("input cannot be empty")
	}
	return nil
}

// parseEditPairs parses edit pairs in format "file:old1:new1:old2:new2:..."
// For multiline content, use file::old1::new1::old2::new2
func parseEditPairs(input string) (file string, pairs []struct{ old, new string }, err error) {
	// Try JSON parsing first
	var data struct {
		Path  string `json:"path"`
		Edits []struct {
			Old string `json:"old"`
			New string `json:"new"`
		} `json:"edits"`
	}

	if jsonErr := json.Unmarshal([]byte(input), &data); jsonErr == nil && data.Path != "" {
		file = data.Path
		for _, edit := range data.Edits {
			pairs = append(pairs, struct{ old, new string }{old: edit.Old, new: edit.New})
		}
		return
	}

	// Legacy format: file:old1:new1...
	parts := strings.SplitN(input, "::", 2)
	if len(parts) == 2 {
		// Multiline format
		file = strings.TrimSpace(parts[0])
		rest := parts[1]
		// Split by :: to get old1:new1:old2:new2
		editStrings := strings.Split(rest, "::")
		for i := 0; i < len(editStrings)-1; i += 2 {
			if i+1 < len(editStrings) {
				pairs = append(pairs, struct{ old, new string }{
					old: editStrings[i],
					new: editStrings[i+1],
				})
			}
		}
	} else {
		// Simple format: file:old1:new1:old2:new2
		parts = strings.Split(input, ":")
		if len(parts) < 3 {
			err = fmt.Errorf("invalid format")
			return
		}
		file = strings.TrimSpace(parts[0])
		for i := 1; i < len(parts)-1; i += 2 {
			if i+1 < len(parts) {
				pairs = append(pairs, struct{ old, new string }{
					old: parts[i],
					new: parts[i+1],
				})
			}
		}
	}
	return
}

func (t *MultiEditTool) Run(input string) func(chan<- StreamMsg) {
	return func(ch chan<- StreamMsg) {
		file, pairs, err := parseEditPairs(input)
		if err != nil || file == "" || len(pairs) == 0 {
			ch <- StreamMsg{
				Status:    "MultiEdit failed: invalid format (use file:old1:new1:old2:new2 or file::old1::new1::old2::new2)",
				LogAction: "multiedit",
				LogResult: "error: invalid format",
				LogNext:   "check format",
				Done:      true,
			}
			return
		}

		// Read file
		data, err := os.ReadFile(file)
		if err != nil {
			ch <- StreamMsg{
				Status:    fmt.Sprintf("MultiEdit failed: %v", err),
				LogAction: "multiedit " + file,
				LogResult: "error: " + err.Error(),
				LogNext:   "verify file exists",
				Done:      true,
			}
			return
		}

		content := string(data)
		originalContent := content
		replacementCount := 0

		// Apply all edits
		for _, pair := range pairs {
			if strings.Contains(content, pair.old) {
				content = strings.Replace(content, pair.old, pair.new, 1)
				replacementCount++
			}
		}

		if content == originalContent {
			ch <- StreamMsg{
				Status:    "MultiEdit warning: no replacements made (old strings not found)",
				LogAction: "multiedit " + file,
				LogResult: "warning: no matches found",
				LogNext:   "verify old strings exist",
				Done:      true,
			}
			return
		}

		// Write back
		err = os.WriteFile(file, []byte(content), 0644)
		if err != nil {
			ch <- StreamMsg{
				Status:    fmt.Sprintf("MultiEdit failed: %v", err),
				LogAction: "multiedit " + file,
				LogResult: "error: " + err.Error(),
				LogNext:   "check permissions",
				Done:      true,
			}
			return
		}

		ch <- StreamMsg{
			Chunk: fmt.Sprintf("MultiEdit complete: %d replacements in %s\n", replacementCount, file),
		}

		ch <- StreamMsg{
			Status:    fmt.Sprintf("Successfully applied %d edits", replacementCount),
			LogAction: "multiedit " + file,
			LogResult: fmt.Sprintf("%d replacements", replacementCount),
			LogNext:   "verify all changes",
			Done:      true,
		}
	}
}
