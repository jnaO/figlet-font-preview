package tui

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jnaO/figlet-font-preview/internal/svgexport"
)

// exportDir is where ctrl+s writes SVGs.
var exportDir = "."

// exportedMsg reports a finished SVG export.
type exportedMsg struct {
	path string
	err  error
}

// exportCursor writes the cursor block as <font>.svg in exportDir, coloured
// with the current paint style, never overwriting an existing file.
func (m *Model) exportCursor() tea.Cmd {
	b, msg := m.cursorBlock()
	if msg != "" {
		return m.setStatus(msg)
	}
	doc := svgexport.Render(b.lines, b.top, b.left, m.state.Paint)
	return func() tea.Msg {
		path, err := writeNew(exportDir, safeName(b.name), ".svg", doc)
		return exportedMsg{path, err}
	}
}

func (m *Model) exported(msg exportedMsg) tea.Cmd {
	if msg.err != nil {
		return m.setStatus(fmt.Sprintf("export failed: %v", msg.err))
	}
	return m.setStatus("saved " + msg.path)
}

// safeName replaces characters that are unsafe in file names with '_'.
func safeName(name string) string {
	return strings.Map(func(r rune) rune {
		if r < 0x20 || strings.ContainsRune(`/\:*?"<>|`, r) {
			return '_'
		}
		return r
	}, name)
}

// writeNew creates dir/base+ext, or base-2, base-3, … when taken, and
// returns its absolute path. O_EXCL makes the existence check atomic.
func writeNew(dir, base, ext string, data []byte) (string, error) {
	for n := 1; ; n++ {
		name := base + ext
		if n > 1 {
			name = fmt.Sprintf("%s-%d%s", base, n, ext)
		}
		path := filepath.Join(dir, name)
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if errors.Is(err, fs.ErrExist) {
			continue
		}
		if err != nil {
			return "", err
		}
		_, werr := f.Write(data)
		if cerr := f.Close(); werr == nil {
			werr = cerr
		}
		if werr != nil {
			return "", werr
		}
		return filepath.Abs(path)
	}
}
