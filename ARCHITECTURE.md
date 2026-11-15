# Architecture

Planning lives in GitHub issues on `jnaO/figlet-font-preview`: Stories carry the plan, Task sub-issues track the steps.

`figlet-font-preview` is a Go terminal app that renders a sample string in every installed FIGlet/TOIlet font, with a fuzzy-searchable, multi-select font sidebar. It is standalone: it reads font files from disk and never runs the `figlet` binary, which is only the test oracle.

## Dependency direction

```
cmd/figlet-font-preview → tui → display → fontclass → catalog → figfont
                       │           └──→ paint
                       ├──→ catalog
                       ├──→ svgexport → paint
                       └──→ figfont (Trim, TrimBlock)
```

Nothing below `internal/tui` imports Bubble Tea, bubbles or lipgloss (check: `go list -deps ./internal/figfont ./internal/catalog ./internal/display ./internal/fontclass ./internal/paint ./internal/svgexport | grep charmbracelet` prints nothing). A second front-end (e.g. a desktop GUI) drives `display.State` and `catalog.Catalog` directly.

## Map

```
.github/workflows/release.yml     on v* tag push: go test, then goreleaser release (actions pinned to SHAs; GITHUB_TOKEN,
                                  HOMEBREW_TAP_TOKEN and the five MACOS_SIGN_* / MACOS_NOTARY_* secrets)
.goreleaser.yaml                  release builds (CGO off; darwin/linux/windows × amd64/arm64; -X main.version), darwin
                                  sign + notarize (quill, after build and before archiving; off in snapshots), archives
                                  (with man/ and completions/), checksums, homebrew_casks → jnaO/homebrew-tap (HOMEBREW_TAP_TOKEN, figlet dependency, manpage + bash/zsh/fish completion stanzas)
cmd/figlet-font-preview/main.go   registerFlags(fs) defines the flags (--font-dir repeatable, --center/--right write one display.Align, the last one given wins, --version: -X main.version ldflag, else build-info module version, else "dev"), positional text (default "Hello"); roots = catalog.DefaultRoots() + --font-dir
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
                   unique names; Entry.Load (once, concurrent-safe), Entry.State/Font (non-blocking), Preload(workers)
internal/fontclass/
  analyze.go       Analyze(font) Profile: Stroke/Weight/Size/Slant from fixed probes; Axes() is THE axis list
  index.go         Index: duplicate groups + representative, facet values, Similar(name) by fingerprint
internal/paint/paint.go
                   THE colour model: RGB, Rainbow.At(row, col) (lolcat formula), Background.At(x, y, w, h)
                   (theme → false; solid/linear/radial sampled at cell centres in box-normalised coordinates, the
                   same geometry as SVG objectBoundingBox gradients; LinearVector gives the SVG endpoints), Style
internal/svgexport/svg.go
                   Render(lines, top, left, style): one <tspan> per non-space rune at its own x, coloured
                   Rainbow.At(top+r, left+c) so it matches the preview cell; background rect/gradient, none in theme
internal/display/display.go
                   State: text, Align (left/centre/right; Align.Apply maps to figfont.Center/Right), Paint (paint.Style), query, selection; Sidebar() and RenderSet(); THE render-set rule
internal/tui/
  tui.go           Bubble Tea model: framed, labelled text bar (textBar; grows to maxTextLines rows), Fonts pane (sidebarView:
                   search, filter window, count, list; focusFilters sits between search and list), preview viewport, footer help.
                   fontsHidden: Fonts steps aside while Style is open or sliding and width < bothFitWidth (76 = Fonts + minPreviewWidth
                   preview + Style); previewWidth and View read it, nextFocus skips Fonts zones while hidden
  options.go       row vocabulary shared by both panes (optRow, optKind, rowCursor, applyOption, optionLine(row, cursor, width)):
                   filterRows (Fonts pane: Collapse, Group by, Similar to the cursor font, one row per facet value, Clear filters;
                   optAxis/optHeading rows are not selectable) and styleRows (the Align cycler, then paintRows). Style is the
                   right-hand column toggled by global ctrl+o (toggleStyle; slides in/out via slideMsg, panelW feeds previewWidth;
                   refused below styleFitWidth = 45), itemsChanged (keeps the cursor font), buildIndex/indexMsg
  paintopts.go     colour rows of the Style pane: slider specs (range, step, wrap period, which background modes
                   enable them), stepSlider, slider/heading/mode line rendering
  copy.go          cursorBlock: THE cursor-font render shared by copy and export (preview width, State.Align.Apply, figfont.TrimBlock);
                   global ctrl+y copyCursor → writeClipboard (atotto/clipboard: pbcopy on macOS; a var tests replace);
                   setStatus footer messages clear after statusFor
  export.go        global ctrl+s exportCursor → svgexport.Render with State.Paint → writeNew (O_EXCL, -2, -3 … suffixes)
                   in exportDir (a var tests replace); safeName for font names with '/'
  nofonts.go       noFontsMessage: the empty-catalog preview (searched roots, per-GOOS install hint, fontSources
                   links mirrored in README.md, disclaimer)
  lolcat.go        the cursor font's colouring from State.Paint: per-rune Rainbow, background flat (theme tint or From)
                   or per cell over preview width × block rows (linear/radial); bgStyle for its header row
  styles.go        all lipgloss styles (adaptive light/dark); tintColour + themeSeed (background sliders start at the tint)
```

## Patterns

- **Universe pipeline** (`display.State.universe`): with a `fontclass.Index` set, the font list first goes through Collapse (representatives only), then facet filters (OR within an axis, AND across axes), then the Similar anchor plus up to `SimilarCap` look-alikes in distance order; selected fonts are always added back. With no index or every option off, the universe is all fonts.
- **Render set** (`display.State.RenderSet`), over the universe: no query + no selection → the sidebar's fonts in sidebar order; no query → selection (sorted); query → selection first, then fuzzy matches not already listed. The sidebar shows the universe without a query, otherwise only matches in `sahilm/fuzzy` rank order; with `GroupBy` set, a header `Item` precedes each non-empty group (axis value order, `Unclassified` last) and `ToggleGroup` selects or deselects the whole group.
- **Font names** (`catalog.uniqueNames`): basename without extension; on a basename collision, `parentdir/basename`; on a further collision, the full path minus extension. Names come from the resolved real path, so a symlink alias collapses into its target's name.
- **Cursor font in the preview** (`tui.Model.revealCursor`, `revealOffset`): `previewContent` records each rendered font's line span; on cursor move, query change or toggle the preview scrolls the minimum needed to show the cursor font in full (its top when taller than the view), and never when it is already fully visible or not in the render set. The cursor font is drawn with `lolcat`.
- **Preview anchor** (`tui.Model.refreshPreview`, `anchorFont`): every re-render that is not a reset keeps the anchor font (the cursor font if rendered, else the font at the top of the view) on the same screen row, so text edits, resizes and fonts finishing loading do not shift it.
- **Text area sizing** (`tui.Model.handleKey`, `resetTextScroll`): the bubbles textarea scrolls its own view during `Update` from its current height, so it is updated at `maxTextLines` and then fitted; after lines are deleted its stale scroll offset is reset via `SetValue` with the cursor restored. Tests call `View()` after each update because bubbles components only learn their content size when rendered.
- **Lazy loading**: the TUI calls `Catalog.Preload` at start and polls on a tick until the returned channel closes; preview entries show `loading…` or the parse error until ready. When loading finishes, `buildIndex` builds the `fontclass.Index` off the UI goroutine and `indexMsg` hands it to `display.State.SetIndex`; until then the Fonts pane's filter rows show `analysing…` and do nothing. The full preview is re-rendered on every change.
- **Colour in tests**: lipgloss emits no colour escapes under `go test` (no TTY profile); tests asserting colours call `trueColour(t)` in `tui_test.go`, which sets `termenv.TrueColor` and restores it.
- **Sidebar headers** (`display.Item.IsHeader`): a header has no `Name`, so `cursorName()` is empty on it and the preview reveal/anchor ignore it; space on a header calls `ToggleGroup`.

## Drill-downs

- [.agent-docs/figfont.md](.agent-docs/figfont.md) — font format quirks, figlet-compatibility rules, golden-test exclusions.
- [.agent-docs/fontclass.md](.agent-docs/fontclass.md) — duplicate key, facet measures and thresholds, slant calibration, fingerprint.
