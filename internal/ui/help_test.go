package ui

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// helpFooter is the help page's last line: the keys that close it.
const helpFooter = "? esc q close"

// helpBlock is one titled block of the help page: its title and its rows,
// each a key and what it does, top to bottom.
type helpBlock struct {
	title string
	rows  [][2]string
}

// helpBlocks are the help page's blocks, in reading order: three to a row
// at 80 columns.
var helpBlocks = []helpBlock{
	{"Navigate", [][2]string{{"↑↓/jk", "move"}, {"←→/hl", "page"}, {"g/G", "top/end"}, {"tab/⇧tab", "group"}, {".", "groups only"}}},
	{"Bookmarks", [][2]string{{"a", "add"}, {"r", "remove"}, {"c", "clear all"}, {"J/K", "reorder"}}},
	{"Names", [][2]string{{"n", "rename"}, {"N", "rename group"}}},
	{"Groups", [][2]string{{"m", "move to group"}, {"u", "ungroup"}}},
	{"Actions", [][2]string{{"enter", "cd"}, {"e", "edit"}, {"y", "copy path"}, {"o", "open"}}},
	{"Menu", [][2]string{{"/", "filter"}, {"t", "appearance"}, {"?", "help"}, {"q", "quit"}}},
}

// openHelp presses ? and checks that the help page opened.
func openHelp(t *testing.T, m Model) Model {
	t.Helper()
	m, _ = press(m, runes("?"))
	if !strings.Contains(plain(m.screen()), helpFooter) {
		t.Fatalf("? should open the help page:\n%s", plain(m.screen()))
	}
	return m
}

// pageBlock is a block as found on the page.
type pageBlock struct {
	line, col int         // where its title is
	rows      [][2]string // key and description, as shown
	descCols  []int       // the column each description starts in
}

// findBlock finds the block with this title on the page: its title, a ─ rule
// of the title's length right under it, then its rows, read from the title's
// column up to the next block's title on that line, until a blank one (or
// the … and footer of a cut page).
func findBlock(t *testing.T, view, title string) pageBlock {
	t.Helper()
	lines := strings.Split(plain(view), "\n")
	for i, l := range lines {
		col := wordAt(l, title)
		if col < 0 {
			continue
		}
		end := 0 // the page's widest line, unless another title comes first
		for _, row := range lines {
			end = max(end, len([]rune(row)))
		}
		for _, other := range helpBlocks {
			if c := wordAt(l, other.title); c > col && c < end {
				end = c
			}
		}
		b := pageBlock{line: i, col: col}
		rule := strings.Repeat("─", len([]rune(title)))
		if i+1 >= len(lines) || cut(lines[i+1], col, col+len([]rune(rule))) != rule {
			t.Fatalf("%s: want a %d-column rule right under the title:\n%s", title, len([]rune(rule)), plain(view))
		}
		for _, row := range lines[i+2:] {
			seg := cut(row, col, end)
			text := strings.TrimSpace(seg)
			if text == "" || text == "…" || strings.HasPrefix(helpFooter, text) {
				break // the blank line after the block, or a cut page's last lines
			}
			key, desc, _ := strings.Cut(text, " ")
			desc = strings.TrimSpace(desc)
			b.rows = append(b.rows, [2]string{key, desc})
			b.descCols = append(b.descCols, col+len([]rune(seg[:strings.Index(seg, desc)])))
		}
		return b
	}
	t.Fatalf("no %s block on the page:\n%s", title, plain(view))
	return pageBlock{}
}

// wordAt is the column where word starts in line as a whole word, or -1.
func wordAt(line, word string) int {
	rs, ws := []rune(line), []rune(word)
	for i := 0; i+len(ws) <= len(rs); i++ {
		if string(rs[i:i+len(ws)]) != word {
			continue
		}
		before := i == 0 || rs[i-1] == ' '
		after := i+len(ws) == len(rs) || rs[i+len(ws)] == ' '
		if before && after {
			return i
		}
	}
	return -1
}

// cut is the text of line from column from up to column to (runes, which
// are columns for everything the page draws), padded if the line is short.
func cut(line string, from, to int) string {
	rs := []rune(line)
	for len(rs) < to {
		rs = append(rs, ' ')
	}
	return string(rs[from:to])
}

// checkFits checks that no line of the view is wider than width, and that
// it has at most height lines.
func checkFits(t *testing.T, name string, m Model, width, height int) {
	t.Helper()
	lines := strings.Split(m.screen(), "\n")
	if len(lines) > height {
		t.Errorf("%s: %d lines, want at most %d:\n%s", name, len(lines), height, plain(m.screen()))
	}
	for _, l := range lines {
		if w := lipgloss.Width(l); w > width {
			t.Errorf("%s: a line is %d wide, want at most %d: %q", name, w, width, plain(l))
		}
	}
}

// ? opens a page of its own: the badge, the blocks and the footer, with the
// list, tabs and help line out of view. It fits the window.
func TestHelpPageOpens(t *testing.T) {
	m := openHelp(t, groupModel(t).sized(80, 24))
	v := plain(m.screen())
	// The rows, the status, the help line and the tab bar (Other is a tab).
	for _, gone := range []string{"~/bash", "6 bookmarks", "↑/k up", "Other"} {
		if strings.Contains(v, gone) {
			t.Errorf("the page should hide the menu, but shows %q:\n%s", gone, v)
		}
	}
	for _, line := range strings.Split(v, "\n") {
		if s := strings.TrimSpace(line); s != "" {
			if s != "PATHPIN" {
				t.Errorf("want the PATHPIN badge first, got %q", s)
			}
			break
		}
	}
	lines := strings.Split(strings.TrimRight(v, " \n"), "\n")
	if last := strings.TrimSpace(lines[len(lines)-1]); last != helpFooter {
		t.Errorf("want %q as the last line, got %q", helpFooter, last)
	}
	checkFits(t, "80x24", m, 80, 24)
}

// Each block has its rows in order, and its descriptions line up. Blocks for
// keys that need groups (tab, N) show without groups too, so the page never
// changes shape and shows that groups exist.
func TestHelpPageBlocks(t *testing.T) {
	for name, m := range map[string]Model{"groups": groupModel(t), "no groups": twoBookmarks(t)} {
		m = openHelp(t, m.sized(80, 24))
		for _, want := range helpBlocks {
			b := findBlock(t, m.screen(), want.title)
			if !slices.Equal(b.rows, want.rows) {
				t.Errorf("%s, %s: rows %q, want %q", name, want.title, b.rows, want.rows)
			}
			for _, c := range b.descCols {
				if c != b.descCols[0] {
					t.Errorf("%s, %s: descriptions start in columns %v, want one column", name, want.title, b.descCols)
					break
				}
			}
		}
	}
}

// At 80 columns the blocks are three to a row, in reading order: Navigate,
// Bookmarks, Names, then Groups, Actions, Menu. Narrower, two to a row, then
// one; every line fits the window.
func TestHelpPageReflows(t *testing.T) {
	for _, tc := range []struct {
		width   int
		perLine int
	}{{80, 3}, {50, 2}, {30, 1}} {
		m := openHelp(t, groupModel(t).sized(tc.width, 60)) // tall enough for one per row
		checkFits(t, "reflow", m, tc.width, 60)
		var rows [][]string // titles by line, in order
		at := -1
		for _, want := range helpBlocks {
			b := findBlock(t, m.screen(), want.title)
			if b.line != at {
				rows, at = append(rows, nil), b.line
			}
			rows[len(rows)-1] = append(rows[len(rows)-1], want.title)
		}
		for i, r := range rows {
			if want := min(tc.perLine, len(helpBlocks)-i*tc.perLine); len(r) != want {
				t.Errorf("%d columns: want %d blocks to a row, got %q", tc.width, tc.perLine, rows)
				break
			}
		}
	}
}

// A page taller than the window is cut, with a … line before the footer.
func TestHelpPageCut(t *testing.T) {
	for _, size := range [][2]int{{80, 12}, {30, 24}} {
		m := openHelp(t, groupModel(t).sized(size[0], size[1]))
		checkFits(t, "cut", m, size[0], size[1])
		v := plain(m.screen())
		lines := strings.Split(strings.TrimRight(v, " \n"), "\n")
		if len(lines) < 2 || strings.TrimSpace(lines[len(lines)-2]) != "…" || strings.TrimSpace(lines[len(lines)-1]) != helpFooter {
			t.Errorf("%dx%d: want the page cut with … above the footer:\n%s", size[0], size[1], v)
		}
	}
}

// ?, esc and q close the page (q doesn't quit), back to the menu as it was,
// filling the window. ctrl+c still quits. Other keys do nothing: enter
// doesn't cd, r doesn't ask, j doesn't move the cursor.
func TestHelpPageCloses(t *testing.T) {
	for _, k := range []tea.KeyMsg{runes("?"), keyMsg("esc"), runes("q")} {
		m := openHelp(t, groupModel(t).sized(80, 24))
		m, cmd := press(m, k)
		if isQuit(cmd) || m.mode != modeBrowse || strings.Contains(plain(m.screen()), helpFooter) {
			t.Errorf("%s should close the page without quitting:\n%s", k, plain(m.screen()))
		}
		if n := len(strings.Split(m.screen(), "\n")); n != 24 || !strings.Contains(plain(m.screen()), "~/bash") {
			t.Errorf("%s: want the menu back, 24 lines, got %d:\n%s", k, n, plain(m.screen()))
		}
	}

	m := openHelp(t, groupModel(t).sized(80, 24))
	for _, k := range []tea.KeyMsg{keyMsg("enter"), runes("r"), runes("j"), runes("a"), runes("m")} {
		var cmd tea.Cmd
		m, cmd = press(m, k)
		if isQuit(cmd) || m.chosen != "" || !strings.Contains(plain(m.screen()), helpFooter) {
			t.Fatalf("%s on the page should do nothing:\n%s", k, plain(m.screen()))
		}
	}
	if m, _ = press(m, runes("?")); selected(m) != "BASH" {
		t.Errorf("keys on the page shouldn't move the cursor, got %s", selected(m))
	}
	if _, cmd := press(openHelp(t, groupModel(t)), keyMsg("ctrl+c")); !isQuit(cmd) {
		t.Error("ctrl+c should quit from the page")
	}
}

// While a filter is typed, ? is a letter in it.
func TestQuestionMarkWhileFilteringIsTyped(t *testing.T) {
	m, _ := press(groupModel(t), runes("/"))
	m, _ = press(m, runes("?"))
	if m.list.FilterValue() != "?" || strings.Contains(plain(m.screen()), helpFooter) {
		t.Errorf("want ? typed into the filter, got filter %q:\n%s", m.list.FilterValue(), plain(m.screen()))
	}
}

// With a filter applied, the Menu block lists esc clear filter under
// / filter. esc closes the page; the filter stays applied.
func TestHelpPageFilterApplied(t *testing.T) {
	m := groupModel(t).sized(80, 24)
	for _, k := range []tea.KeyMsg{runes("/"), runes("cs"), keyMsg("enter")} {
		m, _ = press(m, k)
	}
	filtered := shown(m)
	m = openHelp(t, m)
	want := [][2]string{{"/", "filter"}, {"esc", "clear filter"}, {"t", "appearance"}, {"?", "help"}, {"q", "quit"}}
	if b := findBlock(t, m.screen(), "Menu"); !slices.Equal(b.rows, want) {
		t.Errorf("Menu with a filter applied: rows %q, want %q", b.rows, want)
	}
	m, _ = press(m, keyMsg("esc"))
	if strings.Contains(plain(m.screen()), helpFooter) || !slices.Equal(shown(m), filtered) {
		t.Errorf("esc should close the page and keep the filter (%q), got %q", filtered, shown(m))
	}
}

// Colors: titles in the text color, their rules muted, keys in the accent,
// descriptions muted, and the … and footer like every key hint.
func TestHelpPageColors(t *testing.T) {
	withColor(t)
	m := openHelp(t, groupModel(t).sized(80, 24))
	view := m.screen()

	_, title, _ := find(t, cells(rawLine(t, view, "Navigate")), "Navigate")
	if !allFg(title, fg(nordText)) {
		t.Error("a block title should be in the text color")
	}
	_, rule, _ := find(t, cells(rawLine(t, view, "────────")), "────────")
	if !allFg(rule, fg(nordMuted)) {
		t.Error("a title's rule should be muted")
	}
	row := cells(rawLine(t, view, "move to group"))
	_, key, _ := find(t, row, "enter")
	_, desc, _ := find(t, row, "move to group")
	if !allFg(key, fg(nordAccent)) || !allFg(desc, fg(nordMuted)) {
		t.Error("keys should be in the accent and descriptions muted")
	}
	footer := cells(rawLine(t, view, "close"))
	_, k, _ := find(t, footer, "esc")
	_, d, _ := find(t, footer, "close")
	if !allFg(k, fg(nordAccent)) || !allFg(d, fg(nordMuted)) {
		t.Error("the footer should look like a key hint")
	}

	short := openHelp(t, groupModel(t).sized(80, 12))
	_, dots, _ := find(t, cells(rawLine(t, short.screen(), "…")), "…")
	if !allFg(dots, fg(nordMuted)) {
		t.Error("the … of a cut page should be muted")
	}
}

// The short help line says ? help. Without groups, where tab does nothing,
// m group takes tab group's place, so the group key shows to whoever hasn't
// found it.
func TestShortLine(t *testing.T) {
	for name, tc := range map[string]struct {
		m        Model
		has, not string
	}{
		"groups":    {groupModel(t), "tab group", "m group"},
		"no groups": {twoBookmarks(t), "m group", "tab group"},
	} {
		line := helpLine(t, tc.m.sized(80, 24))
		if !strings.Contains(line, tc.has) || strings.Contains(line, tc.not) {
			t.Errorf("%s: want %q and not %q in %q", name, tc.has, tc.not, line)
		}
		if !strings.HasSuffix(line, "q quit • ? help") {
			t.Errorf("%s: want the line to end with q quit • ? help, got %q", name, line)
		}
	}
}

// At any height from the smallest that holds the badge, … and the footer,
// the page fits the window and is cut between rows of blocks: a block that
// shows has all its rows.
func TestHelpPageCutsWholeRows(t *testing.T) {
	for _, width := range []int{80, 30} {
		for height := 5; height <= 30; height++ {
			m := openHelp(t, groupModel(t).sized(width, height))
			checkFits(t, fmt.Sprintf("%dx%d", width, height), m, width, height)
			for _, want := range helpBlocks {
				if !slices.ContainsFunc(strings.Split(plain(m.screen()), "\n"), func(l string) bool { return wordAt(l, want.title) >= 0 }) {
					continue // cut off whole
				}
				if b := findBlock(t, m.screen(), want.title); !slices.Equal(b.rows, want.rows) {
					t.Errorf("%dx%d: %s shows rows %q, want all of %q", width, height, want.title, b.rows, want.rows)
				}
			}
		}
	}
}

// The page doesn't show the bottom line, so an action's error that arrives
// while it's open (o with no file manager) waits for it to close, instead of
// being cleared by the key that closes it.
func TestHelpPageKeepsMessage(t *testing.T) {
	m := openHelp(t, groupModel(t).sized(80, 24))
	m = deliver(m, openedMsg{path: "/x", err: errors.New("no file manager found")})
	m, _ = press(m, keyMsg("esc"))
	if !strings.Contains(plain(m.screen()), "no file manager found") {
		t.Errorf("want the error once the page closes:\n%s", plain(m.screen()))
	}
}
