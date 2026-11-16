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
	b, err := m.state.Block(m.cat, m.cursorName(), m.previewWidth())
	if err != nil {
		return m.setStatus(err.Error())
	}
	doc := svgexport.Render(b.Lines, b.Top, b.Left, m.state.Paint)
	return func() tea.Msg {
		path, err := newfile.Write(exportDir, newfile.SafeName(b.Name), ".svg", doc)
		return exportedMsg{path, err}
	}
}

func (m *Model) exported(msg exportedMsg) tea.Cmd {
	if msg.err != nil {
		return m.setStatus(fmt.Sprintf("export failed: %v", msg.err))
	}
	return m.setStatus("saved " + msg.path)
}
