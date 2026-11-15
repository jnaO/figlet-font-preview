package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// keyEntry binds app-wide keys (msg.String() values) to one action. A key
// matched here never reaches the focused component. show is the key text the
// footer prints; an empty show keeps the entry out of the footer.
type keyEntry struct {
	keys       []string
	show, help string
	run        func(m *Model, key string) tea.Cmd
}

// globalKeys is in footer order; the last entry is always shown.
var globalKeys = []keyEntry{
	{keys: []string{"ctrl+k"}}, // reserved: swallowed so the focused field never deletes text
	{keys: []string{"tab", "shift+tab"}, show: "tab", help: "pane", run: func(m *Model, k string) tea.Cmd {
		if k == "tab" {
			return m.setFocus(m.nextFocus(1))
		}
		return m.setFocus(m.nextFocus(-1))
	}},
	{keys: []string{"ctrl+o"}, show: "ctrl+o", help: "style", run: func(m *Model, _ string) tea.Cmd { return m.toggleStyle() }},
	{keys: []string{"ctrl+y"}, show: "ctrl+y", help: "copy", run: func(m *Model, _ string) tea.Cmd { return m.copyCursor() }},
	{keys: []string{"ctrl+s"}, show: "ctrl+s", help: "svg", run: func(m *Model, _ string) tea.Cmd { return m.exportCursor() }},
	{keys: []string{"ctrl+e"}, show: "ctrl+e", help: "text", run: func(m *Model, _ string) tea.Cmd { return m.setFocus(focusText) }},
	{keys: []string{"ctrl+f"}, show: "ctrl+f", help: "fonts", run: func(m *Model, _ string) tea.Cmd { return m.focusFonts(focusSearch) }},
	{keys: []string{"ctrl+t"}, show: "ctrl+t", help: "filters", run: func(m *Model, _ string) tea.Cmd { return m.focusFonts(focusFilters) }},
	{keys: []string{"pgup", "pgdown", "ctrl+u", "ctrl+d"}, show: "pgup/pgdn", help: "scroll", run: func(m *Model, k string) tea.Cmd {
		m.scrollPreview(k)
		return nil
	}},
	{keys: []string{"ctrl+c"}, show: "ctrl+c", help: "quit", run: func(*Model, string) tea.Cmd { return tea.Quit }},
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
		if e.show != "" {
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
