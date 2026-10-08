package tui

import (
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
//   - overwrite: Whether an existing file at the destination may be replaced.
//     With false an existing file yields an error wrapping ErrFileExists.
//
// Returns:
//   - nil on success
//   - error if filename is invalid, file cannot be written, or config update fails
//
// Example:
//
//	art := "HELLO\nWORLD"
//	err := ExportToSyscWalls("greeting.txt", art, false)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Security:
//   - Filenames are sanitized to prevent directory traversal
//   - Only safe characters allowed in filenames
//   - Files created with 0600 permissions (user-only read/write)
//   - Directories created with 0700 permissions (user-only access)
func ExportToSyscWalls(filename, content string, overwrite bool) error {
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
	if err := refuseExisting(artPath, overwrite); err != nil {
		return err
	}
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

// syscWallsDefaultConfig is written when no config exists yet.  It mirrors
// sysc-walls' own default layout so the daemon still shows its help comments.
// The four [animation] values an export owns are filled in by the caller.
const syscWallsDefaultConfig = `# sysc-walls daemon configuration

[idle]
timeout = 5m
min_duration = 30s

[daemon]
debug = false

[animation]
# Available effects: fire, matrix, rain, beam-text, ring-text, pour, print,
# matrix-art, rain-art, blackhole-text
# Available themes: dracula, gruvbox, nord, tokyo-night, catppuccin, material,
# solarized, monochrome, transishardjob, rama, eldritch, dark
effect = %EFFECT%
file = %FILE%
theme = %THEME%
datetime = false
cycle = false
cycle_interval = 5m

[datetime]
position = bottom
interval = 1s

[terminal]
kitty = true
fullscreen = true
`

// syscWallsOwnedKeys are the [animation] keys an export owns.  sysc-walls
// reads `effect`, not `type`: an unknown key is silently ignored, which is why
// exporting used to leave the screensaver running whatever effect the user
// already had, showing no exported art at all.
//
// An existing config only ever gets effect and file.  An export picks what
// plays; rewriting the user's theme or cycle would restate a setting they
// chose (#114).  A config created from scratch has nothing to preserve, so it
// gets the full default set.
func syscWallsOwnedKeys(artPath string, fresh bool) [][2]string {
	if !fresh {
		return [][2]string{
			{"effect", "beam-text"},
			{"file", artPath},
		}
	}
	return [][2]string{
		{"effect", "beam-text"},
		{"file", artPath},
		{"theme", "dracula"},
		{"cycle", "false"},
	}
}

// updateSyscWallsConfig points the [animation] section at artPath, creating a
// default config when there is none.  An existing config is edited in place:
// comments, blank lines, key order and every other section survive untouched.
func updateSyscWallsConfig(configPath, artPath string) error {
	// Create config directory with user-only permissions
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// An existing file keeps its own mode; a new one is user-only.
	mode := os.FileMode(0600)

	data, readErr := os.ReadFile(configPath)
	if readErr != nil {
		if !os.IsNotExist(readErr) {
			return fmt.Errorf("failed to read config: %w", readErr)
		}
		content := syscWallsDefaultConfig
		for _, kv := range syscWallsOwnedKeys(artPath, true) {
			content = strings.Replace(content, "%"+strings.ToUpper(kv[0])+"%", kv[1], 1)
		}
		return writeConfigAtomic(configPath, []byte(content), mode)
	}

	if info, err := os.Stat(configPath); err == nil {
		mode = info.Mode().Perm()
	}

	updated := setAnimationKeys(string(data), syscWallsOwnedKeys(artPath, false))
	return writeConfigAtomic(configPath, []byte(updated), mode)
}

// writeConfigAtomic replaces path via a temp file in the same directory, so a
// reader never sees a half-written config.  A symlinked config is resolved
// first: renaming onto the link itself would replace the link.
func writeConfigAtomic(path string, content []byte, mode os.FileMode) error {
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		path = resolved
	}

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".daemon.conf.tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	committed := false
	defer func() {
		tmp.Close()
		if !committed {
			os.Remove(tmpName)
		}
	}()

	if _, err := tmp.Write(content); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	// Chmod before the rename so the file is never visible at the wrong mode.
	if err := tmp.Chmod(mode); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	committed = true
	return nil
}

// setAnimationKeys rewrites the owned keys inside [animation] of an INI config
// and returns the result.  Only the owned keys change; every other byte of the
// file, including comments and blank lines, is preserved.
func setAnimationKeys(content string, owned [][2]string) string {
	lines := splitLinesKeepEndings(content)

	// Which owned keys still need a value written.
	missing := make([][2]string, len(owned))
	copy(missing, owned)
	seen := map[string]bool{}
	has := map[string]bool{}

	inAnimation := false
	sawAnimation := false
	lastContent := -1 // index of the last non-blank line of the section

	ending := "\n"
	if strings.Contains(content, "\r\n") {
		ending = "\r\n"
	}

	for i, line := range lines {
		body := strings.TrimRight(line, "\r\n")

		trimmed := strings.TrimSpace(body)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			if inAnimation {
				// Section ended: append anything still missing after its last
				// real line, so trailing blank lines stay trailing.
				lines, _ = insertAt(lines, insertIndex(lines, i, lastContent), missing, ending)
				missing = nil
			}
			inAnimation = strings.EqualFold(strings.TrimSpace(trimmed[1:len(trimmed)-1]), "animation")
			if inAnimation {
				sawAnimation = true
				lastContent = -1
			}
			continue
		}

		if !inAnimation {
			continue
		}

		if !strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, ";") {
			lastContent = i
		}

		key, prefix, ok := splitConfigAssignment(body)
		if !ok {
			continue
		}

		// `type` is what older sysc-Go versions wrote; sysc-walls never read it.
		if key == "type" {
			lines[i] = ""
			continue
		}

		if !isOwnedKey(owned, key) {
			continue
		}
		if seen[key] {
			lines[i] = "" // Collapse a repeated owned key.
			continue
		}

		seen[key] = true
		has[key] = true
		missing = dropKey(missing, key)
		lines[i] = prefix + valueOf(owned, key) + lineEnding(line)
	}

	if sawAnimation {
		lines, _ = insertAt(lines, insertIndex(lines, len(lines), lastContent), missing, ending)
	} else {
		if len(lines) > 0 && lines[len(lines)-1] != "\n" && lines[len(lines)-1] != "\r\n" {
			lines = append(lines, ending)
		}
		lines = append(lines, ending, "[animation]"+ending)
		for _, kv := range owned {
			lines = append(lines, kv[0]+" = "+kv[1]+ending)
		}
	}

	return strings.Join(lines, "")
}

// splitConfigAssignment splits "  key = value" into its trimmed key and the
// prefix to keep verbatim: indentation, the key, and the spacing up to and
// including "=".  That keeps "  effect = x" and "effect=x" spelled the way the
// user spelled them.
func splitConfigAssignment(body string) (key, prefix string, ok bool) {
	eq := strings.Index(body, "=")
	if eq < 0 {
		return "", "", false
	}
	prefix = body[:eq+1]
	rest := body[eq+1:]
	// Carry the spacing that followed "=" onto the new value.
	i := 0
	for i < len(rest) && (rest[i] == ' ' || rest[i] == '\t') {
		i++
	}
	prefix += rest[:i]
	key = strings.TrimSpace(body[:eq])
	if key == "" {
		return "", "", false
	}
	return key, prefix, true
}

func lineEnding(line string) string {
	if strings.HasSuffix(line, "\r\n") {
		return "\r\n"
	}
	if strings.HasSuffix(line, "\n") {
		return "\n"
	}
	return ""
}

func splitLinesKeepEndings(content string) []string {
	var lines []string
	for len(content) > 0 {
		i := strings.IndexByte(content, '\n')
		if i < 0 {
			lines = append(lines, content)
			break
		}
		lines = append(lines, content[:i+1])
		content = content[i+1:]
	}
	return lines
}

// insertIndex is where a section's missing keys belong: straight after its last
// non-blank line, or at the next header (EOF) when the section has none.
func insertIndex(lines []string, at, lastContent int) int {
	if lastContent >= 0 && lastContent+1 <= len(lines) {
		return lastContent + 1
	}
	return at
}

func insertAt(lines []string, at int, keys [][2]string, ending string) ([]string, int) {
	if len(keys) == 0 {
		return lines, at
	}
	added := make([]string, 0, len(keys))
	for _, kv := range keys {
		added = append(added, kv[0]+" = "+kv[1]+ending)
	}
	out := make([]string, 0, len(lines)+len(added))
	out = append(out, lines[:at]...)
	out = append(out, added...)
	out = append(out, lines[at:]...)
	return out, at + len(added)
}

func isOwnedKey(owned [][2]string, key string) bool {
	for _, kv := range owned {
		if strings.EqualFold(kv[0], key) {
			return true
		}
	}
	return false
}

func valueOf(owned [][2]string, key string) string {
	for _, kv := range owned {
		if strings.EqualFold(kv[0], key) {
			return kv[1]
		}
	}
	return ""
}

func dropKey(keys [][2]string, key string) [][2]string {
	out := keys[:0]
	for _, kv := range keys {
		if !strings.EqualFold(kv[0], key) {
			out = append(out, kv)
		}
	}
	return out
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
//   - overwrite: Whether an existing destination file may be replaced
//
// Returns:
//   - nil on success
//   - error if the export fails or target is unknown
//
// Example:
//
//	art := []string{"\x1b[31mHello\x1b[0m", "\x1b[32mWorld\x1b[0m"}
//	err := ExportBitArt("greeting", art, 1, false)
//	if err != nil {
//	    log.Fatal(err)
//	}
func ExportBitArt(filename string, content []string, target int, overwrite bool) error {
	// Strip ANSI codes and any trailing alignment padding
	plainContent := ""
	for _, line := range content {
		plainContent += strings.TrimRight(stripANSI(line), " ") + "\n"
	}

	// Add .txt extension if not present
	if !strings.HasSuffix(filename, ".txt") {
		filename += ".txt"
	}

	switch target {
	case 0: // syscgo
		return saveToAssets(filename, plainContent, overwrite)

	case 1: // sysc-walls
		return ExportToSyscWalls(filename, plainContent, overwrite)

	default:
		return fmt.Errorf("unknown export target: %d", target)
	}
}
