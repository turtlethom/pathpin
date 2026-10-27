# The README's demo GIF. Record it from the repo's root with:
#
#   vhs demo/demo.tape
#
# Needs vhs, ttyd and ffmpeg (https://github.com/charmbracelet/vhs). It builds
# pathpin and runs it in a throwaway $HOME with demo/config.toml, so your own
# config is never touched. It builds whatever is checked out, uncommitted
# changes included, so record from a committed tree.

Output demo/demo.gif

Require go

Set Shell bash
# Tall enough for all 7 bookmarks on All, and for several rows above the t panel.
Set FontSize 18
Set Width 1100
Set Height 900
Set Padding 24
Set Theme "Catppuccin Mocha"
Set TypingSpeed 80ms

# Setup, not recorded: build into a temp folder (with the real $HOME, so Go's
# build cache is used), then switch $HOME and load the shell function as pp.
# Every later line first checks that $HOME is the temp one: if the build
# failed, they do nothing, so the real ~/.config/pathpin is never overwritten.
Hide
Type@1ms `d=$(mktemp -d) && go build -o "$d/bin/pathpin" . && export PATH="$d/bin:$PATH" HOME="$d/home" XDG_CONFIG_HOME=` Enter
Type@1ms `test "$HOME" = "$d/home" && mkdir -p ~/.config/pathpin ~/notes ~/dotfiles/nvim ~/dotfiles/bash ~/school/cs310 ~/school/thesis ~/work/api ~/work/web` Enter
Type@1ms `test "$HOME" = "$d/home" && cp demo/config.toml ~/.config/pathpin/ && cd ~ && eval "$(pathpin init bash --cmd pp)"` Enter
Type@1ms `test "$HOME" = "$d/home" && PS1='\[\e[1;34m\]\w\[\e[0m\] $ ' && clear` Enter
# The demo prompt ("~ $") shows only once all of that ran, however long the
# build takes. If setup failed, it never shows and vhs stops with an error.
Wait@60s /~ \$/
Show

# Open the menu and look around.
Type "pp" Sleep 400ms Enter
Sleep 1.5s
Down Sleep 500ms Down Sleep 500ms Down
Sleep 1s

# Groups are tabs: All, Config, School, Work, Other, and around to All again.
Tab Sleep 900ms Tab Sleep 900ms Tab Sleep 900ms Tab
Sleep 900ms
Tab Sleep 1s

# Filter All, then go there.
Type "/" Sleep 400ms Type "thes"
Sleep 1s
Enter Sleep 600ms Enter
Sleep 1.5s

# Or jump straight from the shell.
Type "pp api" Sleep 400ms Enter
Sleep 1.5s

# The appearance panel previews themes live; esc leaves everything as it was.
Type "pp" Sleep 400ms Enter
Sleep 1s
Type "t"
Sleep 1.2s
Right Sleep 1.2s Right Sleep 1.2s Right
Sleep 1.5s
Escape Sleep 800ms
Type "q"
Sleep 1.5s

# Not recorded: remove the temp folder (binary and fake home).
Hide
Type@1ms `cd / && rm -rf "$d"` Enter
Sleep 500ms
