package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jnaO/figlet-font-preview/internal/action"
)

type helpState struct {
	open bool
	view viewport.Model
}

type paneKey struct{ keys, what string }

// paneKeys is the per-pane reference the help screen shows beside the
// actions; the README Keys table is the full list.
var paneKeys = []struct {
	pane string
	keys []paneKey
}{
	{"Anywhere", []paneKey{
		{"tab shift+tab", "next, previous pane"},
		{"pgup pgdn ctrl+u ctrl+d", "scroll the preview"},
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

const helpSentence = "› is the cursor: copy and SVG use it. ● is a pick: it narrows the preview."

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

// helpActionLines is the Actions column: each group's heading, then its
// actions with their keys.
func helpActionLines() []string {
	acts := action.All()
	kw := 0
	for _, a := range acts {
		e, _ := binding(a.ID)
		kw = max(kw, lipgloss.Width(e.show))
	}
	lines := []string{styles.label.Render("Actions")}
	group := ""
	for _, a := range acts {
		if a.Group != group {
			group = a.Group
			lines = append(lines, "", styles.header.Render(group))
		}
		e, _ := binding(a.ID)
		lines = append(lines, "  "+e.show+strings.Repeat(" ", kw-lipgloss.Width(e.show))+"  "+a.Title)
	}
	return lines
}

// helpPaneLines is the Keys by pane column.
func helpPaneLines() []string {
	kw := 0
	for _, p := range paneKeys {
		for _, k := range p.keys {
			kw = max(kw, lipgloss.Width(k.keys))
		}
	}
	lines := []string{styles.label.Render("Keys by pane")}
	for _, p := range paneKeys {
		lines = append(lines, "", styles.header.Render(p.pane))
		for _, k := range p.keys {
			lines = append(lines, "  "+k.keys+strings.Repeat(" ", kw-lipgloss.Width(k.keys))+"  "+k.what)
		}
	}
	return lines
}

// helpContent lays the two columns side by side when both fit in avail
// columns, else stacks them, and appends the legend.
func helpContent(avail int) string {
	left, right := strings.Join(helpActionLines(), "\n"), strings.Join(helpPaneLines(), "\n")
	var cols string
	if lipgloss.Width(left)+3+lipgloss.Width(right) <= avail {
		cols = lipgloss.JoinHorizontal(lipgloss.Top, left, "   ", right)
	} else {
		cols = left + "\n\n" + right
	}
	w := max(lipgloss.Width(cols), lipgloss.Width(legendLine()))
	w = min(w, avail)
	note := styles.dim.Width(w).Render(helpSentence + " esc closes; ↑ ↓ scroll.")
	return cols + "\n\n" + legendLine() + "\n" + note
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
