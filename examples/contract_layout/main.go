package main

import (
	"github.com/tinywasm/pdf"
)

func main() {
	doc := pdf.NewDocument(
		pdf.WithLogger(func(args ...any) { /* … */ }),
	)
	// doc.RegisterImage("logo", "examples/contract_layout/logo.png")
	doc.Load(func(err error) {
		if err != nil {
			panic(err)
		}
	})

	theme := pdf.DefaultTheme
	theme.Accent = "#1E3C78"
	theme.Header = "#F0F4FA"
	theme.Gray = "#646464"
	doc.SetTheme(theme)
	doc.AddPage()

	// Header band — replaces the legacy 6-call sequence
	doc.AddTable().
		Cols("auto", "60%", "auto").
		Background(theme.Header).
		BorderBottom(0.6, theme.Accent).
		Row(
			"LOGO", // pdf.Image("logo").Width(40),
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
