// Package paint holds the colour model shared by every front-end and the SVG
// writer: the lolcat-style rainbow for glyph colours and the background
// (theme, solid, linear or radial gradient). It has no UI dependencies.
package paint

import (
	"fmt"
	"math"
)

// RGB is an sRGB colour.
type RGB struct{ R, G, B uint8 }

// Hex returns the colour as #rrggbb.
func (c RGB) Hex() string { return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B) }

// ParseHex parses #rrggbb (the leading # is optional).
func ParseHex(s string) (RGB, error) {
	var c RGB
	if len(s) > 0 && s[0] == '#' {
		s = s[1:]
	}
	if len(s) != 6 {
		return c, fmt.Errorf("paint: %q is not #rrggbb", s)
	}
	if _, err := fmt.Sscanf(s, "%02x%02x%02x", &c.R, &c.G, &c.B); err != nil {
		return c, fmt.Errorf("paint: %q is not #rrggbb", s)
	}
	return c, nil
}

// Lerp blends a toward b by t in [0, 1], per channel.
func Lerp(a, b RGB, t float64) RGB {
	t = clamp01(t)
	ch := func(x, y uint8) uint8 { return uint8(math.Round(float64(x) + (float64(y)-float64(x))*t)) }
	return RGB{ch(a.R, b.R), ch(a.G, b.G), ch(a.B, b.B)}
}

// Ranges of the tunable values.
const (
	SeedMax   = 255
	SpreadMin = 0.5
	SpreadMax = 20.0
	FreqMin   = 0.01
	FreqMax   = 0.50
	AngleMax  = 345 // degrees; angles wrap at 360
)

// Rainbow is lolcat's colouring: rune (row, col) gets the colour at
// Seed + row + col/Spread, a sine wave of frequency Freq per channel.
type Rainbow struct {
	Seed   int
	Spread float64
	Freq   float64
}

// DefaultRainbow is equivalent to `lolcat -S 20 -p 7 -F 0.05`.
var DefaultRainbow = Rainbow{Seed: 20, Spread: 7, Freq: 0.05}

// At returns the colour of the rune at row, col (col counts runes).
func (r Rainbow) At(row, col int) RGB {
	i := float64(r.Seed) + float64(row) + float64(col)/r.Spread
	c := func(phase float64) uint8 { return uint8(int(math.Sin(r.Freq*i+phase)*127 + 128)) }
	return RGB{c(0), c(2 * math.Pi / 3), c(4 * math.Pi / 3)}
}

// Mode is how the background is painted.
type Mode int

const (
	Theme  Mode = iota // the front-end's own theme colour
	Solid              // From everywhere
	Linear             // From to To along Angle
	Radial             // From at the centre to To at the edges of the box's ellipse
	NumModes
)

var modeNames = [NumModes]string{"theme", "solid", "linear", "radial"}

func (m Mode) String() string { return modeNames[m] }

// Background is the backdrop behind a rendered block.
type Background struct {
	Mode     Mode
	From, To RGB
	Angle    int // degrees, Linear only; 0 runs left to right, 90 top to bottom
}

// At returns the background of cell (x, y) in a w×h cell box, or false in
// Theme mode. Cells are sampled at their centres in box-normalised
// coordinates, matching SVG gradients in objectBoundingBox units (see
// LinearVector and the svgexport package).
func (b Background) At(x, y, w, h int) (RGB, bool) {
	switch b.Mode {
	case Theme:
		return RGB{}, false
	case Solid:
		return b.From, true
	}
	u := (float64(x) + .5) / float64(max(w, 1))
	v := (float64(y) + .5) / float64(max(h, 1))
	var t float64
	if b.Mode == Linear {
		s, c := math.Sincos(float64(b.Angle) * math.Pi / 180)
		t = (u-.5)*c + (v-.5)*s + .5
	} else {
		t = math.Hypot((u-.5)/.5, (v-.5)/.5)
	}
	return Lerp(b.From, b.To, t), true
}

// LinearVector returns the gradient's start and end points in the unit box:
// the t = 0 and t = 1 points of At's Linear formula.
func (b Background) LinearVector() (x1, y1, x2, y2 float64) {
	s, c := math.Sincos(float64(b.Angle) * math.Pi / 180)
	return .5 - c/2, .5 - s/2, .5 + c/2, .5 + s/2
}

// Style is the full colouring of a render.
type Style struct {
	Rainbow    Rainbow
	Background Background
}

// Default is today's look: the default rainbow on the theme background.
func Default() Style { return Style{Rainbow: DefaultRainbow} }

func clamp01(t float64) float64 { return math.Max(0, math.Min(1, t)) }
