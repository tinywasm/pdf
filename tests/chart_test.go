package pdf_test

import (
	"os"
	"testing"

	"github.com/tinywasm/pdf"
)

func TestCharts(t *testing.T) {
	const out = "test_charts.pdf"
	_ = os.Remove(out)

	tf, err := pdf.LoadTypeface(
		fontDir+"Roboto-Regular.ttf",
		fontDir+"Roboto-Bold.ttf",
		fontDir+"Roboto-Italic.ttf",
		fontDir+"Roboto-BoldItalic.ttf",
	)
	if err != nil {
		t.Fatalf("loading typeface: %v", err)
	}

	doc := pdf.NewDocument(tf)
	doc.AddPage()

	doc.AddHeader1("Chart Examples")

	// Bar Chart
	doc.AddHeader2("Bar Chart")
	doc.Chart().Bar().
		Title("Monthly Sales").
		Height(100).
		AddBar(120, "Jan", "#3264C8").
		AddBar(140, "Feb", "#C86432").
		AddBar(110, "Mar", "#32C864").
		Draw()

	doc.SpaceBefore(10)

	// Line Chart
	doc.AddHeader2("Line Chart")
	doc.Chart().Line().
		Title("Growth Trends").
		Height(100).
		AddSeries("Revenue", []float64{10, 15, 13, 17, 20, 25, 22}, "#0000FF").
		Draw()

	doc.SpaceBefore(10)

	// Pie Chart
	doc.AddHeader2("Pie Chart")
	doc.Chart().Pie().
		Title("Market Share").
		Height(120).
		AddSlice("A", 40, "#FF0000").
		AddSlice("B", 30, "#00FF00").
		AddSlice("C", 30, "#0000FF").
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
