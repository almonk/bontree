package ui

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/alasdairmonk/bontree/config"
	"github.com/alasdairmonk/bontree/tree"
	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

// flash sets a temporary flash message that auto-clears.
func flash(m *Model, msg string) tea.Cmd {
	m.flashMsg = msg
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return clearFlashMsg{}
	})
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
	key := msg.String()
	action := m.cfg.ActionFor(key)

	// Search mode has its own action set plus hardcoded essentials
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
		return m, tea.Quit

	case action == config.ActionMoveDown:
		m.moveCursor(1)

	case action == config.ActionMoveUp:
		m.moveCursor(-1)

	case action == config.ActionSearchNextMatch:
		m.jumpToMatch(1)

	case action == config.ActionSearchPrevMatch:
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
	action := m.cfg.ActionFor(msg.String())

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
		return m, tea.Quit

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
		if err := clipboard.WriteAll(relPath); err != nil {
			return m, flash(&m, fmt.Sprintf("✗ Failed to copy: %s", err))
		}
		return m, flash(&m, fmt.Sprintf("✓ Copied path: %s", relPath))

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
	}

	return m, nil
}
