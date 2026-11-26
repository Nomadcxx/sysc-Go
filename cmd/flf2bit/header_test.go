package main

import "testing"

// A hostile font header must be refused before it sizes an allocation
// (audit sysc-2373).
func TestParseHeaderBoundsEveryField(t *testing.T) {
	for _, tc := range []struct {
		header string
		ok     bool
	}{
		{"flf2a$ 6 5 16 15 11", true},
		{"flf2a$ 8 6 20 -1 3", true},
		{"flf2a$ -1 0 0 0 0", false},
		{"flf2a$ 0 0 0 0 0", false},
		{"flf2a$ 2000000000 1 1 1 1", false},
		{"flf2a$ six 5 16 15 11", false},
		{"flf2a$ 6 5 16 15 2000000000", false},
		{"flf2a$ 6 5 99999999 15 11", false},
		{"flf2a$", false},
		{"notafont 6 5 16 15 11", false},
	} {
		meta, err := parseHeader(tc.header)
		if (err == nil) != tc.ok {
			t.Errorf("parseHeader(%q) = %+v, %v; want ok=%v", tc.header, meta, err, tc.ok)
		}
	}
}
