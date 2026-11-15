package paint

import "testing"

func TestHexRoundTrip(t *testing.T) {
	c := RGB{0x32, 0x38, 0x46}
	if c.Hex() != "#323846" {
		t.Fatalf("Hex = %s", c.Hex())
	}
	got, err := ParseHex("#323846")
	if err != nil || got != c {
		t.Fatalf("ParseHex = %v, %v", got, err)
	}
	if _, err := ParseHex("#12345"); err == nil {
		t.Fatal("short hex accepted")
	}
}

func TestLerpClamps(t *testing.T) {
	a, b := RGB{0, 0, 0}, RGB{255, 100, 10}
	if Lerp(a, b, -1) != a || Lerp(a, b, 2) != b {
		t.Fatal("Lerp does not clamp t")
	}
	if got := Lerp(a, b, .5); got != (RGB{128, 50, 5}) {
		t.Fatalf("Lerp(.5) = %v", got)
	}
}

// sin(0.05·20) = 0.841 → R 234; the other channels are phase-shifted by 2π/3.
func TestDefaultRainbow(t *testing.T) {
	r := DefaultRainbow
	if got := r.At(0, 0).Hex(); got != "#ea850f" {
		t.Fatalf("At(0,0) = %s", got)
	}
	if r.At(0, 0) == r.At(1, 0) {
		t.Fatal("adjacent rows share a colour")
	}
	if r.At(0, 7) != r.At(1, 0) {
		t.Fatal("Spread runes along a row should equal one row down")
	}
}

func TestBackgroundModes(t *testing.T) {
	from, to := RGB{0, 0, 0}, RGB{200, 200, 200}
	if _, ok := (Background{Mode: Theme}).At(0, 0, 10, 4); ok {
		t.Fatal("Theme should defer to the front-end")
	}
	if c, _ := (Background{Mode: Solid, From: from, To: to}).At(9, 3, 10, 4); c != from {
		t.Fatalf("Solid = %v", c)
	}

	lin := Background{Mode: Linear, From: from, To: to}
	l, _ := lin.At(0, 0, 10, 4)
	r, _ := lin.At(9, 0, 10, 4)
	d, _ := lin.At(0, 3, 10, 4)
	if !(l.R < r.R) || d != l {
		t.Fatalf("angle 0 should run left to right only: %v %v %v", l, r, d)
	}
	lin.Angle = 90
	top, _ := lin.At(0, 0, 10, 4)
	bottom, _ := lin.At(0, 3, 10, 4)
	if !(top.R < bottom.R) {
		t.Fatalf("angle 90 should run top to bottom: %v %v", top, bottom)
	}

	rad := Background{Mode: Radial, From: from, To: to}
	centre, _ := rad.At(5, 2, 11, 5)
	corner, _ := rad.At(0, 0, 11, 5)
	if centre != from || corner != to {
		t.Fatalf("radial centre %v corner %v", centre, corner)
	}
}

// LinearVector's endpoints are where At's Linear formula gives t = 0 and 1.
func TestLinearVectorMatchesAt(t *testing.T) {
	for _, a := range []int{0, 45, 90, 135, 270} {
		b := Background{Angle: a}
		x1, y1, x2, y2 := b.LinearVector()
		dx, dy := x2-x1, y2-y1
		if l := dx*dx + dy*dy; l < .999 || l > 1.001 {
			t.Fatalf("angle %d: vector length² %v", a, l)
		}
		t0 := (x1-.5)*dx + (y1-.5)*dy + .5
		if t0 > 1e-9 || t0 < -1e-9 {
			t.Fatalf("angle %d: start point t = %v", a, t0)
		}
	}
}
