package tui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/jnaO/figlet-font-preview/internal/paint"
)

// tintColour is the cursor font's theme background; themeSeed reads it too.
var tintColour = lipgloss.AdaptiveColor{Light: "#EEF0F6", Dark: "#323846"}

// themeSeed returns the From and To colours the background sliders start at:
// the resolved tint, and black on a dark terminal or white on a light one.
func themeSeed() (from, to paint.RGB) {
	hex, to := tintColour.Light, paint.RGB{R: 255, G: 255, B: 255}
	if lipgloss.HasDarkBackground() {
		hex, to = tintColour.Dark, paint.RGB{}
	}
	from, _ = paint.ParseHex(hex)
	return from, to
}

var accent = lipgloss.AdaptiveColor{Light: "#8839EF", Dark: "#CBA6F7"}

var styles = struct {
	label, dim, err, rule, divider, header, headerSelected,
	selected, match, cursor, cursorInactive, tint lipgloss.Style
}{
	label:          lipgloss.NewStyle().Foreground(accent).Bold(true),
	dim:            lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "245", Dark: "243"}),
	err:            lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "160", Dark: "203"}),
	rule:           lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "252", Dark: "238"}),
	divider:        lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "240"}),
	header:         lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "240", Dark: "250"}).Bold(true),
	headerSelected: lipgloss.NewStyle().Foreground(accent).Bold(true),
	selected:       lipgloss.NewStyle().Foreground(accent),
	match:          lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#D20F39", Dark: "#F9E2AF"}).Bold(true),
	cursor:         lipgloss.NewStyle().Foreground(accent).Bold(true),
	cursorInactive: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "240"}),
	tint:           lipgloss.NewStyle().Background(tintColour),
}
