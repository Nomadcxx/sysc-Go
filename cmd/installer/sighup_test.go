package main

import (
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

const sighupHelperEnv = "SYSCGO_INSTALLER_SIGHUP_HELPER"

// TestSighupHelper is not a real test. The SIGHUP tests below re-exec this
// binary with sighupHelperEnv set so the install runs in its own process: a
// real SIGHUP can then be delivered and the exit status observed, which is
// impossible from inside the process that is being signalled.
func TestSighupHelper(t *testing.T) {
	mode := os.Getenv(sighupHelperEnv)
	if mode == "" {
		t.Skip("helper process for the SIGHUP installer tests")
	}
	os.Exit(sighupHelperRun(mode))
}

// sighupHelperRun mirrors runInstaller's non-interactive path minus the
// privilege and system-write tasks. The signal wiring, the build cancellation
// and the cleanup are the real ones.
func sighupHelperRun(mode string) int {
	defer cleanupBuildOutput()

	if mode == "nohup" {
		// nohup leaves SIGHUP as SIG_IGN across exec; the installer must not
		// override that, or closing the terminal would kill a detached install.
		signal.Ignore(syscall.SIGHUP)
	}
	ctx, stop := installContext()
	defer stop()

	// Written after the signal context exists, so the test may signal as soon
	// as the file appears without racing the handler registration.
	if pidFile := os.Getenv("SYSCGO_HELPER_PIDFILE"); pidFile != "" {
		_ = os.WriteFile(pidFile, []byte(strconv.Itoa(os.Getpid())), 0o644)
	}

	m := newModel(ctx)
	m.tasks = []installTask{{
		name:        "Build syscgo",
		description: "Building syscgo binary",
		execute:     buildBinary,
		status:      statusRunning,
	}}
	return installExitCode(ctx, runConfiguredTasks(&m))
}

func TestSighupCancelsBuildAndCleansUp(t *testing.T) {
	bin := t.TempDir()
	tmp := t.TempDir()
	buildPidFile := filepath.Join(bin, "build.pid")
	installerPidFile := filepath.Join(bin, "installer.pid")
	writeStubGoSleep(t, bin, buildPidFile, 60)

	cmd := sighupHelperCommand(t, "sighup", bin, tmp, installerPidFile)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	buildPID := waitForFilePID(t, buildPidFile)
	t.Cleanup(func() { stopPID(buildPID) })
	installerPID := waitForFilePID(t, installerPidFile)

	if err := syscall.Kill(installerPID, syscall.SIGHUP); err != nil {
		t.Fatalf("send SIGHUP: %v", err)
	}

	err := cmd.Wait()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("installer wait = %v, want a 130 exit after SIGHUP", err)
	}
	if got := exitErr.ExitCode(); got != 130 {
		t.Fatalf("installer exit = %d, want 130", got)
	}
	if err := waitUntilDead(buildPID); err != nil {
		t.Fatal(err)
	}
	assertNoBuildDir(t, tmp)
}

func TestIgnoredSighupDoesNotCancelInstall(t *testing.T) {
	bin := t.TempDir()
	tmp := t.TempDir()
	buildPidFile := filepath.Join(bin, "build.pid")
	installerPidFile := filepath.Join(bin, "installer.pid")
	writeStubGoSleep(t, bin, buildPidFile, 2)

	cmd := sighupHelperCommand(t, "nohup", bin, tmp, installerPidFile)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	installerPID := waitForFilePID(t, installerPidFile)
	if err := syscall.Kill(installerPID, syscall.SIGHUP); err != nil {
		t.Fatalf("send SIGHUP: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		t.Fatalf("installer exited after an ignored SIGHUP: %v", err)
	}
	buildPID := waitForFilePID(t, buildPidFile)
	t.Cleanup(func() { stopPID(buildPID) })
	if err := waitUntilDead(buildPID); err != nil {
		t.Fatal(err)
	}
	assertNoBuildDir(t, tmp)
}

func sighupHelperCommand(t *testing.T, mode, bin, tmp, installerPidFile string) *exec.Cmd {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(exe, "-test.run=^TestSighupHelper$")
	cmd.Env = append(os.Environ(),
		sighupHelperEnv+"="+mode,
		"PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
		"TMPDIR="+tmp,
		"SYSCGO_HELPER_PIDFILE="+installerPidFile,
	)
	return cmd
}

func writeStubGoSleep(t *testing.T, bin, pidFile string, seconds int) {
	t.Helper()
	path := filepath.Join(bin, "go")
	script := "#!/bin/sh\n" +
		"if [ \"$1\" = \"build\" ]; then\n" +
		"  sleep " + strconv.Itoa(seconds) + " &\n" +
		"  echo $! > " + strconv.Quote(pidFile) + "\n" +
		"  wait\n" +
		"fi\n" +
		"exit 0\n"
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
}

func waitForFilePID(t *testing.T, path string) int {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		body, err := os.ReadFile(path)
		if err == nil {
			if pid, convErr := strconv.Atoi(strings.TrimSpace(string(body))); convErr == nil && pid > 0 {
				return pid
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for pid file %s", path)
	return 0
}

func assertNoBuildDir(t *testing.T, tmp string) {
	t.Helper()
	entries, err := os.ReadDir(tmp)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "syscgo-install-") {
			t.Fatalf("build dir %s left behind under %s", entry.Name(), tmp)
		}
	}
}
