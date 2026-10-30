#!/bin/sh
# shellcheck disable=SC2088 # every ~ here is text printed for the user's shell to expand
# Installs or updates pathpin from its GitHub releases:
#
#   curl -fsSL https://github.com/turtlethom/pathpin/releases/latest/download/install.sh | sh
#   curl -fsSL .../install.sh | sh -s -- --bin-dir ~/bin --version 2.0.1
#
# It downloads the archive for this system and the release's checksums.txt,
# installs only if the checksum matches, and never edits shell config files:
# on a first install it prints the setup commands for your shell instead.
# POSIX sh, since sh is dash on Debian and Ubuntu. Tests: install_test.sh.

repo=https://github.com/turtlethom/pathpin
docs=$repo/blob/main/docs/install.md#set-up-your-shell

say() { printf 'pathpin install: %s\n' "$*"; }
die() {
	printf 'pathpin install: %s\n' "$*" >&2
	exit 1
}

usage() {
	cat <<EOF
Install or update pathpin from $repo/releases.

Usage: sh install.sh [--bin-dir DIR] [--version VERSION]

  --bin-dir DIR       where to put pathpin (default: ~/.local/bin)
  --version VERSION   install this release, like 2.0.1 (default: the latest)
  -h, --help          show this help

Piped from curl, pass options after "sh -s --":
  curl -fsSL $repo/releases/latest/download/install.sh | sh -s -- --bin-dir ~/bin
EOF
}

# tilde DIR: DIR with $HOME shown as ~, for messages and the commands printed.
tilde() {
	case $1 in
	"$HOME" | "$HOME"/*) printf '~%s\n' "${1#"$HOME"}" ;;
	*) printf '%s\n' "$1" ;;
	esac
}

[ -n "${HOME:-}" ] || die "HOME isn't set"

# --- options

bin_dir=$HOME/.local/bin
version=
while [ $# -gt 0 ]; do
	case $1 in
	-h | --help)
		usage
		exit 0
		;;
	--bin-dir=*) bin_dir=${1#*=} ;;
	--version=*) version=${1#*=} ;;
	--bin-dir | --version)
		if [ $# -lt 2 ]; then
			usage >&2
			die "$1 needs a value"
		fi
		case $1 in
		--bin-dir) bin_dir=$2 ;;
		*) version=$2 ;;
		esac
		shift
		;;
	*)
		usage >&2
		die "unknown option: $1"
		;;
	esac
	shift
done
[ -n "$bin_dir" ] || die "--bin-dir needs a folder"

# --- this system, and what's needed, before anything is downloaded or written

system=$(uname -s)
case $system in
Linux) os=linux ;;
Darwin) os=darwin ;;
*) die "pathpin runs on Linux and macOS, not $system" ;;
esac

cpu=$(uname -m)
case $cpu in
x86_64 | amd64) arch=amd64 ;;
aarch64 | arm64) arch=arm64 ;;
*) die "there's no pathpin build for $cpu, only for x86-64 and ARM64" ;;
esac

command -v curl >/dev/null 2>&1 || die "curl is needed to download pathpin"

# The download is installed only if its checksum matches, so no tool means
# no install, never an unchecked one.
if command -v sha256sum >/dev/null 2>&1; then
	check() { sha256sum -c "$@"; }
elif command -v shasum >/dev/null 2>&1; then
	check() { shasum -a 256 -c "$@"; }
else
	die "can't check the download: neither sha256sum nor shasum is installed"
fi

# An absolute folder, since the PATH line printed below runs in other shells.
case $bin_dir in
/*) ;;
*) bin_dir=$(pwd)/$bin_dir ;;
esac
[ "$bin_dir" = / ] || bin_dir=${bin_dir%/}

if ! mkdir -p "$bin_dir" 2>/dev/null || [ ! -w "$bin_dir" ]; then
	die "can't write to $(tilde "$bin_dir")
  Pick a folder you own with --bin-dir, or install there with sudo:
  curl -fsSL $repo/releases/latest/download/install.sh | sudo sh -s -- --bin-dir '$bin_dir'"
fi

# --- which release

if [ -z "$version" ]; then
	# Where GitHub redirects releases/latest (.../tag/v2.0.3); no API call.
	url=$(curl -fsSLo /dev/null -w '%{url_effective}' "$repo/releases/latest") ||
		die "can't reach $repo/releases/latest"
	version=${url##*/tag/}
fi
version=${version#v}
case $version in
"" | *[!0-9A-Za-z.+-]*) die "not a pathpin version: $version" ;;
esac

old=
installed=
if [ -e "$bin_dir/pathpin" ]; then
	installed=1
	if old=$("$bin_dir/pathpin" --version 2>/dev/null); then
		old=${old#pathpin }
	else
		old= # another system's build, or not pathpin at all
	fi
fi

# --- download, check, install

tmp=$(mktemp -d) || die "can't make a temporary folder"
staged=
trap 'rm -rf "$tmp"; [ -z "$staged" ] || rm -f "$staged"' EXIT
trap 'exit 1' HUP INT TERM

archive=pathpin_${os}_${arch}.tar.gz
base=$repo/releases/download/v$version
say "downloading pathpin $version for $os/$arch"
curl -fsSL -o "$tmp/$archive" "$base/$archive" ||
	die "can't download $base/$archive (is $version a release? see $repo/releases)"
curl -fsSL -o "$tmp/checksums.txt" "$base/checksums.txt" ||
	die "can't download $base/checksums.txt"

# Both files come from the same tag, so a release published meanwhile can't mix them.
(cd "$tmp" && grep "  $archive\$" checksums.txt | check >/dev/null 2>&1) ||
	die "checksum mismatch: $archive is damaged or was changed, so nothing was installed"
say "checksum OK"

tar -xzf "$tmp/$archive" -C "$tmp" pathpin || die "can't unpack $archive"

# A new file renamed over the old one: writing into the old one fails with
# "Text file busy" while a pathpin menu is open, and a rename is all or nothing.
staged=$bin_dir/.pathpin.$$
if ! { cp "$tmp/pathpin" "$staged" && chmod 755 "$staged" && mv -f "$staged" "$bin_dir/pathpin"; }; then
	die "can't install into $(tilde "$bin_dir")"
fi
staged=

where=$(tilde "$bin_dir")
if [ -z "$installed" ]; then
	say "installed pathpin $version in $where"
elif [ "$old" = "$version" ]; then
	say "reinstalled pathpin $version in $where"
elif [ -n "$old" ]; then
	say "updated pathpin $old → $version in $where"
else
	say "updated pathpin to $version in $where"
fi

# Another pathpin found first on PATH (say ~/go/bin's, from go install) is the
# one that runs, so "updated" would look like it did nothing. Folders are
# compared with symlinks and trailing slashes resolved.
if found=$(command -v pathpin 2>/dev/null) &&
	[ "$(cd "${found%/*}" && pwd -P)" != "$(cd "$bin_dir" && pwd -P)" ]; then
	other=$(tilde "$found")
	if v=$("$found" --version 2>/dev/null); then
		other="$other (${v#pathpin })"
	fi
	say "note: pathpin still runs $other, found first on your PATH:"
	echo "  remove it, or put $where before its folder on PATH"
fi

# --- first install: how to set up the shell (never done for you)

[ -z "$installed" ] || exit 0

case :$PATH: in
*:"$bin_dir":*) on_path=1 ;;
*) on_path= ;;
esac

# The folder as the PATH line writes it: $HOME/... for sh, ~/... for fish.
case $bin_dir in
"$HOME" | "$HOME"/*)
	sh_dir="\$HOME${bin_dir#"$HOME"}"
	fish_dir=$where
	;;
*)
	sh_dir=$bin_dir
	fish_dir="'$bin_dir'"
	;;
esac

shell=${SHELL:-}
shell=${shell##*/}
case $shell in
bash)
	rc='~/.bashrc'
	[ "$os" = darwin ] && rc='~/.bash_profile' # Terminal starts bash as a login shell
	path_line="echo 'export PATH=\"$sh_dir:\$PATH\"' >> $rc"
	init_line="echo 'eval \"\$(pathpin init bash)\"' >> $rc"
	;;
zsh)
	rc='~/.zshrc'
	case ${ZDOTDIR:-} in /*) rc=$(tilde "$ZDOTDIR/.zshrc") ;; esac
	path_line="echo 'export PATH=\"$sh_dir:\$PATH\"' >> $rc"
	init_line="echo 'eval \"\$(pathpin init zsh)\"' >> $rc"
	;;
fish)
	rc='~/.config/fish/config.fish'
	case ${XDG_CONFIG_HOME:-} in /*) rc=$(tilde "$XDG_CONFIG_HOME/fish/config.fish") ;; esac
	path_line="fish_add_path $fish_dir"
	init_line="echo 'pathpin init fish | source' >> $rc"
	;;
*)
	echo
	echo "Next, set up your shell: $docs"
	exit 0
	;;
esac

echo
echo "Next, set up your shell ($shell) by running these once:"
echo
[ -n "$on_path" ] || echo "  $path_line"
echo "  $init_line"
echo "  source $rc"
echo
echo "Other shells, and a shorter name than pathpin: $docs"
