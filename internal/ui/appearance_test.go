package ui

import (
	"strings"
	"testing"

	"github.com/turtlethom/pathpin/v2/internal/config"
)

// arrowModel is a two-bookmark menu (NOTES selected) in nord with
// [menu] selection = "arrow".
func arrowModel(t *testing.T, bookmarks ...config.Bookmark) Model {
	t.Helper()
	if bookmarks == nil {
		bookmarks = []config.Bookmark{{Label: "NOTES", Path: "notes"}, {Label: "WORK", Path: "work"}}
	}
	m := newTestModel(t, bookmarks...)
	cfg := &config.Config{Colors: config.Colors{Theme: "nord"}, Menu: config.Menu{Selection: "arrow"}, Bookmarks: m.cfg.Bookmarks}
	return started(New(cfg, m.configPath)).sized(80, 24)
}

// In arrow mode the selected row has a › where the bar is, on the label line
// only, and the label is bold in the accent with no background; the path is
// underlined as in the other styles. Other rows are as usual.
func TestArrowSelectionRendered(t *testing.T) {
	withColor(t)
	m := arrowModel(t)
	line := rawLine(t, m.screen(), "NOTES")
	if got := strings.TrimSpace(plain(line)); !strings.HasPrefix(got, "› NOTES") {
		t.Errorf("want the label line to start with › NOTES, got %q", got)
	}
	_, arrow, _ := find(t, cells(line), "›")
	_, label, _ := find(t, cells(line), "NOTES")
	if arrow[0].fg != fg(nordAccent) {
		t.Errorf("want the › in the accent, got %q", arrow[0].fg)
	}
	for _, c := range label {
		if c.fg != fg(nordAccent) || c.bg != "" || !c.bold {
			t.Fatalf("want the label bold in the accent with no background, got fg %q bg %q bold %v", c.fg, c.bg, c.bold)
		}
	}
	path := rawLine(t, m.screen(), "~/notes")
	if strings.ContainsAny(plain(path), "│›") {
		t.Errorf("the path line should have no bar or arrow, got %q", plain(path))
	}
	if _, p, _ := find(t, cells(path), "notes"); !p[0].under {
		t.Error("the selected path should be underlined, as in the other styles")
	}
	if got := strings.TrimSpace(plain(rawLine(t, m.screen(), "WORK"))); got != "WORK" {
		t.Errorf("an unselected row should have no marker, got %q", got)
	}
}

// A selected missing row keeps the › (the cursor is there) but its label
// stays dimmed, since enter won't cd there.
func TestArrowSelectionMissingRow(t *testing.T) {
	withColor(t)
	m := arrowModel(t, config.Bookmark{Label: "GONE", Path: "!gone"}, config.Bookmark{Label: "NOTES", Path: "notes"})
	line := rawLine(t, m.screen(), "GONE")
	if got := strings.TrimSpace(plain(line)); !strings.HasPrefix(got, "› ") {
		t.Errorf("want the › on the missing row, got %q", got)
	}
	_, label, _ := find(t, cells(line), "GONE")
	for _, c := range label {
		if c.fg != fg(nordMuted) {
			t.Fatalf("want a selected missing label muted, got %q", c.fg)
		}
	}
}

// headerModel is groupModel (tabs All Config School Other, All open) started
// with [menu] header = header.
func headerModel(t *testing.T, header string) Model {
	t.Helper()
	return restartedWith(t, groupModel(t), "header = \""+header+"\"")
}

// checkWord checks every cell of word on a rendered line.
func checkWord(t *testing.T, name, line, word, wantFg, wantBg string, bold, under bool) {
	t.Helper()
	_, cs, _ := find(t, cells(line), word)
	for _, c := range cs {
		if c.fg != wantFg || c.bg != wantBg || c.bold != bold || c.under != under {
			t.Errorf("%s: %q drawn fg %q bg %q bold %v under %v, want fg %q bg %q bold %v under %v",
				name, word, c.fg, c.bg, c.bold, c.under, wantFg, wantBg, bold, under)
			return
		}
	}
}

// [menu] header styles the title and the open tab together: badge (a block,
// as before), text (bold accent), bracket ([PATHPIN], [All]) and rule
// (PATHPIN, then a muted rule to the list's edge; the open tab underlined).
// Other tabs stay muted, and PATHPIN starts in the same column in every style.
func TestHeaderStyles(t *testing.T) {
	withColor(t)
	accent, onAccent, muted := fg(nordAccent), fg(nordOnAccent), fg(nordMuted)
	for _, tc := range []struct {
		header       string
		title, tab   string // the text that holds the title and the open tab
		bg           string
		fg           string
		under        bool
		titleRuleEnd bool // the title line ends in a rule
	}{
		{"badge", "PATHPIN", "All", bg(nordAccent), onAccent, false, false},
		{"text", "PATHPIN", "All", "", accent, false, false},
		{"bracket", "[PATHPIN]", "[All]", "", accent, false, false},
		{"rule", "PATHPIN", "All", "", accent, true, true},
	} {
		m := headerModel(t, tc.header)
		view := m.screen()
		titleLine := rawLine(t, view, "PATHPIN")
		if !strings.Contains(plain(titleLine), tc.title) {
			t.Errorf("%s: want %q on the title line, got %q", tc.header, tc.title, plain(titleLine))
		}
		checkWord(t, tc.header+" title", titleLine, "PATHPIN", tc.fg, tc.bg, true, false)
		if col, want := strings.Index(plain(titleLine), "PATHPIN"), strings.Index(plain(rawLine(t, groupModel(t).screen(), "PATHPIN")), "PATHPIN"); col != want {
			t.Errorf("%s: PATHPIN starts in column %d, want %d as with the badge", tc.header, col, want)
		}
		bar := strings.Split(view, "\n")[tabBarAt(t, m)]
		if !strings.Contains(plain(bar), tc.tab) {
			t.Errorf("%s: want the open tab as %q, got %q", tc.header, tc.tab, plain(bar))
		}
		checkWord(t, tc.header+" open tab", bar, "All", tc.fg, tc.bg, true, tc.under)
		checkWord(t, tc.header+" other tab", bar, "Config", muted, "", false, false)
		if tc.header == "bracket" {
			checkWord(t, "bracket title", titleLine, "[", muted, "", false, false)
			checkWord(t, "bracket tab", bar, "[", muted, "", false, false)
		}
		got := strings.TrimRight(plain(titleLine), " ")
		if ends := strings.HasSuffix(got, "───"); ends != tc.titleRuleEnd {
			t.Errorf("%s: title line ends in a rule = %v, want %v: %q", tc.header, ends, tc.titleRuleEnd, got)
		}
		if tc.titleRuleEnd {
			if !strings.Contains(plain(titleLine), "PATHPIN ─") {
				t.Errorf("rule: want the rule right after the title, got %q", got)
			}
			_, rule, _ := find(t, cells(titleLine), "───")
			tabRule := strings.TrimRight(plain(strings.Split(view, "\n")[tabBarAt(t, m)+1]), " ")
			if rule[0].fg != muted || len([]rune(got)) != len([]rune(tabRule)) {
				t.Errorf("rule: want a muted rule out to the tab rule's edge, got %q", got)
			}
		}
	}
}

// In every header style, switching tabs moves nothing on the tab bar.
func TestHeaderTabsDontShift(t *testing.T) {
	for _, header := range []string{"badge", "text", "bracket", "rule"} {
		m := headerModel(t, header)
		before := strings.TrimRight(plain(strings.Split(m.screen(), "\n")[tabBarAt(t, m)]), " ")
		m = pressTab(m)
		after := strings.TrimRight(plain(strings.Split(m.screen(), "\n")[tabBarAt(t, m)]), " ")
		// The open tab's decoration moves; the names stay in their columns.
		strip := strings.NewReplacer("[", " ", "]", " ")
		if strip.Replace(after) != strip.Replace(before) {
			t.Errorf("%s: the tab bar moved:\nbefore %q\nafter  %q", header, before, after)
		}
	}
}

// The header style also applies to the empty menu's title and to the ?
// page's, so the title looks the same everywhere.
func TestHeaderEverywhere(t *testing.T) {
	withColor(t)
	empty := restartedWith(t, newTestModel(t), `header = "text"`)
	checkWord(t, "empty menu", rawLine(t, empty.screen(), "PATHPIN"), "PATHPIN", fg(nordAccent), "", true, false)
	help, _ := press(headerModel(t, "text"), runes("?"))
	checkWord(t, "? page", rawLine(t, help.screen(), "PATHPIN"), "PATHPIN", fg(nordAccent), "", true, false)
}

// Names match in any case, and an unknown header falls back to badge (cmd
// warns about it before the menu opens).
func TestHeaderFallback(t *testing.T) {
	withColor(t)
	checkWord(t, "Text", rawLine(t, headerModel(t, "Text").screen(), "PATHPIN"), "PATHPIN", fg(nordAccent), "", true, false)
	checkWord(t, "unknown", rawLine(t, headerModel(t, "banner").screen(), "PATHPIN"), "PATHPIN", fg(nordOnAccent), bg(nordAccent), true, false)
}

// The picker's header row cycles badge → text → bracket → rule (wrapping),
// previewing on the menu behind; d resets to badge; enter saves it.
func TestPickerHeaderRow(t *testing.T) {
	withColor(t)
	m := toRow(t, openPicker(t, groupModel(t).sized(80, 30)), "header")
	if line := pickerLine(t, m, "header"); !strings.Contains(line, "badge") || !strings.Contains(line, "title and open tab") {
		t.Errorf("want badge and the hint: %q", line)
	}
	for _, want := range []string{"text", "bracket", "rule", "badge"} {
		m = right(m)
		if line := pickerLine(t, m, "header"); !strings.Contains(line, want) {
			t.Fatalf("→ should show %s: %q", want, line)
		}
	}
	m = right(m) // text
	if _, title, _ := find(t, cells(rawLine(t, m.screen(), "PATHPIN")), "PATHPIN"); title[0].bg != "" {
		t.Errorf("text should preview with no badge, got bg %q", title[0].bg)
	}
	m = enter(m)
	if !saved(m) || !strings.Contains(readConfig(t, m), `header = "text"`) {
		t.Errorf("enter should save header = \"text\":\n%s", readConfig(t, m))
	}
}

// pickerRowShown reports whether the picker shows the row with this name.
func pickerRowShown(m Model, name string) bool {
	for line := range strings.SplitSeq(plain(m.screen()), "\n") {
		if f := strings.Fields(strings.ReplaceAll(line, "›", " ")); len(f) > 0 && f[0] == name {
			return true
		}
	}
	return false
}

// In a window too short for all its rows, the t picker scrolls like m's:
// the menu fits with its title and the selected bookmark in view, the
// highlighted row always shows, and lines say how many rows are hidden.
func TestThemePickerScrolls(t *testing.T) {
	m := openPicker(t, groupModel(t).sized(80, 18))
	check := func(step string) {
		t.Helper()
		v := plain(m.screen())
		if n := len(strings.Split(m.screen(), "\n")); n > 18 {
			t.Fatalf("%s: the menu is %d lines, want at most 18:\n%s", step, n, v)
		}
		if !strings.Contains(v, "PATHPIN") || !strings.Contains(v, "BASH") {
			t.Fatalf("%s: want the title and the selected bookmark in view:\n%s", step, v)
		}
	}
	check("opened")
	if v := plain(m.screen()); !moreBelow.MatchString(v) || moreAbove.MatchString(v) || pickerRowShown(m, "tabs") {
		t.Errorf("opened at the top: want a ↓ more line and tabs out of view:\n%s", v)
	}
	for i, name := range pickerRows[1:] {
		m, _ = press(m, keyMsg("down"))
		check("down to " + name)
		if !strings.Contains(pickerLine(t, m, name), "›") {
			t.Fatalf("step %d: want %s highlighted:\n%s", i+1, name, plain(m.screen()))
		}
	}
	if v := plain(m.screen()); !moreAbove.MatchString(v) || moreBelow.MatchString(v) {
		t.Errorf("at the bottom: want only an ↑ more line:\n%s", v)
	}
	m, _ = press(m, keyMsg("down")) // wraps to theme
	check("wrapped")
	if !strings.Contains(pickerLine(t, m, "theme"), "›") {
		t.Errorf("want theme highlighted after wrapping:\n%s", plain(m.screen()))
	}
}
