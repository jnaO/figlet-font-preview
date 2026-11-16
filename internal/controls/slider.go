// Package controls holds the option rows of the Filters and Style panes and
// the logic that applies them, so every front-end offers the same controls.
package controls

import (
	"math"

	"github.com/jnaO/figlet-font-preview/internal/paint"
)

// Slider is one numeric colour setting in the Style pane.
type Slider int

const (
	Seed Slider = iota
	Spread
	Freq
	FromR
	FromG
	FromB
	ToR
	ToG
	ToB
	Angle
)

// Spec describes a Slider: its range, step, display format and the
// background modes in which it does anything.
type Spec struct {
	Label        string
	Lo, Hi, Step float64
	Period       float64 // > 0: values wrap modulo Period instead of clamping
	Format       string  // value format
	field        func(st *paint.Style) any
	active       func(paint.Mode) bool // nil: always active
}

// FromActive reports whether the From colour is in use in background mode m.
func FromActive(m paint.Mode) bool { return m != paint.Theme }

// ToActive reports whether the To colour is in use in background mode m.
func ToActive(m paint.Mode) bool { return m == paint.Linear || m == paint.Radial }

var specs = map[Slider]Spec{
	Seed:   {Label: "Seed", Hi: paint.SeedMax, Step: 4, Period: paint.SeedMax + 1, Format: "%.0f", field: func(st *paint.Style) any { return &st.Rainbow.Seed }},
	Spread: {Label: "Spread", Lo: paint.SpreadMin, Hi: paint.SpreadMax, Step: .5, Format: "%.1f", field: func(st *paint.Style) any { return &st.Rainbow.Spread }},
	Freq:   {Label: "Frequency", Lo: paint.FreqMin, Hi: paint.FreqMax, Step: .01, Format: "%.2f", field: func(st *paint.Style) any { return &st.Rainbow.Freq }},
	FromR:  rgbSpec("R", FromActive, func(st *paint.Style) *uint8 { return &st.Background.From.R }),
	FromG:  rgbSpec("G", FromActive, func(st *paint.Style) *uint8 { return &st.Background.From.G }),
	FromB:  rgbSpec("B", FromActive, func(st *paint.Style) *uint8 { return &st.Background.From.B }),
	ToR:    rgbSpec("R", ToActive, func(st *paint.Style) *uint8 { return &st.Background.To.R }),
	ToG:    rgbSpec("G", ToActive, func(st *paint.Style) *uint8 { return &st.Background.To.G }),
	ToB:    rgbSpec("B", ToActive, func(st *paint.Style) *uint8 { return &st.Background.To.B }),
	Angle: {Label: "Angle", Hi: paint.AngleMax, Step: 15, Period: 360, Format: "%.0f°", field: func(st *paint.Style) any { return &st.Background.Angle },
		active: func(m paint.Mode) bool { return m == paint.Linear }},
}

func rgbSpec(label string, active func(paint.Mode) bool, f func(*paint.Style) *uint8) Spec {
	return Spec{Label: label, Hi: 255, Step: 8, Format: "%.0f", field: func(st *paint.Style) any { return f(st) }, active: active}
}

// Spec returns the slider's spec.
func (k Slider) Spec() Spec { return specs[k] }

// Enabled reports whether the slider does anything in background mode m.
func (sp Spec) Enabled(m paint.Mode) bool { return sp.active == nil || sp.active(m) }

// Get reads the slider's current value from st.
func (sp Spec) Get(st *paint.Style) float64 {
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

func (sp Spec) set(st *paint.Style, v float64) {
	switch p := sp.field(st).(type) {
	case *int:
		*p = int(math.Round(v))
	case *uint8:
		*p = uint8(math.Round(v))
	case *float64:
		*p = math.Round(v*100) / 100
	}
}

// StepSlider moves k by dir steps, wrapping or clamping at the ends.
func StepSlider(st *paint.Style, k Slider, dir int) {
	sp := k.Spec()
	v := sp.Get(st) + float64(dir)*sp.Step
	if sp.Period > 0 {
		v = math.Mod(math.Mod(v, sp.Period)+sp.Period, sp.Period)
	} else {
		v = math.Max(sp.Lo, math.Min(sp.Hi, v))
	}
	sp.set(st, v)
}
