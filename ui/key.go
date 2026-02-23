package ui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/almonk/bontree/config"
	"github.com/almonk/bontree/tree"
)

// KeyResult holds the outcome of HandleKey for the caller to act on.
type KeyResult struct {
	Quit       bool
	FlashMsg   string // non-empty = set flash message
	CopyPath   string // non-empty = copy this path to clipboard
	OpenEditor string // non-empty = open file at this path in $EDITOR
}

// HandleKey processes a key event given as a string name (e.g. "j", "esc", "ctrl+f").
// isRune should be true when the key is a printable character (not a control/special key).
// This is the single source of truth for keyboard input handling, used by both
// the Bubble Tea TUI and the WASM bridge.
func (m *Model) HandleKey(key string, isRune bool) KeyResult {
	if m.showHelp {
		action := m.cfg.ActionFor(key)
		if action == config.ActionHelp || action == config.ActionQuit || key == "esc" {
			m.showHelp = false
		}
		if action == config.ActionQuit {
			return KeyResult{Quit: true}
		}
		return KeyResult{}
	}

	if m.searching {
		return m.handleSearchKey(key, isRune)
	}
	return m.handleNormalKey(key)
}

func (m *Model) handleSearchKey(key string, isRune bool) KeyResult {
	// In search mode, printable characters are always typed into the query —
	// never dispatched as normal-mode actions (e.g. j/k/h/l/g/G/q).
	if isRune {
		m.searchQuery += key
		m.applySearchFilter()
		return KeyResult{}
	}

	action := m.cfg.ActionFor(key)

	switch {
	case key == "esc" || action == config.ActionSearchCancel:
		if m.searchQuery == "" {
			// Empty search — go straight back to normal mode
			m.searching = false
			m.filtered = false
			m.flatSearch = false
			m.searchNodes = nil
			m.searchMatchIndices = nil
			m.searchPathIndices = nil
			m.restoreExpandedState()
		} else {
			// First escape confirms the search (same as enter)
			m.searching = false
			m.filtered = true
			if m.searchNodes != nil {
				m.flatNodes = m.searchNodes
			}
			m.clampCursor()
		}

	case key == "enter" || action == config.ActionSearchConfirm:
		m.searching = false
		m.filtered = true
		if m.searchNodes != nil {
			m.flatNodes = m.searchNodes
		}
		m.clampCursor()

	case key == "backspace" || action == config.ActionSearchBackspace:
		if len(m.searchQuery) > 0 {
			_, size := utf8.DecodeLastRuneInString(m.searchQuery)
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-size]
			m.applySearchFilter()
		}

	case action == config.ActionQuit:
		return KeyResult{Quit: true}

	case action == config.ActionMoveDown || key == "down":
		m.moveCursor(1)

	case action == config.ActionMoveUp || key == "up":
		m.moveCursor(-1)

	case action == config.ActionSearchNextMatch || action == config.ActionExpand || key == "right":
		m.jumpToMatch(1)

	case action == config.ActionSearchPrevMatch || action == config.ActionCollapse || key == "left":
		m.jumpToMatch(-1)
	}

	return KeyResult{}
}

func (m *Model) handleNormalKey(key string) KeyResult {
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

	case config.ActionQuit:
		return KeyResult{Quit: true}

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
				node.Expand()
				m.refreshFlatNodes()
			}
		}

	case config.ActionCollapse:
		if m.filtered {
			m.jumpToMatch(-1)
		} else {
			node := m.flatNodes[m.cursor]
			if node.IsDir && node.Expanded {
				node.Collapse()
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
			node.Toggle()
			m.refreshFlatNodes()
		}

	case config.ActionCopyPath:
		node := m.flatNodes[m.cursor]
		relPath := strings.TrimPrefix(node.Path, "./")
		return KeyResult{CopyPath: relPath, FlashMsg: fmt.Sprintf("✓ Copied path: %s", relPath)}

	case config.ActionExpandAll:
		m.setExpandAll(m.root, true)
		m.refreshFlatNodes()

	case config.ActionCollapseAll:
		m.setExpandAll(m.root, false)
		m.root.Expanded = true
		m.refreshFlatNodes()
		m.cursor = 0
		m.scrollOff = 0

	case config.ActionToggleHidden:
		m.showHidden = !m.showHidden
		tree.ShowHidden = m.showHidden
		if root, err := tree.BuildTree(m.rootPath); err == nil {
			m.root = root
			m.refreshFlatNodes()
			m.ensureVisible()
		}

	case config.ActionSearch:
		m.startSearch(false)

	case config.ActionFlatSearch:
		m.startSearch(true)

	case config.ActionHelp:
		m.showHelp = !m.showHelp

	case config.ActionOpenEditor:
		node := m.flatNodes[m.cursor]
		if node.IsDir {
			node.Toggle()
			m.refreshFlatNodes()
		} else {
			return KeyResult{OpenEditor: node.AbsPath}
		}
	}

	return KeyResult{}
}
