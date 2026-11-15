package main

import (
	"flag"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/jnaO/figlet-font-preview/internal/catalog"
)

func TestResolveVersion(t *testing.T) {
	info := func(v string) *debug.BuildInfo {
		return &debug.BuildInfo{Main: debug.Module{Version: v}}
	}
	tests := []struct {
		name   string
		ldflag string
		info   *debug.BuildInfo
		ok     bool
		want   string
	}{
		{"ldflag set", "0.1.0", info("v0.0.9"), true, "0.1.0"},
		{"build info version", "dev", info("v0.1.0"), true, "0.1.0"},
		{"pseudo-version", "dev", info("v0.0.0-20261115164800-724e308abcde+dirty"), true, "0.0.0-20261115164800-724e308abcde+dirty"},
		{"devel", "dev", info("(devel)"), true, "dev"},
		{"empty build info version", "dev", info(""), true, "dev"},
		{"no build info", "dev", nil, false, "dev"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveVersion(tt.ldflag, tt.info, tt.ok); got != tt.want {
				t.Errorf("resolveVersion(%q) = %q, want %q", tt.ldflag, got, tt.want)
			}
		})
	}
}

func readRepoFile(t *testing.T, rel string) string {
	t.Helper()
	b, err := os.ReadFile("../../" + rel)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestManPageCoversFlagsAndRoots(t *testing.T) {
	page := strings.ReplaceAll(readRepoFile(t, "man/figlet-font-preview.1"), `\-`, "-")
	fs := flag.NewFlagSet("figlet-font-preview", flag.ContinueOnError)
	registerFlags(fs)
	fs.VisitAll(func(f *flag.Flag) {
		if !strings.Contains(page, "--"+f.Name) {
			t.Errorf("man page does not mention --%s", f.Name)
		}
	})
	if runtime.GOOS == "windows" {
		return
	}
	for _, root := range append(append([]string{}, catalog.SystemRoots...), "~/"+catalog.UserDir) {
		if !strings.Contains(page, root) {
			t.Errorf("man page does not mention font root %s", root)
		}
	}
}
