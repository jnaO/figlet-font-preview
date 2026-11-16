package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jnaO/figlet-font-preview/internal/action"
)

// keyEntry binds keys (msg.String() values) to one handler. A key matched in
// globalKeys never reaches the focused component. action is the shared action
// id, empty for navigation entries. show is the key text the footer and help
// print; help is the footer's short label, and an entry with an empty help
// stays out of the footer.
type keyEntry struct {
	keys       []string
	action     action.ID
	show, help string
	run        func(m *Model, key string) tea.Cmd
}

// globalKeys is in footer order; the last entry is always shown.
var globalKeys = []keyEntry{
	{keys: []string{"ctrl+k"}, action: action.Palette, show: "ctrl+k", help: "actions", run: func(m *Model, _ string) tea.Cmd { return m.openPalette() }},
	{keys: []string{"tab", "shift+tab"}, show: "tab", help: "pane", run: func(m *Model, k string) tea.Cmd {
		if k == "tab" {
			return m.setFocus(m.nextFocus(1))
		}
		return m.setFocus(m.nextFocus(-1))
	}},
	{keys: []string{"ctrl+o"}, action: action.ToggleStyle, show: "ctrl+o", help: "style", run: func(m *Model, _ string) tea.Cmd { return m.toggleStyle() }},
	{keys: []string{"ctrl+y"}, action: action.Copy, show: "ctrl+y", help: "copy", run: func(m *Model, _ string) tea.Cmd { return m.copyCursor() }},
	{keys: []string{"ctrl+s"}, action: action.ExportSVG, show: "ctrl+s", help: "svg", run: func(m *Model, _ string) tea.Cmd { return m.exportCursor() }},
	{keys: []string{"ctrl+e"}, action: action.EditText, show: "ctrl+e", help: "text", run: func(m *Model, _ string) tea.Cmd { return m.setFocus(focusText) }},
	{keys: []string{"ctrl+f"}, action: action.SearchFonts, show: "ctrl+f", help: "fonts", run: func(m *Model, _ string) tea.Cmd { return m.focusFonts(focusSearch) }},
	{keys: []string{"ctrl+t"}, action: action.Filters, show: "ctrl+t", help: "filters", run: func(m *Model, _ string) tea.Cmd { return m.focusFonts(focusFilters) }},
	{keys: []string{"pgup", "pgdown", "ctrl+u", "ctrl+d"}, show: "pgup/pgdn", help: "scroll", run: func(m *Model, k string) tea.Cmd {
		m.scrollPreview(k)
		return nil
	}},
	{keys: []string{"ctrl+c"}, action: action.Quit, show: "ctrl+c", help: "quit", run: func(*Model, string) tea.Cmd { return tea.Quit }},
}

// zoneKeys holds the actions whose key a zone handles itself. Their entries
// have no keys, so lookupKey never matches them; they run from the palette.
var zoneKeys []keyEntry

// init fills zoneKeys: help's handler reads binding, which reads zoneKeys.
func init() {
	zoneKeys = []keyEntry{
		{action: action.TogglePick, show: "space", run: func(m *Model, _ string) tea.Cmd { m.toggleCursor(); return nil }},
		{action: action.ClearPicks, show: "x", run: func(m *Model, _ string) tea.Cmd { m.clearPicks(); return nil }},
		{action: action.Help, show: "?", run: func(m *Model, _ string) tea.Cmd { return m.openHelp() }},
	}
}

// binding returns the entry bound to an action id.
func binding(id action.ID) (keyEntry, bool) {
	for _, tbl := range [][]keyEntry{globalKeys, zoneKeys} {
		for _, e := range tbl {
			if e.action == id {
				return e, true
			}
		}
	}
	return keyEntry{}, false
}

// runEntry runs an entry's handler; entries without one only swallow the key.
func runEntry(m *Model, e keyEntry, key string) tea.Cmd {
	if e.run == nil {
		return nil
	}
	return e.run(m, key)
}

func lookupKey(key string) (keyEntry, bool) {
	for _, e := range globalKeys {
		for _, k := range e.keys {
			if k == key {
				return e, true
			}
		}
	}
	return keyEntry{}, false
}

// footerKeys joins the shown entries with " · ", dropping whole entries from
// the end of the table (never the last) until the line fits width.
func footerKeys(width int) string {
	var shown []string
	for _, e := range globalKeys {
		if e.help != "" {
			shown = append(shown, e.show+": "+e.help)
		}
	}
	head, last := shown[:len(shown)-1], shown[len(shown)-1]
	line := last
	for i := range head {
		next := strings.Join(append(append([]string{}, head[:i+1]...), last), " · ")
		if len([]rune(next)) > width {
			break
		}
		line = next
	}
	return line
}

// focusFonts focuses a Fonts position, closing Style first when it has
// hidden the Fonts pane.
func (m *Model) focusFonts(f focus) tea.Cmd {
	var cmd tea.Cmd
	if m.fontsHidden() && m.panelOpen {
		cmd = m.toggleStyle()
	}
	return tea.Batch(cmd, m.setFocus(f))
}
