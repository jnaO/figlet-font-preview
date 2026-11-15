package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/jnaO/figlet-font-preview/internal/paint"
)

func colour(c paint.RGB) lipgloss.Color { return lipgloss.Color(c.Hex()) }

// bgStyle is the cursor font's flat background: the theme tint, or From in
// every other mode (headers and placeholder lines are not gradient-painted).
func bgStyle(st paint.Style) lipgloss.Style {
	if st.Background.Mode == paint.Theme {
		return styles.tint
	}
	return lipgloss.NewStyle().Background(colour(st.Background.From))
}

// lolcat colours rune col of row with st.Rainbow (col counts runes, centring
// spaces included; spaces get background only) and pads every line to width.
// The background is painted per cell over the width × len(lines) box, except
// in Theme and Solid mode where it is flat.
func lolcat(lines []string, width int, st paint.Style) []string {
	bg := st.Background
	flat := bg.Mode == paint.Theme || bg.Mode == paint.Solid
	base := bgStyle(st)
	out := make([]string, len(lines))
	for row, line := range lines {
		var b strings.Builder
		cell := func(col int, r rune) {
			s := base
			if !flat {
				c, _ := bg.At(col, row, width, len(lines))
				s = lipgloss.NewStyle().Background(colour(c))
			}
			if r != ' ' {
				s = s.Foreground(colour(st.Rainbow.At(row, col)))
			}
			b.WriteString(s.Render(string(r)))
		}
		col := 0
		for _, r := range line {
			cell(col, r)
			col++
		}
		switch {
		case col >= width:
		case flat:
			b.WriteString(base.Render(strings.Repeat(" ", width-col)))
		default:
			for ; col < width; col++ {
				cell(col, ' ')
			}
		}
		out[row] = b.String()
	}
	return out
}

// tintLine renders an already-plain line on the flat background, padded.
func tintLine(s string, width int, st paint.Style) string {
	return bgStyle(st).Render(s + strings.Repeat(" ", max(width-lipgloss.Width(s), 0)))
}
