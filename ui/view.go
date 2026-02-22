package ui

import (
	"fmt"
	"strings"

	"github.com/alasdairmonk/bontree/icons"
	"github.com/alasdairmonk/bontree/tree"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}
	if m.showHelp {
		return m.helpView()
	}

	var b strings.Builder

	viewH := m.viewportHeight()
	end := m.scrollOff + viewH
	if end > len(m.flatNodes) {
		end = len(m.flatNodes)
	}

	contentWidth := max(m.width, 20)

	// Render visible tree lines
	for i := m.scrollOff; i < end; i++ {
		if i > m.scrollOff {
			b.WriteString("\n")
		}
		b.WriteString(m.renderNode(m.flatNodes[i], i == m.cursor, contentWidth))
	}

	// Pad remaining lines
	for i := end - m.scrollOff; i < viewH; i++ {
		b.WriteString("\n")
	}

	// Search input (above status bar)
	if m.searching {
		b.WriteString("\n")
		prompt := "/"
		if m.flatSearch {
			prompt = "find:"
		}
		searchLine := searchPromptStyle.Render(prompt) + searchInputStyle.Render(m.searchQuery+"█")
		if sw := lipgloss.Width(searchLine); sw < m.width {
			searchLine += statusBase.Render(strings.Repeat(" ", m.width-sw))
		}
		b.WriteString(searchLine)
	}

	// Status bar
	b.WriteString("\n")
	b.WriteString(m.renderStatusBar())

	return b.String()
}

func (m Model) renderStatusBar() string {
	w := max(m.width, 80)

	var left string
	if m.flashMsg != "" {
		left = statusFlashStyle.Render(" " + m.flashMsg)
	} else if m.gitBranch != "" {
		left = statusBranchStyle.Render(fmt.Sprintf(" \ue725 %s", m.gitBranch))
	}

	right := statusHelpStyle.Render(" ?:help  c:copy  q:quit ")

	padding := w - lipgloss.Width(left) - lipgloss.Width(right)
	if padding < 0 {
		padding = 0
	}

	return left + statusBase.Render(strings.Repeat(" ", padding)) + right
}

func (m Model) renderNode(node *tree.Node, selected bool, maxWidth int) string {
	var prefix string
	if !m.flatSearch {
		prefix = m.getDisplayPrefix(node, node.Depth)
	}
	icon := icons.GetIcon(node.Name, node.IsDir, node.Expanded)
	matchIndices := m.searchMatchIndices[node]

	// In flat search mode, append the relative parent dir after the name
	var dirPath string
	if m.flatSearch && node.Parent != nil && node.Parent != m.root {
		dirPath = strings.TrimPrefix(node.Parent.Path, "./")
	}

	// Calculate available width for name (+ dirPath) and truncate if needed
	// Layout: " " + prefix + icon + " " + name [+ "  " + dirPath]
	prefixWidth := lipgloss.Width(prefix)
	iconWidth := lipgloss.Width(icon)
	fixedWidth := 1 + prefixWidth + iconWidth + 1
	available := maxWidth - fixedWidth
	if available < 4 {
		available = 4
	}

	displayName := node.Name
	nameIndices := matchIndices
	displayDirPath := dirPath
	pathIndices := m.searchPathIndices[node]

	if dirPath != "" {
		// "  " separator costs 2 columns
		nameWidth := runeLen(displayName)
		dirWidth := runeLen(displayDirPath)
		total := nameWidth + 2 + dirWidth
		if total > available {
			// Truncate dirPath first, then name if still needed
			dirAlloc := available - nameWidth - 2
			if dirAlloc < 3 {
				dirAlloc = 3
			}
			if dirWidth > dirAlloc {
				displayDirPath, pathIndices = middleTruncate(displayDirPath, dirAlloc, pathIndices)
			}
			nameAlloc := available - 2 - runeLen(displayDirPath)
			if nameAlloc < 4 {
				nameAlloc = 4
			}
			if nameWidth > nameAlloc {
				displayName, nameIndices = middleTruncate(displayName, nameAlloc, nameIndices)
			}
		}
	} else if runeLen(displayName) > available {
		displayName, nameIndices = middleTruncate(displayName, available, nameIndices)
	}

	if selected {
		treeLineSelectedStyle := treeLineStyle.Background(colorSelection)
		var parts []string
		parts = append(parts, selectedStyle.Render(" "))
		if prefix != "" {
			parts = append(parts, treeLineSelectedStyle.Render(prefix))
		}
		parts = append(parts, selectedStyle.Render(icon+" "))
		parts = append(parts, m.renderNameHighlighted(displayName, nameIndices, selectedStyle, matchHighlightSelectedStyle))

		if displayDirPath != "" {
			// Right-align the dir path
			leftWidth := lipgloss.Width(strings.Join(parts, ""))
			dirRendered := m.renderNameHighlighted(displayDirPath, pathIndices, flatPathSelectedStyle, matchHighlightSelectedStyle)
			dirWidth := lipgloss.Width(dirRendered)
			gap := maxWidth - leftWidth - dirWidth
			if gap < 2 {
				gap = 2
			}
			parts = append(parts, selectedStyle.Render(strings.Repeat(" ", gap)))
			parts = append(parts, dirRendered)
		}

		if plainLen := lipgloss.Width(strings.Join(parts, "")); plainLen < maxWidth {
			parts = append(parts, selectedStyle.Render(strings.Repeat(" ", maxWidth-plainLen)))
		}
		return strings.Join(parts, "")
	}

	var parts []string
	parts = append(parts, " ")
	if prefix != "" {
		parts = append(parts, treeLineStyle.Render(prefix))
	}

	iconStyle, nameStyle := m.gitNodeStyles(node)
	parts = append(parts, iconStyle.Render(icon)+" ")
	parts = append(parts, m.renderNameHighlighted(displayName, nameIndices, nameStyle, matchHighlightStyle))

	if displayDirPath != "" {
		// Right-align the dir path
		leftWidth := lipgloss.Width(strings.Join(parts, ""))
		dirRendered := m.renderNameHighlighted(displayDirPath, pathIndices, flatPathStyle, matchHighlightStyle)
		dirWidth := lipgloss.Width(dirRendered)
		gap := maxWidth - leftWidth - dirWidth
		if gap < 2 {
			gap = 2
		}
		parts = append(parts, strings.Repeat(" ", gap))
		parts = append(parts, dirRendered)
	}

	return strings.Join(parts, "")
}

// gitNodeStyles returns the icon and name styles for a node based on its git status.
func (m Model) gitNodeStyles(node *tree.Node) (lipgloss.Style, lipgloss.Style) {
	if m.gitFiles != nil {
		relPath := strings.TrimPrefix(node.Path, "./")
		if status, ok := m.gitFiles[relPath]; ok {
			switch status {
			case gitModified:
				return lipgloss.NewStyle().Foreground(colorBlue), lipgloss.NewStyle().Foreground(colorBlue)
			case gitAdded, gitUntracked:
				return lipgloss.NewStyle().Foreground(colorGreen), lipgloss.NewStyle().Foreground(colorGreen)
			case gitDeleted:
				return lipgloss.NewStyle().Foreground(colorRed), lipgloss.NewStyle().Foreground(colorRed)
			case gitIgnored:
				return lipgloss.NewStyle().Foreground(colorGutter), lipgloss.NewStyle().Foreground(colorGutter)
			}
		}
	}
	// Default styles
	if node.IsDir {
		return iconDirStyle, dirStyle
	}
	return iconFileStyle, fileStyle
}

func (m Model) renderNameHighlighted(name string, matchIndices []int, baseStyle, highlightStyle lipgloss.Style) string {
	if len(matchIndices) == 0 {
		return baseStyle.Render(name)
	}

	matchSet := make(map[int]bool, len(matchIndices))
	for _, idx := range matchIndices {
		matchSet[idx] = true
	}

	var result strings.Builder
	for i, ch := range name {
		if matchSet[i] {
			result.WriteString(highlightStyle.Render(string(ch)))
		} else {
			result.WriteString(baseStyle.Render(string(ch)))
		}
	}
	return result.String()
}

// getDisplayPrefix builds tree-drawing characters for a node, accounting for root being hidden
func (m Model) getDisplayPrefix(node *tree.Node, displayDepth int) string {
	if displayDepth <= 0 {
		return ""
	}

	var parts []string

	if node.IsLastChild() {
		parts = append(parts, "└─")
	} else {
		parts = append(parts, "├─")
	}

	current := node
	ancestor := node.Parent
	for ancestor != nil && ancestor.Depth > 0 {
		if !current.Parent.IsLastChild() {
			parts = append(parts, "│ ")
		} else {
			parts = append(parts, "  ")
		}
		current = ancestor
		ancestor = ancestor.Parent
	}

	// Reverse
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}

	return strings.Join(parts, "")
}

// runeLen returns the number of runes in a string.
func runeLen(s string) int {
	return len([]rune(s))
}

// middleTruncate truncates a string in the middle with "…" if it exceeds maxWidth runes.
// It also remaps match indices to their new positions in the truncated string.
func middleTruncate(s string, maxWidth int, indices []int) (string, []int) {
	runes := []rune(s)
	if len(runes) <= maxWidth {
		return s, indices
	}
	if maxWidth <= 1 {
		return "…", nil
	}

	rightLen := (maxWidth - 1) / 2
	leftLen := maxWidth - 1 - rightLen
	truncated := string(runes[:leftLen]) + "…" + string(runes[len(runes)-rightLen:])

	if len(indices) == 0 {
		return truncated, nil
	}

	rightStart := len(runes) - rightLen
	var remapped []int
	for _, idx := range indices {
		if idx < leftLen {
			remapped = append(remapped, idx)
		} else if idx >= rightStart {
			remapped = append(remapped, leftLen+1+(idx-rightStart))
		}
		// Indices in the truncated middle are dropped
	}
	return truncated, remapped
}

// --- Help ---

func (m Model) helpView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  Keybindings"))
	b.WriteString("\n\n")

	bindings := []struct{ key, desc string }{
		{"j / ↓", "Move down"},
		{"k / ↑", "Move up"},
		{"g", "Go to top"},
		{"G", "Go to bottom"},
		{"Ctrl+d", "Half page down"},
		{"Ctrl+u", "Half page up"},
		{"l / →", "Expand directory"},
		{"h / ←", "Collapse directory / go to parent"},
		{"Enter", "Toggle directory open/close"},
		{"Space", "Toggle directory open/close"},
		{"c", "Copy relative path to clipboard"},
		{"E", "Expand all"},
		{"W", "Collapse all"},
		{"/", "Fuzzy search (tree)"},
		{"Ctrl+f", "Flat file search"},
		{".", "Toggle hidden files"},
		{"?", "Toggle help"},
		{"q / Ctrl+c", "Quit"},
	}

	keyStyle := lipgloss.NewStyle().
		Foreground(colorPurple).
		Bold(true).
		Width(16).
		Align(lipgloss.Left).
		PaddingLeft(2)

	descStyle := lipgloss.NewStyle().
		Foreground(colorFgDim)

	for _, bind := range bindings {
		b.WriteString(keyStyle.Render(bind.key))
		b.WriteString(descStyle.Render(bind.desc))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(colorComment).PaddingLeft(2).Render("Press ? to return"))

	return b.String()
}
