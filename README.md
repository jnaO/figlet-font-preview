# figlet-font-preview

An interactive figlet font renderer: type a short text and see it rendered in every FIGlet and TOIlet font installed on your computer, with a fuzzy-searchable, multi-select font sidebar for comparing fonts side by side.

It renders fonts itself, so the `figlet` program does not need to be installed. Only the font files do.

## Install

Homebrew (macOS; Linux is best effort):

```sh
brew install --cask jnaO/tap/figlet-font-preview
```

This also installs `figlet`, which brings its fonts. From v0.1.1 the macOS binaries are signed and notarized by Apple.

Prebuilt archives for macOS, Linux and Windows (amd64 and arm64) are on the [Releases](https://github.com/jnaO/figlet-font-preview/releases) page. Unpack one and put `figlet-font-preview` on your `PATH`.

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
```

### Which fonts are shown

| Search | Selection | Preview shows |
|---|---|---|
| empty | none | every font |
| empty | some | the selected fonts |
| active | any | the selected fonts, then every fuzzy match |

The selection is kept while you search and when you clear the search. The preview follows the highlighted font in the list, scrolling only when that font is not fully visible, and keeps it on the same row when you edit the text.

### Keys

| Where | Key | Action |
|---|---|---|
| text | type, `enter` | edit the text; `enter` starts a new line (the box grows to 5 rows, then scrolls) |
| anywhere | `tab` / `shift+tab` | cycle focus: text → search → list |
| anywhere | `pgup` / `pgdn`, `ctrl+u` / `ctrl+d`, mouse wheel | scroll the preview |
| anywhere | `ctrl+c` | quit |
| search | type | fuzzy-filter fonts |
| search | `↑` / `↓` | move in the list |
| search | `enter` | select or deselect the highlighted font |
| search | `esc` | clear the search |
| list | `↑` / `↓`, `j` / `k`, `g` / `G` | move |
| list | `space` / `enter` | select or deselect |
| list | `x` | clear the selection |
| list | `/` | jump to search |
| list | `esc` | clear the search |
| anywhere | `ctrl+o` | open or close the options pane |
| anywhere | `ctrl+y` | copy the cursor font's render to the clipboard as plain text, wrapped and centred as shown |
| anywhere | `ctrl+s` | save the cursor font's render as a coloured `<font>.svg` in the current directory (never overwrites: adds `-2`, `-3`, …) |
| list | `q` | quit |
| options | `↑` / `↓`, `j` / `k` | move |
| options | `space` / `enter` | toggle the row |
| options | `←` / `→`, `h` / `l` | change Group by, background Mode, or a slider's value |
| options | `tab` | back to the text, leaving the pane open |
| options | `esc` / `ctrl+o` | close the pane |

### Options panel

Press `ctrl+o` anywhere to slide the options pane in from the right. It has its own column, and changes apply immediately. `tab` moves between the text, search, list and pane.

- **Center text**: centres the text in the preview. `--center` sets the starting value.
- **Rainbow**: Seed, Spread and Frequency of the lolcat-style colouring of the highlighted font (defaults match `lolcat -S 20 -p 7 -F 0.05`).
- **Background**: Mode cycles `theme` (the terminal-adaptive tint; transparent in SVG), `solid` (From), `linear` (From → To along Angle; 0° runs left to right, 90° top to bottom) and `radial` (From at the centre to To at the edges). From and To are RGB sliders with a swatch; rows the mode does not use are dimmed. SVG export uses the same colours.
- **Collapse duplicates**: shows fonts that are identical (the same font in two folders) only once, with `+N` showing how many copies are hidden.
- **Group by**: groups the font list by Stroke, Weight, Size or Slant. Press space on a group header to select or deselect the whole group.
- **Similar to**: lists the font under the list cursor, followed by its closest look-alikes, most similar first.
- **Filters**: every value of the four measured axes, with font counts. Within an axis, any checked value matches; across axes, all must match.
  - **Stroke**: what the letters are drawn with: line art (`/ \ | _`), solid fill (`# * @` or letters), block characters (`█`), box-drawing characters (`═ ║`), or other Unicode.
  - **Weight**: how much of the letter area is inked.
  - **Size**: how many rows tall the letters are.
  - **Slant**: upright or slanted.

Selected fonts always stay visible, whatever is filtered or collapsed. Categories are measured from how each font renders, not from its name.

## Development

```sh
go test ./...
```

The golden tests compare the renderer byte for byte with the real `figlet` for every installed font, and skip when `figlet` or the fonts are missing. See [ARCHITECTURE.md](ARCHITECTURE.md).

### Releasing

Push a `vX.Y.Z` tag. The release workflow runs GoReleaser (`.goreleaser.yaml`), which publishes the archives to GitHub Releases and the cask to `jnaO/homebrew-tap`. It signs and notarizes the darwin binaries using the `MACOS_*` repo secrets; a run missing them fails before publishing anything.

## License

[MIT](LICENSE). Font files are not part of this project and keep their own licenses.
