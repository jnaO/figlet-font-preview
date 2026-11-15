// Command figlet-font-preview renders text in every installed FIGlet/TOIlet font
// in an interactive terminal UI.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jnaO/figlet-font-preview/internal/catalog"
	"github.com/jnaO/figlet-font-preview/internal/tui"
)

// version is set by release builds with -ldflags "-X main.version=…".
var version = "dev"

// resolveVersion returns the version to print: the ldflag value when set,
// else the module version from build info (go install), else "dev".
func resolveVersion(ldflag string, info *debug.BuildInfo, ok bool) string {
	if ldflag != "dev" {
		return ldflag
	}
	if ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return strings.TrimPrefix(info.Main.Version, "v")
	}
	return "dev"
}

type dirList []string

func (d *dirList) String() string     { return strings.Join(*d, ",") }
func (d *dirList) Set(v string) error { *d = append(*d, v); return nil }

// registerFlags defines the command-line flags on fs.
func registerFlags(fs *flag.FlagSet) (dirs *dirList, center, showVersion *bool) {
	dirs = new(dirList)
	fs.Var(dirs, "font-dir", "extra directory to search for fonts (repeatable)")
	center = fs.Bool("center", false, "centre the rendered text in the preview")
	showVersion = fs.Bool("version", false, "print version and exit")
	return dirs, center, showVersion
}

func main() {
	dirs, center, showVersion := registerFlags(flag.CommandLine)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: figlet-font-preview [--center] [--font-dir DIR]... [text]\n\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nDefault font roots: %s\n", strings.Join(catalog.DefaultRoots(), ", "))
	}
	flag.Parse()
	if *showVersion {
		info, ok := debug.ReadBuildInfo()
		fmt.Println("figlet-font-preview", resolveVersion(version, info, ok))
		return
	}

	text := "Hello"
	if flag.NArg() > 0 {
		text = strings.Join(flag.Args(), " ")
	}

	cat := catalog.Discover(append(catalog.DefaultRoots(), *dirs...))
	p := tea.NewProgram(tui.New(cat, text, *center), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
