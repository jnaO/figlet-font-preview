package controls

import "github.com/jnaO/figlet-font-preview/internal/fontclass"

// Kind says what a Row is and what applying it does.
type Kind int

const (
	KindAlign Kind = iota
	KindCollapse
	KindGroupBy
	KindSimilar
	KindAxis    // non-selectable axis label
	KindHeading // non-selectable free-text label
	KindSlider
	KindBgMode
	KindFacet
	KindClear
)

// Row is one row of the Fonts pane filters or the Style pane.
type Row struct {
	Kind   Kind
	Axis   fontclass.Axis
	Value  int
	Label  string // KindHeading
	Slider Slider // KindSlider
}

// Selectable reports whether the row can hold the cursor.
func (r Row) Selectable() bool { return r.Kind != KindAxis && r.Kind != KindHeading }

// FilterRows lists the Fonts pane's filter rows in display order.
func FilterRows() []Row {
	rows := []Row{{Kind: KindCollapse}, {Kind: KindGroupBy}, {Kind: KindSimilar}}
	for _, a := range fontclass.Axes() {
		rows = append(rows, Row{Kind: KindAxis, Axis: a.Axis})
		for v := range a.Values {
			rows = append(rows, Row{Kind: KindFacet, Axis: a.Axis, Value: v})
		}
	}
	return append(rows, Row{Kind: KindClear})
}

// StyleRows lists the Style pane's rows in display order.
func StyleRows() []Row {
	return append([]Row{{Kind: KindAlign}}, paintRows()...)
}

// paintRows lists the colour rows of the Style pane.
func paintRows() []Row {
	return []Row{
		{Kind: KindHeading, Label: "Rainbow"},
		{Kind: KindSlider, Slider: Seed},
		{Kind: KindSlider, Slider: Spread},
		{Kind: KindSlider, Slider: Freq},
		{Kind: KindHeading, Label: "Background"},
		{Kind: KindBgMode},
		{Kind: KindHeading, Label: "From"},
		{Kind: KindSlider, Slider: FromR},
		{Kind: KindSlider, Slider: FromG},
		{Kind: KindSlider, Slider: FromB},
		{Kind: KindHeading, Label: "To"},
		{Kind: KindSlider, Slider: ToR},
		{Kind: KindSlider, Slider: ToG},
		{Kind: KindSlider, Slider: ToB},
		{Kind: KindSlider, Slider: Angle},
	}
}
