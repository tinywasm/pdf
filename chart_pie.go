package pdf

import (
	"math"
)

type PieChart struct {
	doc    *Document
	title  string
	width  float64
	height float64
	slices []pieSlice
}

type pieSlice struct {
	label string
	value float64
	color Color
}

func (c *PieChart) Title(t string) *PieChart {
	c.title = t
	return c
}

func (c *PieChart) Height(h float64) *PieChart {
	c.height = h
	return c
}

func (c *PieChart) Width(w float64) *PieChart {
	c.width = w
	return c
}

func (c *PieChart) AddSlice(label string, val float64, col Color) *PieChart {
	c.slices = append(c.slices, pieSlice{
		label: label,
		value: val,
		color: col,
	})
	return c
}

func (c *PieChart) Draw() {
	if c.width == 0 {
		w, _ := c.doc.internal.GetPageSize()
		l, _, r, _ := c.doc.internal.GetMargins()
		c.width = w - l - r
	}
	if c.height == 0 {
		c.height = 100
	}

	x := c.doc.internal.GetX()
	y := c.doc.internal.GetY()

	// Title
	if c.title != "" {
		c.doc.internal.SetFont(c.doc.fontFamily, "B", 12)
		c.doc.internal.CellFormat(c.width, 10, c.title, "", 1, "C", false, 0, "")
		y = c.doc.internal.GetY() + 5
	}

	total := 0.0
	for _, s := range c.slices {
		total += s.value
	}

	cx := x + c.width/2
	cy := y + c.height/2
	radius := c.height / 2
	if c.width < c.height {
		radius = c.width / 2
	}
	radius -= 10 // Padding

	startAngle := 0.0

	c.doc.internal.SetLineWidth(0.2)
	c.doc.internal.SetDrawColor(255, 255, 255) // White borders

	for _, s := range c.slices {
		angle := (s.value / total) * 360.0
		endAngle := startAngle + angle

		r, g, b, err := s.color.parse()
		if err != nil {
			r, g, b = 100, 100, 100
		}
		c.doc.internal.SetFillColor(r, g, b)

		c.doc.internal.MoveTo(cx, cy)
		c.doc.internal.ArcTo(cx, cy, radius, radius, 0, startAngle, endAngle)
		c.doc.internal.LineTo(cx, cy)
		c.doc.internal.DrawPath("F")

		// Label (Radial)
		midAngle := startAngle + angle/2
		midRad := midAngle * math.Pi / 180

		// Using standard trigonometry assuming standard orientation
		// Adjust signs if necessary after visual inspection
		tx := cx + (radius * 0.7) * math.Cos(midRad)
		ty := cy - (radius * 0.7) * math.Sin(midRad)

		c.doc.internal.SetTextColor(255, 255, 255)
		c.doc.internal.SetFont(c.doc.fontFamily, "B", 10)
		txt := s.label
		if len(txt) > 0 {
			wTxt := c.doc.internal.GetStringWidth(txt)
			c.doc.internal.Text(tx-wTxt/2, ty+3, txt)
		}

		startAngle += angle
	}

	c.doc.internal.SetY(y + c.height + 20)
}
