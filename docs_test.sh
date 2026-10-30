#!/usr/bin/env bash
# Smoke test for the shell setup in the docs: ./docs_test.sh
#
# docs/install.md's per-shell sections promise that pasting their commands
# sets pathpin up. For each one, this pastes the commands as written (read
# from the file, so the docs can't drift from what's tested) into the shell,
# with a throwaway $HOME, then starts a new shell the way a terminal would and
# checks that `pathpin <label>` changes directory. It does the same with the
# section's `--cmd pp` line and its EDITOR commands, and checks that the
# README's collapsed blocks match docs/install.md.
#
# Both bash sections run on every system: they differ only in which file the
# terminal reads (~/.bashrc for an interactive shell, ~/.bash_profile for a
# login shell, as on macOS). Builds pathpin from this checkout; no network.
# A shell that isn't installed is skipped, or fails in CI (CI=true).

here=$(cd "$(dirname "$0")" && pwd)
docs=$here/docs/install.md
readme=$here/README.md

t=$(mktemp -d) && t=$(cd "$t" && pwd -P) || exit 1 # physical, so pwd prints what pathpin stores
trap 'rm -rf "$t"' EXIT
(cd "$here" && go build -o "$t/pathpin" .) || exit 1

# Apple's bash 3.2 on macOS (/bin/bash), not a newer one from Homebrew: it's
# what Terminal runs.
bash=/bin/bash
zsh=$(command -v zsh)
fish=$(command -v fish)

# The new shells' PATH: the system folders and the shells' own, never a
# folder holding a pathpin, so only the docs' PATH line can make it work.
base=/usr/bin:/bin:/usr/sbin:/sbin
[ -z "$fish" ] || base=$base:${fish%/*}
if PATH=$base command -v pathpin >/dev/null; then
	echo "a pathpin is already on $base: $(PATH=$base command -v pathpin)" >&2
	exit 1
fi

# block FILE START N: the Nth fenced code block after the line START, before
# the next heading or </details>.
block() {
	awk -v start="$2" -v n="$3" '
		$0 == start { on = 1; next }
		!on { next }
		!inside && (/^#/ || /^<\/details>/) { exit }
		/^```/ {
			if (inside) { inside = 0; if (++done == n) exit } else inside = 1
			next
		}
		inside && done == n - 1 { print }
	' "$1"
}

# paste SHELL HOME COMMANDS: the commands run as if pasted into SHELL. Every
# shell here reads /dev/null, never the terminal this runs in: fish, for one,
# queries the terminal when it starts.
paste() {
	local sh=$1 home=$2
	shift 2
	case $sh in
	bash) env -i HOME="$home" PATH="$base" TERM=dumb "$bash" --norc --noprofile -c "$1" ;;
	zsh) env -i HOME="$home" PATH="$base" TERM=dumb "$zsh" -f -c "$1" ;;
	fish) env -i HOME="$home" PATH="$base" TERM=dumb "$fish" -c "$1" ;;
	esac </dev/null
}

# terminal SECTION HOME COMMANDS: the commands in a new shell started the way
# a terminal starts it, reading the files the section wrote to.
terminal() {
	local section=$1 home=$2
	shift 2
	case $section in
	"bash on Linux") env -i HOME="$home" PATH="$base" TERM=dumb "$bash" -i -c "$1" ;;
	"bash on macOS") env -i HOME="$home" PATH="$base" TERM=dumb "$bash" -l -i -c "$1" ;;
	zsh*) env -i HOME="$home" PATH="$base" TERM=dumb "$zsh" -i -c "$1" ;;
	fish*) env -i HOME="$home" PATH="$base" TERM=dumb "$fish" -c "$1" ;;
	esac </dev/null
}

failures=0 skipped=0
fail() {
	echo "FAIL  $1"
	shift
	printf '    %s\n' "$@"
	failures=$((failures + 1))
}

# jump_works SECTION HOME CMD: in a new terminal, `CMD add` in ~/proj and
# `CMD proj` from / land in ~/proj. Checks pwd's line, marked: without the
# shell function, the bare binary prints the folder instead of going there.
jump_works() {
	local section=$1 home=$2 cmd=$3 out
	mkdir -p "$home/proj"
	out=$(terminal "$section" "$home" "cd ~/proj && $cmd add >/dev/null && cd / && $cmd proj && pwd | sed s/^/pwd=/ && echo \"editor=\$EDITOR\"" 2>&1)
	last_out=$out
	printf '%s\n' "$out" | grep -Fqx "pwd=$home/proj"
}

# --- the README's blocks are the docs' blocks

while IFS='|' read -r heading summary; do
	want=$(block "$docs" "$heading" 1)
	got=$(block "$readme" "$summary" 1)
	if [ -z "$want" ]; then
		fail "docs: $heading" "no code block after it in docs/install.md"
	elif [ "$got" != "$want" ]; then
		fail "README: $summary" "its block differs from docs/install.md's $heading:" "$(diff <(echo "$want") <(echo "$got"))"
	else
		echo "ok    README matches: ${heading#\#\#\# }"
	fi
done <<'EOF'
### bash on Linux|<summary><b>bash</b> on Linux</summary>
### bash on macOS|<summary><b>bash</b> on macOS</summary>
### zsh on Linux and macOS|<summary><b>zsh</b> on Linux and macOS (macOS's default shell)</summary>
### fish on Linux and macOS|<summary><b>fish</b> on Linux and macOS</summary>
EOF

# --- each section, pasted for real

for section in "bash on Linux" "bash on macOS" "zsh on Linux and macOS" "fish on Linux and macOS"; do
	sh=${section%% *}
	if [ -z "${!sh}" ] || [ ! -x "${!sh}" ]; then
		if [ "${CI:-}" = true ]; then
			fail "$section" "$sh isn't installed"
		else
			echo "skip  $section ($sh isn't installed)"
			skipped=$((skipped + 1))
		fi
		continue
	fi
	setup=$(block "$docs" "### $section" 1)
	pp_line=$(block "$docs" "### $section" 2)
	editor=$(block "$docs" "### $section" 3)
	init=$(printf '%s\n' "$setup" | sed -n 2p)
	case $init in
	*"pathpin init"*) ;;
	*)
		fail "$section" "the setup's second line should be the pathpin init line (the --cmd pp line replaces it), not: $init"
		continue
		;;
	esac

	# As after the install script: the binary in ~/.local/bin, not yet on PATH.
	home=$t/$sh-${section##* }-1
	mkdir -p "$home/.local/bin" && cp "$t/pathpin" "$home/.local/bin/"
	paste "$sh" "$home" "$setup" >"$t/paste.log" 2>&1 || fail "$section: setup" "pasting it failed:" "$(cat "$t/paste.log")"
	paste "$sh" "$home" "$editor" >"$t/paste.log" 2>&1 || fail "$section: EDITOR" "pasting it failed:" "$(cat "$t/paste.log")"
	if ! jump_works "$section" "$home" pathpin; then
		fail "$section: setup" "in a new terminal, pathpin proj didn't land in ~/proj. It printed:" "$last_out"
	elif ! printf '%s\n' "$last_out" | grep -Fqx "editor=nvim"; then
		fail "$section: EDITOR" "in a new terminal, EDITOR isn't nvim. It printed:" "$last_out"
	else
		echo "ok    $section: setup, EDITOR"
	fi

	# The --cmd pp line in place of the setup's second line (the init line).
	home=$t/$sh-${section##* }-pp
	mkdir -p "$home/.local/bin" && cp "$t/pathpin" "$home/.local/bin/"
	with_pp=$(printf '%s\n' "$setup" | sed -n 1p; printf '%s\n' "$pp_line"; printf '%s\n' "$setup" | sed -n '3,$p')
	paste "$sh" "$home" "$with_pp" >"$t/paste.log" 2>&1 || fail "$section: --cmd pp" "pasting it failed:" "$(cat "$t/paste.log")"
	if jump_works "$section" "$home" pp; then
		echo "ok    $section: --cmd pp"
	else
		fail "$section: --cmd pp" "in a new terminal, pp proj didn't land in ~/proj. It printed:" "$last_out"
	fi
done

echo
if [ "$failures" != 0 ]; then
	echo "$failures failed"
	exit 1
fi
if [ "$skipped" = 0 ]; then echo "all passed"; else echo "all passed, $skipped skipped"; fi
