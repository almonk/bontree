package ui

import (
	"time"

	"github.com/alasdairmonk/bontree/config"
	"github.com/alasdairmonk/bontree/tree"
	tea "github.com/charmbracelet/bubbletea"
)

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
	cfg        *config.Config

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

// New creates a new Model with the given config. If cfg is nil, defaults are used.
func New(rootPath string, cfg *config.Config) (Model, error) {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	tree.ShowHidden = cfg.ShowHidden
	tree.RefreshGitIgnored(rootPath)
	root, err := tree.BuildTree(rootPath)
	if err != nil {
		return Model{}, err
	}

	return Model{
		root:       root,
		flatNodes:  flattenTree(root),
		rootPath:   rootPath,
		showHidden: cfg.ShowHidden,
		cfg:        cfg,
	}, nil
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(fetchGitInfo(m.rootPath), gitRefreshTick())
}

// --- Helpers ---

// refreshFlatNodes rebuilds the flat node list and clamps the cursor.
func (m *Model) refreshFlatNodes() {
	m.flatNodes = flattenTree(m.root)
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

// --- Utility ---

func flattenTree(root *tree.Node) []*tree.Node {
	return tree.Flatten(root)
}
