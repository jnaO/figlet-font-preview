package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/jnaO/figlet-font-preview/internal/catalog"
	"github.com/jnaO/figlet-font-preview/internal/figfont"
)

func embedded(t *testing.T, name string) []byte {
	t.Helper()
	data, err := bannerFonts.ReadFile("fonts/" + name + ".flf")
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestEmbeddedFontsParse(t *testing.T) {
	for name, height := range map[string]int{"banner": 8, "small": 5, "mini": 4} {
		f, err := figfont.Parse(embedded(t, name))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if f.Height != height {
			t.Errorf("%s height = %d, want %d", name, f.Height, height)
		}
	}
}

func TestWidthTableMatchesFonts(t *testing.T) {
	prev := 1 << 30
	for _, e := range bannerTable {
		f, err := figfont.Parse(embedded(t, e.name))
		if err != nil {
			t.Fatal(err)
		}
		if got := widestWord(f); got != e.min {
			t.Errorf("%s: widestWord = %d, table min = %d", e.name, got, e.min)
		}
		if e.min >= prev {
			t.Errorf("%s: min %d does not decrease", e.name, e.min)
		}
		prev = e.min
	}
}

func emptyCatalog(t *testing.T) *catalog.Catalog {
	t.Helper()
	return catalog.Discover([]string{t.TempDir()})
}

func nameAt(t *testing.T, cat *catalog.Catalog, width int) string {
	t.Helper()
	f, name := bannerFont(cat, width)
	if (f == nil) != (name == "") {
		t.Fatalf("width %d: font nil=%v but name %q", width, f == nil, name)
	}
	return name
}

func TestBannerFontEmptyRoot(t *testing.T) {
	cat := emptyCatalog(t)
	for width, want := range map[int]string{200: "banner", 52: "banner", 51: "small", 36: "small", 35: "mini", 23: "mini", 22: ""} {
		if got := nameAt(t, cat, width); got != want {
			t.Errorf("width %d = %q, want %q", width, got, want)
		}
	}
}

func writeFont(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestBannerFontAMCFirst(t *testing.T) {
	root := t.TempDir()
	writeFont(t, filepath.Join(root, "AMC AAA01.flf"), embedded(t, "mini"))
	cat := catalog.Discover([]string{root})
	if got := nameAt(t, cat, 30); got != "AMC AAA01" {
		t.Errorf("width 30 = %q, want AMC AAA01", got)
	}
	if got := nameAt(t, cat, 22); got != "" {
		t.Errorf("width 22 = %q, want none", got)
	}
}

func TestBannerFontAMCTooWide(t *testing.T) {
	root := t.TempDir()
	writeFont(t, filepath.Join(root, "AMC AAA01.flf"), embedded(t, "banner"))
	cat := catalog.Discover([]string{root})
	if got := nameAt(t, cat, 51); got != "small" {
		t.Errorf("width 51 = %q, want small", got)
	}
}

func TestBannerFontAMCCaseInsensitive(t *testing.T) {
	root := t.TempDir()
	writeFont(t, filepath.Join(root, "amc aaa01.flf"), embedded(t, "mini"))
	if got := nameAt(t, catalog.Discover([]string{root}), 30); got != "amc aaa01" {
		t.Errorf("got %q", got)
	}
}

func TestBannerFontAMCQualifiedName(t *testing.T) {
	root := t.TempDir()
	writeFont(t, filepath.Join(root, "a", "AMC AAA01.flf"), embedded(t, "mini"))
	writeFont(t, filepath.Join(root, "b", "AMC AAA01.flf"), embedded(t, "mini"))
	got := nameAt(t, catalog.Discover([]string{root}), 30)
	if filepath.Base(got) != "AMC AAA01" || got == "AMC AAA01" {
		t.Errorf("got %q, want a qualified AMC AAA01 name", got)
	}
}

func TestVersionLinesPlain(t *testing.T) {
	got := versionLines(nil, "figlet-font-preview 1", 80)
	if len(got) != 1 || got[0] != "figlet-font-preview 1" {
		t.Errorf("got %q", got)
	}
}

func TestVersionLinesLayout(t *testing.T) {
	const line = "figlet-font-preview 1.2.3"
	cat := emptyCatalog(t)
	for width, blocks := range map[int]int{200: 1, 80: 2, 52: 3} {
		f, _ := bannerFont(cat, width)
		got := versionLines(f, line, width)
		for i, row := range got {
			if n := utf8.RuneCountInString(row); n > width {
				t.Errorf("width %d row %d is %d wide", width, i, n)
			}
			if row != strings.TrimRight(row, " ") {
				t.Errorf("width %d row %d has trailing spaces", width, i)
			}
		}
		last := got[len(got)-1]
		if wantLast := strings.Repeat(" ", (width-len(line))/2) + line; last != wantLast {
			t.Errorf("width %d last row = %q, want %q", width, last, wantLast)
		}
		if got[len(got)-2] != "" {
			t.Errorf("width %d: no empty row before the version line", width)
		}
		runs, in := 0, false
		for _, row := range got[:len(got)-2] {
			if row != "" && !in {
				runs++
			}
			in = row != ""
		}
		if runs != blocks {
			t.Errorf("width %d: %d blocks, want %d", width, runs, blocks)
		}
	}
}
