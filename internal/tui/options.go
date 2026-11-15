package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jnaO/figlet-font-preview/internal/catalog"
	"github.com/jnaO/figlet-font-preview/internal/display"
	"github.com/jnaO/figlet-font-preview/internal/fontclass"
	"github.com/jnaO/figlet-font-preview/internal/paint"
)

type optKind int

const (
	optAlign optKind = iota
	optCollapse
	optGroupBy
	optSimilar
	optAxis    // non-selectable axis label
	optHeading // non-selectable free-text label
	optSlider
	optBgMode
	optFacet
	optClear
)

// optRow is one row of the Fonts pane filters or the Style pane.
type optRow struct {
	kind   optKind
	axis   fontclass.Axis
	value  int
	label  string // optHeading
	slider slider // optSlider
}

func (r optRow) selectable() bool { return r.kind != optAxis && r.kind != optHeading }

const (
	styleWidth = 32 // Style pane content columns, excluding its frame
	slideStep  = 8  // columns the pane grows or shrinks per frame
	slideEvery = 16 * time.Millisecond

	minPreviewWidth = 10
	fontsCols       = sidebarWidth + 2 // left column plus its borders
	styleFrameCols  = 3                // Style pane's two borders and one column of right padding
	styleCols       = styleWidth + styleFrameCols
	frameCols       = 2 // the preview's left padding and the free last column
	// bothFitWidth is the narrowest terminal that shows Fonts, the minimum
	// preview and Style; styleFitWidth is the narrowest that shows Style and
	// the minimum preview.
	bothFitWidth  = fontsCols + frameCols + minPreviewWidth + styleCols
	styleFitWidth = frameCols + minPreviewWidth + styleCols
)

// slideMsg advances the Style pane's slide in or out.
type slideMsg struct{}

func slide() tea.Cmd {
	return tea.Tick(slideEvery, func(time.Time) tea.Msg { return slideMsg{} })
}

// fontsHidden reports whether the Fonts pane steps aside for Style: Style is
// open or sliding and the three columns do not fit.
func (m Model) fontsHidden() bool {
	return (m.panelOpen || m.panelW > 0) && m.width < bothFitWidth
}

// toggleStyle opens the Style pane (taking focus) or closes it (returning
// focus to where it was); the pane slides in from the right. Opening it where
// it does not fit beside the minimum preview is refused with a status.
func (m *Model) toggleStyle() tea.Cmd {
	if m.panelOpen {
		m.panelOpen = false
		return tea.Batch(m.setFocus(m.prevFocus), slide())
	}
	if m.width < styleFitWidth {
		return m.setStatus(fmt.Sprintf("Style needs %d columns", styleFitWidth))
	}
	m.panelOpen = true
	m.prevFocus = m.focus
	m.styleCur.move(styleRows(), 0, m.styleHeight())
	cmds := []tea.Cmd{m.setFocus(focusStyle), slide()}
	if m.fontsHidden() {
		m.layout()
		m.refreshPreview(false)
		cmds = append(cmds, m.setStatus("Fonts hidden to fit Style · ctrl+o brings them back"))
	}
	return tea.Batch(cmds...)
}

// stepSlide moves the pane width one frame toward its target.
func (m *Model) stepSlide() tea.Cmd {
	target := 0
	if m.panelOpen {
		target = styleWidth
	}
	switch {
	case m.panelW < target:
		m.panelW = min(m.panelW+slideStep, target)
	case m.panelW > target:
		m.panelW = max(m.panelW-slideStep, target)
	default:
		return nil
	}
	m.layout()
	m.refreshPreview(false)
	if m.panelW == target {
		return nil
	}
	return slide()
}

// indexMsg delivers the fontclass index built after every font has loaded.
type indexMsg struct{ index *fontclass.Index }

func buildIndex(cat *catalog.Catalog) tea.Cmd {
	return func() tea.Msg { return indexMsg{fontclass.Build(cat)} }
}

// filterRows lists the Fonts pane's filter rows in display order.
func filterRows() []optRow {
	rows := []optRow{{kind: optCollapse}, {kind: optGroupBy}, {kind: optSimilar}}
	for _, a := range fontclass.Axes() {
		rows = append(rows, optRow{kind: optAxis, axis: a.Axis})
		for v := range a.Values {
			rows = append(rows, optRow{kind: optFacet, axis: a.Axis, value: v})
		}
	}
	return append(rows, optRow{kind: optClear})
}

// styleRows lists the Style pane's rows in display order.
func styleRows() []optRow {
	return append([]optRow{{kind: optAlign}}, paintRows()...)
}

// rowCursor is a selectable-row cursor and scroll offset over a row list.
type rowCursor struct{ cursor, offset int }

// move moves the cursor by d selectable rows (0 just snaps it onto a
// selectable row) and keeps it within a window of height rows.
func (c *rowCursor) move(rows []optRow, d, height int) {
	step := 1
	if d < 0 {
		step = -1
	}
	i := max(0, min(c.cursor, len(rows)-1))
	for moved := 0; moved < max(d, -d); moved++ {
		j := i + step
		for j >= 0 && j < len(rows) && !rows[j].selectable() {
			j += step
		}
		if j < 0 || j >= len(rows) {
			break
		}
		i = j
	}
	for !rows[i].selectable() {
		i++
	}
	c.cursor = i
	if c.cursor < c.offset {
		c.offset = c.cursor
		if c.cursor > 0 && !rows[c.cursor-1].selectable() {
			c.offset-- // keep the axis label above its first value
		}
	}
	if c.cursor >= c.offset+height {
		c.offset = c.cursor - height + 1
	}
	c.offset = max(0, min(c.offset, max(len(rows)-height, 0)))
}

// rowKey applies the shared row keys to c and reports whether key was one.
func (m *Model) rowKey(c *rowCursor, rows []optRow, height int, key string) bool {
	switch key {
	case "up", "k":
		c.move(rows, -1, height)
	case "down", "j":
		c.move(rows, 1, height)
	case " ", "enter", "right", "l":
		m.applyOption(rows[c.cursor], 1)
	case "left", "h":
		m.applyOption(rows[c.cursor], -1)
	default:
		return false
	}
	return true
}

func (m Model) handleFiltersKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		return m, m.setFocus(focusList)
	}
	m.rowKey(&m.filterCur, filterRows(), m.filterHeight(), key)
	return m, nil
}

func (m Model) handleStyleKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		return m, m.focusFonts(focusList)
	}
	m.rowKey(&m.styleCur, styleRows(), m.styleHeight(), key)
	return m, nil
}

// applyOption toggles a row; dir only matters for Group by.
func (m *Model) applyOption(r optRow, dir int) {
	s := m.state
	ready := s.Index() != nil
	switch r.kind {
	case optAlign:
		n := display.NumAligns
		s.Align = display.Align((int(s.Align) + dir + n) % n)
		m.refreshPreview(false)
		return
	case optSlider:
		if sliders[r.slider].enabled(s.Paint.Background.Mode) {
			stepSlider(&s.Paint, r.slider, dir)
			m.refreshPreview(false)
		}
		return
	case optBgMode:
		n := int(paint.NumModes)
		s.Paint.Background.Mode = paint.Mode((int(s.Paint.Background.Mode) + dir + n) % n)
		m.refreshPreview(false)
		return
	case optCollapse:
		if !ready || dir < 0 {
			return
		}
		s.Collapse = !s.Collapse
	case optGroupBy:
		if !ready {
			return
		}
		n := int(fontclass.NumAxes) + 1 // NoGroup plus each axis
		s.GroupBy = fontclass.Axis((int(s.GroupBy)+1+dir+n)%n - 1)
	case optSimilar:
		if !ready || dir < 0 {
			return
		}
		switch {
		case s.Similar() != "":
			s.SetSimilar("")
		case m.cursorName() != "":
			s.SetSimilar(m.cursorName())
		default:
			return
		}
	case optFacet:
		if !ready || dir < 0 {
			return
		}
		s.ToggleFilter(r.axis, r.value)
	case optClear:
		if !ready || dir < 0 {
			return
		}
		s.ClearFilters()
	default:
		return
	}
	m.itemsChanged()
}

// itemsChanged rebuilds the sidebar after the universe changed, keeping the
// cursor on the same font when it is still listed.
func (m *Model) itemsChanged() {
	name := m.cursorName()
	m.items = m.state.Sidebar()
	m.cursor = 0
	for i, it := range m.items {
		if name != "" && it.Name == name {
			m.cursor = i
			break
		}
	}
	m.clampOffset()
	m.refreshPreview(false)
	m.revealCursor()
}

// styleHeight is the number of Style rows the pane shows.
func (m Model) styleHeight() int { return max(m.bodyHeight()-2, 1) }

// styleView renders the framed pane, cropped to the slide's current width.
func (m Model) styleView() string {
	var lines []string
	rows := styleRows()
	end := min(m.styleCur.offset+m.styleHeight(), len(rows))
	for i := m.styleCur.offset; i < end; i++ {
		lines = append(lines, m.optionLine(rows[i], i == m.styleCur.cursor && m.focus == focusStyle, styleWidth))
	}
	pane := frame("Style", "ctrl+o", strings.Join(lines, "\n"), styleCols, m.styleHeight(), 0, 1, m.focus == focusStyle)
	return lipgloss.NewStyle().MaxWidth(m.panelW + styleFrameCols).Height(m.bodyHeight()).MaxHeight(m.bodyHeight()).Render(pane)
}

// optionLine renders a row for a pane width columns wide.
func (m Model) optionLine(r optRow, cursor bool, width int) string {
	s := m.state
	ready := s.Index() != nil
	check := func(on bool) string {
		if on {
			return styles.selected.Render("[x]")
		}
		return styles.dim.Render("[ ]")
	}
	pending := styles.dim.Render("analysing…")
	var line string
	switch r.kind {
	case optAlign:
		line = alignLine(s.Align)
	case optCollapse:
		line = check(s.Collapse) + " Collapse duplicates"
		if !ready {
			line = "    Collapse  " + pending
		}
	case optGroupBy:
		group := "none"
		if s.GroupBy != display.NoGroup {
			group = fontclass.Axes()[s.GroupBy].Name
		}
		line = "    Group by ‹ " + styles.label.Render(group) + " ›"
		if !ready {
			line = "    Group by  " + pending
		}
	case optSimilar:
		switch {
		case !ready:
			line = "    Similar   " + pending
		case s.Similar() != "":
			line = check(true) + " Similar to " + truncate(s.Similar(), width-17)
		case m.cursorName() != "":
			line = check(false) + " Similar to " + truncate(m.cursorName(), width-17)
		default:
			line = styles.dim.Render("    Similar to: pick a font")
		}
	case optAxis:
		return "  " + styles.header.Render(fontclass.Axes()[r.axis].Name)
	case optHeading:
		return headingLine(&s.Paint, r.label)
	case optSlider:
		line = sliderLine(&s.Paint, r.slider)
	case optBgMode:
		line = bgModeLine(&s.Paint)
	case optFacet:
		name := fontclass.Axes()[r.axis].Values[r.value]
		if !ready {
			line = "    " + styles.dim.Render(name)
		} else {
			count := fmt.Sprint(s.FacetCount(r.axis, r.value))
			pad := max(width-8-len([]rune(name))-len(count), 1)
			line = check(s.Filtered(r.axis, r.value)) + " " + name + strings.Repeat(" ", pad) + styles.dim.Render(count)
		}
	case optClear:
		line = "    Clear filters"
		if !s.HasFilters() {
			line = styles.dim.Render(line)
		}
	}
	if cursor {
		return styles.cursor.Render("›") + " " + line
	}
	return "  " + line
}
