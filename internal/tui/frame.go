package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// frame draws body in a rounded border outer columns wide and height rows
// tall, with "title  key" in the top edge. The border is dim and the title
// plain while unfocused; both take the accent while focused. lipgloss has no
// border-title API, so the top edge is built by hand.
func frame(title, key, body string, outer, height, padL, padR int, focused bool) string {
	border, label := styles.divider, styles.header
	if focused {
		border, label = styles.label, styles.label
	}
	used := lipgloss.Width("╭─ " + title + "  " + key + " ╮")
	top := border.Render("╭─ ") + label.Render(title) + border.Render("  "+key+" "+strings.Repeat("─", max(outer-used, 0))+"╮")
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderTop(false).
		BorderForeground(border.GetForeground()).PaddingLeft(padL).PaddingRight(padR).
		Width(outer - 2).Height(height).MaxHeight(height + 1)
	return top + "\n" + box.Render(body)
}

// overlay writes box over bg with its top-left corner at column x, row y.
// Each covered row keeps bg's cells left and right of the box, with their
// colours; rows of box that fall below bg are dropped.
func overlay(bg, box string, x, y int) string {
	lines := strings.Split(bg, "\n")
	for i, b := range strings.Split(box, "\n") {
		if y+i < 0 || y+i >= len(lines) {
			continue
		}
		l := lines[y+i]
		left := ansi.Truncate(l, x, "")
		left += strings.Repeat(" ", max(x-ansi.StringWidth(left), 0))
		lines[y+i] = left + ansi.ResetStyle + b + ansi.TruncateLeft(l, x+ansi.StringWidth(b), "")
	}
	return strings.Join(lines, "\n")
}
