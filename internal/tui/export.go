package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jnaO/figlet-font-preview/internal/newfile"
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
		path, err := newfile.Write(exportDir, newfile.SafeName(b.name), ".svg", doc)
		return exportedMsg{path, err}
	}
}

func (m *Model) exported(msg exportedMsg) tea.Cmd {
	if msg.err != nil {
		return m.setStatus(fmt.Sprintf("export failed: %v", msg.err))
	}
	return m.setStatus("saved " + msg.path)
}
