package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"

	"github.com/turtlethom/pathpin/internal/config"
)

// Styles holds every style the menu uses. Build it with NewStyles and pass it
// in; there are no package-level styles. The model keeps copies in several
// places (the list, the delegate, each text field), so restyle in model.go
// updates all of them when the picker changes a color.
type Styles struct {
	app     lipgloss.Style // frame around the whole menu
	title   lipgloss.Style
	status  lipgloss.Style
	empty   lipgloss.Style
	filter  lipgloss.Style         // filter prompt and cursor
	prompt  lipgloss.Style         // y/N prompts on the bottom line
	note    lipgloss.Style         // messages on the bottom line
	errText lipgloss.Style         // errors on the bottom line
	muted   lipgloss.Style         // hints and placeholders
	key     lipgloss.Style         // the keys in key hints
	help    lipgloss.Style         // the help under the list
	item    list.DefaultItemStyles // bubbles' row styles: Normal, Selected, Dimmed (while typing a filter)
	missing list.DefaultItemStyles // rows whose folder is gone
	pathEnd lipgloss.TerminalColor // a path's last folder: the label color
	tabBar  lipgloss.Style         // the line of group tabs, and the rule under it
	tab     lipgloss.Style         // a tab
	tabOpen lipgloss.Style         // the open tab
	rule    lipgloss.Style         // the line under the tabs
	picker  lipgloss.Style         // a picker's panel (t, m)
	picked  lipgloss.Style         // the picker's highlighted row

	header     string         // [menu] header, as one of the header* names
	headerWord lipgloss.Style // the title and open tab's word when it isn't a block
}

// The ways to show the selected row, set with [menu] selection. Empty means
// block.
const (
	selectBlock = "block" // the label is a block of selected_text on primary (fcd's look)
	selectBar   = "bar"   // just the bar; the label is bold in primary
	selectArrow = "arrow" // a › where the bar is (like the pickers' cursor); the label as bar's
)

// KnownSelection reports whether s is a [menu] selection the menu knows, in
// any case. Empty counts, since it means block; anything else unknown gives
// block too, and cmd warns about it.
func KnownSelection(s string) bool {
	return s == "" || indexFold([]string{selectBlock, selectBar, selectArrow}, s) >= 0
}

// The ways to draw the title and the open tab, set with [menu] header. Empty
// means badge.
const (
	headerBadge   = "badge"   // a block of selected_text on primary
	headerText    = "text"    // bold primary
	headerBracket = "bracket" // [ PATHPIN ] and [All], the brackets muted
	headerRule    = "rule"    // bold primary; the title line ends in a muted rule, the open tab underlined
)

// headerNames are the header styles. A function, not a var, so there's no
// package-level state.
func headerNames() []string { return []string{headerBadge, headerText, headerBracket, headerRule} }

// KnownHeader reports whether h is a [menu] header the menu knows, in any
// case. Empty counts, since it means badge; anything else unknown gives badge
// too, and cmd warns about it.
func KnownHeader(h string) bool {
	return h == "" || indexFold(headerNames(), h) >= 0
}

// NewStyles builds the menu's styles from the user's theme (default:
// terminal, which follows the terminal's own palette), then applies any colors
// set in [colors] on top, one at a time:
//
//   - primary: accent (selected row's background, title, filter).
//   - secondary: normal text.
//   - tertiary: muted (descriptions, missing rows, status bar).
//   - selected_text: text on the accent.
//   - error: errors on the bottom line.
//
// An unknown theme gives terminal, and a value that isn't a color is ignored;
// cmd warns about both. menu's selection and header pick how the selected row,
// the title and the open tab look; unknown values give block and badge.
func NewStyles(c config.Colors, menu config.Menu) Styles {
	theme, _ := lookupTheme(c.Theme)
	primary := pick(c.Primary, theme.accent)
	onPrimary := pick(c.SelectedText, theme.onAccent)
	text := pick(c.Secondary, theme.text)
	muted := pick(c.Tertiary, theme.muted)
	errColor := pick(c.Error, theme.err)

	// A lipgloss.Style is a value, and each method returns a changed copy, so
	// bar and plain are bases: bar.Foreground(x) leaves bar itself alone.
	bar := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(primary).
		Padding(0, 0, 0, 1)
	plain := lipgloss.NewStyle().Padding(0, 0, 0, 2)

	// The selected label is a block on the accent color (padding included, so
	// one space shows on each side), and its path is underlined. The list cuts
	// titles to fit NormalTitle's padding, so it gets the same right padding
	// (invisible without a background) or the block would overflow the window
	// by a column. In bar mode the label is just in the accent color.
	selected := bar.Foreground(onPrimary).Background(primary).Bold(true).PaddingRight(1)
	selectedDesc := bar.Foreground(muted).Underline(true)
	switch {
	case strings.EqualFold(menu.Selection, selectBar):
		selected = bar.Foreground(primary).Bold(true).PaddingRight(1)
	case strings.EqualFold(menu.Selection, selectArrow):
		// A border whose left side is ›, so the arrow sits where the bar is,
		// on the label line only; the path keeps the bar's width as padding.
		arrow := lipgloss.Border{Left: "›"}
		selected = bar.Border(arrow, false, false, false, true).Foreground(primary).Bold(true).PaddingRight(1)
		selectedDesc = plain.Foreground(muted).Underline(true)
	}
	item := list.DefaultItemStyles{
		NormalTitle:   plain.Foreground(text).PaddingRight(1),
		NormalDesc:    plain.Foreground(muted),
		SelectedTitle: selected,
		SelectedDesc:  selectedDesc,
		DimmedTitle:   plain.Foreground(muted).PaddingRight(1), // like NormalTitle, so a tag after it keeps its space
		DimmedDesc:    plain.Foreground(muted),
		FilterMatch:   lipgloss.NewStyle().Underline(true),
	}
	// The block means "enter goes here", so a missing row never gets it. In
	// bar mode there's no block to remove, so missing rows look the same.
	missing := item
	missing.NormalTitle = item.NormalTitle.Foreground(muted)
	missing.SelectedTitle = item.SelectedTitle.UnsetBackground().Foreground(muted)

	// The title and the open tab: a block (badge), or bold primary with the
	// same padding, so switching styles moves nothing; rule also underlines
	// the open tab. bracket draws its brackets in titleText and openTab, where
	// the padding would be, so it moves nothing either.
	header := headerBadge
	if i := indexFold(headerNames(), menu.Header); i >= 0 {
		header = headerNames()[i]
	}
	word := lipgloss.NewStyle().Foreground(primary).Bold(true)
	title := lipgloss.NewStyle().Foreground(onPrimary).Background(primary).Bold(true).Padding(0, 1) // a badge, like the selected row
	tabOpen := title
	switch header {
	case headerText, headerBracket:
		title, tabOpen = word.Padding(0, 1), word.Padding(0, 1)
	case headerRule:
		title, tabOpen = word.Padding(0, 1), word.Underline(true).Padding(0, 1)
	}

	return Styles{
		header:     header,
		headerWord: word,
		// No right padding: resize keeps the list a column short of the right
		// edge, and only the help line may use that column.
		app:     lipgloss.NewStyle().Padding(1, 0, 0, 1),
		title:   title,
		status:  lipgloss.NewStyle().Foreground(muted).Padding(0, 0, 1, 2),
		empty:   lipgloss.NewStyle().Foreground(muted).Padding(0, 0, 0, 2),
		filter:  lipgloss.NewStyle().Foreground(primary),
		prompt:  lipgloss.NewStyle().Foreground(primary).Bold(true).Padding(0, 0, 0, 2),
		note:    lipgloss.NewStyle().Foreground(text).Padding(0, 0, 0, 2),
		errText: lipgloss.NewStyle().Foreground(errColor).Padding(0, 0, 0, 2),
		muted:   lipgloss.NewStyle().Foreground(muted),
		key:     lipgloss.NewStyle().Foreground(primary),
		help:    lipgloss.NewStyle().Padding(1, 0, 0, 2), // a blank line, then in line with the labels
		item:    item,
		missing: missing,
		pathEnd: text,
		// The open tab is a block like the title badge, and lines up with it.
		// Every tab has the same padding, so switching tabs moves nothing.
		tabBar:  lipgloss.NewStyle().Padding(0, 0, 0, 1),
		tab:     lipgloss.NewStyle().Foreground(muted).Padding(0, 1),
		tabOpen: tabOpen,
		rule:    lipgloss.NewStyle().Foreground(muted),
		picker:  lipgloss.NewStyle().MarginTop(1), // its › sits where the selected row's bar is
		picked:  lipgloss.NewStyle().Foreground(primary).Bold(true),
	}
}

// pick is the color set in [colors], or fallback if it's unset or not a color.
func pick(set string, fallback lipgloss.TerminalColor) lipgloss.TerminalColor {
	if set != "" && validColor(set) {
		return lipgloss.Color(set)
	}
	return fallback
}

// titleText is PATHPIN in the header style. (rule's line isn't part of it:
// View adds it after whatever else is on the title line; see withRule.)
func (s Styles) titleText() string {
	if s.header == headerBracket {
		return s.muted.Render("[") + s.headerWord.Render(menuTitle) + s.muted.Render("]")
	}
	return s.title.Render(menuTitle)
}

// openTab draws the open tab in the header style, as wide as the other tabs
// (one column each side), so switching tabs moves nothing.
func (s Styles) openTab(name string) string {
	if s.header == headerBracket {
		return s.muted.Render("[") + s.headerWord.Render(name) + s.muted.Render("]")
	}
	return s.tabOpen.Render(name)
}
