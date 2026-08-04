package pdf

import (
	"github.com/tinywasm/color"
)

// Canvas defines the drawing surface interface used by external packages
// (such as tinywasm/chart) to render graphics on the PDF document.
type Canvas interface {
	GetX() float64
	GetY() float64
	SetY(y float64)
	GetPageSize() (width, height float64)
	GetMargins() (left, top, right, bottom float64)
	SetLineWidth(width float64)
	SetDrawColor(c color.Color)
	SetFillColor(c color.Color)
	SetTextColor(c color.Color)
	Line(x1, y1, x2, y2 float64)
	Rect(x, y, w, h float64, style string)
	Circle(x, y, r float64, style string)
	ArcTo(x, y, rx, ry, angle, startAngle, endAngle float64)
	DrawPath(style string)
	LineTo(x, y float64)
	MoveTo(x, y float64)
	Text(x, y float64, txt string)
	GetStringWidth(txt string) float64
	CellFormat(w, h float64, txt, border string, ln int, align string, fill bool, link int, linkStr string)
	SetDrawingFont(style string, size float64)
}

// Ensure *Document implements Canvas.
var _ Canvas = (*Document)(nil)

func (d *Document) GetX() float64 {
	return d.internal.GetX()
}

func (d *Document) GetY() float64 {
	return d.internal.GetY()
}

func (d *Document) SetY(y float64) {
	d.internal.SetY(y)
}

func (d *Document) GetPageSize() (width, height float64) {
	return d.internal.GetPageSize()
}

func (d *Document) GetMargins() (left, top, right, bottom float64) {
	return d.internal.GetMargins()
}

func (d *Document) SetLineWidth(width float64) {
	d.internal.SetLineWidth(width)
}

func (d *Document) SetDrawColor(c color.Color) {
	r, g, b, err := c.RGB()
	if err != nil {
		d.addError(err)
		return
	}
	d.internal.SetDrawColor(r, g, b)
}

func (d *Document) SetFillColor(c color.Color) {
	r, g, b, err := c.RGB()
	if err != nil {
		d.addError(err)
		return
	}
	d.internal.SetFillColor(r, g, b)
}

func (d *Document) SetTextColor(c color.Color) {
	r, g, b, err := c.RGB()
	if err != nil {
		d.addError(err)
		return
	}
	d.internal.SetTextColor(r, g, b)
}

func (d *Document) Line(x1, y1, x2, y2 float64) {
	d.internal.Line(x1, y1, x2, y2)
}

func (d *Document) Rect(x, y, w, h float64, style string) {
	d.internal.Rect(x, y, w, h, style)
}

func (d *Document) Circle(x, y, r float64, style string) {
	d.internal.Circle(x, y, r, style)
}

func (d *Document) ArcTo(x, y, rx, ry, angle, startAngle, endAngle float64) {
	d.internal.ArcTo(x, y, rx, ry, angle, startAngle, endAngle)
}

func (d *Document) DrawPath(style string) {
	d.internal.DrawPath(style)
}

func (d *Document) LineTo(x, y float64) {
	d.internal.LineTo(x, y)
}

func (d *Document) MoveTo(x, y float64) {
	d.internal.MoveTo(x, y)
}

func (d *Document) Text(x, y float64, txt string) {
	d.internal.Text(x, y, txt)
}

func (d *Document) GetStringWidth(txt string) float64 {
	return d.internal.GetStringWidth(txt)
}

func (d *Document) CellFormat(w, h float64, txt, border string, ln int, align string, fill bool, link int, linkStr string) {
	d.internal.CellFormat(w, h, txt, border, ln, align, fill, link, linkStr)
}

// SetDrawingFont activates the currently active typeface with the requested
// style and size. This satisfies the drawing interface constraint: the canvas
// does not expose SetFont(familyStr, ...), preventing untyped font family
// strings from being used in drawing contracts.
func (d *Document) SetDrawingFont(style string, size float64) {
	d.internal.SetFont(d.getActiveFontName(), style, size)
}
