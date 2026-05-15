# PLAN: Flow-First Layout API for `tinywasm/pdf`

## 1. Context

The current public API of `tinywasm/pdf` exposes two parallel models — **flow** (`AddText().Draw()`, `AddHeader2()`) and **absolute** (`DrawTextAt`, `CellAt`, `SetCursorX/Y`, `DrawFilledRect`, `DrawLineH`). Real documents are written using a mix of both because no single model covers all cases. The result is verbose, error-prone, and forces the document author to compute coordinates and row heights by hand.

The motivating evidence comes from an external consumer (`veltylabs/contracts`, a separate repository that depends on this one). Inside that repo, a contract-generation file is built from low-level absolute calls with hardcoded `y+6`, `y+11`, `y+16`, … offsets: the reader cannot tell *why* a block of cells exists, only *where* each cell is placed. **That repository is out of scope for this plan** — it has its own migration plan that runs after this work ships. References to "the contract" in §4 and §10 are descriptive, not action items.

A previous attempt (`docpdfOLD`, also in a separate archived repository) explored a declarative flow API and was deprecated because it accumulated too much hidden state. This plan revives the **flow-first, declarative** concept while learning from the past failure: keep state explicit, expose a single layout primitive (a table), and prefer composition over hidden modes.

## 2. Goals

1. **Flow is the only public model.** Authors describe *what* they want; the library computes *where* it goes.
2. **Tables are the layout primitive.** The same `Table` builder handles header bands, two-column blocks, signature blocks, and tabular data. This mirrors how Word users already think about positioning.
3. **Theme is a document-level concept.** Colors, fonts and spacing live in one place; components read from the theme instead of receiving raw `R, G, B` triples on every call.
4. **No hidden global state in the public API.** Color, font and cursor side-effects stay inside components and reset automatically.
5. **TinyGo/WASM parity.** Every public symbol must compile and run identically in the backend and in a WASM frontend; no `time.Time`, `fmt.Sprintf`-with-reflection, or stdlib-only imports.

## 3. Non-Goals

- Backwards compatibility with the current public API. The package is pre-1.0 and the only known consumer has its own migration plan that gates on this release; it is not part of *this* plan.
- Replacing the lower-level `fpdf` engine. We are only redesigning the wrapper surface.
- Implementing rich text inside a single line (mixed bold/italic runs). One paragraph = one style for v1.

## 3a. Standard-library constraints (TinyGo / WASM parity)

This package must compile and run identically under Go on the backend and TinyGo in a WASM frontend. The rule of thumb is **prefer the project's own packages over stdlib**, and never reach for stdlib APIs that are missing or partial under TinyGo:

| Need | Use | Do **not** use |
|------|-----|----------------|
| Formatting / `Sprintf` / int-to-string / any-to-string | `github.com/tinywasm/fmt` (dot-imported, as the rest of the package already does — see `document.go`, `table.go`) | stdlib `fmt` |
| Dates and time | `github.com/tinywasm/time` (works on `int64` UnixNano / `string "YYYY-MM-DD"`) | stdlib `time.Time` in the public API |
| String building | `tinywasm/fmt` builders / plain `+` concatenation | `strings.Builder` (OK in non-hot paths; not preferred) |
| Conversion helpers (`strconv.Itoa`, `Atoi`) | `tinywasm/fmt.Convert(x).Int()` / `tinywasm/fmt.Sprint(x)` | `strconv` |

Test files (`*_test.go`) may import stdlib `testing` — that is the only stdlib package allowed unconditionally because `tinywasm/test` does not replace it.

The error type returned by `Draw()` and `WritePdf()` is plain `error`; constructed via `tinywasm/fmt.Errorf` (when the package gains it) or plain string concatenation otherwise. No `errors.New` from stdlib in hot paths.

Any pull request that introduces a new stdlib import must justify it in the PR description; reviewers reject silently-added stdlib usage.

## 4. Pain points to fix (concrete)

Symptoms observed in real consumer code (taken from the external `veltylabs/contracts` repo — not part of this repository, only quoted as motivation):

| # | Symptom | Root cause |
|---|---------|------------|
| 1 | `y := GetCursorY()` followed by `CellAt(20, y+6, …)`, `CellAt(20, y+11, …)`, `CellAt(20, y+16, …)` | No row/column primitive; author tracks Y manually |
| 2 | `SetCursorY(y + 40)` after a multi-row block | No "block consumed N units" feedback from `CellAt` |
| 3 | `DrawLineH` depends on whatever `GetX()` is at call time | Hidden state; required a `SetCursorX(20)` patch in consumer code |
| 4 | `SetTextColor(accent); AddHeader2(…); SetTextColor(0,0,0)` | Color is sticky global state |
| 5 | Logo + title + subtitle + accent rule built from 5 separate primitives | No composite "row of cells" available |
| 6 | Two identical `DrawLineH` calls for a signature block (duplicated-call bug, easy to introduce) | No semantic "signature line" abstraction |
| 7 | Consumer-side `AccentR, AccentG, AccentB` ints threaded through every call site | No theme concept; no `Color` type |

Each row should disappear by construction once this plan is implemented: there is no way to write pattern #1 against an API that has no public `CellAt`.

## 5. The new API — by example

### 5.1 Composition model

The new API is **constructor-style**, not fluent. Each builder takes its children as arguments:

```
Row(...Element)            // Row contains Cells (or shorthand elements)
Cell(...Element)           // Cell contains Texts, Images, Lines, sub-Tables
Text(string).Bold().Size() // leaf element; fluent only on leaves where ambiguity is impossible
```

Rationale: making `Cell` a fluent container with chained `.Text(…)` calls forces every leaf modifier (`.Bold()`, `.Size()`) to be ambiguous between "apply to the cell" and "apply to the last text added". This was the kind of hidden state that broke `docpdfOLD`. Constructors with explicit children remove the ambiguity.

### 5.2 Shorthand: when `Cell` is implicit

To avoid `pdf.Cell(pdf.Text("PRESTADOR"))` boilerplate for the common case, `Row` accepts `...any` and applies these rules:

| Argument type | Treated as |
|---------------|------------|
| `string`      | `pdf.Cell(pdf.Text(s))` |
| `*Text`       | `pdf.Cell(t)` |
| `*Image`      | `pdf.Cell(img)` |
| `*Line`       | `pdf.Cell(l)` |
| `*Cell`       | itself (used as-is) |

If you need cell-level properties (`Padding`, `Background`, `Border`, `Span`, vertical alignment) or heterogeneous content (text + image), use `pdf.Cell(...)` explicitly.

Anything else (e.g. an `int`, a custom struct) is a programmer error: `Draw()` returns a non-`nil` `error` of the form `"row arg %d: unsupported type %T (want string, *Text, *Image, *Line, *Cell)"`. No panics.

### 5.3 Self-contained example (target shape)

The example below is the **reference** for the `examples/contract_layout/main.go` deliverable in stage S5. It uses synthetic placeholder data and the package's own `DefaultTheme` so that it compiles inside this repository without any external dependency.

```go
package main

import "github.com/tinywasm/pdf"

func main() {
    doc := pdf.NewDocument(
        pdf.WithLogger(func(args ...any) { /* … */ }),
    )
    doc.RegisterImage("logo", "examples/contract_layout/logo.png")
    doc.Load(func(err error) {
        if err != nil { panic(err) }
    })

    theme := pdf.DefaultTheme
    theme.Accent = "#1E3C78"
    theme.Header = "#F0F4FA"
    theme.Gray   = "#646464"
    doc.SetTheme(theme)
    doc.AddPage()

    // Header band — replaces the legacy 6-call sequence
    // (DrawFilledRect + DrawImageAt + 2× DrawTextAt + SetCursorX + DrawLineH).
    doc.AddTable().
        Cols("auto", "60%", "auto").
        Background(theme.Header).               // hex "#F0F4FA"
        BorderBottom(0.6, theme.Accent).        // hex "#1E3C78"
        Row(
            pdf.Image("logo").Width(40),
            pdf.Cell(
                pdf.Text("EXAMPLE DOCUMENT TITLE").Bold().Size(14),
                pdf.Text("Subtitle — descriptive line").Size(9),
            ),
            pdf.Text("2025-10-23").AlignRight().Color(theme.Gray),
        ).
        Draw()

    // Two-column block — replaces 14 CellAt calls + manual y+6/y+11/…
    doc.AddHeader2("Section: Two-Column Block")
    doc.AddTable().
        Cols("50%", "50%").
        Row(
            pdf.Cell(
                pdf.Text("LEFT HEADER").Bold(),
                pdf.Text("first line"),
                pdf.Text("second line"),
                pdf.Text("third line"),
            ),
            pdf.Cell(
                pdf.Text("RIGHT HEADER").Bold(),
                pdf.Text("first line"),
                pdf.Text("second line"),
            ),
        ).
        Draw()

    // Signature block — replaces 2× DrawLineH + 2× DrawTextAt + 4× CellAt
    doc.AddTable().
        Cols("50%", "50%").
        KeepTogether().
        Row(
            pdf.Cell(
                pdf.Line(),
                pdf.Text("Party A"),
                pdf.Text("Date: 23/10/2025"),
            ),
            pdf.Cell(
                pdf.Line(),
                pdf.Text("Party B"),
                pdf.Text("Date: ___/___/______"),
            ),
        ).
        Draw()

    // Pure-string data table (shorthand)
    doc.AddTable().
        Cols("20%", "60%", "20%").
        Row("Code", "Product", "Price").
        Row("001", "Item A", "1299.99").
        Row("002", "Item B", "349.99").
        Draw()

    if err := doc.WritePdf("contract_layout.pdf"); err != nil {
        panic(err)
    }
}
```

A section header that today needs `SetTextColor + AddHeader2 + SetTextColor(0,0,0)` becomes one line:

```go
doc.AddHeader2("Section title")  // reads theme.Accent automatically, resets after
```

## 6. Architecture

### 6.1 Public surface (post-refactor)

The public API has three layers: **leaf elements** (constructed at package level), **containers** (Cell, Table) and the **Document**. Leaf elements expose fluent style methods because there is no ambiguity about what `.Bold()` applies to. Containers take their children as arguments — never as a fluent chain.

```
// Element interface — everything that can live inside a Cell or be passed to Row
type Element interface { /* internal */ }

// Leaf elements (package-level constructors)
pdf.Text(s string) *Text
  ├── Bold() *Text
  ├── Italic() *Text
  ├── Size(pt float64) *Text
  ├── Color(Color) *Text                  // Color is a CSS-style hex string
  ├── AlignLeft/Center/Right() *Text
  └── Justify() *Text

pdf.Image(name string) *Image
  ├── Width(mm float64) *Image
  ├── Height(mm float64) *Image
  └── AlignLeft/Center/Right() *Image

pdf.Line() *Line                          // horizontal rule (signature line, etc.)
  ├── Width(mm float64) *Line             // optional fixed width
  ├── Color(Color) *Line                  // hex: "#1E3C78"
  └── Thickness(mm float64) *Line

// Containers (constructor-style)
pdf.Cell(children ...Element) *Cell
  ├── Padding(top, right, bottom, left float64) *Cell
  ├── Background(Color) *Cell             // hex: "#F0F4FA"
  ├── Border(side, width float64, color Color) *Cell
  ├── Span(cols, rows int) *Cell
  ├── AlignLeft/Center/Right() *Cell      // horizontal alignment of children
  └── AlignTop/Middle/Bottom() *Cell      // vertical alignment within row height

// Document
NewDocument(...Option) *Document
  ├── SetTheme(Theme) *Document
  ├── AddPage() *Document
  ├── AddText(s) *Text                    // shorthand: flow-mode text
  ├── AddHeader1/2/3(s) *Document         // theme-aware
  ├── AddSeparator() *Document
  ├── AddSpace(units float64) *Document
  ├── AddImage(name) *Image               // shorthand: flow-mode image
  ├── AddTable() *TableBuilder            // <-- the layout primitive
  ├── SetPageHeader() *PageBand
  ├── SetPageFooter() *PageBand
  ├── RegisterFont(family, path string) *Document
  ├── RegisterImage(name, path string) *Document
  ├── Load(cb func(error))                // async resource loading (preserved from current API)
  ├── WritePdf(path) error
  └── OutputTo(w io.Writer) error

// TableBuilder
TableBuilder
  ├── Cols(...string) *TableBuilder       // "30%", "auto", "40mm"
  ├── Row(items ...any) *TableBuilder     // items: string | *Text | *Image | *Cell | *Line
  ├── Background(Color) *TableBuilder
  ├── BorderBottom(width float64, c Color) *TableBuilder
  ├── Border(...) *TableBuilder
  ├── KeepTogether() *TableBuilder
  └── Draw() *Document                    // errors accumulate on doc; surfaced at WritePdf/OutputTo
```

**Note on `Cell` alignment**: alignment methods come in two axes. `AlignLeft/Center/Right()` controls the horizontal position of children inside the cell width; `AlignTop/Middle/Bottom()` controls vertical position inside the row height (the row height is `max(cell heights)`, so vertical alignment matters when one cell in the row is taller than its sibling — e.g. a tall image next to a short text). Defaults: `AlignLeft + AlignTop`.

**Note on the `Row(...any)` shorthand**: an argument that is not already a `*Cell` is auto-wrapped via the rules in §5.2. The implementation is a small switch over the dynamic type; `any` is acceptable here because the alternative — forcing `pdf.Cell(pdf.Text("Code"))` for every data-table cell — is the verbosity we explicitly want to remove. Type checking happens at `Draw()` with a clear error message for unsupported types.

**Note on fluent vs. constructor**: leaves keep fluent style because `pdf.Text("foo").Bold()` has no ambiguity — `.Bold()` can only mean one thing. Containers (`Cell`, `Table`) refuse fluent child-appending precisely to prevent the "what does `.Bold()` apply to" question.

**Note on when content is rendered (the drawing trigger)**: the package has three render triggers:

| Builder | Returned by | Render trigger |
|---------|-------------|----------------|
| `*Text` | `doc.AddText(s)` | Explicit `.Draw()` — required before the next flow element |
| `*Image` | `doc.AddImage(name)` | Explicit `.Draw()` |
| `*Document` (header/separator/space) | `doc.AddHeader2(s)`, `AddSeparator()`, `AddSpace(u)` | **Immediate** — returns the `*Document` so chaining keeps reading top-to-bottom |
| `*TableBuilder` | `doc.AddTable()` | Explicit `.Draw()` |

The asymmetry (`AddHeader2` is immediate, `AddText` needs `.Draw()`) is deliberate: a header has no fluent style methods to apply (color comes from the theme), so requiring `.Draw()` would only be noise. Text has many style methods, so the explicit `.Draw()` boundary tells the reader "the chain is finished, draw it now".

**Options for `NewDocument`**:

```go
pdf.WithLogger(fn func(...any)) Option
pdf.WithPageSize(w, h float64, unit string) Option   // "mm", "in"
pdf.WithMargins(left, top, right, bottom float64) Option
pdf.WithDefaultFont(family, path string) Option
```

Theme is **not** an option (it is set after construction with `SetTheme(...)`) because the default theme has to be valid before any setup runs and the consumer commonly wants to update it later.

**Error story**: the public API never panics on bad input. Builder methods (`Draw`, `Row`, `AddText`, …) return `*Document` for chainability and **accumulate** errors on the document. Terminal calls surface them:

- `Document.Load(cb func(error))` — async; the callback receives the first resource-load error
- `Document.WritePdf(path) error` and `Document.OutputTo(w io.Writer) error` — return the first accumulated error, if any, before/after writing
- `Document.Err() error` — explicit accessor to inspect the accumulated error without writing

Rationale: forcing `if err := doc.AddTable()…Draw(); err != nil` on every block would destroy the readability the new API exists for. The error is *always* surfaced at write time, so it cannot be silently lost; the chain itself stays clean. This mirrors how the underlying `fpdf` engine already accumulates errors.

There is **no** custom `LayoutError` type. Errors are plain `error` constructed with `tinywasm/fmt.Errorf` (when available) or string concatenation. The value contains the offending input (column index, cell content excerpt, hex string) so it diagnoses itself.

### 6.2 Private (internal) primitives

These move from `*Document` methods to **unexported** helpers that components use. They are no longer part of the public surface:

```
d.drawTextAt(x, y, text, style, size)
d.cellAt(x, y, w, h, text, style, size, align)
d.drawFilledRect(x, y, w, h, color Color)
d.drawLineH(x, y, width float64, color Color, thickness float64)
d.setCursorX(x float64)
d.setCursorY(y float64)
d.getCursorX() float64
d.getCursorY() float64
d.measureText(text, style, size) (width, height float64)
```

Rationale: the contract demonstrates that authors should never call these directly. Components — `TableBuilder`, `TextComponent`, `PageBand` — are the only callers.

### 6.3 Theme and colors

Colors in the public API are **hex strings** (CSS-style). No `R, G, B int` triples anywhere on the public surface.

```go
// Type alias for self-documenting signatures
type Color string  // "#RRGGBB" or "#RGB" (shorthand expands to #RRGGBB)

type Theme struct {
    Accent     Color   // "#1E3C78"
    Header     Color   // "#F0F4FA"
    Gray       Color   // "#646464"
    Body       Color   // "#000000"
    FontFamily string
    Sizes      struct {
        H1, H2, H3, Body, Small float64
    }
    Spacing    struct {
        Paragraph, Section, Page float64
    }
}

var DefaultTheme = Theme{
    Accent: "#1E3C78",
    Header: "#F0F4FA",
    Gray:   "#646464",
    Body:   "#000000",
    // …
}
```

Every public function that previously took `r, g, b int` now takes a single `Color`:

```go
// Before
BorderBottom(0.6, 30, 60, 120)
Background(240, 244, 250)
pdf.Text(s).Color(100, 100, 100)

// After
BorderBottom(0.6, "#1E3C78")
Background("#F0F4FA")
pdf.Text(s).Color("#646464")
// …or, using the theme:
BorderBottom(0.6, theme.Accent)
Background(theme.Header)
pdf.Text(s).Color(theme.Gray)
```

Parsing rules at the boundary (`#fff` → `#FFFFFF`):

1. Strip leading `#`. Accept lowercase or uppercase.
2. If length is 3, duplicate each char: `#abc` → `#aabbcc`.
3. If length is 6, use as-is. Anything else → fail with a clear error at `Draw()` time.
4. Convert to `(r, g, b)` ints internally for the `fpdf` engine.

Invalid hex strings are not silently treated as black; they surface a `LayoutError` at `Draw()`. This keeps the contract: "the public surface is CSS-style hex; if you typed it wrong, you find out at render time, not as a mystery color."

Consumers that today thread loose `R, G, B int` constants through every call site replace them with a single `pdf.Theme` value passed to `SetTheme` once at document setup.

### 6.4 Layout engine for `Table`

A `Table` lays out in two passes:

1. **Resolve column widths**: percentages → mm based on content area; `auto` → measured intrinsic width of the cell; `Npx`/`Nmm` → literal.
2. **Layout rows**: for each row, lay out each cell's content vertically inside its column width, compute the natural cell height, take the row height = max of cell heights. If `KeepTogether` is set and the row does not fit on the current page, emit a page break first. Otherwise rows break individually.

This is the same algorithm Word uses for layout tables. It also subsumes the "two-column block" pattern because a 2-column / 1-row table with no borders *is* a two-column block.

### 6.5 Avoiding the `docpdfOLD` failure mode

The previous attempt failed because of hidden state: `inlineMode`, `lastInlineWidth`, `inHeaderFooterDraw`. Each method's behavior depended on flags set elsewhere.

This plan keeps state **inside builders, not inside the document**:

- `TableBuilder` owns rows/cells until `Draw()` is called; nothing leaks to `Document`.
- `*Text` / `*Image` / `*Line` own their style; calling `Draw()` is the only side effect on `Document`.
- `Document` only owns: current page, current cursor, theme, accumulated error, registered fonts/images, page bands.
- No `inlineMode`, no `lastWidth`, no "what was the last thing I drew" flags.

Concretely, the **forbidden patterns** during code review:

1. A field on `*Document` named `last*`, `current*Element`, `pending*` (except `currentCursor`, `currentPage` which are intrinsic to a streaming document).
2. A public method that *only* makes sense if a previous method was called first (other than the obvious `AddPage` → `AddText` ordering, which is intrinsic to PDFs).
3. A builder field that mutates after `Draw()` returns. Builders are write-once: configured, then drawn, then discarded.

## 7. Implementation stages

Each stage produces a green test suite before moving on.

| Stage | Deliverable | Validation |
|-------|-------------|------------|
| **S1 — Theme + colors + private primitives** | Add `Theme` type, `SetTheme`, `DefaultTheme`. Add `Color` type (string alias) with hex parser (`#RGB` / `#RRGGBB`, case-insensitive, errors on invalid). Add `NewDocument` options (`WithLogger`, `WithPageSize`, `WithMargins`, `WithDefaultFont`). Make all `Draw*`/`Cell*`/`SetCursor*`/`GetCursor*` methods package-private (lowercase). Update `AddHeader1/2/3`, `AddSeparator` to consume the theme (and reset color after). Add error accumulation: `Document.Err()`, `WritePdf` / `OutputTo` surface accumulated errors. | Existing `pdf` package tests pass. New `theme_test.go` covers default + override. New `color_test.go` covers `#fff`/`#ffffff` expansion, case-insensitivity, and 3 representative invalid inputs. |
| **S2 — Leaf elements + `Cell`** | New `element.go` with `pdf.Text`, `pdf.Image`, `pdf.Line` (leaves, fluent) and `pdf.Cell(children ...Element)` (container, constructor-style). Pure data; no rendering yet. | Unit tests on builder state and on the `Cell` children list. |
| **S3 — `TableBuilder` layout engine** | `table_layout.go` with column-width resolution, row-height measurement and the `Row(...any)` shorthand resolver (string/`*Text`/`*Image`/`*Cell`). Render via the now-private primitives. Support `Cols`, `Row`, `Background`, `BorderBottom`, `KeepTogether`. | Golden-PDF test for: (a) 3-col header band, (b) 2-col parties block, (c) signature row with `KeepTogether`, (d) data table using string shorthand. |
| **S4 — Nested tables in cells** | `pdf.Cell(pdf.Table(sub))` paints a sub-table inside a cell (recursive layout). Add the `Table(*TableBuilder)` element so a sub-table can appear as a cell child. | Test in this repo: a 3-cell row where the middle cell contains a 2-row mini-table (e.g. title + subtitle). Verify both intrinsic-width and percentage-width modes. |
| **S5 — Reference example + documentation** | Add `examples/contract_layout/main.go` (in this repo) that reproduces the *shape* of a real-world contract (header band + two-column parties + signature row) using **synthetic** placeholder data — no external repos referenced. This is both a smoke test and the documentation example. Update `README.md` to point at it. Move the `Draw*`/`Cell*` removal into `CHANGELOG.md`. | The example compiles and produces a PDF whose visual structure matches the description above. README example compiles. |

**Out of this plan**: actually migrating any *consumer* of this library (e.g. `veltylabs/contracts`) is **not** part of this work. The consumer owns its own migration plan that gates on the release of this one. That plan describes the consumer-side migration steps and is the place where consumer files like `2025_CONTRACT_MANPC19.go` are touched.

## 8. Risks & mitigations

| Risk | Mitigation |
|------|------------|
| `auto` column width measurement is incorrect for nested content | Two-pass layout: measure with infinite width first, then constrain. Add explicit test for nested tables and long words. |
| `KeepTogether` row that exceeds page height would loop forever | Detect oversize row → render anyway with a logged warning; never silently truncate. |
| Removing public `DrawTextAt/CellAt/SetCursorX` breaks downstream code | The only known downstream consumer has its own migration plan that gates on this release. The break is intentional and made explicit with a tagged release; the `CHANGELOG.md` entry from S5 lists every removed symbol. |
| The plan re-creates the `docpdfOLD` failure | Discipline: no hidden state in `Document`; every component owns its own builder; state is reset at `Draw()` end. Reviewed at the end of each stage. |

## 9. Open questions for later (not blockers)

- Do we need `Row().Height("auto"|fixed)`? Probably yes; deferred to S3 if needed.
- Page numbering placeholders inside `PageBand` — already exists in old API as `WithPageTotal(...)`; port in a later stage, not blocking S1–S5.
- Should `pdf.Cell` accept a callback variant (`Cell(func(c *Cell) { c.Add(…) })`) for procedurally-generated children (e.g. looping over a slice)? Idiomatic Go is to build the slice first and spread it (`pdf.Cell(children...)`), so a callback variant is not added by default. Revisit if real call sites end up awkward.

## 10. Success criteria (verifiable inside this repo)

- No exported function in the `pdf` package takes `x` or `y` coordinates, and there is no exported `GetCursorY` / `SetCursor*`. (`grep -E '^\s*func \(.*\) [A-Z].*\b(x|y)\b float64' *.go` returns empty.)
- No exported function takes `r, g, b int` triples; every color parameter has type `pdf.Color`. (`grep -E '^\s*func \(.*\) [A-Z].* int, int, int' *.go` returns empty.)
- `examples/contract_layout/main.go` compiles and produces a PDF whose visual structure matches §7 S5.
- All package-level tests pass under both Go and TinyGo, on Linux. No new stdlib imports were introduced (see §3a); CI greps for forbidden imports.

The above criteria are checkable on this repository alone — no external consumer is required to validate the work.
