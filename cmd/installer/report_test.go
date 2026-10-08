package main

import (
	"context"
	"testing"
)

// The step alone decides the outcome: stepInstalling means the run ended before
// finishing, stepComplete means it ran to the end. Only then can a task have
// failed, and an optional skip is not a failure.
func TestReportExitCode(t *testing.T) {
	tests := []struct {
		name  string
		step  installStep
		tasks []installTask
		want  int
	}{
		{"cancelled mid-run", stepInstalling, nil, 130},
		{"required task failed", stepComplete, []installTask{{status: statusFailed}}, 1},
		{"optional skip is not a failure", stepComplete, []installTask{{status: statusSkipped}}, 0},
		{"every task complete", stepComplete, []installTask{{status: statusComplete}, {status: statusSkipped}}, 0},
		{"quit before starting", stepWelcome, nil, 0},
	}
	for _, tt := range tests {
		m := newModel()
		m.step = tt.step
		m.tasks = tt.tasks
		if got := m.report(); got != tt.want {
			t.Errorf("%s: report() = %d, want %d", tt.name, got, tt.want)
		}
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
