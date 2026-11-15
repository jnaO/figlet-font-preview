package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func TestQuestionMarkTypesInTextFieldsAndOpensHelpElsewhere(t *testing.T) {
	m := key(testModel(t, 3), "?")
	if m.help.open || m.text.Value() != "ab?" {
		t.Fatalf("text bar: open=%v value %q", m.help.open, m.text.Value())
	}
	m = key(testModel(t, 3), "tab", "ctrl+f", "?")
	if m.help.open || m.search.Value() != "?" {
		t.Fatalf("search: open=%v value %q", m.help.open, m.search.Value())
	}
	for _, f := range []focus{focusList, focusFilters, focusStyle} {
		m := testModel(t, 3)
		m.setFocus(f)
		m = key(m, "?")
		if !m.help.open || m.focus != f || m.text.Value() != "ab" || m.search.Value() != "" {
			t.Errorf("focus %v: open=%v focus=%v text=%q search=%q", f, m.help.open, m.focus, m.text.Value(), m.search.Value())
		}
	}
}

func TestHelpClosesOnEscAndQuestionMarkWithoutQuitting(t *testing.T) {
	for _, k := range []tea.KeyMsg{{Type: tea.KeyEsc}, {Type: tea.KeyRunes, Runes: []rune("?")}} {
		m := key(testModel(t, 3), "tab", "?")
		m, cmd := send(m, k)
		if m.help.open || m.focus != focusList || emits(cmd, tea.Quit()) {
			t.Errorf("%v: open=%v focus=%v", k, m.help.open, m.focus)
		}
	}
	m := key(testModel(t, 3), "tab", "?")
	if _, cmd := send(m, tea.KeyMsg{Type: tea.KeyCtrlC}); !emits(cmd, tea.Quit()) {
		t.Error("ctrl+c inside help should quit")
	}
}

func TestHelpHandsAppKeysOver(t *testing.T) {
	m := key(testModel(t, 3), "tab", "?", "ctrl+k")
	if m.help.open || !m.palette.open {
		t.Fatalf("ctrl+k from help: help=%v palette=%v", m.help.open, m.palette.open)
	}
	m = key(testModel(t, 3), "tab", "?", "ctrl+e")
	if m.help.open || m.focus != focusText || !m.text.Focused() {
		t.Fatalf("ctrl+e from help: help=%v focus=%v", m.help.open, m.focus)
	}
	m = key(testModel(t, 3), "tab", "?", "z", "x")
	if !m.help.open || m.state.SelectedCount() != 0 {
		t.Fatalf("other keys should be ignored: open=%v picks=%d", m.help.open, m.state.SelectedCount())
	}
}

func TestPaletteOpensHelpFromTheTextBar(t *testing.T) {
	m := key(testModel(t, 3), "ctrl+k", "h", "e", "l", "p", "enter")
	if !m.help.open || m.palette.open || m.focus != focusText || m.text.Value() != "ab" {
		t.Fatalf("help=%v palette=%v focus=%v text=%q", m.help.open, m.palette.open, m.focus, m.text.Value())
	}
}

func TestHelpListsEveryActionWithItsKey(t *testing.T) {
	m := key(resize(testModel(t, 3), 200, 60), "tab", "?")
	v := ansi.Strip(m.View())
	for _, want := range []string{"Help", "Save cursor font as SVG", "ctrl+s", "Pick or unpick cursor font", "space", "Keys and help", "Fonts list", "Filters", "Style"} {
		if !strings.Contains(v, want) {
			t.Errorf("help lacks %q", want)
		}
	}
}

func TestHelpFitsAndScrollsOnShortTerminals(t *testing.T) {
	for _, sz := range [][2]int{{100, 30}, {80, 24}, {80, 20}, {77, 24}, {46, 24}} {
		m := key(resize(testModel(t, 3), sz[0], sz[1]), "tab", "?")
		assertFits(t, m, sz[1])
	}
	m := key(resize(testModel(t, 3), 80, 20), "tab", "?")
	top := ansi.Strip(m.View())
	if strings.Contains(top, "closes") {
		t.Fatal("the last line should be off screen at 80x20")
	}
	for range 10 {
		m = key(m, "pgdown")
	}
	end := ansi.Strip(m.View())
	if !strings.Contains(end, "closes") || end == top {
		t.Fatal("pgdown never reached the last line")
	}
	assertFits(t, m, 20)
	if m.preview.YOffset != 0 {
		t.Error("pgdown scrolled the preview, not the help")
	}
	m = key(m, "up")
	if m.help.view.YOffset >= m.help.view.TotalLineCount() {
		t.Error("up did not scroll")
	}
	for range 60 {
		m = key(m, "k")
	}
	if !m.help.view.AtTop() {
		t.Error("k never reached the top")
	}
}
