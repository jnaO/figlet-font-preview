// Package tui is the Bubble Tea front-end over internal/display.
package tui

import (
	"fmt"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jnaO/figlet-font-preview/internal/action"
	"github.com/jnaO/figlet-font-preview/internal/catalog"
	"github.com/jnaO/figlet-font-preview/internal/display"
)

const (
	sidebarWidth = 30 // content columns, excluding the frame
	maxTextLines = 5  // the text area grows up to this many rows, then scrolls
	tickEvery    = 100 * time.Millisecond
)

type focus int

const (
	focusText focus = iota
	focusSearch
	focusFilters
	focusList
	focusStyle // only while the Style pane is open
)

type tickMsg struct{}

// span is a half-open range of preview lines.
type span struct{ start, end int }

// Model is the root Bubble Tea model.
type Model struct {
	cat     *catalog.Catalog
	state   *display.State
	loaded  <-chan struct{}
	loading bool

	text    textarea.Model
	search  textinput.Model
	preview viewport.Model
	focus   focus
	palette paletteState
	help    helpState

	items          []display.Item
	cursor, offset int
	spans          map[string]span // preview line range of each rendered font

	panelOpen bool  // the Style pane is open (or sliding in)
	panelW    int   // the pane's current width while sliding
	prevFocus focus // focus to return to when the pane closes

	filterCur, styleCur rowCursor

	status   string // transient footer message
	statusID int

	width, height int
	ready         bool
}

// New builds the model and starts parsing every font in the background.
func New(cat *catalog.Catalog, text string, align display.Align) Model {
	ti := textarea.New()
	ti.ShowLineNumbers = false
	ti.EndOfBufferCharacter = ' '
	ti.CharLimit = 0
	ti.Prompt = ""
	for _, st := range []*textarea.Style{&ti.FocusedStyle, &ti.BlurredStyle} {
		st.CursorLine = lipgloss.NewStyle()
		st.Prompt = styles.label
		st.Text = lipgloss.NewStyle()
	}
	ti.SetValue(text)
	ti.Focus()

	si := textinput.New()
	si.Prompt = "/ "
	si.PromptStyle = styles.label
	si.Placeholder = "fuzzy search fonts"

	m := Model{
		cat:     cat,
		state:   display.New(cat.Names(), text),
		loaded:  cat.Preload(runtime.NumCPU()),
		loading: true,
		text:    ti,
		search:  si,
		preview: viewport.New(0, 0),
	}
	m.state.Align = align
	m.state.Paint.Background.From, m.state.Paint.Background.To = themeSeed()
	m.items = m.state.Sidebar()
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, tick())
}

func tick() tea.Cmd {
	return tea.Tick(tickEvery, func(time.Time) tea.Msg { return tickMsg{} })
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height, m.ready = msg.Width, msg.Height, true
		var cmd tea.Cmd
		switch {
		case m.panelOpen && m.width < styleFitWidth:
			cmd = m.toggleStyle()
		case m.panelOpen && m.fontsHidden() && m.focus != focusText:
			cmd = m.setFocus(focusStyle)
		}
		m.layout()
		m.refreshPreview(false)
		return m, cmd

	case tickMsg:
		select {
		case <-m.loaded:
			m.loading = false
		default:
		}
		m.refreshPreview(false)
		if m.loading {
			return m, tick()
		}
		return m, buildIndex(m.cat)

	case slideMsg:
		return m, m.stepSlide()

	case indexMsg:
		m.state.SetIndex(msg.index)
		m.itemsChanged()
		return m, nil

	case copiedMsg:
		return m, m.copied(msg)

	case exportedMsg:
		return m, m.exported(msg)

	case clearStatusMsg:
		if msg.id == m.statusID {
			m.status = ""
		}
		return m, nil

	case tea.MouseMsg:
		var cmd tea.Cmd
		m.preview, cmd = m.preview.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	var cmd tea.Cmd
	switch {
	case m.palette.open:
		m.palette.input, cmd = m.palette.input.Update(msg)
		return m, cmd
	}
	switch m.focus {
	case focusText:
		m.text, cmd = m.text.Update(msg)
	case focusSearch:
		m.search, cmd = m.search.Update(msg)
	}
	return m, cmd
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.palette.open {
		return m.paletteKey(msg)
	}
	if m.help.open {
		return m.helpKey(msg)
	}
	if e, ok := lookupKey(msg.String()); ok {
		return m, runEntry(&m, e, msg.String())
	}
	if msg.String() == "?" && m.focus != focusText && m.focus != focusSearch {
		e, _ := binding(action.Help)
		return m, runEntry(&m, e, "?")
	}

	switch m.focus {
	case focusText:
		if msg.String() == "esc" {
			return m, m.focusFonts(focusList)
		}
		var cmd tea.Cmd
		// The textarea scrolls its own view during Update based on its
		// current height, so it updates at full height and is then fitted.
		prevH, prevLines := m.text.Height(), m.text.LineCount()
		m.text.SetHeight(maxTextLines)
		m.text, cmd = m.text.Update(msg)
		if n := m.text.LineCount(); n < prevLines && n <= maxTextLines {
			m.resetTextScroll()
		}
		if m.textHeight() != prevH {
			m.layout()
		} else {
			m.text.SetHeight(prevH)
		}
		if v := m.text.Value(); v != m.state.Text {
			m.state.Text = v
			m.refreshPreview(false)
		}
		return m, cmd

	case focusSearch:
		switch msg.String() {
		case "esc":
			m.search.SetValue("")
			m.applyQuery()
			return m, m.setFocus(focusList)
		case "up", "down":
			m.moveCursor(msg.String())
			return m, nil
		case "enter":
			return m, m.setFocus(focusList)
		}
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
		m.applyQuery()
		return m, cmd

	case focusFilters:
		return m.handleFiltersKey(msg.String())

	case focusStyle:
		return m.handleStyleKey(msg.String())

	case focusList:
		switch msg.String() {
		case "/":
			return m, m.setFocus(focusSearch)
		case "esc":
			m.search.SetValue("")
			m.applyQuery()
		case "up", "k", "down", "j", "home", "g", "end", "G":
			m.moveCursor(msg.String())
		case " ", "enter":
			m.toggleCursor()
		case "x":
			m.clearPicks()
		}
	}
	return m, nil
}

// fontsStop maps the search box to the list, Fonts' tab stop.
func fontsStop(f focus) focus {
	if f == focusSearch {
		return focusList
	}
	return f
}

// nextFocus cycles the tab stops text → Fonts (the list) → Filters → Style
// (while open), skipping the left column while it is hidden.
func (m Model) nextFocus(d int) focus {
	stops := []focus{focusText, focusList, focusFilters, focusStyle}
	i := slices.Index(stops, fontsStop(m.focus))
	for range stops {
		i = (i + d + len(stops)) % len(stops)
		if m.focusVisible(stops[i]) {
			return stops[i]
		}
	}
	return m.focus
}

func (m Model) focusVisible(f focus) bool {
	switch f {
	case focusStyle:
		return m.panelOpen
	case focusSearch, focusFilters, focusList:
		return !m.fontsHidden()
	}
	return true
}

func (m *Model) setFocus(f focus) tea.Cmd {
	m.focus = f
	m.text.Blur()
	m.search.Blur()
	switch f {
	case focusText:
		return m.text.Focus()
	case focusSearch:
		return m.search.Focus()
	case focusFilters:
		m.filterCur.move(filterRows(), 0, m.filterHeight())
	}
	return nil
}

func (m *Model) applyQuery() {
	if q := m.search.Value(); q != m.state.Query() {
		m.state.SetQuery(q)
		m.items = m.state.Sidebar()
		m.cursor, m.offset = 0, 0
		m.refreshPreview(true)
		m.revealCursor()
	}
}

func (m *Model) moveCursor(key string) {
	switch key {
	case "up", "k":
		m.cursor--
	case "down", "j":
		m.cursor++
	case "home", "g":
		m.cursor = 0
	case "end", "G":
		m.cursor = len(m.items) - 1
	}
	m.cursor = max(0, min(m.cursor, len(m.items)-1))
	m.clampOffset()
	m.refreshPreview(false)
	m.revealCursor()
}

func (m Model) cursorName() string {
	if m.cursor < len(m.items) {
		return m.items[m.cursor].Name
	}
	return ""
}

// revealCursor centres the cursor font in the preview as far as scrolling
// allows; fonts outside the render set leave the scroll untouched.
func (m *Model) revealCursor() {
	if s, ok := m.spans[m.cursorName()]; ok {
		m.preview.SetYOffset(revealOffset(m.preview.Height, s.start, s.end))
	}
}

func (m *Model) toggleCursor() {
	if m.cursor >= len(m.items) {
		return
	}
	if it := m.items[m.cursor]; it.IsHeader() {
		m.state.ToggleGroup(it.Header)
	} else {
		m.state.Toggle(it.Name)
	}
	m.refreshPreview(true)
	m.revealCursor()
}

func (m *Model) clearPicks() {
	m.state.ClearSelection()
	m.refreshPreview(true)
}

func (m *Model) scrollPreview(key string) {
	switch key {
	case "pgup":
		m.preview.ViewUp()
	case "pgdown":
		m.preview.ViewDown()
	case "ctrl+u":
		m.preview.HalfViewUp()
	case "ctrl+d":
		m.preview.HalfViewDown()
	}
}

// Layout: text bar (1..maxTextLines rows), body (Fonts | preview | Style), footer.
func (m *Model) layout() {
	m.text.SetWidth(max(m.barWidth()-4, 1))
	m.text.SetHeight(m.textHeight())
	m.search.Width = sidebarWidth - lipgloss.Width(m.search.Prompt) - 1
	m.preview.Width = m.previewWidth()
	m.preview.Height = m.bodyHeight()
	m.filterCur.move(filterRows(), 0, m.filterHeight())
	m.styleCur.move(styleRows(), 0, m.styleHeight())
	m.clampOffset()
}

// resetTextScroll scrolls the textarea back to its first line, keeping the
// cursor: textarea keeps a stale scroll offset when lines are deleted, and
// SetValue is the only way to reset it.
func (m *Model) resetTextScroll() {
	row := m.text.Line()
	info := m.text.LineInfo()
	col := info.StartColumn + info.ColumnOffset
	m.text.SetValue(m.text.Value())
	for m.text.Line() > row {
		m.text.CursorUp()
	}
	m.text.SetCursor(col)
}

func (m Model) textHeight() int { return max(1, min(m.text.LineCount(), maxTextLines)) }

// barWidth is the text bar's outer width; the last terminal column stays free.
func (m Model) barWidth() int { return max(m.width-1, 5) }

// barHeight is the text rows plus the frame's top and bottom edges.
func (m Model) barHeight() int { return m.textHeight() + 2 }

func (m Model) bodyHeight() int { return max(m.height-m.barHeight()-1, 1) }

// textBar frames the text area.
func (m Model) textBar() string {
	return frame("Text", "ctrl+e", m.text.View(), m.barWidth(), m.textHeight(), 1, 1, m.focus == focusText)
}

func (m Model) previewWidth() int {
	w := m.width - frameCols
	if !m.fontsHidden() {
		w -= fontsCols
	}
	if m.panelW > 0 {
		w -= m.panelW + styleFrameCols
	}
	return max(w, 1)
}

// filterHeight is the number of filter rows the Filters frame shows; Fonts
// keeps at least one list row when the body allows.
func (m Model) filterHeight() int {
	return max(min(max(m.bodyHeight()/3, 3), len(filterRows()), m.bodyHeight()-8), 1)
}

// listHeight is the number of font rows the Fonts frame shows: both frames'
// four border rows, the Filters window, search, count and legend take the rest.
func (m Model) listHeight() int { return max(m.bodyHeight()-m.filterHeight()-7, 1) }

func (m *Model) clampOffset() {
	h := m.listHeight()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+h {
		m.offset = m.cursor - h + 1
	}
	m.offset = max(0, min(m.offset, max(len(m.items)-h, 0)))
}

// refreshPreview re-renders the preview. reset scrolls it to the top;
// otherwise the anchor font keeps its screen row while the fonts above it
// change height.
func (m *Model) refreshPreview(reset bool) {
	if !m.ready {
		return
	}
	anchor, hasAnchor := m.anchorFont()
	row := m.spans[anchor].start - m.preview.YOffset
	content, spans := m.previewContent()
	m.spans = spans
	m.preview.SetContent(content)
	switch {
	case reset:
		m.preview.GotoTop()
	case hasAnchor:
		if s, ok := m.spans[anchor]; ok {
			m.preview.SetYOffset(s.start - row)
		}
	}
}

// anchorFont is the cursor font when it is rendered, else the font at the
// top of the preview.
func (m Model) anchorFont() (string, bool) {
	if _, ok := m.spans[m.cursorName()]; ok {
		return m.cursorName(), true
	}
	y := m.preview.YOffset
	for name, s := range m.spans {
		if s.start <= y && y <= s.end {
			return name, true
		}
	}
	return "", false
}

// previewContent renders the render set and records each font's line span.
// The cursor font is drawn in rainbow colours on a tinted background.
func (m Model) previewContent() (string, map[string]span) {
	if len(m.cat.Entries) == 0 {
		return styles.dim.Render(catalog.NoFontsMessage(m.cat.Roots, runtime.GOOS, addFolderHint)), nil
	}
	set := m.state.RenderSet()
	if len(set) == 0 {
		return styles.dim.Render("No fonts match the search."), nil
	}
	w := m.previewWidth()
	cursor := m.cursorName()
	spans := make(map[string]span, len(set))
	var lines []string
	for _, name := range set {
		hl := name == cursor
		start := len(lines)
		lines = append(lines, m.fontHeader(name, w, hl))
		e, _ := m.cat.Get(name)
		f, err := e.Font()
		var body []string
		switch {
		case err != nil:
			body = []string{truncate(err.Error(), w)}
		case f == nil:
			body = []string{"loading…"}
		default:
			body = m.state.Align.Apply(f.Render(m.state.Text, w), f.Height, w)
		}
		switch {
		case hl && f != nil:
			body = lolcat(body, w, m.state.Paint)
		case hl:
			for i, l := range body {
				body[i] = tintLine(l, w, m.state.Paint)
			}
		case err != nil:
			body = []string{styles.err.Render(body[0])}
		case f == nil:
			body = []string{styles.dim.Render(body[0])}
		}
		lines = append(lines, body...)
		spans[name] = span{start, len(lines)}
		lines = append(lines, "")
	}
	return strings.Join(lines[:len(lines)-1], "\n"), spans
}

func (m Model) fontHeader(name string, w int, hl bool) string {
	label := " " + name + " "
	style, rule := styles.header, styles.rule
	if m.state.IsSelected(name) {
		label = " ● " + name + " "
		style = styles.headerSelected
	}
	if hl {
		bg := bgStyle(m.state.Paint)
		style, rule = style.Inherit(bg), rule.Inherit(bg)
	}
	n := max(w-lipgloss.Width(label)-2, 0)
	return style.Render("──"+label) + rule.Render(strings.Repeat("─", n))
}

func (m Model) View() string {
	if !m.ready {
		return ""
	}
	header := m.textBar()
	var cols []string
	if !m.fontsHidden() {
		cols = append(cols, m.sidebarView())
	}
	cols = append(cols, lipgloss.NewStyle().PaddingLeft(1).Width(m.previewWidth()+1).Render(m.preview.View()))
	if m.panelW > 0 {
		cols = append(cols, m.styleView())
	}
	body := lipgloss.JoinHorizontal(lipgloss.Top, cols...)
	screen := lipgloss.JoinVertical(lipgloss.Left, header, body, m.footerView())
	if m.help.open {
		box, w, h := m.helpView()
		screen = overlay(screen, box, (m.width-w)/2, m.barHeight()+(m.bodyHeight()-h)/2)
	}
	if m.palette.open {
		screen = overlay(screen, m.paletteView(), (m.width-m.paletteBox())/2, m.barHeight())
	}
	return screen
}

// sidebarView is the left column: the Fonts frame above the Filters frame.
func (m Model) sidebarView() string {
	col := lipgloss.JoinVertical(lipgloss.Left, m.fontsView(), m.filtersView())
	return lipgloss.NewStyle().Height(m.bodyHeight()).MaxHeight(m.bodyHeight()).Render(col)
}

// fontsView frames the search, count, legend and font list.
func (m Model) fontsView() string {
	lines := []string{m.search.View(), styles.dim.Render(m.countLine()), legendLine()}
	end := min(m.offset+m.listHeight(), len(m.items))
	for i := m.offset; i < end; i++ {
		lines = append(lines, m.itemView(i))
	}
	return frame("Fonts", "ctrl+f", strings.Join(lines, "\n"), fontsCols, 3+m.listHeight(), 0, 0,
		m.focus == focusSearch || m.focus == focusList)
}

// filtersView frames the filter window.
func (m Model) filtersView() string {
	var lines []string
	rows := filterRows()
	for i := m.filterCur.offset; i < min(m.filterCur.offset+m.filterHeight(), len(rows)); i++ {
		lines = append(lines, m.optionLine(rows[i], i == m.filterCur.cursor && m.focus == focusFilters, sidebarWidth))
	}
	return frame("Filters", "ctrl+t", strings.Join(lines, "\n"), fontsCols, m.filterHeight(), 0, 0, m.focus == focusFilters)
}

func (m Model) countLine() string {
	n := 0
	for _, it := range m.items {
		if !it.IsHeader() {
			n++
		}
	}
	return truncate(formatCount(n, len(m.cat.Entries), m.state.SelectedCount(), m.loading), sidebarWidth)
}

// formatCount reads "listed/installed · picked ●", plus " · loading" while fonts load.
func formatCount(listed, installed, picked int, loading bool) string {
	s := fmt.Sprintf("%d/%d · %d ●", listed, installed, picked)
	if loading {
		s += " · loading"
	}
	return s
}

// legendLine explains the two list marks: › drives copy and SVG, ● narrows the preview.
func legendLine() string {
	return styles.cursor.Render("›") + styles.dim.Render(" copy/svg · ") +
		styles.selected.Render("●") + styles.dim.Render(" preview only")
}

func (m Model) itemView(i int) string {
	it := m.items[i]
	var row string
	if it.IsHeader() {
		mark := styles.dim.Render("▸ ")
		if m.state.GroupSelected(it.Header) {
			mark = styles.selected.Render("● ")
		}
		count := fmt.Sprintf(" %d", it.Count)
		row = mark + styles.header.Render(truncate(it.Header, sidebarWidth-4-len(count))) + styles.dim.Render(count)
	} else {
		mark := styles.dim.Render("○ ")
		if m.state.IsSelected(it.Name) {
			mark = styles.selected.Render("● ")
		}
		dups := ""
		if it.Dups > 0 {
			dups = fmt.Sprintf(" +%d", it.Dups)
		}
		row = mark + highlight(truncate(it.Name, sidebarWidth-4-len(dups)), it.Matched) + styles.dim.Render(dups)
	}
	if i == m.cursor {
		cur := styles.cursor
		if m.focus != focusSearch && m.focus != focusList {
			cur = styles.cursorInactive
		}
		return cur.Render("›") + " " + row
	}
	return "  " + row
}

func (m Model) footerView() string {
	if m.status != "" {
		return styles.label.Render(truncate(m.status, m.width-1))
	}
	return styles.dim.Render(truncate(footerKeys(m.width-1), m.width-1))
}

// highlight styles the runes of s at the matched indexes.
func highlight(s string, matched []int) string {
	if len(matched) == 0 {
		return s
	}
	hit := map[int]bool{}
	for _, i := range matched {
		hit[i] = true
	}
	var b strings.Builder
	for i, r := range s {
		if hit[i] {
			b.WriteString(styles.match.Render(string(r)))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func truncate(s string, w int) string {
	r := []rune(s)
	if len(r) <= w {
		return s
	}
	if w <= 1 {
		return string(r[:max(w, 0)])
	}
	return string(r[:w-1]) + "…"
}

// revealOffset returns the scroll offset (viewport height h) that centres
// lines [start, end), with any odd spare row below; a block of h lines or more
// gets its first line at the top. The result may fall outside the scrollable
// range; viewport.SetYOffset clamps it.
func revealOffset(h, start, end int) int {
	if end-start >= h {
		return start
	}
	return start - (h-(end-start))/2
}
