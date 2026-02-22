package ui

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

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
	switch msg.String() {
	case "esc":
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

	case "down":
		m.moveCursor(1)

	case "up":
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
		return m, flash(&m, fmt.Sprintf("✓ Copied path: %s", relPath))

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
