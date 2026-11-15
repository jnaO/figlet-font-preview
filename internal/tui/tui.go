// Package tui is the Bubble Tea front-end over internal/display.
package tui

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jnaO/figlet-font-preview/internal/catalog"
	"github.com/jnaO/figlet-font-preview/internal/display"
)

const (
	sidebarWidth = 30 // content columns, excluding the divider
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
	switch m.focus {
	case focusText:
		m.text, cmd = m.text.Update(msg)
	case focusSearch:
		m.search, cmd = m.search.Update(msg)
	}
	return m, cmd
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "ctrl+o":
		return m, m.toggleStyle()
	case "ctrl+y":
		return m, m.copyCursor()
	case "ctrl+s":
		return m, m.exportCursor()
	case "tab":
		return m, m.setFocus(m.nextFocus(1))
	case "shift+tab":
		return m, m.setFocus(m.nextFocus(-1))
	case "pgup", "pgdown", "ctrl+u", "ctrl+d":
		m.scrollPreview(msg.String())
		return m, nil
	}

	switch m.focus {
	case focusText:
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
		case "up", "down":
			m.moveCursor(msg.String())
			return m, nil
		case "enter":
			m.toggleCursor()
			return m, nil
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
		case "q":
			return m, tea.Quit
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
			m.state.ClearSelection()
			m.refreshPreview(true)
		}
	}
	return m, nil
}

// nextFocus cycles text → search → filters → list (→ Style while open),
// skipping the Fonts zones while the Fonts pane is hidden.
func (m Model) nextFocus(d int) focus {
	const n = int(focusStyle) + 1
	f := m.focus
	for range n {
		f = focus((int(f) + d + n) % n)
		if m.focusVisible(f) {
			return f
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

// revealCursor scrolls the preview only as far as needed to show the cursor
// font in full; fonts outside the render set leave the scroll untouched.
func (m *Model) revealCursor() {
	if s, ok := m.spans[m.cursorName()]; ok {
		m.preview.SetYOffset(revealOffset(m.preview.YOffset, m.preview.Height, s.start, s.end))
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

// textBar frames the text area. lipgloss has no border-title API, so the top
// edge is built by hand.
func (m Model) textBar() string {
	border, label := styles.divider, styles.header
	if m.focus == focusText {
		border, label = styles.label, styles.label
	}
	w := m.barWidth()
	top := border.Render("╭─ ") + label.Render("Text") + border.Render(" "+strings.Repeat("─", max(w-9, 0))+"╮")
	frame := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderTop(false).
		BorderForeground(border.GetForeground()).Padding(0, 1).Width(w - 2)
	return top + "\n" + frame.Render(m.text.View())
}

func (m Model) previewWidth() int {
	w := m.width - frameCols
	if !m.fontsHidden() {
		w -= fontsCols
	}
	if m.panelW > 0 {
		w -= m.panelW + 1 // pane plus its divider
	}
	return max(w, 1)
}

// filterHeight is the number of filter rows the Fonts pane shows.
func (m Model) filterHeight() int {
	return min(max(m.bodyHeight()/3, 3), len(filterRows()))
}

// listHeight is the number of font rows the Fonts pane shows: its search,
// two rules, filter window and count take the rest.
func (m Model) listHeight() int { return max(m.bodyHeight()-m.filterHeight()-4, 1) }

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
		return styles.dim.Render(noFontsMessage(m.cat.Roots, runtime.GOOS)), nil
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
	divider := styles.divider.Render(strings.Repeat("│\n", m.bodyHeight()-1) + "│")
	var cols []string
	if !m.fontsHidden() {
		cols = append(cols, m.sidebarView(), divider)
	}
	cols = append(cols, lipgloss.NewStyle().PaddingLeft(1).Width(m.previewWidth()+1).Render(m.preview.View()))
	if m.panelW > 0 {
		cols = append(cols, divider, m.styleView())
	}
	body := lipgloss.JoinHorizontal(lipgloss.Top, cols...)
	return lipgloss.JoinVertical(lipgloss.Left, header, body, m.footerView())
}

// sidebarView is the Fonts pane: search, filter window, count and font list.
func (m Model) sidebarView() string {
	rule := styles.rule.Render(strings.Repeat("─", sidebarWidth))
	lines := []string{m.search.View(), rule}
	rows := filterRows()
	for i := m.filterCur.offset; i < min(m.filterCur.offset+m.filterHeight(), len(rows)); i++ {
		lines = append(lines, m.optionLine(rows[i], i == m.filterCur.cursor && m.focus == focusFilters, sidebarWidth))
	}
	lines = append(lines, rule, styles.dim.Render(m.countLine()))
	end := min(m.offset+m.listHeight(), len(m.items))
	for i := m.offset; i < end; i++ {
		lines = append(lines, m.itemView(i))
	}
	return lipgloss.NewStyle().Width(sidebarWidth).Height(m.bodyHeight()).
		MaxHeight(m.bodyHeight()).Render(strings.Join(lines, "\n"))
}

func (m Model) countLine() string {
	n := 0
	for _, it := range m.items {
		if !it.IsHeader() {
			n++
		}
	}
	s := fmt.Sprintf("%d of %d fonts · %d selected", n, len(m.cat.Entries), m.state.SelectedCount())
	if m.loading {
		s += " · loading"
	}
	return truncate(s, sidebarWidth)
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
	next, after := "search", "text"
	if m.fontsHidden() {
		next = "style"
	}
	if m.panelOpen {
		after = "style"
	}
	var keys string
	switch m.focus {
	case focusText:
		keys = "ctrl+o: style · ctrl+y: copy · ctrl+s: svg · type text · enter: new line · tab: " + next + " · pgup/pgdn: scroll · ctrl+c: quit"
	case focusSearch:
		keys = "ctrl+o: style · ctrl+y: copy · ctrl+s: svg · type to filter · ↑/↓: move · enter: select · esc: clear · tab: filters"
	case focusFilters:
		keys = "ctrl+o: style · ctrl+y: copy · ctrl+s: svg · ↑/↓ j/k: move · space: toggle · ←/→: group by · esc: list · tab: list · q: quit"
	case focusList:
		keys = "ctrl+o: style · ctrl+y: copy · ctrl+s: svg · ↑/↓ j/k: move · space: select · x: clear selection · /: search · tab: " + after + " · q: quit"
	case focusStyle:
		keys = "ctrl+y: copy · ctrl+s: svg · ↑/↓ j/k: move · space: toggle · ←/→: change · tab: text · esc/ctrl+o: close style"
	}
	if m.status != "" {
		return styles.label.Render(truncate(m.status, m.width-1))
	}
	return styles.dim.Render(truncate(keys, m.width-1))
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

// revealOffset returns the smallest change to the scroll offset y (viewport
// height h) that brings lines [start, end) fully into view, or the block's
// first line when it is taller than the viewport.
func revealOffset(y, h, start, end int) int {
	switch {
	case start < y:
		return start
	case end > y+h:
		return min(start, end-h)
	}
	return y
}
