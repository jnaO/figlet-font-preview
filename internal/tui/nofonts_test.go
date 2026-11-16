package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jnaO/figlet-font-preview/internal/catalog"
	"github.com/jnaO/figlet-font-preview/internal/display"
)

func TestNoFontsMessage(t *testing.T) {
	tests := []struct {
		goos, hint string
	}{
		{"darwin", "brew install figlet toilet"},
		{"windows", `%USERPROFILE%\figlet-fonts`},
		{"linux", "package manager"},
	}
	for _, tt := range tests {
		t.Run(tt.goos, func(t *testing.T) {
			msg := noFontsMessage([]string{"/a", "/b"}, tt.goos)
			for _, want := range []string{"No fonts found.", "/a", "/b", tt.hint, "https://", "not affiliated"} {
				if !strings.Contains(msg, want) {
					t.Errorf("missing %q in:\n%s", want, msg)
				}
			}
		})
	}
}

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
