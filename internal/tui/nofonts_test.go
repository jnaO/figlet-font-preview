package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jnaO/figlet-font-preview/internal/catalog"
	"github.com/jnaO/figlet-font-preview/internal/display"
)

func TestNoFontsScreenFits(t *testing.T) {
	m := New(catalog.Discover(nil), "ab", display.AlignLeft)
	<-m.loaded
	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	view := next.(Model).View()
	t.Log("\n" + view)
	for _, want := range []string{"No fonts found.", "not affiliated"} {
		if !strings.Contains(view, want) {
			t.Errorf("view missing %q", want)
		}
	}
}
