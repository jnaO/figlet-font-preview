package controls

import (
	"testing"

	"github.com/jnaO/figlet-font-preview/internal/paint"
)

func TestSlidersWrapAndClamp(t *testing.T) {
	st := paint.Default()
	st.Rainbow.Seed = 254
	StepSlider(&st, Seed, 1)
	if st.Rainbow.Seed != 2 {
		t.Fatalf("seed should wrap modulo 256: %d", st.Rainbow.Seed)
	}
	st.Background.Angle = 0
	StepSlider(&st, Angle, -1)
	if st.Background.Angle != 345 {
		t.Fatalf("angle should wrap to 345: %d", st.Background.Angle)
	}
	st.Rainbow.Freq = paint.FreqMax
	StepSlider(&st, Freq, 1)
	if st.Rainbow.Freq != paint.FreqMax {
		t.Fatalf("frequency should clamp: %v", st.Rainbow.Freq)
	}
	st.Rainbow.Freq = 0.05
	StepSlider(&st, Freq, 1)
	if st.Rainbow.Freq != 0.06 {
		t.Fatalf("frequency step drifted: %v", st.Rainbow.Freq)
	}
	st.Background.From.R = 250
	StepSlider(&st, FromR, 1)
	if st.Background.From.R != 255 {
		t.Fatalf("RGB should clamp at 255: %d", st.Background.From.R)
	}
}

func TestSpecEnabledByMode(t *testing.T) {
	cases := []struct {
		k    Slider
		mode paint.Mode
		want bool
	}{
		{Seed, paint.Theme, true},
		{FromR, paint.Theme, false},
		{FromR, paint.Solid, true},
		{ToG, paint.Solid, false},
		{ToG, paint.Radial, true},
		{Angle, paint.Radial, false},
		{Angle, paint.Linear, true},
	}
	for _, c := range cases {
		if got := c.k.Spec().Enabled(c.mode); got != c.want {
			t.Errorf("slider %d in %v: enabled = %v, want %v", c.k, c.mode, got, c.want)
		}
	}
}
