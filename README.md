# bontree

A file explorer to pair with your favourite agent.

A fast, interactive terminal file explorer with fuzzy search, git integration, and Nerd Font icons. Built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Features

- **Tree navigation** — expand, collapse, and browse directories with keyboard or mouse
- **Fuzzy search** — hierarchy-aware search (`/`) that auto-expands matching ancestors, or flat file search (`Ctrl+f`) across all files
- **Git status** — files colored by status (modified, added, deleted, untracked, ignored) with branch display in the status bar
- **Nerd Font icons** — language and filetype-specific icons for 50+ file types
- **Theming** — use any Ghostty-compatible theme, or inherit your terminal's colors
- **Configurable keybindings** — remap every key or strip down to a minimal layout
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

Every keybinding listed below is a default — all of them can be remapped or removed in the [config file](#configuration).

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

## Configuration

Bontree uses a Ghostty-style config file at `~/.config/bontree/config` (respects `$XDG_CONFIG_HOME`).

The syntax is `key = value`. Comments start with `#` and must be on their own line. Blank lines are ignored. A fully commented example is available in [`config.example`](config.example).

### Custom keybindings

Keybindings are configured with `keybind = <key>=<action>`. Your bindings are merged with the defaults — you only need to specify what you want to change.

```
# ~/.config/bontree/config

# Remap a key
keybind = ctrl+q=quit

# Remove a binding you don't want
keybind = q=unbind
```

#### Key names

Letters and symbols are used as-is (`a`, `G`, `/`, `?`, `.`). Special keys: `up`, `down`, `left`, `right`, `enter`, `esc`, `backspace`, `tab`, `space`. Modifiers use `+`: `ctrl+c`, `ctrl+d`, `ctrl+f`, `ctrl+u`.

#### Available actions

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

### Settings

| Key | Values | Default | Description |
|-----|--------|---------|-------------|
| `show-hidden` | `true` / `false` | `false` | Show hidden files (dotfiles) on startup |
| `theme` | theme name | *(unset)* | Ghostty-compatible color theme (see [Theming](#theming)) |

### Theming

Bontree can load Ghostty-compatible theme files for consistent colors across tools. By default it inherits your terminal's colors; set `theme` in your config to override:

```
# ~/.config/bontree/config
theme = Catppuccin Mocha
```

Themes are searched in order:

1. `~/.config/bontree/themes/<name>` — your own custom themes
2. `/Applications/Ghostty.app/Contents/Resources/ghostty/themes/<name>` — bundled Ghostty themes (macOS)
3. `~/.config/ghostty/themes/<name>` — Ghostty user themes

If you already use Ghostty, your themes are picked up automatically. To add a custom theme, drop a file in `~/.config/bontree/themes/` using the standard format:

```
palette = 0=#1a1b26
palette = 1=#f7768e
# ... indices 0-15
background = #1a1b26
foreground = #c0caf5
selection-background = #33467c
selection-foreground = #c0caf5
cursor-color = #c0caf5
```

## Requirements

- A terminal with [Nerd Font](https://www.nerdfonts.com/) for icons
- Go 1.24+ to build from source
