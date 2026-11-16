// Package action is the list of commands every front-end offers. It holds ids,
// titles and groups only; each front-end binds its own keys.
package action

// ID is a stable kebab-case identifier.
type ID string

const (
	ExportSVG   ID = "export-svg"
	Quit        ID = "quit"
	Copy        ID = "copy"
	EditText    ID = "edit-text"
	SearchFonts ID = "search-fonts"
	TogglePick  ID = "toggle-pick"
	ClearPicks  ID = "clear-picks"
	Filters     ID = "filters"
	ToggleStyle ID = "toggle-style"
	Palette     ID = "palette"
	Help        ID = "help"
)

// Action is one command. Group is a menu-bar name.
type Action struct {
	ID           ID
	Title, Group string
}

// All returns the actions in menu order.
func All() []Action {
	return []Action{
		{ExportSVG, "Save cursor font as SVG", "File"},
		{Quit, "Quit", "File"},
		{Copy, "Copy cursor font", "Edit"},
		{EditText, "Edit text", "Edit"},
		{SearchFonts, "Search fonts", "Edit"},
		{TogglePick, "Pick or unpick cursor font", "Edit"},
		{ClearPicks, "Clear picks", "Edit"},
		{Filters, "Go to filters", "View"},
		{ToggleStyle, "Show or hide Style", "View"},
		{Palette, "Command palette", "View"},
		{Help, "Keys and help", "Help"},
	}
}

// Find returns the action with the given id.
func Find(id ID) (Action, bool) {
	for _, a := range All() {
		if a.ID == id {
			return a, true
		}
	}
	return Action{}, false
}
