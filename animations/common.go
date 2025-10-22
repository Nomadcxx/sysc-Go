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
	Width  int
	Height int
	Theme  string
}
