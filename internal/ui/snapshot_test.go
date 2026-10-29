package ui

// Temporary, for the Bubble Tea v2 migration (see CLAUDE.md, After v2.0.0):
// each screen below is saved in testdata/snapshots as runs of text with their
// style, decoded from the escape codes, so the migration can show that every
// mode still draws the same even if Lip Gloss v2 writes different escape
// bytes for it. Delete this file and testdata/snapshots when the migration
// merges.
//
// After an intended change to how the menu looks:
//
//	go test ./internal/ui -run TestSnapshots -update-snapshots

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"

	"github.com/turtlethom/pathpin/v2/internal/config"
)

var updateSnapshots = flag.Bool("update-snapshots", false, "rewrite testdata/snapshots from the current screens")

// snapshotScreens are the screens to keep: every mode (the merge and ungroup
// prompts share the remove prompt's style), an error, a missing row selected
// and not, both kinds of theme (terminal's ANSI colors, nord's hex), each
// header and selection style, and a narrow window.
func snapshotScreens() map[string]func(t *testing.T) Model {
	keys := func(m Model, ks ...tea.KeyMsg) Model {
		for _, k := range ks {
			m, _ = press(m, k)
		}
		return m
	}
	return map[string]func(t *testing.T) Model{
		"browse-terminal": func(t *testing.T) Model { return twoBookmarks(t) },
		"browse-groups":   func(t *testing.T) Model { return groupModel(t) },
		"group-tab":       func(t *testing.T) Model { return keys(groupModel(t), keyMsg("tab")) },
		"groups-only":     func(t *testing.T) Model { return keys(groupModel(t), runes(".")) },
		"filter-typing":   func(t *testing.T) Model { return keys(groupModel(t), runes("/"), runes("cs")) },
		"filter-applied":  func(t *testing.T) Model { return keys(groupModel(t), runes("/"), runes("cs"), keyMsg("enter")) },
		"missing-row": func(t *testing.T) Model {
			return nordModel(t, config.Bookmark{Label: "NOTES", Path: "notes"}, config.Bookmark{Label: "OLD", Path: "!old"})
		},
		"empty":          func(t *testing.T) Model { return nordModel(t) },
		"confirm-remove": func(t *testing.T) Model { return keys(groupModel(t), runes("r")) },
		"add-field": func(t *testing.T) Model {
			add, _ := launchFrom(t, twoBookmarks(t), "projects/app")
			return keys(add, runes("a"))
		},
		"rename-field":      func(t *testing.T) Model { return keys(groupModel(t), runes("n")) },
		"group-picker":      func(t *testing.T) Model { return keys(groupModel(t), runes("m")) },
		"appearance-picker": func(t *testing.T) Model { return keys(groupModel(t), runes("t")) },
		"hex-field":         func(t *testing.T) Model { return keys(groupModel(t), runes("t"), keyMsg("down"), runes("#")) },
		"help-page":         func(t *testing.T) Model { return keys(groupModel(t), runes("?")) },
		"header-text":       func(t *testing.T) Model { return headerModel(t, "text") },
		"header-bracket":    func(t *testing.T) Model { return headerModel(t, "bracket") },
		"header-rule":       func(t *testing.T) Model { return headerModel(t, "rule") },
		"selection-bar":     func(t *testing.T) Model { return restartedWith(t, groupModel(t), `selection = "bar"`) },
		"selection-arrow":   func(t *testing.T) Model { return restartedWith(t, groupModel(t), `selection = "arrow"`) },
		"icons-nerd":        func(t *testing.T) Model { return restartedWith(t, groupModel(t), `icons = "nerd"`) },
		"narrow":            func(t *testing.T) Model { return groupModel(t).sized(40, 20) },
		"missing-selected": func(t *testing.T) Model {
			m := nordModel(t, config.Bookmark{Label: "NOTES", Path: "notes"}, config.Bookmark{Label: "OLD", Path: "!old"})
			return keys(m, keyMsg("down"))
		},
		"field-error": func(t *testing.T) Model {
			return keys(clearField(keys(groupModel(t), runes("n"))), runes("cs310"), keyMsg("enter"))
		},
		"new-group-field":            func(t *testing.T) Model { return keys(groupModel(t), runes("m"), keyMsg("up"), keyMsg("enter")) },
		"appearance-picker-terminal": func(t *testing.T) Model { return keys(twoBookmarks(t), runes("t")) },
	}
}

func TestSnapshots(t *testing.T) {
	withColor(t)
	for name, build := range snapshotScreens() {
		t.Run(name, func(t *testing.T) {
			got := snapshot(build(t).screen())
			path := filepath.Join("testdata", "snapshots", name+".txt")
			if *updateSnapshots {
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("%v (run with -update-snapshots to create it)", err)
			}
			if got != string(want) {
				t.Errorf("%s differs from %s:\n%s", name, path, firstDiff(string(want), got))
			}
		})
	}
}

// firstDiff shows the first line where want and got differ.
func firstDiff(want, got string) string {
	w, g := strings.Split(want, "\n"), strings.Split(got, "\n")
	for i := range max(len(w), len(g)) {
		var wl, gl string
		if i < len(w) {
			wl = w[i]
		}
		if i < len(g) {
			gl = g[i]
		}
		if wl != gl {
			return fmt.Sprintf("line %d:\nwant %s\n got %s", i+1, wl, gl)
		}
	}
	return ""
}

// csi matches an escape sequence: SGR (…m) sets the style; any other kind is
// kept in the snapshot as it is, so a new one shows up as a difference.
var csi = regexp.MustCompile(`^\x1b\[([0-9;:?]*)([A-Za-z])`)

// snapStyle is the style a run of text is drawn in, by meaning rather than by
// escape bytes.
type snapStyle struct {
	fg, bg                              string
	bold, faint, italic, under, reverse bool
}

func (s snapStyle) String() string {
	var parts []string
	for _, p := range []struct {
		on   bool
		name string
	}{{s.fg != "", "fg=" + s.fg}, {s.bg != "", "bg=" + s.bg}, {s.bold, "bold"}, {s.faint, "faint"},
		{s.italic, "italic"}, {s.under, "under"}, {s.reverse, "reverse"}} {
		if p.on {
			parts = append(parts, p.name)
		}
	}
	return strings.Join(parts, ",")
}

// snapshot writes a screen as lines of runs: unstyled text as it is, styled
// text as «style|text». Trailing unstyled spaces and blank lines at the end
// are dropped, so padding that doesn't show doesn't count.
func snapshot(raw string) string {
	var lines []string
	var st snapStyle
	for line := range strings.SplitSeq(raw, "\n") {
		var b strings.Builder
		var run strings.Builder
		runStyle := st
		flush := func() {
			if run.Len() == 0 {
				return
			}
			if s := runStyle.String(); s != "" {
				fmt.Fprintf(&b, "«%s|%s»", s, run.String())
			} else {
				b.WriteString(run.String())
			}
			run.Reset()
		}
		for len(line) > 0 {
			if m := csi.FindStringSubmatch(line); m != nil {
				line = line[len(m[0]):]
				if m[2] != "m" {
					flush()
					fmt.Fprintf(&b, "⟨CSI %s%s⟩", m[1], m[2])
					continue
				}
				st = applySGR(st, m[1])
				continue
			}
			// A run ends where the style a character is drawn in changes, not
			// at every escape code: lipgloss may reset and set the same style
			// around each character (underlines), or once around many.
			if st != runStyle {
				flush()
				runStyle = st
			}
			r, size := utf8.DecodeRuneInString(line)
			run.WriteRune(r)
			line = line[size:]
		}
		flush()
		lines = append(lines, strings.TrimRight(b.String(), " "))
	}
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n") + "\n"
}

// applySGR applies one SGR sequence's parameters to st. Colors are named the
// same way whichever form wrote them: ansi:N (0-15), 256:N, or #rrggbb.
func applySGR(st snapStyle, params string) snapStyle {
	ps := strings.Split(params, ";") // ESC[m has one empty parameter
	for i := 0; i < len(ps); i++ {
		p := ps[i]
		// 38:2::r:g:b and 4:3 are the colon forms of the same parameters.
		sub := strings.Split(p, ":")
		n, _ := strconv.Atoi(sub[0])
		switch {
		case p == "0" || p == "": // an empty parameter is 0: reset
			st = snapStyle{}
		case n == 1:
			st.bold = true
		case n == 2:
			st.faint = true
		case n == 3:
			st.italic = true
		case n == 4:
			st.under = len(sub) < 2 || sub[1] != "0"
		case n == 7:
			st.reverse = true
		case n == 22:
			st.bold, st.faint = false, false
		case n == 23:
			st.italic = false
		case n == 24:
			st.under = false
		case n == 27:
			st.reverse = false
		case n >= 30 && n <= 37:
			st.fg = fmt.Sprintf("ansi:%d", n-30)
		case n >= 90 && n <= 97:
			st.fg = fmt.Sprintf("ansi:%d", n-90+8)
		case n == 39:
			st.fg = ""
		case n >= 40 && n <= 47:
			st.bg = fmt.Sprintf("ansi:%d", n-40)
		case n >= 100 && n <= 107:
			st.bg = fmt.Sprintf("ansi:%d", n-100+8)
		case n == 49:
			st.bg = ""
		case n == 38 || n == 48:
			var c string
			if len(sub) > 1 { // colon form: 38:5:n or 38:2:[colorspace]:r:g:b
				c = extColor(sub[1:], true)
			} else {
				c = extColor(ps[i+1:], false)
				if len(ps) > i+1 && ps[i+1] == "5" {
					i += 2
				} else {
					i += 4
				}
			}
			if n == 38 {
				st.fg = c
			} else {
				st.bg = c
			}
		default:
			st.fg += "?" + p // unknown: make it show
		}
	}
	return st
}

// extColor names an extended color from the parameters after 38 or 48:
// 5;n (256 colors) or 2;r;g;b (24-bit). In the colon form (38:2::r:g:b) a
// color space may come before r, g and b.
func extColor(ps []string, colon bool) string {
	if len(ps) >= 2 && ps[0] == "5" {
		n, _ := strconv.Atoi(ps[1])
		if n < 16 {
			return fmt.Sprintf("ansi:%d", n)
		}
		return "256:" + ps[1]
	}
	if len(ps) >= 4 && ps[0] == "2" {
		rgb := ps[1:4]
		if colon && len(ps) >= 5 {
			rgb = ps[2:5]
		}
		var v [3]int
		for j, s := range rgb {
			v[j], _ = strconv.Atoi(s)
		}
		return fmt.Sprintf("#%02x%02x%02x", v[0], v[1], v[2])
	}
	return "?" + strings.Join(ps, ";")
}
