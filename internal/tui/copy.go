package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/jnaO/figlet-font-preview/internal/figfont"
)

const statusFor = 2 * time.Second

// writeClipboard puts text on the system clipboard (pbcopy on macOS).
var writeClipboard = clipboard.WriteAll

// copiedMsg reports a finished clipboard write.
type copiedMsg struct {
	name string
	err  error
}

// clearStatusMsg clears the footer status set with the same id.
type clearStatusMsg struct{ id int }

// block is the cursor font's render as the preview lays it out (width,
// centring), trimmed; top and left are the rows and columns trimmed off.
type block struct {
	name      string
	lines     []string
	top, left int
}

// cursorBlock is THE cursor-font render shared by copy and export. When there
// is nothing to use it returns a footer message instead.
func (m Model) cursorBlock() (block, string) {
	name := m.cursorName()
	e, ok := m.cat.Get(name)
	if !ok {
		return block{}, "no font under the cursor"
	}
	f, err := e.Font()
	if err != nil || f == nil {
		return block{}, name + " is not loaded"
	}
	lines := f.Render(m.state.Text, m.previewWidth())
	if m.state.Center {
		lines = figfont.Center(lines, f.Height, m.previewWidth())
	}
	out, top, left := figfont.TrimBlock(lines)
	return block{name, out, top, left}, ""
}

// copyCursor copies the cursor block as plain text.
func (m *Model) copyCursor() tea.Cmd {
	b, msg := m.cursorBlock()
	if msg != "" {
		return m.setStatus(msg)
	}
	text := strings.Join(b.lines, "\n") + "\n"
	return func() tea.Msg { return copiedMsg{b.name, writeClipboard(text)} }
}

// setStatus shows s in the footer until statusFor passes.
func (m *Model) setStatus(s string) tea.Cmd {
	m.statusID++
	m.status = s
	id := m.statusID
	return tea.Tick(statusFor, func(time.Time) tea.Msg { return clearStatusMsg{id} })
}

func (m *Model) copied(msg copiedMsg) tea.Cmd {
	if msg.err != nil {
		return m.setStatus(fmt.Sprintf("copy failed: %v", msg.err))
	}
	return m.setStatus("copied " + msg.name)
}
