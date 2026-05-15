package pdf

import (
	. "github.com/tinywasm/fmt"
)

type Table struct {
	doc         *Document
	columns     []*TableColumn
	rows        [][]string
	headerStyle Style
}

type TableColumn struct {
	table  *Table
	header string
	width  float64
	align  string
	prefix string
	suffix string
}

func (d *Document) AddTable() *Table {
	return &Table{
		doc:     d,
		columns: make([]*TableColumn, 0),
		rows:    make([][]string, 0),
	}
}

func (t *Table) AddColumn(header string) *TableColumn {
	c := &TableColumn{
		table:  t,
		header: header,
		align:  "L",
	}
	t.columns = append(t.columns, c)
	return c
}

// Delegate methods for TableColumn to support chaining

func (c *TableColumn) AddColumn(header string) *TableColumn {
	return c.table.AddColumn(header)
}

func (c *TableColumn) Width(w float64) *TableColumn {
	c.width = w
	return c
}

func (c *TableColumn) AlignLeft() *TableColumn {
	c.align = "L"
	return c
}

func (c *TableColumn) AlignRight() *TableColumn {
	c.align = "R"
	return c
}

func (c *TableColumn) AlignCenter() *TableColumn {
	c.align = "C"
	return c
}

func (c *TableColumn) Prefix(p string) *TableColumn {
	c.prefix = p
	return c
}

func (c *TableColumn) Suffix(s string) *TableColumn {
	c.suffix = s
	return c
}

// Table methods that can be called from TableColumn too

func (t *Table) HeaderStyle(s Style) *Table {
	t.headerStyle = s
	return t
}

func (c *TableColumn) HeaderStyle(s Style) *Table {
	return c.table.HeaderStyle(s)
}

func (t *Table) AddRow(values ...any) *Table {
	row := make([]string, len(values))
	for i, v := range values {
		row[i] = Sprintf("%v", v)
	}
	t.rows = append(t.rows, row)
	return t
}

func (c *TableColumn) AddRow(values ...any) *Table {
	return c.table.AddRow(values...)
}

func (t *Table) Draw() *Document {
	// Draw Header
	// Save current font settings?
	// For simplicity, we just set what we need.

	fontFamily := t.doc.internal.GetFontFamily()
	if fontFamily == "" {
		fontFamily = t.doc.fontFamily
	}

	headerFont := t.headerStyle.Font
	if headerFont == "" {
		headerFont = "B"
	}

	headerSize := t.headerStyle.FontSize
	if headerSize == 0 {
		headerSize = 12
	}

	t.doc.internal.SetFont(fontFamily, headerFont, headerSize)

	// Apply colors
	if t.headerStyle.FillColor != (Color{}) {
		t.doc.internal.SetFillColor(t.headerStyle.FillColor.R, t.headerStyle.FillColor.G, t.headerStyle.FillColor.B)
	} else {
		t.doc.internal.SetFillColor(255, 255, 255)
	}

	if t.headerStyle.TextColor != (Color{}) {
		t.doc.internal.SetTextColor(t.headerStyle.TextColor.R, t.headerStyle.TextColor.G, t.headerStyle.TextColor.B)
	} else {
		t.doc.internal.SetTextColor(0, 0, 0)
	}

	// Draw Header Row
	for _, col := range t.columns {
		t.doc.internal.CellFormat(col.width, 10, col.header, "1", 0, "C", true, 0, "")
	}
	t.doc.internal.Ln(10)

	// Draw Data
	t.doc.internal.SetFont(fontFamily, "", 12) // Reset to regular 12
	t.doc.internal.SetTextColor(0, 0, 0)
	t.doc.internal.SetFillColor(255, 255, 255)

	for _, row := range t.rows {
		for i, val := range row {
			if i < len(t.columns) {
				col := t.columns[i]
				text := col.prefix + val + col.suffix
				t.doc.internal.CellFormat(col.width, 10, text, "1", 0, col.align, false, 0, "")
			}
		}
		t.doc.internal.Ln(10)
	}

	return t.doc
}

func (c *TableColumn) Draw() *Document {
	return c.table.Draw()
}
