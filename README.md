# TinyPDF
<img src="docs/img/badges.svg">

## Overview

This library is designed for web rendering with WebAssembly. It is optimized for TinyGo compatibility, with all standard library components that are not compatible with TinyGo being removed or replaced.

The fork of go-pdf  https://github.com/jung-kurt/gofpdf

## New Flow-First Layout API

TinyPDF now features a declarative, flow-first API that prioritizes composition over absolute positioning.

### Example

```go
doc := pdf.NewDocument()
doc.AddPage()

doc.AddTable().
    Cols("30%", "auto").
    Row("Key", "Value").
    Row("Name", "John Doe").
    Draw()

doc.WritePdf("output.pdf")
```

See `examples/contract_layout/main.go` for a full example.
