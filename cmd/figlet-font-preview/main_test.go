package main

import (
	"flag"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/jnaO/figlet-font-preview/internal/catalog"
	"github.com/jnaO/figlet-font-preview/internal/display"
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

func TestCompletionsCoverFlags(t *testing.T) {
	fs := flag.NewFlagSet("figlet-font-preview", flag.ContinueOnError)
	registerFlags(fs)
	for _, shell := range []string{"bash", "zsh", "fish"} {
		t.Run(shell, func(t *testing.T) {
			script := readRepoFile(t, "completions/figlet-font-preview."+shell)
			fs.VisitAll(func(f *flag.Flag) {
				if !strings.Contains(script, "--"+f.Name) && !strings.Contains(script, "-l "+f.Name) {
					t.Errorf("%s completion does not mention --%s", shell, f.Name)
				}
			})
			if !strings.Contains(script, "help") {
				t.Errorf("%s completion does not offer --help", shell)
			}
		})
	}
}

func TestAlignFlags(t *testing.T) {
	cases := []struct {
		args []string
		want display.Align
		rest []string
	}{
		{nil, display.AlignLeft, nil},
		{[]string{"--center"}, display.AlignCenter, nil},
		{[]string{"-right"}, display.AlignRight, nil},
		{[]string{"--center", "--right"}, display.AlignRight, nil},
		{[]string{"--right", "--center"}, display.AlignCenter, nil},
		{[]string{"--center", "--center=false"}, display.AlignLeft, nil},
		{[]string{"--right", "--center=false"}, display.AlignRight, nil},
		{[]string{"--center", "Hello"}, display.AlignCenter, []string{"Hello"}},
	}
	for _, c := range cases {
		fs := flag.NewFlagSet("figlet-font-preview", flag.ContinueOnError)
		_, align, _ := registerFlags(fs)
		if err := fs.Parse(c.args); err != nil {
			t.Fatalf("%v: %v", c.args, err)
		}
		if *align != c.want {
			t.Errorf("%v: align = %v, want %v", c.args, *align, c.want)
		}
		if len(fs.Args()) != len(c.rest) || (len(c.rest) > 0 && fs.Arg(0) != c.rest[0]) {
			t.Errorf("%v: args = %v, want %v", c.args, fs.Args(), c.rest)
		}
	}
}
