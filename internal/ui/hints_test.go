package ui

import (
	"slices"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/turtlethom/pathpin/v2/internal/config"
)

const nordErr = "#bf616a" // nord's error color, from themes.go

// helpLine is the visible text of the short help line, trimmed.
func helpLine(t *testing.T, m Model) string {
	t.Helper()
	return strings.TrimSpace(plain(rawLine(t, m.screen(), "↑/k up")))
}

// The help line is never wider than the window. When it doesn't fit, it
// stops before the first key that doesn't, so every key shown is whole, and
// ends with …. At 80 columns nothing is cut, even with groups: the help may
// use the window's last column, while the list keeps a column of margin on
// each side.
func TestHelpLineFits(t *testing.T) {
	full := helpLine(t, groupModel(t).sized(200, 24))
	for _, w := range []int{80, 50, 30, 20} {
		m := groupModel(t).sized(w, 24)
		if m.list.Width() != w-2 {
			t.Errorf("%d columns: the list is %d wide, want %d (a column of margin on each side)", w, m.list.Width(), w-2)
		}
		for l := range strings.SplitSeq(m.screen(), "\n") {
			if got := lipgloss.Width(l); got > w {
				t.Errorf("%d columns: a line is %d wide: %q", w, got, plain(l))
			}
		}
		line := helpLine(t, m)
		if w == 80 {
			if line != full {
				t.Errorf("80 columns: want the whole help line %q, got %q", full, line)
			}
			continue
		}
		shown, cut := strings.CutSuffix(line, " …")
		if !cut || !strings.HasPrefix(full, shown) || !strings.HasPrefix(full[len(shown):], " • ") {
			t.Errorf("%d columns: want whole keys and then …, got %q (full: %q)", w, line, full)
		}
	}
}

// pickerKeysLine is the picker's bottom line, every key in it.
const pickerKeysLine = "↑↓/jk row • ←→/hl change • # hex • d reset • enter save • esc cancel"

// The picker's bottom line names every key, with the letter keys next to the
// arrows (h and l were missing). In a narrow window it's cut like the help
// line: whole keys, then ….
func TestPickerKeysLine(t *testing.T) {
	for _, w := range []int{80, 40} {
		m := openPicker(t, colorsModel(t, config.Colors{}).sized(w, 24))
		line := strings.TrimSpace(plain(rawLine(t, m.screen(), "↑↓/jk")))
		if w == 80 {
			if line != pickerKeysLine {
				t.Errorf("80 columns: want %q, got %q", pickerKeysLine, line)
			}
			continue
		}
		if got := lipgloss.Width(rawLine(t, m.screen(), "↑↓/jk")); got > w {
			t.Errorf("%d columns: the keys line is %d wide", w, got)
		}
		shown, cut := strings.CutSuffix(line, " …")
		if !cut || !strings.HasPrefix(pickerKeysLine, shown) || !strings.HasPrefix(pickerKeysLine[len(shown):], " • ") {
			t.Errorf("%d columns: want whole keys and then …, got %q", w, line)
		}
	}
}

// ? says that shift+tab goes back a tab. The short help line has no room for
// it and keeps "tab group".
func TestTabKeysInHelp(t *testing.T) {
	m := groupModel(t)
	if line := helpLine(t, m); !strings.Contains(line, "tab group") || strings.Contains(line, "⇧") {
		t.Errorf("short help: want tab group as before, got %q", line)
	}
	m, _ = press(m, runes("?"))
	if !helpShows(m, "tab/⇧tab", "group") {
		t.Errorf("? should list tab/⇧tab group:\n%s", plain(m.screen()))
	}
}

// checkHint checks one "key desc" hint on a rendered line: the key in the
// accent, the description muted.
func checkHint(t *testing.T, name, line, key, desc string) {
	t.Helper()
	_, k, _ := find(t, cells(line), key+" "+desc)
	for i, c := range k {
		want := fg(nordMuted)
		if i < len([]rune(key)) {
			want = fg(nordAccent)
		} else if c.r == ' ' {
			continue
		}
		if c.fg != want {
			t.Errorf("%s: in %q, %q is drawn in %q, want %q", name, key+" "+desc, string(c.r), c.fg, want)
			return
		}
	}
}

// Every key hint looks the same: keys in the accent, descriptions and the •
// between hints muted. The picker and the fields use • like the help line.
func TestHintColors(t *testing.T) {
	withColor(t)
	m := nordModel(t, config.Bookmark{Label: "NOTES", Path: "notes"}, config.Bookmark{Label: "WORK", Path: "work"})

	help := rawLine(t, m.screen(), "↑/k up")
	checkHint(t, "help line", help, "q", "quit")
	checkHint(t, "help line", help, "/", "filter")
	for _, c := range cells(help) {
		if c.r == '•' && c.fg != fg(nordMuted) {
			t.Errorf("help line: • is drawn in %q, want muted", c.fg)
		}
	}

	picker, _ := press(m, runes("t"))
	line := rawLine(t, picker.screen(), "reset")
	if want := pickerKeysLine; !strings.Contains(plain(line), want) {
		t.Errorf("picker: want %q, got %q", want, plain(line))
	}
	checkHint(t, "picker", line, "d", "reset")
	checkHint(t, "picker", line, "esc", "cancel")

	field, _ := press(m, runes("n"))
	line = rawLine(t, field.screen(), "Label:")
	if !strings.Contains(plain(line), "enter save • esc cancel") {
		t.Errorf("rename field: want enter save • esc cancel, got %q", plain(line))
	}
	checkHint(t, "rename field", line, "enter", "save")

	group, _ := press(m, runes("m")) // no groups: on + new group…
	line = rawLine(t, group.screen(), "enter new group")
	if !strings.Contains(plain(line), newGroupKeysLine) {
		t.Errorf("group picker: want %q, got %q", newGroupKeysLine, plain(line))
	}
	checkHint(t, "group picker", line, "enter", "new group")
}

// Nothing is drawn in bubbles' own colors: in every state of the menu, each
// colored cell uses one of the theme's five colors. (The picker's swatches
// show other colors on purpose, so they're skipped.)
func TestOnlyThemeColors(t *testing.T) {
	withColor(t)
	allowed := map[string]bool{"": true}
	for _, hex := range []string{nordAccent, nordOnAccent, nordText, nordMuted, nordErr} {
		allowed[fg(hex)], allowed[bg(hex)] = true, true
	}
	keys := func(m Model, ks ...tea.KeyMsg) Model {
		for _, k := range ks {
			m, _ = press(m, k)
		}
		return m
	}
	enterKey := keyMsg("enter")
	groups := groupModel(t).sized(80, 16) // several pages, so the page dots show
	gone := nordModel(t, config.Bookmark{Label: "GONE", Path: "!gone"}, config.Bookmark{Label: "NOTES", Path: "notes"})

	states := map[string]Model{
		"browsing":        groups,
		"next page":       keys(groups, keyMsg("right")),
		"typing a filter": keys(groups, runes("/"), runes("c")),
		"no matches":      keys(groups, runes("/"), runes("zzz")),
		"filter applied":  keys(groups, runes("/"), runes("cs"), enterKey),
		"help page":       keys(groups, runes("?")),
		"tabs hidden":     keys(groups, runes(".")),
		"confirm":         keys(groups, runes("r")),
		"rename field":    keys(groups, runes("n")),
		"group picker":    keys(groups, runes("m")),
		"new group field": keys(groups, runes("m"), keyMsg("up"), enterKey), // ↑ wraps to + new group…
		"ungroup confirm": keys(groups, runes("u")),
		"picker":          keys(groups, runes("t")),
		"hex field":       keys(groups, runes("t"), keyMsg("down"), runes("#")),
		"missing + enter": keys(gone, enterKey),
		"empty":           nordModel(t),
	}
	for name, m := range states {
		for line := range strings.SplitSeq(m.screen(), "\n") {
			for _, c := range cells(line) {
				if c.r == '█' {
					continue // a picker swatch
				}
				if !allowed[c.fg] || !allowed[c.bg] {
					t.Errorf("%s: %q is drawn in fg %q bg %q, not a theme color, in %q", name, string(c.r), c.fg, c.bg, plain(line))
					break
				}
			}
		}
		if slices.Contains([]string{"browsing", "next page"}, name) && !strings.Contains(plain(m.screen()), "•") {
			t.Errorf("%s: want the page dots to show:\n%s", name, plain(m.screen()))
		}
	}
}
