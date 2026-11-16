package controls

import (
	"testing"

	"github.com/jnaO/figlet-font-preview/internal/display"
	"github.com/jnaO/figlet-font-preview/internal/fontclass"
	"github.com/jnaO/figlet-font-preview/internal/paint"
)

func profile(key byte, stroke, weight, size, slant int, fp float64) fontclass.Profile {
	return fontclass.Profile{
		Key:         [32]byte{key},
		Values:      [fontclass.NumAxes]int{stroke, weight, size, slant},
		Fingerprint: []float64{fp},
	}
}

func state(indexed bool) *display.State {
	s := display.New([]string{"a", "b", "c"}, "hi")
	if indexed {
		s.SetIndex(fontclass.New([]fontclass.Input{
			{Name: "a", Profile: profile(1, 0, 1, 2, 0, 0)},
			{Name: "b", Profile: profile(2, 1, 0, 2, 0, 5)},
			{Name: "c", Profile: profile(3, 0, 1, 1, 1, 2)},
		}))
	}
	return s
}

func TestApplyEffects(t *testing.T) {
	cases := []struct {
		name          string
		row           Row
		cursor        string
		before, after [2]Effect // index of dir: [0] dir=+1, [1] dir=-1
	}{
		{"align", Row{Kind: KindAlign}, "a", [2]Effect{PreviewChanged, PreviewChanged}, [2]Effect{PreviewChanged, PreviewChanged}},
		{"mode", Row{Kind: KindBgMode}, "a", [2]Effect{PreviewChanged, PreviewChanged}, [2]Effect{PreviewChanged, PreviewChanged}},
		{"slider enabled", Row{Kind: KindSlider, Slider: Seed}, "a", [2]Effect{PreviewChanged, PreviewChanged}, [2]Effect{PreviewChanged, PreviewChanged}},
		{"slider inert in theme", Row{Kind: KindSlider, Slider: FromR}, "a", [2]Effect{NoEffect, NoEffect}, [2]Effect{NoEffect, NoEffect}},
		{"collapse", Row{Kind: KindCollapse}, "a", [2]Effect{NoEffect, NoEffect}, [2]Effect{ItemsChanged, NoEffect}},
		{"group by", Row{Kind: KindGroupBy}, "a", [2]Effect{NoEffect, NoEffect}, [2]Effect{ItemsChanged, ItemsChanged}},
		{"similar with cursor", Row{Kind: KindSimilar}, "a", [2]Effect{NoEffect, NoEffect}, [2]Effect{ItemsChanged, NoEffect}},
		{"similar without cursor", Row{Kind: KindSimilar}, "", [2]Effect{NoEffect, NoEffect}, [2]Effect{NoEffect, NoEffect}},
		{"facet", Row{Kind: KindFacet, Axis: fontclass.Axis(0), Value: 0}, "a", [2]Effect{NoEffect, NoEffect}, [2]Effect{ItemsChanged, NoEffect}},
		{"clear", Row{Kind: KindClear}, "a", [2]Effect{NoEffect, NoEffect}, [2]Effect{ItemsChanged, NoEffect}},
		{"axis", Row{Kind: KindAxis}, "a", [2]Effect{NoEffect, NoEffect}, [2]Effect{NoEffect, NoEffect}},
		{"heading", Row{Kind: KindHeading}, "a", [2]Effect{NoEffect, NoEffect}, [2]Effect{NoEffect, NoEffect}},
	}
	for _, c := range cases {
		for i, dir := range []int{1, -1} {
			if got := Apply(state(false), c.row, dir, c.cursor); got != c.before[i] {
				t.Errorf("%s before index, dir %d: %v, want %v", c.name, dir, got, c.before[i])
			}
			if got := Apply(state(true), c.row, dir, c.cursor); got != c.after[i] {
				t.Errorf("%s after index, dir %d: %v, want %v", c.name, dir, got, c.after[i])
			}
		}
	}
}

func TestApplyState(t *testing.T) {
	s := state(true)
	Apply(s, Row{Kind: KindGroupBy}, 1, "")
	if s.GroupBy != fontclass.Axis(0) {
		t.Fatalf("group by after +1 = %d, want first axis", s.GroupBy)
	}
	Apply(s, Row{Kind: KindGroupBy}, -1, "")
	Apply(s, Row{Kind: KindGroupBy}, -1, "")
	if s.GroupBy != fontclass.NumAxes-1 {
		t.Fatalf("group by wraps backwards to the last axis, got %d", s.GroupBy)
	}
	Apply(s, Row{Kind: KindSimilar}, 1, "b")
	if s.Similar() != "b" {
		t.Fatalf("similar anchors on the cursor font, got %q", s.Similar())
	}
	Apply(s, Row{Kind: KindSimilar}, 1, "c")
	if s.Similar() != "" {
		t.Fatalf("similar toggles off, got %q", s.Similar())
	}
	Apply(s, Row{Kind: KindBgMode}, 1, "")
	if s.Paint.Background.Mode != paint.Mode(1) {
		t.Fatalf("mode = %v", s.Paint.Background.Mode)
	}
	Apply(s, Row{Kind: KindAlign}, 1, "")
	if s.Align != display.AlignCenter {
		t.Fatalf("align = %v", s.Align)
	}
}

func TestSelectable(t *testing.T) {
	for k, want := range map[Kind]bool{KindAxis: false, KindHeading: false, KindAlign: true, KindFacet: true, KindSlider: true} {
		if got := (Row{Kind: k}).Selectable(); got != want {
			t.Errorf("kind %d selectable = %v, want %v", k, got, want)
		}
	}
}

func TestRowLists(t *testing.T) {
	f := FilterRows()
	if f[0].Kind != KindCollapse || f[1].Kind != KindGroupBy || f[2].Kind != KindSimilar || f[len(f)-1].Kind != KindClear {
		t.Fatalf("filter rows out of order: %+v", f)
	}
	s := StyleRows()
	if s[0].Kind != KindAlign || s[1].Kind != KindHeading || s[1].Label != "Rainbow" || s[len(s)-1].Slider != Angle {
		t.Fatalf("style rows out of order: %+v", s)
	}
}
