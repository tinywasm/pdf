package pdf_test

import (
	"os"
	"testing"

	"webtyp.com/font"
	"webtyp.com/pdf"
)

func TestTable(t *testing.T) {
	const out = "test_table.pdf"
	_ = os.Remove(out)

	d := font.Declare("Roboto", fontDir)
	tf, err := pdf.LoadDeclared(d)
	if err != nil {
		t.Fatalf("loading typeface: %v", err)
	}

	doc := pdf.NewDocument(tf)
	doc.AddPage()

	doc.AddTable().
		Cols("20", "80", "30").
		Row("Code", "Product", "Price").
		Row("001", "Widget A", "10.00").
		Row("002", "Widget B", "20.50").
		Row("003", "Widget C", "5.99").
		Draw()

	writeErr := doc.WritePdf(out)
	if writeErr != nil {
		t.Fatalf("WritePdf failed: %v", writeErr)
	}

	st, err := os.Stat(out)
	if err != nil {
		t.Fatalf("stat %s: %v", out, err)
	}
	if st.Size() == 0 {
		t.Fatalf("PDF %s is empty", out)
	}
}
