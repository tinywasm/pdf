package pdf

// Element is the interface for all layout components.
type Element interface {
	draw(doc *Document, x, y, w float64) (height float64)
	measure(doc *Document, w float64) (width, height float64)
}

// Text element represents a paragraph of text.
type TextElement struct {
	content string
	bold    bool
	italic  bool
	size    float64
	color   Color
	align   string // L, C, R, J
}

func Text(s string) *TextElement {
	return &TextElement{content: s, align: "L"}
}

func (t *TextElement) Bold() *TextElement {
	t.bold = true
	return t
}

func (t *TextElement) Italic() *TextElement {
	t.italic = true
	return t
}

func (t *TextElement) Size(pt float64) *TextElement {
	t.size = pt
	return t
}

func (t *TextElement) Color(c Color) *TextElement {
	t.color = c
	return t
}

func (t *TextElement) AlignLeft() *TextElement {
	t.align = "L"
	return t
}

func (t *TextElement) AlignCenter() *TextElement {
	t.align = "C"
	return t
}

func (t *TextElement) AlignRight() *TextElement {
	t.align = "R"
	return t
}

func (t *TextElement) Justify() *TextElement {
	t.align = "J"
	return t
}

func (t *TextElement) getStyle() string {
	style := ""
	if t.bold {
		style += "B"
	}
	if t.italic {
		style += "I"
	}
	return style
}

func (t *TextElement) getSize(doc *Document) float64 {
	if t.size > 0 {
		return t.size
	}
	return doc.theme.Sizes.Body
}

func (t *TextElement) draw(doc *Document, x, y, w float64) float64 {
	style := t.getStyle()
	size := t.getSize(doc)
	color := t.color
	if color == "" {
		color = doc.theme.Body
	}

	doc.setTextColor(color)
	doc.internal.SetFont(doc.fontFamily, style, size)

	// MultiCell at (x, y)
	doc.setPosition(x, y)

	// We need to calculate height. MultiCell doesn't return it directly.
	// But we can calculate it by splitting lines.
	lines := doc.internal.SplitText(t.content, w)
	lineHeight := size * 0.5 // approximate
	height := float64(len(lines)) * lineHeight

	doc.internal.MultiCell(w, lineHeight, t.content, "", t.align, false)
	return height
}

func (t *TextElement) measure(doc *Document, w float64) (float64, float64) {
	style := t.getStyle()
	size := t.getSize(doc)
	doc.internal.SetFont(doc.fontFamily, style, size)

	if w <= 0 {
		// Infinite width measurement for auto-sizing
		return doc.internal.GetStringWidth(t.content), size * 0.5
	}

	lines := doc.internal.SplitText(t.content, w)
	lineHeight := size * 0.5
	return w, float64(len(lines)) * lineHeight
}

// Image element.
type ImageElement struct {
	name   string
	width  float64
	height float64
	align  string // L, C, R
}

func Image(name string) *ImageElement {
	return &ImageElement{name: name, align: "L"}
}

func (i *ImageElement) Width(mm float64) *ImageElement {
	i.width = mm
	return i
}

func (i *ImageElement) Height(mm float64) *ImageElement {
	i.height = mm
	return i
}

func (i *ImageElement) AlignLeft() *ImageElement {
	i.align = "L"
	return i
}

func (i *ImageElement) AlignCenter() *ImageElement {
	i.align = "C"
	return i
}

func (i *ImageElement) AlignRight() *ImageElement {
	i.align = "R"
	return i
}

func (i *ImageElement) draw(doc *Document, x, y, w float64) float64 {
	imgW := i.width
	imgH := i.height

	info := doc.internal.GetImageInfo(i.name)
	if info != nil {
		if imgW == 0 && imgH == 0 {
			imgW = w // Fit to width by default if not specified?
			// Or keep original size? PLAN says Width(mm) optional.
			// Let's use info if available.
			imgW = info.Width()
			if imgW > w {
				imgW = w
			}
		}
		if imgW == 0 {
			imgW = imgH * info.Width() / info.Height()
		}
		if imgH == 0 {
			imgH = imgW * info.Height() / info.Width()
		}
	} else {
		if imgW == 0 { imgW = w }
		if imgH == 0 { imgH = 10 } // Placeholder
	}

	posX := x
	if i.align == "C" {
		posX = x + (w-imgW)/2
	} else if i.align == "R" {
		posX = x + w - imgW
	}

	doc.drawImageAt(i.name, posX, y, imgW)
	return imgH
}

func (i *ImageElement) measure(doc *Document, w float64) (float64, float64) {
	imgW := i.width
	imgH := i.height
	info := doc.internal.GetImageInfo(i.name)
	if info != nil {
		if imgW == 0 && imgH == 0 {
			imgW = info.Width()
			imgH = info.Height()
		} else if imgW == 0 {
			imgW = imgH * info.Width() / info.Height()
		} else if imgH == 0 {
			imgH = imgW * info.Height() / info.Width()
		}
	} else {
		if imgW == 0 { imgW = 20 }
		if imgH == 0 { imgH = 20 }
	}
	return imgW, imgH
}

// Line element.
type LineElement struct {
	width     float64
	color     Color
	thickness float64
}

func Line() *LineElement {
	return &LineElement{thickness: 0.2}
}

func (l *LineElement) Width(mm float64) *LineElement {
	l.width = mm
	return l
}

func (l *LineElement) Color(c Color) *LineElement {
	l.color = c
	return l
}

func (l *LineElement) Thickness(mm float64) *LineElement {
	l.thickness = mm
	return l
}

func (l *LineElement) draw(doc *Document, x, y, w float64) float64 {
	width := l.width
	if width == 0 {
		width = w
	}
	color := l.color
	if color == "" {
		color = doc.theme.Body
	}

	doc.drawLineH(x, y + l.thickness, width, color, l.thickness)
	return l.thickness * 2 // Some padding
}

func (l *LineElement) measure(doc *Document, w float64) (float64, float64) {
	width := l.width
	if width == 0 {
		width = w
	}
	return width, l.thickness * 2
}

// Cell container.
type CellElement struct {
	children []Element
	padding  [4]float64 // top, right, bottom, left
	bg       Color
	border   struct {
		width float64
		color Color
		sides string // "LTRB"
	}
	span struct {
		cols, rows int
	}
	hAlign string // L, C, R
	vAlign string // T, M, B
}

func Cell(children ...Element) *CellElement {
	return &CellElement{
		children: children,
		hAlign:   "L",
		vAlign:   "T",
	}
}

func (c *CellElement) Padding(top, right, bottom, left float64) *CellElement {
	c.padding = [4]float64{top, right, bottom, left}
	return c
}

func (c *CellElement) Background(color Color) *CellElement {
	c.bg = color
	return c
}

func (c *CellElement) Border(sides string, width float64, color Color) *CellElement {
	c.border.sides = sides
	c.border.width = width
	c.border.color = color
	return c
}

func (c *CellElement) Span(cols, rows int) *CellElement {
	c.span.cols = cols
	c.span.rows = rows
	return c
}

func (c *CellElement) AlignLeft() *CellElement {
	c.hAlign = "L"
	return c
}

func (c *CellElement) AlignCenter() *CellElement {
	c.hAlign = "C"
	return c
}

func (c *CellElement) AlignRight() *CellElement {
	c.hAlign = "R"
	return c
}

func (c *CellElement) AlignTop() *CellElement {
	c.vAlign = "T"
	return c
}

func (c *CellElement) AlignMiddle() *CellElement {
	c.vAlign = "M"
	return c
}

func (c *CellElement) AlignBottom() *CellElement {
	c.vAlign = "B"
	return c
}

func (c *CellElement) draw(doc *Document, x, y, w float64) float64 {
	_, h := c.measure(doc, w)
	return c.drawWithHeight(doc, x, y, w, h)
}

func (c *CellElement) drawWithHeight(doc *Document, x, y, w, h float64) float64 {
	// Background
	if c.bg != "" {
		doc.drawFilledRect(x, y, w, h, c.bg)
	}

	// Border - simple implementation
	if c.border.width > 0 {
		r, g, b, _ := c.border.color.parse()
		doc.internal.SetDrawColor(r, g, b)
		doc.internal.SetLineWidth(c.border.width)
		if c.border.sides == "1" || c.border.sides == "" {
			doc.internal.Rect(x, y, w, h, "D")
		} else {
			// individual sides...
		}
	}

	innerX := x + c.padding[3]
	innerY := y + c.padding[0]
	innerW := w - c.padding[1] - c.padding[3]

	// Vertical alignment
	contentHeight := 0.0
	for _, child := range c.children {
		_, ch := child.measure(doc, innerW)
		contentHeight += ch
	}

	if c.vAlign == "M" {
		innerY += (h - c.padding[0] - c.padding[2] - contentHeight) / 2
	} else if c.vAlign == "B" {
		innerY = y + h - c.padding[2] - contentHeight
	}

	currY := innerY
	for _, child := range c.children {
		ch := child.draw(doc, innerX, currY, innerW)
		currY += ch
	}

	return h
}

func (c *CellElement) measure(doc *Document, w float64) (float64, float64) {
	innerW := w - c.padding[1] - c.padding[3]
	if w <= 0 {
		innerW = 0 // auto
	}

	contentH := 0.0
	maxChildW := 0.0
	for _, child := range c.children {
		cw, ch := child.measure(doc, innerW)
		contentH += ch
		if cw > maxChildW {
			maxChildW = cw
		}
	}

	finalW := w
	if w <= 0 {
		finalW = maxChildW + c.padding[1] + c.padding[3]
	}
	finalH := contentH + c.padding[0] + c.padding[2]

	return finalW, finalH
}
