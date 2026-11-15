package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"

	"github.com/jnaO/figlet-font-preview/internal/action"
)

const paletteWidth = 56

// paletteRow is one listed action and the title indexes the query matched.
type paletteRow struct {
	act     action.Action
	matched []int
}

type paletteState struct {
	open           bool
	input          textinput.Model
	rows           []paletteRow
	cursor, offset int
}

func newPaletteInput() textinput.Model {
	in := textinput.New()
	in.Prompt = "> "
	in.PromptStyle = styles.label
	in.Placeholder = "type to filter actions"
	return in
}

// paletteOrder is the unfiltered list, common actions first; the palette
// does not list itself.
var paletteOrder = []action.ID{
	action.Copy, action.ExportSVG, action.EditText, action.SearchFonts, action.Filters,
	action.ToggleStyle, action.TogglePick, action.ClearPicks, action.Help, action.Quit,
}

func paletteActions() []action.Action {
	out := make([]action.Action, 0, len(paletteOrder))
	for _, id := range paletteOrder {
		a, _ := action.Find(id)
		out = append(out, a)
	}
	return out
}

// openPalette blurs the text fields without changing m.focus.
func (m *Model) openPalette() tea.Cmd {
	m.help.open = false
	m.text.Blur()
	m.search.Blur()
	m.palette.input = newPaletteInput()
	m.palette.input.Width = max(m.paletteInner()-lipgloss.Width(m.palette.input.Prompt)-1, 1)
	m.palette.open = true
	m.filterPalette()
	return m.palette.input.Focus()
}

// closePalette restores the focus that opening blurred.
func (m *Model) closePalette() tea.Cmd {
	m.palette.open = false
	m.palette.input.Blur()
	return m.setFocus(m.focus)
}

// filterPalette lists every action in menu order for an empty query, else
// the fuzzy matches in rank order.
func (m *Model) filterPalette() {
	acts := paletteActions()
	m.palette.rows = m.palette.rows[:0]
	if q := m.palette.input.Value(); q == "" {
		for _, a := range acts {
			m.palette.rows = append(m.palette.rows, paletteRow{act: a})
		}
	} else {
		titles := make([]string, len(acts))
		for i, a := range acts {
			titles[i] = a.Title
		}
		for _, r := range fuzzy.Find(q, titles) {
			m.palette.rows = append(m.palette.rows, paletteRow{acts[r.Index], r.MatchedIndexes})
		}
	}
	m.palette.cursor, m.palette.offset = 0, 0
}

func (m *Model) movePalette(d int) {
	p := &m.palette
	p.cursor = max(0, min(p.cursor+d, len(p.rows)-1))
	h := m.paletteRows()
	p.cursor = max(p.cursor, 0)
	if p.cursor < p.offset {
		p.offset = p.cursor
	}
	if p.cursor >= p.offset+h {
		p.offset = p.cursor - h + 1
	}
}

// paletteKey handles a key while the palette is open: its own keys first,
// then any app-wide key closes it and runs.
func (m Model) paletteKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := msg.String()
	switch k {
	case "esc", "ctrl+k":
		return m, m.closePalette()
	case "enter":
		var e keyEntry
		var ok bool
		if m.palette.cursor < len(m.palette.rows) {
			e, ok = binding(m.palette.rows[m.palette.cursor].act.ID)
		}
		cmd := m.closePalette()
		if ok && e.run != nil {
			cmd = tea.Batch(cmd, e.run(&m, k))
		}
		return m, cmd
	case "up":
		m.movePalette(-1)
		return m, nil
	case "down":
		m.movePalette(1)
		return m, nil
	}
	if e, ok := lookupKey(k); ok {
		return m, tea.Batch(m.closePalette(), runEntry(&m, e, k))
	}
	var cmd tea.Cmd
	m.palette.input, cmd = m.palette.input.Update(msg)
	m.filterPalette()
	return m, cmd
}

// paletteBox is the box's outer width.
func (m Model) paletteBox() int { return max(min(paletteWidth, m.width-2), 6) }

// paletteInner is the content width inside the frame and its padding.
func (m Model) paletteInner() int { return m.paletteBox() - 4 }

// paletteRows is the number of action rows shown; it does not depend on the query.
func (m Model) paletteRows() int { return max(min(len(paletteActions()), m.bodyHeight()-3), 1) }

func (m Model) paletteView() string {
	inner := m.paletteInner()
	lines := []string{m.palette.input.View()}
	p := m.palette
	end := min(p.offset+m.paletteRows(), len(p.rows))
	for i := p.offset; i < end; i++ {
		lines = append(lines, paletteLine(p.rows[i], i == p.cursor, inner))
	}
	if len(p.rows) == 0 {
		lines = append(lines, styles.dim.Render("  No matching actions"))
	}
	return frame("Actions", "ctrl+k", strings.Join(lines, "\n"), m.paletteBox(), 1+m.paletteRows(), 1, 1, true)
}

// paletteLine is "› Title" with the matched runes styled and the action's
// key right-aligned.
func paletteLine(r paletteRow, cursor bool, width int) string {
	key, _ := binding(r.act.ID)
	title := truncate(r.act.Title, max(width-2-len(key.show)-1, 1))
	pad := max(width-2-len([]rune(title))-len(key.show), 1)
	mark := "  "
	if cursor {
		mark = styles.cursor.Render("›") + " "
	}
	return mark + highlight(title, r.matched) + strings.Repeat(" ", pad) + styles.dim.Render(key.show)
}
