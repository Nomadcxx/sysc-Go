// registry.go - Unified effect registry for sysc-Go animations
// This provides metadata about all available effects and enables
// automatic synchronization with consumers like sysc-walls
package animations

const (
	// LibraryVersion is the sysc-Go animations library version
	LibraryVersion = "1.0.3"
)

// EffectMetadata describes an animation effect
type EffectMetadata struct {
	Name         string // Effect name (e.g., "fire", "matrix")
	RequiresText bool   // Whether effect requires text input
	Description  string // Brief description
	VersionAdded string // Version when effect was added
	Category     string // Effect category (e.g., "particle", "text", "abstract")
}

// EffectRegistry contains metadata for all available effects
var EffectRegistry = []EffectMetadata{
	{
		Name:         "matrix",
		RequiresText: false,
		Description:  "Classic Matrix digital rain effect",
		VersionAdded: "1.0.0",
		Category:     "particle",
	},
	{
		Name:         "matrix-art",
		RequiresText: true,
		Description:  "Matrix rain revealing ASCII art",
		VersionAdded: "1.0.0",
		Category:     "text",
	},
	{
		Name:         "fire",
		RequiresText: false,
		Description:  "Doom-style fire effect",
		VersionAdded: "1.0.0",
		Category:     "particle",
	},
	{
		Name:         "fire-text",
		RequiresText: true,
		Description:  "Fire effect with text as negative space",
		VersionAdded: "1.0.1",
		Category:     "text",
	},
	{
		Name:         "fireworks",
		RequiresText: false,
		Description:  "Animated fireworks display",
		VersionAdded: "1.0.0",
		Category:     "particle",
	},
	{
		Name:         "rain",
		RequiresText: false,
		Description:  "Falling rain droplets",
		VersionAdded: "1.0.0",
		Category:     "particle",
	},
	{
		Name:         "rain-art",
		RequiresText: true,
		Description:  "Rain revealing ASCII art",
		VersionAdded: "1.0.0",
		Category:     "text",
	},
	{
		Name:         "beams",
		RequiresText: false,
		Description:  "Light beams crossing the screen",
		VersionAdded: "1.0.0",
		Category:     "abstract",
	},
	{
		Name:         "beam-text",
		RequiresText: true,
		Description:  "Light beams revealing ASCII art",
		VersionAdded: "1.0.0",
		Category:     "text",
	},
	{
		Name:         "ring-text",
		RequiresText: true,
		Description:  "ASCII art with rotating colored rings",
		VersionAdded: "1.0.0",
		Category:     "text",
	},
	{
		Name:         "blackhole",
		RequiresText: true,
		Description:  "Text consumed by an animated blackhole",
		VersionAdded: "1.0.0",
		Category:     "text",
	},
	{
		Name:         "aquarium",
		RequiresText: false,
		Description:  "Animated underwater scene with fish",
		VersionAdded: "1.0.0",
		Category:     "scene",
	},
	{
		Name:         "skull",
		RequiresText: false,
		Description:  "Skull drip animation with ash particles",
		VersionAdded: "1.0.3",
		Category:     "scene",
	},
	{
		Name:         "skull-text",
		RequiresText: true,
		Description:  "Skull drip animation with animated text overlay",
		VersionAdded: "1.0.3",
		Category:     "text",
	},
	{
		Name:         "pour",
		RequiresText: true,
		Description:  "Text pouring onto screen with color transition",
		VersionAdded: "1.0.0",
		Category:     "text",
	},
	{
		Name:         "print",
		RequiresText: true,
		Description:  "Typewriter-style text printing effect",
		VersionAdded: "1.0.0",
		Category:     "text",
	},
	{
		Name:         "decrypt",
		RequiresText: true,
		Description:  "Text decryption/reveal effect",
		VersionAdded: "1.0.0",
		Category:     "text",
	},
}

// GetEffectNames returns all available effect names
func GetEffectNames() []string {
	names := make([]string, len(EffectRegistry))
	for i, effect := range EffectRegistry {
		names[i] = effect.Name
	}
	return names
}

// GetTextBasedEffects returns names of effects that require text input
func GetTextBasedEffects() []string {
	var textEffects []string
	for _, effect := range EffectRegistry {
		if effect.RequiresText {
			textEffects = append(textEffects, effect.Name)
		}
	}
	return textEffects
}

// GetEffectMetadata returns metadata for a specific effect
func GetEffectMetadata(name string) *EffectMetadata {
	for i := range EffectRegistry {
		if EffectRegistry[i].Name == name {
			return &EffectRegistry[i]
		}
	}
	return nil
}

// IsTextBasedEffect checks if an effect requires text input
func IsTextBasedEffect(name string) bool {
	meta := GetEffectMetadata(name)
	return meta != nil && meta.RequiresText
}

// GetLibraryVersion returns the sysc-Go animations library version
func GetLibraryVersion() string {
	return LibraryVersion
}

// ThemeMetadata describes a color theme
type ThemeMetadata struct {
	Name         string   // Theme name (e.g., "nord", "dracula")
	Aliases      []string // Alternative names for the theme
	Description  string   // Brief description
	VersionAdded string   // Version when theme was added
}

// ThemeRegistry contains metadata for all available themes
var ThemeRegistry = []ThemeMetadata{
	{
		Name:         "dracula",
		Aliases:      []string{},
		Description:  "Dracula dark theme with purple and pink accents",
		VersionAdded: "1.0.0",
	},
	{
		Name:         "catppuccin",
		Aliases:      []string{"catppuccin-mocha"},
		Description:  "Catppuccin Mocha - soothing pastel theme",
		VersionAdded: "1.0.0",
	},
	{
		Name:         "nord",
		Aliases:      []string{},
		Description:  "Nord arctic, north-bluish color palette",
		VersionAdded: "1.0.0",
	},
	{
		Name:         "tokyo-night",
		Aliases:      []string{"tokyonight"},
		Description:  "Tokyo Night dark theme inspired by Tokyo",
		VersionAdded: "1.0.0",
	},
	{
		Name:         "gruvbox",
		Aliases:      []string{},
		Description:  "Gruvbox retro groove color scheme",
		VersionAdded: "1.0.0",
	},
	{
		Name:         "material",
		Aliases:      []string{},
		Description:  "Material Design color palette",
		VersionAdded: "1.0.0",
	},
	{
		Name:         "solarized",
		Aliases:      []string{},
		Description:  "Solarized precision colors for machines and people",
		VersionAdded: "1.0.0",
	},
	{
		Name:         "monochrome",
		Aliases:      []string{},
		Description:  "Grayscale monochrome theme",
		VersionAdded: "1.0.0",
	},
	{
		Name:         "transishardjob",
		Aliases:      []string{},
		Description:  "Trans pride colors",
		VersionAdded: "1.0.0",
	},
	{
		Name:         "rama",
		Aliases:      []string{},
		Description:  "Rama custom color scheme",
		VersionAdded: "1.0.0",
	},
	{
		Name:         "eldritch",
		Aliases:      []string{},
		Description:  "Eldritch dark theme with purple and cyan",
		VersionAdded: "1.0.0",
	},
	{
		Name:         "dark",
		Aliases:      []string{},
		Description:  "Simple dark theme with grayscale",
		VersionAdded: "1.0.0",
	},
}

// GetThemeNames returns all available theme names (including aliases)
func GetThemeNames() []string {
	var names []string
	for _, theme := range ThemeRegistry {
		names = append(names, theme.Name)
		names = append(names, theme.Aliases...)
	}
	return names
}

// GetThemeMetadata returns metadata for a specific theme
func GetThemeMetadata(name string) *ThemeMetadata {
	for i := range ThemeRegistry {
		if ThemeRegistry[i].Name == name {
			return &ThemeRegistry[i]
		}
		// Check aliases
		for _, alias := range ThemeRegistry[i].Aliases {
			if alias == name {
				return &ThemeRegistry[i]
			}
		}
	}
	return nil
}
