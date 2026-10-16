package main

import (
	"flag"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestCLIAvailableEffectsAreAccepted runs the built binary for every name in
// its own "Available:" list and asserts none of them is reported as unknown.
func TestCLIAvailableEffectsAreAccepted(t *testing.T) {
	if effect := os.Getenv("SYSC_CLI_EFFECT"); effect != "" {
		flag.CommandLine = flag.NewFlagSet("syscgo", flag.ExitOnError)
		os.Args = []string{"syscgo", "-effect", effect, "-duration", "1"}
		main()
		return
	}

	probe := exec.Command(os.Args[0], "-test.run=^TestCLIAvailableEffectsAreAccepted$")
	probe.Env = childEnv("SYSC_CLI_EFFECT=__not_an_effect__")
	out, _ := probe.CombinedOutput()
	names := parseAvailableEffects(string(out))
	if len(names) == 0 {
		t.Fatalf("no Available list in output:\n%s", out)
	}
	for _, hidden := range []string{"ak-47-rama", "justice-cross"} {
		for _, name := range names {
			if name == hidden {
				t.Errorf("unfinished effect %q is advertised", hidden)
			}
		}
	}

	for _, name := range names {
		cmd := exec.Command(os.Args[0], "-test.run=^TestCLIAvailableEffectsAreAccepted$")
		cmd.Env = childEnv("SYSC_CLI_EFFECT=" + name)
		out, _ := cmd.CombinedOutput()
		if strings.Contains(string(out), "Unknown effect") {
			t.Errorf("advertised effect %q is rejected as unknown", name)
		}
	}
}

func parseAvailableEffects(out string) []string {
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimRight(line, "\r")
		if rest, ok := strings.CutPrefix(line, "Available: "); ok {
			var names []string
			for _, name := range strings.Split(rest, ",") {
				if name = strings.TrimSpace(name); name != "" {
					names = append(names, name)
				}
			}
			return names
		}
	}
	return nil
}
