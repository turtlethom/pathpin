#!/usr/bin/env bash
# Spec tests for install.sh: ./install_test.sh [test_name...]
#
# They run this checkout's install.sh with sh (dash on Debian and Ubuntu, so
# bash-only syntax fails) against the real releases on GitHub: the latest one,
# and an older one for --version. So they need the network, and they check
# the script against releases that are already published. Each test gets a
# throwaway $HOME, a clean environment and its own folders. Fake commands put
# first on PATH stand in for other systems (uname) and broken downloads (curl).

here=$(cd "$(dirname "$0")" && pwd)
script=$here/install.sh
repo=https://github.com/turtlethom/pathpin
older=2.0.1 # any published release older than the latest
docs="$repo/blob/main/docs/install.md#set-up-your-shell"

# The latest release's version, from where GitHub redirects releases/latest.
latest=$(curl -fsSLo /dev/null -w '%{url_effective}' "$repo/releases/latest") || {
	echo "can't reach $repo/releases/latest" >&2
	exit 1
}
latest=${latest##*/v}

# bash on macOS reads ~/.bash_profile (Terminal starts a login shell).
case $(uname -s) in
Darwin) bash_rc='~/.bash_profile' ;;
*) bash_rc='~/.bashrc' ;;
esac

# Per test: $t (its folder), $home, $bin (the default install folder), $fake
# (first on PATH), and from run_install: $out and $status.

# run_install [VAR=value...] [--] [install.sh arguments...]
# Runs install.sh from $t in a clean environment: HOME, PATH (with $fake
# first), SHELL=/bin/bash, then any VAR=value given (a later one wins).
run_install() {
	local vars=("HOME=$home" "PATH=$fake:$PATH" "SHELL=/bin/bash")
	while [ $# -gt 0 ]; do
		case $1 in
		--) shift && break ;;
		[A-Z]*=*) vars+=("$1") && shift ;;
		*) break ;;
		esac
	done
	out=$(cd "$t" && env -i "${vars[@]}" sh "$script" "$@" 2>&1)
	status=$?
}

# fake NAME BODY: a NAME command first on PATH, running the sh in BODY, where
# $real is the real command.
fake() {
	printf '#!/bin/sh\nreal=%s\n%s\n' "$(command -v "$1")" "$2" >"$fake/$1"
	chmod +x "$fake/$1"
}

# without NAME...: prints a PATH holding every command on today's PATH but
# these (links in $t/path).
without() {
	local dir name
	mkdir -p "$t/path"
	local IFS=:
	for dir in $PATH; do
		[ -d "$dir" ] && ln -s "$dir"/* "$t/path/" 2>/dev/null # the first of a name wins
	done
	for name; do rm -f "$t/path/$name"; done
	echo "$fake:$t/path"
}

fail() {
	printf '    %s\n' "$@"
	failed=1
}

expect_status() { # 0, or error
	if [ "$1" = 0 ]; then
		[ "$status" = 0 ] || fail "exit status $status, want 0"
	else
		[ "$status" != 0 ] || fail "exit status 0, want an error"
	fi
}

expect_out() {
	local s
	for s; do
		case $out in *"$s"*) ;; *) fail "output lacks: $s" ;; esac
	done
}

expect_no_out() {
	local s
	for s; do
		case $out in *"$s"*) fail "output has: $s" ;; esac
	done
}

# expect_lines: every line of stdin is in the output.
expect_lines() {
	local line
	while IFS= read -r line; do expect_out "$line"; done
}

expect_order() { # first second
	case $out in *"$1"*"$2"*) ;; *) fail "output lacks, in this order: $1 / $2" ;; esac
}

expect_version() { # dir version
	local got
	got=$("$1/pathpin" --version 2>&1)
	[ "$got" = "pathpin $2" ] || fail "$1/pathpin --version says: $got, want: pathpin $2"
}

expect_files() { # dir names...: the folder holds exactly these (no temp files left)
	local dir=$1 got want= name
	shift
	got=$(ls -A "$dir" 2>/dev/null | tr '\n' ' ')
	for name; do want+="$name "; done
	[ "$got" = "$want" ] || fail "$dir holds: '$got', want: '$want'"
}

expect_arch() { # file amd64|arm64
	local got pattern
	got=$(file -b "$1")
	case $2 in
	amd64) pattern='x86[-_]64' ;;
	arm64) pattern='aarch64|arm64' ;;
	esac
	echo "$got" | grep -Eq "$pattern" || fail "$1 is: $got, want $2"
}

inode() { ls -i "$1" 2>/dev/null | awk '{print $1}'; }

# --- installing

test_fresh_install() {
	run_install
	expect_status 0
	expect_out "installed pathpin $latest in ~/.local/bin"
	expect_files "$bin" pathpin
	expect_version "$bin" "$latest"
}

test_update_from_older() {
	run_install -- --version "$older"
	run_install
	expect_status 0
	expect_out "updated pathpin $older → $latest"
	expect_version "$bin" "$latest"
	expect_no_out "pathpin init" # the shell was set up on the first install
}

test_reinstall_replaces_the_file() {
	run_install
	local before
	before=$(inode "$bin/pathpin")
	run_install
	expect_status 0
	expect_out "reinstalled pathpin $latest"
	expect_no_out "pathpin init"
	expect_files "$bin" pathpin
	# A new file renamed over the old one: writing into the old one fails
	# with "Text file busy" while a menu is open.
	[ "$(inode "$bin/pathpin")" != "$before" ] || fail "pathpin was rewritten in place, not replaced"
}

test_version_rolls_back() {
	run_install
	run_install -- --version "v$older"
	expect_status 0
	expect_out "updated pathpin $latest → $older"
	expect_version "$bin" "$older"
}

test_version_with_equals() {
	run_install -- --version="$older"
	expect_status 0
	expect_out "installed pathpin $older"
	expect_version "$bin" "$older"
}

test_unknown_version_changes_nothing() {
	run_install
	run_install -- --version 9.9.9
	expect_status error
	expect_out "9.9.9"
	expect_version "$bin" "$latest"
	expect_files "$bin" pathpin
}

# --- the install folder

test_bin_dir_is_created() {
	run_install -- --bin-dir "$t/my bin"
	expect_status 0
	expect_files "$t/my bin" pathpin
	expect_files "$home"
	expect_out "echo 'export PATH=\"$t/my bin:\$PATH\"' >> $bash_rc"
}

test_bin_dir_with_equals() {
	run_install -- --bin-dir="$t/b"
	expect_status 0
	expect_files "$t/b" pathpin
}

test_bin_dir_in_home_uses_HOME() {
	run_install -- --bin-dir "$home/bin"
	expect_status 0
	expect_out "installed pathpin $latest in ~/bin"
	expect_out "echo 'export PATH=\"\$HOME/bin:\$PATH\"' >> $bash_rc"
}

test_relative_bin_dir() {
	run_install -- --bin-dir rel
	expect_status 0
	expect_files "$t/rel" pathpin
	expect_out "/rel:\$PATH\"' >> $bash_rc" # an absolute path: rel means nothing in a new shell
	expect_no_out "\"rel:\$PATH"
}

test_unwritable_bin_dir() {
	if [ "$(id -u)" = 0 ]; then
		echo "    skipped: root can write anywhere"
		return
	fi
	mkdir "$t/locked dir" && chmod 555 "$t/locked dir"
	run_install -- --bin-dir "$t/locked dir"
	expect_status error
	expect_out "sudo sh -s -- --bin-dir '$t/locked dir'" # quoted, so it pastes as one folder
	expect_files "$t/locked dir"
}

# other_pathpin: a pathpin 1.0.0 in $t/other, like one from go install.
other_pathpin() {
	mkdir -p "$t/other"
	printf '#!/bin/sh\necho "pathpin 1.0.0"\n' >"$t/other/pathpin"
	chmod +x "$t/other/pathpin"
}

test_earlier_pathpin_is_named() {
	other_pathpin
	run_install "PATH=$fake:$t/other:$bin:$PATH"
	expect_status 0
	expect_out "pathpin still runs $t/other/pathpin (1.0.0)" "put ~/.local/bin before"
}

test_earlier_pathpin_is_named_on_update() {
	run_install
	other_pathpin
	run_install "PATH=$fake:$t/other:$bin:$PATH"
	expect_status 0
	expect_out "pathpin still runs $t/other/pathpin (1.0.0)"
}

test_later_pathpin_is_fine() {
	other_pathpin
	run_install "PATH=$fake:$bin/:$t/other:$PATH" # bin with a trailing slash is still bin
	expect_status 0
	expect_no_out "still runs"
}

# --- shell setup: printed on a first install, never written

test_setup_bash() {
	run_install SHELL=/bin/bash
	expect_status 0
	expect_lines <<EOF
echo 'export PATH="\$HOME/.local/bin:\$PATH"' >> $bash_rc
echo 'eval "\$(pathpin init bash)"' >> $bash_rc
source $bash_rc
$docs
EOF
	expect_order "export PATH" "pathpin init bash" # init runs pathpin, so PATH comes first
	expect_files "$home" .local                    # no shell config written
}

test_setup_zsh() {
	run_install SHELL=/usr/bin/zsh
	expect_status 0
	expect_lines <<EOF
echo 'export PATH="\$HOME/.local/bin:\$PATH"' >> ~/.zshrc
echo 'eval "\$(pathpin init zsh)"' >> ~/.zshrc
source ~/.zshrc
EOF
	expect_files "$home" .local
}

test_setup_zsh_zdotdir() {
	run_install SHELL=/bin/zsh "ZDOTDIR=$home/.config/zsh"
	expect_status 0
	expect_out "echo 'eval \"\$(pathpin init zsh)\"' >> ~/.config/zsh/.zshrc"
}

test_setup_fish() {
	run_install SHELL=/usr/bin/fish
	expect_status 0
	expect_lines <<EOF
fish_add_path ~/.local/bin
echo 'pathpin init fish | source' >> ~/.config/fish/config.fish
source ~/.config/fish/config.fish
EOF
	expect_files "$home" .local
}

test_setup_fish_xdg() {
	run_install SHELL=/usr/bin/fish "XDG_CONFIG_HOME=$home/xdg"
	expect_status 0
	expect_out "echo 'pathpin init fish | source' >> ~/xdg/fish/config.fish"
}

test_setup_skips_path_when_on_path() {
	run_install "PATH=$fake:$home/.local/bin:$PATH"
	expect_status 0
	expect_out "pathpin init bash"
	expect_no_out "export PATH"
}

test_setup_unknown_shell() {
	run_install SHELL=/bin/tcsh
	expect_status 0
	expect_out "$docs"
	expect_no_out "pathpin init"
}

test_setup_no_shell() {
	run_install SHELL=
	expect_status 0
	expect_out "$docs"
	expect_no_out "pathpin init"
}

# --- systems

test_arch_names() {
	local m want
	for m in x86_64 aarch64 arm64; do
		case $m in x86_64) want=amd64 ;; *) want=arm64 ;; esac
		fake uname "[ \"\$1\" = -m ] && { echo $m; exit; }; exec \"\$real\" \"\$@\""
		rm -rf "$home/.local"
		run_install
		expect_status 0
		expect_arch "$bin/pathpin" "$want"
	done
}

test_unsupported_arch() {
	fake uname "[ \"\$1\" = -m ] && { echo armv7l; exit; }; exec \"\$real\" \"\$@\""
	run_install
	expect_status error
	expect_out "armv7l"
	expect_files "$home"
}

test_unsupported_os() {
	fake uname "[ \"\$1\" = -m ] && exec \"\$real\" -m; echo FreeBSD"
	run_install
	expect_status error
	expect_out "FreeBSD" "Linux and macOS"
	expect_files "$home"
}

# --- refusing a download it can't trust

test_bad_checksum_installs_nothing() {
	# The real curl, then every hash in a downloaded checksums.txt changed,
	# however install.sh saves it (-o FILE, -O, or stdout).
	fake curl '
url= out= remote= next=
for a in "$@"; do
	[ -n "$next" ] && out=$a && next=
	case $a in
	http*) url=$a ;;
	-o | --output) next=1 ;;
	--remote-name) remote=1 ;;
	--*) ;;
	-*o) next=1 ;;
	-*O*) remote=1 ;;
	esac
done
spoil() { awk "{ c = substr(\$1, 1, 1); print (c == \"0\" ? \"1\" : \"0\") substr(\$1, 2) \"  \" \$2 }"; }
case $url in *checksums.txt) ;; *) exec "$real" "$@" ;; esac
[ -z "$out" ] && [ -n "$remote" ] && out=checksums.txt
if [ -n "$out" ]; then
	"$real" "$@" || exit
	spoil <"$out" >"$out.spoiled" && mv "$out.spoiled" "$out"
else
	body=$("$real" "$@") || exit
	printf "%s\n" "$body" | spoil
fi'
	run_install
	expect_status error
	expect_out "checksum"
	expect_files "$bin"
}

test_either_checksum_tool() {
	# sha256sum alone on Linux only: every Mac has shasum, which comes first,
	# and macOS's own sha256sum isn't GNU's.
	if [ "$(uname -s)" = Linux ]; then
		run_install "PATH=$(without shasum)"
		expect_status 0
		expect_version "$bin" "$latest"
		rm -rf "$home/.local" "$t/path"
	fi
	run_install "PATH=$(without sha256sum)"
	expect_status 0
	expect_version "$bin" "$latest"
}

test_no_checksum_tool() {
	run_install "PATH=$(without sha256sum shasum)"
	expect_status error
	expect_out "sha256sum" "shasum"
	expect_files "$bin"
}

# --- arguments

test_help() {
	run_install -- --help
	expect_status 0
	expect_out "--bin-dir" "--version"
	expect_files "$home"
}

test_unknown_flag() {
	run_install -- --frobnicate
	expect_status error
	expect_out "--frobnicate" "--bin-dir" "--version"
	expect_files "$home"
}

test_flag_without_value() {
	run_install -- --bin-dir
	expect_status error
	expect_out "--bin-dir"
	run_install -- --version
	expect_status error
	expect_out "--version"
	expect_files "$home"
}

# --- runner

if [ $# -gt 0 ]; then
	tests=("$@")
else
	tests=($(declare -F | awk '$3 ~ /^test_/ { print $3 }'))
fi
failures=0
for name in "${tests[@]}"; do
	t=$(mktemp -d)
	home=$t/home bin=$t/home/.local/bin fake=$t/fake
	mkdir -p "$home" "$fake"
	out= status= failed=0
	"$name"
	if [ "$failed" = 0 ]; then
		echo "ok    $name"
	else
		echo "FAIL  $name, whose last run printed:"
		printf '%s\n' "$out" | sed 's/^/    | /'
		failures=$((failures + 1))
	fi
	chmod -R u+w "$t" && rm -rf "$t"
done
echo
if [ "$failures" = 0 ]; then
	echo "all ${#tests[@]} passed (latest release: $latest)"
else
	echo "$failures of ${#tests[@]} failed"
	exit 1
fi
