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
func (m *Model) SetGitInfo(branch string, files map[string]GitFileStatus) {
	m.gitBranch = branch
	m.gitFiles = files
}

// SetFlash sets a flash message (caller is responsible for clearing it later).
func (m *Model) SetFlash(msg string) {
	m.flashMsg = msg
}

// GitFileStatus is the exported type for git file status constants.
type GitFileStatus = gitFileStatus

// Exported git status constants for demo use.
const (
	GitModified  = gitModified
	GitAdded     = gitAdded
	GitDeleted   = gitDeleted
	GitUntracked = gitUntracked
	GitIgnored   = gitIgnored
)
