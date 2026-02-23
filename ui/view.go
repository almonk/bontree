package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/almonk/bontree/config"
	"github.com/almonk/bontree/icons"
	"github.com/almonk/bontree/tree"
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
		prompt := "\uf002"
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
	w := m.width
	if w < 20 {
		w = 20
	}
	chevron := "\ue0b0" // 

	// Determine mode label and color
	var modeLabel string
	var modeBg lipgloss.TerminalColor
	if m.filtered || m.searching {
		if m.flatSearch {
			modeLabel = "FFIND"
			modeBg = colorOrange
		} else {
			modeLabel = "FILTER"
			modeBg = colorPurple
		}
	} else {
		modeLabel = "NORMAL"
		modeBg = colorBlue
	}

	// Mode segment
	modeStyle := lipgloss.NewStyle().
		Background(modeBg).
		Foreground(lipgloss.Color("0")).
		Bold(true)
	modeChevronStyle := lipgloss.NewStyle().
		Foreground(modeBg)

	var right string
	if w >= 60 {
		right = statusHelpStyle.Render(" ?:help  c:copy  q:quit ")
	}

	var left string

	if m.flashMsg != "" {
		left = modeStyle.Render(" "+modeLabel+" ") +
			lipgloss.NewStyle().Foreground(modeBg).Background(colorBg).Render(chevron) +
			statusFlashStyle.Render(" "+m.flashMsg)
	} else {
		// Branch segment
		branchText := ""
		if m.gitBranch != "" {
			// Reserve space for mode segment (~10), chevrons (~2), and right side if visible
			reserved := 15 + lipgloss.Width(right)
			branchMax := w - reserved
			if branchMax < 10 {
				branchMax = 10
			}
			branch := m.gitBranch
			if runeLen(branch) > branchMax {
				branch, _ = middleTruncate(branch, branchMax, nil)
			}
			branchText = fmt.Sprintf(" \ue725 %s ", branch)
		}
		branchBg := lipgloss.AdaptiveColor{Light: "252", Dark: "237"}

		branchStyle := lipgloss.NewStyle().
			Background(branchBg).
			Foreground(colorBlue).
			Bold(true)
		branchChevronStyle := lipgloss.NewStyle().
			Foreground(branchBg).
			Background(colorBg)

		left = modeStyle.Render(" "+modeLabel+" ") +
			modeChevronStyle.Background(branchBg).Render(chevron)

		if branchText != "" {
			left += branchStyle.Render(branchText) +
				branchChevronStyle.Render(chevron)
		} else {
			left += lipgloss.NewStyle().Foreground(modeBg).Background(colorBg).Render(chevron)
		}
	}

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
		// Budget: give the name up to 40% of available, the rest to the path.
		nameWidth := runeLen(displayName)
		dirWidth := runeLen(displayDirPath)
		gap := 2 // separator between name and path

		nameMax := available * 2 / 5
		if nameMax < 8 {
			nameMax = 8
		}
		if nameWidth > nameMax {
			displayName, nameIndices = middleTruncate(displayName, nameMax, nameIndices)
			nameWidth = runeLen(displayName)
		}

		pathMax := available - nameWidth - gap
		if pathMax < 10 {
			pathMax = 10
		}
		if dirWidth > pathMax {
			displayDirPath, pathIndices = middleTruncate(displayDirPath, pathMax, pathIndices)
		}
	} else if runeLen(displayName) > available {
		displayName, nameIndices = middleTruncate(displayName, available, nameIndices)
	}

	if selected {
		treeLineSelectedStyle := lipgloss.NewStyle().Foreground(colorFgDim).Background(colorSelection)
		var parts []string
		parts = append(parts, selectedStyle.Render(" "))
		if prefix != "" {
			parts = append(parts, treeLineSelectedStyle.Render(prefix))
		}
		parts = append(parts, selectedStyle.Render(icon+" "))
		parts = append(parts, m.renderNameHighlighted(displayName, nameIndices, selectedStyle, matchHighlightSelectedStyle))

		if displayDirPath != "" {
			// Align dir path at 50% column, same as unselected rows
			leftWidth := lipgloss.Width(strings.Join(parts, ""))
			halfCol := maxWidth / 2
			gap := halfCol - leftWidth
			if gap < 2 {
				gap = 2
			}
			parts = append(parts, selectedStyle.Render(strings.Repeat(" ", gap)))
			parts = append(parts, m.renderNameHighlighted(displayDirPath, pathIndices, flatPathSelectedStyle, matchHighlightSelectedStyle))
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
		// Place dir path at 50% column
		leftWidth := lipgloss.Width(strings.Join(parts, ""))
		dirRendered := m.renderNameHighlighted(displayDirPath, pathIndices, flatPathStyle, matchHighlightStyle)
		halfCol := maxWidth / 2
		gap := halfCol - leftWidth
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

	// Build set of indices that are part of consecutive runs of 2+.
	// Single isolated matches are not highlighted to reduce visual noise.
	sorted := make([]int, len(matchIndices))
	copy(sorted, matchIndices)
	sort.Ints(sorted)

	matchSet := make(map[int]bool, len(sorted))
	for i, idx := range sorted {
		hasPrev := i > 0 && sorted[i-1] == idx-1
		hasNext := i < len(sorted)-1 && sorted[i+1] == idx+1
		if hasPrev || hasNext {
			matchSet[idx] = true
		}
	}

	if len(matchSet) == 0 {
		return baseStyle.Render(name)
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

// formatKeyName makes key names more readable for the help view.
func formatKeyName(key string) string {
	switch key {
	case "up":
		return "↑"
	case "down":
		return "↓"
	case "left":
		return "←"
	case "right":
		return "→"
	case " ":
		return "Space"
	case "enter":
		return "Enter"
	case "esc":
		return "Esc"
	case "backspace":
		return "Backspace"
	}
	// Capitalise ctrl+ shortcuts
	if strings.HasPrefix(key, "ctrl+") {
		return "Ctrl+" + strings.TrimPrefix(key, "ctrl+")
	}
	return key
}

func (m Model) helpView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  Keybindings"))
	b.WriteString("\n\n")

	// Action descriptions in display order
	actionOrder := []struct {
		action config.Action
		desc   string
	}{
		{config.ActionMoveDown, "Move down"},
		{config.ActionMoveUp, "Move up"},
		{config.ActionGoTop, "Go to top"},
		{config.ActionGoBottom, "Go to bottom"},
		{config.ActionHalfPageDown, "Half page down"},
		{config.ActionHalfPageUp, "Half page up"},
		{config.ActionExpand, "Expand directory"},
		{config.ActionCollapse, "Collapse directory / go to parent"},
		{config.ActionToggle, "Toggle directory open/close"},
		{config.ActionCopyPath, "Copy relative path to clipboard"},
		{config.ActionOpenEditor, "Open file in $EDITOR"},
		{config.ActionExpandAll, "Expand all"},
		{config.ActionCollapseAll, "Collapse all"},
		{config.ActionSearch, "Fuzzy search (tree)"},
		{config.ActionFlatSearch, "Flat file search"},
		{config.ActionToggleHidden, "Toggle hidden files"},
		{config.ActionClearFilter, "Clear filter"},
		{config.ActionHelp, "Toggle help"},
		{config.ActionQuit, "Quit"},
	}

	keyStyle := lipgloss.NewStyle().
		Foreground(colorPurple).
		Bold(true).
		Width(20).
		Align(lipgloss.Left).
		PaddingLeft(2)

	descStyle := lipgloss.NewStyle().
		Foreground(colorFgDim)

	for _, entry := range actionOrder {
		keys := m.cfg.KeysFor(entry.action)
		if len(keys) == 0 {
			continue
		}
		// Sort for deterministic display
		sort.Strings(keys)
		formatted := make([]string, len(keys))
		for i, k := range keys {
			formatted[i] = formatKeyName(k)
		}
		b.WriteString(keyStyle.Render(strings.Join(formatted, " / ")))
		b.WriteString(descStyle.Render(entry.desc))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Find the actual help key
	helpKeys := m.cfg.KeysFor(config.ActionHelp)
	helpHint := "Press ? to return"
	if len(helpKeys) > 0 {
		helpHint = fmt.Sprintf("Press %s to return", formatKeyName(helpKeys[0]))
	}
	b.WriteString(lipgloss.NewStyle().Foreground(colorComment).PaddingLeft(2).Render(helpHint))

	return b.String()
}
