package pdf_test

import (
	"testing"

	"github.com/tinywasm/pdf"
)

func TestCharts(t *testing.T) {
	doc := pdf.NewDocument()
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

	err := doc.WritePdf("test_charts.pdf")
	if err != nil {
		t.Errorf("WritePdf failed: %v", err)
	}
}
