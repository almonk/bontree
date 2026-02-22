// Package theme provides Ghostty-compatible theme loading for bontree.
//
// Themes are plain text files using the same format as Ghostty:
//
//	palette = 0=#1a1b26
//	palette = 1=#f7768e
//	...
//	background = #1a1b26
//	foreground = #c0caf5
//	selection-background = #33467c
//	selection-foreground = #c0caf5
//
// Theme search order:
//  1. ~/.config/bontree/themes/<name>
//  2. Ghostty app bundle themes (macOS)
//  3. ~/.config/ghostty/themes/<name>
package theme

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// Theme holds the parsed color palette from a Ghostty-format theme file.
type Theme struct {
	Name string

	// 16-color ANSI palette (indices 0–15). Empty string means "not set".
	Palette [16]string

	Background          string
	Foreground          string
	SelectionBackground string
	SelectionForeground string
	CursorColor         string
	CursorText          string
}

// themeSearchDirs returns directories to search for themes, in priority order.
func themeSearchDirs() []string {
	var dirs []string

	// 1. Bontree's own theme dir
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		if home, err := os.UserHomeDir(); err == nil {
			configHome = filepath.Join(home, ".config")
		}
	}
	if configHome != "" {
		dirs = append(dirs, filepath.Join(configHome, "bontree", "themes"))
	}

	// 2. Ghostty app bundle (macOS)
	if runtime.GOOS == "darwin" {
		dirs = append(dirs, "/Applications/Ghostty.app/Contents/Resources/ghostty/themes")
	}

	// 3. Ghostty user themes
	if configHome != "" {
		dirs = append(dirs, filepath.Join(configHome, "ghostty", "themes"))
	}

	return dirs
}

// Load finds and parses a theme by name. Returns nil, nil if not found.
func Load(name string) (*Theme, error) {
	if name == "" {
		return nil, nil
	}

	// If it's an absolute path, load directly
	if filepath.IsAbs(name) {
		return parseFile(name)
	}

	// Search theme directories
	for _, dir := range themeSearchDirs() {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			return parseFile(path)
		}
	}

	return nil, fmt.Errorf("theme %q not found", name)
}

// parseFile reads a Ghostty-format theme file.
func parseFile(path string) (*Theme, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening theme: %w", err)
	}
	defer f.Close()

	t := &Theme{
		Name: filepath.Base(path),
	}

	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		eqIdx := strings.Index(line, "=")
		if eqIdx < 0 {
			continue // skip lines we don't understand
		}

		key := strings.TrimSpace(line[:eqIdx])
		value := strings.TrimSpace(line[eqIdx+1:])

		switch key {
		case "palette":
			// Format: "N=#rrggbb"
			parts := strings.SplitN(value, "=", 2)
			if len(parts) != 2 {
				continue
			}
			idx, err := strconv.Atoi(strings.TrimSpace(parts[0]))
			if err != nil || idx < 0 || idx > 15 {
				continue
			}
			t.Palette[idx] = strings.TrimSpace(parts[1])

		case "background":
			t.Background = value
		case "foreground":
			t.Foreground = value
		case "selection-background":
			t.SelectionBackground = value
		case "selection-foreground":
			t.SelectionForeground = value
		case "cursor-color":
			t.CursorColor = value
		case "cursor-text":
			t.CursorText = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading theme: %w", err)
	}

	return t, nil
}

// List returns the names of all available themes across all search directories.
func List() []string {
	seen := make(map[string]bool)
	var names []string

	for _, dir := range themeSearchDirs() {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
				continue
			}
			if !seen[e.Name()] {
				seen[e.Name()] = true
				names = append(names, e.Name())
			}
		}
	}

	return names
}
