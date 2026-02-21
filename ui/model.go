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

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7aa2f7")).
			PaddingLeft(1)

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#283457")).
			Foreground(lipgloss.Color("#c0caf5")).
			Bold(true)

	dirStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7aa2f7")).
			Bold(true)

	fileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a9b1d6"))

	treeLineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3b4261"))

	iconDirStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7aa2f7"))

	iconFileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7dcfff"))

	flashStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9ece6a")).
			Bold(true)

	rootIconStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e0af68")).
			Bold(true)

	// Status bar styles
	statusBarBgStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#1a1b26")).
				Foreground(lipgloss.Color("#565f89"))

	statusPathStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1a1b26")).
			Foreground(lipgloss.Color("#a9b1d6")).
			PaddingLeft(1).
			PaddingRight(1)

	statusBranchStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#1a1b26")).
				Foreground(lipgloss.Color("#bb9af7")).
				Bold(true)

	statusCountStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#1a1b26")).
				Foreground(lipgloss.Color("#565f89")).
				PaddingRight(1)

	statusFlashStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#1a1b26")).
				Foreground(lipgloss.Color("#9ece6a")).
				Bold(true).
				PaddingLeft(1)

	statusHelpStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1a1b26")).
			Foreground(lipgloss.Color("#3b4261"))

	statusModifiedStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#1a1b26")).
				Foreground(lipgloss.Color("#e0af68")).
				Bold(true)

	statusStagedStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#1a1b26")).
				Foreground(lipgloss.Color("#9ece6a")).
				Bold(true)

	statusUntrackedStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#1a1b26")).
				Foreground(lipgloss.Color("#f7768e")).
				Bold(true)

	searchInputStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#1a1b26")).
				Foreground(lipgloss.Color("#c0caf5")).
				PaddingLeft(1)

	searchPromptStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#1a1b26")).
				Foreground(lipgloss.Color("#7aa2f7")).
				Bold(true).
				PaddingLeft(1)

	matchHighlightStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ff9e64")).
				Bold(true)

	matchHighlightSelectedStyle = lipgloss.NewStyle().
					Background(lipgloss.Color("#283457")).
					Foreground(lipgloss.Color("#ff9e64")).
					Bold(true)
)

type clearFlashMsg struct{}
type gitRefreshMsg struct{}

func gitRefreshTick() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return gitRefreshMsg{}
	})
}

type gitStatus struct {
	staged    int
	modified  int
	untracked int
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
	scrollOff  int // viewport scroll offset
	gitBranch  string
	gitStatus  gitStatus
	showHidden bool

	// Search
	searching    bool
	searchQuery  string
	searchNodes  []*tree.Node // filtered flat list preserving hierarchy
	searchMatchIndices map[*tree.Node][]int // char indices that matched for highlighting
}

// New creates a new Model
func New(rootPath string) (Model, error) {
	root, err := tree.BuildTree(rootPath)
	if err != nil {
		return Model{}, err
	}

	flat := flattenSkipRoot(root)

	m := Model{
		root:      root,
		flatNodes: flat,
		cursor:    0,
		rootPath:  rootPath,
	}

	return m, nil
}

type gitInfoMsg struct {
	branch string
	status gitStatus
}

func fetchGitInfo(path string) tea.Cmd {
	return func() tea.Msg {
		return gitInfoMsg{
			branch: getGitBranch(path),
			status: getGitStatus(path),
		}
	}
}

// flattenSkipRoot returns the flattened tree without the root node itself
func flattenSkipRoot(root *tree.Node) []*tree.Node {
	all := tree.Flatten(root)
	if len(all) > 1 {
		return all[1:] // skip the root
	}
	return all
}

func getGitBranch(path string) string {
	// Try rev-parse first (works when there are commits)
	cmd := exec.Command("git", "-C", path, "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err == nil {
		branch := strings.TrimSpace(string(out))
		if branch != "" {
			return branch
		}
	}
	// Fall back to symbolic-ref for repos with no commits yet
	cmd = exec.Command("git", "-C", path, "symbolic-ref", "--short", "HEAD")
	out, err = cmd.Output()
	if err == nil {
		branch := strings.TrimSpace(string(out))
		if branch != "" {
			return branch
		}
	}
	return ""
}

func getGitStatus(path string) gitStatus {
	var gs gitStatus
	cmd := exec.Command("git", "-C", path, "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return gs
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		if len(line) < 2 {
			continue
		}
		x := line[0] // index/staged status
		y := line[1] // working tree status
		if x == '?' {
			gs.untracked++
		} else {
			if x != ' ' && x != '?' {
				gs.staged++
			}
			if y != ' ' && y != '?' {
				gs.modified++
			}
		}
	}
	return gs
}

// nodeSource implements fuzzy.Source for tree nodes
type nodeSource []*tree.Node

func (ns nodeSource) String(i int) string { return ns[i].Name }
func (ns nodeSource) Len() int            { return len(ns) }

// updateSearch filters the full tree using fuzzy matching, preserving hierarchy.
func (m *Model) updateSearch() {
	if m.searchQuery == "" {
		m.searchNodes = nil
		m.searchMatchIndices = nil
		return
	}

	// Get all nodes from tree
	allNodes := tree.FlattenAll(m.root)

	// Run fuzzy match
	results := fuzzy.FindFrom(m.searchQuery, nodeSource(allNodes))

	// Build match map: node -> matched char indices
	matchMap := make(map[*tree.Node][]int)
	matchSet := make(map[*tree.Node]bool)
	for _, r := range results {
		node := allNodes[r.Index]
		matchMap[node] = r.MatchedIndexes
		matchSet[node] = true
		// Mark all ancestors as included to preserve hierarchy
		ancestor := node.Parent
		for ancestor != nil {
			if matchSet[ancestor] {
				break
			}
			matchSet[ancestor] = true
			ancestor = ancestor.Parent
		}
	}

	// Build filtered list preserving original order and hierarchy
	var filtered []*tree.Node
	for _, node := range allNodes {
		if matchSet[node] && node != m.root {
			filtered = append(filtered, node)
		}
	}

	m.searchNodes = filtered
	m.searchMatchIndices = matchMap
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(fetchGitInfo(m.rootPath), gitRefreshTick())
}

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
		m.gitStatus = msg.status
		return m, gitRefreshTick()

	case gitRefreshMsg:
		return m, fetchGitInfo(m.rootPath)

	case tea.KeyMsg:
		// Search mode input handling
		if m.searching {
			switch msg.String() {
			case "esc":
				m.searching = false
				m.searchQuery = ""
				m.searchNodes = nil
				m.searchMatchIndices = nil
				m.flatNodes = flattenSkipRoot(m.root)
				if m.cursor >= len(m.flatNodes) {
					m.cursor = len(m.flatNodes) - 1
				}
				if m.cursor < 0 {
					m.cursor = 0
				}
				m.scrollOff = 0
				return m, nil

			case "enter":
				m.searching = false
				// Keep filtered results as current view
				if m.searchNodes != nil {
					m.flatNodes = m.searchNodes
				}
				if m.cursor >= len(m.flatNodes) {
					m.cursor = len(m.flatNodes) - 1
				}
				if m.cursor < 0 {
					m.cursor = 0
				}
				return m, nil

			case "backspace":
				if len(m.searchQuery) > 0 {
					_, size := utf8.DecodeLastRuneInString(m.searchQuery)
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-size]
					m.updateSearch()
					if m.searchNodes != nil {
						m.flatNodes = m.searchNodes
					} else {
						m.flatNodes = flattenSkipRoot(m.root)
					}
					m.cursor = 0
					m.scrollOff = 0
				}
				return m, nil

			case "ctrl+c":
				return m, tea.Quit

			case "j", "down":
				if m.cursor < len(m.flatNodes)-1 {
					m.cursor++
					m.ensureVisible()
				}
				return m, nil

			case "k", "up":
				if m.cursor > 0 {
					m.cursor--
					m.ensureVisible()
				}
				return m, nil

			default:
				// Only add printable characters
				if msg.Type == tea.KeyRunes {
					m.searchQuery += msg.String()
					m.updateSearch()
					if m.searchNodes != nil {
						m.flatNodes = m.searchNodes
					} else {
						m.flatNodes = flattenSkipRoot(m.root)
					}
					m.cursor = 0
					m.scrollOff = 0
				}
				return m, nil
			}
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "j", "down":
			if m.cursor < len(m.flatNodes)-1 {
				m.cursor++
				m.ensureVisible()
			}

		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
				m.ensureVisible()
			}

		case "g":
			m.cursor = 0
			m.scrollOff = 0

		case "G":
			m.cursor = len(m.flatNodes) - 1
			m.ensureVisible()

		case "ctrl+d":
			viewH := m.viewportHeight()
			jump := viewH / 2
			m.cursor += jump
			if m.cursor >= len(m.flatNodes) {
				m.cursor = len(m.flatNodes) - 1
			}
			m.ensureVisible()

		case "ctrl+u":
			viewH := m.viewportHeight()
			jump := viewH / 2
			m.cursor -= jump
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.ensureVisible()

		case "l", "right":
			node := m.flatNodes[m.cursor]
			if node.IsDir && !node.Expanded {
				node.Expand()
				m.flatNodes = flattenSkipRoot(m.root)
			}

		case "enter":
			node := m.flatNodes[m.cursor]
			if node.IsDir {
				node.Toggle()
				m.flatNodes = flattenSkipRoot(m.root)
				if m.cursor >= len(m.flatNodes) {
					m.cursor = len(m.flatNodes) - 1
				}
			}

		case "h", "left":
			node := m.flatNodes[m.cursor]
			if node.IsDir && node.Expanded {
				node.Collapse()
				m.flatNodes = flattenSkipRoot(m.root)
			} else if node.Parent != nil && node.Parent != m.root {
				// Jump to parent (but not the hidden root)
				for i, n := range m.flatNodes {
					if n == node.Parent {
						m.cursor = i
						m.ensureVisible()
						break
					}
				}
			}

		case " ":
			node := m.flatNodes[m.cursor]
			if node.IsDir {
				node.Toggle()
				m.flatNodes = flattenSkipRoot(m.root)
				if m.cursor >= len(m.flatNodes) {
					m.cursor = len(m.flatNodes) - 1
				}
			}

		case "c":
			node := m.flatNodes[m.cursor]
			relPath := node.Path
			// Strip the leading "./" from paths
			relPath = strings.TrimPrefix(relPath, "./")
			err := clipboard.WriteAll(relPath)
			if err != nil {
				m.flashMsg = fmt.Sprintf("✗ Failed to copy: %s", err)
			} else {
				m.flashMsg = fmt.Sprintf("✓ Copied: %s", relPath)
			}
			return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
				return clearFlashMsg{}
			})

		case "E":
			m.expandAll(m.root)
			m.flatNodes = flattenSkipRoot(m.root)

		case "W":
			m.collapseAll(m.root)
			m.root.Expanded = true
			m.flatNodes = flattenSkipRoot(m.root)
			m.cursor = 0
			m.scrollOff = 0

		case ".":
			m.showHidden = !m.showHidden
			tree.ShowHidden = m.showHidden
			root, err := tree.BuildTree(m.rootPath)
			if err == nil {
				m.root = root
				m.flatNodes = flattenSkipRoot(m.root)
				if m.cursor >= len(m.flatNodes) {
					m.cursor = len(m.flatNodes) - 1
				}
				if m.cursor < 0 {
					m.cursor = 0
				}
				m.ensureVisible()
			}

		case "/":
			m.searching = true
			m.searchQuery = ""
			m.searchNodes = nil
			m.searchMatchIndices = nil
			// Expand all for search
			m.expandAll(m.root)
			m.cursor = 0
			m.scrollOff = 0

		case "?":
			m.showHelp = !m.showHelp
		}
	}

	return m, nil
}

func (m *Model) expandAll(node *tree.Node) {
	if node.IsDir {
		node.Expand()
		for _, child := range node.Children {
			m.expandAll(child)
		}
	}
}

func (m *Model) collapseAll(node *tree.Node) {
	if node.IsDir {
		node.Collapse()
		for _, child := range node.Children {
			m.collapseAll(child)
		}
	}
}

func (m *Model) viewportHeight() int {
	// height minus: title(1) + status bar(1) + search bar(1 if searching)
	h := m.height - 2
	if m.searching {
		h-- // search input takes a line
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

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	if m.showHelp {
		return m.helpView()
	}

	var b strings.Builder

	// Title
	titleIcon := rootIconStyle.Render("\uf07c")
	title := titleStyle.Render(fmt.Sprintf("%s %s", titleIcon, m.root.Name))
	b.WriteString(title)
	b.WriteString("\n")

	// Tree content
	viewH := m.viewportHeight()
	end := m.scrollOff + viewH
	if end > len(m.flatNodes) {
		end = len(m.flatNodes)
	}

	contentWidth := m.width
	if contentWidth < 20 {
		contentWidth = 20
	}

	for i := m.scrollOff; i < end; i++ {
		node := m.flatNodes[i]
		line := m.renderNode(node, i == m.cursor, contentWidth)
		b.WriteString(line)
		if i < end-1 {
			b.WriteString("\n")
		}
	}

	// Pad remaining lines
	rendered := end - m.scrollOff
	for i := rendered; i < viewH; i++ {
		if i > 0 || rendered == 0 {
			b.WriteString("\n")
		}
	}

	// Search input (above status bar)
	if m.searching {
		b.WriteString("\n")
		prompt := searchPromptStyle.Render("/")
		input := searchInputStyle.Render(m.searchQuery + "█")
		searchLine := prompt + input
		// Pad to full width
		sw := lipgloss.Width(searchLine)
		if sw < m.width {
			searchLine += lipgloss.NewStyle().Background(lipgloss.Color("#1a1b26")).Render(strings.Repeat(" ", m.width-sw))
		}
		b.WriteString(searchLine)
	}

	// Status bar (full width)
	b.WriteString("\n")
	b.WriteString(m.renderStatusBar())

	return b.String()
}

func (m Model) renderStatusBar() string {
	w := m.width
	if w < 1 {
		w = 80
	}

	var left string
	if m.flashMsg != "" {
		left = statusFlashStyle.Render(" " + m.flashMsg)
	} else if m.gitBranch != "" {
		branchIcon := "\ue725" //  git branch icon
		branchStr := fmt.Sprintf(" %s %s", branchIcon, m.gitBranch)

		// Append git status counts
		var statusParts []string
		if m.gitStatus.staged > 0 {
			statusParts = append(statusParts, statusStagedStyle.Render(fmt.Sprintf("+%d", m.gitStatus.staged)))
		}
		if m.gitStatus.modified > 0 {
			statusParts = append(statusParts, statusModifiedStyle.Render(fmt.Sprintf("~%d", m.gitStatus.modified)))
		}
		if m.gitStatus.untracked > 0 {
			statusParts = append(statusParts, statusUntrackedStyle.Render(fmt.Sprintf("?%d", m.gitStatus.untracked)))
		}

		if len(statusParts) > 0 {
			branchStr += " " + strings.Join(statusParts, " ")
		}

		left = statusBranchStyle.Render(branchStr)
	}

	// Right side: count + help
	var rightParts []string

	countStr := fmt.Sprintf("%d/%d ", m.cursor+1, len(m.flatNodes))
	rightParts = append(rightParts, statusCountStyle.Render(countStr))

	// Help keys
	helpKeys := statusHelpStyle.Render(" ?:help  c:copy  q:quit ")
	rightParts = append(rightParts, helpKeys)

	right := strings.Join(rightParts, statusBranchStyle.Render("│"))

	// Calculate padding
	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	padding := w - leftWidth - rightWidth
	if padding < 0 {
		padding = 0
	}

	padStr := statusBarBgStyle.Render(strings.Repeat(" ", padding))

	return left + padStr + right
}

func (m Model) renderNode(node *tree.Node, selected bool, maxWidth int) string {
	// Since we skip root, adjust depth for display (depth 1 -> no indent for top-level)
	displayDepth := node.Depth - 1

	var parts []string

	// Tree prefix (indentation lines) — recalculated for display
	prefix := m.getDisplayPrefix(node, displayDepth)
	if prefix != "" {
		parts = append(parts, treeLineStyle.Render(prefix))
	}

	// Icon
	icon := icons.GetIcon(node.Name, node.IsDir, node.Expanded)
	if node.IsDir {
		parts = append(parts, iconDirStyle.Render(icon)+" ")
	} else {
		parts = append(parts, iconFileStyle.Render(icon)+" ")
	}

	// Name — with optional fuzzy match highlighting
	name := node.Name
	matchIndices := m.searchMatchIndices[node]

	if selected {
		// Re-render with selection highlight
		var selParts []string

		if prefix != "" {
			selParts = append(selParts, treeLineStyle.Render(prefix))
		}

		selParts = append(selParts, selectedStyle.Render(icon+" "))
		selParts = append(selParts, m.renderNameHighlighted(name, matchIndices, selectedStyle, matchHighlightSelectedStyle))

		// Pad the selected line to fill width
		plainLen := lipgloss.Width(strings.Join(selParts, ""))
		if plainLen < maxWidth {
			padding := strings.Repeat(" ", maxWidth-plainLen)
			selParts = append(selParts, selectedStyle.Render(padding))
		}

		return strings.Join(selParts, "")
	}

	nameStyle := fileStyle
	if node.IsDir {
		nameStyle = dirStyle
	}
	parts = append(parts, m.renderNameHighlighted(name, matchIndices, nameStyle, matchHighlightStyle))

	return strings.Join(parts, "")
}

// renderNameHighlighted renders a name with fuzzy match indices highlighted
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
		s := string(ch)
		if matchSet[i] {
			result.WriteString(highlightStyle.Render(s))
		} else {
			result.WriteString(baseStyle.Render(s))
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

	// Current level connector
	if node.IsLastChild() {
		parts = append(parts, "└─")
	} else {
		parts = append(parts, "├─")
	}

	// Walk up ancestors (skip the hidden root)
	current := node
	ancestor := node.Parent
	for ancestor != nil && ancestor.Parent != nil { // stop before root
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

func (m Model) helpView() string {
	var b strings.Builder

	title := titleStyle.Render("  Keybindings")
	b.WriteString(title)
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
		{"/", "Fuzzy search"},
		{".", "Toggle hidden files"},
		{"?", "Toggle help"},
		{"q / Ctrl+c", "Quit"},
	}

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#bb9af7")).
		Bold(true).
		Width(16).
		Align(lipgloss.Left).
		PaddingLeft(2)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#a9b1d6"))

	for _, bind := range bindings {
		b.WriteString(keyStyle.Render(bind.key))
		b.WriteString(descStyle.Render(bind.desc))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#565f89")).PaddingLeft(2).Render("Press ? to return"))

	return b.String()
}
