package main

import (
	"context"
	"io"
	"os"
	"testing"
)

// captureOutput runs fn with os.Stdout and os.Stderr redirected to pipes and
// returns what each captured. report() writes only a few bytes, well under the
// pipe buffer, so the reads after closing the write ends cannot block.
func captureOutput(t *testing.T, fn func()) (stdout, stderr string) {
	t.Helper()
	origStdout, origStderr := os.Stdout, os.Stderr
	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	errR, errW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout, os.Stderr = outW, errW
	defer func() {
		os.Stdout, os.Stderr = origStdout, origStderr
	}()

	fn()

	outW.Close()
	errW.Close()
	out, _ := io.ReadAll(outR)
	errOut, _ := io.ReadAll(errR)
	outR.Close()
	errR.Close()
	return string(out), string(errOut)
}

// The step alone decides the outcome: stepWelcome means the user quit before
// anything ran, stepInstalling means the run ended before finishing,
// stepComplete means it ran to the end. Only then can a task have failed, and
// an optional skip is not a failure. The uninstall flag picks the wording.
func TestReport(t *testing.T) {
	tests := []struct {
		name      string
		uninstall bool
		step      installStep
		tasks     []installTask
		errors    []string
		wantCode  int
		wantOut   string
		wantErr   string
	}{
		{"welcome quit", false, stepWelcome, nil, nil, 130, "", "Cancelled. Nothing was changed.\n"},
		{"install cancel", false, stepInstalling, nil, nil, 130, "", "Installation cancelled. Nothing further was changed.\n"},
		{"install failure", false, stepComplete, []installTask{{status: statusFailed}}, []string{"boom"}, 1, "", "Installation failed: boom\n"},
		{"install success", false, stepComplete, []installTask{{status: statusComplete}}, nil, 0, "Installation complete.\n", ""},
		{"optional skip is not a failure", false, stepComplete, []installTask{{status: statusComplete}, {status: statusSkipped}}, nil, 0, "Installation complete.\n", ""},
		{"uninstall cancel", true, stepInstalling, nil, nil, 130, "", "Uninstall cancelled. Nothing further was changed.\n"},
		{"uninstall failure", true, stepComplete, []installTask{{status: statusFailed}}, []string{"boom"}, 1, "", "Uninstall failed: boom\n"},
		{"uninstall success", true, stepComplete, []installTask{{status: statusComplete}}, nil, 0, "Uninstall complete.\n", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newModel()
			m.uninstallMode = tt.uninstall
			m.step = tt.step
			m.tasks = tt.tasks
			m.errors = tt.errors

			var code int
			out, errOut := captureOutput(t, func() { code = m.report() })

			if code != tt.wantCode {
				t.Errorf("report() = %d, want %d", code, tt.wantCode)
			}
			if out != tt.wantOut {
				t.Errorf("stdout = %q, want %q", out, tt.wantOut)
			}
			if errOut != tt.wantErr {
				t.Errorf("stderr = %q, want %q", errOut, tt.wantErr)
			}
		})
	}
}

// runInstaller hands the signal-cancelled context to the model so Ctrl+C stops
// the build group. Without the parent, m.ctx is a plain Background child and
// errIfCancelled never fires.
func TestNewModelInheritsParentCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	m := newModel(ctx)
	cancel()
	if err := errIfCancelled(&m); err == nil {
		t.Fatal("errIfCancelled = nil after parent cancel, want context.Canceled")
	}

	fresh := newModel()
	if err := errIfCancelled(&fresh); err != nil {
		t.Fatalf("errIfCancelled = %v on a model with no parent, want nil", err)
	}
}
