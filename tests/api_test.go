package pdf_test

import (
	"os"
	"testing"

	"webtyp.com/font"
	"webtyp.com/pdf"
)

func TestAPI_Basic(t *testing.T) {
	const out = "test_api.pdf"
	_ = os.Remove(out)

	d := font.Declare("Roboto", fontDir)
	tf, err := pdf.LoadDeclared(d)
	if err != nil {
		t.Fatalf("loading typeface: %v", err)
	}

	doc := pdf.NewDocument(tf)
	doc.SetPageHeader().SetLeftText("Test Header")
	doc.SetPageFooter().WithPageTotal("R")
	doc.AddPage()
	doc.AddHeader1("Hello World")
	doc.AddText("This is a test document.").Draw()
	doc.SpaceBefore(10)
	doc.AddText("Another paragraph.").Bold().AlignRight().Draw()

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
