package animations

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestGreeterThemePalettes(t *testing.T) {
	if names := GetThemeNames(); len(names) == 0 || names[0] != "dracula" {
		t.Fatal("existing first theme changed")
	}
	// sysc-shell dark theme tokens, in order: background, active, primary,
	// secondary, accent, warning, danger, foreground, secondary text, muted.
	themes := []struct {
		name   string
		colors [10]string
	}{
		{"rose-pine", [10]string{"#191724", "#26233a", "#ebbcba", "#9ccfd8", "#31748f", "#ebbcba", "#eb6f92", "#e0def4", "#c8c6db", "#a8a6ba"}},
		{"kanagawa", [10]string{"#1f1f28", "#2a2a37", "#76946a", "#c0a36e", "#7e9cd8", "#76946a", "#c34043", "#c8c093", "#b4ad86", "#999375"}},
		{"noctalia", [10]string{"#070722", "#21215f", "#fff59b", "#a9aefe", "#9bfece", "#fff59b", "#fd4663", "#f3edf7", "#d7d1dd", "#b1adbb"}},
		{"eldritch-abyss", [10]string{"#171928", "#474852", "#2dcc82", "#0396b3", "#8b75d9", "#2dcc82", "#cc5860", "#d8e6e6", "#c1cdcf", "#a2adb1"}},
		{"void", [10]string{"#000000", "#1a1a1a", "#ffffff", "#ffffff", "#808080", "#ffffff", "#999999", "#ffffff", "#e0e0e0", "#b8b8b8"}},
		{"red", [10]string{"#1a1110", "#3d3231", "#f44336", "#f28b82", "#f28b82", "#f44336", "#f2b8b5", "#f1dedc", "#d7c5c4", "#b5a5a3"}},
		{"cyan", [10]string{"#0e1416", "#303637", "#00bcd4", "#4dd0e1", "#4dd0e1", "#00bcd4", "#f2b8b5", "#dee3e5", "#c5cacc", "#a4a9ab"}},
		{"coral", [10]string{"#1a1110", "#3d3231", "#ffb4ab", "#f9dedc", "#ffb4ab", "#ffb4ab", "#f2b8b5", "#f1dedc", "#d7c5c4", "#b5a5a3"}},
		{"pink", [10]string{"#191112", "#3c3233", "#e91e63", "#f8bbd9", "#f8bbd9", "#e91e63", "#f2b8b5", "#f0dee0", "#d6c5c7", "#b4a5a6"}},
		{"ayu", [10]string{"#0b0e14", "#1e222a", "#e6b450", "#aad94c", "#39bae6", "#e6b450", "#d95757", "#d1d1c7", "#b9bab2", "#9a9a95"}},
		{"amber", [10]string{"#17130b", "#39342b", "#ffc107", "#ffd54f", "#ffd54f", "#ffc107", "#f2b8b5", "#ebe1d4", "#d2c8bc", "#b0a79c"}},
		{"blue", [10]string{"#101418", "#32353a", "#42a5f5", "#8ab4f8", "#8ab4f8", "#f8d38a", "#f2b8b5", "#e0e2e8", "#c7c9cf", "#a6a8ae"}},
		{"purple", [10]string{"#141218", "#36343a", "#d0bcff", "#ccc2dc", "#d0bcff", "#f1daad", "#f2b8b5", "#e6e0e9", "#cdc7d0", "#aba6ae"}},
		{"green", [10]string{"#10140f", "#323630", "#4caf50", "#81c995", "#81c995", "#e4ba66", "#f2b8b5", "#e0e4db", "#c7cbc3", "#a6aaa2"}},
		{"orange", [10]string{"#1a120e", "#3d332e", "#ff6d00", "#ffb74d", "#ffb74d", "#ff6d00", "#f2b8b5", "#f0dfd8", "#d6c6c0", "#b4a69f"}},
	}
	getters := []struct {
		name  string
		get   func(string) []string
		slots []int
	}{
		{"fire", GetFirePalette, []int{0, 1, 4, 5, 6, 2, 7}},
		{"matrix", GetMatrixPalette, []int{0, 1, 4, 3, 2, 7}},
		{"particle", GetParticlePalette, []int{2, 3, 4, 7}},
		{"rain", GetRainPalette, []int{2, 3, 4, 9}},
		{"fireworks", GetFireworksPalette, []int{2, 3, 4, 5, 6, 7}},
		{"screensaver", GetScreensaverPalette, []int{0, 2, 3, 4, 5, 7}},
		{"burn", GetBurnPalette, []int{7, 8, 3, 4, 2, 5, 6, 1, 0, 9, 1, 0}},
		{"skull", GetSkullPalette, []int{1, 9, 2, 3, 7, 0, 0}},
		{"cracktro", GetCracktroPalette, []int{1, 9, 7, 2, 3, 7, 0}},
		{"logo", GetLogoPalette, []int{0, 1, 9, 2, 7, 4, 0}},
	}
	for _, theme := range themes {
		t.Run(theme.name, func(t *testing.T) {
			meta := GetThemeMetadata(theme.name)
			if meta == nil || meta.Name != theme.name {
				t.Errorf("theme missing from registry")
			}
			found := false
			for _, name := range GetThemeNames() {
				if name == theme.name {
					found = true
				}
			}
			if !found {
				t.Errorf("theme missing from catalog")
			}
			for _, getter := range getters {
				t.Run(getter.name, func(t *testing.T) {
					want := make([]string, len(getter.slots))
					for i, slot := range getter.slots {
						want[i] = theme.colors[slot]
					}
					got := getter.get(theme.name)
					if !reflect.DeepEqual(got, want) {
						t.Fatalf("palette=%v, want %v", got, want)
					}
					if !reflect.DeepEqual(getter.get(strings.ToUpper(theme.name)), want) {
						t.Fatal("case-insensitive palette changed")
					}
					for _, color := range got {
						if len(color) != 7 || color[0] != '#' {
							t.Fatalf("invalid color %q", color)
						}
						if _, err := strconv.ParseUint(color[1:], 16, 24); err != nil {
							t.Fatal(err)
						}
					}
					got[0] = "#000000"
					if !reflect.DeepEqual(getter.get(theme.name), want) {
						t.Fatal("caller mutation changed palette")
					}
				})
			}
		})
	}
}
