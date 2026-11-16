# figlet-font-preview

An interactive figlet font renderer: type a short text and see it rendered in every FIGlet and TOIlet font installed on your computer, with a fuzzy-searchable, multi-select font sidebar for comparing fonts side by side.

It renders fonts itself, so the `figlet` program does not need to be installed. Only the font files do.

## Install

Homebrew (macOS; Linux is best effort):

```sh
brew install --cask jnaO/tap/figlet-font-preview
```

This also installs `figlet`, which brings its fonts, the `figlet-font-preview(1)` man page, and bash, zsh and fish completions. If tab completion does not work, see Homebrew's [shell completion setup](https://docs.brew.sh/Shell-Completion). From v0.1.1 the macOS binaries are signed and notarized by Apple.

Prebuilt archives for macOS, Linux and Windows (amd64 and arm64) are on the [Releases](https://github.com/jnaO/figlet-font-preview/releases) page. Unpack one and put `figlet-font-preview` on your `PATH`. The archives also carry the man page under `man/` and the completions under `completions/`.

With Go 1.26 or newer:

```sh
go install github.com/jnaO/figlet-font-preview/cmd/figlet-font-preview@latest
```

### Fonts

No fonts are bundled. They are read from these directories when present:

- `/opt/homebrew/share/figlet`
- `/usr/local/share/figlet`
- `/usr/share/figlet`
- `~/figlet-fonts`

Windows searches only `%USERPROFILE%\figlet-fonts` and `--font-dir`. Fonts in different directories are listed separately even when they are the same font.

To get fonts:

- macOS: `brew install figlet toilet` provides about 440 fonts.
- Linux: install the `figlet` and `toilet` packages with your package manager.
- Anywhere: put `.flf`/`.tlf` files in `~/figlet-fonts`, or point `--font-dir` at a folder of them.

More fonts: [figlet.org](https://www.figlet.org), [cmatsuoka/figlet-fonts](https://github.com/cmatsuoka/figlet-fonts) and [cacalabs/toilet](https://github.com/cacalabs/toilet). Fonts are made by their own authors; this project is not affiliated with them.

## Usage

```sh
figlet-font-preview                       # renders "Hello"
figlet-font-preview Hello world           # initial text
figlet-font-preview $'two\nlines'         # multiline initial text
figlet-font-preview --font-dir ~/fonts    # also search ~/fonts (repeatable)
figlet-font-preview --center Hello        # centre the text in the preview
figlet-font-preview --right Hello         # right-align it; the last of --center / --right wins
figlet-font-preview --version             # banner on a terminal; plain version line otherwise
```

### Which fonts are shown

| Search | Selection | Preview shows |
|---|---|---|
| empty | none | every font |
| empty | some | the selected fonts |
| active | any | the selected fonts, then every fuzzy match |

The selection is kept while you search and when you clear the search. The preview follows the highlighted font in the list, centring that font vertically as far as scrolling allows, and keeps it on the same row when you edit the text.

### Keys

| Where | Key | Action |
|---|---|---|
| text | type, `enter` | edit the text; `enter` starts a new line (the box grows to 5 rows, then scrolls). `ctrl+e`, `ctrl+f`, `ctrl+t` and `ctrl+k` are app keys here, so `end` and `→` do the line-editing jobs |
| text | `esc` | go to the Fonts list |
| anywhere | `tab` / `shift+tab` | cycle focus: Text → Fonts → Filters → Style (while it is open); Fonts lands on the list |
| anywhere | `ctrl+e` | focus the text |
| anywhere | `ctrl+f` | focus the search; with Fonts hidden behind Style, closes Style first |
| anywhere | `ctrl+t` | focus the filters, with the same rule |
| anywhere | `ctrl+k` | open the command palette: every action with its key, filtered as you type. `↑` / `↓` move, `enter` runs the action, `esc` or `ctrl+k` closes it; any other app key (such as `ctrl+e` or `ctrl+c`) closes it and does its job |
| anywhere except text and search | `?` | open the help screen: every action with its key, the keys of each pane and the cursor/pick legend. `↑` / `↓`, `j` / `k`, `pgup` / `pgdn` scroll it, `esc` or `?` closes it; other app keys close it and do their job. In the text bar and the search box `?` types a question mark |
| anywhere | `pgup` / `pgdn`, `ctrl+u` / `ctrl+d`, mouse wheel | scroll the preview |
| anywhere | `ctrl+c` | quit (`esc` and `q` never quit) |
| search | type | fuzzy-filter fonts |
| search | `↑` / `↓` | move in the list |
| search | `enter` | go to the list, keeping the query |
| search | `esc` | clear the search and go to the list |
| list | `↑` / `↓`, `j` / `k`, `g` / `G` | move |
| list | `space` / `enter` | select or deselect |
| list | `x` | clear the selection |
| list | `/` | jump to search |
| list | `esc` | clear the search |
| anywhere | `ctrl+o` | show or hide the Style pane; below 79 columns the Fonts pane steps aside while it is open, and below 47 columns Style does not open |
| anywhere | `ctrl+y` | copy the cursor font's render to the clipboard as plain text, wrapped and aligned as shown |
| anywhere | `ctrl+s` | save the cursor font's render as a coloured `<font>.svg` in the current directory (never overwrites: adds `-2`, `-3`, …) |
| filters | `↑` / `↓`, `j` / `k` | move |
| filters | `space` / `enter` / `→` / `l` | toggle the row, or step Group by forward |
| filters | `←` / `h` | step Group by back |
| filters | `esc` | back to the list |
| style | `↑` / `↓`, `j` / `k` | move |
| style | `space` / `enter` | change the row |
| style | `←` / `→`, `h` / `l` | change Align, background Mode, or a slider's value |
| style | `esc` | go to the Fonts list, leaving the pane open (below 79 columns Style closes so Fonts can show) |
| style | `ctrl+o` | close the pane |

### Fonts pane

The left column is two framed segments: Fonts (search, a count such as `412/768 · 2 ●` (listed, installed, picked), a legend (`›` is the cursor and drives copy and SVG, `●` is a pick and narrows the preview) and the font list) above Filters (a scrolling checklist). Every pane has a rounded frame that takes the accent colour while it has focus. Changes apply immediately. `tab` moves between the text, the Fonts list, Filters and the Style pane while it is open; `/` or `ctrl+f` jump to the search and `ctrl+t` to the filters.

- **Collapse duplicates**: shows fonts that are identical (the same font in two folders) only once, with `+N` showing how many copies are hidden. It is on at start.
- **Group by**: groups the font list by Stroke, Weight, Size or Slant. Press `esc` to reach the list, then space on a group header to select or deselect the whole group.
- **Similar to**: lists the font under the list cursor, followed by its closest look-alikes, most similar first.
- **Filters**: every value of the four measured axes, with font counts. Within an axis, any checked value matches; across axes, all must match.
  - **Stroke**: what the letters are drawn with: line art (`/ \ | _`), solid fill (`# * @` or letters), block characters (`█`), box-drawing characters (`═ ║`), or other Unicode.
  - **Weight**: how much of the letter area is inked.
  - **Size**: how many rows tall the letters are.
  - **Slant**: upright or slanted.

Selected fonts always stay visible, whatever is filtered or collapsed. Categories are measured from how each font renders, not from its name.

### Style pane

Press `ctrl+o` anywhere to slide the Style pane in from the right; it is closed at start. It has its own column, and changes apply immediately. On terminals narrower than 79 columns the Fonts pane steps aside while Style is open and returns when it closes; below 47 columns Style does not open.

- **Align**: cycles left, center and right for the text in the preview; left is the default. `--center` or `--right` sets the starting value, and the last one given wins.
- **Rainbow**: Seed, Spread and Frequency of the lolcat-style colouring of the highlighted font (defaults match `lolcat -S 20 -p 7 -F 0.05`).
- **Background**: Mode cycles `theme` (the terminal-adaptive tint; transparent in SVG), `solid` (From), `linear` (From → To along Angle; 0° runs left to right, 90° top to bottom) and `radial` (From at the centre to To at the edges). From and To are RGB sliders with a swatch; rows the mode does not use are dimmed. SVG export uses the same colours.

## Development

```sh
go test ./...
```

The golden tests compare the renderer byte for byte with the real `figlet` for every installed font, and skip when `figlet` or the fonts are missing. See [ARCHITECTURE.md](ARCHITECTURE.md).

### Releasing

Push a `vX.Y.Z` tag. The release workflow runs GoReleaser (`.goreleaser.yaml`), which publishes the archives to GitHub Releases and the cask to `jnaO/homebrew-tap`. It signs and notarizes the darwin binaries using the `MACOS_*` repo secrets; a run missing them fails before publishing anything.

## License

[MIT](LICENSE). Font files are not part of this project and keep their own licenses, except three FIGlet fonts (banner, small, mini) embedded for the `--version` banner under BSD-3; see [THIRD_PARTY_NOTICES](THIRD_PARTY_NOTICES).
