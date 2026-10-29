package ui

import (
	"os"
	"slices"
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"

	"github.com/turtlethom/pathpin/v2/internal/config"
)

// pressDot presses . (hide or show All and Other).
func pressDot(m Model) Model {
	m, _ = press(m, runes("."))
	return m
}

// statusText is the list's status line (the bookmark count), trimmed.
func statusText(t *testing.T, m Model) string {
	t.Helper()
	lines := strings.Split(plain(m.screen()), "\n")
	title := slices.IndexFunc(lines, func(l string) bool { return strings.Contains(l, "PATHPIN") })
	i := statusLine(lines, title)
	if i < 0 {
		t.Fatalf("no status line:\n%s", plain(m.screen()))
	}
	return strings.TrimSpace(lines[i])
}

// startedWith builds groupModel's menu from a config file that also has
// these [menu] lines, as pathpin does at start.
func startedWith(t *testing.T, menu string) Model {
	t.Helper()
	return restartedWith(t, groupModel(t), menu)
}

// restartedWith builds m's menu again from its config file plus these
// [menu] lines, in nord (like the test models), at 80x40.
func restartedWith(t *testing.T, m Model, menu string) Model {
	t.Helper()
	f, err := os.OpenFile(m.configPath, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("\n[menu]\n" + menu + "\n"); err != nil {
		t.Fatal(err)
	}
	f.Close()
	cfg, err := config.Load(m.configPath)
	if err != nil {
		t.Fatalf("load [menu] %s: %v", menu, err)
	}
	cfg.Colors.Theme = "nord" // like groupModel's, so activeTab finds the open tab's color
	return started(New(cfg, m.configPath)).sized(80, 40)
}

// . hides the All and Other tabs, so the tab bar shows only the groups; the
// menu moves from All to the first group. After the tabs, a note counts
// what's out of view (the ungrouped PATHPIN and CSSLAB); the status line
// counts the open tab's rows as usual. Nothing is saved.
func TestDotHidesAllAndOther(t *testing.T) {
	m := pressDot(groupModel(t))
	if got, want := tabBar(t, m), []string{"Config", "School"}; !slices.Equal(got, want) {
		t.Errorf("tab bar: got %q, want %q", got, want)
	}
	if got := activeTab(t, m); got != "Config" {
		t.Errorf("want Config open, got %s", got)
	}
	if got, want := shown(m), []string{"BASH", "GHOSTTY"}; !slices.Equal(got, want) {
		t.Errorf("rows: got %q, want %q", got, want)
	}
	if got, want := hiddenNote(t, m), "2 hidden (.)"; got != want {
		t.Errorf("note after the tabs: got %q, want %q", got, want)
	}
	if got, want := statusText(t, m), "2 bookmarks"; got != want {
		t.Errorf("status: got %q, want %q", got, want)
	}
	if _, err := os.Stat(m.configPath + ".bak"); err == nil {
		t.Error(". should not save (a backup was written)")
	}
}

// . again shows All and Other, and the open tab stays open.
func TestDotShowsThemAgain(t *testing.T) {
	m := pressDot(pressTab(pressDot(groupModel(t)))) // hide, School, show
	if got, want := tabBar(t, m), []string{"All", "Config", "School", "Other"}; !slices.Equal(got, want) {
		t.Errorf("tab bar: got %q, want %q", got, want)
	}
	if got := activeTab(t, m); got != "School" {
		t.Errorf("want School still open, got %s", got)
	}
	if got := hiddenNote(t, m); got != "" {
		t.Errorf("nothing is hidden now, but the tab bar says %q", got)
	}
}

// Hiding keeps a group's tab open; from Other, the menu moves to the first
// group, as from All.
func TestDotFromEachTab(t *testing.T) {
	if got := activeTab(t, pressDot(schoolTab(t))); got != "School" {
		t.Errorf("from School: want School open, got %s", got)
	}
	if got := activeTab(t, pressDot(pressShiftTab(groupModel(t)))); got != "Config" {
		t.Errorf("from Other: want Config open, got %s", got)
	}
}

// While hidden, tab and shift+tab go round the groups only.
func TestTabWhileHidden(t *testing.T) {
	m := pressDot(groupModel(t)) // Config
	for _, want := range []string{"School", "Config"} {
		if m = pressTab(m); activeTab(t, m) != want {
			t.Errorf("tab: want %s, got %s", want, activeTab(t, m))
		}
	}
	if m = pressShiftTab(m); activeTab(t, m) != "School" {
		t.Errorf("shift+tab: want School (wrapping), got %s", activeTab(t, m))
	}
}

// With every bookmark in a group there's no Other: . hides All, and nothing
// is out of view, so the tab bar doesn't mention it.
func TestDotAllGrouped(t *testing.T) {
	m := pressDot(groupModelOf(t,
		config.Bookmark{Label: "BASH", Path: "bash", Group: "Config"},
		config.Bookmark{Label: "CS310", Path: "cs310", Group: "School"},
	))
	if got, want := tabBar(t, m), []string{"Config", "School"}; !slices.Equal(got, want) {
		t.Errorf("tab bar: got %q, want %q", got, want)
	}
	if got := hiddenNote(t, m); got != "" {
		t.Errorf("nothing is out of view, but the tab bar says %q", got)
	}
}

// Without groups there's nothing to show instead: . says so and changes
// nothing. On an empty menu it does nothing.
func TestDotWithoutGroups(t *testing.T) {
	m := pressDot(twoBookmarks(t))
	if !strings.Contains(plain(m.screen()), "No groups to show") || len(shown(m)) != 2 || tabBar(t, m) != nil {
		t.Errorf("want a note and the menu as it was:\n%s", plain(m.screen()))
	}
	if m := pressDot(newTestModel(t)); m.mode != modeBrowse || !strings.Contains(plain(m.screen()), emptyMessage) {
		t.Errorf(". on an empty menu should do nothing:\n%s", plain(m.screen()))
	}
}

// While a filter is being typed, . is a character in it.
func TestDotWhileFilteringIsTyped(t *testing.T) {
	m, _ := press(groupModel(t), runes("/"))
	m = pressDot(m)
	if m.list.FilterValue() != "." || m.tab != config.TabAll {
		t.Errorf("want . typed into the filter and All still open, got filter %q on %s", m.list.FilterValue(), m.tab)
	}
}

// Changes while hidden keep the count right: ungrouping a bookmark sends it
// out of view.
func TestHiddenCountFollowsChanges(t *testing.T) {
	m := pressDot(groupModel(t)) // Config: BASH GHOSTTY; PATHPIN, CSSLAB hidden
	m, _ = press(m, runes("u"))  // BASH
	m, _ = press(m, runes("y"))
	if got, want := shown(m), []string{"GHOSTTY"}; !slices.Equal(got, want) {
		t.Errorf("rows: got %q, want %q", got, want)
	}
	if got, want := hiddenNote(t, m), "3 hidden (.)"; got != want {
		t.Errorf("note: got %q, want %q", got, want)
	}
}

// When the last group goes, there's nothing left to hide behind: every
// bookmark shows, as in a menu without groups, and . says there are no groups.
func TestLastGroupGoneShowsAll(t *testing.T) {
	m := pressDot(soloTab(t)) // Solo: SOLO; NOTES hidden
	m, _ = press(m, runes("u"))
	m, _ = press(m, runes("y"))
	if tabBar(t, m) != nil || !slices.Equal(shown(m), []string{"SOLO", "NOTES"}) {
		t.Errorf("want every bookmark and no tab bar:\n%s", plain(m.screen()))
	}
	if got := statusText(t, m); strings.Contains(got, "hidden") {
		t.Errorf("nothing is hidden now, but the status says %q", got)
	}
}

// The note is left out when it doesn't fit after the tabs, so the tab bar
// never runs past the list's width.
func TestHiddenNoteFits(t *testing.T) {
	m := pressDot(groupModel(t).sized(26, 24))
	if got := hiddenNote(t, m); got != "" {
		t.Errorf("26 columns: want no note, got %q", got)
	}
	checkFits(t, "26 columns", m, 26, 24)
}

// With a single group, hiding leaves one tab: tab has nowhere to go, so it
// keeps an applied filter and the cursor.
func TestTabWithOneTab(t *testing.T) {
	m := pressDot(soloTab(t))
	for _, k := range []tea.KeyMsg{runes("/"), runes("so"), keyMsg("enter")} {
		m, _ = press(m, k)
	}
	m = pressTab(m)
	if m.list.FilterState() != list.FilterApplied || m.list.FilterValue() != "so" {
		t.Errorf("tab with one tab should keep the filter, got %v %q", m.list.FilterState(), m.list.FilterValue())
	}
}

// Ungrouping a whole group with N while hidden opens the first group left,
// and the ungrouped bookmarks join the hidden count.
func TestUngroupGroupWhileHidden(t *testing.T) {
	m := pressDot(schoolTab(t))
	m = enter(clearField(pressN(m)))
	m, _ = press(m, runes("y"))
	if got := activeTab(t, m); got != "Config" {
		t.Errorf("want Config open, got %s", got)
	}
	if got, want := tabBar(t, m), []string{"Config"}; !slices.Equal(got, want) {
		t.Errorf("tab bar: got %q, want %q", got, want)
	}
	if got, want := hiddenNote(t, m), "4 hidden (.)"; got != want {
		t.Errorf("note: got %q, want %q", got, want)
	}
}

// [menu] tabs = "groups" starts the menu with All and Other hidden, on the
// first group; . shows them. Without groups it changes nothing.
func TestStartHidden(t *testing.T) {
	m := startedWith(t, `tabs = "groups"`)
	if got, want := tabBar(t, m), []string{"Config", "School"}; !slices.Equal(got, want) {
		t.Errorf("tab bar at start: got %q, want %q", got, want)
	}
	if got := activeTab(t, m); got != "Config" {
		t.Errorf("want Config open at start, got %s", got)
	}
	if got, want := tabBar(t, pressDot(m)), []string{"All", "Config", "School", "Other"}; !slices.Equal(got, want) {
		t.Errorf(". should show All and Other: got %q, want %q", got, want)
	}

	if m := startedWith(t, `tabs = "all"`); len(tabBar(t, m)) != 4 || activeTab(t, m) != "All" {
		t.Errorf(`tabs = "all" should start as usual, got %q on %s`, tabBar(t, m), activeTab(t, m))
	}
}

// The t picker's tabs row sets how the menu starts. It doesn't preview: .
// is the live switch, so the menu behind keeps its tabs. esc saves nothing;
// enter saves [menu] tabs.
func TestTabsRow(t *testing.T) {
	m := openPicker(t, groupModel(t).sized(80, 30))
	m = toRow(t, m, "tabs")
	if line := pickerLine(t, m, "tabs"); !strings.Contains(line, "all") {
		t.Errorf("tabs row should start at all, got %q", line)
	}
	m = right(m)
	if line := pickerLine(t, m, "tabs"); !strings.Contains(line, "groups") {
		t.Errorf("→ should choose groups, got %q", line)
	}
	if got := tabBar(t, m); len(got) != 4 {
		t.Errorf("the row shouldn't change the menu's tabs, got %q", got)
	}
	if cancelled, _ := press(m, keyMsg("esc")); saved(cancelled) {
		t.Error("esc should save nothing")
	}
	m = enter(m)
	if !saved(m) || !strings.Contains(readConfig(t, m), `tabs = "groups"`) {
		t.Errorf("enter should save tabs = \"groups\":\n%s", readConfig(t, m))
	}
	if got := tabBar(t, m); len(got) != 4 {
		t.Errorf("saving shouldn't change this run's tabs, got %q", got)
	}
}

// readConfig is the config file's text.
func readConfig(t *testing.T, m Model) string {
	t.Helper()
	b, err := os.ReadFile(m.configPath)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// pathOf is the saved folder of the bookmark with this label.
func pathOf(t *testing.T, m Model, label string) string {
	t.Helper()
	b, ok := m.cfg.Find(label)
	if !ok {
		t.Fatalf("no bookmark %s", label)
	}
	return b.Path
}

// a in a folder that's already saved, while All and Other are hidden: a
// grouped bookmark opens its group's tab, selected; an ungrouped one turns
// the hiding off, since All is the only tab it shows on, and is selected
// there.
func TestAddSavedWhileHidden(t *testing.T) {
	m := pressDot(groupModel(t)) // Config
	m.cwd = pathOf(t, m, "CS310")
	m, _ = press(m, runes("a"))
	if activeTab(t, m) != "School" || selected(m) != "CS310" || len(tabBar(t, m)) != 2 {
		t.Errorf("want CS310 selected on School, still hidden; got %s on %s, tabs %q", selected(m), activeTab(t, m), tabBar(t, m))
	}
	if !strings.Contains(plain(m.screen()), "already saved as CS310") {
		t.Errorf("want the already-saved note:\n%s", plain(m.screen()))
	}

	m.cwd = pathOf(t, m, "PATHPIN")
	m, _ = press(m, runes("a"))
	if activeTab(t, m) != "All" || selected(m) != "PATHPIN" || len(tabBar(t, m)) != 4 {
		t.Errorf("want PATHPIN selected on All, hiding off; got %s on %s, tabs %q", selected(m), activeTab(t, m), tabBar(t, m))
	}
}
