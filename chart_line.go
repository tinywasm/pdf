package pdf

type LineChart struct {
	doc    *Document
	title  string
	width  float64
	height float64
	series []lineSeries
}

type lineSeries struct {
	name  string
	data  []float64
	color Color
	width float64
}

func (c *LineChart) Title(t string) *LineChart {
	c.title = t
	return c
}

func (c *LineChart) Height(h float64) *LineChart {
	c.height = h
	return c
}

func (c *LineChart) Width(w float64) *LineChart {
	c.width = w
	return c
}

func (c *LineChart) AddSeries(name string, data []float64, col Color) *LineChart {
	c.series = append(c.series, lineSeries{
		name:  name,
		data:  data,
		color: col,
		width: 0.5,
	})
	return c
}

func (c *LineChart) Draw() {
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

	// Calculate Max Y
	maxVal := 0.0
	maxPoints := 0
	for _, s := range c.series {
		if len(s.data) > maxPoints {
			maxPoints = len(s.data)
		}
		for _, v := range s.data {
			if v > maxVal {
				maxVal = v
			}
		}
	}
	if maxVal == 0 {
		maxVal = 1
	}
	if maxPoints < 2 {
		return // Not enough data to draw lines
	}

	plotH := c.height - 20
	scaleY := plotH / maxVal
	stepX := (c.width - 20) / float64(maxPoints-1)

	// Draw Axes
	c.doc.internal.SetDrawColor(0, 0, 0)
	c.doc.internal.SetLineWidth(0.2)
	c.doc.internal.Line(x, y, x, y+c.height) // Y
	c.doc.internal.Line(x, y+c.height, x+c.width, y+c.height) // X

	// Draw Series
	for _, s := range c.series {
		c.doc.internal.SetDrawColor(s.color.R, s.color.G, s.color.B)
		c.doc.internal.SetLineWidth(s.width)
		c.doc.internal.SetFillColor(s.color.R, s.color.G, s.color.B)

		for i := 0; i < len(s.data)-1; i++ {
			x1 := x + 10 + float64(i)*stepX
			y1 := y + c.height - (s.data[i] * scaleY)
			x2 := x + 10 + float64(i+1)*stepX
			y2 := y + c.height - (s.data[i+1] * scaleY)

			c.doc.internal.Line(x1, y1, x2, y2)
			// Dot
			c.doc.internal.Circle(x1, y1, s.width*2, "F")
		}
		// Last dot
		if len(s.data) > 0 {
			i := len(s.data) - 1
			x1 := x + 10 + float64(i)*stepX
			y1 := y + c.height - (s.data[i] * scaleY)
			c.doc.internal.Circle(x1, y1, s.width*2, "F")
		}
	}

	c.doc.internal.SetY(y + c.height + 20)
}
