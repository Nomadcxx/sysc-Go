package floydtools

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func init() {
	Register("edit", func() Tool { return &EditTool{} })
}

// EditTool performs exact string replacement in a file.
type EditTool struct{}

func (t *EditTool) Name() string              { return "edit" }
func (t *EditTool) Description() string       { return "Performs exact string replacement in a file." }
func (t *EditTool) FrameDelay() time.Duration { return 0 }

func (t *EditTool) Validate(input string) error {
	parts := strings.SplitN(input, ":", 3)
	if len(parts) < 3 {
		return fmt.Errorf("format: file:old_string:new_string (use :: for multiline)")
	}
	if strings.TrimSpace(parts[0]) == "" {
		return fmt.Errorf("file path cannot be empty")
	}
	return nil
}

func (t *EditTool) Run(input string) func(chan<- StreamMsg) {
	return func(ch chan<- StreamMsg) {
		// New robust JSON parsing
		file := ParseInput(input, "path", "file_path", "filename", "file")
		oldStr := ParseInput(input, "old_str", "oldStr", "target")
		newStr := ParseInput(input, "new_str", "newStr", "replacement")

		// If it wasn't JSON or didn't match, fallback to legacy colon format
		if file == strings.TrimSpace(input) {
			parts := strings.SplitN(input, "::", 3)
			if len(parts) == 3 {
				file, oldStr, newStr = parts[0], parts[1], parts[2]
			} else {
				parts = strings.SplitN(input, ":", 3)
				if len(parts) != 3 {
					ch <- StreamMsg{
						Status:    "Edit failed: invalid format (use JSON or file:old:new)",
						LogAction: "edit " + input,
						Done:      true,
					}
					return
				}
				file, oldStr, newStr = parts[0], parts[1], parts[2]
			}
		}

		file = strings.TrimSpace(file)

		// Read file
		data, err := os.ReadFile(file)
		if err != nil {
			ch <- StreamMsg{
				Status:    fmt.Sprintf("Edit failed: %v", err),
				LogAction: "edit " + file,
				LogResult: "error: " + err.Error(),
				LogNext:   "verify file exists",
				Done:      true,
			}
			return
		}

		content := string(data)

		// Check if old string exists
		if !strings.Contains(content, oldStr) {
			ch <- StreamMsg{
				Status:    "Edit failed: old string not found in file",
				LogAction: "edit " + file,
				LogResult: "error: old string not found",
				LogNext:   "verify old string matches exactly",
				Done:      true,
			}
			return
		}

		// Perform replacement
		newContent := strings.Replace(content, oldStr, newStr, 1)

		// Write back
		err = os.WriteFile(file, []byte(newContent), 0644)
		if err != nil {
			ch <- StreamMsg{
				Status:    fmt.Sprintf("Edit failed: %v", err),
				LogAction: "edit " + file,
				LogResult: "error: " + err.Error(),
				LogNext:   "check permissions",
				Done:      true,
			}
			return
		}

		ch <- StreamMsg{
			Chunk: fmt.Sprintf("Edited %s: replaced occurrence\n", file),
		}

		ch <- StreamMsg{
			Status:    "Successfully edited file",
			LogAction: "edit " + file,
			LogResult: "replacement completed",
			LogNext:   "verify changes",
			Done:      true,
		}
	}
}
