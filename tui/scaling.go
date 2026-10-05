package tui

import (
	"strings"
)

// Helper function to convert ANSI string to expanded binary representation
// Each ANSI character represents a vertical pair of pixels
func ansiToExpandedBinary(ansiLines []string) [][]int {
	if len(ansiLines) == 0 {
		return [][]int{}
	}

	// Each ANSI line represents 2 rows of pixels
	binary := make([][]int, len(ansiLines)*2)

	for lineIdx, line := range ansiLines {
		runes := []rune(line)
		if len(runes) == 0 {
			continue
		}

		// Create two binary rows for this ANSI line
		topRow := make([]int, len(runes))
		bottomRow := make([]int, len(runes))

		for charIdx, r := range runes {
			switch r {
			case '█': // Full block - both pixels on
				topRow[charIdx] = 1
				bottomRow[charIdx] = 1
			case '▀': // Upper half block - top pixel on
				topRow[charIdx] = 1
				bottomRow[charIdx] = 0
			case '▄': // Lower half block - bottom pixel on
				topRow[charIdx] = 0
				bottomRow[charIdx] = 1
			case ' ': // Space - both pixels off
				topRow[charIdx] = 0
				bottomRow[charIdx] = 0
			default: // Any other glyph is ink, approximated as a full block.
				// ASCII-art fonts (3d-ascii, banner, big, ...) draw with characters
				// like $ _ | / \; clearing them blanked scaled output (#102).
				topRow[charIdx] = 1
				bottomRow[charIdx] = 1
			}
		}

		binary[lineIdx*2] = topRow
		binary[lineIdx*2+1] = bottomRow
	}

	return binary
}

// Helper function to convert expanded binary back to ANSI format
func expandedBinaryToAnsi(binary [][]int) []string {
	if len(binary) == 0 {
		return []string{}
	}

	var result []string
	for y := 0; y < len(binary); y += 2 {
		if len(binary) <= y {
			break
		}

		width := len(binary[y])
		if width == 0 {
			continue
		}

		var line strings.Builder
		for x := range width {
			topPixel := 0
			bottomPixel := 0

			if y < len(binary) && x < len(binary[y]) {
				topPixel = binary[y][x]
			}

			if y+1 < len(binary) && x < len(binary[y+1]) {
				bottomPixel = binary[y+1][x]
			}

			// Convert pixel pair to ANSI block character
			if topPixel == 1 && bottomPixel == 1 {
				line.WriteString("█") // Full block
			} else if topPixel == 1 && bottomPixel == 0 {
				line.WriteString("▀") // Upper half block
			} else if topPixel == 0 && bottomPixel == 1 {
				line.WriteString("▄") // Lower half block
			} else {
				line.WriteString(" ") // Empty space
			}
		}

		result = append(result, line.String())
	}

	return result
}

// scaleCharacter scales a character bitmap using ANSI-aware algorithm
func scaleCharacter(bitmap []string, scaleFactor float64) []string {
	if scaleFactor == 1.0 || len(bitmap) == 0 {
		return bitmap
	}

	// Convert float scale factor to integer multiplier for scaleBitmap
	var actualScaleFactor int
	if scaleFactor == 0.5 {
		actualScaleFactor = -2 // 0.5x scaling (downscale by 2)
	} else if scaleFactor == 2.0 {
		actualScaleFactor = 2 // 2x scaling
	} else if scaleFactor == 4.0 {
		actualScaleFactor = 4 // 4x scaling
	} else {
		return bitmap // Unsupported scale factor, return unchanged
	}

	// Convert ANSI to expanded binary (each ANSI char becomes 2 pixel rows)
	expandedBinary := ansiToExpandedBinary(bitmap)
	if len(expandedBinary) == 0 {
		return bitmap
	}

	// Scale the expanded binary representation
	scaledBinary := scaleBitmap(expandedBinary, actualScaleFactor)

	// Convert back to ANSI representation
	return expandedBinaryToAnsi(scaledBinary)
}

// scaleBitmap implements the exact algorithm from convert.py
func scaleBitmap(bitmap [][]int, scaleFactor int) [][]int {
	if len(bitmap) == 0 || len(bitmap[0]) == 0 {
		return bitmap
	}

	originalHeight := len(bitmap)
	originalWidth := len(bitmap[0])

	if scaleFactor == 1 {
		return bitmap
	} else if scaleFactor > 1 { // Upscaling
		scaledBitmap := make([][]int, 0)
		for _, row := range bitmap {
			scaledRow := make([]int, 0)
			for _, pixel := range row {
				// Replicate pixel horizontally
				for range scaleFactor {
					scaledRow = append(scaledRow, pixel)
				}
			}
			// Replicate row vertically
			for range scaleFactor {
				// Make a copy of the scaled row
				rowCopy := make([]int, len(scaledRow))
				copy(rowCopy, scaledRow)
				scaledBitmap = append(scaledBitmap, rowCopy)
			}
		}
		return scaledBitmap
	} else if scaleFactor < 0 { // Downscaling
		downscaleBy := -scaleFactor
		if downscaleBy == 0 {
			return bitmap
		}

		newHeight := originalHeight / downscaleBy
		newWidth := originalWidth / downscaleBy

		if newHeight == 0 || newWidth == 0 {
			return [][]int{{0}}
		}

		scaledBitmap := make([][]int, newHeight)
		for rNew := range newHeight {
			newRow := make([]int, newWidth)
			for cNew := range newWidth {
				blockHasPixelOn := 0
				for rOrig := rNew * downscaleBy; rOrig < (rNew+1)*downscaleBy; rOrig++ {
					for cOrig := cNew * downscaleBy; cOrig < (cNew+1)*downscaleBy; cOrig++ {
						// Rows are not always as wide as row 0, and an empty
						// source line becomes a nil row. Cells past the end are off.
						if rOrig >= originalHeight || cOrig >= originalWidth {
							continue
						}
						row := bitmap[rOrig]
						if cOrig < len(row) && row[cOrig] == 1 {
							blockHasPixelOn = 1
							break
						}
					}
					if blockHasPixelOn == 1 {
						break
					}
				}
				newRow[cNew] = blockHasPixelOn
			}
			scaledBitmap[rNew] = newRow
		}
		return scaledBitmap
	}

	return bitmap
}
