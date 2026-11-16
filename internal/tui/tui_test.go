package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/jnaO/figlet-font-preview/internal/catalog"
	"github.com/jnaO/figlet-font-preview/internal/display"
	"github.com/jnaO/figlet-font-preview/internal/fontclass"
	"github.com/jnaO/figlet-font-preview/internal/paint"
)

func TestRevealOffset(t *testing.T) {
	const h = 10
	cases := []struct {
		name          string
		start, end, w int
	}{
		{"even spare rows centre", 20, 24, 17},
		{"odd spare rows put the extra row below", 20, 25, 18},
		{"exactly the view height shows top", 20, 30, 20},
		{"taller than view shows top", 5, 30, 5},
		{"near the start is negative", 1, 4, -2},
	}
	for _, c := range cases {
		if got := revealOffset(h, c.start, c.end); got != c.w {
			t.Errorf("%s: got %d, want %d", c.name, got, c.w)
		}
	}
}

func centreRow(m Model) int {
	s := m.spans[m.cursorName()]
	return (m.preview.Height - (s.end - s.start)) / 2
}

func TestCursorMoveCentresFont(t *testing.T) {
	m := key(testModel(t, 30), "tab")
	for range 10 {
		m = key(m, "down")
	}
	if got := screenRow(m, m.cursorName()); got != centreRow(m) {
		t.Fatalf("moving down: row %d, want %d", got, centreRow(m))
	}
	before := m.preview.YOffset
	m = key(m, "down")
	if m.preview.YOffset == before {
		t.Fatal("a move between two visible fonts did not scroll")
	}
	if got := screenRow(m, m.cursorName()); got != centreRow(m) {
		t.Fatalf("down between visible fonts: row %d, want %d", got, centreRow(m))
	}
	m = key(m, "up")
	if got := screenRow(m, m.cursorName()); got != centreRow(m) {
		t.Fatalf("moving up: row %d, want %d", got, centreRow(m))
	}
}

func TestCursorJumpsClampAtEnds(t *testing.T) {
	m := key(testModel(t, 30), "tab", "G")
	if want := m.preview.TotalLineCount() - m.preview.Height; m.preview.YOffset != want {
		t.Fatalf("G: offset %d, want %d", m.preview.YOffset, want)
	}
	m = key(m, "g")
	if m.preview.YOffset != 0 {
		t.Fatalf("g: offset %d, want 0", m.preview.YOffset)
	}
}

func TestCursorMoveCentresAfterFreeScroll(t *testing.T) {
	m := key(testModel(t, 30), "tab")
	for range 10 {
		m = key(m, "down")
	}
	m = key(m, "pgdown", "down")
	if got := screenRow(m, m.cursorName()); got != centreRow(m) {
		t.Fatalf("after pgdown: row %d, want %d", got, centreRow(m))
	}
}

func TestPaletteUnpickCentresWhileFontsHidden(t *testing.T) {
	m := key(testModel(t, 30), "tab")
	for range 14 {
		m = key(m, "down")
	}
	m = key(m, "space")
	m = settle(key(resize(m, 70, 24), "ctrl+o"))
	if !m.fontsHidden() {
		t.Fatal("Fonts should be hidden")
	}
	m = key(m, "ctrl+k", "u", "n", "p", "i", "c", "k", "enter")
	if m.preview.YOffset == 0 {
		t.Fatal("preview did not scroll")
	}
	if got := screenRow(m, m.cursorName()); got != centreRow(m) {
		t.Fatalf("row %d, want %d", got, centreRow(m))
	}
}

func TestLolcatPadsAndColours(t *testing.T) {
	out := lolcat([]string{"ab", ""}, 5, paint.Default())
	if len(out) != 2 || out[0] == "ab" {
		t.Fatalf("got %q", out)
	}
}

func testModel(t *testing.T, n int) Model {
	t.Helper()
	data, err := os.ReadFile("../figfont/testdata/tiny.flf")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	for i := range n {
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("f%02d.flf", i)), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	m := New(catalog.Discover([]string{dir}), "ab", display.AlignLeft)
	m.state.Collapse = false // the fixtures are identical copies of one font
	<-m.loaded
	next, _ := m.Update(tea.WindowSizeMsg{Width: 60, Height: 20})
	return next.(Model)
}

func key(m Model, keys ...string) Model {
	for _, k := range keys {
		var msg tea.KeyMsg
		switch k {
		case "down":
			msg = tea.KeyMsg{Type: tea.KeyDown}
		case "up":
			msg = tea.KeyMsg{Type: tea.KeyUp}
		case "pgdown":
			msg = tea.KeyMsg{Type: tea.KeyPgDown}
		case "enter":
			msg = tea.KeyMsg{Type: tea.KeyEnter}
		case "tab":
			msg = tea.KeyMsg{Type: tea.KeyTab}
		case "esc":
			msg = tea.KeyMsg{Type: tea.KeyEsc}
		case "space":
			msg = tea.KeyMsg{Type: tea.KeySpace, Runes: []rune(" ")}
		case "right":
			msg = tea.KeyMsg{Type: tea.KeyRight}
		case "ctrl+o":
			msg = tea.KeyMsg{Type: tea.KeyCtrlO}
		case "ctrl+y":
			msg = tea.KeyMsg{Type: tea.KeyCtrlY}
		case "ctrl+s":
			msg = tea.KeyMsg{Type: tea.KeyCtrlS}
		case "ctrl+t":
			msg = tea.KeyMsg{Type: tea.KeyCtrlT}
		case "ctrl+k":
			msg = tea.KeyMsg{Type: tea.KeyCtrlK}
		case "ctrl+e":
			msg = tea.KeyMsg{Type: tea.KeyCtrlE}
		case "ctrl+f":
			msg = tea.KeyMsg{Type: tea.KeyCtrlF}
		case "shift+tab":
			msg = tea.KeyMsg{Type: tea.KeyShiftTab}
		case "ctrl+c":
			msg = tea.KeyMsg{Type: tea.KeyCtrlC}
		default:
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
		}
		next, _ := m.Update(msg)
		m = next.(Model)
		m.View() // components update scroll state from their last render
	}
	return m
}

func screenRow(m Model, name string) int { return m.spans[name].start - m.preview.YOffset }

func TestCursorFontKeepsRowWhenTextChanges(t *testing.T) {
	m := testModel(t, 30)
	m = key(m, "tab")
	for range 12 {
		m = key(m, "down")
	}
	name := m.cursorName()
	before := screenRow(m, name)
	if m.preview.YOffset == 0 {
		t.Fatal("expected the preview to have scrolled")
	}
	m = key(m, "tab", "tab", "c", "d", "enter", "e", "f") // list → filters → text; add a second line
	if got := screenRow(m, name); got != before {
		t.Fatalf("cursor font moved from row %d to %d", before, got)
	}
	if m.text.Height() != 2 {
		t.Fatalf("text area height %d, want 2", m.text.Height())
	}
}

func withIndex(m Model) Model {
	next, _ := m.Update(indexMsg{fontclass.Build(m.cat)})
	return next.(Model)
}

func settle(m Model) Model {
	for range 10 {
		next, _ := m.Update(slideMsg{})
		m = next.(Model)
	}
	return m
}

func TestStylePaneFromAnyFocus(t *testing.T) {
	next, _ := testModel(t, 3).Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	m := next.(Model)
	full := m.previewWidth()
	m = settle(key(m, "ctrl+o")) // from the text area
	if !m.panelOpen || m.focus != focusStyle || m.panelW != styleWidth {
		t.Fatalf("open=%v focus=%v width=%d", m.panelOpen, m.focus, m.panelW)
	}
	if m.previewWidth() != full-styleWidth-3 {
		t.Fatalf("preview width %d, want %d", m.previewWidth(), full-styleWidth-3)
	}
	m = settle(key(m, "ctrl+o"))
	if m.panelOpen || m.focus != focusText || m.panelW != 0 {
		t.Fatalf("after close: open=%v focus=%v width=%d", m.panelOpen, m.focus, m.panelW)
	}
	m = settle(key(m, "tab", "ctrl+o", "esc")) // esc goes to Fonts, Style stays open
	if !m.panelOpen || m.focus != focusList {
		t.Fatalf("esc: open=%v focus=%v", m.panelOpen, m.focus)
	}
	m = key(m, "tab", "tab", "tab") // list → filters → Style → text, leaving the pane open
	if !m.panelOpen || m.focus != focusText {
		t.Fatalf("tab: open=%v focus=%v", m.panelOpen, m.focus)
	}
}

func TestStyleAlignCycler(t *testing.T) {
	m := settle(key(resize(testModel(t, 3), 100, 30), "ctrl+o")) // first row is Align
	if !strings.Contains(m.styleView(), "‹ left ›") {
		t.Fatal("Align row should read ‹ left ›")
	}
	for _, step := range []struct {
		k    string
		want display.Align
	}{{"space", display.AlignCenter}, {"right", display.AlignRight}, {"right", display.AlignLeft}, {"left", display.AlignRight}} {
		m = key(m, step.k)
		if m.state.Align != step.want {
			t.Fatalf("after %q: Align = %v, want %v", step.k, m.state.Align, step.want)
		}
	}
	if !strings.Contains(m.styleView(), "‹ right ›") {
		t.Fatal("Align row should read ‹ right ›")
	}
}

// toOption moves the matching pane's cursor onto the first row of kind k and focuses it.
func toOption(t *testing.T, m Model, k optKind) Model {
	t.Helper()
	for i, r := range filterRows() {
		if r.kind == k {
			m.filterCur.cursor = i
			m.setFocus(focusFilters)
			return m
		}
	}
	for i, r := range styleRows() {
		if r.kind == k {
			m.styleCur.cursor = i
			m.setFocus(focusStyle)
			return m
		}
	}
	t.Fatalf("no option row of kind %d", k)
	return m
}

func TestOptionsInertBeforeIndex(t *testing.T) {
	m := toOption(t, testModel(t, 3), optFacet)
	m = key(m, "space")
	if m.state.HasFilters() || m.state.Collapse {
		t.Fatal("options must do nothing before the index arrives")
	}
}

func TestOptionsSimilarAnchorsCursorFont(t *testing.T) {
	m := withIndex(testModel(t, 5))
	m = key(m, "tab", "down", "down") // cursor on f02
	m = key(toOption(t, m, optSimilar), "space")
	if m.state.Similar() != "f02" {
		t.Fatalf("similar = %q, want f02", m.state.Similar())
	}
	if m.items[0].Name != "f02" {
		t.Fatalf("anchor should lead the sidebar, got %q", m.items[0].Name)
	}
	if m = key(m, "space"); m.state.Similar() != "" {
		t.Fatal("toggling again should clear the anchor")
	}
}

func TestGroupedSidebarHeaders(t *testing.T) {
	m := withIndex(testModel(t, 4))
	m = key(toOption(t, m, optGroupBy), "right", "esc") // Group by → Stroke, esc → list
	if m.state.GroupBy != fontclass.Stroke || !m.items[0].IsHeader() {
		t.Fatalf("expected a Stroke header first, got %+v", m.items[0])
	}
	m.cursor = 0
	y := m.preview.YOffset
	m = key(m, "space")
	if m.state.SelectedCount() != 4 {
		t.Fatalf("header should select its group, selected %d", m.state.SelectedCount())
	}
	if m.preview.YOffset != y || m.cursorName() != "" {
		t.Fatal("a header has no font and must not scroll the preview")
	}
	if m = key(m, "space"); m.state.SelectedCount() != 0 {
		t.Fatal("header should deselect a fully selected group")
	}
}

func TestCopyCursorFontAlignedAndTrimmed(t *testing.T) {
	var got string
	writeClipboard = func(s string) error { got = s; return nil }
	t.Cleanup(func() { writeClipboard = clipboard.WriteAll })

	for _, c := range []struct {
		a    display.Align
		want string
	}{
		{display.AlignCenter, "  a|b|\n  a|b|\na|b|c|d|\na|b|c|d|\n"},
		{display.AlignRight, "    a|b|\n    a|b|\na|b|c|d|\na|b|c|d|\n"},
		{display.AlignLeft, "a|b|\na|b|\na|b|c|d|\na|b|c|d|\n"},
	} {
		m := testModel(t, 2)
		m.state.Align = c.a
		m.text.SetValue("ab\nabcd")
		m.state.Text = m.text.Value()
		next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlY})
		m = next.(Model)
		next, _ = m.Update(cmd())
		m = next.(Model)

		if got != c.want {
			t.Fatalf("%v: clipboard = %q, want %q", c.a, got, c.want)
		}
		if !strings.Contains(m.footerView(), "copied f00") {
			t.Fatalf("footer %q lacks the copy status", m.footerView())
		}
	}
}

func trueColour(t *testing.T) {
	t.Helper()
	prev := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(prev) })
}

func TestLolcatBackgroundModes(t *testing.T) {
	trueColour(t)
	st := paint.Default()
	st.Background = paint.Background{Mode: paint.Solid, From: paint.RGB{R: 10, G: 20, B: 30}, To: paint.RGB{R: 250, G: 250, B: 250}}
	solid := strings.Join(lolcat([]string{"ab"}, 6, st), "")
	if !strings.Contains(solid, "48;2;10;20;30") || strings.Contains(solid, "48;2;250;250;250") {
		t.Fatalf("solid background wrong: %q", solid)
	}

	st.Background.Mode = paint.Linear
	lin := strings.Join(lolcat([]string{"ab"}, 6, st), "")
	left, _ := st.Background.At(0, 0, 6, 1)
	right, _ := st.Background.At(5, 0, 6, 1)
	for _, c := range []paint.RGB{left, right} {
		if want := fmt.Sprintf("48;2;%d;%d;%d", c.R, c.G, c.B); !strings.Contains(lin, want) {
			t.Fatalf("linear row lacks %s (padding must be painted per cell): %q", want, lin)
		}
	}
	fg := st.Rainbow.At(0, 1)
	if want := fmt.Sprintf("38;2;%d;%d;%d", fg.R, fg.G, fg.B); !strings.Contains(lin, want) {
		t.Fatalf("rune (0,1) lacks rainbow colour %s", want)
	}
}

func TestSlidersRespectMode(t *testing.T) {
	m := toOption(t, key(testModel(t, 2), "ctrl+o"), optBgMode)
	from := m.state.Paint.Background.From
	m = key(m, "j", "l") // j skips the From heading onto R
	if m.state.Paint.Background.From != from {
		t.Fatal("RGB sliders must be inert in theme mode")
	}
	m = key(toOption(t, m, optBgMode), "l")
	if m.state.Paint.Background.Mode != paint.Solid {
		t.Fatalf("mode = %v, want solid", m.state.Paint.Background.Mode)
	}
	m = key(m, "j", "l") // onto From R
	if m.state.Paint.Background.From.R == from.R {
		t.Fatal("From R should step in solid mode")
	}
}

func TestRowsFitTheirPane(t *testing.T) {
	m := withIndex(testModel(t, 2))
	for _, mode := range []paint.Mode{paint.Theme, paint.Linear} {
		m.state.Paint.Background.Mode = mode
		for _, p := range []struct {
			name  string
			rows  []optRow
			width int
		}{{"filter", filterRows(), sidebarWidth}, {"style", styleRows(), styleWidth}} {
			for i, r := range p.rows {
				if w := lipgloss.Width(m.optionLine(r, true, p.width)); w > p.width {
					t.Errorf("%s row %d (%v mode) is %d columns, pane is %d", p.name, i, mode, w, p.width)
				}
			}
		}
	}
	m.state.SetSimilar("f01")
	for _, r := range filterRows() {
		if r.kind == optSimilar && lipgloss.Width(m.optionLine(r, true, sidebarWidth)) > sidebarWidth {
			t.Error("Similar row overflows the Fonts pane")
		}
	}
}

func TestExportWritesNumberedSVGs(t *testing.T) {
	dir := t.TempDir()
	exportDir = dir
	t.Cleanup(func() { exportDir = "." })

	m := testModel(t, 2)
	m.state.Paint.Background.Mode = paint.Radial
	for range 2 {
		next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
		m = next.(Model)
		next, _ = m.Update(cmd())
		m = next.(Model)
	}
	for _, name := range []string{"f00.svg", "f00-2.svg"} {
		doc, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(doc), "<radialGradient") {
			t.Fatalf("%s lacks the radial background", name)
		}
	}
	if want := "saved " + filepath.Join(dir, "f00-2.svg"); m.status != want {
		t.Fatalf("status %q, want %q", m.status, want)
	}
}

func TestTextBarIsFramedAndLabelled(t *testing.T) {
	for _, text := range []string{"ab", "ab\ncd"} {
		m := testModel(t, 3)
		m.text.SetValue(text)
		m.state.Text = text
		m = key(m, "x")
		m.View()
		v := m.View()
		if h := lipgloss.Height(v); h != 20 {
			t.Fatalf("%q: View is %d lines, want 20", text, h)
		}
		lines := strings.Split(v, "\n")
		if !strings.Contains(lines[0], "Text") {
			t.Fatalf("%q: first line %q lacks the Text label", text, lines[0])
		}
		for i, l := range lines {
			if w := lipgloss.Width(l); w > 59 {
				t.Fatalf("%q: line %d is %d columns, want at most 59", text, i, w)
			}
		}
	}
}

func TestFontsPaneFillsTheBody(t *testing.T) {
	for _, h := range []int{12, 20, 40} {
		for _, indexed := range []bool{false, true} {
			m := testModel(t, 30)
			if indexed {
				m = withIndex(m)
			}
			next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: h})
			m = next.(Model)
			if got := lipgloss.Height(m.sidebarView()); got != m.bodyHeight() {
				t.Errorf("height %d indexed=%v: pane is %d rows, body is %d", h, indexed, got, m.bodyHeight())
			}
		}
	}
	m := testModel(t, 3)
	if v := m.sidebarView(); !strings.Contains(v, "analysing…") || !strings.Contains(v, "3/3 · 0 ● · loading") || !strings.Contains(v, "› copy/svg · ● preview only") {
		t.Fatalf("Fonts pane before the index lacks analysing… or the count:\n%s", v)
	}
}

func TestClearFiltersScrollsIntoTheFilterWindow(t *testing.T) {
	m := withIndex(testModel(t, 3))
	if m.filterHeight() >= len(filterRows()) {
		t.Fatal("test needs a filter window shorter than the checklist")
	}
	m = toOption(t, m, optCollapse)
	for range len(filterRows()) {
		m = key(m, "j")
	}
	if r := filterRows()[m.filterCur.cursor]; r.kind != optClear {
		t.Fatalf("cursor on kind %d, want Clear filters", r.kind)
	}
	if !strings.Contains(m.sidebarView(), "Clear filters") {
		t.Fatalf("Clear filters not scrolled into view:\n%s", m.sidebarView())
	}
}

func TestFiltersZoneInTabCycle(t *testing.T) {
	m := testModel(t, 3)
	m = key(m, "tab")
	if m.focus != focusList {
		t.Fatalf("tab from text = %v, want list", m.focus)
	}
	if m = key(m, "f"); m.focus != focusList {
		t.Fatalf("f in the list = %v, want list", m.focus)
	}
	if m = key(m, "tab"); m.focus != focusFilters {
		t.Fatalf("tab from list = %v, want filters", m.focus)
	}
	if m = key(m, "tab"); m.focus != focusText {
		t.Fatalf("tab from filters = %v, want text (Style closed)", m.focus)
	}
	m = key(m, "ctrl+t", "esc")
	if m.focus != focusList {
		t.Fatalf("esc should reach the list, focus = %v", m.focus)
	}
	if m = key(m, "shift+tab"); m.focus != focusText {
		t.Fatalf("shift+tab from list = %v, want text", m.focus)
	}
}

func TestCtrlTFocusesFilters(t *testing.T) {
	m := testModel(t, 3)
	m.text.SetValue("ab")
	m.text.CursorStart()
	m.text.SetCursor(1)
	m = key(m, "ctrl+t")
	if m.focus != focusFilters || m.text.Value() != "ab" {
		t.Fatalf("focus %v text %q", m.focus, m.text.Value())
	}
	m = settle(key(resize(testModel(t, 3), 70, 24), "ctrl+o", "ctrl+t"))
	if m.panelOpen || m.fontsHidden() || m.focus != focusFilters || !strings.Contains(m.View(), "fuzzy search") {
		t.Fatalf("70 cols: open=%v hidden=%v focus=%v", m.panelOpen, m.fontsHidden(), m.focus)
	}
}

func TestTabCyclesTextFontsStyle(t *testing.T) {
	m := settle(key(resize(testModel(t, 3), 100, 24), "ctrl+o", "ctrl+e"))
	var got []focus
	for range 4 {
		m = key(m, "tab")
		got = append(got, m.focus)
	}
	if want := []focus{focusList, focusFilters, focusStyle, focusText}; fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("tab order %v, want %v", got, want)
	}
	got = nil
	for range 4 {
		m = key(m, "shift+tab")
		got = append(got, m.focus)
	}
	if want := []focus{focusStyle, focusFilters, focusList, focusText}; fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("shift+tab order %v, want %v", got, want)
	}
}

func TestSearchAndListZoneKeys(t *testing.T) {
	m := key(testModel(t, 3), "ctrl+f", "f", "0")
	if m.search.Value() != "f0" || m.focus != focusSearch {
		t.Fatalf("typed into search: %q focus %v", m.search.Value(), m.focus)
	}
	m = key(m, "enter")
	if m.focus != focusList || m.search.Value() != "f0" || m.state.SelectedCount() != 0 {
		t.Fatalf("enter: focus %v query %q picked %d", m.focus, m.search.Value(), m.state.SelectedCount())
	}
	if m = key(m, "esc"); m.focus != focusList || m.search.Value() != "" {
		t.Fatalf("esc in list: focus %v query %q", m.focus, m.search.Value())
	}
	if m = key(m, "/", "f", "esc"); m.focus != focusList || m.search.Value() != "" {
		t.Fatalf("esc in search: focus %v query %q", m.focus, m.search.Value())
	}
	if m = key(m, "ctrl+f"); m.focus != focusSearch {
		t.Fatalf("ctrl+f = %v, want search", m.focus)
	}
}

func TestTextBarEscGoesToFonts(t *testing.T) {
	m := key(testModel(t, 3), "esc")
	if m.focus != focusList {
		t.Fatalf("esc in text = %v, want list", m.focus)
	}
}

func TestEscAndQNeverQuit(t *testing.T) {
	for _, f := range []focus{focusText, focusSearch, focusFilters, focusList, focusStyle} {
		for _, k := range []string{"esc", "q"} {
			m := settle(key(resize(testModel(t, 3), 100, 24), "ctrl+o"))
			m.setFocus(f)
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
			if k == "esc" {
				msg = tea.KeyMsg{Type: tea.KeyEsc}
			}
			_, cmd := m.Update(msg)
			if cmd == nil {
				continue
			}
			if _, quit := cmd().(tea.QuitMsg); quit {
				t.Errorf("%s in focus %v quits", k, f)
			}
		}
	}
	_, cmd := testModel(t, 3).Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if _, quit := cmd().(tea.QuitMsg); !quit {
		t.Error("ctrl+c should quit")
	}
}

func TestEscInStyle(t *testing.T) {
	m := settle(key(resize(testModel(t, 3), 100, 24), "ctrl+o", "esc"))
	if !m.panelOpen || m.focus != focusList {
		t.Fatalf("100 cols: open=%v focus=%v", m.panelOpen, m.focus)
	}
	m = settle(key(resize(testModel(t, 3), 70, 24), "ctrl+o", "esc"))
	if m.panelOpen || m.focus != focusList || m.fontsHidden() {
		t.Fatalf("70 cols: open=%v focus=%v hidden=%v", m.panelOpen, m.focus, m.fontsHidden())
	}
}

func resize(m Model, w, h int) Model {
	next, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	return next.(Model)
}

func assertFits(t *testing.T, m Model, h int) {
	t.Helper()
	v := m.View()
	if got := lipgloss.Height(v); got != h {
		t.Errorf("width %d: View is %d lines, want %d", m.width, got, h)
	}
	for i, l := range strings.Split(v, "\n") {
		if w := lipgloss.Width(l); w >= m.width {
			t.Errorf("width %d: line %d is %d columns", m.width, i, w)
		}
	}
}

func TestFitThresholds(t *testing.T) {
	if bothFitWidth != 79 || styleFitWidth != 47 {
		t.Fatalf("thresholds %d and %d, docs say 79 and 47", bothFitWidth, styleFitWidth)
	}
	for w, hidden := range map[int]bool{79: false, 78: true} {
		m := key(resize(testModel(t, 3), w, 24), "ctrl+o")
		if m.fontsHidden() != hidden {
			t.Errorf("width %d: fontsHidden = %v, want %v", w, m.fontsHidden(), hidden)
		}
	}
	if m := key(resize(testModel(t, 3), 47, 24), "ctrl+o"); !m.panelOpen {
		t.Error("Style should open at 47 columns")
	}
}

func TestFontsStayBesideStyleWhenTheyFit(t *testing.T) {
	m := resize(testModel(t, 3), 100, 24)
	m = settle(key(m, "ctrl+o"))
	if m.fontsHidden() || !strings.Contains(m.View(), "fuzzy search") {
		t.Fatal("Fonts should stay at 100 columns")
	}
	assertFits(t, m, 24)
}

func TestFontsStepAsideBelowSeventyEight(t *testing.T) {
	m := resize(testModel(t, 3), 75, 24)
	m = key(m, "tab", "ctrl+o")
	if !m.fontsHidden() || m.focus != focusStyle {
		t.Fatalf("hidden=%v focus=%v", m.fontsHidden(), m.focus)
	}
	if want := "Fonts hidden to fit Style · ctrl+o brings them back"; m.status != want {
		t.Fatalf("status %q", m.status)
	}
	m = settle(m)
	if strings.Contains(m.View(), "fuzzy search") {
		t.Fatal("Fonts should be hidden")
	}
	if m.previewWidth() != 75-frameCols-styleCols {
		t.Fatalf("preview width %d", m.previewWidth())
	}
	assertFits(t, m, 24)
	if m = key(m, "tab"); m.focus != focusText {
		t.Fatalf("tab from Style = %v, want text", m.focus)
	}
	if m = key(m, "tab"); m.focus != focusStyle {
		t.Fatalf("tab from text = %v, want Style (Fonts skipped)", m.focus)
	}
	if m = key(m, "shift+tab"); m.focus != focusText {
		t.Fatalf("shift+tab = %v, want text", m.focus)
	}
}

func TestFontsReturnAfterTheSlideOut(t *testing.T) {
	m := resize(testModel(t, 3), 75, 24)
	m = settle(key(m, "tab", "ctrl+o"))
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlO})
	m = next.(Model)
	if !m.fontsHidden() {
		t.Fatal("Fonts must stay hidden while Style slides out")
	}
	for range 10 {
		assertFits(t, m, 24)
		next, _ := m.Update(slideMsg{})
		m = next.(Model)
	}
	if m.fontsHidden() || m.focus != focusList || !strings.Contains(m.View(), "fuzzy search") {
		t.Fatalf("hidden=%v focus=%v", m.fontsHidden(), m.focus)
	}
	assertFits(t, m, 24)
}

func TestResizeHidesAndRestoresFonts(t *testing.T) {
	m := resize(testModel(t, 3), 100, 24)
	m = settle(key(m, "tab", "ctrl+o", "esc")) // Style open, focus on the list
	if m.focus != focusList {
		t.Fatalf("focus = %v, want list", m.focus)
	}
	m = resize(m, 70, 24)
	if !m.fontsHidden() || m.focus != focusStyle {
		t.Fatalf("hidden=%v focus=%v", m.fontsHidden(), m.focus)
	}
	assertFits(t, m, 24)
	if m = resize(m, 100, 24); m.fontsHidden() || !strings.Contains(m.View(), "fuzzy search") {
		t.Fatal("Fonts should return at 100 columns")
	}
	assertFits(t, m, 24)
}

func TestStyleRefusedWhenItCannotFit(t *testing.T) {
	m := resize(testModel(t, 3), 46, 24)
	m = key(m, "ctrl+o")
	if m.panelOpen || m.status != "Style needs 47 columns" || m.focus != focusText {
		t.Fatalf("open=%v status=%q focus=%v", m.panelOpen, m.status, m.focus)
	}
	m = resize(testModel(t, 3), 80, 24)
	m = settle(key(m, "ctrl+o"))
	if m = resize(m, 46, 24); m.panelOpen {
		t.Fatal("a resize below 47 should close Style")
	}
	m = settle(m)
	if m.fontsHidden() {
		t.Fatal("with Style closed Fonts never hides")
	}
}

func TestEightyColumnsKeepsFontsAndNarrowPreview(t *testing.T) {
	m := resize(testModel(t, 3), 80, 24)
	if m.previewWidth() != 46 {
		t.Fatalf("closed preview width %d, want 46", m.previewWidth())
	}
	m = settle(key(m, "ctrl+o"))
	if m.previewWidth() != 11 || m.fontsHidden() {
		t.Fatalf("open preview width %d hidden=%v, want 11 and visible", m.previewWidth(), m.fontsHidden())
	}
	assertFits(t, m, 24)
}

func TestCtrlKeysDoNotEditText(t *testing.T) {
	for _, k := range []string{"ctrl+k", "ctrl+e", "ctrl+f"} {
		m := testModel(t, 3)
		m.text.CursorStart()
		m = key(m, k)
		if m.text.Value() != "ab" || m.text.LineInfo().CharOffset != 0 || m.palette.open != (k == "ctrl+k") {
			t.Errorf("%s in the text bar: value %q, column %d", k, m.text.Value(), m.text.LineInfo().CharOffset)
		}
	}
	m := testModel(t, 3)
	m.search.SetValue("abc")
	m.search.SetCursor(0)
	m = key(m, "tab", "ctrl+f", "ctrl+k")
	if m.focus != focusSearch || m.search.Value() != "abc" || m.search.Position() != 0 || !m.palette.open {
		t.Fatalf("search: focus %v value %q position %d palette %v", m.focus, m.search.Value(), m.search.Position(), m.palette.open)
	}
	if m = key(m, "ctrl+e"); m.focus != focusText {
		t.Fatalf("ctrl+e from search = %v, want text", m.focus)
	}
	if m.search.Value() != "abc" || m.search.Position() != 0 {
		t.Fatalf("ctrl+e moved the search cursor to %d", m.search.Position())
	}
}

func TestCtrlEAndCtrlFMoveFocus(t *testing.T) {
	m := testModel(t, 3)
	m.setFocus(focusList)
	if m = key(m, "ctrl+e"); m.focus != focusText {
		t.Fatalf("ctrl+e from list = %v, want text", m.focus)
	}
	if m = key(m, "ctrl+f"); m.focus != focusSearch {
		t.Fatalf("ctrl+f from text = %v, want search", m.focus)
	}
}

func TestCtrlFClosesStyleWhenFontsAreHidden(t *testing.T) {
	m := resize(testModel(t, 3), 70, 24)
	m = settle(key(m, "ctrl+o"))
	if !m.fontsHidden() {
		t.Fatal("Fonts should be hidden at 70 columns")
	}
	m = settle(key(m, "ctrl+f"))
	if m.panelOpen || m.fontsHidden() || m.focus != focusSearch || !strings.Contains(m.View(), "fuzzy search") {
		t.Fatalf("open=%v hidden=%v focus=%v", m.panelOpen, m.fontsHidden(), m.focus)
	}
}

func TestFooterIsTheSameInEveryFocus(t *testing.T) {
	m := resize(testModel(t, 3), 100, 24)
	m = settle(key(m, "ctrl+o"))
	want := m.footerView()
	for _, f := range []focus{focusText, focusSearch, focusFilters, focusList, focusStyle} {
		m.setFocus(f)
		if got := m.footerView(); got != want {
			t.Errorf("focus %v footer %q, want %q", f, got, want)
		}
	}
	m = resize(m, 80, 24)
	got := m.footerView()
	if want := "ctrl+k: actions · tab: pane · ctrl+o: style · ctrl+y: copy · ctrl+c: quit"; got != want {
		t.Fatalf("80-column footer %q, want %q", got, want)
	}
	if got := resize(m, 200, 24).footerView(); !strings.HasSuffix(got, "ctrl+c: quit") {
		t.Fatalf("wide footer %q should end in ctrl+c: quit", got)
	}
}

func TestCountAndLegendFitThePane(t *testing.T) {
	for _, loading := range []bool{false, true} {
		for _, l := range []string{formatCount(9999, 9999, 9999, loading), legendLine()} {
			if strings.Contains(l, "…") || lipgloss.Width(l) > sidebarWidth {
				t.Errorf("%q is %d columns wide or truncated", l, lipgloss.Width(l))
			}
		}
	}
	if got := formatCount(412, 768, 2, true); got != "412/768 · 2 ● · loading" {
		t.Fatalf("count %q", got)
	}
}

func TestFrameAccentFollowsFocus(t *testing.T) {
	trueColour(t)
	m := settle(key(resize(testModel(t, 3), 100, 30), "ctrl+o"))
	accent, dim := styles.label.Render("╭─ "), styles.divider.Render("╭─ ")
	for _, c := range []struct {
		f    focus
		name string
	}{{focusText, "text"}, {focusSearch, "fonts"}, {focusList, "fonts"}, {focusFilters, "filters"}, {focusStyle, "style"}} {
		m.setFocus(c.f)
		frames := map[string]string{"text": m.textBar(), "fonts": m.fontsView(), "filters": m.filtersView(), "style": m.styleView()}
		for name, v := range frames {
			top := strings.Split(v, "\n")[0]
			if focused := strings.Contains(top, accent); focused != (name == c.name) {
				t.Errorf("focus %v: %s frame accented=%v", c.f, name, focused)
			}
			if !strings.Contains(top, accent) && !strings.Contains(top, dim) {
				t.Errorf("focus %v: %s frame top %q has neither border colour", c.f, name, top)
			}
		}
	}
}

func TestPaneFramesAreClosedAndTitled(t *testing.T) {
	m := settle(key(resize(testModel(t, 3), 100, 30), "ctrl+o"))
	for title, v := range map[string]string{"Fonts  ctrl+f": m.fontsView(), "Filters  ctrl+t": m.filtersView(), "Style  ctrl+o": m.styleView(), "Text  ctrl+e": m.textBar()} {
		lines := strings.Split(v, "\n")
		if !strings.Contains(lines[0], title) || !strings.HasPrefix(lines[len(lines)-1], "╰") {
			t.Errorf("%s: frame\n%s", title, v)
		}
	}
	if w := lipgloss.Width(m.sidebarView()); w != sidebarWidth+2 {
		t.Errorf("left column is %d columns", w)
	}
}

func TestCollapseStartsOnAndFoldsIdenticalFonts(t *testing.T) {
	m := testModel(t, 4)
	m.state.Collapse = true
	m.itemsChanged()
	if len(m.items) != 4 {
		t.Fatalf("collapse must wait for the index, got %d rows", len(m.items))
	}
	m = withIndex(m)
	if len(m.items) != 1 || m.items[0].Dups != 3 {
		t.Fatalf("items %+v, want one row with 3 duplicates", m.items)
	}
	if !New(m.cat, "ab", display.AlignLeft).state.Collapse {
		t.Fatal("a fresh model should start with Collapse on")
	}
}

func TestStyleFrameHasRightPadding(t *testing.T) {
	m := settle(key(resize(testModel(t, 3), 100, 30), "ctrl+o"))
	lines := strings.Split(m.styleView(), "\n")
	for i, l := range lines[1 : len(lines)-1] {
		if !strings.HasSuffix(l, " │") {
			t.Errorf("row %d %q: no space before the right border", i, l)
		}
	}
	if w := lipgloss.Width(lines[0]); w != styleCols {
		t.Errorf("Style frame is %d columns, styleCols = %d", w, styleCols)
	}
}
