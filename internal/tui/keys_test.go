package tui

import (
	"testing"

	"github.com/jnaO/figlet-font-preview/internal/action"
)

func TestEveryActionHasOneBinding(t *testing.T) {
	for _, a := range action.All() {
		n := 0
		for _, tbl := range [][]keyEntry{globalKeys, zoneKeys} {
			for _, e := range tbl {
				if e.action == a.ID {
					n++
				}
			}
		}
		if n != 1 {
			t.Errorf("%s: %d bindings, want 1", a.ID, n)
			continue
		}
		e, ok := binding(a.ID)
		if !ok || e.show == "" || e.run == nil {
			t.Errorf("%s: binding ok=%v show=%q run nil=%v", a.ID, ok, e.show, e.run == nil)
		}
	}
}

func TestTableActionsAreKnown(t *testing.T) {
	for _, tbl := range [][]keyEntry{globalKeys, zoneKeys} {
		for _, e := range tbl {
			if e.action == "" {
				continue
			}
			if _, ok := action.Find(e.action); !ok {
				t.Errorf("unknown action %q", e.action)
			}
		}
	}
	for _, e := range zoneKeys {
		if len(e.keys) != 0 {
			t.Errorf("zone entry %q has keys %v", e.action, e.keys)
		}
	}
}
