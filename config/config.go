package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Action represents a named action that can be bound to a key.
type Action string

const (
	ActionQuit         Action = "quit"
	ActionMoveDown     Action = "move_down"
	ActionMoveUp       Action = "move_up"
	ActionGoTop        Action = "go_top"
	ActionGoBottom     Action = "go_bottom"
	ActionHalfPageDown Action = "half_page_down"
	ActionHalfPageUp   Action = "half_page_up"
	ActionExpand       Action = "expand"
	ActionCollapse     Action = "collapse"
	ActionToggle       Action = "toggle"
	ActionCopyPath     Action = "copy_path"
	ActionExpandAll    Action = "expand_all"
	ActionCollapseAll  Action = "collapse_all"
	ActionToggleHidden Action = "toggle_hidden"
	ActionSearch       Action = "search"
	ActionFlatSearch   Action = "flat_search"
	ActionHelp         Action = "help"
	ActionClearFilter  Action = "clear_filter"
	ActionOpenEditor   Action = "open_editor"

	// Search mode actions
	ActionSearchConfirm   Action = "search_confirm"
	ActionSearchCancel    Action = "search_cancel"
	ActionSearchBackspace Action = "search_backspace"
	ActionSearchNextMatch Action = "search_next_match"
	ActionSearchPrevMatch Action = "search_prev_match"
)

// Config holds all parsed configuration.
type Config struct {
	// Keybinds maps a key string (e.g. "ctrl+c", "j", "G") to an action.
	Keybinds map[string]Action

	// ShowHidden controls whether hidden files are shown by default.
	ShowHidden bool

	// Theme is the name of a Ghostty-compatible theme to use.
	// Empty string means inherit from the terminal.
	Theme string
}

// DefaultConfig returns the config with all default keybindings.
func DefaultConfig() *Config {
	c := &Config{
		Keybinds:   make(map[string]Action),
		ShowHidden: false,
	}

	// Normal mode defaults
	defaults := map[string]Action{
		"q":      ActionQuit,
		"ctrl+c": ActionQuit,
		"j":      ActionMoveDown,
		"down":   ActionMoveDown,
		"k":      ActionMoveUp,
		"up":     ActionMoveUp,
		"g":      ActionGoTop,
		"G":      ActionGoBottom,
		"ctrl+d": ActionHalfPageDown,
		"ctrl+u": ActionHalfPageUp,
		"l":      ActionExpand,
		"right":  ActionExpand,
		"h":      ActionCollapse,
		"left":   ActionCollapse,
		"enter":  ActionToggle,
		" ":      ActionToggle,
		"c":      ActionCopyPath,
		"E":      ActionExpandAll,
		"W":      ActionCollapseAll,
		".":      ActionToggleHidden,
		"/":      ActionSearch,
		"ctrl+f": ActionFlatSearch,
		"ctrl+_": ActionFlatSearch,
		"?":      ActionHelp,
		"esc":    ActionClearFilter,
	}

	for k, v := range defaults {
		c.Keybinds[k] = v
	}

	return c
}

// ConfigPath returns the default config file path.
func ConfigPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "bontree", "config")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "bontree", "config")
}

// Load reads the config file from the default path. If the file doesn't
// exist, it returns the default config with no error.
func Load() (*Config, error) {
	return LoadFrom(ConfigPath())
}

// LoadFrom reads a config file from the given path. If the file doesn't
// exist, it returns the default config with no error.
func LoadFrom(path string) (*Config, error) {
	cfg := DefaultConfig()

	if path == "" {
		return cfg, nil
	}

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("opening config: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip blank lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse "key = value"
		eqIdx := strings.Index(line, "=")
		if eqIdx < 0 {
			return nil, fmt.Errorf("%s:%d: invalid syntax (expected key = value): %s", path, lineNum, line)
		}

		key := strings.TrimSpace(line[:eqIdx])
		value := strings.TrimSpace(line[eqIdx+1:])

		if key == "" {
			return nil, fmt.Errorf("%s:%d: empty key", path, lineNum)
		}

		switch key {
		case "keybind":
			if err := parseKeybind(cfg, value, path, lineNum); err != nil {
				return nil, err
			}

		case "show-hidden":
			switch value {
			case "true":
				cfg.ShowHidden = true
			case "false":
				cfg.ShowHidden = false
			default:
				return nil, fmt.Errorf("%s:%d: show-hidden must be true or false, got %q", path, lineNum, value)
			}

		case "theme":
			cfg.Theme = value

		default:
			return nil, fmt.Errorf("%s:%d: unknown config key %q", path, lineNum, key)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	return cfg, nil
}

// parseKeybind parses a keybind value like "ctrl+c=quit" or "unbind=j".
func parseKeybind(cfg *Config, value string, path string, lineNum int) error {
	// Find the last '=' to split key from action, since the key itself
	// could be '=' or contain '=' in theory.
	eqIdx := strings.LastIndex(value, "=")
	if eqIdx < 0 {
		return fmt.Errorf("%s:%d: invalid keybind syntax (expected key=action): %s", path, lineNum, value)
	}

	bindKey := strings.TrimSpace(value[:eqIdx])
	actionStr := strings.TrimSpace(value[eqIdx+1:])

	// Support "space" as a named key for the space character
	if bindKey == "space" {
		bindKey = " "
	}

	if bindKey == "" {
		return fmt.Errorf("%s:%d: empty keybind key", path, lineNum)
	}

	// "unbind" removes a binding
	if actionStr == "unbind" {
		delete(cfg.Keybinds, bindKey)
		return nil
	}

	action := Action(actionStr)
	if !isValidAction(action) {
		return fmt.Errorf("%s:%d: unknown action %q", path, lineNum, actionStr)
	}

	cfg.Keybinds[bindKey] = action
	return nil
}

func isValidAction(a Action) bool {
	switch a {
	case ActionQuit, ActionMoveDown, ActionMoveUp, ActionGoTop, ActionGoBottom,
		ActionHalfPageDown, ActionHalfPageUp, ActionExpand, ActionCollapse,
		ActionToggle, ActionCopyPath, ActionExpandAll, ActionCollapseAll,
		ActionToggleHidden, ActionSearch, ActionFlatSearch, ActionHelp,
		ActionClearFilter, ActionOpenEditor, ActionSearchConfirm, ActionSearchCancel,
		ActionSearchBackspace, ActionSearchNextMatch, ActionSearchPrevMatch:
		return true
	}
	return false
}

// ActionFor returns the action bound to the given key string, or "" if unbound.
func (c *Config) ActionFor(key string) Action {
	return c.Keybinds[key]
}

// KeysFor returns all keys bound to the given action.
func (c *Config) KeysFor(action Action) []string {
	var keys []string
	for k, a := range c.Keybinds {
		if a == action {
			keys = append(keys, k)
		}
	}
	return keys
}
