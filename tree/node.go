package tree

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Node represents a file or directory in the tree
type Node struct {
	Name     string
	Path     string // relative path from root
	AbsPath  string
	IsDir    bool
	Children []*Node
	Parent   *Node
	Expanded bool
	Depth    int
	Loaded   bool // whether children have been loaded
}

// Hidden files/dirs to skip
var defaultHidden = map[string]bool{
	".git":          true,
	"node_modules":  true,
	".DS_Store":     true,
	"__pycache__":   true,
	".next":         true,
	".cache":        true,
	".vscode":       true,
	".idea":         true,
}

// ShowHidden controls whether hidden/ignored files are displayed
var ShowHidden = false

// BuildTree creates a tree from a root path (only loads top level initially)
func BuildTree(rootPath string) (*Node, error) {
	absPath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}

	root := &Node{
		Name:     info.Name(),
		Path:     ".",
		AbsPath:  absPath,
		IsDir:    info.IsDir(),
		Expanded: true,
		Depth:    0,
		Loaded:   false,
	}

	if root.IsDir {
		err = loadChildren(root)
		if err != nil {
			return nil, err
		}
	}

	return root, nil
}

// loadChildren loads the immediate children of a directory node
func loadChildren(node *Node) error {
	entries, err := os.ReadDir(node.AbsPath)
	if err != nil {
		return err
	}

	node.Children = nil
	var dirs []*Node
	var files []*Node

	for _, entry := range entries {
		name := entry.Name()

		// Skip hidden defaults unless ShowHidden is on
		if !ShowHidden && defaultHidden[name] {
			continue
		}

		childPath := filepath.Join(node.Path, name)
		childAbsPath := filepath.Join(node.AbsPath, name)

		child := &Node{
			Name:    name,
			Path:    childPath,
			AbsPath: childAbsPath,
			IsDir:   entry.IsDir(),
			Parent:  node,
			Depth:   node.Depth + 1,
			Loaded:  false,
		}

		if entry.IsDir() {
			dirs = append(dirs, child)
		} else {
			files = append(files, child)
		}
	}

	// Sort: dirs first (alphabetical), then files (alphabetical) – case insensitive
	sort.Slice(dirs, func(i, j int) bool {
		return strings.ToLower(dirs[i].Name) < strings.ToLower(dirs[j].Name)
	})
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	node.Children = append(dirs, files...)
	node.Loaded = true
	return nil
}

// Toggle expands or collapses a directory node
func (n *Node) Toggle() error {
	if !n.IsDir {
		return nil
	}
	if !n.Loaded {
		if err := loadChildren(n); err != nil {
			return err
		}
	}
	n.Expanded = !n.Expanded
	return nil
}

// Expand expands a directory node
func (n *Node) Expand() error {
	if !n.IsDir {
		return nil
	}
	if !n.Loaded {
		if err := loadChildren(n); err != nil {
			return err
		}
	}
	n.Expanded = true
	return nil
}

// Collapse collapses a directory node
func (n *Node) Collapse() {
	if !n.IsDir {
		return
	}
	n.Expanded = false
}

// Flatten returns a flat list of visible nodes for rendering
func Flatten(root *Node) []*Node {
	var result []*Node
	flatten(root, &result)
	return result
}

func flatten(node *Node, result *[]*Node) {
	*result = append(*result, node)
	if node.IsDir && node.Expanded {
		for _, child := range node.Children {
			flatten(child, result)
		}
	}
}

// FlattenAll returns all nodes in the tree regardless of expanded state (skips root)
func FlattenAll(root *Node) []*Node {
	var result []*Node
	flattenAll(root, &result)
	if len(result) > 0 {
		return result[1:] // skip root
	}
	return result
}

func flattenAll(node *Node, result *[]*Node) {
	*result = append(*result, node)
	if node.IsDir {
		// Load children if not loaded
		if !node.Loaded {
			loadChildren(node)
		}
		for _, child := range node.Children {
			flattenAll(child, result)
		}
	}
}

// IsLastChild returns whether this node is the last child of its parent
func (n *Node) IsLastChild() bool {
	if n.Parent == nil {
		return true
	}
	children := n.Parent.Children
	return len(children) > 0 && children[len(children)-1] == n
}

// GetTreePrefix returns the tree drawing characters for this node
func (n *Node) GetTreePrefix() string {
	if n.Depth == 0 {
		return ""
	}

	// Build prefix from bottom up
	var parts []string

	// Current level connector
	if n.IsLastChild() {
		parts = append(parts, "└─")
	} else {
		parts = append(parts, "├─")
	}

	// Walk up to build indentation
	parent := n.Parent
	current := n
	for parent != nil && parent.Depth > 0 {
		if current.Parent != nil && !current.Parent.IsLastChild() {
			parts = append(parts, "│ ")
		} else {
			parts = append(parts, "  ")
		}
		current = parent
		parent = parent.Parent
	}

	// Reverse
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}

	return strings.Join(parts, "")
}
