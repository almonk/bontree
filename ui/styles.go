package ui

import "github.com/charmbracelet/lipgloss"

// Color palette using ANSI colors that adapt to the terminal's color scheme.
// Semantic colors (blue, green, red, etc.) use ANSI indices 0-15,
// which are customized by the user's terminal theme.
// Colors that need light/dark awareness use AdaptiveColor.
var (
	// Semantic ANSI colors — themed by the terminal
	colorBlue   lipgloss.TerminalColor = lipgloss.Color("12")
	colorGreen  lipgloss.TerminalColor = lipgloss.Color("10")
	colorRed    lipgloss.TerminalColor = lipgloss.Color("9")
	colorYellow lipgloss.TerminalColor = lipgloss.Color("11")
	colorPurple lipgloss.TerminalColor = lipgloss.Color("13")
	colorCyan   lipgloss.TerminalColor = lipgloss.Color("14")
	colorOrange lipgloss.TerminalColor = lipgloss.Color("208") // 256-color; no ANSI equivalent

	// Adaptive text colors
	colorFg      lipgloss.TerminalColor = lipgloss.AdaptiveColor{Light: "0", Dark: "15"}
	colorFgDim   lipgloss.TerminalColor = lipgloss.AdaptiveColor{Light: "8", Dark: "7"}
	colorComment lipgloss.TerminalColor = lipgloss.Color("8")
	colorGutter  lipgloss.TerminalColor = lipgloss.AdaptiveColor{Light: "248", Dark: "239"}

	// Adaptive background colors
	colorBg        lipgloss.TerminalColor = lipgloss.AdaptiveColor{Light: "254", Dark: "235"}
	colorSelection lipgloss.TerminalColor = lipgloss.AdaptiveColor{Light: "253", Dark: "237"}
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorBlue).
			PaddingLeft(1)

	selectedStyle = lipgloss.NewStyle().
			Background(colorSelection).
			Foreground(colorFg).
			Bold(true)

	dirStyle = lipgloss.NewStyle().
			Foreground(colorBlue).
			Bold(true)

	fileStyle = lipgloss.NewStyle().
			Foreground(colorFgDim)

	treeLineStyle = lipgloss.NewStyle().
			Foreground(colorGutter)

	iconDirStyle = lipgloss.NewStyle().
			Foreground(colorBlue)

	iconFileStyle = lipgloss.NewStyle().
			Foreground(colorFgDim)

	matchHighlightStyle = lipgloss.NewStyle().
				Foreground(colorOrange).
				Bold(true)

	matchHighlightSelectedStyle = lipgloss.NewStyle().
					Background(colorSelection).
					Foreground(colorOrange).
					Bold(true)

	flatPathStyle = lipgloss.NewStyle().
			Foreground(colorComment)

	flatPathSelectedStyle = lipgloss.NewStyle().
				Background(colorSelection).
				Foreground(colorComment)

	// Status bar base style — all status styles inherit this background
	statusBase = lipgloss.NewStyle().
			Background(colorBg)

	statusPathStyle   = statusBase.Foreground(colorFgDim).PaddingLeft(1).PaddingRight(1)
	statusBranchStyle = statusBase.Foreground(colorPurple).Bold(true)
	statusFlashStyle  = statusBase.Foreground(colorGreen).Bold(true).PaddingLeft(1)
	statusHelpStyle   = statusBase.Foreground(colorGutter)

	searchInputStyle  = statusBase.Foreground(colorFg).PaddingLeft(1)
	searchPromptStyle = statusBase.Foreground(colorBlue).Bold(true).PaddingLeft(1)
)
