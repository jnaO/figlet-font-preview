package action

import "testing"

func TestAllIsWellFormed(t *testing.T) {
	seen := map[ID]bool{}
	closed := map[string]bool{}
	prev := ""
	for _, a := range All() {
		if a.ID == "" || a.Title == "" || a.Group == "" {
			t.Errorf("incomplete action %+v", a)
		}
		if seen[a.ID] {
			t.Errorf("duplicate id %q", a.ID)
		}
		seen[a.ID] = true
		if a.Group != prev {
			if closed[a.Group] {
				t.Errorf("group %q is not contiguous", a.Group)
			}
			closed[prev] = true
			prev = a.Group
		}
	}
	if len(seen) != 11 {
		t.Errorf("%d actions, want 11", len(seen))
	}
}

func TestFindRoundTrips(t *testing.T) {
	for _, a := range All() {
		got, ok := Find(a.ID)
		if !ok || got != a {
			t.Errorf("Find(%q) = %+v, %v", a.ID, got, ok)
		}
	}
	if _, ok := Find("nope"); ok {
		t.Error("Find of an unknown id succeeded")
	}
}

func TestAllReturnsAFreshSlice(t *testing.T) {
	a := All()
	a[0].Title = "changed"
	if All()[0].Title == "changed" {
		t.Error("All shares its backing array")
	}
}
