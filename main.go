package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/almonk/bontree/config"
	"github.com/almonk/bontree/theme"
	"github.com/almonk/bontree/ui"
	tea "github.com/charmbracelet/bubbletea"
)

// Version is set at build time via -ldflags
var Version = "dev"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Printf("bontree %s\n", Version)
		os.Exit(0)
	}

	rootPath, focusPath, err := resolveLaunchPaths(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %s\n", err)
		os.Exit(1)
	}

	// Load theme if configured
	if cfg.Theme != "" {
		t, err := theme.Load(cfg.Theme)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Theme error: %s\n", err)
			os.Exit(1)
		}
		ui.ApplyTheme(t)
	}

	model, err := ui.NewWithFocus(rootPath, focusPath, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building tree: %s\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseAllMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func resolveLaunchPaths(args []string) (string, string, error) {
	switch len(args) {
	case 0:
		return ".", "", nil
	case 1:
		info, err := os.Stat(args[0])
		if err != nil {
			return "", "", err
		}
		if info.IsDir() {
			return args[0], "", nil
		}
		return filepath.Dir(args[0]), args[0], nil
	case 2:
		rootPath := args[0]
		info, err := os.Stat(rootPath)
		if err != nil {
			return "", "", err
		}
		if !info.IsDir() {
			return "", "", fmt.Errorf("%s is not a directory", rootPath)
		}

		focusPath := args[1]
		if !filepath.IsAbs(focusPath) {
			focusPath = filepath.Join(rootPath, focusPath)
		}
		if _, err := os.Stat(focusPath); err != nil {
			return "", "", err
		}
		return rootPath, focusPath, nil
	default:
		return "", "", fmt.Errorf("usage: bontree [path] [focus-path]")
	}
}
