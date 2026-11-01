package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallScriptRunsNonInteractiveAndCleansUp(t *testing.T) {
	requireSudo(t)

	dir := t.TempDir()
	logPath := filepath.Join(dir, "log")
	statePath := filepath.Join(dir, "temp-path")
	bin := filepath.Join(dir, "bin")
	if err := os.Mkdir(bin, 0755); err != nil {
		t.Fatal(err)
	}
	writeStub(t, bin, "git", `#!/bin/bash
echo "git $*" >> "$LOG"
if [ "$1" = "clone" ]; then
  mkdir -p sysc-Go
  pwd > "$STATE"
fi
`)
	writeStub(t, bin, "go", `#!/bin/bash
echo "go $*" >> "$LOG"
if [ "$1" = "build" ]; then
  cat > install-syscgo << 'EOF'
#!/bin/bash
echo "installer $*" >> "$LOG"
EOF
  chmod +x install-syscgo
fi
`)

	cmd := scriptCmd(t, bin, logPath, statePath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("install.sh failed: %v\n%s", err, out)
	}
	log, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(log), "installer --yes") {
		t.Fatalf("installer was not started with --yes\nlog:\n%s\noutput:\n%s", log, out)
	}
	tempPath, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(strings.TrimSpace(string(tempPath))); !os.IsNotExist(err) {
		t.Fatalf("temp dir %s still exists after success", strings.TrimSpace(string(tempPath)))
	}
}

func TestInstallScriptRequiresGit(t *testing.T) {
	requireSudo(t)

	dir := t.TempDir()
	bin := filepath.Join(dir, "bin")
	if err := os.Mkdir(bin, 0755); err != nil {
		t.Fatal(err)
	}
	writeStub(t, bin, "go", "#!/bin/bash\nexit 0\n")

	cmd := scriptCmd(t, bin, filepath.Join(dir, "log"), filepath.Join(dir, "state"))
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected failure without git, output:\n%s", out)
	}
	if !strings.Contains(string(out), "Error: git is not installed") {
		t.Fatalf("missing git error\n%s", out)
	}
}

func TestInstallScriptCleansUpWhenBuildFails(t *testing.T) {
	requireSudo(t)

	dir := t.TempDir()
	logPath := filepath.Join(dir, "log")
	statePath := filepath.Join(dir, "temp-path")
	bin := filepath.Join(dir, "bin")
	if err := os.Mkdir(bin, 0755); err != nil {
		t.Fatal(err)
	}
	writeStub(t, bin, "git", `#!/bin/bash
if [ "$1" = "clone" ]; then
  mkdir -p sysc-Go
  pwd > "$STATE"
fi
`)
	writeStub(t, bin, "go", "#!/bin/bash\nexit 1\n")

	cmd := scriptCmd(t, bin, logPath, statePath)
	_, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected build failure")
	}
	tempPath, readErr := os.ReadFile(statePath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if _, statErr := os.Stat(strings.TrimSpace(string(tempPath))); !os.IsNotExist(statErr) {
		t.Fatalf("temp dir %s left behind after failure", strings.TrimSpace(string(tempPath)))
	}
}

// TestInstallScriptRefusesUserOwnedToolchain is #183: the script runs as root,
// so a git or go the invoking user owns must never be executed.
func TestInstallScriptRefusesUserOwnedToolchain(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "bin")
	if err := os.Mkdir(bin, 0755); err != nil {
		t.Fatal(err)
	}
	logPath := filepath.Join(dir, "log")
	writeStub(t, bin, "sudo", `#!/bin/bash
printf '__SYSC_PATH__%s\n' "$TEST_USER_PATH"
`)
	for _, name := range []string{"git", "go"} {
		writeStub(t, bin, name, "#!/bin/bash\necho \"ran $0\" >> \"$TEST_LOG\"\n")
	}

	cmd := exec.Command("bash", unprivilegedScript(t, dir, false))
	cmd.Env = append(os.Environ(),
		"PATH="+isolatedPath(t, bin),
		"SUDO_USER=sysc-test-user",
		"TEST_USER_PATH="+bin,
		"TEST_LOG="+logPath,
	)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("install.sh accepted a user-owned toolchain\n%s", out)
	}
	if !strings.Contains(string(out), "Error: no Go that root can safely run") {
		t.Fatalf("missing untrusted toolchain error\n%s", out)
	}
	if log, err := os.ReadFile(logPath); err == nil {
		t.Fatalf("a user-owned tool ran\nlog:\n%s\noutput:\n%s", log, out)
	}
}

// TestInstallScriptSkipsLinkIntoWritableDirectory checks the resolved file is
// what gets judged: a go symlink in a safe directory that points into a
// world-writable one is passed over for the next trusted go.
func TestInstallScriptSkipsLinkIntoWritableDirectory(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "bin")
	linkDir := filepath.Join(dir, "link")
	shimDir := filepath.Join(dir, "shared")
	realDir := filepath.Join(dir, "real")
	for _, path := range []string{bin, linkDir, shimDir, realDir} {
		if err := os.Mkdir(path, 0755); err != nil {
			t.Fatal(err)
		}
	}
	logPath := filepath.Join(dir, "log")

	writeStub(t, bin, "sudo", `#!/bin/bash
printf '__SYSC_PATH__%s\n' "$TEST_USER_PATH"
`)
	writeStub(t, shimDir, "go", "#!/bin/bash\necho \"shim $*\" >> \"$TEST_LOG\"\n")
	if err := os.Chmod(shimDir, 0777); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(shimDir, "go"), filepath.Join(linkDir, "go")); err != nil {
		t.Fatal(err)
	}
	writeStub(t, realDir, "git", `#!/bin/bash
if [ "$1" = "clone" ]; then mkdir -p sysc-Go; fi
`)
	writeStub(t, realDir, "go", `#!/bin/bash
echo "go $*" >> "$TEST_LOG"
if [ "$1" = "build" ]; then
  cat > install-syscgo << 'EOF'
#!/bin/bash
echo "installer $*" >> "$TEST_LOG"
EOF
  chmod +x install-syscgo
fi
`)

	cmd := exec.Command("bash", unprivilegedScript(t, dir, true))
	cmd.Env = append(os.Environ(),
		"PATH="+isolatedPath(t, bin),
		"SUDO_USER=sysc-test-user",
		"TEST_USER_PATH="+linkDir+":"+realDir,
		"TEST_LOG="+logPath,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("install.sh failed: %v\n%s", err, out)
	}
	log, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(log), "shim") {
		t.Fatalf("the go in a world-writable directory ran\nlog:\n%s", log)
	}
	if !strings.Contains(string(log), "go build") || !strings.Contains(string(log), "installer --yes") {
		t.Fatalf("trusted go did not build and run the installer\nlog:\n%s", log)
	}
	if !strings.Contains(string(out), "Skipping "+filepath.Join(linkDir, "go")) {
		t.Fatalf("skipped link not reported\n%s", out)
	}
}

// unprivilegedScript writes a copy of install.sh into dir with the root check
// removed. trustSelf adds the current uid to the trusted owners, standing in
// for root so files the test creates count as root-owned.
func unprivilegedScript(t *testing.T, dir string, trustSelf bool) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "install.sh"))
	if err != nil {
		t.Fatal(err)
	}
	edits := [][2]string{{`if [ "$EUID" -ne 0 ]; then`, "if false; then"}}
	if trustSelf {
		edits = append(edits, [2]string{`TRUSTED_UIDS="0"`, fmt.Sprintf(`TRUSTED_UIDS="0 %d"`, os.Getuid())})
	}
	script := string(data)
	for _, edit := range edits {
		if !strings.Contains(script, edit[0]) {
			t.Fatalf("install.sh no longer contains %q; update this test's unprivileged harness", edit[0])
		}
		script = strings.Replace(script, edit[0], edit[1], 1)
	}
	path := filepath.Join(dir, "install.sh")
	if err := os.WriteFile(path, []byte(script), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func requireSudo(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("sudo"); err != nil {
		t.Skip("sudo not available")
	}
	if err := exec.Command("sudo", "-n", "true").Run(); err != nil {
		t.Skip("passwordless sudo not available")
	}
}

func scriptCmd(t *testing.T, bin, logPath, statePath string) *exec.Cmd {
	t.Helper()
	script, err := filepath.Abs(filepath.Join("..", "..", "install.sh"))
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("sudo", "-n", "env",
		"PATH="+isolatedPath(t, rootOwnedCopy(t, bin)),
		"LOG="+logPath,
		"STATE="+statePath,
		"bash", script,
	)
	return cmd
}

// rootOwnedCopy copies the stubs in dir into a fresh root-owned directory,
// since install.sh run as root refuses tools anyone else could change.
func rootOwnedCopy(t *testing.T, dir string) string {
	t.Helper()
	out, err := exec.Command("sudo", "-n", "mktemp", "-d").Output()
	if err != nil {
		t.Fatal(err)
	}
	root := strings.TrimSpace(string(out))
	t.Cleanup(func() { _ = exec.Command("sudo", "-n", "rm", "-rf", "--", root).Run() })
	for _, args := range [][]string{
		{"cp", "-r", "--", dir + "/.", root},
		{"chown", "-R", "0:0", "--", root},
		{"chmod", "0755", "--", root},
	} {
		if out, err := exec.Command("sudo", append([]string{"-n"}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("sudo %v: %v\n%s", args, err, out)
		}
	}
	return root
}

// installScriptTools is every external command install.sh reaches for. Keep it
// in step with install.sh: a tool missing here is a 127 that aborts the script
// under set -e, which looks like an install.sh bug rather than a harness gap.
// git and go stay out on purpose so a test can never clone or build for real.
var installScriptTools = []string{
	"bash", "mktemp", "rm", "chmod", "mkdir", "cat", "pwd", "cp", "mv",
	"dirname", "getent", "cut", "sed", "tail", "stat", "readlink",
}

// isolatedPath keeps stub binaries ahead of a small tool directory that does
// not include git, so tests cannot accidentally clone the real repository.
func isolatedPath(t *testing.T, stub string) string {
	t.Helper()
	tools := t.TempDir()
	for _, name := range installScriptTools {
		src, err := exec.LookPath(name)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(src, filepath.Join(tools, name)); err != nil {
			t.Fatal(err)
		}
	}
	return stub + ":" + tools
}

func writeStub(t *testing.T, dir, name, body string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(body), 0755); err != nil {
		t.Fatal(err)
	}
}

// TestIsolatedPathCoversInstallScript guards the tool list the sudo tests skip
// past on hosts without passwordless sudo: it runs install.sh under the same
// isolated PATH without sudo, so a command missing from installScriptTools
// surfaces here as a 127 instead of silently skipping on every other machine.
func TestIsolatedPathCoversInstallScript(t *testing.T) {
	currentUser, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	bin := filepath.Join(dir, "bin")
	goDir := filepath.Join(dir, "go")
	for _, path := range []string{bin, goDir} {
		if err := os.Mkdir(path, 0755); err != nil {
			t.Fatal(err)
		}
	}
	logPath := filepath.Join(dir, "log")

	writeStub(t, bin, "sudo", `#!/bin/bash
printf '__SYSC_PATH__%s\n' "$PATH"
`)
	writeStub(t, bin, "git", `#!/bin/bash
if [ "$1" = "clone" ]; then mkdir -p sysc-Go; fi
`)
	writeStub(t, bin, "curl", `#!/bin/bash
exit 1
`)
	writeStub(t, goDir, "go", `#!/bin/bash
if [ "$1" = "build" ]; then
  cat > install-syscgo << 'INNER'
#!/bin/bash
exit 0
INNER
  chmod +x install-syscgo
fi
`)

	cmd := exec.Command("bash", unprivilegedScript(t, dir, true))
	cmd.Env = append(os.Environ(),
		"PATH="+isolatedPath(t, bin+":"+goDir),
		"SUDO_USER="+currentUser.Username,
		"TEST_LOG="+logPath,
	)
	out, err := cmd.CombinedOutput()
	if strings.Contains(string(out), "command not found") {
		t.Fatalf("isolated PATH is missing a tool install.sh needs; add it to installScriptTools\n%s", out)
	}
	// install.sh exits non-zero here for unrelated reasons (no network, nothing
	// installed), so only the missing-tool signal is fatal here.
	_ = err
}
