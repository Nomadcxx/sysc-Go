package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CreateAnimationPreview creates a static preview of the animation config
func CreateAnimationPreview(animName, theme, file string, width, height int) string {
	var preview strings.Builder

	// Center the preview text
	centerY := height / 2
	emptyLines := centerY - 6

	for i := 0; i < emptyLines; i++ {
		preview.WriteString("\n")
	}

	// Animation info
	preview.WriteString(centerText("╔════════════════════════════════╗", width))
	preview.WriteString("\n")
	preview.WriteString(centerText("║    ANIMATION READY TO START    ║", width))
	preview.WriteString("\n")
	preview.WriteString(centerText("╚════════════════════════════════╝", width))
	preview.WriteString("\n\n")

	preview.WriteString(centerText(fmt.Sprintf("Animation: %s", animName), width))
	preview.WriteString("\n")
	preview.WriteString(centerText(fmt.Sprintf("Theme: %s", theme), width))
	preview.WriteString("\n")
	preview.WriteString(centerText(fmt.Sprintf("File: %s", file), width))
	preview.WriteString("\n\n")

	preview.WriteString(centerText("Press ENTER to launch animation", width))
	preview.WriteString("\n")
	preview.WriteString(centerText("(opens in new view)", width))

	return preview.String()
}

// centerText centers text within the given width
func centerText(text string, width int) string {
	// Remove ANSI codes for length calculation
	plainText := stripANSI(text)
	textLen := len(plainText)

	if textLen >= width {
		return text
	}

	padding := (width - textLen) / 2
	return strings.Repeat(" ", padding) + text
}

// stripANSI removes ANSI escape codes from text
func stripANSI(text string) string {
	// Simple ANSI stripper - good enough for our purposes
	inEscape := false
	var result strings.Builder

	for _, ch := range text {
		if ch == '\033' {
			inEscape = true
		} else if inEscape {
			if ch == 'm' {
				inEscape = false
			}
		} else {
			result.WriteRune(ch)
		}
	}

	return result.String()
}

// LaunchAnimation launches the actual animation in the CLI
func LaunchAnimation(animName, theme, file, duration string) error {
	// Find syscgo binary
	syscgoPath := findSyscgoBinary()
	if syscgoPath == "" {
		return fmt.Errorf("could not find syscgo binary")
	}

	// Build command arguments
	args := []string{
		"-effect", animName,
		"-theme", theme,
	}

	// Add file if it's for a text-based animation
	if animationNeedsFile(animName) {
		args = append(args, "-file", getAssetPath(file))
	}

	// Add duration
	if duration != "infinite" {
		args = append(args, "-duration", duration)
	}

	// Create command
	cmd := exec.Command(syscgoPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Run animation
	return cmd.Run()
}

// findSyscgoBinary locates the syscgo binary
func findSyscgoBinary() string {
	// Try multiple locations
	locations := []string{
		"./syscgo",              // Current directory
		"../syscgo",             // Parent directory
		"/usr/local/bin/syscgo", // System install
		"/usr/bin/syscgo",       // System install
		filepath.Join(os.Getenv("HOME"), "sysc-Go", "syscgo"), // Home directory
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc
		}
	}

	// Try PATH
	if path, err := exec.LookPath("syscgo"); err == nil {
		return path
	}

	return ""
}
