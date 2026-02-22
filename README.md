# bontree

A fast, interactive terminal file explorer with fuzzy search, git integration, and Nerd Font icons. Built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Features

- **Tree navigation** — expand, collapse, and browse directories with keyboard or mouse
- **Fuzzy search** — hierarchy-aware search (`/`) that auto-expands matching ancestors, or flat file search (`Ctrl+f`) across all files
- **Git status** — files colored by status (modified, added, deleted, untracked, ignored) with branch display in the status bar
- **Nerd Font icons** — language and filetype-specific icons for 50+ file types
- **Clipboard** — copy relative file paths with `c`
- **Hidden files** — toggle visibility with `.`
- **Mouse support** — scroll, click to select, double-click to toggle directories

## Install

```bash
go install github.com/alasdairmonk/bontree@latest
```

Or build from source:

```bash
make build    # produces ./bontree
make install  # installs to /usr/local/bin/bontree
```

## Usage

```bash
bontree [path]
```

Defaults to the current directory if no path is given.

## Keybindings

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `g` | Go to top |
| `G` | Go to bottom |
| `Ctrl+d` | Half page down |
| `Ctrl+u` | Half page up |
| `l` / `→` | Expand directory |
| `h` / `←` | Collapse / go to parent |
| `Enter` / `Space` | Toggle directory |
| `/` | Fuzzy search (tree) |
| `Ctrl+f` | Flat file search |
| `c` | Copy relative path |
| `E` | Expand all |
| `W` | Collapse all |
| `.` | Toggle hidden files |
| `?` | Help |
| `q` / `Ctrl+c` | Quit |

**In search mode:** type to filter, `↑`/`↓` to navigate results, `←`/`→` to jump between matches, `Enter` to confirm, `Esc` to cancel.

## Requirements

- A terminal with [Nerd Font](https://www.nerdfonts.com/) for icons
- Go 1.24+ to build from source
