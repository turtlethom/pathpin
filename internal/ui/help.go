package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// The help page (?) fills the screen with every key, in titled blocks laid
// out as a grid. It's a mode of its own (modeHelp) whose View replaces the
// menu entirely: the list keeps its cursor, filter and size meanwhile, so
// closing the page puts everything back as it was.

// helpSection is one titled block of the help page.
type helpSection struct {
	title string
	keys  []hint
}

// helpSections are the help page's blocks, in reading order. They're the
// one list of every key: adding a key means adding its row to the block it
// belongs to. tab and N show without groups too, so the page never changes
// shape and shows that groups exist.
func (m Model) helpSections() []helpSection {
	menu := []hint{{"/", "filter"}}
	if m.list.FilterState() == list.FilterApplied {
		menu = append(menu, hint{"esc", "clear filter"})
	}
	menu = append(menu, hint{"t", "appearance"}, hint{"?", "help"}, hint{"q", "quit"})
	return []helpSection{
		{"Navigate", []hint{{"↑↓/jk", "move"}, {"←→/hl", "page"}, {"g/G", "top/end"}, {"tab/⇧tab", "group"}, {".", "groups only"}}},
		{"Bookmarks", []hint{{"a", "add"}, {"r", "remove"}, {"c", "clear all"}, {"J/K", "reorder"}}},
		{"Names", []hint{{"n", "rename"}, {"N", "rename group"}}},
		{"Groups", []hint{{"m", "move to group"}, {"u", "ungroup"}}},
		{"Actions", []hint{{"enter", "cd"}, {"e", "edit"}, {"y", "copy path"}, {"o", "open"}}},
		{"Menu", menu},
	}
}

// helpGap is the space between two blocks side by side.
const helpGap = 6

// updateHelp handles keys while the help page is open (modeHelp): ?, esc
// and q close it, and every other key does nothing, so nothing changes
// behind the page. (ctrl+c quits from any mode; see Update.)
func (m Model) updateHelp(msg tea.Msg) (tea.Model, tea.Cmd) {
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg) // status message timeouts
		return m, cmd
	}
	switch k.String() {
	case "?", "esc", "q":
		m.mode = modeBrowse
	}
	return m, nil
}

// helpPage draws the help page: the badge (where the menu draws it, so it
// doesn't move), the blocks as a grid, and the keys that close the page. A
// grid taller than the window is cut between rows of blocks, never through
// one, and a muted … takes the place of the blank line above the keys.
func (m Model) helpPage() string {
	indent := m.styles.tabBar // the blocks start under the badge's block, like the tabs
	width := m.width - m.styles.app.GetHorizontalFrameSize() - indent.GetHorizontalFrameSize()
	badge := m.titleBar()

	// Room for the grid: the window, but for the frame, the badge (with the
	// blank line under it), the line above the keys and the keys.
	room := max(0, m.height-m.styles.app.GetVerticalFrameSize()-lipgloss.Height(badge)-2)
	grid := strings.Split(m.helpGrid(width), "\n")
	sep := ""
	if len(grid) > room {
		// Keep the rows of blocks that fit whole: up to the last blank line
		// between rows within room, or none if even the first doesn't fit.
		keep := 0
		for i := range min(room+1, len(grid)) {
			if strings.TrimSpace(grid[i]) == "" {
				keep = i
			}
		}
		grid, sep = grid[:keep], m.styles.muted.Render("…")
	}
	page := sep + "\n" + m.hints(width, hint{"? esc q", "close"})
	if len(grid) > 0 {
		page = strings.Join(grid, "\n") + "\n" + page
	}
	return badge + "\n" + indent.Render(lipgloss.NewStyle().MaxWidth(max(width, 1)).Render(page))
}

// helpGrid lays the blocks out as many to a row as fit width (at least
// one), in reading order, with a blank line between rows. Every column is as
// wide as the widest block plus helpGap, so the blocks line up across rows.
func (m Model) helpGrid(width int) string {
	var blocks []string
	widest := 0
	for _, s := range m.helpSections() {
		b := m.helpBlock(s)
		blocks = append(blocks, b)
		widest = max(widest, lipgloss.Width(b))
	}
	// n blocks take n*widest + (n-1)*helpGap columns.
	perRow := max(1, (width+helpGap)/(widest+helpGap))
	column := lipgloss.NewStyle().Width(widest + helpGap)
	var rows []string
	for i := 0; i < len(blocks); i += perRow {
		row := blocks[i:min(i+perRow, len(blocks))]
		cells := make([]string, len(row))
		for j, b := range row {
			cells[j] = b
			if j < len(row)-1 {
				cells[j] = column.Render(b) // the last block needs no gap after it
			}
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cells...))
	}
	return strings.Join(rows, "\n\n")
}

// helpBlock draws one block: its title in the text color, a muted rule as
// long as the title, then a row per key, with the keys padded to the widest
// so the descriptions line up. Keys are in the accent and descriptions muted,
// like every key hint.
func (m Model) helpBlock(s helpSection) string {
	keyWidth := 0
	for _, h := range s.keys {
		keyWidth = max(keyWidth, lipgloss.Width(h.key))
	}
	lines := []string{
		m.styles.note.UnsetPadding().Render(s.title),
		m.styles.rule.Render(strings.Repeat("─", lipgloss.Width(s.title))),
	}
	for _, h := range s.keys {
		pad := strings.Repeat(" ", keyWidth-lipgloss.Width(h.key)+2)
		lines = append(lines, m.styles.key.Render(h.key)+pad+m.styles.muted.Render(h.desc))
	}
	return strings.Join(lines, "\n")
}
