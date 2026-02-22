package ui

import (
	"github.com/alasdairmonk/bontree/theme"
	"github.com/charmbracelet/lipgloss"
)

// colorScheme holds all the resolved colors used by the UI.
// When no theme is loaded, these use ANSI indices that adapt to the terminal.
// When a theme is loaded, these use the theme's hex colors directly.
type colorScheme struct {
	blue   lipgloss.TerminalColor
	green  lipgloss.TerminalColor
	red    lipgloss.TerminalColor
	yellow lipgloss.TerminalColor
	purple lipgloss.TerminalColor
	cyan   lipgloss.TerminalColor
	orange lipgloss.TerminalColor

	fg      lipgloss.TerminalColor
	fgDim   lipgloss.TerminalColor
	comment lipgloss.TerminalColor
	gutter  lipgloss.TerminalColor

	bg        lipgloss.TerminalColor
	selection lipgloss.TerminalColor
}

// defaultColors returns ANSI-based colors that inherit from the terminal.
func defaultColors() colorScheme {
	return colorScheme{
		blue:   lipgloss.Color("12"),
		green:  lipgloss.Color("10"),
		red:    lipgloss.Color("9"),
		yellow: lipgloss.Color("11"),
		purple: lipgloss.Color("13"),
		cyan:   lipgloss.Color("14"),
		orange: lipgloss.Color("208"),

		fg:      lipgloss.AdaptiveColor{Light: "0", Dark: "7"},
		fgDim:   lipgloss.AdaptiveColor{Light: "8", Dark: "7"},
		comment: lipgloss.Color("8"),
		gutter:  lipgloss.AdaptiveColor{Light: "248", Dark: "239"},

		bg:        lipgloss.AdaptiveColor{Light: "254", Dark: "235"},
		selection: lipgloss.AdaptiveColor{Light: "253", Dark: "237"},
	}
}

// themedColors creates a color scheme from a parsed Ghostty theme.
func themedColors(t *theme.Theme) colorScheme {
	c := defaultColors()

	// Map ANSI palette indices to semantic colors.
	// Standard mapping: 0=black 1=red 2=green 3=yellow 4=blue 5=purple 6=cyan 7=white
	// Bright variants: 8-15
	set := func(target *lipgloss.TerminalColor, idx int) {
		if t.Palette[idx] != "" {
			*target = lipgloss.Color(t.Palette[idx])
		}
	}

	set(&c.red, 9)
	set(&c.green, 10)
	set(&c.yellow, 11)
	set(&c.blue, 12)
	set(&c.purple, 13)
	set(&c.cyan, 14)

	// Use palette 8 (bright black) for comments/dim text
	if t.Palette[8] != "" {
		c.comment = lipgloss.Color(t.Palette[8])
		c.gutter = lipgloss.Color(t.Palette[8])
	}

	// Foreground
	if t.Foreground != "" {
		c.fg = lipgloss.Color(t.Foreground)
		c.fgDim = lipgloss.Color(t.Foreground)
	}
	// Use palette 7 (white) for dim foreground if available
	if t.Palette[7] != "" {
		c.fgDim = lipgloss.Color(t.Palette[7])
	}

	// Background — use palette 0 (black) lightened as the status bar bg,
	// and derive a subtle selection highlight from it.
	if t.Background != "" {
		c.bg = lipgloss.Color(t.Background)
	}
	// Use palette 8 (bright black / comment) as the selection highlight.
	// This is always a muted mid-tone that works as a subtle bar highlight,
	// unlike selection-background which is designed for text selection and
	// is often too bright/inverted.
	if t.Palette[8] != "" {
		c.selection = lipgloss.Color(t.Palette[8])
	}

	return c
}

// --- Current active colors and styles (package-level) ---

var colors colorScheme

// Styles — these reference `colors` and are rebuilt by initStyles().
var (
	titleStyle                  lipgloss.Style
	selectedStyle               lipgloss.Style
	dirStyle                    lipgloss.Style
	fileStyle                   lipgloss.Style
	treeLineStyle               lipgloss.Style
	iconDirStyle                lipgloss.Style
	iconFileStyle               lipgloss.Style
	matchHighlightStyle         lipgloss.Style
	matchHighlightSelectedStyle lipgloss.Style
	flatPathStyle               lipgloss.Style
	flatPathSelectedStyle       lipgloss.Style
	statusBase                  lipgloss.Style
	statusPathStyle             lipgloss.Style
	statusBranchStyle           lipgloss.Style
	statusFlashStyle            lipgloss.Style
	statusHelpStyle             lipgloss.Style
	statusFilterTagStyle        lipgloss.Style
	searchInputStyle            lipgloss.Style
	searchPromptStyle           lipgloss.Style
)

// Color aliases used by view.go for status bar rendering.
var (
	colorBlue   lipgloss.TerminalColor
	colorGreen  lipgloss.TerminalColor
	colorRed    lipgloss.TerminalColor
	colorYellow lipgloss.TerminalColor
	colorPurple lipgloss.TerminalColor
	colorCyan   lipgloss.TerminalColor
	colorOrange lipgloss.TerminalColor
	colorFg     lipgloss.TerminalColor
	colorFgDim  lipgloss.TerminalColor
	colorComment lipgloss.TerminalColor
	colorGutter lipgloss.TerminalColor
	colorBg     lipgloss.TerminalColor
	colorSelection lipgloss.TerminalColor
)

func init() {
	ApplyTheme(nil)
}

// ApplyTheme sets the global color scheme and rebuilds all styles.
// Pass nil to use the default terminal-inherited colors.
func ApplyTheme(t *theme.Theme) {
	if t != nil {
		colors = themedColors(t)
	} else {
		colors = defaultColors()
	}

	// Publish color aliases for view.go
	colorBlue = colors.blue
	colorGreen = colors.green
	colorRed = colors.red
	colorYellow = colors.yellow
	colorPurple = colors.purple
	colorCyan = colors.cyan
	colorOrange = colors.orange
	colorFg = colors.fg
	colorFgDim = colors.fgDim
	colorComment = colors.comment
	colorGutter = colors.gutter
	colorBg = colors.bg
	colorSelection = colors.selection

	// Rebuild styles
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(colors.blue).
		PaddingLeft(1)

	selectedStyle = lipgloss.NewStyle().
		Background(colors.selection).
		Foreground(colors.fg).
		Bold(true)

	dirStyle = lipgloss.NewStyle().
		Foreground(colors.blue).
		Bold(true)

	fileStyle = lipgloss.NewStyle().
		Foreground(colors.fgDim)

	treeLineStyle = lipgloss.NewStyle().
		Foreground(colors.gutter)

	iconDirStyle = lipgloss.NewStyle().
		Foreground(colors.blue)

	iconFileStyle = lipgloss.NewStyle().
		Foreground(colors.fgDim)

	matchHighlightStyle = lipgloss.NewStyle().
		Foreground(colors.orange).
		Bold(true)

	matchHighlightSelectedStyle = lipgloss.NewStyle().
		Background(colors.selection).
		Foreground(colors.orange).
		Bold(true)

	flatPathStyle = lipgloss.NewStyle().
		Foreground(colors.comment)

	flatPathSelectedStyle = lipgloss.NewStyle().
		Background(colors.selection).
		Foreground(colors.fgDim)

	statusBase = lipgloss.NewStyle().
		Background(colors.bg)

	statusPathStyle = statusBase.Foreground(colors.fgDim).PaddingLeft(1).PaddingRight(1)
	statusBranchStyle = statusBase.Foreground(colors.purple).Bold(true)
	statusFlashStyle = statusBase.Foreground(colors.green).Bold(true).PaddingLeft(1)
	statusHelpStyle = statusBase.Foreground(colors.gutter)

	statusFilterTagStyle = lipgloss.NewStyle().
		Background(colors.cyan).
		Foreground(lipgloss.Color("0")).
		Bold(true)

	searchInputStyle = statusBase.Foreground(colors.fg).PaddingLeft(1)
	searchPromptStyle = statusBase.Foreground(colors.gutter).PaddingLeft(1)
}
