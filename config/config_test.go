package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Check some default bindings
	if cfg.ActionFor("q") != ActionQuit {
		t.Errorf("expected q=quit, got %q", cfg.ActionFor("q"))
	}
	if cfg.ActionFor("j") != ActionMoveDown {
		t.Errorf("expected j=move_down, got %q", cfg.ActionFor("j"))
	}
	if cfg.ActionFor("?") != ActionHelp {
		t.Errorf("expected ?=help, got %q", cfg.ActionFor("?"))
	}
	if cfg.ShowHidden {
		t.Error("expected show_hidden=false by default")
	}
}

func TestLoadMissing(t *testing.T) {
	cfg, err := LoadFrom("/nonexistent/path/config")
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	// Should return defaults
	if cfg.ActionFor("q") != ActionQuit {
		t.Error("expected defaults when file missing")
	}
}

func TestLoadKeybinds(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")

	content := `# My bontree config

keybind = q=quit
keybind = ctrl+c=quit
keybind = j=move_down
keybind = k=move_up
keybind = x=expand_all
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// When keybinds are present, defaults are cleared
	if cfg.ActionFor("G") != "" {
		t.Error("expected G to be unbound when user provides keybinds")
	}

	if cfg.ActionFor("q") != ActionQuit {
		t.Errorf("expected q=quit, got %q", cfg.ActionFor("q"))
	}
	if cfg.ActionFor("x") != ActionExpandAll {
		t.Errorf("expected x=expand_all, got %q", cfg.ActionFor("x"))
	}
}

func TestLoadUnbind(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")

	content := `keybind = q=quit
keybind = q=unbind
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ActionFor("q") != "" {
		t.Error("expected q to be unbound")
	}
}

func TestLoadShowHidden(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")

	content := `show-hidden = true
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !cfg.ShowHidden {
		t.Error("expected show-hidden=true")
	}

	// Defaults should still be present since no keybind lines
	if cfg.ActionFor("q") != ActionQuit {
		t.Error("expected defaults preserved when no keybind lines")
	}
}

func TestLoadInvalidAction(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")

	content := `keybind = q=does_not_exist
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadFrom(path)
	if err == nil {
		t.Fatal("expected error for unknown action")
	}
}

func TestLoadInvalidSyntax(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")

	content := `this has no equals sign
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadFrom(path)
	if err == nil {
		t.Fatal("expected error for invalid syntax")
	}
}

func TestLoadUnknownKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")

	content := `foobar = baz
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadFrom(path)
	if err == nil {
		t.Fatal("expected error for unknown config key")
	}
}

func TestKeysFor(t *testing.T) {
	cfg := DefaultConfig()
	keys := cfg.KeysFor(ActionQuit)

	found := map[string]bool{}
	for _, k := range keys {
		found[k] = true
	}

	if !found["q"] || !found["ctrl+c"] {
		t.Errorf("expected q and ctrl+c bound to quit, got %v", keys)
	}
}

func TestCommentsAndBlankLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")

	content := `# This is a comment
# Another comment

   # Indented comment

keybind = q=quit

# Trailing comment
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ActionFor("q") != ActionQuit {
		t.Error("expected q=quit")
	}
}
