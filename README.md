# tinywasm/pdf
<img src="docs/img/badges.svg">

## Overview

This library is designed for web rendering with WebAssembly. It is optimized for TinyGo compatibility, with all standard library components that are not compatible with TinyGo being removed or replaced.

The rendering engine (`fpdf/`) is adapted from [gofpdf](https://github.com/jung-kurt/gofpdf), not a fork kept in sync with upstream: the legacy Type1/AFM font lineage and `internal/` machinery were removed, and the public surface was rebuilt around a typed, WASM-safe API (`LoadDeclared`, `Canvas`, `color.Color`) — see `docs/ARCHITECTURE.md`.

## New Flow-First Layout API

TwPDF now features a declarative, flow-first API that prioritizes composition over absolute positioning.

### Example

```go
import (
	"github.com/tinywasm/font"
	"github.com/tinywasm/pdf"
)

// Declare the font family and path
d := font.Declare("Roboto", "fonts/")
tf, err := pdf.LoadDeclared(d)
if err != nil {
	panic(err)
}

// Create a new document with the typeface
doc := pdf.NewDocument(tf)
doc.AddPage()

doc.AddTable().
    Cols("30%", "auto").
    Row("Key", "Value").
    Row("Name", "John Doe").
    Draw()

doc.WritePdf("output.pdf")
```

See `examples/contract_layout/main.go` for a full example.
