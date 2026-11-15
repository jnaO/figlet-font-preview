package tui

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/jnaO/figlet-font-preview/internal/paint"
)

// slider is one numeric colour setting in the options pane.
type slider int

const (
	sSeed slider = iota
	sSpread
	sFreq
	sFromR
	sFromG
	sFromB
	sToR
	sToG
	sToB
	sAngle
)

type sliderSpec struct {
	label        string
	lo, hi, step float64
	period       float64 // > 0: values wrap modulo period instead of clamping
	format       string  // value format
	field        func(st *paint.Style) any
	active       func(paint.Mode) bool // nil: always active
}

func fromActive(m paint.Mode) bool { return m != paint.Theme }
func toActive(m paint.Mode) bool   { return m == paint.Linear || m == paint.Radial }

var sliders = map[slider]sliderSpec{
	sSeed:   {label: "Seed", hi: paint.SeedMax, step: 4, period: paint.SeedMax + 1, format: "%.0f", field: func(st *paint.Style) any { return &st.Rainbow.Seed }},
	sSpread: {label: "Spread", lo: paint.SpreadMin, hi: paint.SpreadMax, step: .5, format: "%.1f", field: func(st *paint.Style) any { return &st.Rainbow.Spread }},
	sFreq:   {label: "Frequency", lo: paint.FreqMin, hi: paint.FreqMax, step: .01, format: "%.2f", field: func(st *paint.Style) any { return &st.Rainbow.Freq }},
	sFromR:  rgbSpec("R", fromActive, func(st *paint.Style) *uint8 { return &st.Background.From.R }),
	sFromG:  rgbSpec("G", fromActive, func(st *paint.Style) *uint8 { return &st.Background.From.G }),
	sFromB:  rgbSpec("B", fromActive, func(st *paint.Style) *uint8 { return &st.Background.From.B }),
	sToR:    rgbSpec("R", toActive, func(st *paint.Style) *uint8 { return &st.Background.To.R }),
	sToG:    rgbSpec("G", toActive, func(st *paint.Style) *uint8 { return &st.Background.To.G }),
	sToB:    rgbSpec("B", toActive, func(st *paint.Style) *uint8 { return &st.Background.To.B }),
	sAngle: {label: "Angle", hi: paint.AngleMax, step: 15, period: 360, format: "%.0f°", field: func(st *paint.Style) any { return &st.Background.Angle },
		active: func(m paint.Mode) bool { return m == paint.Linear }},
}

func rgbSpec(label string, active func(paint.Mode) bool, f func(*paint.Style) *uint8) sliderSpec {
	return sliderSpec{label: label, hi: 255, step: 8, format: "%.0f", field: func(st *paint.Style) any { return f(st) }, active: active}
}

// enabled reports whether the slider does anything in background mode m.
func (sp sliderSpec) enabled(m paint.Mode) bool { return sp.active == nil || sp.active(m) }

func (sp sliderSpec) get(st *paint.Style) float64 {
	switch p := sp.field(st).(type) {
	case *int:
		return float64(*p)
	case *uint8:
		return float64(*p)
	case *float64:
		return *p
	}
	return 0
}

func (sp sliderSpec) set(st *paint.Style, v float64) {
	switch p := sp.field(st).(type) {
	case *int:
		*p = int(math.Round(v))
	case *uint8:
		*p = uint8(math.Round(v))
	case *float64:
		*p = math.Round(v*100) / 100
	}
}

// stepSlider moves k by dir steps, wrapping or clamping at the ends.
func stepSlider(st *paint.Style, k slider, dir int) {
	sp := sliders[k]
	v := sp.get(st) + float64(dir)*sp.step
	if sp.period > 0 {
		v = math.Mod(math.Mod(v, sp.period)+sp.period, sp.period)
	} else {
		v = math.Max(sp.lo, math.Min(sp.hi, v))
	}
	sp.set(st, v)
}

// paintRows lists the colour rows of the options pane.
func paintRows() []optRow {
	return []optRow{
		{kind: optHeading, label: "Rainbow"},
		{kind: optSlider, slider: sSeed},
		{kind: optSlider, slider: sSpread},
		{kind: optSlider, slider: sFreq},
		{kind: optHeading, label: "Background"},
		{kind: optBgMode},
		{kind: optHeading, label: "From"},
		{kind: optSlider, slider: sFromR},
		{kind: optSlider, slider: sFromG},
		{kind: optSlider, slider: sFromB},
		{kind: optHeading, label: "To"},
		{kind: optSlider, slider: sToR},
		{kind: optSlider, slider: sToG},
		{kind: optSlider, slider: sToB},
		{kind: optSlider, slider: sAngle},
	}
}

const barCells = 10

// sliderLine is "    Label     ━━━●──────  value": 4 + 10 + 10 + 6 columns.
func sliderLine(st *paint.Style, k slider) string {
	sp := sliders[k]
	v := sp.get(st)
	pos := int(math.Round((v - sp.lo) / (sp.hi - sp.lo) * (barCells - 1)))
	bar := strings.Repeat("━", pos) + "●" + strings.Repeat("─", barCells-1-pos)
	value := fmt.Sprintf("%6s", fmt.Sprintf(sp.format, v))
	label := fmt.Sprintf("    %-10s", sp.label)
	if !sp.enabled(st.Background.Mode) {
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
	if active := map[string]func(paint.Mode) bool{"From": fromActive, "To": toActive}[label]; !active(st.Background.Mode) {
		return "  " + styles.dim.Render(label+"  "+c.Hex())
	}
	swatch := lipgloss.NewStyle().Background(colour(c)).Render("  ")
	return h + " " + swatch + " " + styles.dim.Render(c.Hex())
}

func bgModeLine(st *paint.Style) string {
	return "    Mode      ‹ " + styles.label.Render(st.Background.Mode.String()) + " ›"
}
