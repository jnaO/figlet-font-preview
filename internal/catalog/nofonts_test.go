package catalog

import (
	"strings"
	"testing"
)

func TestNoFontsMessage(t *testing.T) {
	tests := []struct {
		goos, hint string
	}{
		{"darwin", "brew install figlet toilet"},
		{"windows", `%USERPROFILE%\figlet-fonts`},
		{"linux", "package manager"},
	}
	for _, tt := range tests {
		t.Run(tt.goos, func(t *testing.T) {
			msg := NoFontsMessage([]string{"/a", "/b"}, tt.goos, "ADD HINT")
			for _, want := range []string{"No fonts found.", "/a", "/b", tt.hint, "https://", "ADD HINT\n\nMore fonts:", "not affiliated"} {
				if !strings.Contains(msg, want) {
					t.Errorf("missing %q in:\n%s", want, msg)
				}
			}
		})
	}
}
