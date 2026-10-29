package ui

import (
	"fmt"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/turtlethom/pathpin/v2/internal/config"
)

// The group picker (m) puts the highlighted bookmark in a group: a panel
// under the list with a row per group, then + new group…, which opens a field
// for the new group's name. It's built like the appearance picker (t): its
// own mode, a row index on the Model, and a field that belongs to it, so esc
// in the field steps back to the picker rather than out of it.

// newGroupItem is the group picker's last row.
const newGroupItem = "+ new group…"

// groupPickerKeys are the group picker's keys, on the bottom line. enter's
// hint says what it does on the highlighted row.
func (m Model) groupPickerKeys() []hint {
	enter := hint{"enter", "move"}
	if m.groupRow == len(m.cfg.Groups()) {
		enter = hint{"enter", "new group"}
	}
	return []hint{{"↑↓/jk", "pick"}, enter, {"esc", "cancel"}}
}

// startMove opens the group picker for the highlighted bookmark, on its group
// (or the first row if it has none; with no groups, that's + new group…).
func (m Model) startMove() (tea.Model, tea.Cmd) {
	it, ok := m.list.SelectedItem().(item)
	if !ok {
		return m, nil // empty list
	}
	names := m.cfg.Groups()
	m.mode, m.editing, m.inputErr, m.groupRow, m.groupTop = modeGroupPick, it.Label, "", 0, 0
	if g := m.cfg.GroupOf(it.Bookmark); g != "" {
		m.groupRow = slices.Index(names, g)
	}
	// Scroll the rows to show its group, and shrink the list to make room for
	// them (the help hides meanwhile).
	m.groupTop, _, _ = m.groupWindow(len(names))
	m.resize()
	return m, nil
}

// updateGroupPick handles keys while the group picker is open
// (modeGroupPick). The rows come from the config each time, never stored, so
// they can't go stale.
func (m Model) updateGroupPick(msg tea.Msg) (tea.Model, tea.Cmd) {
	k, ok := msg.(tea.KeyPressMsg)
	if !ok {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg) // status message timeouts
		return m, cmd
	}
	m.inputErr = "" // an error shows until the next key, as in the fields
	names := m.cfg.Groups()
	n := len(names) + 1 // the groups, then + new group…
	switch k.String() {
	case "up", "k":
		// Wraps to the bottom; adding n first keeps it from going negative.
		m.groupRow = (m.groupRow + n - 1) % n
	case "down", "j":
		m.groupRow = (m.groupRow + 1) % n
	case "enter":
		if m.groupRow == len(names) {
			return m.startNewGroup()
		}
		return m.move(names[m.groupRow])
	case "esc":
		return m.closePicker(), nil
	}
	// Remember where the rows are scrolled to, so they only move when the
	// cursor reaches an edge.
	m.groupTop, _, _ = m.groupWindow(len(names))
	return m, nil
}

// minListHeight is the fewest lines the list draws itself in: its title, its
// status, the highlighted row and the page dots. Given less, it draws that
// many anyway, and the menu would be taller than the window.
const minListHeight = 7

// panelRoom is how many lines a picker may use: what the window has left
// after the list at its shortest and the rest of the menu (as resize counts
// them; the help hides while picking).
func (m Model) panelRoom() int {
	_, v := m.styles.app.GetFrameSize()
	room := m.height - v - minListHeight - 1 - 1 // the bottom line, and the blank line above the picker
	if m.tabNames() != nil {
		room -= 3 // a blank line, the tab bar and its rule
	}
	return max(room, 4) // a row between the two scroll lines, and + new group…
}

// window is the part of n rows in view, in room lines with pinned more
// lines always under them: w rows from top. When they don't all fit, scrolls
// is true: two lines around the rows say how many are hidden, and top moves
// only as far as keeps the cursor's row in view (a cursor past the rows, on a
// pinned line, keeps top). Both pickers use it; it only does arithmetic, so
// it's easy to test and share.
func window(n, room, pinned, cursor, top int) (int, int, bool) {
	if n+pinned <= room {
		return 0, n, false
	}
	w := max(room-pinned-2, 1) // the two scroll lines
	if cursor < n {
		top = max(min(top, cursor), cursor-w+1)
	}
	return min(max(top, 0), n-w), w, true
}

// groupWindow is the part of the n groups in view (see window). + new group…
// is pinned under them, so it always shows.
func (m Model) groupWindow(n int) (top, w int, scrolls bool) {
	return window(n, m.panelRoom(), 1, m.groupRow, m.groupTop)
}

// startNewGroup opens the new-group field, empty. N shares the field, so its
// placeholder and suggestions are cleared.
func (m Model) startNewGroup() (tea.Model, tea.Cmd) {
	m.mode, m.inputErr = modeGroup, ""
	m.group.Reset()
	m.group.Prompt = "Group: "
	m.group.Placeholder = ""
	m.group.SetSuggestions(nil)
	m.group.SetWidth(12) // longer names scroll sideways
	return m, m.group.Focus()
}

// createGroup puts the bookmark in a new group named by the field. The field
// only makes new groups (the picker joins existing ones), so an existing name
// is refused, as is an empty one; SetGroup refuses the tab names. Errors keep
// the field open to fix.
func (m Model) createGroup(name string) (tea.Model, tea.Cmd) {
	name = strings.TrimSpace(name)
	if name == "" {
		return m, nil // nothing typed: the field stays open
	}
	names := m.cfg.Groups()
	if i := indexFold(names, name); i >= 0 {
		m.inputErr = names[i] + " already exists"
		return m, nil
	}
	return m.move(name)
}

// move puts the bookmark in group (one from the picker, or a new one from the
// field), like add: change a copy, save it, then swap it in. Errors show in
// place of the picker's keys, or beside the field, until the next key.
func (m Model) move(group string) (tea.Model, tea.Cmd) {
	before, _ := m.cfg.Find(m.editing)
	if current := m.cfg.GroupOf(before); current != "" && strings.EqualFold(group, current) {
		// Its own group: don't save, not even to re-case a second spelling.
		// Every save replaces config.toml.bak, so a save that changes nothing
		// would throw away the last real backup.
		// closePicker returns a copy, so close first: the status message
		// changes m.list, and must change the Model that's returned.
		m.group.Blur()
		m = m.closePicker()
		return m, m.list.NewStatusMessage(displayLabel(before.Label) + " stays in " + current)
	}
	next := m.edit()
	if err := next.SetGroup(m.editing, group); err != nil {
		m.inputErr = err.Error()
		return m, nil
	}
	if err := config.Save(m.configPath, next); err != nil {
		m.inputErr = "couldn't save: " + err.Error()
		return m, nil
	}

	m.cfg = next
	m.group.Blur()
	m = m.closePicker()
	moved, _ := next.Find(m.editing)
	status := "Moved " + displayLabel(moved.Label) + " to " + moved.Group

	// The moved row may leave the open tab; syncTabs then opens All if the tab
	// emptied. Go evaluates the arguments left to right, so rows come first.
	return m, tea.Batch(m.reloadRows(moved.Path), m.list.NewStatusMessage(status), m.syncTabs())
}

// groupPickerView draws the group picker's rows: the groups in view, then
// + new group…, with "↑ 3 more" and "↓ 5 more" lines around the groups when
// they scroll.
func (m Model) groupPickerView() string {
	b, _ := m.cfg.Find(m.editing)
	current := m.cfg.GroupOf(b)
	names := m.cfg.Groups()
	top, w, scrolls := m.groupWindow(len(names))
	var lines []string
	if scrolls {
		lines = append(lines, m.moreLine("↑", top, 4))
	}
	for i := top; i < top+w; i++ {
		lines = append(lines, m.groupLine(i, names[i], names[i] == current))
	}
	if scrolls {
		lines = append(lines, m.moreLine("↓", len(names)-top-w, 4))
	}
	lines = append(lines, m.groupLine(len(names), newGroupItem, false))
	return m.styles.picker.MaxWidth(m.list.Width()).Render(strings.Join(lines, "\n"))
}

// groupLine is row i of the group picker: the cursor (›) if it's highlighted,
// a mark (•) if it's the bookmark's group, each in a column of its own so the
// names line up, and + new group… muted, since it's an action.
func (m Model) groupLine(i int, name string, marked bool) string {
	cursor, mark, style := "  ", "  ", m.styles.note.UnsetPadding()
	if marked {
		mark = m.styles.key.Render("• ")
	}
	if name == newGroupItem {
		style = m.styles.muted
	}
	if i == m.groupRow {
		cursor, style = m.styles.picked.Render("› "), m.styles.picked
	}
	return cursor + mark + style.Render(name)
}

// moreLine says how many rows are hidden past one end of a picker, indented
// to line up with its names. It's blank when none are, so the picker's height
// stays the same while it scrolls.
func (m Model) moreLine(arrow string, hidden, indent int) string {
	if hidden == 0 {
		return ""
	}
	return m.styles.muted.Render(fmt.Sprintf("%s%s %d more", strings.Repeat(" ", indent), arrow, hidden))
}
