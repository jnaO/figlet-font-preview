package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/jnaO/figlet-font-preview/internal/action"
)

func send(m Model, msg tea.Msg) (Model, tea.Cmd) {
	next, cmd := m.Update(msg)
	return next.(Model), cmd
}

// emits reports whether running cmd, or any command in a batch, yields want.
func emits(cmd tea.Cmd, want tea.Msg) bool {
	if cmd == nil {
		return false
	}
	switch msg := cmd().(type) {
	case tea.BatchMsg:
		for _, c := range msg {
			if emits(c, want) {
				return true
			}
		}
		return false
	default:
		return msg == want
	}
}

func paletteTitles(m Model) []string {
	var out []string
	for _, r := range m.palette.rows {
		out = append(out, r.act.Title)
	}
	return out
}

func TestOverlaySplicesTheBoxAndKeepsWidths(t *testing.T) {
	trueColour(t)
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	blue := lipgloss.NewStyle().Background(lipgloss.Color("#0000ff"))
	line := red.Render("abcdefghij") + blue.Render("klmnopqrst")
	bg := strings.Join([]string{line, line, line, line}, "\n")
	box := "[12]\n[34]"
	got := strings.Split(overlay(bg, box, 8, 1), "\n")
	if len(got) != 4 || got[0] != line || got[3] != line {
		t.Fatalf("rows outside the box changed: %q", got)
	}
	if plain := ansi.Strip(got[2]); plain != "abcdefgh[34]mnopqrst" {
		t.Errorf("row 2 = %q", plain)
	}
	if plain := ansi.Strip(got[1]); plain != "abcdefgh[12]mnopqrst" {
		t.Errorf("row 1 = %q", plain)
	}
	for i, l := range got {
		if ansi.StringWidth(l) != 20 {
			t.Errorf("row %d is %d columns, want 20", i, ansi.StringWidth(l))
		}
	}
	if !strings.Contains(got[1], "\x1b[m[12]") {
		t.Errorf("box not preceded by a reset: %q", got[1])
	}
	// the remainder keeps the background colour of the cut run
	if !strings.Contains(got[1], "48;2;0;0;255") {
		t.Errorf("right remainder lost its colour: %q", got[1])
	}
}

func TestOverlayPadsShortLinesAndClipsBelow(t *testing.T) {
	got := strings.Split(overlay("ab\ncd", "XY\nZZ\nQQ", 4, 1), "\n")
	if len(got) != 2 || got[0] != "ab" || ansi.Strip(got[1]) != "cd  XY" {
		t.Fatalf("got %q", got)
	}
}

func TestCtrlKOpensFromEveryFocus(t *testing.T) {
	for _, f := range []focus{focusText, focusSearch, focusFilters, focusList, focusStyle} {
		m := testModel(t, 3)
		m.search.SetValue("ab")
		m.setFocus(f)
		m = key(m, "ctrl+k")
		if !m.palette.open || m.focus != f || m.text.Value() != "ab" || m.search.Value() != "ab" {
			t.Errorf("focus %v: open=%v focus=%v text=%q search=%q", f, m.palette.open, m.focus, m.text.Value(), m.search.Value())
		}
		if m.text.Focused() || m.search.Focused() {
			t.Errorf("focus %v: a field stays focused under the palette", f)
		}
		m = key(m, "esc")
		if m.palette.open || m.focus != f {
			t.Errorf("focus %v: esc left open=%v focus=%v", f, m.palette.open, m.focus)
		}
		if (f == focusText) != m.text.Focused() || (f == focusSearch) != m.search.Focused() {
			t.Errorf("focus %v: cursor not restored", f)
		}
	}
}

func TestPaletteListsActionsInPaletteOrder(t *testing.T) {
	m := key(testModel(t, 3), "ctrl+k")
	want := []string{"Copy cursor font", "Save cursor font as SVG", "Edit text", "Search fonts", "Go to filters",
		"Show or hide Style", "Pick or unpick cursor font", "Clear picks", "Keys and help", "Quit"}
	if got := paletteTitles(m); strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("empty query lists %q, want %q", got, want)
	}
}

func TestPaletteOrderCoversEveryActionOnce(t *testing.T) {
	seen := map[action.ID]int{}
	for _, id := range paletteOrder {
		seen[id]++
	}
	for _, a := range action.All() {
		want := 1
		if a.ID == action.Palette {
			want = 0
		}
		if seen[a.ID] != want {
			t.Errorf("%s appears %d times in paletteOrder, want %d", a.ID, seen[a.ID], want)
		}
	}
	if len(seen) != len(action.All())-1 {
		t.Errorf("paletteOrder has %d distinct ids, want %d", len(seen), len(action.All())-1)
	}
}

func TestPaletteFiltersAsYouType(t *testing.T) {
	m := key(testModel(t, 3), "ctrl+k", "s", "t", "y")
	if got := paletteTitles(m); len(got) == 0 || got[0] != "Show or hide Style" {
		t.Fatalf("sty lists %q", got)
	}
	m = key(m, "z", "z", "z")
	if len(m.palette.rows) != 0 || !strings.Contains(m.View(), "No matching actions") {
		t.Fatalf("zzz: %d rows", len(m.palette.rows))
	}
	m = key(m, "enter")
	if m.palette.open {
		t.Fatal("enter on no matches should close")
	}
}

func TestPaletteArrowsMoveTheCursor(t *testing.T) {
	m := key(testModel(t, 3), "ctrl+k", "down", "down", "up")
	if m.palette.cursor != 1 {
		t.Fatalf("cursor %d, want 1", m.palette.cursor)
	}
	if m.palette.input.Value() != "" {
		t.Fatalf("arrows typed %q", m.palette.input.Value())
	}
	m = key(m, "up", "up")
	if m.palette.cursor != 0 {
		t.Fatalf("cursor %d, want 0", m.palette.cursor)
	}
}

func TestPaletteEnterRunsStyleThroughToggleStyle(t *testing.T) {
	for _, w := range []int{100, 77, 46} {
		want := settle(key(resize(testModel(t, 3), w, 24), "ctrl+o"))
		m := settle(key(resize(testModel(t, 3), w, 24), "ctrl+k", "s", "t", "y", "enter"))
		if m.palette.open || m.panelOpen != want.panelOpen || m.fontsHidden() != want.fontsHidden() || m.status != want.status || m.focus != want.focus {
			t.Errorf("width %d: open=%v hidden=%v status=%q focus=%v; ctrl+o gives open=%v hidden=%v status=%q focus=%v",
				w, m.panelOpen, m.fontsHidden(), m.status, m.focus, want.panelOpen, want.fontsHidden(), want.status, want.focus)
		}
	}
	if m := key(resize(testModel(t, 3), 77, 24), "ctrl+k", "s", "t", "y", "enter"); !m.panelOpen || !m.fontsHidden() {
		t.Error("77 columns: Style should open and hide Fonts")
	}
	if m := key(resize(testModel(t, 3), 46, 24), "ctrl+k", "s", "t", "y", "enter"); m.panelOpen || m.status == "" {
		t.Errorf("46 columns: open=%v status=%q, want refused with a status", m.panelOpen, m.status)
	}
}

func TestPaletteRunsZoneActionsFromAnyFocus(t *testing.T) {
	m := testModel(t, 3)
	m = key(m, "ctrl+k", "u", "n", "p", "i", "c", "k", "enter")
	if m.state.SelectedCount() != 1 || m.focus != focusText {
		t.Fatalf("pick: selected %d focus %v", m.state.SelectedCount(), m.focus)
	}
	m = key(m, "ctrl+k", "c", "l", "e", "a", "r", "enter")
	if m.state.SelectedCount() != 0 {
		t.Fatalf("clear picks left %d", m.state.SelectedCount())
	}
}

func TestPaletteEscNeverQuitsAndCtrlCQuits(t *testing.T) {
	m := key(testModel(t, 3), "ctrl+k")
	m, cmd := send(m, tea.KeyMsg{Type: tea.KeyEsc})
	if emits(cmd, tea.Quit()) || m.palette.open {
		t.Fatal("esc quit or left the palette open")
	}
	m = key(m, "ctrl+k")
	_, cmd = send(m, tea.KeyMsg{Type: tea.KeyCtrlC})
	if !emits(cmd, tea.Quit()) {
		t.Fatal("ctrl+c inside the palette should quit")
	}
	m = key(m, "ctrl+k") // ctrl+k closes
	if m.palette.open {
		t.Fatal("ctrl+k should close the palette")
	}
}

func TestPaletteLetsOtherAppKeysThroughWithTheirMeaning(t *testing.T) {
	m := key(testModel(t, 3), "tab", "ctrl+k", "ctrl+e")
	if m.palette.open || m.focus != focusText || !m.text.Focused() {
		t.Fatalf("ctrl+e: open=%v focus=%v", m.palette.open, m.focus)
	}
	m = key(m, "ctrl+k", "ctrl+t")
	if m.palette.open || m.focus != focusFilters {
		t.Fatalf("ctrl+t: open=%v focus=%v", m.palette.open, m.focus)
	}
	m = key(m, "ctrl+k", "tab")
	if m.palette.open || m.focus == focusFilters {
		t.Fatalf("tab: open=%v focus=%v", m.palette.open, m.focus)
	}
}

func TestPaletteViewFitsTheScreen(t *testing.T) {
	for _, w := range []int{100, 80, 77, 46} {
		m := key(resize(testModel(t, 3), w, 24), "ctrl+k")
		assertFits(t, m, 24)
		if !strings.Contains(m.View(), "Actions") {
			t.Errorf("width %d: no palette in view", w)
		}
		m = key(m, "s", "t", "y")
		assertFits(t, m, 24)
	}
}

func TestPaletteTypingEditsTheQueryNotTheText(t *testing.T) {
	m := key(testModel(t, 3), "ctrl+k", "a", "b")
	if m.palette.input.Value() != "ab" || m.text.Value() != "ab" || m.search.Value() != "" {
		t.Fatalf("query %q text %q search %q", m.palette.input.Value(), m.text.Value(), m.search.Value())
	}
}
