package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// discoverAssetFiles finds all .txt files in the assets directory
func discoverAssetFiles() []string {
	var files []string
	seen := make(map[string]bool) // Deduplicate files

	// Get executable directory for better path resolution
	exePath, err := os.Executable()
	var binaryDir string
	if err == nil {
		binaryDir = filepath.Dir(exePath)
	}

	// Try multiple possible asset paths (prioritize user-writable locations)
	assetPaths := []string{
		filepath.Join(os.Getenv("HOME"), "sysc-Go", "assets"), // User home (writable)
		"assets",    // Current directory
		"./assets",  // Explicit relative
		"../assets", // Parent directory
		filepath.Join("/usr/local/share/syscgo", "assets"), // Local install (matches installer)
		filepath.Join("/usr/share/syscgo", "assets"),       // System install (matches installer)
	}

	// Add binary-relative path if available
	if binaryDir != "" {
		assetPaths = append(assetPaths, filepath.Join(binaryDir, "assets"))
	}

	for _, assetPath := range assetPaths {
		entries, err := os.ReadDir(assetPath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if strings.HasSuffix(strings.ToLower(name), ".txt") && !seen[name] {
				files = append(files, name)
				seen[name] = true
			}
		}
	}

	return files
}

// getAssetPath returns the full path to an asset file
func getAssetPath(filename string) string {
	// Get executable directory
	exePath, err := os.Executable()
	var binaryDir string
	if err == nil {
		binaryDir = filepath.Dir(exePath)
	}

	assetPaths := []string{
		filepath.Join(os.Getenv("HOME"), "sysc-Go", "assets", filename), // User home (writable, TUI saves here)
		filepath.Join("assets", filename),                               // ./assets/ (current dir)
		filepath.Join("../assets", filename),                            // ../assets/ (parent dir)
		filename,                                                        // Bare filename in current directory
	}

	// Add binary-relative path if available
	if binaryDir != "" {
		assetPaths = append(assetPaths, filepath.Join(binaryDir, "assets", filename))
	}

	// Add system paths last (read-only fallback)
	assetPaths = append(assetPaths,
		filepath.Join("/usr/local/share/syscgo", "assets", filename), // Local install (matches installer)
		filepath.Join("/usr/share/syscgo", "assets", filename),       // System install (matches installer)
	)

	for _, path := range assetPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return filename // fallback
}

// validateFilename validates a filename for safety and correctness
func validateFilename(filename string) error {
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	// Check for path traversal attempts
	if strings.Contains(filename, "..") {
		return fmt.Errorf("filename cannot contain '..'")
	}

	if strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return fmt.Errorf("filename cannot contain path separators")
	}

	// Check for invalid characters (allow alphanumeric, dash, underscore, space, dot)
	validName := regexp.MustCompile(`^[a-zA-Z0-9_\- .]+$`)
	if !validName.MatchString(filename) {
		return fmt.Errorf("filename contains invalid characters (only letters, numbers, spaces, dash, underscore, and dot allowed)")
	}

	// Check length
	if len(filename) > 255 {
		return fmt.Errorf("filename too long (max 255 characters)")
	}

	return nil
}

// saveToAssets saves content to a file in the assets directory
func saveToAssets(filename, content string) error {
	// Validate filename
	if err := validateFilename(filename); err != nil {
		return fmt.Errorf("invalid filename: %w", err)
	}

	// Validate content is not empty
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("content cannot be empty")
	}

	// Try to find writable assets directory
	assetPaths := []string{
		filepath.Join(os.Getenv("HOME"), "sysc-Go", "assets"), // User home (writable)
		"assets",    // Current directory
		"./assets",  // Explicit relative
		"../assets", // Parent directory
	}

	var targetPath string
	for _, assetPath := range assetPaths {
		// Check if directory exists
		if info, err := os.Stat(assetPath); err == nil && info.IsDir() {
			// Check if writable by trying to create a temp file
			testFile := filepath.Join(assetPath, ".write_test")
			if err := os.WriteFile(testFile, []byte("test"), 0644); err == nil {
				os.Remove(testFile)
				targetPath = assetPath
				break
			}
		}
	}

	// If no writable directory found, try to create ./assets
	if targetPath == "" {
		targetPath = "assets"
		if err := os.MkdirAll(targetPath, 0755); err != nil {
			return fmt.Errorf("could not create assets directory: %w", err)
		}
	}

	// Write file
	filePath := filepath.Join(targetPath, filename)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("could not write file: %w", err)
	}

	return nil
}
