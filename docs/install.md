# Installing pathpin

pathpin runs on Linux and macOS, on x86-64 and ARM, in bash, zsh and fish. Installing takes two
steps: get the `pathpin` binary, then [set up your shell](#set-up-your-shell) so pathpin can
change your directory.

- [Install script](#install-script) (recommended)
- [With Go](#with-go)
- [Download a release by hand](#download-a-release-by-hand)
- [Build from source](#build-from-source)
- [Set up your shell](#set-up-your-shell)
- [Updating](#updating)
- [Coming from fcd](#coming-from-fcd)
- [Uninstall](#uninstall)

## Install script

```bash
curl -fsSL https://github.com/turtlethom/pathpin/releases/latest/download/install.sh | sh
```

The script picks the build for your system, checks it against the release's `checksums.txt`, and
puts `pathpin` in `~/.local/bin`. Run the same line again to update. It never edits your shell's
config: on a first install, it prints the commands from [Set up your shell](#set-up-your-shell)
for the shell you use.

Options go after `sh -s --`:

```bash
# install somewhere else than ~/.local/bin
curl -fsSL https://github.com/turtlethom/pathpin/releases/latest/download/install.sh | sh -s -- --bin-dir ~/bin
# install a specific release, for example to go back to an older one
curl -fsSL https://github.com/turtlethom/pathpin/releases/latest/download/install.sh | sh -s -- --version 2.0.1
```

The script never runs `sudo` itself. For a folder only root can write to, like `/usr/local/bin`,
run it as `… | sudo sh -s -- --bin-dir /usr/local/bin`.

To read the script before running it:

```bash
curl -fsSLO https://github.com/turtlethom/pathpin/releases/latest/download/install.sh
less install.sh
sh install.sh
```

## With Go

Needs Go 1.26 or newer:

```bash
go install github.com/turtlethom/pathpin/v2@latest
```

Go puts `pathpin` in `~/go/bin` (or `$(go env GOPATH)/bin`), not `~/.local/bin`. In
[Set up your shell](#set-up-your-shell), put that folder on your `PATH` instead.

## Download a release by hand

Each [release](https://github.com/turtlethom/pathpin/releases) has an archive per system and a
`checksums.txt`. Find your CPU with `uname -m`: `x86_64` is x86-64 (Intel and AMD), `aarch64`
or `arm64` is ARM (Apple Silicon on a Mac). Then run the commands for your system. They download
into a temporary folder, check the archive, and only if it checks out, unpack `pathpin` into
`~/.local/bin`.

### Linux, x86-64

```bash
cd "$(mktemp -d)"
curl -fsSLO https://github.com/turtlethom/pathpin/releases/latest/download/pathpin_linux_amd64.tar.gz
curl -fsSLO https://github.com/turtlethom/pathpin/releases/latest/download/checksums.txt
mkdir -p ~/.local/bin
grep pathpin_linux_amd64.tar.gz checksums.txt | sha256sum -c &&
  tar -xzf pathpin_linux_amd64.tar.gz -C ~/.local/bin pathpin
```

### Linux, ARM

```bash
cd "$(mktemp -d)"
curl -fsSLO https://github.com/turtlethom/pathpin/releases/latest/download/pathpin_linux_arm64.tar.gz
curl -fsSLO https://github.com/turtlethom/pathpin/releases/latest/download/checksums.txt
mkdir -p ~/.local/bin
grep pathpin_linux_arm64.tar.gz checksums.txt | sha256sum -c &&
  tar -xzf pathpin_linux_arm64.tar.gz -C ~/.local/bin pathpin
```

### macOS, Apple Silicon

```bash
cd "$(mktemp -d)"
curl -fsSLO https://github.com/turtlethom/pathpin/releases/latest/download/pathpin_darwin_arm64.tar.gz
curl -fsSLO https://github.com/turtlethom/pathpin/releases/latest/download/checksums.txt
mkdir -p ~/.local/bin
grep pathpin_darwin_arm64.tar.gz checksums.txt | shasum -a 256 -c &&
  tar -xzf pathpin_darwin_arm64.tar.gz -C ~/.local/bin pathpin
```

### macOS, Intel

```bash
cd "$(mktemp -d)"
curl -fsSLO https://github.com/turtlethom/pathpin/releases/latest/download/pathpin_darwin_amd64.tar.gz
curl -fsSLO https://github.com/turtlethom/pathpin/releases/latest/download/checksums.txt
mkdir -p ~/.local/bin
grep pathpin_darwin_amd64.tar.gz checksums.txt | shasum -a 256 -c &&
  tar -xzf pathpin_darwin_amd64.tar.gz -C ~/.local/bin pathpin
```

The check prints `pathpin_…tar.gz: OK`. Anything else means the download is damaged, and nothing
is unpacked: run the commands again.

If you download an archive in a browser on macOS, run `xattr -d com.apple.quarantine pathpin`
once after unpacking: pathpin isn't signed by Apple, so macOS refuses to open a browser download
otherwise. Files fetched with `curl` aren't affected.

## Build from source

Needs Go 1.26 or newer:

```bash
git clone https://github.com/turtlethom/pathpin.git
cd pathpin
go build -o ~/.local/bin/pathpin .
```

## Set up your shell

A program can't change its shell's directory by itself, so pathpin comes with a small shell
function that does the `cd` for it: the function runs pathpin, and when you pick a folder,
the function `cd`s there. Without it, pathpin prints the folder instead and says how to set it up.
The same line also sets up tab completion: `pathpin no<TAB>` completes `notes` (in bash, that
needs the `bash-completion` package; see your shell's section).

Not sure which shell you use? Run `echo $SHELL`. Then open your shell's section below and run
its commands **once**: each `echo … >>` adds a line to your shell's config file, so running it
again adds the line twice. That file is also where to change or remove them later.

The first command puts `~/.local/bin` on your `PATH`. Leave it out if `command -v pathpin`
already prints a path; if you installed [with Go](#with-go), use `$HOME/go/bin` instead.

- [bash on Linux](#bash-on-linux)
- [bash on macOS](#bash-on-macos)
- [zsh on Linux and macOS](#zsh-on-linux-and-macos)
- [fish on Linux and macOS](#fish-on-linux-and-macos)
- [Pick your own command name](#pick-your-own-command-name)

### bash on Linux

bash reads `~/.bashrc` when it starts in a terminal. Tab completion needs the `bash-completion`
package, which most distributions install by default.

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
echo 'eval "$(pathpin init bash)"' >> ~/.bashrc
source ~/.bashrc
```

To call it `pp` instead of `pathpin` (see [Pick your own command name](#pick-your-own-command-name)),
use this as the second line:

```bash
echo 'eval "$(pathpin init bash --cmd pp)"' >> ~/.bashrc
```

The menu's `e` key opens a folder in `$VISUAL`, else `$EDITOR`, else `vi`. To set your editor
(here Neovim; use your editor's command):

```bash
echo 'export EDITOR=nvim' >> ~/.bashrc
source ~/.bashrc
```

### bash on macOS

macOS's Terminal starts bash as a login shell, which reads `~/.bash_profile`, not `~/.bashrc`.
Tab completion needs the `bash-completion` package, which macOS doesn't include: install it
with Homebrew or MacPorts. Without it, everything else works, but tab prints
`_get_comp_words_by_ref: command not found`.

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bash_profile
echo 'eval "$(pathpin init bash)"' >> ~/.bash_profile
source ~/.bash_profile
```

To call it `pp` instead of `pathpin`, use this as the second line:

```bash
echo 'eval "$(pathpin init bash --cmd pp)"' >> ~/.bash_profile
```

To set the editor that `e` opens (here Neovim):

```bash
echo 'export EDITOR=nvim' >> ~/.bash_profile
source ~/.bash_profile
```

### zsh on Linux and macOS

zsh is macOS's default shell. It reads `~/.zshrc`, or `$ZDOTDIR/.zshrc` if you set `ZDOTDIR`
(then use that path in the commands below).

```zsh
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
echo 'eval "$(pathpin init zsh)"' >> ~/.zshrc
source ~/.zshrc
```

To call it `pp` instead of `pathpin`, use this as the second line:

```zsh
echo 'eval "$(pathpin init zsh --cmd pp)"' >> ~/.zshrc
```

To set the editor that `e` opens (here Neovim):

```zsh
echo 'export EDITOR=nvim' >> ~/.zshrc
source ~/.zshrc
```

### fish on Linux and macOS

fish reads `~/.config/fish/config.fish` (under `$XDG_CONFIG_HOME` if you set it).
`fish_add_path` saves the `PATH` change itself, so it needs no line in `config.fish`.

```fish
fish_add_path ~/.local/bin
echo 'pathpin init fish | source' >> ~/.config/fish/config.fish
source ~/.config/fish/config.fish
```

To call it `pp` instead of `pathpin`, use this as the second line:

```fish
echo 'pathpin init fish --cmd pp | source' >> ~/.config/fish/config.fish
```

To set the editor that `e` opens (here Neovim). `set -Ux` saves it for every fish session, so
it needs no line in `config.fish` either:

```fish
set -Ux EDITOR nvim
```

### Pick your own command name

The command is `pathpin` by default. `--cmd` on the `init` line names it something else: with
`--cmd pp`, `pp` opens the menu and `pp notes` jumps. Any name made of letters, digits, `_` and
`-` works, as long as it starts with a letter or `_`.

The function hides any other command with the same name in your shell, so pick one you don't
already use: before you set it up, `type pp` should say it's not found. The binary itself stays
`pathpin`, so the `init` line and updating keep using that name.

## Updating

Update the same way you installed:

| Installed with | To update |
|---|---|
| the install script | run the install line again (with the same `--bin-dir`, if you gave one) |
| Go | run `go install github.com/turtlethom/pathpin/v2@latest` again |
| a download by hand | run your system's commands again |
| source | `git pull`, then `go build -o ~/.local/bin/pathpin .` again |

`pathpin --version` shows the version you have, and the
[releases page](https://github.com/turtlethom/pathpin/releases) lists what's changed. Your shell
setup and bookmarks don't need anything.

## Coming from fcd

The first time you run pathpin, it imports your fcd bookmarks and leaves fcd's file in place.

## Uninstall

1. Remove the lines the setup added to your shell's config file (`~/.bashrc`, `~/.bash_profile`,
   `~/.zshrc` or `~/.config/fish/config.fish`): the `pathpin init` line, and the `PATH` and
   `EDITOR` lines if nothing else needs them. In fish, `set -Ue EDITOR` removes the editor.
2. Delete the binary: `rm ~/.local/bin/pathpin` (or `rm ~/go/bin/pathpin` if you installed with
   Go, or from the folder you gave `--bin-dir`).
3. To drop your bookmarks too, delete `~/.config/pathpin/` (or `$XDG_CONFIG_HOME/pathpin/`).
