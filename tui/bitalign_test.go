package tui

import (
	"strings"
	"testing"
)

func bannerFont() *BitFont {
	return &BitFont{
		Name: "banner-test",
		Characters: map[string][]string{
			"H": {"# #", "###", "# #"},
			"I": {"#", "#", "#"},
			"W": {"####", "####", "####"},
		},
	}
}

func bannerOpts(text string, align, maxWidth int) TUIRenderOptions {
	return TUIRenderOptions{
		Font:      bannerFont(),
		Text:      text,
		Alignment: align,
		Color:     "#FFFFFF",
		Scale:     1,
		MaxWidth:  maxWidth,
	}
}

// inkIndent is the leftmost ink column across rendered rows.
func inkIndent(lines []string) int {
	min := -1
	for _, line := range lines {
		plain := stripANSI(line)
		if strings.TrimSpace(plain) == "" {
			continue
		}
		n := 0
		for _, r := range plain {
			if r != ' ' {
				break
			}
			n++
		}
		if min < 0 || n < min {
			min = n
		}
	}
	return min
}

func TestBitAlignmentPadsSingleLineToCanvas(t *testing.T) {
	const canvas = 40
	left := RenderBitText(bannerOpts("HI", 0, canvas))
	center := RenderBitText(bannerOpts("HI", 1, canvas))
	right := RenderBitText(bannerOpts("HI", 2, canvas))

	if len(left) == 0 || len(center) == 0 || len(right) == 0 {
		t.Fatal("expected rendered banner lines")
	}

	li, ci, ri := inkIndent(left), inkIndent(center), inkIndent(right)
	if !(li < ci && ci < ri) {
		t.Fatalf("canvas alignment indents left=%d center=%d right=%d; want left < center < right", li, ci, ri)
	}

	// Without a canvas wider than the glyphs, a single line cannot move.
	plainLeft := inkIndent(RenderBitText(bannerOpts("HI", 0, 0)))
	plainCenter := inkIndent(RenderBitText(bannerOpts("HI", 1, 0)))
	plainRight := inkIndent(RenderBitText(bannerOpts("HI", 2, 0)))
	if plainLeft != plainCenter || plainCenter != plainRight {
		t.Fatalf("MaxWidth 0 should not pad a single line, got left=%d center=%d right=%d", plainLeft, plainCenter, plainRight)
	}
}

func TestBitPreviewAlignmentFollowsControl(t *testing.T) {
	m := NewModel()
	m.width = 80
	m.bitCurrentFont = bannerFont()
	m.bitTextInput.SetValue("HI")
	m.bitScale = 1

	m.bitAlignment = 0
	m = m.updateBitPreview()
	left := inkIndent(m.bitPreviewLines)

	m.bitAlignment = 1
	m = m.updateBitPreview()
	center := inkIndent(m.bitPreviewLines)

	m.bitAlignment = 2
	m = m.updateBitPreview()
	right := inkIndent(m.bitPreviewLines)

	if !(left < center && center < right) {
		t.Fatalf("preview indents left=%d center=%d right=%d; want left < center < right", left, center, right)
	}
}

func TestBitAlignmentStillPadsNarrowerSiblingLine(t *testing.T) {
	// MaxWidth 0 keeps the old rule: pad a short line toward the widest sibling.
	center := RenderBitText(bannerOpts("W\nI", 1, 0))
	if inkIndent(center) != 0 {
		t.Fatalf("wider line indent = %d, want 0", inkIndent(center))
	}

	// The narrow second glyph is the only row whose every line is a single column
	// of ink after the wider block. Find a row whose ink starts after column 0.
	shifted := false
	for _, line := range center {
		plain := strings.TrimRight(stripANSI(line), " ")
		if strings.TrimSpace(plain) == "" {
			continue
		}
		if strings.TrimLeft(plain, " ") == "#" && inkOf(plain) > 0 {
			shifted = true
			break
		}
	}
	if !shifted {
		t.Fatalf("narrower line was not centered against its sibling:\n%s", strings.Join(center, "\n"))
	}
}

func inkOf(plain string) int {
	n := 0
	for _, r := range plain {
		if r != ' ' {
			break
		}
		n++
	}
	return n
}
