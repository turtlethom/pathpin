# Appearance

pathpin starts in your terminal's own colors. Everything below can be changed live from the menu
with `t`, or in the config file.

- [The `t` panel](#the-t-panel)
- [The config file](#the-config-file)
- [Themes and colors](#themes-and-colors)
- [Styles](#styles)
- [Icons](#icons)
- [Values pathpin doesn't know](#values-pathpin-doesnt-know)

## The `t` panel

Press `t` in the menu. A panel opens under the list; the menu itself is the preview, redrawn as
you change things:

| Keys | |
|---|---|
| `↑` `↓` / `j` `k` | pick a row |
| `←` `→` / `h` `l` | change it |
| `#` | type any color (on a color row) |
| `d` | put the row back to its default |
| `enter` / `esc` | save / leave everything as it was |

## The config file

Everything the panel sets lives in the config file (`~/.config/pathpin/config.toml`, see
[Using pathpin](usage.md#the-config-file)) too:

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

## Themes and colors

| Theme | |
|---|---|
| `terminal` (default) | your terminal's own colors, so it follows your terminal's theme |
| `catppuccin-latte` | light |
| `catppuccin-mocha`, `gruvbox-dark`, `monokai-pro-ristretto`, `nord`, `one-dark`, `tokyo-night` | dark |

Each color takes a hex color (`"#56B6C2"`) or an ANSI number from 0 to 255 (`"6"`; 0–15 follow
your terminal's palette) and overrides that one color of the theme. In the panel, colors cycle
through your terminal's 16 and a set of named ones (`aqua · #56B6C2`); the file stores the value.

The terminal's 16 are named after their slots (`bright-blue · 12`), but your terminal's theme
decides how they look: Monokai Pro, for one, makes its blues orange. The panel says
`your terminal's palette` next to them; for an exact color, pick a named one or type a hex with `#`.

If the selected row's label is hard to read on the `terminal` theme (it's your palette's color 0
on color 6), set `selected_text`, or try one of the presets.

## Styles

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

## Icons

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

## Values pathpin doesn't know

A value pathpin doesn't know doesn't break the menu: an unknown theme or style falls back to the
default, a value that isn't a color is ignored, and pathpin names the setting to fix when the menu
closes. A mistyped key (`selecton`) or a broken file stops pathpin with an error naming it, and
the file is left as it is.
