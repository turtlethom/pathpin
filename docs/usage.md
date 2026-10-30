# Using pathpin

These examples use the command name `pathpin`; if you picked your own with `--cmd` (see
[Installing pathpin](install.md#pick-your-own-command-name)), type that instead.

- [The command line](#the-command-line)
- [The menu](#the-menu)
- [Groups](#groups)
- [The config file](#the-config-file)
- [Known limitations](#known-limitations)

For the menu's colors and styles, see [Appearance](appearance.md).

## The command line

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

## The menu

Labels show in uppercase; paths show with `~` for your home folder. A bookmark whose folder no
longer exists is dimmed and marked `(missing)`. Press `?` in the menu for this list of keys:

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
| `t` | appearance: theme, colors, styles, icons (see [Appearance](appearance.md)) |
| `?` | help |
| `q` / `esc` | quit without changing directory |

`e` opens `$VISUAL`, else `$EDITOR`, else `vi`; [Installing pathpin](install.md#set-up-your-shell)
shows how to set it in each shell. `y` uses `pbcopy` on macOS, and on Linux `wl-copy` (Wayland),
`xclip` or `xsel`, whichever is installed (`wl-clipboard` or `xclip` from your package manager).
`o` uses `open` on macOS and `xdg-open` on Linux.

## Groups

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
  [Appearance](appearance.md#the-config-file)).
- Group names match in any case. Don't name a group `All` or `Other`: those are the tab names.

## The config file

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

The `[colors]` and `[menu]` settings are in [Appearance](appearance.md#the-config-file).

## Known limitations

- **No copying over SSH.** `y` uses the clipboard of the machine pathpin runs on.
- **Editor paths with spaces.** `$EDITOR` is split on spaces (so `code -w` works), which means an
  editor whose path contains a space won't start.
- **One backup.** `config.toml.bak` holds only the file as it was before the last change.
- **Unresponsive network folders.** Going to, opening or adding a folder whose network mount
  stopped answering (`enter`, `e`, `o`, `a`, `pathpin <label>`) waits until the mount does.
- **Windows** isn't supported yet.
