package display

import (
	"reflect"
	"testing"
)

var fonts = []string{"banner", "big", "slant", "smslant", "standard"}

func TestRenderSet(t *testing.T) {
	cases := []struct {
		name     string
		query    string
		selected []string
		want     []string
	}{
		{"all", "", nil, fonts},
		{"selected only", "", []string{"standard", "big"}, []string{"big", "standard"}},
		{"search only", "slnt", nil, []string{"slant", "smslant"}},
		{"search plus selection", "slnt", []string{"banner"}, []string{"banner", "slant", "smslant"}},
		{"selected match listed once", "slnt", []string{"slant"}, []string{"slant", "smslant"}},
		{"no matches keeps selection", "zzz", []string{"big"}, []string{"big"}},
		{"no matches no selection", "zzz", nil, []string{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := New(fonts, "hi")
			s.SetQuery(c.query)
			for _, n := range c.selected {
				s.Toggle(n)
			}
			if got := s.RenderSet(); !reflect.DeepEqual(got, c.want) {
				t.Errorf("got %v, want %v", got, c.want)
			}
		})
	}
}

func TestSidebar(t *testing.T) {
	s := New(fonts, "")
	if got := len(s.Sidebar()); got != len(fonts) {
		t.Fatalf("got %d items", got)
	}
	s.SetQuery("slnt")
	items := s.Sidebar()
	if len(items) != 2 || items[0].Name != "slant" || len(items[0].Matched) != 4 {
		t.Fatalf("got %+v", items)
	}
}

func TestSelectionSurvivesSearch(t *testing.T) {
	s := New(fonts, "")
	s.Toggle("big")
	s.SetQuery("sl")
	s.Toggle("slant")
	s.SetQuery("")
	if got := s.RenderSet(); !reflect.DeepEqual(got, []string{"big", "slant"}) {
		t.Fatalf("got %v", got)
	}
	s.Toggle("big")
	s.ClearSelection()
	if s.SelectedCount() != 0 {
		t.Fatal("selection not cleared")
	}
}

func TestAlignApply(t *testing.T) {
	lines := []string{"ab", "abcd", "x", "x"}
	cases := []struct {
		a    Align
		want []string
	}{
		{AlignLeft, lines},
		{AlignCenter, []string{"   ab", "   abcd", "    x", "    x"}},
		{AlignRight, []string{"      ab", "      abcd", "         x", "         x"}},
	}
	for _, c := range cases {
		if got := c.a.Apply(lines, 2, 10); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%v: got %q, want %q", c.a, got, c.want)
		}
	}
	if New(fonts, "").Align != AlignLeft {
		t.Error("default should be left")
	}
}

func TestAlignString(t *testing.T) {
	for a, want := range map[Align]string{AlignLeft: "left", AlignCenter: "center", AlignRight: "right"} {
		if a.String() != want {
			t.Errorf("%d: got %q, want %q", a, a.String(), want)
		}
	}
}
