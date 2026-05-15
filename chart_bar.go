package pdf

import (
	. "github.com/tinywasm/fmt"
)

type BarChart struct {
	doc    *Document
	title  string
	width  float64
	height float64
	bars   []barData
}

type barData struct {
	label string
	value float64
	color Color
}

func (c *BarChart) Title(t string) *BarChart {
	c.title = t
	return c
}

func (c *BarChart) Height(h float64) *BarChart {
	c.height = h
	return c
}

func (c *BarChart) Width(w float64) *BarChart {
	c.width = w
	return c
}

func (c *BarChart) AddBar(val float64, label string, col ...Color) *BarChart {
	var color Color
	if len(col) > 0 {
		color = col[0]
	} else {
		color = ColorRGB(100, 100, 100) // Default color
	}
	c.bars = append(c.bars, barData{
		label: label,
		value: val,
		color: color,
	})
	return c
}

func (c *BarChart) Draw() {
	if c.width == 0 {
		// Use available width
		w, _ := c.doc.internal.GetPageSize()
		l, _, r, _ := c.doc.internal.GetMargins()
		c.width = w - l - r
	}
	if c.height == 0 {
		c.height = 100 // Default height
	}

	x := c.doc.internal.GetX()
	y := c.doc.internal.GetY()

	// Title
	if c.title != "" {
		c.doc.internal.SetFont(c.doc.fontFamily, "B", 12)
		c.doc.internal.CellFormat(c.width, 10, c.title, "", 1, "C", false, 0, "")
		y = c.doc.internal.GetY() + 5
	}

	// Calculate Scale
	maxVal := 0.0
	for _, b := range c.bars {
		if b.value > maxVal {
			maxVal = b.value
		}
	}
	if maxVal == 0 {
		maxVal = 1
	}

	// Margins inside chart area
	margin := 20.0
	plotH := c.height - margin
	scaleY := plotH / maxVal
	barWidth := (c.width - margin) / float64(len(c.bars))

	// Draw Axes
	c.doc.internal.SetDrawColor(0, 0, 0)
	c.doc.internal.SetLineWidth(0.2)
	c.doc.internal.Line(x, y, x, y+c.height) // Y Axis
	c.doc.internal.Line(x, y+c.height, x+c.width, y+c.height) // X Axis

	// Draw Bars
	for i, bar := range c.bars {
		h := bar.value * scaleY
		bx := x + 10 + float64(i)*barWidth // 10 offset from Y axis
		by := y + c.height - h

		c.doc.internal.SetFillColor(bar.color.R, bar.color.G, bar.color.B)
		// Use simple rect
		c.doc.internal.Rect(bx+2, by, barWidth-4, h, "F")

		// Draw Text
		c.doc.internal.SetTextColor(0, 0, 0)
		c.doc.internal.SetFont(c.doc.fontFamily, "", 8)

		// Value on top
		valStr := Sprintf("%.1f", bar.value)
		wVal := c.doc.internal.GetStringWidth(valStr)
		c.doc.internal.Text(bx+(barWidth-wVal)/2, by-2, valStr)

		// Label on bottom
		wLbl := c.doc.internal.GetStringWidth(bar.label)
		c.doc.internal.Text(bx+(barWidth-wLbl)/2, y+c.height+4+4, bar.label) // +4 to descend below axis
	}

	c.doc.internal.SetY(y + c.height + 20) // Move below chart
}
