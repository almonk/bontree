// WASM entry point for the bontree interactive demo.
// This is a self-contained reimplementation of the TUI that runs in the browser
// via xterm.js, using the real tree, icons, styles, and rendering logic.
package main

import (
	"fmt"
	"sort"
	"strings"
	"syscall/js"
	"unicode/utf8"

	"github.com/alasdairmonk/bontree/config"
	"github.com/alasdairmonk/bontree/icons"
	"github.com/alasdairmonk/bontree/tree"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

// ── Colors (Tokyo Night) ──

var (
	colorBlue      lipgloss.TerminalColor = lipgloss.Color("#7aa2f7")
	colorGreen     lipgloss.TerminalColor = lipgloss.Color("#9ece6a")
	colorRed       lipgloss.TerminalColor = lipgloss.Color("#f7768e")
	colorYellow    lipgloss.TerminalColor = lipgloss.Color("#e0af68")
	colorPurple    lipgloss.TerminalColor = lipgloss.Color("#bb9af7")
	colorCyan      lipgloss.TerminalColor = lipgloss.Color("#7dcfff")
	colorOrange    lipgloss.TerminalColor = lipgloss.Color("#ff9e64")
	colorFg        lipgloss.TerminalColor = lipgloss.Color("#c0caf5")
	colorFgDim     lipgloss.TerminalColor = lipgloss.Color("#9aa5ce")
	colorComment   lipgloss.TerminalColor = lipgloss.Color("#565f89")
	colorGutter    lipgloss.TerminalColor = lipgloss.Color("#3b4261")
	colorBg        lipgloss.TerminalColor = lipgloss.Color("#1f2335")
	colorSelection lipgloss.TerminalColor = lipgloss.Color("#33467c")
)

// ── Styles ──

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
	statusFlashStyle            lipgloss.Style
	statusHelpStyle             lipgloss.Style
	searchInputStyle            lipgloss.Style
	searchPromptStyle           lipgloss.Style
)

func initStyles() {
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(colorBlue).PaddingLeft(1)
	selectedStyle = lipgloss.NewStyle().Background(colorSelection).Foreground(colorFg).Bold(true)
	dirStyle = lipgloss.NewStyle().Foreground(colorBlue).Bold(true)
	fileStyle = lipgloss.NewStyle().Foreground(colorFgDim)
	treeLineStyle = lipgloss.NewStyle().Foreground(colorGutter)
	iconDirStyle = lipgloss.NewStyle().Foreground(colorBlue)
	iconFileStyle = lipgloss.NewStyle().Foreground(colorFgDim)
	matchHighlightStyle = lipgloss.NewStyle().Foreground(colorOrange).Bold(true)
	matchHighlightSelectedStyle = lipgloss.NewStyle().Background(colorSelection).Foreground(colorOrange).Bold(true)
	flatPathStyle = lipgloss.NewStyle().Foreground(colorComment)
	flatPathSelectedStyle = lipgloss.NewStyle().Background(colorSelection).Foreground(colorFgDim)
	statusBase = lipgloss.NewStyle().Background(colorBg)
	statusFlashStyle = statusBase.Foreground(colorGreen).Bold(true).PaddingLeft(1)
	statusHelpStyle = statusBase.Foreground(colorGutter)
	searchInputStyle = statusBase.Foreground(colorFg).PaddingLeft(1)
	searchPromptStyle = statusBase.Foreground(colorGutter).PaddingLeft(1)
}

// ── Git status ──

type gitFileStatus int

const (
	gitUnchanged gitFileStatus = iota
	gitModified
	gitAdded
	gitDeleted
	gitUntracked
	gitIgnored
)

// ── Model ──

type model struct {
	root      *tree.Node
	flatNodes []*tree.Node
	cursor    int
	width     int
	height    int
	scrollOff int

	gitBranch string
	gitFiles  map[string]gitFileStatus

	showHelp  bool
	flashMsg  string
	searching bool
	flatSearch bool
	filtered  bool
	searchQuery string
	searchNodes []*tree.Node
	searchMatchIndices map[*tree.Node][]int
	searchPathIndices  map[*tree.Node][]int
	savedExpanded  map[*tree.Node]bool
	savedCursor    int
	savedScrollOff int

	cfg *config.Config
}

// ── Fuzzy sources (same as ui/search.go) ──

type nodeSource []*tree.Node
func (ns nodeSource) String(i int) string { return strings.TrimPrefix(ns[i].Path, "./") }
func (ns nodeSource) Len() int            { return len(ns) }

type nodeNameSource []*tree.Node
func (ns nodeNameSource) String(i int) string { return ns[i].Name }
func (ns nodeNameSource) Len() int            { return len(ns) }

func splitMatchIndices(indices []int, path, name string) (nameIndices, pathIndices []int) {
	nameOffset := len(path) - len(name)
	for _, idx := range indices {
		if idx >= nameOffset {
			nameIndices = append(nameIndices, idx-nameOffset)
		} else {
			pathIndices = append(pathIndices, idx)
		}
	}
	return
}

// ── Tree helpers ──

func flattenVisible(root *tree.Node) []*tree.Node {
	var result []*tree.Node
	var walk func(n *tree.Node)
	walk = func(n *tree.Node) {
		result = append(result, n)
		if n.IsDir && n.Expanded {
			for _, c := range n.Children {
				walk(c)
			}
		}
	}
	walk(root)
	return result
}

func flattenAll(root *tree.Node) []*tree.Node {
	var result []*tree.Node
	var walk func(n *tree.Node)
	walk = func(n *tree.Node) {
		result = append(result, n)
		if n.IsDir {
			for _, c := range n.Children {
				walk(c)
			}
		}
	}
	walk(root)
	if len(result) > 0 {
		return result[1:]
	}
	return result
}

func setExpandAll(node *tree.Node, expanded bool) {
	if node.IsDir {
		node.Expanded = expanded
		for _, c := range node.Children {
			setExpandAll(c, expanded)
		}
	}
}

// ── Build demo tree ──

func buildDemoTree() *tree.Node {
	root := &tree.Node{Name: "my-project", Path: ".", IsDir: true, Expanded: true, Depth: 0}

	add := func(parent *tree.Node, name string, isDir bool, expanded bool) *tree.Node {
		n := &tree.Node{
			Name:     name,
			Path:     strings.TrimPrefix(parent.Path+"/"+name, "./"),
			IsDir:    isDir,
			Expanded: expanded,
			Parent:   parent,
			Depth:    parent.Depth + 1,
			Loaded:   true,
		}
		parent.Children = append(parent.Children, n)
		return n
	}

	// .github
	gh := add(root, ".github", true, false)
	wf := add(gh, "workflows", true, false)
	add(wf, "ci.yml", false, false)

	// cmd
	cmd := add(root, "cmd", true, true)
	srv := add(cmd, "server", true, true)
	add(srv, "main.go", false, false)
	add(srv, "routes.go", false, false)
	add(srv, "middleware.go", false, false)
	cli := add(cmd, "cli", true, false)
	add(cli, "root.go", false, false)
	add(cli, "serve.go", false, false)

	// internal
	internal := add(root, "internal", true, true)
	auth := add(internal, "auth", true, true)
	add(auth, "jwt.go", false, false)
	add(auth, "jwt_test.go", false, false)
	add(auth, "oauth.go", false, false)
	db := add(internal, "db", true, false)
	add(db, "postgres.go", false, false)
	add(db, "migrations.go", false, false)
	add(db, "schema.sql", false, false)
	handlers := add(internal, "handlers", true, false)
	add(handlers, "users.go", false, false)
	add(handlers, "posts.go", false, false)
	add(handlers, "health.go", false, false)
	models := add(internal, "models", true, false)
	add(models, "user.go", false, false)
	add(models, "post.go", false, false)

	// web
	web := add(root, "web", true, false)
	src := add(web, "src", true, false)
	add(src, "App.tsx", false, false)
	add(src, "index.tsx", false, false)
	comps := add(src, "components", true, false)
	add(comps, "Header.tsx", false, false)
	add(comps, "Footer.tsx", false, false)
	add(web, "package.json", false, false)
	add(web, "tsconfig.json", false, false)
	add(web, "vite.config.ts", false, false)

	// docs
	docs := add(root, "docs", true, false)
	add(docs, "architecture.md", false, false)
	add(docs, "api.md", false, false)
	add(docs, "deployment.md", false, false)

	// root files
	add(root, ".env.example", false, false)
	add(root, ".gitignore", false, false)
	add(root, "docker-compose.yml", false, false)
	add(root, "Dockerfile", false, false)
	add(root, "go.mod", false, false)
	add(root, "go.sum", false, false)
	add(root, "Makefile", false, false)
	add(root, "README.md", false, false)

	return root
}

func demoGitFiles() map[string]gitFileStatus {
	return map[string]gitFileStatus{
		"cmd/server/main.go":       gitModified,
		"cmd/server/middleware.go":  gitAdded,
		"cmd/server":               gitModified,
		"cmd":                      gitModified,
		"internal/auth/jwt.go":     gitModified,
		"internal/auth/jwt_test.go": gitModified,
		"internal/auth/oauth.go":   gitAdded,
		"internal/auth":            gitModified,
		"internal/handlers/posts.go": gitModified,
		"internal/handlers":        gitModified,
		"internal":                 gitModified,
		"docs/api.md":              gitAdded,
		"docs":                     gitAdded,
		"Dockerfile":               gitModified,
	}
}

// ── Model methods ──

func (m *model) viewportHeight() int {
	h := m.height - 1
	if m.searching {
		h--
	}
	if h < 1 {
		h = 1
	}
	return h
}

func (m *model) refreshFlatNodes() {
	m.flatNodes = flattenVisible(m.root)
	m.clampCursor()
}

func (m *model) clampCursor() {
	if m.cursor >= len(m.flatNodes) {
		m.cursor = len(m.flatNodes) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m *model) ensureVisible() {
	vh := m.viewportHeight()
	if m.cursor < m.scrollOff {
		m.scrollOff = m.cursor
	}
	if m.cursor >= m.scrollOff+vh {
		m.scrollOff = m.cursor - vh + 1
	}
}

func (m *model) moveCursor(delta int) {
	m.cursor += delta
	m.clampCursor()
	m.ensureVisible()
}

// ── View (ported from ui/view.go) ──

func (m *model) view() string {
	if m.width == 0 || m.height == 0 {
		return ""
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

	for i := m.scrollOff; i < end; i++ {
		if i > m.scrollOff {
			b.WriteString("\n")
		}
		b.WriteString(m.renderNode(m.flatNodes[i], i == m.cursor, contentWidth))
	}
	for i := end - m.scrollOff; i < viewH; i++ {
		b.WriteString("\n")
	}
	if m.searching {
		b.WriteString("\n")
		prompt := "\uf002"
		searchLine := searchPromptStyle.Render(prompt) + searchInputStyle.Render(m.searchQuery+"█")
		if sw := lipgloss.Width(searchLine); sw < m.width {
			searchLine += statusBase.Render(strings.Repeat(" ", m.width-sw))
		}
		b.WriteString(searchLine)
	}
	b.WriteString("\n")
	b.WriteString(m.renderStatusBar())
	return b.String()
}

func (m *model) renderStatusBar() string {
	w := m.width
	if w < 20 {
		w = 20
	}
	chevron := "\ue0b0"

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

	modeStyle := lipgloss.NewStyle().Background(modeBg).Foreground(lipgloss.Color("0")).Bold(true)
	modeChevronStyle := lipgloss.NewStyle().Foreground(modeBg)

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
		branchText := ""
		if m.gitBranch != "" {
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
		branchBg := lipgloss.Color("237")
		branchStyle := lipgloss.NewStyle().Background(branchBg).Foreground(colorBlue).Bold(true)
		branchChevronStyle := lipgloss.NewStyle().Foreground(branchBg).Background(colorBg)

		left = modeStyle.Render(" "+modeLabel+" ") + modeChevronStyle.Background(branchBg).Render(chevron)
		if branchText != "" {
			left += branchStyle.Render(branchText) + branchChevronStyle.Render(chevron)
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

func (m *model) renderNode(node *tree.Node, selected bool, maxWidth int) string {
	var prefix string
	if !m.flatSearch {
		prefix = getDisplayPrefix(node, node.Depth)
	}
	icon := icons.GetIcon(node.Name, node.IsDir, node.Expanded)
	matchIndices := m.searchMatchIndices[node]

	var dirPath string
	if m.flatSearch && node.Parent != nil && node.Parent != m.root {
		dirPath = strings.TrimPrefix(node.Parent.Path, "./")
	}

	prefixWidth := lipgloss.Width(prefix)
	iconWidth := lipgloss.Width(icon)
	thinSpace := 0
	if prefixWidth > 0 {
		thinSpace = 1
	}
	fixedWidth := 1 + prefixWidth + thinSpace + iconWidth + 1
	available := maxWidth - fixedWidth
	if available < 4 {
		available = 4
	}

	displayName := node.Name
	nameIndices := matchIndices
	displayDirPath := dirPath
	pathIndices := m.searchPathIndices[node]

	if dirPath != "" {
		nameWidth := runeLen(displayName)
		dirWidth := runeLen(displayDirPath)
		gap := 2
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
			parts = append(parts, treeLineSelectedStyle.Render(prefix)+selectedStyle.Render("\u2009"))
		}
		parts = append(parts, selectedStyle.Render(icon+" "))
		parts = append(parts, renderNameHighlighted(displayName, nameIndices, selectedStyle, matchHighlightSelectedStyle))

		if displayDirPath != "" {
			leftWidth := lipgloss.Width(strings.Join(parts, ""))
			halfCol := maxWidth / 2
			gap := halfCol - leftWidth
			if gap < 2 {
				gap = 2
			}
			parts = append(parts, selectedStyle.Render(strings.Repeat(" ", gap)))
			parts = append(parts, renderNameHighlighted(displayDirPath, pathIndices, flatPathSelectedStyle, matchHighlightSelectedStyle))
		}

		if plainLen := lipgloss.Width(strings.Join(parts, "")); plainLen < maxWidth {
			parts = append(parts, selectedStyle.Render(strings.Repeat(" ", maxWidth-plainLen)))
		}

		// Strip intermediate RESETs so the selection background is continuous.
		// Each lipgloss.Render() fully specifies its style (fg+bg+bold), so
		// removing RESETs between segments is safe — the next SGR overrides.
		// This prevents wide Nerd Font glyphs clipping in xterm.js.
		line := strings.Join(parts, "")
		line = strings.ReplaceAll(line, "\x1b[0m", "")
		line += "\x1b[0m"
		return line
	}

	var parts []string
	parts = append(parts, " ")
	if prefix != "" {
		parts = append(parts, treeLineStyle.Render(prefix)+"\u2009")
	}

	iconStyle, nameStyle := m.gitNodeStyles(node)
	parts = append(parts, iconStyle.Render(icon)+" ")
	parts = append(parts, renderNameHighlighted(displayName, nameIndices, nameStyle, matchHighlightStyle))

	if displayDirPath != "" {
		leftWidth := lipgloss.Width(strings.Join(parts, ""))
		dirRendered := renderNameHighlighted(displayDirPath, pathIndices, flatPathStyle, matchHighlightStyle)
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

func (m *model) gitNodeStyles(node *tree.Node) (lipgloss.Style, lipgloss.Style) {
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
	if node.IsDir {
		return iconDirStyle, dirStyle
	}
	return iconFileStyle, fileStyle
}

func renderNameHighlighted(name string, matchIndices []int, baseStyle, highlightStyle lipgloss.Style) string {
	if len(matchIndices) == 0 {
		return baseStyle.Render(name)
	}
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

func getDisplayPrefix(node *tree.Node, displayDepth int) string {
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
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	return strings.Join(parts, "")
}

func runeLen(s string) int { return len([]rune(s)) }

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
	}
	return truncated, remapped
}

// ── Help view ──

func (m *model) helpView() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("  Keybindings"))
	b.WriteString("\n\n")

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
		{config.ActionCollapse, "Collapse / go to parent"},
		{config.ActionToggle, "Toggle directory open/close"},
		{config.ActionCopyPath, "Copy relative path to clipboard"},
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
		Foreground(colorPurple).Bold(true).Width(20).Align(lipgloss.Left).PaddingLeft(2)
	descStyle := lipgloss.NewStyle().Foreground(colorFgDim)

	for _, entry := range actionOrder {
		keys := m.cfg.KeysFor(entry.action)
		if len(keys) == 0 {
			continue
		}
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
	helpKeys := m.cfg.KeysFor(config.ActionHelp)
	helpHint := "Press ? to return"
	if len(helpKeys) > 0 {
		helpHint = fmt.Sprintf("Press %s to return", formatKeyName(helpKeys[0]))
	}
	b.WriteString(lipgloss.NewStyle().Foreground(colorComment).PaddingLeft(2).Render(helpHint))
	return b.String()
}

func formatKeyName(key string) string {
	switch key {
	case "up":    return "↑"
	case "down":  return "↓"
	case "left":  return "←"
	case "right": return "→"
	case " ":     return "Space"
	case "enter": return "Enter"
	case "esc":   return "Esc"
	}
	if strings.HasPrefix(key, "ctrl+") {
		return "Ctrl+" + strings.TrimPrefix(key, "ctrl+")
	}
	return key
}

// ── Search ──

func (m *model) saveExpandedState() {
	m.savedExpanded = make(map[*tree.Node]bool)
	for _, n := range flattenAll(m.root) {
		if n.IsDir {
			m.savedExpanded[n] = n.Expanded
		}
	}
	m.savedCursor = m.cursor
	m.savedScrollOff = m.scrollOff
}

func (m *model) restoreExpandedState() {
	if m.savedExpanded == nil {
		return
	}
	for node, expanded := range m.savedExpanded {
		node.Expanded = expanded
	}
	m.refreshFlatNodes()
	m.cursor = m.savedCursor
	m.clampCursor()
	m.scrollOff = m.savedScrollOff
	m.savedExpanded = nil
	m.ensureVisible()
}

func (m *model) startSearch(flat bool) {
	m.saveExpandedState()
	m.searching = true
	m.filtered = false
	m.flatSearch = flat
	m.searchQuery = ""
	m.searchNodes = nil
	m.searchMatchIndices = nil
	m.searchPathIndices = nil
	if !flat {
		setExpandAll(m.root, true)
	}
	m.refreshFlatNodes()
	m.cursor = 0
	m.scrollOff = 0
}

func (m *model) applySearchFilter() {
	if m.flatSearch {
		m.updateFlatSearch()
	} else {
		m.updateSearch()
	}
	if m.searchNodes != nil {
		m.flatNodes = m.searchNodes
	} else {
		m.flatNodes = flattenVisible(m.root)
	}
	m.cursor = 0
	m.scrollOff = 0
}

func (m *model) jumpToMatch(dir int) {
	if m.searchMatchIndices == nil || len(m.flatNodes) == 0 {
		return
	}
	n := len(m.flatNodes)
	for step := 1; step < n; step++ {
		idx := (m.cursor + dir*step%n + n) % n
		if _, ok := m.searchMatchIndices[m.flatNodes[idx]]; ok {
			m.cursor = idx
			m.ensureVisible()
			return
		}
	}
}

func (m *model) updateSearch() {
	if m.searchQuery == "" {
		m.searchNodes = nil
		m.searchMatchIndices = nil
		m.searchPathIndices = nil
		return
	}
	allNodes := flattenAll(m.root)
	results := fuzzy.FindFrom(m.searchQuery, nodeNameSource(allNodes))
	nameMap := make(map[*tree.Node][]int)
	matchSet := make(map[*tree.Node]bool)
	for _, r := range results {
		node := allNodes[r.Index]
		nameMap[node] = r.MatchedIndexes
		matchSet[node] = true
		for ancestor := node.Parent; ancestor != nil && !matchSet[ancestor]; ancestor = ancestor.Parent {
			matchSet[ancestor] = true
		}
	}
	var filtered []*tree.Node
	for _, node := range allNodes {
		if matchSet[node] && node != m.root {
			filtered = append(filtered, node)
		}
	}
	m.searchNodes = filtered
	m.searchMatchIndices = nameMap
	m.searchPathIndices = nil
}

func (m *model) updateFlatSearch() {
	if m.searchQuery == "" {
		m.searchNodes = nil
		m.searchMatchIndices = nil
		m.searchPathIndices = nil
		return
	}
	allNodes := flattenAll(m.root)
	results := fuzzy.FindFrom(m.searchQuery, nodeSource(allNodes))
	nameMap := make(map[*tree.Node][]int)
	pathMap := make(map[*tree.Node][]int)
	var files, dirs []*tree.Node
	for _, r := range results {
		node := allNodes[r.Index]
		path := strings.TrimPrefix(node.Path, "./")
		var nameIdx, pathIdx []int
		if nr := fuzzy.Find(m.searchQuery, []string{node.Name}); len(nr) > 0 {
			nameIdx = nr[0].MatchedIndexes
		} else {
			nameIdx, _ = splitMatchIndices(r.MatchedIndexes, path, node.Name)
		}
		if node.Parent != nil && node.Parent != m.root {
			dirPath := strings.TrimPrefix(node.Parent.Path, "./")
			if pr := fuzzy.Find(m.searchQuery, []string{dirPath}); len(pr) > 0 {
				pathIdx = pr[0].MatchedIndexes
			} else {
				_, pathIdx = splitMatchIndices(r.MatchedIndexes, path, node.Name)
			}
		}
		nameMap[node] = nameIdx
		pathMap[node] = pathIdx
		if node.IsDir {
			dirs = append(dirs, node)
		} else {
			files = append(files, node)
		}
	}
	m.searchNodes = append(files, dirs...)
	m.searchMatchIndices = nameMap
	m.searchPathIndices = pathMap
}

// ── Input handling ──

func (m *model) handleNormalKey(key string) {
	action := m.cfg.ActionFor(key)
	switch action {
	case config.ActionClearFilter:
		if m.filtered {
			m.filtered = false
			m.flatSearch = false
			m.searchQuery = ""
			m.searchNodes = nil
			m.searchMatchIndices = nil
			m.searchPathIndices = nil
			m.restoreExpandedState()
		}
	case config.ActionMoveDown:
		m.moveCursor(1)
	case config.ActionMoveUp:
		m.moveCursor(-1)
	case config.ActionGoTop:
		m.cursor = 0
		m.scrollOff = 0
	case config.ActionGoBottom:
		m.cursor = len(m.flatNodes) - 1
		m.ensureVisible()
	case config.ActionHalfPageDown:
		m.moveCursor(m.viewportHeight() / 2)
	case config.ActionHalfPageUp:
		m.moveCursor(-m.viewportHeight() / 2)
	case config.ActionExpand:
		if m.filtered {
			m.jumpToMatch(1)
		} else {
			node := m.flatNodes[m.cursor]
			if node.IsDir && !node.Expanded {
				node.Expanded = true
				m.refreshFlatNodes()
			}
		}
	case config.ActionCollapse:
		if m.filtered {
			m.jumpToMatch(-1)
		} else {
			node := m.flatNodes[m.cursor]
			if node.IsDir && node.Expanded {
				node.Expanded = false
				m.refreshFlatNodes()
			} else if node.Parent != nil {
				for i, n := range m.flatNodes {
					if n == node.Parent {
						m.cursor = i
						m.ensureVisible()
						break
					}
				}
			}
		}
	case config.ActionToggle:
		node := m.flatNodes[m.cursor]
		if node.IsDir {
			node.Expanded = !node.Expanded
			m.refreshFlatNodes()
		}
	case config.ActionCopyPath:
		node := m.flatNodes[m.cursor]
		relPath := strings.TrimPrefix(node.Path, "./")
		m.flashMsg = fmt.Sprintf("✓ Copied path: %s", relPath)
		// Copy to clipboard via JS
		js.Global().Get("navigator").Get("clipboard").Call("writeText", relPath)
	case config.ActionExpandAll:
		setExpandAll(m.root, true)
		m.refreshFlatNodes()
	case config.ActionCollapseAll:
		setExpandAll(m.root, false)
		m.root.Expanded = true
		m.refreshFlatNodes()
		m.cursor = 0
		m.scrollOff = 0
	case config.ActionSearch:
		m.startSearch(false)
	case config.ActionFlatSearch:
		m.startSearch(true)
	case config.ActionHelp:
		m.showHelp = !m.showHelp
	}
}

func (m *model) handleSearchKey(key string) {
	action := m.cfg.ActionFor(key)
	switch {
	case key == "esc":
		if m.searchQuery == "" {
			m.searching = false
			m.filtered = false
			m.flatSearch = false
			m.searchNodes = nil
			m.searchMatchIndices = nil
			m.searchPathIndices = nil
			m.restoreExpandedState()
		} else {
			m.searching = false
			m.filtered = true
			if m.searchNodes != nil {
				m.flatNodes = m.searchNodes
			}
			m.clampCursor()
		}
	case key == "enter":
		m.searching = false
		m.filtered = true
		if m.searchNodes != nil {
			m.flatNodes = m.searchNodes
		}
		m.clampCursor()
	case key == "backspace":
		if len(m.searchQuery) > 0 {
			_, size := utf8.DecodeLastRuneInString(m.searchQuery)
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-size]
			m.applySearchFilter()
		}
	case action == config.ActionMoveDown || key == "down":
		m.moveCursor(1)
	case action == config.ActionMoveUp || key == "up":
		m.moveCursor(-1)
	case action == config.ActionExpand || key == "right":
		m.jumpToMatch(1)
	case action == config.ActionCollapse || key == "left":
		m.jumpToMatch(-1)
	default:
		// Printable character
		if len(key) == 1 && key[0] >= 32 {
			m.searchQuery += key
			m.applySearchFilter()
		}
	}
}

// ── JS bridge ──

var app *model

func main() {
	initStyles()

	root := buildDemoTree()
	cfg := config.DefaultConfig()

	app = &model{
		root:      root,
		flatNodes: flattenVisible(root),
		gitBranch: "main",
		gitFiles:  demoGitFiles(),
		cfg:       cfg,
	}

	// Expose functions to JS
	js.Global().Set("bontreeInit", js.FuncOf(bontreeInit))
	js.Global().Set("bontreeKey", js.FuncOf(bontreeKey))
	js.Global().Set("bontreeClick", js.FuncOf(bontreeClick))
	js.Global().Set("bontreeScroll", js.FuncOf(bontreeScroll))
	js.Global().Set("bontreeClearFlash", js.FuncOf(bontreeClearFlash))

	// Keep alive
	select {}
}

func bontreeInit(_ js.Value, args []js.Value) interface{} {
	cols := args[0].Int()
	rows := args[1].Int()
	app.width = cols
	app.height = rows
	return app.view()
}

func bontreeKey(_ js.Value, args []js.Value) interface{} {
	key := args[0].String()

	if app.showHelp {
		if key == "?" || key == "q" || key == "esc" {
			app.showHelp = false
		}
		return app.view()
	}

	if app.searching {
		app.handleSearchKey(key)
	} else {
		app.handleNormalKey(key)
	}

	result := js.Global().Get("Object").New()
	result.Set("view", app.view())
	result.Set("flash", app.flashMsg != "")
	return result
}

func bontreeClick(_ js.Value, args []js.Value) interface{} {
	row := args[0].Int()
	doubleClick := args[1].Bool()

	if app.showHelp {
		app.showHelp = false
		return app.view()
	}

	target := row + app.scrollOff
	if target < 0 || target >= len(app.flatNodes) {
		return app.view()
	}

	if doubleClick {
		node := app.flatNodes[target]
		if node.IsDir {
			node.Expanded = !node.Expanded
			app.refreshFlatNodes()
		}
	} else {
		app.cursor = target
		app.ensureVisible()
	}
	return app.view()
}

func bontreeScroll(_ js.Value, args []js.Value) interface{} {
	dir := args[0].Int()
	app.moveCursor(dir * 3)
	return app.view()
}

func bontreeClearFlash(_ js.Value, _ []js.Value) interface{} {
	app.flashMsg = ""
	return app.view()
}
