//go:build darwin || netbsd || freebsd || openbsd || linux || dragonfly || solaris

package termcell

import "github.com/charmbracelet/x/termios"

// Apply uses pixel geometry when the terminal supplies it. Non-TTY descriptors
// and terminals without pixel dimensions leave the effect's aspect unchanged.
func Apply(fd int, effect any) {
	if _, ok := effect.(interface{ SetCellAspect(int, int) }); !ok {
		return
	}
	size, err := termios.GetWinsize(fd)
	if err == nil {
		applySize(effect, int(size.Col), int(size.Row), int(size.Xpixel), int(size.Ypixel))
	}
}
