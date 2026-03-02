// Package animations provides terminal animation effects for Go applications.
//
// sysc-Go includes multiple animation effects with customizable color themes:
//   - Fire: DOOM PSX-style fire effect
//   - Matrix: Digital rain with falling character streaks
//   - Fireworks: Physics-based particle fireworks
//   - Rain: ASCII character rain effect
//
// Basic usage:
//
//	palette := animations.GetFirePalette("dracula")
//	fire := animations.NewFireEffect(80, 24, palette)
//
//	for frame := 0; frame < 200; frame++ {
//	    fire.Update(frame)
//	    output := fire.Render()
//	    fmt.Print(output)
//	}
//
// See GUIDE.md for detailed usage examples and integration patterns.
package animations

// Animation interface that all effects implement
type Animation interface {
	// Update advances the animation by one frame
	Update()

	// Render returns the current frame as a string
	Render() string

	// Reset restarts the animation from the beginning
	Reset()
}

// Config holds common animation settings
type Config struct {
	Width  int    // Terminal width in characters
	Height int    // Terminal height in characters
	Theme  string // Color theme name
}

// TextUpdatable is implemented by effects that can change their displayed text at runtime.
type TextUpdatable interface {
	SetText(text string)
}

// IsTextUpdatable reports whether an Animation supports dynamic text changes.
func IsTextUpdatable(anim Animation) bool {
	_, ok := anim.(TextUpdatable)
	return ok
}
