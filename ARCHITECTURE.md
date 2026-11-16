# Architecture

Planning lives in GitHub issues on `jnaO/figlet-font-preview`: Stories carry the plan, Task sub-issues track the steps.

`figlet-font-preview` is a Go terminal app that renders a sample string in every installed FIGlet/TOIlet font, with a fuzzy-searchable, multi-select font sidebar. It is standalone: it reads font files from disk and never runs the `figlet` binary, which is only the test oracle.

## Dependency direction

```
cmd/figlet-font-preview → tui → display → fontclass → catalog → figfont
                       │   │       ├──→ catalog
                       │   │       └──→ paint
                       │   ├──→ action
                       │   └──→ newfile
                       ├──→ catalog
                       ├──→ svgexport → paint
                       └──→ figfont (Parse, Render, Center, Trim, TrimBlock)
```

Nothing below `internal/tui` imports Bubble Tea, bubbles or lipgloss (check: `go list -deps $(go list ./internal/... | grep -v /internal/tui$) | grep charmbracelet` prints nothing; it covers every package added below `internal/tui` automatically). A second front-end (e.g. a desktop GUI) drives `display.State` and `catalog.Catalog` directly and reads `action.All()` for its menus.

## Map

```
.github/workflows/release.yml     on v* tag push: go test, then goreleaser release (actions pinned to SHAs; GITHUB_TOKEN,
                                  HOMEBREW_TAP_TOKEN and the five MACOS_SIGN_* / MACOS_NOTARY_* secrets)
.goreleaser.yaml                  release builds (CGO off; darwin/linux/windows × amd64/arm64; -X main.version), darwin
                                  sign + notarize (quill, after build and before archiving; off in snapshots), archives
                                  (with man/, completions/ and THIRD_PARTY_NOTICES), checksums, homebrew_casks → jnaO/homebrew-tap (HOMEBREW_TAP_TOKEN, figlet dependency, manpage + bash/zsh/fish completion stanzas)
cmd/figlet-font-preview/main.go   registerFlags(fs) defines the flags (--font-dir repeatable, --center/--right write one display.Align, the last one given wins, --version: -X main.version ldflag, else build-info module version, else "dev"), positional text (default "Hello"); roots = catalog.DefaultRoots() + --font-dir; --version on a terminal prints the banner (printVersion: term.IsTerminal/GetSize, else the plain line)
cmd/figlet-font-preview/banner.go the --version banner: bannerFont(cat, width) picks AMC AAA01 (from the roots, measured at call time, used when widestWord fits), else the
                                  embedded banner/small/mini by bannerTable (min width = widestWord, drift-tested), else none; versionLines centres and appends the version
cmd/figlet-font-preview/fonts/    banner.flf, small.flf, mini.flf: FIGlet 2.2.5 fonts, byte for byte, go:embed'd (BSD-3 notice in THIRD_PARTY_NOTICES)
THIRD_PARTY_NOTICES               credits and the FIGlet licence for the embedded fonts; listed in the goreleaser archives
man/figlet-font-preview.1         man(7) page: OPTIONS/FILES/KEYS mirror README's Usage, Fonts and Keys sections; a drift test in
                                  cmd/figlet-font-preview/main_test.go checks every flag and font root appears. Edit with README
completions/figlet-font-preview.{bash,zsh,fish}
                                  static shell completions (flags + directories after --font-dir); the same drift test checks
                                  each file names every flag. Brew links the zsh file as _figlet-font-preview
internal/figfont/
  font.go          Parse/ParseFile: flf2a + tlf2a headers, zip-archived fonts, glyph table; THE font parser
  reader.go        byte-level line reader reproducing figlet's myfgets, header sscanf subset, C strtol (code tags)
  font.go/GlyphKey exact-duplicate key (glyph table is private to figfont)
  render.go        Font.Render(text, width): THE renderer — port of figlet's layout, smushing, RTL and word wrap; Center/Right/Trim/TrimBlock block helpers
  golden_test.go   every installed font vs the figlet binary, unwrapped and wrapped; skips without figlet/fonts
  font_test.go     portable fixture tests (testdata/: tiny flf, zipped flf, 0x7F-hardblank tlf, RTL flf)
internal/catalog/catalog.go
                   SystemRoots (package-manager dirs, empty on Windows; the only roots tests use), DefaultRoots() = SystemRoots + ~/figlet-fonts;
                   Discover(roots): recursive walk following symlinks, *.flf/*.tlf only, dedupe by real path
                   (Entry.Root = index of the first root that reached the file),
                   unique names; Entry.Load (once, concurrent-safe), Entry.State/Font (non-blocking), Preload(workers);
                   NoFontsMessage(roots, goos, addHint): the empty-catalog text (searched roots, per-GOOS install hint, the front-end's addHint line, FontSources mirrored in README.md, disclaimer)
internal/fontclass/
  analyze.go       Analyze(font) Profile: Stroke/Weight/Size/Slant from fixed probes; Axes() is THE axis list
  index.go         Index: duplicate groups + representative, facet values, Similar(name) by fingerprint
internal/paint/paint.go
                   THE colour model: RGB, Rainbow.At(row, col) (lolcat formula), Background.At(x, y, w, h)
                   (theme → false; solid/linear/radial sampled at cell centres in box-normalised coordinates, the
                   same geometry as SVG objectBoundingBox gradients; LinearVector gives the SVG endpoints), Style
internal/action/action.go
                   THE action list: id, title, group, in menu order; no keys (each front-end binds its own)
internal/svgexport/svg.go
                   Render(lines, top, left, style): one <tspan> per non-space rune at its own x, coloured
                   Rainbow.At(top+r, left+c) so it matches the preview cell; background rect/gradient, none in theme
internal/newfile/newfile.go
                   Write(dir, base, ext, data): never-overwrite file writer (O_EXCL; -2, -3 … suffixes; returns the absolute path); SafeName(name) maps unsafe file-name characters to '_'
internal/display/display.go
                   State: text, Align (left/centre/right; Align.Apply maps to figfont.Center/Right), Paint (paint.Style), query, selection; Sidebar() and RenderSet(); THE render-set rule;
                   Lines(font, width) and Block(catalog, name, width): THE cursor-font render (Align.Apply over Font.Render; Block trims with figfont.TrimBlock and returns Top/Left, or the footer error "no font under the cursor" / "<name> is not loaded")
internal/tui/
  tui.go           Bubble Tea model: framed, labelled text bar (textBar; grows to maxTextLines rows), left column (sidebarView: fontsView = search, count,
                   cursor/pick legend, list; filtersView = filter window, each in its own frame), unframed preview viewport, footer.
                   Five focus values, four tab stops (Text, Fonts = focusList, Filters, Style; nextFocus/fontsStop): search is a
                   position inside Fonts (/ and ctrl+f; ctrl+t reaches Filters; esc/enter return to the list). focusFonts (keys.go) closes
                   Style first when it has hidden Fonts. esc and q never quit
                   fontsHidden: Fonts steps aside while Style is open or sliding and width < bothFitWidth (79 = Fonts + minPreviewWidth
                   preview + Style); previewWidth and View read it, nextFocus skips Fonts zones while hidden
  frame.go         frame(title, key, body, outer, height, padL, padR, focused): THE pane border (rounded, title and key in the top edge,
                   dim unfocused, accent focused) used by the text bar, Fonts, Filters and Style (Style has one column of right padding: styleFrameCols = 3);
                   bottomEdge(label, outer): the focused bottom edge carrying a label (help's scroll hint);
                   overlay(bg, box, x, y): splices a box over the screen string with ansi.Truncate/TruncateLeft, keeping the colours around it
  palette.go       ctrl+k command palette: paletteState (input, fuzzy rows over action.All() minus palette, cursor/offset), an accent-framed
                   overlay on the first body row; its own keys first, any other globalKeys key closes it and runs (close before run)
  help.go          ? help screen: helpState (bubbles viewport); helpSections = Marks (cursor/pick legend), Actions (action.All() + binding().show),
                   Keys by pane (static paneKeys table), laid out by helpContent in one shared key column with descriptions ansi.Wrap'd and hung
                   under their column; the bottom edge reads "↓ more · esc closes" until the end, then "esc closes"; opens outside focusText/focusSearch (or from the palette)
  keys.go          globalKeys: THE app-wide key-to-action table (tab, ctrl+o/y/s/e/f/t/k, pgup/pgdn, ctrl+c); handleKey looks keys
                   up here before any zone code, which takes ctrl+e/f/t/k from the textarea and textinput; footerKeys builds the
                   one footer from the entries that have a help label (whole entries, ctrl+c always last). keyEntry.action is the
                   internal/action id (empty for navigation); zoneKeys holds the actions a zone handles itself (space, x, ?;
                   no keys, so lookupKey skips them); binding(id) finds an action's entry; a test checks every action.All() id
                   has one binding
  options.go       row vocabulary shared by both panes (optRow, optKind, rowCursor, applyOption, optionLine(row, cursor, width)):
                   filterRows (Fonts pane: Collapse, Group by, Similar to the cursor font, one row per facet value, Clear filters;
                   optAxis/optHeading rows are not selectable) and styleRows (the Align cycler, then paintRows). Style is the
                   right-hand column toggled by global ctrl+o (toggleStyle; slides in/out via slideMsg, panelW feeds previewWidth;
                   refused below styleFitWidth = 47), itemsChanged (keeps the cursor font), buildIndex/indexMsg
  paintopts.go     colour rows of the Style pane: slider specs (range, step, wrap period, which background modes
                   enable them), stepSlider, slider/heading/mode line rendering
  copy.go          global ctrl+y copyCursor (State.Block at the preview width, THE cursor-font render, shared with export) → writeClipboard (atotto/clipboard: pbcopy on macOS; a var tests replace);
                   setStatus footer messages clear after statusFor
  export.go        global ctrl+s exportCursor → svgexport.Render with State.Paint → newfile.Write (SafeName for the file name)
                   in exportDir (a var tests replace)
  nofonts.go       addFolderHint: the CLI's add-folder line, passed to catalog.NoFontsMessage by previewContent
  lolcat.go        the cursor font's colouring from State.Paint: per-rune Rainbow, background flat (theme tint or From)
                   or per cell over preview width × block rows (linear/radial); bgStyle for its header row
  styles.go        all lipgloss styles (adaptive light/dark); tintColour + themeSeed (background sliders start at the tint)
```

## Patterns

- **Universe pipeline** (`display.State.universe`): with a `fontclass.Index` set, the font list first goes through Collapse (representatives only), then facet filters (OR within an axis, AND across axes), then the Similar anchor plus up to `SimilarCap` look-alikes in distance order; selected fonts are always added back. With no index or every option off, the universe is all fonts. Collapse starts on (`display.New`) and takes effect when the index arrives.
- **Render set** (`display.State.RenderSet`), over the universe: no query + no selection → the sidebar's fonts in sidebar order; no query → selection (sorted); query → selection first, then fuzzy matches not already listed. The sidebar shows the universe without a query, otherwise only matches in `sahilm/fuzzy` rank order; with `GroupBy` set, a header `Item` precedes each non-empty group (axis value order, `Unclassified` last) and `ToggleGroup` selects or deselects the whole group.
- **Font names** (`catalog.uniqueNames`): basename without extension; on a basename collision, `parentdir/basename`; on a further collision, the full path minus extension. Names come from the resolved real path, so a symlink alias collapses into its target's name.
- **Cursor font in the preview** (`tui.Model.revealCursor`, `revealOffset`): `previewContent` records each rendered font's line span; on cursor move, query change, toggle or filter change the preview centres the cursor font as far as the viewport's clamp allows (its top when taller than the view), even when it is already visible, and never when it is not in the render set. The cursor font is drawn with `lolcat`.
- **Preview anchor** (`tui.Model.refreshPreview`, `anchorFont`): every re-render that is not a reset keeps the anchor font (the cursor font if rendered, else the font at the top of the view) on the same screen row, so text edits, resizes and fonts finishing loading do not shift it.
- **Text area sizing** (`tui.Model.handleKey`, `resetTextScroll`): the bubbles textarea scrolls its own view during `Update` from its current height, so it is updated at `maxTextLines` and then fitted; after lines are deleted its stale scroll offset is reset via `SetValue` with the cursor restored. Tests call `View()` after each update because bubbles components only learn their content size when rendered.
- **Lazy loading**: the TUI calls `Catalog.Preload` at start and polls on a tick until the returned channel closes; preview entries show `loading…` or the parse error until ready. When loading finishes, `buildIndex` builds the `fontclass.Index` off the UI goroutine and `indexMsg` hands it to `display.State.SetIndex`; until then the Fonts pane's filter rows show `analysing…` and do nothing. The full preview is re-rendered on every change.
- **Colour in tests**: lipgloss emits no colour escapes under `go test` (no TTY profile); tests asserting colours call `trueColour(t)` in `tui_test.go`, which sets `termenv.TrueColor` and restores it.
- **Sidebar headers** (`display.Item.IsHeader`): a header has no `Name`, so `cursorName()` is empty on it and the preview reveal/anchor ignore it; space on a header calls `ToggleGroup`.

## Drill-downs

- [.agent-docs/figfont.md](.agent-docs/figfont.md) — font format quirks, figlet-compatibility rules, golden-test exclusions.
- [.agent-docs/fontclass.md](.agent-docs/fontclass.md) — duplicate key, facet measures and thresholds, slant calibration, fingerprint.
