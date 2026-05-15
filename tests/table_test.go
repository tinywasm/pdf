package pdf_test

import (
	"testing"

	"github.com/tinywasm/pdf"
)

func TestTable(t *testing.T) {
	doc := pdf.NewDocument()
	doc.AddPage()

	doc.AddTable().
		Cols("20", "80", "30").
		Row("Code", "Product", "Price").
		Row("001", "Widget A", "10.00").
		Row("002", "Widget B", "20.50").
		Row("003", "Widget C", "5.99").
		Draw()

	err := doc.WritePdf("test_table.pdf")
	if err != nil {
		t.Errorf("WritePdf failed: %v", err)
	}
}
