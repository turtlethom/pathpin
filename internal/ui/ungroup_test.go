package ui

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/turtlethom/pathpin/v2/internal/config"
)

// u asks before ungrouping the highlighted bookmark. Any key but y cancels,
// saving nothing (q included: it doesn't quit). ? lists u.
func TestUngroupOneAsks(t *testing.T) {
	m := selectLabel(t, groupModel(t), "BASH") // Config
	if help, _ := press(m, runes("?")); !helpShows(help, "u", "ungroup") {
		t.Errorf("the help page (?) should list u ungroup:\n%s", plain(help.View()))
	}
	for _, k := range []tea.KeyMsg{runes("n"), runes("q"), {Type: tea.KeyEsc}} {
		m, _ := press(m, runes("u"))
		if m.mode != modeConfirm || !strings.Contains(plain(m.View()), "Ungroup BASH? y/N") {
			t.Fatalf("u should ask Ungroup BASH? y/N:\n%s", plain(m.View()))
		}
		m, cmd := press(m, k)
		if m.mode != modeBrowse || isQuit(cmd) {
			t.Errorf("%s should cancel without quitting", k)
		}
		if _, err := os.Stat(m.configPath + ".bak"); err == nil || groupOnDisk(t, m, "BASH") != "Config" {
			t.Errorf("%s: cancelling should not save", k)
		}
	}
}

// y ungroups: on a group's tab the bookmark leaves it (the tab stays open while
// it has bookmarks), and the file keeps its order. When the tab empties, All
// opens and the tab is gone.
func TestUngroupOneOnGroupTab(t *testing.T) {
	m := selectLabel(t, schoolTab(t), "CS310")
	m, _ = press(m, runes("u"))
	m, _ = press(m, runes("y"))
	if m.mode != modeBrowse {
		t.Fatalf("y should close the prompt:\n%s", plain(m.View()))
	}
	if got := groupOnDisk(t, m, "CS310"); got != "" {
		t.Errorf("want CS310 ungrouped, got %q", got)
	}
	if want := []string{"BASH", "CS310", "PATHPIN", "GHOSTTY", "NOTES", "CSSLAB"}; !slices.Equal(fileOrder(t, m), want) {
		t.Errorf("order on disk %q, want %q", fileOrder(t, m), want)
	}
	if got := activeTab(t, m); got != "School" {
		t.Errorf("want School still open, got %s", got)
	}
	if got := shown(m); !slices.Equal(got, []string{"NOTES"}) {
		t.Errorf("School after ungrouping CS310: got %q", got)
	}
	if !strings.Contains(plain(m.View()), "CS310 has no group now") {
		t.Errorf("want the no-group status:\n%s", plain(m.View()))
	}

	m, _ = press(m, runes("u")) // NOTES, the last in School
	m, _ = press(m, runes("y"))
	if got := activeTab(t, m); got != "All" {
		t.Errorf("want All once School is empty, got %s", got)
	}
	if got, want := tabBar(t, m), []string{"All", "Config", "Other"}; !slices.Equal(got, want) {
		t.Errorf("tab bar: got %q, want %q", got, want)
	}
}

// On All the bookmark stays in view and selected, without its group tag.
func TestUngroupOneOnAll(t *testing.T) {
	m := selectLabel(t, groupModel(t), "BASH")
	m, _ = press(m, runes("u"))
	m, _ = press(m, runes("y"))
	if got := selected(m); got != "BASH" {
		t.Errorf("want BASH still selected, got %q", got)
	}
	if got := row(t, m, "BASH"); strings.Contains(got, "·") {
		t.Errorf("want BASH's tag gone, got %q", got)
	}
}

// u on a bookmark with no group says so and saves nothing. No bookmarks: u
// does nothing. While a filter is being typed, u is a letter in it.
func TestUngroupOneNothingToDo(t *testing.T) {
	m, _ := press(selectLabel(t, groupModel(t), "PATHPIN"), runes("u"))
	if m.mode != modeBrowse || !strings.Contains(plain(m.View()), "PATHPIN has no group") {
		t.Errorf("want a note and no prompt:\n%s", plain(m.View()))
	}
	if _, err := os.Stat(m.configPath + ".bak"); err == nil {
		t.Error("u on an ungrouped bookmark should not save")
	}

	if m, _ := press(newTestModel(t), runes("u")); m.mode != modeBrowse {
		t.Error("u on an empty menu should do nothing")
	}

	m, _ = press(groupModel(t), runes("/"))
	m, _ = press(m, runes("u"))
	if m.mode != modeBrowse || m.list.FilterValue() != "u" {
		t.Errorf("want u typed into the filter, got mode %v, filter %q", m.mode, m.list.FilterValue())
	}
}

// bubbles binds u to the previous page (hidden from its help). pathpin drops
// that, so u only ungroups; pgup, h and ← still page up.
func TestUIsNotPageUp(t *testing.T) {
	var bookmarks []config.Bookmark
	for i := range 30 {
		bookmarks = append(bookmarks, config.Bookmark{Label: fmt.Sprintf("B%02d", i), Path: fmt.Sprintf("b%02d", i)})
	}
	m, _ := press(newTestModel(t, bookmarks...), tea.KeyMsg{Type: tea.KeyRight})
	if m.list.Paginator.Page != 1 {
		t.Fatalf("→ should open the second page, got page %d", m.list.Paginator.Page)
	}
	m, _ = press(m, runes("u"))
	if m.list.Paginator.Page != 1 {
		t.Errorf("u should not page up, got page %d", m.list.Paginator.Page)
	}
	m, _ = press(m, runes("h"))
	if m.list.Paginator.Page != 0 {
		t.Errorf("h should still page up, got page %d", m.list.Paginator.Page)
	}
}
