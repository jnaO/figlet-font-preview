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
		name             string
		y, start, end, w int
	}{
		{"fully visible stays", 0, 2, 8, 0},
		{"exactly fits at bottom stays", 0, 4, 10, 0},
		{"below scrolls just enough", 0, 8, 14, 4},
		{"above scrolls to its top", 20, 12, 18, 12},
		{"taller than view shows top", 0, 5, 30, 5},
		{"taller than view above shows top", 40, 5, 30, 5},
	}
	for _, c := range cases {
		if got := revealOffset(c.y, h, c.start, c.end); got != c.w {
			t.Errorf("%s: got %d, want %d", c.name, got, c.w)
		}
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
	m = key(m, "tab", "tab", "tab")
	for range 12 {
		m = key(m, "down")
	}
	name := m.cursorName()
	before := screenRow(m, name)
	if m.preview.YOffset == 0 {
		t.Fatal("expected the preview to have scrolled")
	}
	m = key(m, "tab", "c", "d", "enter", "e", "f") // list → text; add a second line
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
	if m.previewWidth() != full-styleWidth-1 {
		t.Fatalf("preview width %d, want %d", m.previewWidth(), full-styleWidth-1)
	}
	m = settle(key(m, "ctrl+o"))
	if m.panelOpen || m.focus != focusText || m.panelW != 0 {
		t.Fatalf("after close: open=%v focus=%v width=%d", m.panelOpen, m.focus, m.panelW)
	}
	m = settle(key(m, "tab", "ctrl+o", "esc")) // esc closes too, back to search
	if m.panelOpen || m.focus != focusSearch {
		t.Fatalf("esc: open=%v focus=%v", m.panelOpen, m.focus)
	}
	m = key(m, "ctrl+o", "tab") // tab leaves the pane open
	if !m.panelOpen || m.focus != focusText {
		t.Fatalf("tab: open=%v focus=%v", m.panelOpen, m.focus)
	}
}

func TestStyleAlignCycler(t *testing.T) {
	m := key(testModel(t, 3), "ctrl+o") // first row is Align
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
	m = key(m, "tab", "tab", "tab", "down", "down") // cursor on f02
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

func TestSlidersWrapClampAndRespectMode(t *testing.T) {
	st := paint.Default()
	st.Rainbow.Seed = 254
	stepSlider(&st, sSeed, 1)
	if st.Rainbow.Seed != 2 {
		t.Fatalf("seed should wrap modulo 256: %d", st.Rainbow.Seed)
	}
	st.Background.Angle = 0
	stepSlider(&st, sAngle, -1)
	if st.Background.Angle != 345 {
		t.Fatalf("angle should wrap to 345: %d", st.Background.Angle)
	}
	st.Rainbow.Freq = paint.FreqMax
	stepSlider(&st, sFreq, 1)
	if st.Rainbow.Freq != paint.FreqMax {
		t.Fatalf("frequency should clamp: %v", st.Rainbow.Freq)
	}
	st.Rainbow.Freq = 0.05
	stepSlider(&st, sFreq, 1)
	if st.Rainbow.Freq != 0.06 {
		t.Fatalf("frequency step drifted: %v", st.Rainbow.Freq)
	}
	st.Background.From.R = 250
	stepSlider(&st, sFromR, 1)
	if st.Background.From.R != 255 {
		t.Fatalf("RGB should clamp at 255: %d", st.Background.From.R)
	}

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

func TestSafeName(t *testing.T) {
	if got := safeName(`C64-fonts/gothic:a*b?"<>|\\x`); got != "C64-fonts_gothic_a_b_______x" {
		t.Fatalf("safeName = %q", got)
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
	if v := m.sidebarView(); !strings.Contains(v, "analysing…") || !strings.Contains(v, "3 of 3 fonts · 0 selected") {
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
	m = key(m, "tab", "tab")
	if m.focus != focusFilters {
		t.Fatalf("focus = %v, want filters", m.focus)
	}
	if m = key(m, "esc"); m.focus != focusList {
		t.Fatalf("esc should reach the list, focus = %v", m.focus)
	}
	if m = key(m, "shift+tab"); m.focus != focusFilters {
		t.Fatalf("shift+tab from list, focus = %v", m.focus)
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
	if bothFitWidth != 76 || styleFitWidth != 45 {
		t.Fatalf("thresholds %d and %d, docs say 76 and 45", bothFitWidth, styleFitWidth)
	}
	for w, hidden := range map[int]bool{76: false, 75: true} {
		m := key(resize(testModel(t, 3), w, 24), "ctrl+o")
		if m.fontsHidden() != hidden {
			t.Errorf("width %d: fontsHidden = %v, want %v", w, m.fontsHidden(), hidden)
		}
	}
	if m := key(resize(testModel(t, 3), 45, 24), "ctrl+o"); !m.panelOpen {
		t.Error("Style should open at 45 columns")
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

func TestFontsStepAsideBelowSeventySix(t *testing.T) {
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
	if m.fontsHidden() || m.focus != focusSearch || !strings.Contains(m.View(), "fuzzy search") {
		t.Fatalf("hidden=%v focus=%v", m.fontsHidden(), m.focus)
	}
	assertFits(t, m, 24)
}

func TestResizeHidesAndRestoresFonts(t *testing.T) {
	m := resize(testModel(t, 3), 100, 24)
	m = settle(key(m, "tab", "tab", "tab", "ctrl+o", "tab", "tab", "tab", "tab")) // Style open, focus on the list
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
	m := resize(testModel(t, 3), 44, 24)
	m = key(m, "ctrl+o")
	if m.panelOpen || m.status != "Style needs 45 columns" || m.focus != focusText {
		t.Fatalf("open=%v status=%q focus=%v", m.panelOpen, m.status, m.focus)
	}
	m = resize(testModel(t, 3), 80, 24)
	m = settle(key(m, "ctrl+o"))
	if m = resize(m, 44, 24); m.panelOpen {
		t.Fatal("a resize below 45 should close Style")
	}
	m = settle(m)
	if m.fontsHidden() {
		t.Fatal("with Style closed Fonts never hides")
	}
}

func TestEightyColumnsKeepsFontsAndNarrowPreview(t *testing.T) {
	m := resize(testModel(t, 3), 80, 24)
	if m.previewWidth() != 47 {
		t.Fatalf("closed preview width %d, want 47", m.previewWidth())
	}
	m = settle(key(m, "ctrl+o"))
	if m.previewWidth() != 14 || m.fontsHidden() {
		t.Fatalf("open preview width %d hidden=%v, want 14 and visible", m.previewWidth(), m.fontsHidden())
	}
	assertFits(t, m, 24)
}
