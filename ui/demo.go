package ui

import (
	"github.com/almonk/bontree/config"
	"github.com/almonk/bontree/tree"
)

// NewDemo creates a Model from a pre-built tree root for demo/WASM use.
// No filesystem or git access is performed.
func NewDemo(root *tree.Node, cfg *config.Config) Model {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	m := Model{
		root:      root,
		flatNodes: flattenTree(root),
		cfg:       cfg,
	}
	return m
}

// SetGitInfo sets the git branch and file status for demo display.
func (m *Model) SetGitInfo(branch string, files map[string]gitFileStatus) {
	m.gitBranch = branch
	m.gitFiles = files
}

// Exported git status constants for demo use.
const (
	GitModified  = gitModified
	GitAdded     = gitAdded
	GitDeleted   = gitDeleted
	GitUntracked = gitUntracked
	GitIgnored   = gitIgnored
)
