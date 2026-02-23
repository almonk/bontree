//go:build !js

package ui

import (
	"os/exec"
	"strings"
	"time"

	"github.com/almonk/bontree/tree"
	tea "github.com/charmbracelet/bubbletea"
)

type gitRefreshMsg struct{}
type gitInfoMsg struct {
	branch     string
	fileStatus map[string]gitFileStatus // relative path -> status
}

func gitRefreshTick() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return gitRefreshMsg{}
	})
}

func fetchGitInfo(path string) tea.Cmd {
	return func() tea.Msg {
		tree.RefreshGitIgnored(path)
		return gitInfoMsg{
			branch:     getGitBranch(path),
			fileStatus: getGitFileStatus(path),
		}
	}
}

func getGitBranch(path string) string {
	// Try rev-parse first (works when there are commits)
	for _, args := range [][]string{
		{"rev-parse", "--abbrev-ref", "HEAD"},
		{"symbolic-ref", "--short", "HEAD"}, // fallback for repos with no commits
	} {
		cmd := exec.Command("git", append([]string{"-C", path}, args...)...)
		if out, err := cmd.Output(); err == nil {
			if branch := strings.TrimSpace(string(out)); branch != "" {
				return branch
			}
		}
	}
	return ""
}

func getGitFileStatus(path string) map[string]gitFileStatus {
	out, err := exec.Command("git", "-C", path, "status", "--porcelain", "--ignored").Output()
	if err != nil {
		return nil
	}
	result := make(map[string]gitFileStatus)
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if len(line) < 4 {
			continue
		}
		x, y := line[0], line[1]
		file := strings.TrimSpace(line[3:])
		// Handle renames: "R  old -> new"
		if idx := strings.Index(file, " -> "); idx >= 0 {
			file = file[idx+4:]
		}

		var status gitFileStatus
		switch {
		case x == '!' || y == '!':
			status = gitIgnored
		case x == '?' || y == '?':
			status = gitUntracked
		case x == 'D' || y == 'D':
			status = gitDeleted
		case x == 'A' || x == '?':
			status = gitAdded
		case x == 'M' || y == 'M' || x == 'R':
			status = gitModified
		default:
			status = gitModified
		}
		result[file] = status

		// Propagate to parent directories (skip ignored — don't dim parent dirs)
		if status != gitIgnored {
			dir := file
			for {
				dir = parentDir(dir)
				if dir == "" {
					break
				}
				// Dirs get the highest-priority status of their children
				if existing, ok := result[dir]; !ok || status > existing {
					result[dir] = status
				}
			}
		}
	}
	return result
}

func parentDir(path string) string {
	if i := strings.LastIndexByte(path, '/'); i > 0 {
		return path[:i]
	}
	return ""
}
