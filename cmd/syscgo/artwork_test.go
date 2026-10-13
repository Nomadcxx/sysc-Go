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
	for _, id := range []string{"great-wave", "wasteland", "ak-47", "ak-47-rama", "ak-47-text"} {
		cmd := exec.Command(os.Args[0], "-test.run=^TestArtworkCLI$")
		cmd.Env = append(os.Environ(), "SYSC_ARTWORK_CLI_CHECK="+id)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%s: %v: %s", id, err, out)
		}
		if !strings.Contains(string(out), "\x1b[38;2;") || strings.Contains(string(out), "Unknown effect") {
			t.Fatalf("%s did not render", id)
		}
	}
	for _, id := range []string{"justice-cross", "cyberpunk-girl", "ak-47-text"} {
		p := filepath.Join(t.TempDir(), "invalid.txt")
		if err := os.WriteFile(p, []byte("\x1b[31mX"), 0600); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(os.Args[0], "-test.run=^TestArtworkCLI$")
		cmd.Env = append(os.Environ(), "SYSC_ARTWORK_CLI_CHECK="+id, "SYSC_ARTWORK_FILE="+p)
		if out, err := cmd.CombinedOutput(); err == nil {
			t.Fatalf("accepted %s: %s", id, out)
		}
	}
}
