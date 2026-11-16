// Command figlet-font-preview renders text in every installed FIGlet/TOIlet font
// in an interactive terminal UI.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/term"

	"github.com/jnaO/figlet-font-preview/internal/catalog"
	"github.com/jnaO/figlet-font-preview/internal/display"
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

// alignFlag is a boolean flag that sets *a to to; flags sharing one *a are
// mutually exclusive and the last one parsed wins.
type alignFlag struct {
	a  *display.Align
	to display.Align
}

func (f alignFlag) IsBoolFlag() bool { return true }

func (f alignFlag) String() string {
	if f.a == nil {
		return "false"
	}
	return strconv.FormatBool(*f.a == f.to)
}

func (f alignFlag) Set(v string) error {
	on, err := strconv.ParseBool(v)
	if err != nil {
		return err
	}
	if on {
		*f.a = f.to
	} else if *f.a == f.to {
		*f.a = display.AlignLeft
	}
	return nil
}

// registerFlags defines the command-line flags on fs.
func registerFlags(fs *flag.FlagSet) (dirs *dirList, align *display.Align, showVersion *bool) {
	dirs = new(dirList)
	fs.Var(dirs, "font-dir", "extra directory to search for fonts (repeatable)")
	align = new(display.Align)
	fs.Var(alignFlag{align, display.AlignCenter}, "center", "centre the rendered text in the preview")
	fs.Var(alignFlag{align, display.AlignRight}, "right", "right-align the rendered text in the preview")
	showVersion = fs.Bool("version", false, "print version and exit")
	return dirs, align, showVersion
}

// printVersion prints line, as a centred banner above it when stdout is a
// terminal whose width fits a banner font.
func printVersion(roots []string, line string) {
	fd := os.Stdout.Fd()
	if !term.IsTerminal(fd) {
		fmt.Println(line)
		return
	}
	width, _, err := term.GetSize(fd)
	if err != nil {
		fmt.Println(line)
		return
	}
	f, _ := bannerFont(catalog.Discover(roots), width)
	fmt.Println(strings.Join(versionLines(f, line, width), "\n"))
}

func main() {
	dirs, align, showVersion := registerFlags(flag.CommandLine)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: figlet-font-preview [--center | --right] [--font-dir DIR]... [text]\n\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nDefault font roots: %s\n", strings.Join(catalog.DefaultRoots(), ", "))
	}
	flag.Parse()
	roots := append(catalog.DefaultRoots(), *dirs...)
	if *showVersion {
		info, ok := debug.ReadBuildInfo()
		printVersion(roots, "figlet-font-preview "+resolveVersion(version, info, ok))
		return
	}

	text := "Hello"
	if flag.NArg() > 0 {
		text = strings.Join(flag.Args(), " ")
	}

	cat := catalog.Discover(roots)
	p := tea.NewProgram(tui.New(cat, text, *align), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
