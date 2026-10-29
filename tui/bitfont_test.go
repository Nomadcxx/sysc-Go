package tui

import "testing"

func TestLoadEmbeddedBitFont(t *testing.T) {
	hasBlock := false
	for _, name := range embeddedFontNames() {
		if name == "block" {
			hasBlock = true
			break
		}
	}
	if !hasBlock {
		t.Fatal("embedded font list does not contain block")
	}

	font, err := loadEmbeddedBitFont("block")
	if err != nil {
		t.Fatalf("load embedded font: %v", err)
	}
	if font.Name != "block" || len(font.Characters) == 0 {
		t.Fatalf("loaded invalid font: name=%q characters=%d", font.Name, len(font.Characters))
	}
}

func TestParseBitFontNormalizesFIGletHardblanksOnly(t *testing.T) {
	for _, tc := range []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "FIGlet markers become spaces",
			data: []byte(`{"name":"figlet","license":"FIGlet font license","characters":{"A":["$##$"]}}`),
			want: " ## ",
		},
		{
			name: "other font keeps dollar glyphs",
			data: []byte(`{"name":"custom","license":"MIT","characters":{"A":["$##$"]}}`),
			want: "$##$",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			font, err := parseBitFont(tc.data)
			if err != nil {
				t.Fatal(err)
			}
			if got := font.RenderText("A")[0]; got != tc.want {
				t.Fatalf("rendered glyph = %q, want %q", got, tc.want)
			}
		})
	}
}
