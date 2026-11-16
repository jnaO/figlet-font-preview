package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
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

// copyCursor copies the cursor block as plain text.
func (m *Model) copyCursor() tea.Cmd {
	b, err := m.state.Block(m.cat, m.cursorName(), m.previewWidth())
	if err != nil {
		return m.setStatus(err.Error())
	}
	text := strings.Join(b.Lines, "\n") + "\n"
	return func() tea.Msg { return copiedMsg{b.Name, writeClipboard(text)} }
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
