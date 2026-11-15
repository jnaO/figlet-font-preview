package svgexport

import (
	"bytes"
	"encoding/xml"
	"io"
	"strings"
	"testing"

	"github.com/jnaO/figlet-font-preview/internal/paint"
)

func wellFormed(t *testing.T, doc []byte) {
	t.Helper()
	d := xml.NewDecoder(bytes.NewReader(doc))
	for {
		if _, err := d.Token(); err == io.EOF {
			return
		} else if err != nil {
			t.Fatalf("not well-formed XML: %v\n%s", err, doc)
		}
	}
}

func TestRenderColoursMatchPreviewCells(t *testing.T) {
	st := paint.Default()
	doc := Render([]string{"a<b", " &"}, 2, 3, st)
	wellFormed(t, doc)
	s := string(doc)
	// Rune (1,1) '&' sits at preview cell (2+1, 3+1).
	want := `fill="` + st.Rainbow.At(3, 4).Hex() + `">&amp;</tspan>`
	if !strings.Contains(s, want) {
		t.Fatalf("missing %s in\n%s", want, s)
	}
	if !strings.Contains(s, "&lt;") || strings.Count(s, "<tspan") != 4 {
		t.Fatalf("expected 4 escaped tspans (spaces skipped):\n%s", s)
	}
	if strings.Contains(s, "<rect") {
		t.Fatal("theme mode must leave the background transparent")
	}
	// 3 columns + 2 margin cells at 9.6px, 2 rows + 2 at 19.2px.
	if !strings.Contains(s, `width="48" height="76.8"`) {
		t.Fatalf("unexpected canvas size:\n%s", s)
	}
}

func TestRenderBackgrounds(t *testing.T) {
	bg := paint.Background{From: paint.RGB{R: 1, G: 2, B: 3}, To: paint.RGB{R: 250, G: 251, B: 252}}
	cases := []struct {
		mode  paint.Mode
		angle int
		want  []string
	}{
		{paint.Solid, 0, []string{`<rect width="100%" height="100%" fill="#010203"/>`}},
		{paint.Linear, 0, []string{`<linearGradient id="bg" x1="0" y1="0.5" x2="1" y2="0.5">`, `stop-color="#fafbfc"`, `fill="url(#bg)"`}},
		{paint.Linear, 90, []string{`x1="0.5" y1="0" x2="0.5" y2="1"`}},
		{paint.Radial, 0, []string{`<radialGradient id="bg" cx="0.5" cy="0.5" r="0.5">`, `fill="url(#bg)"`}},
	}
	for _, c := range cases {
		st := paint.Default()
		st.Background = bg
		st.Background.Mode, st.Background.Angle = c.mode, c.angle
		doc := Render([]string{"x"}, 0, 0, st)
		wellFormed(t, doc)
		for _, w := range c.want {
			if !strings.Contains(string(doc), w) {
				t.Errorf("%v %d°: missing %s\n%s", c.mode, c.angle, w, doc)
			}
		}
	}
}
