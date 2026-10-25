package ui

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/turtlethom/pathpin/internal/config"
)

// selectLabel moves the cursor to the row with this label on the open tab.
func selectLabel(t *testing.T, m Model, label string) Model {
	t.Helper()
	i := slices.Index(shown(m), label)
	if i < 0 {
		t.Fatalf("%s isn't on this tab: %q", label, shown(m))
	}
	m.list.Select(i)
	return m
}

// groupOnDisk is the saved group of the bookmark with this label.
func groupOnDisk(t *testing.T, m Model, label string) string {
	t.Helper()
	for _, b := range onDisk(t, m) {
		if b.Label == label {
			return b.Group
		}
	}
	t.Fatalf("%s isn't in the file", label)
	return ""
}

func clearField(m Model) Model {
	m, _ = press(m, tea.KeyMsg{Type: tea.KeyCtrlU}) // delete to the start of the field
	return m
}

func enter(m Model) Model {
	m, _ = press(m, tea.KeyMsg{Type: tea.KeyEnter})
	return m
}

func esc(m Model) Model {
	m, _ = press(m, tea.KeyMsg{Type: tea.KeyEsc})
	return m
}

// newGroupRow is the group picker's last row.
const newGroupRow = "+ new group…"

// groupPickerKeysLine is the group picker's bottom line on a group, and
// newGroupKeysLine on + new group…, where enter opens a field.
const (
	groupPickerKeysLine = "↑↓/jk pick • enter move • esc cancel"
	newGroupKeysLine    = "↑↓/jk pick • enter new group • esc cancel"
)

// groupRowAt is the line number of the group picker's row for name (a group,
// or newGroupRow), or -1. A row is its name plus the cursor (›) and the mark
// on the bookmark's current group (•), in any places.
func groupRowAt(m Model, name string) int {
	for i, line := range strings.Split(plain(m.View()), "\n") {
		line = strings.NewReplacer("›", " ", "•", " ").Replace(line)
		if strings.Join(strings.Fields(line), " ") == name {
			return i
		}
	}
	return -1
}

// groupRow is the visible text of the group picker's row for name.
func groupRow(t *testing.T, m Model, name string) string {
	t.Helper()
	i := groupRowAt(m, name)
	if i < 0 {
		t.Fatalf("no %q row in the group picker:\n%s", name, plain(m.View()))
	}
	return strings.Split(plain(m.View()), "\n")[i]
}

// pickerOrder checks that the group picker shows exactly these rows, in this
// order: each one below the last, and nothing else between them.
func pickerOrder(t *testing.T, m Model, names ...string) {
	t.Helper()
	for i, name := range names {
		at := groupRowAt(m, name)
		if at < 0 {
			t.Fatalf("no %q row in the group picker:\n%s", name, plain(m.View()))
		}
		if i > 0 && at != groupRowAt(m, names[i-1])+1 {
			t.Errorf("want %q right under %q:\n%s", name, names[i-1], plain(m.View()))
		}
	}
}

func highlighted(t *testing.T, m Model, name string) bool {
	t.Helper()
	return strings.Contains(groupRow(t, m, name), "›")
}

func current(t *testing.T, m Model, name string) bool {
	t.Helper()
	return strings.Contains(groupRow(t, m, name), "•")
}

// pressM presses m and checks that the group picker opened.
func pressM(t *testing.T, m Model) Model {
	t.Helper()
	m, _ = press(m, runes("m"))
	if m.mode == modeBrowse || groupRowAt(m, newGroupRow) < 0 {
		t.Fatalf("m should open the group picker:\n%s", plain(m.View()))
	}
	return m
}

// toGroupRow highlights the picker's row for name, going down (the picker
// wraps around, so every row is reached).
func toGroupRow(t *testing.T, m Model, name string) Model {
	t.Helper()
	for range 10 {
		if highlighted(t, m, name) {
			return m
		}
		m, _ = press(m, tea.KeyMsg{Type: tea.KeyDown})
	}
	t.Fatalf("couldn't reach %q:\n%s", name, plain(m.View()))
	return m
}

// moveTo moves the highlighted bookmark to an existing group with the picker.
func moveTo(t *testing.T, m Model, group string) Model {
	t.Helper()
	return enter(toGroupRow(t, pressM(t, m), group))
}

// newGroup opens the new-group field from the picker.
func newGroup(t *testing.T, m Model) Model {
	t.Helper()
	m = enter(toGroupRow(t, pressM(t, m), newGroupRow))
	if m.mode == modeBrowse || !strings.Contains(plain(m.View()), "Group:") {
		t.Fatalf("%s should open a Group field:\n%s", newGroupRow, plain(m.View()))
	}
	return m
}

// m opens a picker under the list: a row per group in tab order (each in its
// first spelling: GHOSTTY's "config" is Config), then + new group…, with the
// bookmark's group marked and highlighted. The list stays in view, the
// picker's keys are on the bottom line, and ? lists m.
func TestGroupPickerOpens(t *testing.T) {
	m := selectLabel(t, groupModel(t), "CS310") // School
	if help, _ := press(m, runes("?")); !helpShows(help, "m", "move to group") {
		t.Errorf("the help page (?) should list m move to group:\n%s", plain(help.View()))
	}
	m = pressM(t, m)
	pickerOrder(t, m, "Config", "School", newGroupRow)
	if !current(t, m, "School") || current(t, m, "Config") || current(t, m, newGroupRow) {
		t.Errorf("want only School marked as CS310's group:\n%s", plain(m.View()))
	}
	if !highlighted(t, m, "School") || highlighted(t, m, "Config") {
		t.Errorf("want the cursor on School:\n%s", plain(m.View()))
	}
	if !strings.Contains(plain(m.View()), "CS310") {
		t.Errorf("the list should stay in view above the picker:\n%s", plain(m.View()))
	}
	if !strings.Contains(plain(m.View()), groupPickerKeysLine) {
		t.Errorf("want the picker's keys %q on the bottom line:\n%s", groupPickerKeysLine, plain(m.View()))
	}
	if strings.Contains(plain(m.View()), "q quit") {
		t.Errorf("the browse help line should hide while picking:\n%s", plain(m.View()))
	}
}

// An ungrouped bookmark has no mark, and the cursor starts on the first group.
func TestGroupPickerUngrouped(t *testing.T) {
	m := pressM(t, selectLabel(t, groupModel(t), "PATHPIN"))
	for _, name := range []string{"Config", "School", newGroupRow} {
		if current(t, m, name) {
			t.Errorf("an ungrouped bookmark marks no row, but %q is marked:\n%s", name, plain(m.View()))
		}
	}
	if !highlighted(t, m, "Config") {
		t.Errorf("want the cursor on the first group:\n%s", plain(m.View()))
	}
}

// With no groups yet, the picker is just + new group…, highlighted, so the way
// to make the first group is right there. No bookmarks: m does nothing.
func TestGroupPickerNoGroups(t *testing.T) {
	m := pressM(t, twoBookmarks(t))
	if !highlighted(t, m, newGroupRow) {
		t.Errorf("want %s highlighted:\n%s", newGroupRow, plain(m.View()))
	}
	if !strings.Contains(plain(m.View()), newGroupKeysLine) {
		t.Errorf("on %s, want the keys %q:\n%s", newGroupRow, newGroupKeysLine, plain(m.View()))
	}
	if m, _ := press(newTestModel(t), runes("m")); m.mode != modeBrowse {
		t.Error("m on an empty menu should do nothing")
	}
}

// With a picker open, the menu still fills the window exactly, as it does
// while browsing: the list gets every line the picker doesn't use.
func TestPickersFillWindow(t *testing.T) {
	for name, m := range map[string]Model{"groups": groupModel(t).sized(80, 24), "no groups": twoBookmarks(t).sized(80, 24)} {
		for _, k := range []string{"", "t", "m"} {
			v := m
			if k != "" {
				v, _ = press(m, runes(k))
			}
			if n := len(strings.Split(v.View(), "\n")); n != 24 {
				t.Errorf("%s, %q: the menu is %d lines, want 24:\n%s", name, k, n, plain(v.View()))
			}
		}
	}
}

// ↑↓ and j/k move the cursor, wrapping around at both ends.
func TestGroupPickerCursor(t *testing.T) {
	m := pressM(t, selectLabel(t, groupModel(t), "BASH")) // on Config, the top
	steps := []struct {
		key  tea.KeyMsg
		want string
	}{
		{tea.KeyMsg{Type: tea.KeyUp}, newGroupRow}, // wraps to the bottom
		{runes("j"), "Config"},                     // wraps to the top
		{runes("j"), "School"},
		{tea.KeyMsg{Type: tea.KeyDown}, newGroupRow},
		{runes("k"), "School"},
	}
	for _, s := range steps {
		m, _ = press(m, s.key)
		if !highlighted(t, m, s.want) {
			t.Fatalf("after %s: want %q highlighted:\n%s", s.key, s.want, plain(m.View()))
		}
	}
}

// enter on another group moves the bookmark there, keeping its group's first
// spelling, and the bookmark stays selected with its new tag.
func TestGroupPickerMoves(t *testing.T) {
	m := moveTo(t, selectLabel(t, groupModel(t), "PATHPIN"), "Config")
	if m.mode != modeBrowse {
		t.Fatalf("enter should close the picker:\n%s", plain(m.View()))
	}
	if got := groupOnDisk(t, m, "PATHPIN"); got != "Config" {
		t.Errorf("saved group: got %q, want Config", got)
	}
	if got := selected(m); got != "PATHPIN" {
		t.Errorf("PATHPIN should stay selected, got %q", got)
	}
	if !strings.Contains(plain(m.View()), "Moved PATHPIN to Config") {
		t.Errorf("want a Moved status:\n%s", plain(m.View()))
	}
	if got := row(t, m, "PATHPIN"); !strings.HasSuffix(got, " · Config") {
		t.Errorf("want PATHPIN tagged Config, got %q", got)
	}
}

// enter on the bookmark's own group saves nothing, even for a second spelling
// (GHOSTTY's "config"): every save replaces config.toml.bak, and a save that
// changes nothing would throw away the last real backup.
func TestGroupPickerSameGroup(t *testing.T) {
	for _, label := range []string{"BASH", "GHOSTTY"} {
		m := enter(pressM(t, selectLabel(t, groupModel(t), label)))
		if m.mode != modeBrowse {
			t.Fatalf("%s: enter should close the picker:\n%s", label, plain(m.View()))
		}
		if _, err := os.Stat(m.configPath + ".bak"); err == nil {
			t.Errorf("%s: an unchanged group should not save (a backup was written)", label)
		}
		if !strings.Contains(plain(m.View()), label+" stays in Config") {
			t.Errorf("%s: want a stays-in status:\n%s", label, plain(m.View()))
		}
	}
}

// esc closes the picker without saving. In the picker, other keys don't act:
// q doesn't quit, r doesn't ask to remove.
func TestGroupPickerCancel(t *testing.T) {
	m := pressM(t, selectLabel(t, groupModel(t), "BASH"))
	m, _ = press(m, tea.KeyMsg{Type: tea.KeyDown}) // School
	for _, k := range []string{"q", "r", "c", "a"} {
		var cmd tea.Cmd
		m, cmd = press(m, runes(k))
		if isQuit(cmd) || m.mode == modeBrowse || m.mode == modeConfirm {
			t.Fatalf("%s in the picker should do nothing:\n%s", k, plain(m.View()))
		}
	}
	m = esc(m)
	if m.mode != modeBrowse || groupRowAt(m, newGroupRow) >= 0 {
		t.Fatalf("esc should close the picker:\n%s", plain(m.View()))
	}
	if _, err := os.Stat(m.configPath + ".bak"); err == nil || groupOnDisk(t, m, "BASH") != "Config" || len(onDisk(t, m)) != 6 {
		t.Error("esc should close the picker without saving")
	}
}

// While a filter is being typed, m is a letter in it.
func TestMWhileFilteringIsTyped(t *testing.T) {
	m, _ := press(groupModel(t), runes("/"))
	m, _ = press(m, runes("m"))
	if m.mode != modeBrowse || m.list.FilterValue() != "m" {
		t.Errorf("want m typed into the filter, got mode %v, filter %q", m.mode, m.list.FilterValue())
	}
}

// + new group… opens an empty Group field; enter makes the group with the
// bookmark in it, and its tab appears before Other.
func TestNewGroup(t *testing.T) {
	m := newGroup(t, selectLabel(t, groupModel(t), "PATHPIN"))
	m, _ = press(m, runes("Tools"))
	m = enter(m)
	if m.mode != modeBrowse {
		t.Fatalf("enter should close the field:\n%s", plain(m.View()))
	}
	if got := groupOnDisk(t, m, "PATHPIN"); got != "Tools" {
		t.Errorf("saved group: got %q, want Tools", got)
	}
	if got, want := tabBar(t, m), []string{"All", "Config", "School", "Tools", "Other"}; !slices.Equal(got, want) {
		t.Errorf("tab bar: got %q, want %q", got, want)
	}
	if got := selected(m); got != "PATHPIN" {
		t.Errorf("PATHPIN should stay selected, got %q", got)
	}
	if !strings.Contains(plain(m.View()), "Moved PATHPIN to Tools") {
		t.Errorf("want a Moved status:\n%s", plain(m.View()))
	}
}

// The first group works the same way: the tab bar appears.
func TestNewGroupFirst(t *testing.T) {
	m := newGroup(t, twoBookmarks(t)) // NVIM highlighted
	m, _ = press(m, runes("Config"))
	m = enter(m)
	if got, want := tabBar(t, m), []string{"All", "Config", "Other"}; !slices.Equal(got, want) {
		t.Errorf("tab bar: got %q, want %q", got, want)
	}
}

// The field makes new groups only: an existing group's name (in any case),
// a tab name, and an empty or blank name are refused with the field kept
// open. Nothing is saved.
func TestNewGroupRefuses(t *testing.T) {
	for typed, want := range map[string]string{
		"school": "School already exists",
		"CONFIG": "Config already exists",
		"All":    "tab names",
		"other":  "tab names",
		"":       "",
		"   ":    "",
	} {
		m := newGroup(t, selectLabel(t, groupModel(t), "PATHPIN"))
		if typed != "" {
			m, _ = press(m, runes(typed))
		}
		m = enter(m)
		if m.mode == modeBrowse || !strings.Contains(plain(m.View()), "Group:") {
			t.Errorf("%q: the field should stay open:\n%s", typed, plain(m.View()))
		}
		if want != "" && !strings.Contains(plain(m.View()), want) {
			t.Errorf("%q: want %q:\n%s", typed, want, plain(m.View()))
		}
		if _, err := os.Stat(m.configPath + ".bak"); err == nil || groupOnDisk(t, m, "PATHPIN") != "" {
			t.Errorf("%q: a refused name must not be saved", typed)
		}
	}
}

// esc in the field goes back to the picker, still on + new group…; esc again
// closes it. In the field, keys are letters: j, k, r, q don't act.
func TestNewGroupEscAndTyping(t *testing.T) {
	m := newGroup(t, selectLabel(t, groupModel(t), "PATHPIN"))
	m, _ = press(m, runes("jkrq"))
	if m.mode == modeBrowse || !strings.Contains(plain(m.View()), "Group: jkrq") {
		t.Errorf("want j, k, r, q typed into the field:\n%s", plain(m.View()))
	}
	m = esc(m)
	if m.mode == modeBrowse || strings.Contains(plain(m.View()), "Group:") || !highlighted(t, m, newGroupRow) {
		t.Fatalf("esc should go back to the picker, on %s:\n%s", newGroupRow, plain(m.View()))
	}
	m = esc(m)
	if m.mode != modeBrowse {
		t.Errorf("esc again should close the picker:\n%s", plain(m.View()))
	}
	if _, err := os.Stat(m.configPath + ".bak"); err == nil {
		t.Error("nothing should be saved")
	}

	// A refused name's error belongs to the field: esc takes it away, and
	// the picker shows its keys again.
	m = newGroup(t, selectLabel(t, groupModel(t), "PATHPIN"))
	m, _ = press(m, runes("school"))
	m = esc(enter(m))
	if v := plain(m.View()); strings.Contains(v, "already exists") || !strings.Contains(v, newGroupKeysLine) {
		t.Errorf("after esc, want the picker's keys, not the field's error:\n%s", v)
	}
}

// The field's keys are on its line, themed like every key hint.
func TestNewGroupFieldKeys(t *testing.T) {
	withColor(t)
	m := newGroup(t, selectLabel(t, groupModel(t), "PATHPIN"))
	line := rawLine(t, m.View(), "Group:")
	if !strings.Contains(plain(line), "enter save • esc back") {
		t.Errorf("want enter save • esc back on the field's line, got %q", plain(line))
	}
	checkHint(t, "new group field", line, "esc", "back")
}

// On a group's tab, a bookmark moved to another group leaves the tab; if the
// tab empties, the menu shows All.
func TestMoveOffGroupTab(t *testing.T) {
	m := selectLabel(t, schoolTab(t), "CS310") // School: CS310 NOTES
	m = moveTo(t, m, "Config")
	if got := shown(m); !slices.Equal(got, []string{"NOTES"}) {
		t.Errorf("School after moving CS310 away: got %q", got)
	}
	m = moveTo(t, m, "Config") // NOTES, the last in School
	if got := activeTab(t, m); got != "All" {
		t.Errorf("want All once School is empty, got %s", got)
	}
	if got, want := tabBar(t, m), []string{"All", "Config", "Other"}; !slices.Equal(got, want) {
		t.Errorf("tab bar: got %q, want %q", got, want)
	}
}

// Moving keeps every other setting in the file, like add and remove do.
func TestMoveKeepsSettings(t *testing.T) {
	m := selectLabel(t, groupModel(t), "PATHPIN")
	m.cfg.Menu.Icons = "nerd"
	m = moveTo(t, m, "School")
	cfg, err := config.Load(m.configPath)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Menu.Icons == "" || cfg.Colors.Theme != "nord" {
		t.Errorf("moving lost settings: menu %+v, colors %+v", cfg.Menu, cfg.Colors)
	}
}

// moreAbove and moreBelow match the group picker's scroll lines.
var (
	moreAbove = regexp.MustCompile(`↑ \d+ more`)
	moreBelow = regexp.MustCompile(`↓ \d+ more`)
)

// manyGroups is a menu with 25 groups of one bookmark each (B00 in G00, …),
// in an 80x24 window: more groups than the picker has room for.
func manyGroups(t *testing.T) Model {
	t.Helper()
	var bookmarks []config.Bookmark
	for i := range 25 {
		bookmarks = append(bookmarks, config.Bookmark{Label: fmt.Sprintf("B%02d", i), Path: fmt.Sprintf("b%02d", i), Group: fmt.Sprintf("G%02d", i)})
	}
	return groupModelOf(t, bookmarks...).sized(80, 24)
}

// With more groups than fit, the menu still fits the window with its title
// and the highlighted bookmark in view. The groups scroll to keep the cursor
// in view, with lines saying how many are hidden; + new group… always shows.
func TestGroupPickerScrolls(t *testing.T) {
	m := pressM(t, manyGroups(t)) // B00, in G00
	check := func(step string) {
		t.Helper()
		v := plain(m.View())
		if n := len(strings.Split(m.View(), "\n")); n > 24 {
			t.Fatalf("%s: the menu is %d lines, want at most 24:\n%s", step, n, v)
		}
		if !strings.Contains(v, "PATHPIN") || !strings.Contains(v, "B00") {
			t.Fatalf("%s: want the title and B00 in view:\n%s", step, v)
		}
		if groupRowAt(m, newGroupRow) < 0 || !strings.Contains(v, "›") {
			t.Fatalf("%s: want + new group… and the cursor in view:\n%s", step, v)
		}
	}
	check("opened")
	if v := plain(m.View()); !moreBelow.MatchString(v) || moreAbove.MatchString(v) {
		t.Errorf("opened at the top: want only a ↓ more line:\n%s", v)
	}
	if groupRowAt(m, "G24") >= 0 {
		t.Errorf("G24 shouldn't fit at first:\n%s", plain(m.View()))
	}
	for i := 1; i < 25; i++ {
		m, _ = press(m, runes("j"))
		check(fmt.Sprintf("down to G%02d", i))
		if name := fmt.Sprintf("G%02d", i); !highlighted(t, m, name) {
			t.Fatalf("want %s highlighted:\n%s", name, plain(m.View()))
		}
	}
	if v := plain(m.View()); !moreAbove.MatchString(v) || moreBelow.MatchString(v) {
		t.Errorf("at the bottom: want only an ↑ more line:\n%s", v)
	}
	m, _ = press(m, runes("j")) // + new group…
	m, _ = press(m, runes("j")) // wraps to G00
	check("wrapped")
	if !highlighted(t, m, "G00") {
		t.Errorf("want G00 highlighted after wrapping:\n%s", plain(m.View()))
	}

	// Opened on a bookmark far down, the picker starts scrolled to its group.
	m = pressM(t, selectLabel(t, manyGroups(t), "B20"))
	if !highlighted(t, m, "G20") || !current(t, m, "G20") {
		t.Errorf("want G20 in view, marked and highlighted:\n%s", plain(m.View()))
	}
}
