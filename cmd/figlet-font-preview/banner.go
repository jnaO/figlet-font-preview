package main

import (
	"embed"
	"path/filepath"
	"strings"

	"github.com/jnaO/figlet-font-preview/internal/catalog"
	"github.com/jnaO/figlet-font-preview/internal/figfont"
)

// bannerFonts holds the FIGlet 2.2.5 banner, small and mini fonts, unmodified.
//
//go:embed fonts/*.flf
var bannerFonts embed.FS

const (
	bannerText = "FIGLET FONT PREVIEW"
	amcName    = "AMC AAA01"
)

// bannerTable lists the embedded fonts widest first; min is the font's
// widestWord, the narrowest terminal it fits.
var bannerTable = []struct {
	name string
	min  int
}{
	{"banner", 52},
	{"small", 36},
	{"mini", 23},
}

// widestWord is the widest row of any word of bannerText rendered in f.
func widestWord(f *figfont.Font) int {
	widest := 0
	for _, w := range strings.Fields(bannerText) {
		for _, row := range f.Render(w, 0) {
			widest = max(widest, len([]rune(row)))
		}
	}
	return widest
}

// bannerFont returns the font to render the banner in at width columns and
// its name: AMC AAA01 from cat when it loads and fits, else the widest
// embedded font whose min fits. It returns nil, "" when nothing fits.
func bannerFont(cat *catalog.Catalog, width int) (*figfont.Font, string) {
	for _, e := range cat.Entries {
		if !strings.EqualFold(filepath.Base(e.Name), amcName) {
			continue
		}
		if f, err := e.Load(); err == nil && widestWord(f) <= width {
			return f, e.Name
		}
		break
	}
	for _, t := range bannerTable {
		if t.min > width {
			continue
		}
		data, err := bannerFonts.ReadFile("fonts/" + t.name + ".flf")
		if err != nil {
			return nil, ""
		}
		f, err := figfont.Parse(data)
		if err != nil {
			return nil, ""
		}
		return f, t.name
	}
	return nil, ""
}

// versionLines returns the --version output rows: line alone when f is nil,
// else the centred banner, an empty row and the centred line.
func versionLines(f *figfont.Font, line string, width int) []string {
	if f == nil {
		return []string{line}
	}
	rows := figfont.Center(f.Render(bannerText, width), f.Height, width)
	for i, r := range rows {
		rows[i] = strings.TrimRight(r, " ")
	}
	for len(rows) > 0 && rows[len(rows)-1] == "" {
		rows = rows[:len(rows)-1]
	}
	indent := strings.Repeat(" ", max((width-len(line))/2, 0))
	return append(rows, "", indent+line)
}
