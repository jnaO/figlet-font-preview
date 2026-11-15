// Package svgexport writes a rendered block as an SVG document coloured with
// a paint.Style. It has no UI dependencies.
package svgexport

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"unicode/utf8"

	"github.com/jnaO/figlet-font-preview/internal/paint"
)

// Text geometry: a monospace cell is 0.6em wide and 1.2em tall.
const (
	fontSize = 16.0
	cellW    = 0.6 * fontSize
	cellH    = 1.2 * fontSize
	baseline = 0.9 * fontSize // from the cell top
	margin   = 1              // cells of padding on every side
)

const fontFamily = "ui-monospace, SFMono-Regular, Menlo, Consolas, monospace"

// Render writes lines as SVG. top and left are the rows and columns trimmed
// off the block before it reached here: rune (r, c) gets st.Rainbow.At(top+r,
// left+c), the colour the preview gives that cell. Each rune is placed at its
// own x so the grid holds whatever monospace font the viewer substitutes. The
// background follows st.Background over the whole canvas; Theme leaves it
// transparent.
func Render(lines []string, top, left int, st paint.Style) []byte {
	cols := 0
	for _, l := range lines {
		cols = max(cols, utf8.RuneCountInString(l))
	}
	w := float64(cols+2*margin) * cellW
	h := float64(len(lines)+2*margin) * cellH

	var b bytes.Buffer
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%s" height="%s" viewBox="0 0 %s %s">`+"\n", num(w), num(h), num(w), num(h))
	writeBackground(&b, st.Background)
	fmt.Fprintf(&b, `<g font-family="%s" font-size="%s">`+"\n", fontFamily, num(fontSize))
	for r, line := range lines {
		y := float64(r+margin)*cellH + baseline
		fmt.Fprintf(&b, `<text y="%s">`, num(y))
		c := 0
		for _, ch := range line {
			if ch != ' ' {
				x := float64(c+margin) * cellW
				fmt.Fprintf(&b, `<tspan x="%s" fill="%s">%s</tspan>`, num(x), st.Rainbow.At(top+r, left+c).Hex(), escape(ch))
			}
			c++
		}
		b.WriteString("</text>\n")
	}
	b.WriteString("</g>\n</svg>\n")
	return b.Bytes()
}

// writeBackground paints the canvas. Gradients use the default
// objectBoundingBox units, pad spread and sRGB interpolation, which is the
// geometry paint.Background.At samples.
func writeBackground(b *bytes.Buffer, bg paint.Background) {
	fill := bg.From.Hex()
	switch bg.Mode {
	case paint.Theme:
		return
	case paint.Linear:
		x1, y1, x2, y2 := bg.LinearVector()
		fmt.Fprintf(b, `<defs><linearGradient id="bg" x1="%s" y1="%s" x2="%s" y2="%s">%s</linearGradient></defs>`+"\n",
			num(x1), num(y1), num(x2), num(y2), stops(bg))
		fill = "url(#bg)"
	case paint.Radial:
		fmt.Fprintf(b, `<defs><radialGradient id="bg" cx="0.5" cy="0.5" r="0.5">%s</radialGradient></defs>`+"\n", stops(bg))
		fill = "url(#bg)"
	}
	fmt.Fprintf(b, `<rect width="100%%" height="100%%" fill="%s"/>`+"\n", fill)
}

func stops(bg paint.Background) string {
	return fmt.Sprintf(`<stop offset="0" stop-color="%s"/><stop offset="1" stop-color="%s"/>`, bg.From.Hex(), bg.To.Hex())
}

func escape(r rune) string {
	switch r {
	case '&':
		return "&amp;"
	case '<':
		return "&lt;"
	case '>':
		return "&gt;"
	}
	return string(r)
}

// num formats v with at most four decimals and no trailing zeros.
func num(v float64) string {
	return strconv.FormatFloat(math.Round(v*1e4)/1e4, 'f', -1, 64)
}
