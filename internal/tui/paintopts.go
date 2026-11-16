package tui

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/jnaO/figlet-font-preview/internal/controls"
	"github.com/jnaO/figlet-font-preview/internal/display"
	"github.com/jnaO/figlet-font-preview/internal/paint"
)

// paintRows lists the colour rows of the Style pane.
func paintRows() []optRow {
	return []optRow{
		{kind: optHeading, label: "Rainbow"},
		{kind: optSlider, slider: controls.Seed},
		{kind: optSlider, slider: controls.Spread},
		{kind: optSlider, slider: controls.Freq},
		{kind: optHeading, label: "Background"},
		{kind: optBgMode},
		{kind: optHeading, label: "From"},
		{kind: optSlider, slider: controls.FromR},
		{kind: optSlider, slider: controls.FromG},
		{kind: optSlider, slider: controls.FromB},
		{kind: optHeading, label: "To"},
		{kind: optSlider, slider: controls.ToR},
		{kind: optSlider, slider: controls.ToG},
		{kind: optSlider, slider: controls.ToB},
		{kind: optSlider, slider: controls.Angle},
	}
}

const barCells = 10

// sliderLine is "    Label     ━━━●──────  value": 4 + 10 + 10 + 6 columns.
func sliderLine(st *paint.Style, k controls.Slider) string {
	sp := k.Spec()
	v := sp.Get(st)
	pos := int(math.Round((v - sp.Lo) / (sp.Hi - sp.Lo) * (barCells - 1)))
	bar := strings.Repeat("━", pos) + "●" + strings.Repeat("─", barCells-1-pos)
	value := fmt.Sprintf("%6s", fmt.Sprintf(sp.Format, v))
	label := fmt.Sprintf("    %-10s", sp.Label)
	if !sp.Enabled(st.Background.Mode) {
		return styles.dim.Render(label + bar + value)
	}
	return label + styles.selected.Render(bar) + value
}

// headingLine renders a colour heading; From and To carry a swatch and hex.
func headingLine(st *paint.Style, label string) string {
	h := "  " + styles.header.Render(label)
	var c paint.RGB
	switch label {
	case "From":
		c = st.Background.From
	case "To":
		c = st.Background.To
	default:
		return h
	}
	if active := map[string]func(paint.Mode) bool{"From": controls.FromActive, "To": controls.ToActive}[label]; !active(st.Background.Mode) {
		return "  " + styles.dim.Render(label+"  "+c.Hex())
	}
	swatch := lipgloss.NewStyle().Background(colour(c)).Render("  ")
	return h + " " + swatch + " " + styles.dim.Render(c.Hex())
}

func alignLine(a display.Align) string {
	return "    Align     ‹ " + styles.label.Render(a.String()) + " ›"
}

func bgModeLine(st *paint.Style) string {
	return "    Mode      ‹ " + styles.label.Render(st.Background.Mode.String()) + " ›"
}
