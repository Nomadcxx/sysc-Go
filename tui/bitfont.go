package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// BitFont represents a bitmap font loaded from a .bit JSON file
type BitFont struct {
	Name       string              `json:"name"`
	Author     string              `json:"author"`
	License    string              `json:"license"`
	Characters map[string][]string `json:"characters"`
}

// LoadBitFont loads a .bit font file from the given path
func LoadBitFont(path string) (*BitFont, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read font file: %w", err)
	}

	var font BitFont
	if err := json.Unmarshal(data, &font); err != nil {
		return nil, fmt.Errorf("failed to parse font JSON: %w", err)
	}

	// Validate font
	if font.Name == "" {
		return nil, fmt.Errorf("font name is required")
	}
	if len(font.Characters) == 0 {
		return nil, fmt.Errorf("font must have at least one character")
	}

	return &font, nil
}

// ListAvailableFonts returns a list of .bit font files from the assets/fonts directory
func ListAvailableFonts() []string {
	var fonts []string

	// Try multiple paths - prioritize system-wide install locations
	searchPaths := []string{
		"assets/fonts",                  // Relative to working directory (dev mode)
		"/usr/local/share/syscgo/fonts", // System-wide install (preferred)
		"/usr/share/syscgo/fonts",       // System-wide install (alternative)
		filepath.Join(os.Getenv("HOME"), ".local", "share", "syscgo", "fonts"), // User local
	}

	for _, basePath := range searchPaths {
		entries, err := os.ReadDir(basePath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".bit") {
				// Remove .bit extension for display
				fontName := strings.TrimSuffix(entry.Name(), ".bit")
				fonts = append(fonts, fontName)
			}
		}

		// If we found fonts, stop searching
		if len(fonts) > 0 {
			break
		}
	}

	return fonts
}

// FindFontPath returns the full path to a font file by name
func FindFontPath(fontName string) (string, error) {
	filename := fontName
	if !strings.HasSuffix(filename, ".bit") {
		filename += ".bit"
	}

	// Try multiple paths - prioritize system-wide install locations
	searchPaths := []string{
		"assets/fonts",                  // Relative to working directory (dev mode)
		"/usr/local/share/syscgo/fonts", // System-wide install (preferred)
		"/usr/share/syscgo/fonts",       // System-wide install (alternative)
		filepath.Join(os.Getenv("HOME"), ".local", "share", "syscgo", "fonts"), // User local
	}

	for _, basePath := range searchPaths {
		fullPath := filepath.Join(basePath, filename)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, nil
		}
	}

	return "", fmt.Errorf("font not found: %s", fontName)
}

// RenderText converts a string to ASCII art using this font
func (f *BitFont) RenderText(text string) []string {
	if text == "" {
		return []string{}
	}

	// Split into lines
	inputLines := strings.Split(text, "\n")
	var outputLines []string

	fontHeight := f.GetHeight()

	for _, line := range inputLines {
		// Initialize output lines for this input line
		lineOutput := make([]string, fontHeight)

		// Process each character
		for _, char := range line {
			charStr := string(char)

			// Get character glyph
			glyph, ok := f.Characters[charStr]
			if !ok {
				// Use space for unknown characters
				glyph, ok = f.Characters[" "]
				if !ok {
					// Create empty glyph if space not defined
					glyph = make([]string, fontHeight)
					for i := range glyph {
						glyph[i] = "  " // Two spaces width
					}
				}
			}

			// Append character to output lines
			for i := 0; i < fontHeight && i < len(glyph); i++ {
				lineOutput[i] += glyph[i]
			}
			// Fill remaining lines if glyph is shorter
			for i := len(glyph); i < fontHeight; i++ {
				lineOutput[i] += strings.Repeat(" ", f.GetCharWidth(char))
			}
		}

		// Add this input line's output to result
		outputLines = append(outputLines, lineOutput...)
	}

	return outputLines
}

// GetHeight returns the height of characters in this font
func (f *BitFont) GetHeight() int {
	// Find the maximum height from any character
	maxHeight := 0
	for _, glyph := range f.Characters {
		if len(glyph) > maxHeight {
			maxHeight = len(glyph)
		}
	}
	return maxHeight
}

// GetCharWidth returns the width of a character in this font
func (f *BitFont) GetCharWidth(char rune) int {
	charStr := string(char)
	glyph, ok := f.Characters[charStr]
	if !ok || len(glyph) == 0 {
		return 2 // Default width
	}
	return len([]rune(glyph[0]))
}

// GetMaxWidth returns the maximum width needed for the given text
func (f *BitFont) GetMaxWidth(text string) int {
	text = strings.ToUpper(text)
	lines := strings.Split(text, "\n")
	maxWidth := 0

	for _, line := range lines {
		width := 0
		for _, char := range line {
			width += f.GetCharWidth(char)
		}
		if width > maxWidth {
			maxWidth = width
		}
	}

	return maxWidth
}
