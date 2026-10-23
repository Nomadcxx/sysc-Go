//go:build !darwin && !netbsd && !freebsd && !openbsd && !linux && !dragonfly && !solaris

package termcell

// Apply retains the effect's default aspect on platforms without pixel winsize.
func Apply(fd int, effect any) {}
