package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
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
