package tui

// Gradient direction constants for TUI usage
const (
	GradientUpDown = iota
	GradientDownUp
	GradientLeftRight
	GradientRightLeft
)

// Shadow style constants for TUI usage
const (
	ShadowLight = iota
	ShadowMedium
	ShadowDark
)

// TUIRenderOptions holds simplified configuration for rendering text in the TUI
// This is our wrapper around BIT's RenderOptions
type TUIRenderOptions struct {
	Font          *BitFont
	Text          string
	Alignment     int
	Color         string
	Scale         float64
	Shadow        bool
	ShadowOffsetX int
	ShadowOffsetY int
	ShadowStyle   int
	CharSpacing   int
	WordSpacing   int
	LineSpacing   int
	UseGradient   bool
	GradientColor string
	GradientDir   int
	MaxWidth      int // Canvas width for alignment
}

// RenderOptions is BIT's full rendering options structure
type RenderOptions struct {
	CharSpacing            int
	WordSpacing            int
	LineSpacing            int
	Alignment              TextAlignment
	TextColor              string
	GradientColor          string
	GradientDirection      GradientDirection
	UseGradient            bool
	ScaleFactor            float64
	ShadowEnabled          bool
	ShadowHorizontalOffset int
	ShadowVerticalOffset   int
	ShadowStyle            ShadowStyle
	TextLines              []string
	MaxWidth               int // Canvas width; 0 keeps alignment relative to sibling lines
}

// RenderBitText renders text using a bitmap font with styling options
// This wraps BIT's proven rendering engine
func RenderBitText(opts TUIRenderOptions) []string {
	if opts.Font == nil || opts.Text == "" {
		return []string{}
	}

	// Convert our simplified options to BIT's RenderOptions format
	bitOpts := convertToBITOptions(opts)

	// Use BIT's rendering engine
	fontData := FontData{
		Name:       opts.Font.Name,
		Author:     opts.Font.Author,
		License:    opts.Font.License,
		Characters: opts.Font.Characters,
	}

	return RenderTextWithFont(opts.Text, fontData, bitOpts)
}

// convertToBITOptions converts our TUIRenderOptions to BIT's RenderOptions format
func convertToBITOptions(opts TUIRenderOptions) RenderOptions {
	bitOpts := RenderOptions{
		CharSpacing:            opts.CharSpacing,
		WordSpacing:            opts.WordSpacing,
		LineSpacing:            opts.LineSpacing,
		TextColor:              opts.Color,
		ScaleFactor:            opts.Scale,
		ShadowEnabled:          opts.Shadow,
		ShadowHorizontalOffset: opts.ShadowOffsetX,
		ShadowVerticalOffset:   opts.ShadowOffsetY,
		UseGradient:            opts.UseGradient,
		GradientColor:          opts.GradientColor,
		MaxWidth:               opts.MaxWidth,
	}

	// Default values
	if bitOpts.ScaleFactor == 0 {
		bitOpts.ScaleFactor = 1.0
	}
	if bitOpts.TextColor == "" {
		bitOpts.TextColor = "#FFFFFF"
	}

	// Convert alignment (use the actual BIT alignment constants from alignment.go)
	bitOpts.Alignment = TextAlignment(opts.Alignment)

	// Convert gradient direction
	switch opts.GradientDir {
	case GradientUpDown:
		bitOpts.GradientDirection = UpDown
	case GradientDownUp:
		bitOpts.GradientDirection = DownUp
	case GradientLeftRight:
		bitOpts.GradientDirection = LeftRight
	case GradientRightLeft:
		bitOpts.GradientDirection = RightLeft
	default:
		bitOpts.GradientDirection = UpDown
	}

	// Convert shadow style
	switch opts.ShadowStyle {
	case ShadowLight:
		bitOpts.ShadowStyle = LightShade
	case ShadowMedium:
		bitOpts.ShadowStyle = MediumShade
	case ShadowDark:
		bitOpts.ShadowStyle = DarkShade
	default:
		bitOpts.ShadowStyle = LightShade
	}

	return bitOpts
}

// FontData represents BIT's font structure
type FontData struct {
	Name       string
	Author     string
	License    string
	Characters map[string][]string
}

// TextAlignment from BIT - using the same values as HorizontalAlignment
type TextAlignment int

const (
	LeftAlign TextAlignment = iota
	CenterAlign
	RightAlign
)

// GradientDirection from BIT
type GradientDirection int

const (
	UpDown GradientDirection = iota
	DownUp
	LeftRight
	RightLeft
)

// ShadowStyle from BIT
type ShadowStyle int

const (
	LightShade ShadowStyle = iota
	MediumShade
	DarkShade
)

// GetRenderedDimensions calculates the final dimensions of rendered text
func GetRenderedDimensions(opts TUIRenderOptions) (width, height int) {
	lines := RenderBitText(opts)
	if len(lines) == 0 {
		return 0, 0
	}

	height = len(lines)
	for _, line := range lines {
		// Strip ANSI codes for accurate width
		plainLine := stripANSI(line)
		w := len([]rune(plainLine))
		if w > width {
			width = w
		}
	}

	return width, height
}

// stripANSI is already defined in animation.go
