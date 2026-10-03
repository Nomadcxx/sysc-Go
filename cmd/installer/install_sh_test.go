package main

import (
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

func TestInstallScriptPassesInvokingUsersMiseDataDir(t *testing.T) {
	currentUser, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	bin := filepath.Join(dir, "bin")
	goDir := filepath.Join(dir, "mise-shims")
	for _, path := range []string{bin, goDir} {
		if err := os.Mkdir(path, 0755); err != nil {
			t.Fatal(err)
		}
	}
	logPath := filepath.Join(dir, "log")
	miseDataDir := filepath.Join(dir, "mise-data")

	writeStub(t, bin, "sudo", `#!/bin/bash
printf '__SYSC_PATH__%s\n' "$TEST_USER_PATH"
printf '__SYSC_MISE_DATA_DIR__%s\n' "$TEST_MISE_DATA_DIR"
`)
	writeStub(t, bin, "git", `#!/bin/bash
if [ "$1" = "clone" ]; then mkdir -p sysc-Go; fi
`)
	writeStub(t, goDir, "go", `#!/bin/bash
printf 'go MISE_DATA_DIR=%s args=%s\n' "$MISE_DATA_DIR" "$*" >> "$TEST_LOG"
if [ "$1" = "build" ]; then
  cat > install-syscgo << 'EOF'
#!/bin/bash
printf 'installer MISE_DATA_DIR=%s args=%s\n' "$MISE_DATA_DIR" "$*" >> "$TEST_LOG"
EOF
  chmod +x install-syscgo
fi
`)

	script, err := os.ReadFile(filepath.Join("..", "..", "install.sh"))
	if err != nil {
		t.Fatal(err)
	}
	rootCheck := `if [ "$EUID" -ne 0 ]; then`
	if !strings.Contains(string(script), rootCheck) {
		t.Fatal("install.sh root check changed; update this test's unprivileged harness")
	}
	testScript := filepath.Join(dir, "install.sh")
	if err := os.WriteFile(testScript, []byte(strings.Replace(string(script), rootCheck, "if false; then", 1)), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("bash", testScript)
	cmd.Env = append(os.Environ(),
		"PATH="+goDir+":"+bin+":"+os.Getenv("PATH"),
		"SUDO_USER="+currentUser.Username,
		"TEST_USER_PATH="+goDir+":"+bin,
		"TEST_MISE_DATA_DIR="+miseDataDir,
		"TEST_LOG="+logPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("install.sh failed: %v\n%s", err, out)
	}

	log, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(log), "go MISE_DATA_DIR="+miseDataDir) || !strings.Contains(string(log), "installer MISE_DATA_DIR="+miseDataDir) {
		t.Fatalf("user MISE_DATA_DIR not passed to both Go invocations\nlog:\n%s", log)
	}
	if !strings.Contains(string(log), "installer MISE_DATA_DIR="+miseDataDir+" args=--yes") {
		t.Fatalf("installer did not run with --yes\nlog:\n%s", log)
	}
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
		"PATH="+isolatedPath(t, bin),
		"LOG="+logPath,
		"STATE="+statePath,
		"bash", script,
	)
	return cmd
}

// isolatedPath keeps stub binaries ahead of a small tool directory that does
// not include git, so tests cannot accidentally clone the real repository.
func isolatedPath(t *testing.T, stub string) string {
	t.Helper()
	tools := t.TempDir()
	for _, name := range []string{"bash", "mktemp", "rm", "chmod", "mkdir", "cat", "pwd", "cp", "mv", "basename", "dirname"} {
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
