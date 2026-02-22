# bontree

A file explorer for your agents.

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

All keybindings can be overridden via the [config file](#configuration).

## Configuration

Bontree uses a Ghostty-style config file at `~/.config/bontree/config` (respects `$XDG_CONFIG_HOME`).

The syntax is `key = value`. Comments start with `#` and must be on their own line. Blank lines are ignored.

```
# ~/.config/bontree/config

# Show hidden files by default
show-hidden = true

# Keybindings use the format: keybind = <key>=<action>
keybind = q=quit
keybind = ctrl+c=quit
keybind = j=move_down
keybind = down=move_down
keybind = k=move_up
keybind = up=move_up
```

> **Note:** When you add _any_ `keybind` line, all default keybindings are cleared and only your specified bindings will be active. This gives you full control — there are no hidden bindings you can't remove.

To remove a single binding, use `unbind`:

```
keybind = q=unbind
```

### Settings

| Key | Values | Default | Description |
|-----|--------|---------|-------------|
| `show-hidden` | `true` / `false` | `false` | Show hidden files (dotfiles) on startup |

### Available actions

| Action | Description |
|--------|-------------|
| `quit` | Exit bontree |
| `move_down` | Move cursor down |
| `move_up` | Move cursor up |
| `go_top` | Jump to first item |
| `go_bottom` | Jump to last item |
| `half_page_down` | Scroll half page down |
| `half_page_up` | Scroll half page up |
| `expand` | Expand directory (next match in filter mode) |
| `collapse` | Collapse directory / go to parent (prev match in filter mode) |
| `toggle` | Toggle directory open/close |
| `copy_path` | Copy relative path to clipboard |
| `expand_all` | Expand all directories |
| `collapse_all` | Collapse all directories |
| `toggle_hidden` | Toggle hidden file visibility |
| `search` | Start fuzzy search (tree mode) |
| `flat_search` | Start flat file search |
| `help` | Toggle help screen |
| `clear_filter` | Clear active search filter |

Search mode also supports: `search_confirm`, `search_cancel`, `search_backspace`, `search_next_match`, `search_prev_match`.

### Key names

Letters and symbols are used as-is (`a`, `G`, `/`, `?`, `.`). Special keys: `up`, `down`, `left`, `right`, `enter`, `esc`, `backspace`, `tab`, ` ` (space). Modifiers use `+`: `ctrl+c`, `ctrl+d`, `ctrl+f`, `ctrl+u`.

A fully commented example config is available in [`config.example`](config.example).

## Requirements

- A terminal with [Nerd Font](https://www.nerdfonts.com/) for icons
- Go 1.24+ to build from source
