package main

import (
	"flag"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestArtworkCLI(t *testing.T) {
	if id := os.Getenv("SYSC_ARTWORK_CLI_CHECK"); id != "" {
		flag.CommandLine = flag.NewFlagSet("syscgo", flag.ExitOnError)
		os.Args = []string{"syscgo", "-effect", id, "-theme", "kanagawa", "-duration", "1"}
		main()
		return
	}
	for _, id := range []string{"great-wave", "cyberpunk-girl", "wasteland"} {
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
}
