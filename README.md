# pathpin

**pin your paths.** A bookmark-first directory navigator for the terminal: pin the folders you
use, then jump to them from a menu or by label. Works in bash, zsh and fish on Linux and macOS.
Written in Go with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

![pathpin: open the menu, switch group tabs, filter to a bookmark and go there, jump by label, and preview themes live](https://raw.githubusercontent.com/turtlethom/pathpin/main/demo/demo.gif)

- **Bookmarks you act on in place:** add, remove, rename and reorder them from the menu.
- **Groups as tabs:** `Config  School  Work`, switched with `tab`, filtered with `/`.
- **Actions on any bookmark:** open it in your editor, copy its path, or open it in your file
  manager, without leaving the menu.
- **Your look:** seven themes or your terminal's own colors, previewed live with `t`.
- **One plain TOML file** that works across machines and in a dotfiles repo.

## Install

```bash
curl -fsSL https://github.com/turtlethom/pathpin/releases/latest/download/install.sh | sh
```

The script picks the build for your system, checks it against the release's checksums, and puts
`pathpin` in `~/.local/bin`. Run it again to update. To install with Go, download by hand, or
build from source, see [Installing pathpin](https://github.com/turtlethom/pathpin/blob/main/docs/install.md).

## Set up your shell

A program can't change its shell's directory by itself, so pathpin comes with a small shell
function that does the `cd` for it. Open your shell's section and run its commands **once**
(running them again adds the lines again). Not sure which shell you use? Run `echo $SHELL`.

The first command puts `~/.local/bin` on your `PATH`; leave it out if `command -v pathpin`
already prints a path.

<details>
<summary><b>bash</b> on Linux</summary>

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
echo 'eval "$(pathpin init bash)"' >> ~/.bashrc
source ~/.bashrc
```

</details>

<details>
<summary><b>bash</b> on macOS</summary>

macOS's Terminal starts bash as a login shell, which reads `~/.bash_profile`, not `~/.bashrc`.
Tab completion needs the `bash-completion` package, which macOS doesn't include: install it
with Homebrew or MacPorts.

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bash_profile
echo 'eval "$(pathpin init bash)"' >> ~/.bash_profile
source ~/.bash_profile
```

</details>

<details>
<summary><b>zsh</b> on Linux and macOS (macOS's default shell)</summary>

```zsh
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
echo 'eval "$(pathpin init zsh)"' >> ~/.zshrc
source ~/.zshrc
```

If you set `ZDOTDIR`, use `$ZDOTDIR/.zshrc` instead of `~/.zshrc`.

</details>

<details>
<summary><b>fish</b> on Linux and macOS</summary>

```fish
fish_add_path ~/.local/bin
echo 'pathpin init fish | source' >> ~/.config/fish/config.fish
source ~/.config/fish/config.fish
```

`fish_add_path` saves the `PATH` change itself, so it needs no line in `config.fish`. If you set
`XDG_CONFIG_HOME`, use `$XDG_CONFIG_HOME/fish/config.fish` instead of `~/.config/fish/config.fish`.

</details>

This also sets up tab completion (in bash, through the `bash-completion` package). To type
something shorter than `pathpin` (say `pp`), or to set the editor that `e` opens, see your
shell's section in
[Installing pathpin](https://github.com/turtlethom/pathpin/blob/main/docs/install.md#set-up-your-shell).

## First use

```bash
cd ~/projects        # any folder you use often
pathpin add          # bookmark it, labeled PROJECTS
cd ~
pathpin projects     # jump back to it from anywhere
pathpin              # or open the menu: pick a bookmark, press enter, you're there
```

In the menu, `a` adds the folder you're in, `/` filters, `t` changes the look, and `?` shows
every key. Coming from fcd? pathpin imports your fcd bookmarks the first time it runs.

## Learn more

- [Installing pathpin](https://github.com/turtlethom/pathpin/blob/main/docs/install.md): every
  way to install, updating, checksums, uninstalling
- [Using pathpin](https://github.com/turtlethom/pathpin/blob/main/docs/usage.md): the menu's
  keys, groups, the command line, the config file, known limitations
- [Appearance](https://github.com/turtlethom/pathpin/blob/main/docs/appearance.md): themes,
  colors, styles and icons
- [Architecture](https://github.com/turtlethom/pathpin/blob/main/docs/ARCHITECTURE.md): how the
  code fits together, for contributors

## License

MIT
