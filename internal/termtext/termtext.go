// Package termtext makes untrusted strings safe to write to a terminal.
package termtext

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Printable returns s with every C0 and C1 control character, DEL and invalid
// UTF-8 byte shown as a visible \xNN escape. A file name can contain ESC, so
// printing one raw lets a downloaded art pack set the terminal title, write
// the clipboard (OSC 52) or move the cursor (audit sysc-2372).
func Printable(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		switch {
		case r == utf8.RuneError && size == 1:
			fmt.Fprintf(&b, `\x%02x`, s[i])
		case r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f):
			fmt.Fprintf(&b, `\x%02x`, r)
		default:
			b.WriteString(s[i : i+size])
		}
		i += size
	}
	return b.String()
}
