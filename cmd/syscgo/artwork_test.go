package main

import (
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestArtworkCLI(t *testing.T) {
	if id := os.Getenv("SYSC_ARTWORK_CLI_CHECK"); id != "" {
		flag.CommandLine = flag.NewFlagSet("syscgo", flag.ExitOnError)
		os.Args = []string{"syscgo", "-effect", id, "-theme", "kanagawa", "-duration", "1"}
		if file := os.Getenv("SYSC_ARTWORK_FILE"); file != "" {
			os.Args = append(os.Args, "-file", file)
		}
		main()
		return
	}
	repoSYSC, err := os.ReadFile(filepath.Join("..", "..", "assets", "SYSC.txt"))
	if err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	seed := filepath.Join(home, "sysc-Go", "assets")
	if err := os.MkdirAll(seed, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(seed, "SYSC.txt"), repoSYSC, 0o644); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"great-wave", "wasteland", "ak-47", "ak-47-text"} {
		cmd := exec.Command(os.Args[0], "-test.run=^TestArtworkCLI$")
		cmd.Env = childEnv("SYSC_ARTWORK_CLI_CHECK="+id, "HOME="+home)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%s: %v: %s", id, err, out)
		}
		if !strings.Contains(string(out), "\x1b[38;2;") || strings.Contains(string(out), "Unknown effect") {
			t.Fatalf("%s did not render", id)
		}
	}
	for _, id := range []string{"ak-47-rama", "justice-cross", "cyberpunk-girl", "ak-47-text"} {
		p := filepath.Join(t.TempDir(), "invalid.txt")
		if err := os.WriteFile(p, []byte("\x1b[31mX"), 0600); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(os.Args[0], "-test.run=^TestArtworkCLI$")
		cmd.Env = childEnv("SYSC_ARTWORK_CLI_CHECK="+id, "SYSC_ARTWORK_FILE="+p)
		if out, err := cmd.CombinedOutput(); err == nil {
			t.Fatalf("accepted %s: %s", id, out)
		}
	}
}

// childEnv drops the artwork and home overrides inherited from the developer's
// shell so the default asset search runs hermetically, exactly as on a clean
// runner.
func childEnv(extra ...string) []string {
	env := make([]string, 0, len(os.Environ())+len(extra))
	for _, kv := range os.Environ() {
		if strings.HasPrefix(kv, "HOME=") || strings.HasPrefix(kv, "SYSC_ARTWORK_") {
			continue
		}
		env = append(env, kv)
	}
	return append(env, extra...)
}
