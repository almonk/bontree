package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/almonk/bontree/config"
	"github.com/almonk/bontree/tree"
	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

// editorFinishedMsg is sent when the external editor process exits.
type editorFinishedMsg struct{ err error }

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

	case editorFinishedMsg:
		// Editor exited — refresh the tree in case files changed
		m.refreshTree()
		if msg.err != nil {
			return m, flash(&m, fmt.Sprintf("✗ Editor error: %s", msg.err))
		}
		return m, nil

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

		m.cursor = row
		m.ensureVisible()

		if doubleClick {
			node := m.flatNodes[row]
			if node.IsDir {
				node.Toggle()
				m.refreshFlatNodes()
			} else if action := m.cfg.ActionFor("enter"); action != "" {
				return m.dispatchAction(action)
			}
		}
	}
	return m, nil
}

func (m Model) updateSearchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// In search mode, printable characters are always typed into the query —
	// never dispatched as normal-mode actions (e.g. j/k/g/G/q).
	if msg.Type == tea.KeyRunes {
		m.searchQuery += msg.String()
		m.applySearchFilter()
		return m, nil
	}

	key := msg.String()
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
		return m, tea.Quit

	case action == config.ActionMoveDown || key == "down":
		m.moveCursor(1)

	case action == config.ActionMoveUp || key == "up":
		m.moveCursor(-1)

	case action == config.ActionSearchNextMatch || action == config.ActionExpand || key == "right":
		m.jumpToMatch(1)

	case action == config.ActionSearchPrevMatch || action == config.ActionCollapse || key == "left":
		m.jumpToMatch(-1)
	}

	return m, nil
}

// dispatchAction executes a single action. Used by updateNormalMode and mouse handlers.
func (m Model) dispatchAction(action config.Action) (tea.Model, tea.Cmd) {
	switch action {
	case config.ActionOpenEditor:
		node := m.flatNodes[m.cursor]
		if node.IsDir {
			node.Toggle()
			m.refreshFlatNodes()
			return m, nil
		}
		editor := os.Getenv("EDITOR")
		if editor == "" {
			return m, flash(&m, "✗ $EDITOR is not set")
		}
		c := exec.Command(editor, node.AbsPath)
		return m, tea.ExecProcess(c, func(err error) tea.Msg {
			return editorFinishedMsg{err}
		})
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

	case config.ActionOpenEditor:
		return m.dispatchAction(action)
	}

	return m, nil
}
