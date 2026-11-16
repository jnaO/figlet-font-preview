package controls

import (
	"github.com/jnaO/figlet-font-preview/internal/display"
	"github.com/jnaO/figlet-font-preview/internal/fontclass"
	"github.com/jnaO/figlet-font-preview/internal/paint"
)

// Effect tells the front-end what Apply changed.
type Effect int

const (
	NoEffect       Effect = iota // nothing to redraw
	PreviewChanged               // the preview re-renders; the font list is as before
	ItemsChanged                 // the universe changed: rebuild the sidebar
)

// Apply toggles or steps row r on s; dir only matters for the cycling rows.
// cursor is the font under the cursor, the anchor Similar falls back to.
// Filter rows do nothing until the font index exists.
func Apply(s *display.State, r Row, dir int, cursor string) Effect {
	ready := s.Index() != nil
	switch r.Kind {
	case KindAlign:
		n := display.NumAligns
		s.Align = display.Align((int(s.Align) + dir + n) % n)
		return PreviewChanged
	case KindSlider:
		if r.Slider.Spec().Enabled(s.Paint.Background.Mode) {
			StepSlider(&s.Paint, r.Slider, dir)
			return PreviewChanged
		}
		return NoEffect
	case KindBgMode:
		n := int(paint.NumModes)
		s.Paint.Background.Mode = paint.Mode((int(s.Paint.Background.Mode) + dir + n) % n)
		return PreviewChanged
	case KindCollapse:
		if !ready || dir < 0 {
			return NoEffect
		}
		s.Collapse = !s.Collapse
	case KindGroupBy:
		if !ready {
			return NoEffect
		}
		n := int(fontclass.NumAxes) + 1 // NoGroup plus each axis
		s.GroupBy = fontclass.Axis((int(s.GroupBy)+1+dir+n)%n - 1)
	case KindSimilar:
		if !ready || dir < 0 {
			return NoEffect
		}
		switch {
		case s.Similar() != "":
			s.SetSimilar("")
		case cursor != "":
			s.SetSimilar(cursor)
		default:
			return NoEffect
		}
	case KindFacet:
		if !ready || dir < 0 {
			return NoEffect
		}
		s.ToggleFilter(r.Axis, r.Value)
	case KindClear:
		if !ready || dir < 0 {
			return NoEffect
		}
		s.ClearFilters()
	default:
		return NoEffect
	}
	return ItemsChanged
}
