# pathpin

**pin your paths.** A bookmark-first directory navigator for the terminal: pin the folders you
use, then jump to them from a menu or by label. Written in Go with
[Bubble Tea](https://github.com/charmbracelet/bubbletea). Works in bash, zsh and fish on Linux
and macOS.

![pathpin: open the menu, switch group tabs, filter to a bookmark and go there, jump by label, and preview themes live](https://raw.githubusercontent.com/turtlethom/pathpin/main/demo/demo.gif)

```bash
pathpin              # open the menu: pick a folder, press enter, you're there
pathpin notes        # or jump straight to the bookmark labeled NOTES
```

- **Bookmarks you act on in place:** add, remove, rename and reorder them from the menu.
- **Groups as tabs:** `Config  School  Work`, switched with `tab`, filtered with `/`.
- **Actions on any bookmark:** open it in your editor, copy its path, or open it in your file
  manager, without leaving the menu.
- **Your look:** seven themes or your terminal's own colors, every color adjustable, three
  selection styles, four header styles and five icon sets, all previewed live with `t`.
- **One plain TOML file** that works across machines and in a dotfiles repo.

## Install

Download the latest release into `~/.local/bin` (make sure that folder is on your `PATH`), with
the line for your system:

```bash
mkdir -p ~/.local/bin

# Linux, x86-64
curl -fsSL https://github.com/turtlethom/pathpin/releases/latest/download/pathpin_linux_amd64.tar.gz | tar -xzf - -C ~/.local/bin pathpin
# Linux, ARM
curl -fsSL https://github.com/turtlethom/pathpin/releases/latest/download/pathpin_linux_arm64.tar.gz | tar -xzf - -C ~/.local/bin pathpin
# macOS, Apple Silicon
curl -fsSL https://github.com/turtlethom/pathpin/releases/latest/download/pathpin_darwin_arm64.tar.gz | tar -xzf - -C ~/.local/bin pathpin
# macOS, Intel
curl -fsSL https://github.com/turtlethom/pathpin/releases/latest/download/pathpin_darwin_amd64.tar.gz | tar -xzf - -C ~/.local/bin pathpin
```

Each [release](https://github.com/turtlethom/pathpin/releases) also lists the archives and a
`checksums.txt` to verify them. If you download one in a browser on macOS, run
`xattr -d com.apple.quarantine pathpin` once after unpacking: pathpin isn't signed by Apple, so
macOS refuses to open a browser download otherwise (files fetched with `curl` aren't affected).

Or build it from source (needs Go 1.25.4 or newer):

```bash
git clone https://github.com/turtlethom/pathpin.git
cd pathpin && go build -o ~/.local/bin/pathpin .
```

### Set up your shell

A program can't change its shell's directory by itself, so pathpin comes with a small shell
function that does the `cd` for it. Add one line to your shell's config, then open a new terminal:

| Shell | File | Line |
|---|---|---|
| bash | `~/.bashrc` (macOS: `~/.bash_profile`) | `eval "$(pathpin init bash)"` |
| zsh | `~/.zshrc` (or `$ZDOTDIR/.zshrc`) | `eval "$(pathpin init zsh)"` |
| fish | `~/.config/fish/config.fish` | `pathpin init fish \| source` |

This also sets up tab completion: `pathpin no<TAB>` completes `notes`. Without the line,
pathpin prints the folder instead of going there, and says how to set it up.

### Pick your own command name

The command is `pathpin` by default. To type something shorter, name it with `--cmd`:

```bash
eval "$(pathpin init bash --cmd pp)"
```

Now `pp` opens the menu and `pp notes` jumps. Any name made of letters, digits, `_` and `-`
works, as long as it starts with a letter or `_`. The function hides any other command with the same name in your interactive shell, so
pick one you don't already use (check with `type pp`). The examples below use `pathpin`.

### Set your editor

`e` in the menu opens a folder in `$VISUAL`, else `$EDITOR`, else `vi`. If you haven't set one,
add it to your shell's config:

```bash
export EDITOR=nvim       # bash, zsh
set -gx EDITOR nvim      # fish
```

### Coming from fcd?

The first time you run pathpin, it imports your fcd bookmarks and leaves fcd's file in place.

## Use

```bash
pathpin              # open the menu
pathpin notes        # jump to the bookmark labeled NOTES (any case)
pathpin day job      # labels with spaces need no quotes
pathpin add          # bookmark this folder, labeled with its name in uppercase
pathpin add work     # ...or with your own label
pathpin add -g Work  # ...in a group (--group; tab completes group names)
pathpin ls           # list bookmarks by group
pathpin --version
```

`pathpin ls` prints a table in a terminal. Piped, each line is
`LABEL<tab>/absolute/path<tab>GROUP` (the group is empty if there's none), so `cut -f1,2` gives
label and path:

```bash
pathpin ls | while IFS=$'\t' read -r label dir group; do echo "$label → $dir"; done
```

### The menu

Labels show in uppercase; paths show with `~` for your home folder. A bookmark whose folder no
longer exists is dimmed and marked `(missing)`. Press `?` for the full list of keys:

**Navigate**

| Key | Action |
|---|---|
| `↑` `↓` / `j` `k` | move |
| `←` `→` / `h` `l` | previous / next page |
| `g` / `G` | top / end |
| `tab` / `shift+tab` | next / previous group tab |
| `.` | show only the group tabs (hide All and Other) until you quit |

**Bookmarks**

| Key | Action |
|---|---|
| `a` | add the folder you opened pathpin from (a field suggests a label; `enter` saves) |
| `r` | remove the highlighted bookmark (asks first) |
| `c` | clear all bookmarks (asks first) |
| `J` / `K` (or `shift+↓` / `shift+↑`) | move the highlighted bookmark down / up |

**Names**

| Key | Action |
|---|---|
| `n` | rename the highlighted bookmark |
| `N` | rename the open group's tab (onto another group's name, it merges them, after asking) |

**Groups**

| Key | Action |
|---|---|
| `m` | move the highlighted bookmark to a group, picked from a list, or `+ new group…` |
| `u` | take the highlighted bookmark out of its group (asks first) |

**Actions**

| Key | Action |
|---|---|
| `enter` | go to the highlighted folder |
| `e` | open it in your editor; the menu comes back when the editor closes |
| `y` | copy its full path to the clipboard |
| `o` | open it in your file manager |

**Menu**

| Key | Action |
|---|---|
| `/` | filter (`esc` clears it) |
| `t` | appearance: theme, colors, styles, icons (see [Appearance](#appearance)) |
| `?` | help |
| `q` / `esc` | quit without changing directory |

`y` uses `pbcopy` on macOS, and on Linux `wl-copy` (Wayland), `xclip` or `xsel`, whichever is
installed (`wl-clipboard` or `xclip` from your package manager). `o` uses `open` on macOS and
`xdg-open` on Linux.

### Groups

Press `m` on a bookmark and pick `+ new group…` to make your first group. The menu then shows a
tab for each group, plus All and, for bookmarks without a group, Other:

```
   PATHPIN

   All   Config   School   Other
  ───────────────────────────────
```

- `tab` and `shift+tab` switch tabs, and `/` filters the tab you're on. On All it matches group
  names too, so `school` finds everything in School, and each row shows its group (`· School`).
- `a` on a group's tab adds the new bookmark to that group.
- `N` renames a group. Renaming it to another group's name merges the two, and an empty name
  ungroups all its bookmarks (they move to Other, nothing is deleted); both ask first.
- `.` hides All and Other, so only your groups show. To start that way, set `tabs = "groups"` (see
  [Appearance](#appearance)).
- Group names match in any case. Don't name a group `All` or `Other`: those are the tab names.

## Appearance

Press `t` in the menu. A panel opens under the list; the menu itself is the preview, redrawn as
you change things:

| Keys | |
|---|---|
| `↑` `↓` / `j` `k` | pick a row |
| `←` `→` / `h` `l` | change it |
| `#` | type any color (on a color row) |
| `d` | put the row back to its default |
| `enter` / `esc` | save / leave everything as it was |

Everything the panel sets lives in the config file too:

```toml
[colors]
theme = "nord"
primary = "#56B6C2"        # accent: the selected row, title and filter
secondary = "#ECEFF4"      # labels
tertiary = "#7B88A1"       # paths and other muted text
selected_text = "#2E3440"  # text on the accent
error = "#BF616A"          # error messages

[menu]
selection = "bar"          # the selected row: block (default), bar or arrow
header = "rule"            # the title and open tab: badge (default), text, bracket or rule
icons = "triangle"         # off (default), nerd, dot, triangle, diamond or star
tabs = "groups"            # how the menu starts: all (default), or groups (All and Other hidden)
```

### Themes and colors

| Theme | |
|---|---|
| `terminal` (default) | your terminal's own colors, so it follows your terminal's theme |
| `catppuccin-latte` | light |
| `catppuccin-mocha`, `gruvbox-dark`, `monokai-pro-ristretto`, `nord`, `one-dark`, `tokyo-night` | dark |

Each color takes a hex color (`"#56B6C2"`) or an ANSI number from 0 to 255 (`"6"`; 0–15 follow
your terminal's palette) and overrides that one color of the theme. In the panel, colors cycle through your
terminal's 16 and a set of named ones (`aqua · #56B6C2`); the file stores the value.

The terminal's 16 are named after their slots (`bright-blue · 12`), but your terminal's theme
decides how they look: Monokai Pro, for one, makes its blues orange. The panel says
`your terminal's palette` next to them; for an exact color, pick a named one or type a hex with `#`.

If the selected row's label is hard to read on the `terminal` theme (it's your palette's color 0
on color 6), set `selected_text`, or try one of the presets.

### Styles

| `selection` | The highlighted row |
|---|---|
| `block` | the label as a solid block in the accent color |
| `bar` | a bar at the left, the label bold in the accent |
| `arrow` | a `›` at the left, the label bold in the accent |

| `header` | The title and the open tab |
|---|---|
| `badge` | solid blocks in the accent color |
| `text` | bold, in the accent |
| `bracket` | `[PATHPIN]`, `[All]` |
| `rule` | `PATHPIN ────` across the window, the open tab underlined |

### Icons

| `icons` | Folder | Home | Needs |
|---|---|---|---|
| `dot` | `●` | `⌂` | any font |
| `triangle` | `▸` | `⌂` | any font |
| `diamond` | `◆` | `⌂` | any font |
| `star` | `★` | `⌂` | any font |
| `nerd` | folder icon | house icon | a [Nerd Font](https://www.nerdfonts.com/) |

A folder that no longer exists gets `✗` in every set. `●`, `◆` and `★` are "ambiguous width"
characters: if your terminal draws those two columns wide (an option in some terminals, and common
with CJK locales), the rows shift. Use `triangle` or `nerd` there.

A value pathpin doesn't know doesn't break the menu: an unknown theme or style falls back to the
default, a value that isn't a color is ignored, and pathpin names the setting to fix when the menu
closes. A mistyped key (`selecton`) or a broken file stops pathpin with an error naming it, and
the file is left as it is.

## Config

Bookmarks live in `~/.config/pathpin/config.toml` (or `$XDG_CONFIG_HOME/pathpin/config.toml`):

```toml
[[bookmark]]
label = "NVIM"
path = "~/.config/nvim"
group = "Config"
```

- Paths inside your home folder are stored as `~/...`, so the file works across machines and in
  a dotfiles repo (symlinks are followed).
- Before every change, pathpin copies the previous file to `config.toml.bak`. It keeps **one**
  step back, so the next change replaces it.
- pathpin rewrites the file on every change (from the menu or `pathpin add`), so **comments
  aren't kept**. Use groups to organize.
- A file pathpin can't read is reported and left untouched, never reset.

## Known limitations

- **Slow start in some terminals.** Before it starts, pathpin asks the terminal for its background
  color. A terminal that never answers (such as the Linux text console, or a session recorded with
  `script`) makes every pathpin command wait about 5 seconds. Most terminals answer at once, and
  inside tmux or screen nothing is asked. This comes from Bubble Tea v1 and should go away with v2.
- **No copying over SSH.** `y` uses the clipboard of the machine pathpin runs on.
- **Editor paths with spaces.** `$EDITOR` is split on spaces (so `code -w` works), which means an
  editor whose path contains a space won't start.
- **One backup.** `config.toml.bak` holds only the file as it was before the last change.
- **Unresponsive network folders.** Going to, opening or adding a folder whose network mount
  stopped answering (`enter`, `e`, `o`, `a`, `pathpin <label>`) waits until the mount does.
- **Windows** isn't supported yet.

## How it works

pathpin is a Go binary plus a small shell function that does the `cd` for it: the function runs
pathpin with a temporary file, pathpin writes the chosen folder there, and the function `cd`s to
it. [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) maps the code: the packages, how a key press
becomes a `cd`, and where to make changes.

## Uninstall

Remove the `pathpin init` line from your shell's config, delete the `pathpin` binary, and, to
drop your bookmarks too, `~/.config/pathpin/` (or `$XDG_CONFIG_HOME/pathpin/`).

## License

MIT
