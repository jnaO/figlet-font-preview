package tui

import "strings"

// fontSources are shown on the no-fonts screen and mirrored in README.md.
var fontSources = []string{
	"https://github.com/cmatsuoka/figlet-fonts",
	"https://github.com/cacalabs/toilet",
}

// noFontsMessage is the preview text for an empty catalog: the searched
// roots, an install hint for goos, font sources and a disclaimer.
func noFontsMessage(roots []string, goos string) string {
	var b strings.Builder
	b.WriteString("No fonts found.\n\n")
	if len(roots) > 0 {
		b.WriteString("Searched:\n")
		for _, r := range roots {
			b.WriteString("  " + r + "\n")
		}
		b.WriteString("\n")
	}
	switch goos {
	case "darwin":
		b.WriteString("Install fonts with:\n  brew install figlet toilet\n")
	case "windows":
		b.WriteString("Put .flf/.tlf files in:\n  %USERPROFILE%\\figlet-fonts\n")
	default:
		b.WriteString("Install the figlet and toilet\npackages with your package manager.\n")
	}
	b.WriteString("Or add a folder with --font-dir <dir>.\n\nMore fonts:\n")
	for _, s := range fontSources {
		b.WriteString("  " + s + "\n")
	}
	b.WriteString("\nFonts are made by their own authors;\nthis project is not affiliated with them.")
	return b.String()
}
