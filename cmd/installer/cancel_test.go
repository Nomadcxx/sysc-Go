package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func isQuit(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	_, ok := cmd().(tea.QuitMsg)
	return ok
}

func TestInstallingHelpOffersCancel(t *testing.T) {
	m := newModel()
	m.step = stepInstalling
	if got := m.getHelpText(); got != "Q/Ctrl+C: Cancel" {
		t.Fatalf("help = %q, want Q/Ctrl+C: Cancel", got)
	}
}

func TestWelcomeAndCompleteStillQuitImmediately(t *testing.T) {
	for _, step := range []installStep{stepWelcome, stepComplete} {
		m := newModel()
		m.step = step
		for _, key := range []tea.KeyMsg{
			{Type: tea.KeyRunes, Runes: []rune{'q'}},
			{Type: tea.KeyRunes, Runes: []rune{'Q'}},
			{Type: tea.KeyCtrlC},
		} {
			_, cmd := m.Update(key)
			if !isQuit(cmd) {
				t.Fatalf("step %d key %s should quit immediately", step, key.String())
			}
		}
	}
}

func TestInstallingCancelStopsWorkAndSkipsLaterTasks(t *testing.T) {
	keys := []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune{'q'}},
		{Type: tea.KeyRunes, Runes: []rune{'Q'}},
		{Type: tea.KeyCtrlC},
	}
	for _, key := range keys {
		t.Run(key.String(), func(t *testing.T) {
			m := newModel()
			m.step = stepInstalling
			m.width = 80
			m.height = 24

			dir := t.TempDir()
			pidFile := filepath.Join(dir, "child.pid")
			// A build-like command: a shell that leaves a grandchild running.
			// Cancelling must kill that child, not only the direct process.
			script := "sleep 60 & echo $! > " + strconv.Quote(pidFile) + "; wait"
			started := make(chan struct{})
			var nextRan bool
			m.currentTaskIndex = 0
			m.tasks = []installTask{
				{
					name:        "Build syscgo",
					description: "Building syscgo binary",
					status:      statusRunning,
					execute: func(m *model) error {
						close(started)
						cmd := commandContext(m, "sh", "-c", script)
						return cmd.Run()
					},
				},
				{
					name: "Install syscgo",
					execute: func(*model) error {
						nextRan = true
						return nil
					},
				},
			}

			done := make(chan tea.Msg, 1)
			go func() {
				done <- executeTask(0, &m)()
			}()

			pid := waitForPID(t, pidFile)
			select {
			case <-started:
			case <-time.After(5 * time.Second):
				stopPID(pid)
				t.Fatal("task did not start")
			}

			updated, cmd := m.Update(key)
			if isQuit(cmd) {
				stopPID(pid)
				t.Fatal("install cancel quit before the in-flight task returned")
			}
			got := updated.(model)
			if !got.cancelled {
				stopPID(pid)
				t.Fatal("cancel key did not mark the install cancelled")
			}
			if got.getHelpText() != "Cancelling..." {
				stopPID(pid)
				t.Fatalf("help = %q, want Cancelling...", got.getHelpText())
			}
			if got.ctx.Err() == nil {
				stopPID(pid)
				t.Fatal("cancel key did not cancel the install context")
			}

			var msg tea.Msg
			select {
			case msg = <-done:
			case <-time.After(5 * time.Second):
				stopPID(pid)
				t.Fatal("in-flight task kept running after cancel")
			}

			if err := waitUntilDead(pid); err != nil {
				stopPID(pid)
				t.Fatal(err)
			}

			updated, cmd = got.Update(msg)
			if !isQuit(cmd) {
				t.Fatal("expected quit after the cancelled task returned")
			}
			if nextRan {
				t.Fatal("later install task ran after cancel")
			}
			final := updated.(model)
			if final.currentTaskIndex != 0 {
				t.Fatalf("current task index = %d, want 0", final.currentTaskIndex)
			}
			if final.step != stepInstalling {
				t.Fatalf("step = %d, want installing until quit", final.step)
			}
		})
	}
}

func TestBuildCancelKillsChildProcesses(t *testing.T) {
	builds := []struct {
		name string
		run  func(*model) error
	}{
		{name: "syscgo", run: buildBinary},
		{name: "syscgo-tui", run: buildTuiBinary},
	}
	for _, tc := range builds {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(cleanupBuildOutput)

			bin := t.TempDir()
			pidFile := filepath.Join(bin, "child.pid")
			writeStubGo(t, bin, pidFile)
			t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

			m := newModel()
			done := make(chan error, 1)
			go func() {
				done <- tc.run(&m)
			}()

			pid := waitForPID(t, pidFile)
			m.cancel()

			var err error
			select {
			case err = <-done:
			case <-time.After(5 * time.Second):
				stopPID(pid)
				t.Fatal("build kept running after cancel")
			}
			if !errors.Is(err, context.Canceled) {
				stopPID(pid)
				t.Fatalf("build err = %v, want context.Canceled", err)
			}
			if err := waitUntilDead(pid); err != nil {
				stopPID(pid)
				t.Fatal(err)
			}
		})
	}
}

func TestCommandContextCompletesWhenNotCancelled(t *testing.T) {
	m := newModel()
	cmd := commandContext(&m, "sh", "-c", "exit 0")
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
}

func TestCancelledInstallDoesNotClobberExistingAssets(t *testing.T) {
	src := t.TempDir()
	mustWrite(t, filepath.Join(src, "a.txt"), "new-a")
	mustWrite(t, filepath.Join(src, "b.txt"), "new-b")

	share := t.TempDir()
	if err := os.MkdirAll(filepath.Join(share, "assets"), 0755); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(share, "assets", "a.txt"), "old-a")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := installAssetFiles(ctx, src, share)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	body, err := os.ReadFile(filepath.Join(share, "assets", "a.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "old-a" {
		t.Fatalf("existing asset = %q, want old-a", body)
	}
	if _, err := os.Stat(filepath.Join(share, "assets", "b.txt")); !os.IsNotExist(err) {
		t.Fatal("cancelled install copied a file that was not there yet")
	}
}

func writeStubGo(t *testing.T, bin, pidFile string) {
	t.Helper()
	path := filepath.Join(bin, "go")
	script := "#!/bin/sh\n" +
		"if [ \"$1\" = \"build\" ]; then\n" +
		"  sleep 60 &\n" +
		"  echo $! > " + strconv.Quote(pidFile) + "\n" +
		"  wait\n" +
		"fi\n" +
		"exit 0\n"
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
}

func waitForPID(t *testing.T, path string) int {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		body, err := os.ReadFile(path)
		if err == nil {
			pid, convErr := strconv.Atoi(strings.TrimSpace(string(body)))
			if convErr == nil && pid > 0 && alive(pid) {
				t.Cleanup(func() { stopPID(pid) })
				return pid
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for pid file %s", path)
	return 0
}

func alive(pid int) bool {
	err := syscall.Kill(pid, 0)
	return err == nil
}

func waitUntilDead(pid int) error {
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if !alive(pid) {
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return fmt.Errorf("child process %d still running after cancel", pid)
}

func stopPID(pid int) {
	if pid <= 0 {
		return
	}
	_ = syscall.Kill(-pid, syscall.SIGKILL)
	_ = syscall.Kill(pid, syscall.SIGKILL)
}

// A task can call something the context cannot interrupt (os.RemoveAll, one
// large WriteFile). The first Q must stay graceful so a compiler or copy is
// not torn down mid-write, but a second press has to be an unconditional exit
// or the TUI is wedged with no key that works.
func TestSecondCancelQuitsEvenWhenTaskIgnoresContext(t *testing.T) {
	keys := []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune{'q'}},
		{Type: tea.KeyRunes, Runes: []rune{'Q'}},
		{Type: tea.KeyCtrlC},
	}
	for _, first := range keys {
		t.Run("first="+first.String(), func(t *testing.T) {
			release := make(chan struct{})
			defer close(release)
			entered := make(chan struct{})

			m := newModel()
			m.step = stepInstalling
			m.width, m.height = 80, 24
			m.currentTaskIndex = 0
			m.tasks = []installTask{{
				name:   "Uninterruptible",
				status: statusRunning,
				execute: func(*model) error {
					close(entered)
					<-release
					return nil
				},
			}}

			go func() { executeTask(0, &m)() }()
			<-entered

			// First press: graceful, no quit.
			updated, cmd := m.Update(first)
			if isQuit(cmd) {
				t.Fatal("first cancel quit before the in-flight task returned")
			}
			cur := updated.(model)
			if !cur.cancelled {
				t.Fatal("first cancel did not mark the install cancelled")
			}

			// Second press, whichever key: must exit.
			for _, second := range keys {
				_, cmd := cur.Update(second)
				if !isQuit(cmd) {
					t.Fatalf("second %s did not quit while a task ignored the cancel", second.String())
				}
			}
		})
	}
}

// The per-entry guards inside the copy loop are the only thing that stops a
// cancel from continuing on to the rest of the asset tree. A context that is
// already cancelled exits at the first guard and never reaches them.
//
// b.txt is a FIFO, so the loop parks inside copyFile reading it. That holds
// the copy open long enough to cancel deterministically instead of racing a
// goroutine against three tiny writes.
func TestCancelMidCopyStopsBeforeTheRestOfTheAssets(t *testing.T) {
	src := t.TempDir()
	share := t.TempDir()
	mustWrite(t, filepath.Join(src, "a.txt"), "a")
	mustWrite(t, filepath.Join(src, "c.txt"), "c")
	if err := syscall.Mkfifo(filepath.Join(src, "b.txt"), 0o600); err != nil {
		t.Skipf("mkfifo unsupported: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ReadDir yields sorted names, so a.txt copies first and the loop then
	// blocks on the b.txt FIFO. Unblock it once the cancel is in place.
	go func() {
		first := filepath.Join(share, "assets", "a.txt")
		for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); {
			if _, err := os.Stat(first); err == nil {
				break
			}
			time.Sleep(time.Millisecond)
		}
		cancel()
		if f, err := os.OpenFile(filepath.Join(src, "b.txt"), os.O_WRONLY, 0o600); err == nil {
			_, _ = f.WriteString("b")
			_ = f.Close()
		}
	}()

	err := installAssetFiles(ctx, src, share)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	if _, err := os.Stat(filepath.Join(share, "assets", "c.txt")); !os.IsNotExist(err) {
		t.Fatal("c.txt was copied after the cancel")
	}
}
