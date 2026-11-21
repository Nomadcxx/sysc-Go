package termtext

import "testing"

func TestPrintable(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"SYSC.txt", "SYSC.txt"},
		{"skull ☠ art.txt", "skull ☠ art.txt"},
		{"evil\x1b]0;pwned\x07.txt", `evil\x1b]0;pwned\x07.txt`},
		{"clip\x1b]52;c;ZXZpbA==\x07", `clip\x1b]52;c;ZXZpbA==\x07`},
		{"line\nbreak\r", `line\x0abreak\x0d`},
		{"c1\u009b31m", `c1\x9b31m`},
		{"del\x7f", `del\x7f`},
		{"bad\xffbyte", `bad\xffbyte`},
	} {
		if got := Printable(tc.in); got != tc.want {
			t.Errorf("Printable(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
