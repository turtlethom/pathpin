package ui

import (
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/turtlethom/pathpin/v2/internal/config"
)

// pickerRows are the picker's rows, top to bottom, named as in the TOML.
var pickerRows = []string{"theme", "primary", "secondary", "tertiary", "selected_text", "error", "selection", "icons", "header", "tabs"}

// colorsModel builds a menu with two bookmarks and these [colors].
func colorsModel(t *testing.T, c config.Colors) Model {
	t.Helper()
	return lookModel(t, c, config.Menu{})
}

// lookModel builds a menu with two bookmarks (NVIM, NOTES), these [colors]
// and this [menu].
func lookModel(t *testing.T, c config.Colors, menu config.Menu) Model {
	t.Helper()
	m := twoBookmarks(t) // creates the folders and the config file
	cfg := &config.Config{Colors: c, Menu: menu, Bookmarks: m.cfg.Bookmarks}
	if err := config.Save(m.configPath, cfg); err != nil {
		t.Fatal(err)
	}
	os.Remove(m.configPath + ".bak") // so tests can tell whether the picker saved
	return started(New(cfg, m.configPath)).sized(80, 24)
}

func openPicker(t *testing.T, m Model) Model {
	t.Helper()
	m, _ = press(m, runes("t"))
	if m.mode != modeTheme {
		t.Fatalf("t should open the theme picker:\n%s", plain(m.screen()))
	}
	return m
}

// toRow highlights the picker row with this name, from the top.
func toRow(t *testing.T, m Model, name string) Model {
	t.Helper()
	i := slices.Index(pickerRows, name)
	if i < 0 {
		t.Fatalf("no picker row %q", name)
	}
	for range len(pickerRows) { // up to the theme row, whether ↑ stops at the top or wraps
		m, _ = press(m, keyMsg("up"))
		if strings.Contains(pickerLine(t, m, "theme"), "›") {
			break
		}
	}
	for range i {
		m, _ = press(m, keyMsg("down"))
	}
	if !strings.Contains(pickerLine(t, m, name), "›") {
		t.Fatalf("want %s highlighted:\n%s", name, plain(m.screen()))
	}
	return m
}

// pickerLine is the visible text of the picker row with this name.
func pickerLine(t *testing.T, m Model, name string) string {
	t.Helper()
	for line := range strings.SplitSeq(plain(m.screen()), "\n") {
		if f := strings.Fields(strings.ReplaceAll(line, "›", " ")); len(f) > 0 && f[0] == name {
			return line
		}
	}
	t.Fatalf("no %s row in the picker:\n%s", name, plain(m.screen()))
	return ""
}

func right(m Model) Model { m, _ = press(m, keyMsg("right")); return m }
func left(m Model) Model  { m, _ = press(m, keyMsg("left")); return m }

// saved reports whether the menu has written the config since colorsModel.
func saved(m Model) bool {
	_, err := os.Stat(m.configPath + ".bak")
	return err == nil
}

// t opens the picker with a row per [colors] key plus selection and icons,
// the theme row highlighted, and its keys on the bottom line. The help
// mentions t; the browse help line is hidden while picking.
func TestThemePickerOpens(t *testing.T) {
	m := colorsModel(t, config.Colors{})
	if help, _ := press(m, runes("?")); !helpShows(help, "t", "appearance") {
		t.Errorf("the help page (?) should mention t:\n%s", plain(help.screen()))
	}
	m = openPicker(t, m)
	for _, row := range pickerRows {
		pickerLine(t, m, row)
	}
	if !strings.Contains(pickerLine(t, m, "theme"), "›") {
		t.Errorf("the theme row should be highlighted first:\n%s", plain(m.screen()))
	}
	if !strings.Contains(pickerLine(t, m, "theme"), "terminal") {
		t.Errorf("no theme set should show terminal: %q", pickerLine(t, m, "theme"))
	}
	if !strings.Contains(pickerLine(t, m, "primary"), "theme's") {
		t.Errorf("an unset color should show theme's: %q", pickerLine(t, m, "primary"))
	}
	view := plain(m.screen())
	for _, s := range []string{"d reset", "enter save", "esc cancel"} {
		if !strings.Contains(view, s) {
			t.Errorf("want %q on the picker's bottom line:\n%s", s, view)
		}
	}
	if strings.Contains(view, "a add") {
		t.Errorf("the browse help shouldn't show while picking:\n%s", view)
	}
	if saved(m) {
		t.Error("opening the picker shouldn't save")
	}

	// An empty menu has colors too, with one blank line above the rows as usual.
	lines := strings.Split(plain(openPicker(t, newTestModel(t)).screen()), "\n")
	i := slices.IndexFunc(lines, func(l string) bool { return strings.Contains(l, "No bookmarks yet") })
	if i < 0 || i+2 >= len(lines) || strings.TrimSpace(lines[i+1]) != "" || !strings.Contains(lines[i+2], "theme") {
		t.Errorf("want one blank line between the empty message and the picker:\n%s", strings.Join(lines, "\n"))
	}
}

// While typing a filter, t is typed.
func TestTWhileFilteringIsTyped(t *testing.T) {
	m, _ := press(twoBookmarks(t), runes("/"))
	m, _ = press(m, runes("t"))
	if m.mode != modeBrowse || m.list.FilterValue() != "t" {
		t.Errorf("t should be typed into the filter; mode %v, filter %q", m.mode, m.list.FilterValue())
	}
}

// The panel fits the window: the list shrinks to make room, with or without
// the tab bar, and after a resize.
func TestThemePickerFitsWindow(t *testing.T) {
	for name, m := range map[string]Model{
		"no groups": colorsModel(t, config.Colors{}),
		"groups":    groupModel(t).sized(80, 24),
		"empty":     newTestModel(t),
	} {
		m = openPicker(t, m)
		for size, m := range map[string]Model{"opened": m, "resized": m.sized(80, 24)} {
			lines := strings.Split(m.screen(), "\n")
			if len(lines) > 24 {
				t.Errorf("%s, %s: %d lines, want at most 24:\n%s", name, size, len(lines), plain(m.screen()))
			}
			for _, l := range lines {
				if w := lipgloss.Width(l); w > 80 {
					t.Errorf("%s, %s: a line is %d columns: %q", name, size, w, plain(l))
				}
			}
			pickerLine(t, m, "tabs") // the last row still shows
		}
	}
}

// ↑↓ move between rows, not the list's cursor; q doesn't quit.
func TestThemePickerKeysStayInPicker(t *testing.T) {
	m := openPicker(t, colorsModel(t, config.Colors{}))
	for _, k := range []tea.KeyMsg{runes("j"), keyMsg("down")} {
		m, _ = press(m, k)
	}
	if !strings.Contains(pickerLine(t, m, "secondary"), "›") {
		t.Errorf("j and ↓ should move down two rows:\n%s", plain(m.screen()))
	}
	m, _ = press(m, runes("k"))
	if !strings.Contains(pickerLine(t, m, "primary"), "›") {
		t.Errorf("k should move up a row:\n%s", plain(m.screen()))
	}
	if m.list.Index() != 0 {
		t.Errorf("the list's cursor moved to %d", m.list.Index())
	}
	m, cmd := press(m, runes("q"))
	if m.mode != modeTheme || isQuit(cmd) {
		t.Error("q shouldn't quit or close the picker")
	}
}

// ←→ cycle the theme through ThemeNames, wrapping, and the menu is redrawn in
// it right away: the title badge and the selected row use the new accent.
func TestThemeRowCyclesAndPreviews(t *testing.T) {
	withColor(t)
	names := ThemeNames()
	m := openPicker(t, colorsModel(t, config.Colors{}))

	m = right(m)
	if !strings.Contains(pickerLine(t, m, "theme"), names[1]) {
		t.Errorf("→ should show %s: %q", names[1], pickerLine(t, m, "theme"))
	}
	if colorsOf(m.styles) != colorsOf(NewStyles(config.Colors{Theme: names[1]}, config.Menu{})) {
		t.Errorf("→ should switch the styles to %s", names[1])
	}
	m = left(left(m))
	if !strings.Contains(pickerLine(t, m, "theme"), names[len(names)-1]) {
		t.Errorf("← from terminal should wrap to %s: %q", names[len(names)-1], pickerLine(t, m, "theme"))
	}

	for range names {
		if strings.Contains(pickerLine(t, m, "theme"), "nord") {
			break
		}
		m = right(m)
	}
	view := m.screen()
	for _, s := range []string{"PATHPIN", "NVIM"} { // the badge and the selected row's block
		_, word, _ := find(t, cells(rawLine(t, view, s)), s)
		if word[0].bg != bg(nordAccent) {
			t.Errorf("%s should be drawn on nord's accent while previewing, got bg %q", s, word[0].bg)
		}
	}
	if saved(m) {
		t.Error("cycling shouldn't save")
	}
}

// esc closes the picker without saving and puts the saved colors back.
func TestThemePickerEscRestores(t *testing.T) {
	withColor(t)
	m := openPicker(t, colorsModel(t, config.Colors{Theme: "nord"}))
	m = toRow(t, right(m), "primary")
	m = right(m)
	m, _ = press(m, keyMsg("esc"))
	if m.mode != modeBrowse {
		t.Fatalf("esc should close the picker:\n%s", plain(m.screen()))
	}
	if saved(m) || m.cfg.Colors != (config.Colors{Theme: "nord"}) {
		t.Errorf("esc shouldn't save or change the colors, got %+v", m.cfg.Colors)
	}
	if colorsOf(m.styles) != colorsOf(NewStyles(m.cfg.Colors, config.Menu{})) {
		t.Error("esc should restore the saved colors")
	}
	_, word, _ := find(t, cells(rawLine(t, m.screen(), "PATHPIN")), "PATHPIN")
	if word[0].bg != bg(nordAccent) {
		t.Errorf("the badge should be back on nord's accent, got bg %q", word[0].bg)
	}
}

// colorChoices are the terminal's 16 ANSI colors, then the named colors in
// hue order (reds round to pinks, then browns and neutrals light to dark).
// Every named color is one lowercase word that no other color uses, ANSI
// names included, and every value is a valid, unrepeated color.
func TestColorChoices(t *testing.T) {
	choices := colorChoices()
	ansi := []string{"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white"}
	named := [][2]string{
		{"crimson", "#DC143C"}, {"blush", "#E06C75"}, {"salmon", "#F08080"},
		{"copper", "#D19A66"}, {"orange", "#FF9500"},
		{"honey", "#E5C078"}, {"mustard", "#DAA520"}, {"amber", "#FFBF00"},
		{"lime", "#A6E22E"}, {"pistachio", "#98C379"}, {"emerald", "#50C878"}, {"mint", "#3EB489"},
		{"teal", "#008080"}, {"aqua", "#56B6C2"}, {"sky", "#87CEEB"},
		{"cobalt", "#0047AB"}, {"cornflower", "#6495ED"}, {"navy", "#000080"},
		{"amethyst", "#9966CC"}, {"indigo", "#4B0082"}, {"lavender", "#B57EDC"}, {"orchid", "#C678DD"},
		{"rose", "#FF007F"},
		{"sienna", "#A0522D"},
		{"snow", "#FFFFFF"}, {"silver", "#C0C0C0"}, {"slate", "#5C6370"}, {"charcoal", "#36454F"}, {"ink", "#000000"},
	}
	if len(choices) != 16+len(named) {
		t.Fatalf("want %d choices, got %d", 16+len(named), len(choices))
	}
	for i, c := range choices[:16] {
		want := ansi[i%8]
		if i >= 8 {
			want = "bright-" + want
		}
		if c.name != want || c.value != strconv.Itoa(i) {
			t.Errorf("choice %d: want %s · %d, got %s · %s", i, want, i, c.name, c.value)
		}
	}
	for i, c := range choices[16:] {
		if c.name != named[i][0] || !strings.EqualFold(c.value, named[i][1]) {
			t.Errorf("choice %d: want %s · %s, got %s · %s", 16+i, named[i][0], named[i][1], c.name, c.value)
		}
	}

	names, values := map[string]bool{}, map[string]bool{}
	for _, c := range choices {
		if names[c.name] {
			t.Errorf("the name %q is used twice", c.name)
		}
		names[c.name] = true
		if !validColor(c.value) || values[strings.ToLower(c.value)] {
			t.Errorf("%s · %s is invalid or repeated", c.name, c.value)
		}
		values[strings.ToLower(c.value)] = true
	}
	for _, c := range choices[16:] {
		if c.name != strings.ToLower(c.name) || strings.ContainsAny(c.name, "- ") {
			t.Errorf("%q should be one lowercase word", c.name)
		}
	}
}

// ←→ on a color row cycle theme's → the choices → theme's, showing name ·
// value, and the slot's color changes with it.
func TestColorRowCycles(t *testing.T) {
	choices := colorChoices()
	m := toRow(t, openPicker(t, colorsModel(t, config.Colors{})), "primary")

	m = right(m)
	if line := pickerLine(t, m, "primary"); !strings.Contains(line, "black · 0") {
		t.Errorf("→ from theme's should show the first choice: %q", line)
	}
	if got := colorsOf(m.styles).accent; got != lipgloss.Color("0") {
		t.Errorf("primary should preview as 0, got %v", got)
	}
	m = left(m)
	if line := pickerLine(t, m, "primary"); !strings.Contains(line, "theme's") {
		t.Errorf("← should go back to theme's: %q", line)
	}
	m = left(m)
	last := choices[len(choices)-1]
	if line := pickerLine(t, m, "primary"); !strings.Contains(line, last.name+" · "+last.value) {
		t.Errorf("← from theme's should wrap to %s · %s: %q", last.name, last.value, line)
	}

	// h and l work too, on any color row.
	m = toRow(t, m, "error")
	m, _ = press(m, runes("l"))
	if got := colorsOf(m.styles).err; got != lipgloss.Color("0") {
		t.Errorf("l on error should preview 0, got %v", got)
	}
}

// An ANSI color's name is a palette slot's, not its hue (Monokai Pro's 12,
// bright-blue, is orange), so while a color row shows one, its hint says the
// terminal decides. A hex color keeps the row's own hint.
func TestPickerPaletteHint(t *testing.T) {
	m := openPicker(t, colorsModel(t, config.Colors{Theme: "nord", Primary: "12", Error: "#0047AB"}))
	if line := pickerLine(t, m, "primary"); !strings.Contains(line, "bright-blue · 12") ||
		!strings.Contains(line, "your terminal's palette") || strings.Contains(line, "title, selected row") {
		t.Errorf("an ANSI color should keep its name and say the terminal decides: %q", line)
	}
	if line := pickerLine(t, m, "error"); !strings.Contains(line, "errors") || strings.Contains(line, "palette") {
		t.Errorf("a hex color should keep the row's hint: %q", line)
	}

	// Only a name needs explaining: 16-255 (typed with #) have none, and a
	// default row says theme's, even when the theme's color is ANSI.
	m = openPicker(t, colorsModel(t, config.Colors{Tertiary: "200"}))
	if line := pickerLine(t, m, "tertiary"); !strings.Contains(line, "paths") {
		t.Errorf("an unnamed ANSI color should keep the row's hint: %q", line)
	}
	if line := pickerLine(t, m, "primary"); !strings.Contains(line, "title, selected row") {
		t.Errorf("a default row should keep its hint on the terminal theme: %q", line)
	}
	m = toRow(t, openPicker(t, colorsModel(t, config.Colors{Theme: "nord", Primary: "12"})), "primary")

	// Cycling onto a hex color brings the row's hint back.
	for range 4 { // 12 → 13, 14, 15, then crimson
		m = right(m)
	}
	if line := pickerLine(t, m, "primary"); !strings.Contains(line, "crimson") || !strings.Contains(line, "title, selected row") {
		t.Errorf("a hex color should show the row's hint again: %q", line)
	}
}

// A color row at its default shows the theme's color as a swatch before
// theme's, so d's result can be seen.
func TestPickerDefaultSwatch(t *testing.T) {
	withColor(t)
	m := openPicker(t, colorsModel(t, config.Colors{Theme: "nord"}))
	for row, want := range map[string]string{
		"primary":   nordAccent,
		"secondary": nordText,
		"tertiary":  nordMuted,
	} {
		_, swatch, after := find(t, cells(rawLine(t, m.screen(), row+" ")), "██")
		if !allFg(swatch, fg(want)) {
			t.Errorf("%s: want a swatch in the theme's %s, got %+v", row, want, swatch)
		}
		if !strings.HasPrefix(strings.TrimSpace(string(runesOf(after))), "theme's") {
			t.Errorf("%s: want theme's after the swatch, got %q", row, string(runesOf(after)))
		}
	}

	// The swatch follows the theme row.
	m = right(m) // nord → one-dark
	_, swatch, _ := find(t, cells(rawLine(t, m.screen(), "primary ")), "██")
	if !allFg(swatch, fg("#56b6c2")) {
		t.Errorf("after switching to one-dark, want its accent, got %+v", swatch)
	}

	// The terminal theme's labels have no color (the terminal's own text
	// color), and so does their swatch.
	m = openPicker(t, colorsModel(t, config.Colors{}))
	_, swatch, _ = find(t, cells(rawLine(t, m.screen(), "secondary ")), "██")
	if !allFg(swatch, "") {
		t.Errorf("want the secondary swatch in the terminal's text color, got %+v", swatch)
	}
}

// selected_text's swatch is Aa on the accent, like the badge: a text color
// alone says little (the terminal theme's 0 is often the background itself).
// The accent is primary's, chosen or the theme's.
func TestPickerTextSwatch(t *testing.T) {
	withColor(t)
	for _, tc := range []struct {
		colors   config.Colors
		text, on string
	}{
		{config.Colors{Theme: "nord"}, nordOnAccent, nordAccent},
		{config.Colors{Theme: "nord", Primary: "#123456", SelectedText: "#ABCDEF"}, "#ABCDEF", "#123456"},
	} {
		m := openPicker(t, colorsModel(t, tc.colors))
		_, aa, _ := find(t, cells(rawLine(t, m.screen(), "selected_text ")), "Aa")
		for _, c := range aa {
			if c.fg != fg(tc.text) || c.bg != bg(tc.on) {
				t.Errorf("%+v: want Aa in %s on %s, got fg %q bg %q", tc.colors, tc.text, tc.on, c.fg, c.bg)
				break
			}
		}
	}
}

func runesOf(cs []cell) []rune {
	out := make([]rune, len(cs))
	for i, c := range cs {
		out[i] = c.r
	}
	return out
}

// enter saves the draft, keeping the bookmarks and other settings, and the
// menu keeps the new colors.
func TestThemePickerSaves(t *testing.T) {
	m := colorsModel(t, config.Colors{})
	m.cfg.Menu.Icons = "nerd"
	m = openPicker(t, m)
	m = right(m) // the first preset
	m = right(toRow(t, m, "primary"))
	m = enter(m)
	if m.mode != modeBrowse {
		t.Fatalf("enter should close the picker:\n%s", plain(m.screen()))
	}
	want := config.Colors{Theme: ThemeNames()[1], Primary: "0"}
	cfg, err := config.Load(m.configPath)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Colors != want || cfg.Menu.Icons == "" || len(cfg.Bookmarks) != 2 {
		t.Errorf("want colors %+v saved with icons and 2 bookmarks, got %+v, %+v, %d", want, cfg.Colors, cfg.Menu, len(cfg.Bookmarks))
	}
	if m.cfg.Colors != want || colorsOf(m.styles) != colorsOf(NewStyles(want, config.Menu{})) {
		t.Error("the menu should keep the saved colors")
	}
	if !strings.Contains(plain(m.screen()), "Appearance saved") {
		t.Errorf("want a status:\n%s", plain(m.screen()))
	}
}

// enter with nothing changed (or changed and changed back) saves nothing,
// since every save replaces the one backup.
func TestThemePickerUnchangedDoesNotSave(t *testing.T) {
	m := enter(openPicker(t, colorsModel(t, config.Colors{Theme: "nord"})))
	if m.mode != modeBrowse || saved(m) {
		t.Errorf("unchanged colors shouldn't save; mode %v, saved %v", m.mode, saved(m))
	}
	if !strings.Contains(plain(m.screen()), "Appearance unchanged") {
		t.Errorf("want a status:\n%s", plain(m.screen()))
	}
	m = enter(left(right(openPicker(t, m))))
	if saved(m) {
		t.Error("changed and changed back shouldn't save")
	}
}

// Switching the theme keeps colors set on the other rows.
func TestThemeSwitchKeepsOverrides(t *testing.T) {
	m := right(openPicker(t, colorsModel(t, config.Colors{Primary: "#ff0000"})))
	if line := pickerLine(t, m, "primary"); !strings.Contains(line, "#ff0000") {
		t.Errorf("primary should keep its override: %q", line)
	}
	if got := colorsOf(m.styles).accent; got != lipgloss.Color("#ff0000") {
		t.Errorf("the override should still win over the theme, got %v", got)
	}
}

// d resets a row: a color to the theme's, the theme to terminal, which saves
// as no theme at all.
func TestThemePickerResetRow(t *testing.T) {
	m := openPicker(t, colorsModel(t, config.Colors{Theme: "nord", Primary: "6"}))
	if line := pickerLine(t, m, "primary"); !strings.Contains(line, "cyan · 6") {
		t.Errorf("want name · value for an ANSI color: %q", line)
	}
	m, _ = press(m, runes("d"))
	if line := pickerLine(t, m, "theme"); !strings.Contains(line, "terminal") {
		t.Errorf("d on the theme row should pick terminal: %q", line)
	}
	m, _ = press(toRow(t, m, "primary"), runes("d"))
	if line := pickerLine(t, m, "primary"); !strings.Contains(line, "theme's") {
		t.Errorf("d on a color row should reset it: %q", line)
	}
	m = enter(m)
	cfg, err := config.Load(m.configPath)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Colors != (config.Colors{}) {
		t.Errorf("want no colors saved, got %+v", cfg.Colors)
	}
}

// # on a color row opens a field for any color: it's filled with the row's
// hex (or just #), a bad value keeps it open, enter puts the value in the draft
// (not saved yet), and esc goes back to the picker. The theme row has no field.
func TestThemePickerHexField(t *testing.T) {
	m := openPicker(t, colorsModel(t, config.Colors{}))
	if m, _ = press(m, runes("#")); m.mode != modeTheme {
		t.Errorf("# on the theme row should do nothing:\n%s", plain(m.screen()))
	}

	m, _ = press(toRow(t, m, "primary"), runes("#"))
	if m.mode == modeTheme || !strings.Contains(plain(m.screen()), "primary: #") {
		t.Fatalf("# should open a primary: field:\n%s", plain(m.screen()))
	}
	m, _ = press(m, runes("12"))
	m = enter(m)
	if m.mode == modeTheme || m.inputErr == "" {
		t.Fatalf("#12 should be refused with the field open:\n%s", plain(m.screen()))
	}
	m, _ = press(clearField(m), runes("#ABCDEF"))
	m = enter(m)
	if m.mode != modeTheme {
		t.Fatalf("a good value should go back to the picker:\n%s", plain(m.screen()))
	}
	if line := pickerLine(t, m, "primary"); !strings.Contains(line, "#ABCDEF") {
		t.Errorf("want the typed value on the row: %q", line)
	}
	if got := colorsOf(m.styles).accent; got != lipgloss.Color("#ABCDEF") {
		t.Errorf("want #ABCDEF previewed, got %v", got)
	}
	if saved(m) {
		t.Error("the field shouldn't save; enter in the picker does")
	}

	m, _ = press(m, runes("#"))
	if !strings.Contains(plain(m.screen()), "primary: #ABCDEF") {
		t.Errorf("the field should start with the row's value:\n%s", plain(m.screen()))
	}
	m, _ = press(m, keyMsg("esc"))
	if m.mode != modeTheme || !strings.Contains(pickerLine(t, m, "primary"), "#ABCDEF") {
		t.Errorf("esc should go back to the picker with the draft kept:\n%s", plain(m.screen()))
	}
}

// The # field starts with an ANSI number too, and takes one typed in.
func TestThemePickerHexFieldTakesANSI(t *testing.T) {
	m := toRow(t, openPicker(t, colorsModel(t, config.Colors{Tertiary: "8"})), "tertiary")
	m, _ = press(m, runes("#"))
	if !strings.Contains(plain(m.screen()), "tertiary: 8") {
		t.Errorf("the field should start with the row's 8:\n%s", plain(m.screen()))
	}
	m, _ = press(clearField(m), runes("6"))
	m = enter(m)
	if m.mode != modeTheme || colorsOf(m.styles).muted != lipgloss.Color("6") {
		t.Errorf("6 should be taken; mode %v, muted %v", m.mode, colorsOf(m.styles).muted)
	}
}

// The # field doesn't leave its prompt on the label field used by a and n.
func TestHexFieldKeepsLabelPrompt(t *testing.T) {
	m := toRow(t, openPicker(t, colorsModel(t, config.Colors{})), "primary")
	m, _ = press(m, runes("#"))
	m, _ = press(m, keyMsg("esc"))
	m, _ = press(m, keyMsg("esc"))
	m, _ = press(m, runes("n"))
	if view := plain(m.screen()); !strings.Contains(view, "Label: NVIM") {
		t.Errorf("n should show the Label: field after using #:\n%s", view)
	}
}

// Values the menu ignores (from a hand-edited file) are shown as such.
func TestThemePickerShowsBadValues(t *testing.T) {
	m := openPicker(t, colorsModel(t, config.Colors{Theme: "dracula", Primary: "aqua"}))
	if line := pickerLine(t, m, "theme"); !strings.Contains(line, "dracula (unknown)") {
		t.Errorf("want the unknown theme marked: %q", line)
	}
	if line := pickerLine(t, m, "primary"); !strings.Contains(line, "aqua (ignored)") {
		t.Errorf("want the bad color marked: %q", line)
	}
}

// A failed save keeps the picker open with the error, and the file's colors.
func TestThemePickerFailedSave(t *testing.T) {
	m := colorsModel(t, config.Colors{})
	blocker := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(blocker, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	m.configPath = filepath.Join(blocker, "config.toml")

	m = enter(right(openPicker(t, m)))
	if m.mode != modeTheme || !strings.Contains(plain(m.screen()), "couldn't save") {
		t.Errorf("want the picker open with the error:\n%s", plain(m.screen()))
	}
	if m.cfg.Colors != (config.Colors{}) {
		t.Errorf("a failed save shouldn't change the colors, got %+v", m.cfg.Colors)
	}
}

// hasBlock reports whether the menu draws the selected row as a block (else
// with just the bar).
func hasBlock(m Model) bool {
	_, none := m.styles.item.SelectedTitle.GetBackground().(lipgloss.NoColor)
	return !none
}

// The selection row shows block (the default), bar or arrow, with a hint. ←→
// cycle them (wrapping), the menu previews the change, and d resets to block.
func TestPickerSelectionRow(t *testing.T) {
	m := openPicker(t, colorsModel(t, config.Colors{}))
	if line := pickerLine(t, m, "selection"); !strings.Contains(line, "block") || !strings.Contains(line, "block, bar or arrow") {
		t.Errorf("want block and the hint: %q", line)
	}
	m = right(toRow(t, m, "selection"))
	if line := pickerLine(t, m, "selection"); !strings.Contains(line, "bar") || hasBlock(m) {
		t.Errorf("→ should show and preview bar: %q, block %v", line, hasBlock(m))
	}
	m = right(m)
	if line := pickerLine(t, m, "selection"); !strings.Contains(line, "arrow") || hasBlock(m) || !strings.Contains(plain(m.screen()), "› NVIM") {
		t.Errorf("→ from bar should show and preview arrow: %q, block %v", line, hasBlock(m))
	}
	m = right(m)
	if line := pickerLine(t, m, "selection"); !strings.Contains(line, "block") || !hasBlock(m) {
		t.Errorf("→ from arrow should wrap to block: %q, block %v", line, hasBlock(m))
	}
	m = left(m)
	if line := pickerLine(t, m, "selection"); !strings.Contains(line, "arrow") {
		t.Errorf("← from block should wrap to arrow: %q", line)
	}
	m, _ = press(m, runes("d"))
	if line := pickerLine(t, m, "selection"); !strings.Contains(line, "block") || !hasBlock(m) {
		t.Errorf("d should reset to block: %q, block %v", line, hasBlock(m))
	}
	if m, _ = press(m, runes("#")); m.mode != modeTheme {
		t.Error("# is for color rows only")
	}
}

// The icons row shows off (the default) or a set, with a hint. ←→ cycle
// off → nerd → dot → triangle → diamond → star (wrapping), the rows above show
// the icons right away, and d turns them off.
func TestPickerIconsRow(t *testing.T) {
	m := openPicker(t, colorsModel(t, config.Colors{}))
	if line := pickerLine(t, m, "icons"); !strings.Contains(line, "off") || !strings.Contains(line, "nerd needs a Nerd Font") {
		t.Errorf("want off and the hint: %q", line)
	}
	m = toRow(t, m, "icons")
	for _, step := range [][2]string{{"nerd", folderIcon}, {"dot", "●"}, {"triangle", "▸"}, {"diamond", "◆"}, {"star", "★"}} {
		m = right(m)
		if line := pickerLine(t, m, "icons"); !strings.Contains(line, step[0]) || !strings.Contains(m.screen(), step[1]+" NVIM") {
			t.Errorf("→ should show and preview %s: %q\n%s", step[0], line, plain(m.screen()))
		}
	}
	m = right(m)
	if line := pickerLine(t, m, "icons"); !strings.Contains(line, "off") || strings.Contains(m.screen(), "★") {
		t.Errorf("→ from star should wrap to off: %q\n%s", line, plain(m.screen()))
	}
	if line := pickerLine(t, left(m), "icons"); !strings.Contains(line, "star") {
		t.Errorf("← from off should wrap to star: %q", line)
	}
	m, _ = press(right(m), runes("d"))
	if line := pickerLine(t, m, "icons"); !strings.Contains(line, "off") || strings.Contains(m.screen(), folderIcon) {
		t.Errorf("d should turn icons off: %q\n%s", line, m.screen())
	}
	if m, _ = press(m, runes("#")); m.mode != modeTheme {
		t.Error("# is for color rows only")
	}
}

// Toggling icons keeps the cursor on its bookmark, and an applied filter
// stays applied, showing the same rows (now with icons).
func TestPickerIconsKeepCursorAndFilter(t *testing.T) {
	m := selectLabel(t, colorsModel(t, config.Colors{}), "NOTES")
	m = right(toRow(t, openPicker(t, m), "icons"))
	if got := selected(m); got != "NOTES" {
		t.Errorf("want the cursor still on NOTES, got %s", got)
	}

	m = colorsModel(t, config.Colors{})
	m.list.SetFilterText("notes")
	m = right(toRow(t, openPicker(t, m), "icons"))
	if m.list.FilterState() != list.FilterApplied || !slices.Equal(shown(m), []string{"NOTES"}) || selected(m) != "NOTES" {
		t.Errorf("want the filter kept with NOTES selected, got state %v, rows %q, selected %s",
			m.list.FilterState(), shown(m), selected(m))
	}
	if !strings.Contains(plain(m.screen()), folderIcon+" NOTES") {
		t.Errorf("want the filtered row with its icon:\n%s", m.screen())
	}
	m, _ = press(m, keyMsg("esc"))
	if m.list.FilterState() != list.FilterApplied || !slices.Equal(shown(m), []string{"NOTES"}) || strings.Contains(m.screen(), folderIcon) {
		t.Errorf("esc should drop the icons and keep the filter, got state %v, rows %q:\n%s",
			m.list.FilterState(), shown(m), m.screen())
	}
}

// enter saves selection and icons with the colors, keeping the bookmarks; the
// menu keeps them. Going back to block and off saves the defaults, which the
// file leaves out.
func TestPickerSavesMenuOptions(t *testing.T) {
	m := openPicker(t, colorsModel(t, config.Colors{Theme: "nord"}))
	m = right(toRow(t, m, "selection"))
	m = enter(right(toRow(t, m, "icons")))
	if m.mode != modeBrowse {
		t.Fatalf("enter should close the picker:\n%s", plain(m.screen()))
	}
	cfg, err := config.Load(m.configPath)
	if err != nil {
		t.Fatal(err)
	}
	if want := (config.Menu{Icons: "nerd", Selection: "bar"}); cfg.Menu != want || m.cfg.Menu != want {
		t.Errorf("want %+v saved and kept, got file %+v, menu %+v", want, cfg.Menu, m.cfg.Menu)
	}
	if cfg.Colors.Theme != "nord" || len(cfg.Bookmarks) != 2 {
		t.Errorf("saving should keep the colors and bookmarks, got %+v, %d bookmarks", cfg.Colors, len(cfg.Bookmarks))
	}
	if hasBlock(m) || !strings.Contains(m.screen(), folderIcon+" NVIM") {
		t.Errorf("after saving, the menu should keep bar and icons:\n%s", m.screen())
	}
	if !strings.Contains(plain(m.screen()), "Appearance saved") {
		t.Errorf("want a status:\n%s", plain(m.screen()))
	}

	m = openPicker(t, m)
	m = left(toRow(t, m, "selection"))
	m = enter(left(toRow(t, m, "icons")))
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		t.Fatal(err)
	}
	if m.cfg.Menu != (config.Menu{}) || strings.Contains(string(data), "selection") || strings.Contains(string(data), "icons") {
		t.Errorf("block and off are the defaults and shouldn't be written, got %+v:\n%s", m.cfg.Menu, data)
	}
}

// esc puts back the saved selection and icons, and saves nothing.
func TestPickerEscRestoresMenuOptions(t *testing.T) {
	m := openPicker(t, colorsModel(t, config.Colors{}))
	m = right(toRow(t, m, "selection"))
	m = right(toRow(t, m, "icons"))
	m, _ = press(m, keyMsg("esc"))
	if m.mode != modeBrowse || saved(m) {
		t.Errorf("esc should close the picker without saving; mode %v, saved %v", m.mode, saved(m))
	}
	if !hasBlock(m) || strings.Contains(m.screen(), folderIcon) || m.cfg.Menu != (config.Menu{}) {
		t.Errorf("esc should restore block and no icons:\n%s", m.screen())
	}
}

// A hand-edited selection the menu doesn't know shows as unknown and steps
// from block; one in another case counts as unchanged.
func TestPickerHandEditedSelection(t *testing.T) {
	m := openPicker(t, lookModel(t, config.Colors{}, config.Menu{Selection: "bars"}))
	if line := pickerLine(t, m, "selection"); !strings.Contains(line, "bars (unknown)") {
		t.Errorf("want the unknown value marked: %q", line)
	}
	if line := pickerLine(t, right(toRow(t, m, "selection")), "selection"); !strings.Contains(line, "bar") || strings.Contains(line, "unknown") {
		t.Errorf("→ from an unknown value should step from block to bar: %q", line)
	}

	m = openPicker(t, lookModel(t, config.Colors{}, config.Menu{Selection: "Bar"}))
	if line := pickerLine(t, m, "selection"); !strings.Contains(line, "bar") || strings.Contains(line, "unknown") {
		t.Errorf("Bar should show as bar: %q", line)
	}
	m = enter(m)
	if saved(m) || !strings.Contains(plain(m.screen()), "Appearance unchanged") {
		t.Errorf("Bar is unchanged and shouldn't save; saved %v:\n%s", saved(m), plain(m.screen()))
	}
}
