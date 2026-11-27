package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/Nomadcxx/sysc-Go/animations"
)

func TestSyscgoMainHelper(t *testing.T) {
	if os.Getenv("SYSCGO_MAIN_HELPER") != "1" {
		t.Skip("helper process only")
	}
	dir := os.Getenv("SYSCGO_HELPER_DIR")
	assetShareDirs = []string{filepath.Join(dir, "empty-share")}
	os.Args = append([]string{"syscgo"}, strings.Fields(os.Getenv("SYSCGO_ARGS"))...)
	main()
	os.Exit(0)
}

func TestErrorExitsRestoreCursor(t *testing.T) {
	dir := t.TempDir()
	empty := filepath.Join(dir, "empty.txt")
	if err := os.WriteFile(empty, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name        string
		args        string
		wantMessage string
		wantHide    bool
	}{
		{
			name:        "unknown effect fails before setup",
			args:        "-effect bogus",
			wantMessage: "Unknown effect: bogus",
		},
		{
			name:        "missing text assets",
			args:        "-effect fire-text",
			wantMessage: "Could not find SYSC.txt",
			wantHide:    true,
		},
		{
			name:        "missing explicit file",
			args:        "-effect pour -file /nonexistent-syscgo-art.txt",
			wantMessage: "Could not read file",
			wantHide:    true,
		},
		{
			name:        "empty beam text file",
			args:        "-effect beam-text -file " + empty,
			wantMessage: "beam-text effect requires non-empty text (-file)",
			wantHide:    true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestSyscgoMainHelper$")
			cmd.Dir = dir
			cmd.Env = append(os.Environ(),
				"SYSCGO_MAIN_HELPER=1",
				"SYSCGO_ARGS="+test.args,
				"SYSCGO_HELPER_DIR="+dir,
				"HOME="+dir,
				"XDG_DATA_HOME="+dir,
			)
			out, err := cmd.CombinedOutput()
			exitErr, ok := err.(*exec.ExitError)
			if !ok || exitErr.ExitCode() != 1 {
				t.Fatalf("exit = %v, want status 1; output: %q", err, out)
			}
			output := string(out)
			if !strings.Contains(output, test.wantMessage) {
				t.Fatalf("output = %q, want message %q", output, test.wantMessage)
			}
			hide := strings.Index(output, "\033[?25l")
			show := strings.Index(output, "\033[?25h")
			if !test.wantHide {
				if hide >= 0 || strings.Contains(output, "\033[2J") {
					t.Fatalf("output = %q, want no terminal setup before an unknown effect", output)
				}
				return
			}
			if hide < 0 {
				t.Fatalf("output = %q, want cursor hide before the error", output)
			}
			if show < hide {
				t.Fatalf("output = %q, want cursor restore after hide", output)
			}
		})
	}
}

func TestKnownEffectsMatchSwitch(t *testing.T) {
	// Keep this list in sync with main's effect switch. Every name must be
	// accepted by isKnownEffect and exported by the animation registry.
	switchEffects := []string{
		"higgs-field",
		"ak-47-text",
		"ak-47",
		"l2a1-sar",
		"jesus-christ-statue", "jason", "nymphe-der-flora", "lowe", "roman-gladiator", "grabfigur",
		"laocoon-and-his-sons", "napoleon-bust", "brandenburg-gate", "louis-xiv", "engel", "st-george", "santa-muerte", "aspasia", "st-nepomuk", "lion-statue",
		"f-14-tomcat", "f-15e-strike-eagle", "f-22-raptor", "fa-18e-super-hornet",
		"great-wave", "wasteland",
		"fire", "fire-text",
		"matrix", "matrix-art",
		"fireworks",
		"rain", "rain-art",
		"pour",
		"print",
		"decrypt",
		"beams", "beam-text", "ring-text",
		"blackhole",
		"aquarium",
		"skull",
		"sonar",
		"cracktro",
		"sysc-logo", "cross-logo", "logo-morph",
	}
	registered := append([]string(nil), animations.GetEffectNames()...)
	sort.Strings(registered)
	sort.Strings(switchEffects)
	if fmt.Sprint(registered) != fmt.Sprint(switchEffects) {
		t.Fatalf("animation registry = %v, switch = %v; keep them in sync", registered, switchEffects)
	}
	for _, name := range switchEffects {
		if !isKnownEffect(name) {
			t.Errorf("isKnownEffect(%q) = false", name)
		}
	}
	if isKnownEffect("bogus") {
		t.Error("isKnownEffect(bogus) = true")
	}
}
