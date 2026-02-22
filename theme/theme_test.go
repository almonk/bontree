package theme

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadGhosttyTheme(t *testing.T) {
	th, err := Load("TokyoNight")
	if err != nil {
		t.Skipf("TokyoNight theme not available: %v", err)
	}
	if th == nil {
		t.Skip("TokyoNight theme not found")
	}

	if th.Background == "" {
		t.Error("expected background color to be set")
	}
	if th.Foreground == "" {
		t.Error("expected foreground color to be set")
	}
	if th.Palette[12] == "" {
		t.Error("expected palette[12] (bright blue) to be set")
	}

	t.Logf("Theme: %s, BG: %s, FG: %s", th.Name, th.Background, th.Foreground)
}

func TestParseFile(t *testing.T) {
	// Create a temp theme file
	dir := t.TempDir()
	path := filepath.Join(dir, "test-theme")
	content := `# Test theme
palette = 0=#000000
palette = 1=#ff0000
palette = 12=#0000ff
background = #1a1b26
foreground = #c0caf5
selection-background = #33467c
selection-foreground = #c0caf5
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	th, err := parseFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if th.Palette[0] != "#000000" {
		t.Errorf("palette[0] = %q, want #000000", th.Palette[0])
	}
	if th.Palette[1] != "#ff0000" {
		t.Errorf("palette[1] = %q, want #ff0000", th.Palette[1])
	}
	if th.Palette[12] != "#0000ff" {
		t.Errorf("palette[12] = %q, want #0000ff", th.Palette[12])
	}
	if th.Background != "#1a1b26" {
		t.Errorf("background = %q, want #1a1b26", th.Background)
	}
	if th.Foreground != "#c0caf5" {
		t.Errorf("foreground = %q, want #c0caf5", th.Foreground)
	}
	if th.SelectionBackground != "#33467c" {
		t.Errorf("selection-background = %q, want #33467c", th.SelectionBackground)
	}
}

func TestList(t *testing.T) {
	names := List()
	t.Logf("Found %d themes", len(names))
	// On a system with Ghostty installed, we should find themes
	if len(names) == 0 {
		t.Skip("No themes found (Ghostty may not be installed)")
	}
}

func TestLoadNotFound(t *testing.T) {
	_, err := Load("this-theme-definitely-does-not-exist-xyz")
	if err == nil {
		t.Error("expected error for non-existent theme")
	}
}

func TestLoadEmpty(t *testing.T) {
	th, err := Load("")
	if err != nil {
		t.Errorf("unexpected error for empty theme: %v", err)
	}
	if th != nil {
		t.Error("expected nil theme for empty name")
	}
}
