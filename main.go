package main

import (
	"fmt"
	"os"

	"github.com/alasdairmonk/bontree/config"
	"github.com/alasdairmonk/bontree/ui"
	tea "github.com/charmbracelet/bubbletea"
)

// Version is set at build time via -ldflags
var Version = "dev"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Printf("bontree %s\n", Version)
		os.Exit(0)
	}

	path := "."
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	// Verify path exists
	info, err := os.Stat(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: %s is not a directory\n", path)
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %s\n", err)
		os.Exit(1)
	}

	model, err := ui.New(path, cfg)
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
