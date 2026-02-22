package ui

import (
	"strings"

	"github.com/almonk/bontree/tree"
	"github.com/sahilm/fuzzy"
)

// nodeSource implements fuzzy.Source for tree nodes, matching against relative path
type nodeSource []*tree.Node

func (ns nodeSource) String(i int) string {
	return strings.TrimPrefix(ns[i].Path, "./")
}
func (ns nodeSource) Len() int { return len(ns) }

// nodeNameSource implements fuzzy.Source for tree nodes, matching against name only
type nodeNameSource []*tree.Node

func (ns nodeNameSource) String(i int) string { return ns[i].Name }
func (ns nodeNameSource) Len() int            { return len(ns) }

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
		m.flatNodes = flattenTree(m.root)
	}
	m.cursor = 0
	m.scrollOff = 0
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

func (m *Model) updateSearch() {
	if m.searchQuery == "" {
		m.searchNodes = nil
		m.searchMatchIndices = nil
		m.searchPathIndices = nil
		m.searchPathIndices = nil
		return
	}

	allNodes := tree.FlattenAll(m.root)
	// Tree mode: match against node names only for stricter results
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

		// Prefer direct fuzzy match against name/path for better highlights;
		// fall back to splitting full-path match indices.
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
