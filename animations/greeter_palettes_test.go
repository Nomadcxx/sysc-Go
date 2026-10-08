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
