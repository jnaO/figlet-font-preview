package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/jnaO/figlet-font-preview/internal/action"
)

type helpState struct {
	open bool
	view viewport.Model
}

type paneKey struct{ keys, what string }

// paneKeys is the per-pane reference the help screen lists after the
// actions; the README Keys table is the full list.
var paneKeys = []struct {
	pane string
	keys []paneKey
}{
	{"Anywhere", []paneKey{
		{"tab shift+tab", "next, previous pane"},
		{"pgup pgdn", "scroll the preview a page"},
		{"ctrl+u ctrl+d", "scroll half a page"},
	}},
	{"Text", []paneKey{
		{"enter", "new line"},
		{"esc", "go to the Fonts list"},
	}},
	{"Fonts list", []paneKey{
		{"↑ ↓ j k g G", "move"},
		{"/", "jump to search"},
		{"esc", "clear the search"},
	}},
	{"Search", []paneKey{
		{"type", "fuzzy-filter fonts"},
		{"↑ ↓", "move in the list"},
		{"enter", "go to the list"},
		{"esc", "clear, go to the list"},
	}},
	{"Filters", []paneKey{
		{"↑ ↓ j k", "move"},
		{"space enter → l", "toggle, step Group by on"},
		{"← h", "step Group by back"},
		{"esc", "back to the list"},
	}},
	{"Style", []paneKey{
		{"↑ ↓ j k", "move"},
		{"space enter", "change the row"},
		{"← → h l", "change Align, Mode, slider"},
		{"esc", "go to the Fonts list"},
		{"ctrl+o", "close the pane"},
	}},
}

// openHelp shows the help screen; opening it closes the palette.
func (m *Model) openHelp() tea.Cmd {
	m.palette.open = false
	m.palette.input.Blur()
	m.text.Blur()
	m.search.Blur()
	m.help.open = true
	m.help.view = viewport.New(0, 0)
	m.syncHelp()
	return nil
}

// closeHelp restores the focus that opening blurred.
func (m *Model) closeHelp() tea.Cmd {
	m.help.open = false
	return m.setFocus(m.focus)
}

// helpKey handles a key while help is open: scrolling and closing first,
// then any app-wide key closes help and runs.
func (m Model) helpKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := msg.String()
	m.syncHelp()
	switch k {
	case "esc", "?":
		return m, m.closeHelp()
	case "up", "k":
		m.help.view.LineUp(1)
	case "down", "j":
		m.help.view.LineDown(1)
	case "pgup":
		m.help.view.ViewUp()
	case "pgdown":
		m.help.view.ViewDown()
	default:
		if e, ok := lookupKey(k); ok {
			return m, tea.Batch(m.closeHelp(), runEntry(&m, e, k))
		}
	}
	return m, nil
}

type helpGroup struct {
	heading string
	rows    []paneKey
}

type helpSection struct {
	label  string
	groups []helpGroup
}

// helpSections is the help list in order: Marks, Actions, Keys by pane.
func helpSections() []helpSection {
	marks := helpSection{"Marks", []helpGroup{{rows: []paneKey{
		{styles.cursor.Render("›"), "the cursor: copy and SVG use it"},
		{styles.selected.Render("●"), "a pick: narrows the preview"},
	}}}}
	acts := helpSection{label: "Actions"}
	for _, a := range action.All() {
		if n := len(acts.groups); n == 0 || acts.groups[n-1].heading != a.Group {
			acts.groups = append(acts.groups, helpGroup{heading: a.Group})
		}
		g := &acts.groups[len(acts.groups)-1]
		e, _ := binding(a.ID)
		g.rows = append(g.rows, paneKey{e.show, a.Title})
	}
	panes := helpSection{label: "Keys by pane"}
	for _, p := range paneKeys {
		panes.groups = append(panes.groups, helpGroup{p.pane, p.keys})
	}
	return []helpSection{marks, acts, panes}
}

// helpRows is every row of every section.
func helpRows() []paneKey {
	var rows []paneKey
	for _, s := range helpSections() {
		for _, g := range s.groups {
			rows = append(rows, g.rows...)
		}
	}
	return rows
}

// helpContent lays every section out in one key column and one description
// column; descriptions wrap inside avail columns and hang under the column.
func helpContent(avail int) string {
	kw := 0
	for _, r := range helpRows() {
		kw = max(kw, lipgloss.Width(r.keys))
	}
	dw := max(avail-4-kw, 1)
	hang := strings.Repeat(" ", 4+kw)
	var lines []string
	for i, s := range helpSections() {
		if i > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, styles.label.Render(s.label))
		for _, g := range s.groups {
			if g.heading != "" {
				lines = append(lines, "", styles.header.Render(g.heading))
			}
			for _, r := range g.rows {
				desc := strings.Split(ansi.Wrap(r.what, dw, ""), "\n")
				lines = append(lines, "  "+r.keys+strings.Repeat(" ", kw-lipgloss.Width(r.keys))+"  "+desc[0])
				for _, d := range desc[1:] {
					lines = append(lines, hang+d)
				}
			}
		}
	}
	return strings.Join(lines, "\n")
}

// syncHelp sizes the viewport to the screen and loads the content, and
// returns the box's outer width and height.
func (m *Model) syncHelp() (int, int) {
	avail := max(m.width-6, 1)
	content := helpContent(avail)
	inner := min(lipgloss.Width(content), avail)
	h := max(min(lipgloss.Height(content), m.bodyHeight()-2), 1)
	m.help.view.Width, m.help.view.Height = inner, h
	m.help.view.SetContent(content)
	return inner + 4, h + 2
}

func (m Model) helpView() (string, int, int) {
	w, h := m.syncHelp()
	return frame("Help", "?", m.help.view.View(), w, h-2, 1, 1, true), w, h
}
