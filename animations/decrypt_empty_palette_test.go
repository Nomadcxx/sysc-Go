package animations

import "testing"

// An exported constructor must not panic on an empty palette a host forwards
// from its config (audit sysc-2371).
func TestNewDecryptEffectSurvivesEmptyPalettes(t *testing.T) {
	for _, cfg := range []DecryptConfig{
		{Width: 20, Height: 5, Text: "hello"},
		{Width: 20, Height: 5, Text: "hello", CiphertextColors: []string{}, FinalGradientStops: []string{}},
	} {
		e := NewDecryptEffect(cfg)
		for i := 0; i < 5; i++ {
			e.Update()
		}
		_ = e.Render()
	}
}
