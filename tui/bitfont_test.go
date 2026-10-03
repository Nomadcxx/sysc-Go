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
