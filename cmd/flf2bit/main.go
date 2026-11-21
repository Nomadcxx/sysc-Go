package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// BitFont represents the .bit font format
type BitFont struct {
	Name       string              `json:"name"`
	Author     string              `json:"author"`
	License    string              `json:"license"`
	Characters map[string][]string `json:"characters"`
}

// FIGletFont represents parsed FIGlet font metadata
type FIGletFont struct {
	Signature    string
	Hardblank    rune
	Height       int
	Baseline     int
	MaxLength    int
	OldLayout    int
	CommentLines int
	PrintDir     int
	FullLayout   int
	CodetagCount int
	Comments     []string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: flf2bit <figlet-font.flf> [output.bit]")
		fmt.Println("Converts FIGlet .flf fonts to .bit JSON format")
		os.Exit(1)
	}

	inputPath := os.Args[1]
	outputPath := ""
	if len(os.Args) > 2 {
		outputPath = os.Args[2]
	} else {
		// Auto-generate output name
		base := filepath.Base(inputPath)
		name := strings.TrimSuffix(base, filepath.Ext(base))
		outputPath = name + ".bit"
	}

	fmt.Printf("Converting %s to %s...\n", inputPath, outputPath)

	// Parse FIGlet font
	font, err := parseFIGletFont(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing FIGlet font: %v\n", err)
		os.Exit(1)
	}

	// Write .bit font
	if err := writeBitFont(outputPath, font); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing .bit font: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully converted! %d characters\n", len(font.Characters))
}

func parseFIGletFont(path string) (*BitFont, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Read header line
	if !scanner.Scan() {
		return nil, fmt.Errorf("empty file")
	}

	header := scanner.Text()
	meta, err := parseHeader(header)
	if err != nil {
		return nil, err
	}

	// Skip comment lines
	for i := 0; i < meta.CommentLines; i++ {
		if !scanner.Scan() {
			return nil, fmt.Errorf("unexpected EOF in comments")
		}
		meta.Comments = append(meta.Comments, scanner.Text())
	}

	// Extract font name and author from comments
	fontName := filepath.Base(path)
	fontName = strings.TrimSuffix(fontName, filepath.Ext(fontName))
	author := "Unknown"

	for _, comment := range meta.Comments {
		if strings.Contains(strings.ToLower(comment), "by ") {
			author = strings.TrimSpace(strings.Split(comment, "by ")[1])
			break
		}
	}

	// Create BitFont
	bitFont := &BitFont{
		Name:       fontName,
		Author:     author,
		License:    "See original FIGlet font license",
		Characters: make(map[string][]string),
	}

	// Read character definitions
	// Standard ASCII printable characters: 32-126
	for ascii := 32; ascii <= 126; ascii++ {
		char := string(rune(ascii))
		lines, err := readCharacter(scanner, meta)
		if err != nil {
			// If we can't read a character, skip it
			continue
		}

		if len(lines) > 0 {
			// Clean up lines (remove hardblank, trim trailing spaces)
			cleaned := make([]string, len(lines))
			for i, line := range lines {
				cleaned[i] = strings.ReplaceAll(line, string(meta.Hardblank), " ")
			}
			bitFont.Characters[char] = cleaned
		}
	}

	return bitFont, nil
}

func parseHeader(header string) (*FIGletFont, error) {
	parts := strings.Fields(header)
	if len(parts) < 1 {
		return nil, fmt.Errorf("invalid header")
	}

	signature := parts[0]
	if !strings.HasPrefix(signature, "flf2") {
		return nil, fmt.Errorf("not a FIGlet font file")
	}

	// Extract hardblank from signature
	hardblank := ' '
	if len(signature) > 4 {
		hardblank = rune(signature[4])
	}

	meta := &FIGletFont{
		Signature: signature,
		Hardblank: hardblank,
	}

	// Parse numeric fields. A font is untrusted input (art packs ship them),
	// and Height sizes allocations, so every field is checked before use
	// (audit sysc-2373).
	fields := []struct {
		name     string
		dst      *int
		min, max int
	}{
		{"height", &meta.Height, 1, maxFontHeight},
		{"baseline", &meta.Baseline, 0, maxFontHeight},
		{"max length", &meta.MaxLength, 0, maxFontLineLength},
		{"old layout", &meta.OldLayout, -1, 63},
		{"comment lines", &meta.CommentLines, 0, maxCommentLines},
	}
	for i, f := range fields {
		if len(parts) <= i+1 {
			if f.name == "height" {
				return nil, fmt.Errorf("header %q has no height", header)
			}
			break
		}
		v, err := strconv.Atoi(parts[i+1])
		if err != nil || v < f.min || v > f.max {
			return nil, fmt.Errorf("header %q: %s %q is not between %d and %d", header, f.name, parts[i+1], f.min, f.max)
		}
		*f.dst = v
	}

	return meta, nil
}

// Bounds on a FIGlet header. Real fonts are under 20 rows high with lines
// under a few hundred columns.
const (
	maxFontHeight     = 1024
	maxFontLineLength = 4096
	maxCommentLines   = 100000
)

func readCharacter(scanner *bufio.Scanner, meta *FIGletFont) ([]string, error) {
	lines := make([]string, 0, meta.Height)

	for i := 0; i < meta.Height; i++ {
		if !scanner.Scan() {
			return nil, fmt.Errorf("unexpected EOF reading character")
		}

		line := scanner.Text()

		// Remove end markers (@ or @@)
		line = strings.TrimRight(line, "@")

		// Convert to visual representation
		// FIGlet uses space and hardblank, we want actual characters
		lines = append(lines, line)
	}

	return lines, nil
}

func writeBitFont(path string, font *BitFont) error {
	data, err := json.MarshalIndent(font, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
