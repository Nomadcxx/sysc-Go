package tui

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ExportToSyscWalls exports ASCII art to the sysc-walls screensaver daemon.
//
// The function saves the provided content to ~/.local/share/syscgo/walls/filename
// and updates the sysc-walls configuration file at ~/.config/sysc-walls/daemon.conf
// to use the exported artwork.
//
// The filename is automatically sanitized to prevent path traversal attacks.
// Only alphanumeric characters, dots, hyphens, underscores, and spaces are allowed.
// The .txt extension is added automatically if not present.
//
// Files and directories are created with user-only permissions (0600 for files,
// 0700 for directories) to protect user privacy.
//
// Parameters:
//   - filename: The name for the exported file (e.g., "my-art.txt").
//     Path separators and shell metacharacters are automatically stripped.
//   - content: The ASCII art content to export (plain text).
//
// Returns:
//   - nil on success
//   - error if filename is invalid, file cannot be written, or config update fails
//
// Example:
//
//	art := "HELLO\nWORLD"
//	err := ExportToSyscWalls("greeting.txt", art)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Security:
//   - Filenames are sanitized to prevent directory traversal
//   - Only safe characters allowed in filenames
//   - Files created with 0600 permissions (user-only read/write)
//   - Directories created with 0700 permissions (user-only access)
func ExportToSyscWalls(filename, content string) error {
	// Sanitize filename: strip any directory components to prevent path traversal
	filename = filepath.Base(filename)

	// Validate filename is not empty or special directory names
	if filename == "" || filename == "." || filename == ".." {
		return fmt.Errorf("invalid filename: %s", filename)
	}

	// Validate filename contains only safe characters
	// Allow: alphanumeric, hyphens, underscores, dots, spaces
	// Block: shell metacharacters and path separators
	safeFilename, _ := regexp.MatchString(`^[a-zA-Z0-9_. -]+$`, filename)
	if !safeFilename {
		return fmt.Errorf("filename contains unsafe characters: %s", filename)
	}

	// Ensure .txt extension
	if !strings.HasSuffix(filename, ".txt") {
		filename += ".txt"
	}

	// Create walls directory with user-only permissions
	wallsDir := filepath.Join(os.Getenv("HOME"), ".local", "share", "syscgo", "walls")
	if err := os.MkdirAll(wallsDir, 0700); err != nil {
		return fmt.Errorf("failed to create walls directory: %w", err)
	}

	// Build and validate final path
	artPath := filepath.Join(wallsDir, filename)

	// Final safety check: ensure path is within walls directory
	if !strings.HasPrefix(filepath.Clean(artPath), filepath.Clean(wallsDir)) {
		return fmt.Errorf("path traversal detected: %s", filename)
	}

	// Save ASCII art file with user-only permissions
	if err := os.WriteFile(artPath, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to save ASCII art: %w", err)
	}

	// Update daemon.conf
	configPath := filepath.Join(os.Getenv("HOME"), ".config", "sysc-walls", "daemon.conf")
	if err := updateSyscWallsConfig(configPath, artPath); err != nil {
		// Non-fatal - file saved successfully
		return fmt.Errorf("ASCII art saved to %s, but failed to update config: %w", artPath, err)
	}

	return nil
}

// updateSyscWallsConfig updates or creates the sysc-walls daemon config
func updateSyscWallsConfig(configPath, artPath string) error {
	// Create config directory with user-only permissions
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Read existing config or create default
	config := make(map[string]map[string]string)
	config["idle"] = map[string]string{"timeout": "300s", "min_duration": "30s"}
	config["daemon"] = map[string]string{"debug": "false"}
	config["terminal"] = map[string]string{"fullscreen": "true", "kitty": "true"}
	config["animation"] = map[string]string{
		"type":  "beam-text",
		"theme": "dracula",
		"file":  artPath,
		"cycle": "false",
	}

	// If config exists, read and merge
	if data, err := os.ReadFile(configPath); err == nil {
		parseINI(string(data), config)
		// Update animation section with new file
		if config["animation"] == nil {
			config["animation"] = make(map[string]string)
		}
		config["animation"]["file"] = artPath
		config["animation"]["type"] = "beam-text"
		config["animation"]["theme"] = "dracula"
		config["animation"]["cycle"] = "false"
	}

	// Write config
	if err := writeINI(configPath, config); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// parseINI parses a simple INI format config
func parseINI(content string, config map[string]map[string]string) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	var currentSection string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Section header
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line[1 : len(line)-1]
			if config[currentSection] == nil {
				config[currentSection] = make(map[string]string)
			}
			continue
		}

		// Key-value pair
		if currentSection != "" && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			config[currentSection][key] = value
		}
	}
}

// writeINI writes config to INI format
func writeINI(path string, config map[string]map[string]string) error {
	var content strings.Builder

	// Write sections in a specific order
	sectionOrder := []string{"idle", "daemon", "animation", "terminal"}

	for _, section := range sectionOrder {
		if values, ok := config[section]; ok {
			content.WriteString(fmt.Sprintf("[%s]\n", section))
			for key, value := range values {
				content.WriteString(fmt.Sprintf("%s = %s\n", key, value))
			}
			content.WriteString("\n")
			delete(config, section) // Mark as written
		}
	}

	// Write any remaining sections
	for section, values := range config {
		content.WriteString(fmt.Sprintf("[%s]\n", section))
		for key, value := range values {
			content.WriteString(fmt.Sprintf("%s = %s\n", key, value))
		}
		content.WriteString("\n")
	}

	// Write config file with user-only permissions
	return os.WriteFile(path, []byte(content.String()), 0600)
}

// ExportBitArt handles export target selection and saves accordingly.
//
// This function serves as a router for exporting ASCII art to different targets.
// It automatically strips ANSI color codes and ensures the filename has a .txt extension.
//
// Parameters:
//   - filename: The base filename for the export (extension added if missing)
//   - content: The ASCII art content as an array of lines (may contain ANSI codes)
//   - target: The export destination (0 = syscgo assets, 1 = sysc-walls daemon)
//
// Returns:
//   - nil on success
//   - error if the export fails or target is unknown
//
// Example:
//
//	art := []string{"\x1b[31mHello\x1b[0m", "\x1b[32mWorld\x1b[0m"}
//	err := ExportBitArt("greeting", art, 1)
//	if err != nil {
//	    log.Fatal(err)
//	}
func ExportBitArt(filename string, content []string, target int) error {
	// Strip ANSI codes
	plainContent := ""
	for _, line := range content {
		plainContent += stripANSI(line) + "\n"
	}

	// Add .txt extension if not present
	if !strings.HasSuffix(filename, ".txt") {
		filename += ".txt"
	}

	switch target {
	case 0: // syscgo
		return saveToAssets(filename, plainContent)

	case 1: // sysc-walls
		return ExportToSyscWalls(filename, plainContent)

	default:
		return fmt.Errorf("unknown export target: %d", target)
	}
}
