//go:build !js

package ui

import (
	"fmt"
	"os"
	"os/exec"
	"time"

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
		m.refreshTree()
		reEnableMouse := func() tea.Msg { return tea.EnableMouseAllMotion() }
		if msg.err != nil {
			return m, tea.Batch(reEnableMouse, flash(&m, fmt.Sprintf("✗ Editor error: %s", msg.err)))
		}
		return m, reEnableMouse

	case tea.KeyMsg:
		isRune := msg.Type == tea.KeyRunes
		result := m.HandleKey(msg.String(), isRune)
		return m.applyKeyResult(result)
	}

	return m, nil
}

// applyKeyResult converts a KeyResult into Bubble Tea commands.
func (m Model) applyKeyResult(r KeyResult) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if r.Quit {
		cmds = append(cmds, tea.Quit)
	}

	if r.CopyPath != "" {
		if err := clipboard.WriteAll(r.CopyPath); err != nil {
			cmds = append(cmds, flash(&m, fmt.Sprintf("✗ Failed to copy: %s", err)))
			return m, tea.Batch(cmds...)
		}
	}

	if r.FlashMsg != "" {
		cmds = append(cmds, flash(&m, r.FlashMsg))
	}

	if r.OpenEditor != "" {
		editor := os.Getenv("EDITOR")
		if editor == "" {
			cmds = append(cmds, flash(&m, "✗ $EDITOR is not set"))
		} else {
			c := exec.Command(editor, r.OpenEditor)
			cmds = append(cmds, tea.ExecProcess(c, func(err error) tea.Msg {
				return editorFinishedMsg{err}
			}))
		}
	}

	if len(cmds) == 0 {
		return m, nil
	}
	return m, tea.Batch(cmds...)
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
				r := m.handleNormalKey("enter")
				return m.applyKeyResult(r)
			}
		}
	}
	return m, nil
}
