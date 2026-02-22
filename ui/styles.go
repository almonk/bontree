package ui

import "github.com/charmbracelet/lipgloss"

// Color palette
const (
	colorBg        = "#1a1b26"
	colorBlue      = "#7aa2f7"
	colorFg        = "#c0caf5"
	colorFgDim     = "#a9b1d6"
	colorComment   = "#565f89"
	colorGutter    = "#3b4261"
	colorGreen     = "#9ece6a"
	colorYellow    = "#e0af68"
	colorOrange    = "#ff9e64"
	colorPurple    = "#bb9af7"
	colorRed       = "#f7768e"
	colorCyan      = "#7dcfff"
	colorSelection = "#283457"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorBlue)).
			PaddingLeft(1)

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(colorSelection)).
			Foreground(lipgloss.Color(colorFg)).
			Bold(true)

	dirStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorBlue)).
			Bold(true)

	fileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorFgDim))

	treeLineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorGutter))

	iconDirStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorBlue))

	iconFileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorFgDim))

	matchHighlightStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorOrange)).
				Bold(true)

	matchHighlightSelectedStyle = lipgloss.NewStyle().
					Background(lipgloss.Color(colorSelection)).
					Foreground(lipgloss.Color(colorOrange)).
					Bold(true)

	flatPathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorComment))

	flatPathSelectedStyle = lipgloss.NewStyle().
				Background(lipgloss.Color(colorSelection)).
				Foreground(lipgloss.Color(colorComment))

	// Status bar base style — all status styles inherit this background
	statusBase = lipgloss.NewStyle().
			Background(lipgloss.Color(colorBg))

	statusPathStyle     = statusBase.Foreground(lipgloss.Color(colorFgDim)).PaddingLeft(1).PaddingRight(1)
	statusBranchStyle   = statusBase.Foreground(lipgloss.Color(colorPurple)).Bold(true)
	statusFlashStyle = statusBase.Foreground(lipgloss.Color(colorGreen)).Bold(true).PaddingLeft(1)
	statusHelpStyle  = statusBase.Foreground(lipgloss.Color(colorGutter))

	searchInputStyle  = statusBase.Foreground(lipgloss.Color(colorFg)).PaddingLeft(1)
	searchPromptStyle = statusBase.Foreground(lipgloss.Color(colorBlue)).Bold(true).PaddingLeft(1)
)
