package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// userAssetsDir is where the TUI saves art: $XDG_DATA_HOME/syscgo/assets,
// defaulting to ~/.local/share/syscgo/assets.  A relative or empty
// XDG_DATA_HOME is ignored per the XDG spec, so art never lands in the cwd.
// ponytail: the XDG rule is duplicated in cmd/syscgo/main.go; share it via a
// package if a third copy shows up.
func userAssetsDir() string {
	base := os.Getenv("XDG_DATA_HOME")
	if base == "" || !filepath.IsAbs(base) {
		base = filepath.Join(os.Getenv("HOME"), ".local", "share")
	}
	return filepath.Join(base, "syscgo", "assets")
}

// legacyAssetsDir is the assets folder of a source checkout, still honoured for
// users who already save there.
func legacyAssetsDir() string {
	return filepath.Join(os.Getenv("HOME"), "sysc-Go", "assets")
}

// isDir reports whether path is an existing directory.
func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

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
		legacyAssetsDir(), // Legacy source checkout
		userAssetsDir(),   // Where the TUI saves new art
		"assets",          // Current directory
		"./assets",        // Explicit relative
		"../assets",       // Parent directory
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
		filepath.Join(legacyAssetsDir(), filename), // Legacy source checkout
		filepath.Join(userAssetsDir(), filename),   // Where the TUI saves new art
		filepath.Join("assets", filename),          // ./assets/ (current dir)
		filepath.Join("../assets", filename),       // ../assets/ (parent dir)
		filename,                                   // Bare filename in current directory
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

// saveToAssets saves content to a file in the assets directory.  An existing
// file is left alone unless overwrite is set.
func saveToAssets(filename, content string, overwrite bool) error {
	// Validate filename
	if err := validateFilename(filename); err != nil {
		return fmt.Errorf("invalid filename: %w", err)
	}

	// Validate content is not empty
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("content cannot be empty")
	}

	// Save to a stable per-user directory so art is found from any working
	// directory.  A legacy ~/sysc-Go/assets keeps priority for users who
	// already have one; the cwd is never written to, which would otherwise
	// drop the file into whatever unrelated project the TUI was started in.
	targetPath := userAssetsDir()
	if legacy := legacyAssetsDir(); isDir(legacy) {
		targetPath = legacy
	}
	if err := os.MkdirAll(targetPath, 0o755); err != nil {
		return fmt.Errorf("could not create assets directory: %w", err)
	}

	// Write file
	filePath := filepath.Join(targetPath, filename)
	if err := refuseExisting(filePath, overwrite); err != nil {
		return err
	}
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("could not write file: %w", err)
	}

	return nil
}

// exportDirLabel names where an export target actually writes, for the save
// prompt.  0 lands in the assets folder, 1 under the sysc-walls data dir.
func exportDirLabel(target int) string {
	if target == 1 {
		return filepath.Join("~", ".local", "share", "syscgo", "walls")
	}
	return "the assets/ folder"
}

// refuseExisting returns an *ExistsError when path is already there and the
// caller did not opt into overwriting it.
func refuseExisting(path string, overwrite bool) error {
	if overwrite {
		return nil
	}
	if _, err := os.Stat(path); err == nil {
		return &ExistsError{Path: path}
	}
	return nil
}
