package ui

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/alasdairmonk/altree/icons"
	"github.com/alasdairmonk/altree/tree"
	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

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
			Foreground(lipgloss.Color(colorCyan))

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

// gitFileStatus represents the git state of a file
type gitFileStatus int

const (
	gitUnchanged  gitFileStatus = iota
	gitModified                         // working tree modified
	gitAdded                            // staged / new tracked file
	gitDeleted                          // deleted
	gitUntracked                        // untracked (?)
)

type clearFlashMsg struct{}
type gitRefreshMsg struct{}
type gitInfoMsg struct {
	branch     string
	fileStatus map[string]gitFileStatus // relative path -> status
}



// Model is the Bubble Tea model
type Model struct {
	root      *tree.Node
	flatNodes []*tree.Node
	cursor    int
	width     int
	height    int
	rootPath  string
	flashMsg  string
	showHelp  bool
	scrollOff int
	gitBranch  string
	gitFiles   map[string]gitFileStatus // relative path -> status
	showHidden bool

	// Mouse
	lastClickTime time.Time
	lastClickRow  int



	// Search
	searching          bool
	filtered           bool
	flatSearch         bool // true = flat file search (ctrl+f), false = hierarchy search (/)
	searchQuery        string
	searchNodes        []*tree.Node
	searchMatchIndices map[*tree.Node][]int // match indices within the node name
	searchPathIndices  map[*tree.Node][]int // match indices within the parent path (flat search)

	// Saved state before search
	savedExpanded  map[*tree.Node]bool
	savedCursor    int
	savedScrollOff int
}

// New creates a new Model
func New(rootPath string) (Model, error) {
	root, err := tree.BuildTree(rootPath)
	if err != nil {
		return Model{}, err
	}

	return Model{
		root:       root,
		flatNodes:  flattenSkipRoot(root),
		rootPath:   rootPath,
		showHidden: tree.ShowHidden,
	}, nil
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(fetchGitInfo(m.rootPath), gitRefreshTick())
}

// --- Helpers to reduce repetition ---

// refreshFlatNodes rebuilds the flat node list and clamps the cursor.
func (m *Model) refreshFlatNodes() {
	m.flatNodes = flattenSkipRoot(m.root)
	m.clampCursor()
}

// clampCursor ensures cursor is within bounds.
func (m *Model) clampCursor() {
	if m.cursor >= len(m.flatNodes) {
		m.cursor = len(m.flatNodes) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

// moveCursor moves the cursor by delta and ensures it stays visible.
func (m *Model) moveCursor(delta int) {
	m.cursor += delta
	m.clampCursor()
	m.ensureVisible()
}

// setExpandAll recursively sets the expanded state on all directory nodes.
func (m *Model) setExpandAll(node *tree.Node, expanded bool) {
	if node.IsDir {
		if expanded {
			node.Expand()
		} else {
			node.Collapse()
		}
		for _, child := range node.Children {
			m.setExpandAll(child, expanded)
		}
	}
}

// jumpToMatch moves cursor to the next (dir=+1) or previous (dir=-1) fuzzy match.
func (m *Model) jumpToMatch(dir int) {
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

// applySearchFilter runs the appropriate fuzzy search and updates flatNodes/cursor.
func (m *Model) applySearchFilter() {
	if m.flatSearch {
		m.updateFlatSearch()
	} else {
		m.updateSearch()
	}
	if m.searchNodes != nil {
		m.flatNodes = m.searchNodes
	} else {
		m.flatNodes = flattenSkipRoot(m.root)
	}
	m.cursor = 0
	m.scrollOff = 0
}

// flash sets a temporary flash message that auto-clears.
func flash(m *Model, msg string) tea.Cmd {
	m.flashMsg = msg
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return clearFlashMsg{}
	})
}

// --- Update ---

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case clearFlashMsg:
		m.flashMsg = ""
		return m, nil

	case gitInfoMsg:
		m.gitBranch = msg.branch
		m.gitFiles = msg.fileStatus
		if !m.searching {
			m.refreshTree()
		}
		return m, gitRefreshTick()

	case gitRefreshMsg:
		return m, fetchGitInfo(m.rootPath)

	case tea.MouseMsg:
		return m.updateMouse(msg)

	case tea.KeyMsg:
		if m.searching {
			return m.updateSearchMode(msg)
		}
		return m.updateNormalMode(msg)
	}

	return m, nil
}

func (m Model) updateMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.Button == tea.MouseButtonWheelUp:
		m.moveCursor(-3)
	case msg.Button == tea.MouseButtonWheelDown:
		m.moveCursor(3)
	case msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress:
		row := msg.Y + m.scrollOff
		if row < 0 || row >= len(m.flatNodes) {
			break
		}

		now := time.Now()
		doubleClick := row == m.lastClickRow && now.Sub(m.lastClickTime) < 400*time.Millisecond
		m.lastClickTime = now
		m.lastClickRow = row

		if doubleClick {
			node := m.flatNodes[row]
			if node.IsDir {
				node.Toggle()
				m.refreshFlatNodes()
			}
		} else {
			m.cursor = row
			m.ensureVisible()
		}
	}
	return m, nil
}

func (m Model) updateSearchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.searching = false
		m.filtered = false
		m.flatSearch = false
		m.searchQuery = ""
		m.searchNodes = nil
		m.searchMatchIndices = nil
		m.searchPathIndices = nil
		m.restoreExpandedState()

	case "enter":
		m.searching = false
		m.filtered = true
		if m.searchNodes != nil {
			m.flatNodes = m.searchNodes
		}
		m.clampCursor()

	case "backspace":
		if len(m.searchQuery) > 0 {
			_, size := utf8.DecodeLastRuneInString(m.searchQuery)
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-size]
			m.applySearchFilter()
		}

	case "ctrl+c":
		return m, tea.Quit

	case "j", "down":
		m.moveCursor(1)

	case "k", "up":
		m.moveCursor(-1)

	case "right":
		m.jumpToMatch(1)

	case "left":
		m.jumpToMatch(-1)

	default:
		if msg.Type == tea.KeyRunes {
			m.searchQuery += msg.String()
			m.applySearchFilter()
		}
	}

	return m, nil
}

func (m Model) updateNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		if m.filtered {
			m.filtered = false
			m.flatSearch = false
			m.searchQuery = ""
			m.searchNodes = nil
			m.searchMatchIndices = nil
			m.searchPathIndices = nil
			m.restoreExpandedState()
		}

	case "q", "ctrl+c":
		return m, tea.Quit

	case "j", "down":
		m.moveCursor(1)

	case "k", "up":
		m.moveCursor(-1)

	case "g":
		m.cursor = 0
		m.scrollOff = 0

	case "G":
		m.cursor = len(m.flatNodes) - 1
		m.ensureVisible()

	case "ctrl+d":
		m.moveCursor(m.viewportHeight() / 2)

	case "ctrl+u":
		m.moveCursor(-m.viewportHeight() / 2)

	case "l", "right":
		if m.filtered {
			m.jumpToMatch(1)
		} else {
			node := m.flatNodes[m.cursor]
			if node.IsDir && !node.Expanded {
				node.Expand()
				m.refreshFlatNodes()
			}
		}

	case "h", "left":
		if m.filtered {
			m.jumpToMatch(-1)
		} else {
			node := m.flatNodes[m.cursor]
			if node.IsDir && node.Expanded {
				node.Collapse()
				m.refreshFlatNodes()
			} else if node.Parent != nil && node.Parent != m.root {
				for i, n := range m.flatNodes {
					if n == node.Parent {
						m.cursor = i
						m.ensureVisible()
						break
					}
				}
			}
		}

	case "enter", " ":
		node := m.flatNodes[m.cursor]
		if node.IsDir {
			node.Toggle()
			m.refreshFlatNodes()
		}

	case "c":
		node := m.flatNodes[m.cursor]
		relPath := strings.TrimPrefix(node.Path, "./")
		if err := clipboard.WriteAll(relPath); err != nil {
			return m, flash(&m, fmt.Sprintf("✗ Failed to copy: %s", err))
		}
		return m, flash(&m, fmt.Sprintf("✓ Copied: %s", relPath))

	case "E":
		m.setExpandAll(m.root, true)
		m.refreshFlatNodes()

	case "W":
		m.setExpandAll(m.root, false)
		m.root.Expanded = true
		m.refreshFlatNodes()
		m.cursor = 0
		m.scrollOff = 0

	case ".":
		m.showHidden = !m.showHidden
		tree.ShowHidden = m.showHidden
		if root, err := tree.BuildTree(m.rootPath); err == nil {
			m.root = root
			m.refreshFlatNodes()
			m.ensureVisible()
		}

	case "/":
		m.startSearch(false)

	case "ctrl+_", "ctrl+f":
		m.startSearch(true)

	case "?":
		m.showHelp = !m.showHelp
	}

	return m, nil
}

// startSearch enters search mode. flat=true for flat file search, false for hierarchy search.
func (m *Model) startSearch(flat bool) {
	m.saveExpandedState()
	m.searching = true
	m.filtered = false
	m.flatSearch = flat
	m.searchQuery = ""
	m.searchNodes = nil
	m.searchMatchIndices = nil
	m.searchPathIndices = nil
	if !flat {
		m.setExpandAll(m.root, true)
	}
	m.refreshFlatNodes()
	m.cursor = 0
	m.scrollOff = 0
}

// refreshTree rebuilds the file tree from disk, preserving expanded state and cursor.
func (m *Model) refreshTree() {
	// Capture expanded paths and cursor path
	expandedPaths := make(map[string]bool)
	for _, n := range tree.FlattenAll(m.root) {
		if n.IsDir && n.Expanded {
			expandedPaths[n.Path] = true
		}
	}
	var cursorPath string
	if m.cursor >= 0 && m.cursor < len(m.flatNodes) {
		cursorPath = m.flatNodes[m.cursor].Path
	}

	// Rebuild
	root, err := tree.BuildTree(m.rootPath)
	if err != nil {
		return
	}
	m.root = root

	// Restore expanded state
	for _, n := range tree.FlattenAll(m.root) {
		if n.IsDir {
			if expandedPaths[n.Path] {
				n.Expand()
			} else {
				n.Collapse()
			}
		}
	}
	m.root.Expanded = true

	if m.filtered && m.searchNodes != nil {
		m.applySearchFilter()
	} else {
		m.refreshFlatNodes()
	}

	// Restore cursor position
	if cursorPath != "" {
		for i, n := range m.flatNodes {
			if n.Path == cursorPath {
				m.cursor = i
				break
			}
		}
	}
	m.clampCursor()
	m.ensureVisible()
}

// --- Search ---

// nodeSource implements fuzzy.Source for tree nodes, matching against relative path
type nodeSource []*tree.Node

func (ns nodeSource) String(i int) string {
	return strings.TrimPrefix(ns[i].Path, "./")
}
func (ns nodeSource) Len() int { return len(ns) }

// splitMatchIndices splits full-path match indices into name indices and dir-path indices.
// path is "dir/name", nameOffset is len(path) - len(name).
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

func (m *Model) updateSearch() {
	if m.searchQuery == "" {
		m.searchNodes = nil
		m.searchMatchIndices = nil
		m.searchPathIndices = nil
		m.searchPathIndices = nil
		return
	}

	allNodes := tree.FlattenAll(m.root)
	results := fuzzy.FindFrom(m.searchQuery, nodeSource(allNodes))

	nameMap := make(map[*tree.Node][]int)
	matchSet := make(map[*tree.Node]bool)
	for _, r := range results {
		node := allNodes[r.Index]
		path := strings.TrimPrefix(node.Path, "./")
		nameIdx, _ := splitMatchIndices(r.MatchedIndexes, path, node.Name)
		nameMap[node] = nameIdx
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

// updateFlatSearch does a flat fuzzy search — no hierarchy, files first then dirs.
func (m *Model) updateFlatSearch() {
	if m.searchQuery == "" {
		m.searchNodes = nil
		m.searchMatchIndices = nil
		m.searchPathIndices = nil
		m.searchPathIndices = nil
		return
	}

	allNodes := tree.FlattenAll(m.root)
	results := fuzzy.FindFrom(m.searchQuery, nodeSource(allNodes))

	nameMap := make(map[*tree.Node][]int)
	pathMap := make(map[*tree.Node][]int)
	var files, dirs []*tree.Node
	for _, r := range results {
		node := allNodes[r.Index]
		path := strings.TrimPrefix(node.Path, "./")
		nameIdx, pathIdx := splitMatchIndices(r.MatchedIndexes, path, node.Name)
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

// --- Expand/Collapse state save/restore ---

func (m *Model) saveExpandedState() {
	m.savedExpanded = make(map[*tree.Node]bool)
	for _, n := range tree.FlattenAll(m.root) {
		if n.IsDir {
			m.savedExpanded[n] = n.Expanded
		}
	}
	m.savedCursor = m.cursor
	m.savedScrollOff = m.scrollOff
}

func (m *Model) restoreExpandedState() {
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

// --- Viewport ---

func (m *Model) viewportHeight() int {
	h := m.height - 1 // status bar
	if m.searching {
		h-- // search input
	}
	if h < 1 {
		h = 1
	}
	return h
}

func (m *Model) ensureVisible() {
	viewH := m.viewportHeight()
	if m.cursor < m.scrollOff {
		m.scrollOff = m.cursor
	}
	if m.cursor >= m.scrollOff+viewH {
		m.scrollOff = m.cursor - viewH + 1
	}
}

// --- View ---

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
		prefix = m.getDisplayPrefix(node, node.Depth-1)
	}
	icon := icons.GetIcon(node.Name, node.IsDir, node.Expanded)
	matchIndices := m.searchMatchIndices[node]

	// In flat search mode, append the relative parent dir after the name
	var dirPath string
	if m.flatSearch && node.Parent != nil && node.Parent != m.root {
		dirPath = strings.TrimPrefix(node.Parent.Path, "./")
	}

	if selected {
		treeLineSelectedStyle := treeLineStyle.Background(lipgloss.Color(colorSelection))
		var parts []string
		parts = append(parts, selectedStyle.Render(" "))
		if prefix != "" {
			parts = append(parts, treeLineSelectedStyle.Render(prefix))
		}
		parts = append(parts, selectedStyle.Render(icon+" "))
		parts = append(parts, m.renderNameHighlighted(node.Name, matchIndices, selectedStyle, matchHighlightSelectedStyle))
		if dirPath != "" {
			parts = append(parts, flatPathSelectedStyle.Render("  "))
			parts = append(parts, m.renderNameHighlighted(dirPath, m.searchPathIndices[node], flatPathSelectedStyle, matchHighlightSelectedStyle))
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
	parts = append(parts, m.renderNameHighlighted(node.Name, matchIndices, nameStyle, matchHighlightStyle))
	if dirPath != "" {
		parts = append(parts, flatPathStyle.Render("  "))
		parts = append(parts, m.renderNameHighlighted(dirPath, m.searchPathIndices[node], flatPathStyle, matchHighlightStyle))
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
				color := lipgloss.Color(colorBlue)
				return lipgloss.NewStyle().Foreground(color), lipgloss.NewStyle().Foreground(color)
			case gitAdded, gitUntracked:
				color := lipgloss.Color(colorGreen)
				return lipgloss.NewStyle().Foreground(color), lipgloss.NewStyle().Foreground(color)
			case gitDeleted:
				color := lipgloss.Color(colorRed)
				return lipgloss.NewStyle().Foreground(color), lipgloss.NewStyle().Foreground(color)
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
	for ancestor != nil && ancestor.Parent != nil {
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
		Foreground(lipgloss.Color(colorPurple)).
		Bold(true).
		Width(16).
		Align(lipgloss.Left).
		PaddingLeft(2)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorFgDim))

	for _, bind := range bindings {
		b.WriteString(keyStyle.Render(bind.key))
		b.WriteString(descStyle.Render(bind.desc))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(colorComment)).PaddingLeft(2).Render("Press ? to return"))

	return b.String()
}

// --- Git ---

func gitRefreshTick() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return gitRefreshMsg{}
	})
}

func fetchGitInfo(path string) tea.Cmd {
	return func() tea.Msg {
		return gitInfoMsg{
			branch:     getGitBranch(path),
			fileStatus: getGitFileStatus(path),
		}
	}
}

func getGitBranch(path string) string {
	// Try rev-parse first (works when there are commits)
	for _, args := range [][]string{
		{"rev-parse", "--abbrev-ref", "HEAD"},
		{"symbolic-ref", "--short", "HEAD"}, // fallback for repos with no commits
	} {
		cmd := exec.Command("git", append([]string{"-C", path}, args...)...)
		if out, err := cmd.Output(); err == nil {
			if branch := strings.TrimSpace(string(out)); branch != "" {
				return branch
			}
		}
	}
	return ""
}



func getGitFileStatus(path string) map[string]gitFileStatus {
	out, err := exec.Command("git", "-C", path, "status", "--porcelain").Output()
	if err != nil {
		return nil
	}
	result := make(map[string]gitFileStatus)
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if len(line) < 4 {
			continue
		}
		x, y := line[0], line[1]
		file := strings.TrimSpace(line[3:])
		// Handle renames: "R  old -> new"
		if idx := strings.Index(file, " -> "); idx >= 0 {
			file = file[idx+4:]
		}

		var status gitFileStatus
		switch {
		case x == '?' || y == '?':
			status = gitUntracked
		case x == 'D' || y == 'D':
			status = gitDeleted
		case x == 'A' || x == '?':
			status = gitAdded
		case x == 'M' || y == 'M' || x == 'R':
			status = gitModified
		default:
			status = gitModified
		}
		result[file] = status

		// Propagate to parent directories
		dir := file
		for {
			dir = parentDir(dir)
			if dir == "" {
				break
			}
			// Dirs get the highest-priority status of their children
			if existing, ok := result[dir]; !ok || status > existing {
				result[dir] = status
			}
		}
	}
	return result
}

func parentDir(path string) string {
	if i := strings.LastIndexByte(path, '/'); i > 0 {
		return path[:i]
	}
	return ""
}

// --- Utility ---

func flattenSkipRoot(root *tree.Node) []*tree.Node {
	all := tree.Flatten(root)
	if len(all) > 1 {
		return all[1:]
	}
	return all
}
