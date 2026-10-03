package animations

import "testing"

func TestCracktroHandlesNonPositiveDimensions(t *testing.T) {
	for _, dimensions := range []struct {
		width, height int
	}{
		{width: 80, height: 0},
		{width: 80, height: -1},
		{width: 0, height: 24},
	} {
		c := NewCracktroEffect(dimensions.width, dimensions.height, nil, "")
		if len(c.stars) != 0 {
			t.Fatalf("constructor initialized %d stars for dimensions %dx%d", len(c.stars), dimensions.width, dimensions.height)
		}

		c.Resize(dimensions.width, dimensions.height)
		if len(c.stars) != 0 {
			t.Fatalf("resize initialized %d stars for height %d", len(c.stars), dimensions.height)
		}
	}
}
