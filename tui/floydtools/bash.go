package floydtools

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func init() {
	Register("bash", func() Tool { return &BashTool{} })
}

// BashTool executes shell commands.
type BashTool struct{}

func (t *BashTool) Name() string { return "bash" }
func (t *BashTool) Description() string {
	return "Run shell commands (bash -lc) and capture stdout/stderr."
}
func (t *BashTool) FrameDelay() time.Duration { return 20 * time.Millisecond }

func (t *BashTool) Validate(input string) error {
	if strings.TrimSpace(input) == "" {
		return fmt.Errorf("command cannot be empty")
	}
	return nil
}

func (t *BashTool) Run(input string) func(chan<- StreamMsg) {
	return func(ch chan<- StreamMsg) {
		cmdStr := ParseInput(input, "command", "input", "code", "cmd")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		cmd := exec.CommandContext(ctx, "bash", "-lc", cmdStr)

		var output bytes.Buffer
		cmd.Stdout = &output
		cmd.Stderr = &output

		err := cmd.Run()
		outStr := output.String()

		if strings.TrimSpace(outStr) != "" {
			scanner := bufio.NewScanner(strings.NewReader(outStr))
			for scanner.Scan() {
				ch <- StreamMsg{Chunk: scanner.Text() + "\n"}
				time.Sleep(t.FrameDelay())
			}
		}

		status := "completed"
		logResult := "success"
		if err != nil {
			status = fmt.Sprintf("Error: %v", err)
			logResult = "error: " + err.Error()
		}

		ch <- StreamMsg{
			Status:    status,
			LogAction: "bash " + input,
			LogResult: logResult,
			LogNext:   "review output",
			Done:      true,
		}
	}
}
