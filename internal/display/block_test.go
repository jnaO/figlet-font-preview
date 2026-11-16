package display

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/jnaO/figlet-font-preview/internal/catalog"
)

func tinyCatalog(t *testing.T) *catalog.Catalog {
	t.Helper()
	font, err := os.ReadFile("../figfont/testdata/tiny.flf")
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "tiny.flf"), font, 0o644); err != nil {
		t.Fatal(err)
	}
	return catalog.Discover([]string{root})
}

func TestBlockErrors(t *testing.T) {
	cat := tinyCatalog(t)
	s := New(cat.Names(), "a")
	for _, name := range []string{"", "missing"} {
		if _, err := s.Block(cat, name, 40); err == nil || err.Error() != "no font under the cursor" {
			t.Errorf("name %q: err = %v", name, err)
		}
	}
	if _, err := s.Block(cat, "tiny", 40); err == nil || err.Error() != "tiny is not loaded" {
		t.Errorf("unloaded: err = %v", err)
	}
}

func TestBlockTrimsAndReportsOffsets(t *testing.T) {
	cat := tinyCatalog(t)
	<-cat.Preload(1)
	s := New(cat.Names(), "a")
	e, _ := cat.Get("tiny")
	f, _ := e.Font()

	left, err := s.Block(cat, "tiny", 40)
	if err != nil || left.Name != "tiny" {
		t.Fatalf("left: %+v, %v", left, err)
	}
	s.Align = AlignCenter
	centre, err := s.Block(cat, "tiny", 40)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(centre.Lines, left.Lines) || centre.Left <= left.Left {
		t.Fatalf("centre %+v should be left %+v shifted right", centre, left)
	}
	if got := s.Lines(f, 40); len(got) != f.Height {
		t.Fatalf("Lines = %d rows, want %d", len(got), f.Height)
	}
}
