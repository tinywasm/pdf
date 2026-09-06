package fpdf

import (
	"bytes"
	"math"

	. "webtyp.com/fmt"
)

// GetXY returns the abscissa and ordinate of the current position.
//
// Note: the value returned for the abscissa will be affected by the current
// cell margin. To account for this, you may need to either add the value
// returned by GetCellMargin() to it or call SetCellMargin(0) to remove the
// cell margin.
func (f *Fpdf) GetXY() (float64, float64) {
	return f.x, f.y
}

// GetX returns the abscissa of the current position.
//
// Note: the value returned will be affected by the current cell margin. To
// account for this, you may need to either add the value returned by
// GetCellMargin() to it or call SetCellMargin(0) to remove the cell margin.
func (f *Fpdf) GetX() float64 {
	return f.x
}

// SetX defines the abscissa of the current position. If the passed value is
// negative, it is relative to the right of the page.
func (f *Fpdf) SetX(x float64) {
	if x >= 0 {
		f.x = x
	} else {
		f.x = f.w + x
	}
}

// GetY returns the ordinate of the current position.
func (f *Fpdf) GetY() float64 {
	return f.y
}

// SetY moves the current abscissa back to the left margin and sets the
// ordinate. If the passed value is negative, it is relative to the bottom of
// the page.
func (f *Fpdf) SetY(y float64) {
	// dbg("SetY x %.2f, lMargin %.2f", f.x, f.lMargin)
	f.x = f.lMargin
	if y >= 0 {
		f.y = y
	} else {
		f.y = f.h + y
	}
}

// SetXY defines the abscissa and ordinate of the current position. If the
// passed values are negative, they are relative respectively to the right and
// bottom of the page.
func (f *Fpdf) SetXY(x, y float64) {
	f.SetY(y)
	f.SetX(x)
}

// SetDrawColor defines the color used for all drawing operations (lines,
// rectangles and cell borders). It is expressed in RGB components (0 - 255).
// The method can be called before the first page is created. The value is
// retained from page to page.
func (f *Fpdf) SetDrawColor(r, g, b int) {
	f.setDrawColor(r, g, b)
}

func (f *Fpdf) setDrawColor(r, g, b int) {
	f.color.draw = f.rgbColorValue(r, g, b, "G", "RG")
	if f.page > 0 {
		f.out(f.color.draw.str)
	}
}

// GetDrawColor returns the most recently set draw color as RGB components (0 -
// 255). This will not be the current value if a draw color of some other type
// (for example, spot) has been more recently set.
func (f *Fpdf) GetDrawColor() (int, int, int) {
	return f.color.draw.ir, f.color.draw.ig, f.color.draw.ib
}

// SetFillColor defines the color used for all filling operations (filled
// rectangles and cell backgrounds). It is expressed in RGB components (0
// -255). The method can be called before the first page is created and the
// value is retained from page to page.
func (f *Fpdf) SetFillColor(r, g, b int) {
	f.setFillColor(r, g, b)
}

func (f *Fpdf) setFillColor(r, g, b int) {
	f.color.fill = f.rgbColorValue(r, g, b, "g", "rg")
	f.colorFlag = f.color.fill.str != f.color.text.str
	if f.page > 0 {
		f.out(f.color.fill.str)
	}
}

// GetFillColor returns the most recently set fill color as RGB components (0 -
// 255). This will not be the current value if a fill color of some other type
// (for example, spot) has been more recently set.
func (f *Fpdf) GetFillColor() (int, int, int) {
	return f.color.fill.ir, f.color.fill.ig, f.color.fill.ib
}

// SetTextColor defines the color used for text. It is expressed in RGB
// components (0 - 255). The method can be called before the first page is
// created. The value is retained from page to page.
func (f *Fpdf) SetTextColor(r, g, b int) {
	f.setTextColor(r, g, b)
}

func (f *Fpdf) setTextColor(r, g, b int) {
	f.color.text = f.rgbColorValue(r, g, b, "g", "rg")
	f.colorFlag = f.color.fill.str != f.color.text.str
}

// GetTextColor returns the most recently set text color as RGB components (0 -
// 255). This will not be the current value if a text color of some other type
// (for example, spot) has been more recently set.
func (f *Fpdf) GetTextColor() (int, int, int) {
	return f.color.text.ir, f.color.text.ig, f.color.text.ib
}

// GetStringWidth returns the length of a string in user units. A font must be
// currently selected.
func (f *Fpdf) GetStringWidth(s string) float64 {
	if f.err != nil {
		return 0
	}
	w := f.GetStringSymbolWidth(s)
	return float64(w) * f.fontSize / 1000
}

// GetStringSymbolWidth returns the length of a string in glyf units. A font must be
// currently selected.
func (f *Fpdf) GetStringSymbolWidth(s string) int {
	if f.err != nil {
		return 0
	}
	w := 0
	if f.isCurrentUTF8 {
		for _, char := range s {
			intChar := int(char)
			if len(f.currentFont.Cw) >= intChar && f.currentFont.Cw[intChar] > 0 {
				if f.currentFont.Cw[intChar] != 65535 {
					w += f.currentFont.Cw[intChar]
				}
			} else if f.currentFont.Desc.MissingWidth != 0 {
				w += f.currentFont.Desc.MissingWidth
			} else {
				w += 500
			}
		}
	} else {
		for _, ch := range []byte(s) {
			if ch == 0 {
				break
			}
			w += f.currentFont.Cw[ch]
		}
	}
	return w
}

// SetLineWidth defines the line width. By default, the value equals 0.2 mm.
// The method can be called before the first page is created. The value is
// retained from page to page.
func (f *Fpdf) SetLineWidth(width float64) {
	f.setLineWidth(width)
}

func (f *Fpdf) setLineWidth(width float64) {
	f.lineWidth = width
	if f.page > 0 {
		f.out(f.fmtF64(width*f.k, 2) + " w")
	}
}

// GetLineWidth returns the current line thickness.
func (f *Fpdf) GetLineWidth() float64 {
	return f.lineWidth
}

// GetLineCapStyle returns the current line cap style.
func (f *Fpdf) GetLineCapStyle() string {
	switch f.capStyle {
	case 1:
		return "round"
	case 2:
		return "square"
	default:
		return "butt"
	}
}

// SetLineCapStyle defines the line cap style. styleStr should be "butt",
// "round" or "square". A square style projects from the end of the line. The
// method can be called before the first page is created. The value is
// retained from page to page.
func (f *Fpdf) SetLineCapStyle(styleStr string) {
	var capStyle int
	switch styleStr {
	case "round":
		capStyle = 1
	case "square":
		capStyle = 2
	default:
		capStyle = 0
	}
	f.capStyle = capStyle
	if f.page > 0 {
		f.outf("%d J", f.capStyle)
	}
}

// GetLineJoinStyle returns the current line join style.
func (f *Fpdf) GetLineJoinStyle() string {
	switch f.joinStyle {
	case 1:
		return "round"
	case 2:
		return "bevel"
	default:
		return "miter"
	}
}

// SetLineJoinStyle defines the line cap style. styleStr should be "miter",
// "round" or "bevel". The method can be called before the first page
// is created. The value is retained from page to page.
func (f *Fpdf) SetLineJoinStyle(styleStr string) {
	var joinStyle int
	switch styleStr {
	case "round":
		joinStyle = 1
	case "bevel":
		joinStyle = 2
	default:
		joinStyle = 0
	}
	f.joinStyle = joinStyle
	if f.page > 0 {
		f.outf("%d j", f.joinStyle)
	}
}

// Path Drawing

// MoveTo moves the stylus to (x, y) without drawing the path from the
// previous point. Paths must start with a MoveTo to set the original
// stylus location or the result is undefined.
//
// Create a "path" by moving a virtual stylus around the page (with
// MoveTo, LineTo, CurveTo, CurveBezierCubicTo, ArcTo & ClosePath)
// then draw it or  fill it in (with DrawPath). The main advantage of
// using the path drawing routines rather than multiple Fpdf.Line is
// that PDF creates nice line joins at the angles, rather than just
// overlaying the lines.
func (f *Fpdf) MoveTo(x, y float64) {
	f.point(x, y)
	f.x, f.y = x, y
}

// LineTo creates a line from the current stylus location to (x, y), which
// becomes the new stylus location. Note that this only creates the line in
// the path; it does not actually draw the line on the page.
//
// The MoveTo() example demonstrates this method.
func (f *Fpdf) LineTo(x, y float64) {
	// f.outf("%.2f %.2f l", x*f.k, (f.h-y)*f.k)
	const prec = 2
	f.putF64(x*f.k, prec)
	f.put(" ")

	f.putF64((f.h-y)*f.k, prec)
	f.put(" l\n")

	f.x, f.y = x, y
}

// CurveTo creates a single-segment quadratic Bézier curve. The curve starts at
// the current stylus location and ends at the point (x, y). The control point
// (cx, cy) specifies the curvature. At the start point, the curve is tangent
// to the straight line between the current stylus location and the control
// point. At the end point, the curve is tangent to the straight line between
// the end point and the control point.
//
// The MoveTo() example demonstrates this method.
func (f *Fpdf) CurveTo(cx, cy, x, y float64) {
	// f.outf("%.5f %.5f %.5f %.5f v", cx*f.k, (f.h-cy)*f.k, x*f.k, (f.h-y)*f.k)
	const prec = 5
	f.putF64(cx*f.k, prec)
	f.put(" ")
	f.putF64((f.h-cy)*f.k, prec)
	f.put(" ")
	f.putF64(x*f.k, prec)
	f.put(" ")
	f.putF64((f.h-y)*f.k, prec)
	f.put(" v\n")
	f.x, f.y = x, y
}

// CurveBezierCubicTo creates a single-segment cubic Bézier curve. The curve
// starts at the current stylus location and ends at the point (x, y). The
// control points (cx0, cy0) and (cx1, cy1) specify the curvature. At the
// current stylus, the curve is tangent to the straight line between the
// current stylus location and the control point (cx0, cy0). At the end point,
// the curve is tangent to the straight line between the end point and the
// control point (cx1, cy1).
//
// The MoveTo() example demonstrates this method.
func (f *Fpdf) CurveBezierCubicTo(cx0, cy0, cx1, cy1, x, y float64) {
	f.curve(cx0, cy0, cx1, cy1, x, y)
	f.x, f.y = x, y
}

// ClosePath creates a line from the current location to the last MoveTo point
// (if not the same) and mark the path as closed so the first and last lines
// join nicely.
//
// The MoveTo() example demonstrates this method.
func (f *Fpdf) ClosePath() {
	f.outf("h")
}

// DrawPath actually draws the path on the page.
//
// styleStr can be "F" for filled, "D" for outlined only, or "DF" or "FD" for
// outlined and filled. An empty string will be replaced with "D".
// Path-painting operators as defined in the PDF specification are also
// allowed: "S" (Stroke the path), "s" (Close and stroke the path),
// "f" (fill the path, using the nonzero winding number), "f*"
// (Fill the path, using the even-odd rule), "B" (Fill and then stroke
// the path, using the nonzero winding number rule), "B*" (Fill and
// then stroke the path, using the even-odd rule), "b" (Close, fill,
// and then stroke the path, using the nonzero winding number rule) and
// "b*" (Close, fill, and then stroke the path, using the even-odd
// rule).
// Drawing uses the current draw color, line width, and cap style
// centered on the
// path. Filling uses the current fill color.
//
// The MoveTo() example demonstrates this method.
func (f *Fpdf) DrawPath(styleStr string) {
	f.outf("%s", fillDrawOp(styleStr))
}

// ArcTo draws an elliptical arc centered at point (x, y). rx and ry specify its
// horizontal and vertical radii. If the start of the arc is not at
// the current position, a connecting line will be drawn.
//
// degRotate specifies the angle that the arc will be rotated. degStart and
// degEnd specify the starting and ending angle of the arc. All angles are
// specified in degrees and measured counter-clockwise from the 3 o'clock
// position.
//
// styleStr can be "F" for filled, "D" for outlined only, or "DF" or "FD" for
// outlined and filled. An empty string will be replaced with "D". Drawing uses
// the current draw color, line width, and cap style centered on the arc's
// path. Filling uses the current fill color.
//
// The MoveTo() example demonstrates this method.
func (f *Fpdf) ArcTo(x, y, rx, ry, degRotate, degStart, degEnd float64) {
	f.arc(x, y, rx, ry, degRotate, degStart, degEnd, "", true)
}

func (f *Fpdf) arc(x, y, rx, ry, degRotate, degStart, degEnd float64,
	styleStr string, path bool) {
	x *= f.k
	y = (f.h - y) * f.k
	rx *= f.k
	ry *= f.k
	segments := int(degEnd-degStart) / 60
	if segments < 2 {
		segments = 2
	}
	angleStart := degStart * math.Pi / 180
	angleEnd := degEnd * math.Pi / 180
	angleTotal := angleEnd - angleStart
	dt := angleTotal / float64(segments)
	dtm := dt / 3
	if degRotate != 0 {
		a := -degRotate * math.Pi / 180
		sin, cos := math.Sincos(a)
		//	f.outf("q %.5f %.5f %.5f %.5f %.5f %.5f cm",
		//		math.Cos(a), -1*math.Sin(a),
		//		math.Sin(a), math.Cos(a), x, y)
		const prec = 5
		f.put("q ")
		f.putF64(cos, prec)
		f.put(" ")
		f.putF64(-1*sin, prec)
		f.put(" ")
		f.putF64(sin, prec)
		f.put(" ")
		f.putF64(cos, prec)
		f.put(" ")
		f.putF64(x, prec)
		f.put(" ")
		f.putF64(y, prec)
		f.put(" cm\n")

		x = 0
		y = 0
	}
	t := angleStart
	a0 := x + rx*math.Cos(t)
	b0 := y + ry*math.Sin(t)
	c0 := -rx * math.Sin(t)
	d0 := ry * math.Cos(t)
	sx := a0 / f.k // start point of arc
	sy := f.h - (b0 / f.k)
	if path {
		if f.x != sx || f.y != sy {
			// Draw connecting line to start point
			f.LineTo(sx, sy)
		}
	} else {
		f.point(sx, sy)
	}
	for j := 1; j <= segments; j++ {
		// Draw this bit of the total curve
		t = (float64(j) * dt) + angleStart
		a1 := x + rx*math.Cos(t)
		b1 := y + ry*math.Sin(t)
		c1 := -rx * math.Sin(t)
		d1 := ry * math.Cos(t)
		f.curve((a0+(c0*dtm))/f.k,
			f.h-((b0+(d0*dtm))/f.k),
			(a1-(c1*dtm))/f.k,
			f.h-((b1-(d1*dtm))/f.k),
			a1/f.k,
			f.h-(b1/f.k))
		a0 = a1
		b0 = b1
		c0 = c1
		d0 = d1
		if path {
			f.x = a1 / f.k
			f.y = f.h - (b1 / f.k)
		}
	}
	if !path {
		f.out(fillDrawOp(styleStr))
	}
	if degRotate != 0 {
		f.out("Q")
	}
}

// Text prints a character string. The origin (x, y) is on the left of the
// first character at the baseline. This method permits a string to be placed
// precisely on the page, but it is usually easier to use Cell(), MultiCell()
// or Write() which are the standard methods to print text.
func (f *Fpdf) Text(x, y float64, txtStr string) {
	var txt2 string
	if f.isCurrentUTF8 {
		if f.isRTL {
			txtStr = reverseText(txtStr)
			x -= f.GetStringWidth(txtStr)
		}
		txt2 = f.escape(utf8toutf16(txtStr, false))
		for _, uni := range txtStr {
			f.currentFont.usedRunes[int(uni)] = int(uni)
		}
	} else {
		txt2 = f.escape(txtStr)
	}
	s := sprintf("BT %.2f %.2f Td (%s) Tj ET", x*f.k, (f.h-y)*f.k, txt2)
	if f.underline && txtStr != "" {
		s += " " + f.dounderline(x, y, txtStr)
	}
	if f.strikeout && txtStr != "" {
		s += " " + f.dostrikeout(x, y, txtStr)
	}
	if f.colorFlag {
		s = sprintf("q %s %s Q", f.color.text.str, s)
	}
	f.out(s)
}

// Line draws a line between points (x1, y1) and (x2, y2) using the current
// draw color, line width and cap style.
func (f *Fpdf) Line(x1, y1, x2, y2 float64) {
	// f.outf("%.2f %.2f m %.2f %.2f l S", x1*f.k, (f.h-y1)*f.k, x2*f.k, (f.h-y2)*f.k)
	const prec = 2
	f.putF64(x1*f.k, prec)
	f.put(" ")
	f.putF64((f.h-y1)*f.k, prec)
	f.put(" m ")
	f.putF64(x2*f.k, prec)
	f.put(" ")
	f.putF64((f.h-y2)*f.k, prec)
	f.put(" l S\n")
}

// fillDrawOp corrects path painting operators
func fillDrawOp(styleStr string) (opStr string) {
	switch Convert(styleStr).ToUpper().String() {
	case "", "D":
		// Stroke the path.
		opStr = "S"
	case "F":
		// fill the path, using the nonzero winding number rule
		opStr = "f"
	case "F*":
		// fill the path, using the even-odd rule
		opStr = "f*"
	case "FD", "DF":
		// fill and then stroke the path, using the nonzero winding number rule
		opStr = "B"
	case "FD*", "DF*":
		// fill and then stroke the path, using the even-odd rule
		opStr = "B*"
	default:
		opStr = styleStr
	}
	return
}

// Rect outputs a rectangle of width w and height h with the upper left corner
// positioned at point (x, y).
//
// It can be drawn (border only), filled (with no border) or both. styleStr can
// be "F" for filled, "D" for outlined only, or "DF" or "FD" for outlined and
// filled. An empty string will be replaced with "D". Drawing uses the current
// draw color and line width centered on the rectangle's perimeter. Filling
// uses the current fill color.
func (f *Fpdf) Rect(x, y, w, h float64, styleStr string) {
	// f.outf("%.2f %.2f %.2f %.2f re %s", x*f.k, (f.h-y)*f.k, w*f.k, -h*f.k, fillDrawOp(styleStr))
	const prec = 2
	f.putF64(x*f.k, prec)
	f.put(" ")
	f.putF64((f.h-y)*f.k, prec)
	f.put(" ")
	f.putF64(w*f.k, prec)
	f.put(" ")
	f.putF64(-h*f.k, prec)
	f.put(" re " + fillDrawOp(styleStr) + "\n")
}

// RoundedRect outputs a rectangle of width w and height h with the upper left
// corner positioned at point (x, y). It can be drawn (border only), filled
// (with no border) or both. styleStr can be "F" for filled, "D" for outlined
// only, or "DF" or "FD" for outlined and filled. An empty string will be
// replaced with "D". Drawing uses the current draw color and line width
// centered on the rectangle's perimeter. Filling uses the current fill color.
// The rounded corners of the rectangle are specified by radius r. corners is a
// string that includes "1" to round the upper left corner, "2" to round the
// upper right corner, "3" to round the lower right corner, and "4" to round
// the lower left corner. The RoundedRect example demonstrates this method.
func (f *Fpdf) RoundedRect(x, y, w, h, r float64, corners string, stylestr string) {
	// This routine was adapted by Brigham Thompson from a script by Christophe Prugnaud
	var rTL, rTR, rBR, rBL float64 // zero means no rounded corner
	if Contains(corners, "1") {
		rTL = r
	}
	if Contains(corners, "2") {
		rTR = r
	}
	if Contains(corners, "3") {
		rBR = r
	}
	if Contains(corners, "4") {
		rBL = r
	}
	f.RoundedRectExt(x, y, w, h, rTL, rTR, rBR, rBL, stylestr)
}

// RoundedRectExt behaves the same as RoundedRect() but supports a different
// radius for each corner. A zero radius means squared corner. See
// RoundedRect() for more details. This method is demonstrated in the
// RoundedRect() example.
func (f *Fpdf) RoundedRectExt(x, y, w, h, rTL, rTR, rBR, rBL float64, stylestr string) {
	f.roundedRectPath(x, y, w, h, rTL, rTR, rBR, rBL)
	f.out(fillDrawOp(stylestr))
	f.out("Q")
}

// Circle draws a circle centered on point (x, y) with radius r.
//
// styleStr can be "F" for filled, "D" for outlined only, or "DF" or "FD" for
// outlined and filled. An empty string will be replaced with "D". Drawing uses
// the current draw color and line width centered on the circle's perimeter.
// Filling uses the current fill color.
func (f *Fpdf) Circle(x, y, r float64, styleStr string) {
	f.Ellipse(x, y, r, r, 0, styleStr)
}

// Ellipse draws an ellipse centered at point (x, y). rx and ry specify its
// horizontal and vertical radii.
//
// degRotate specifies the counter-clockwise angle in degrees that the ellipse
// will be rotated.
//
// styleStr can be "F" for filled, "D" for outlined only, or "DF" or "FD" for
// outlined and filled. An empty string will be replaced with "D". Drawing uses
// the current draw color and line width centered on the ellipse's perimeter.
// Filling uses the current fill color.
//
// The Circle() example demonstrates this method.
func (f *Fpdf) Ellipse(x, y, rx, ry, degRotate float64, styleStr string) {
	f.arc(x, y, rx, ry, degRotate, 0, 360, styleStr, false)
}

// Polygon draws a closed figure defined by a series of vertices specified by
// points. The x and y fields of the points use the units established in New().
// The last point in the slice will be implicitly joined to the first to close
// the polygon.
//
// styleStr can be "F" for filled, "D" for outlined only, or "DF" or "FD" for
// outlined and filled. An empty string will be replaced with "D". Drawing uses
// the current draw color and line width centered on the ellipse's perimeter.
// Filling uses the current fill color.
func (f *Fpdf) Polygon(points []PointType, styleStr string) {
	if len(points) > 2 {
		const prec = 5
		for j, pt := range points {
			if j == 0 {
				f.point(pt.X, pt.Y)
			} else {
				// f.outf("%.5f %.5f l ", pt.X*f.k, (f.h-pt.Y)*f.k)
				f.putF64(pt.X*f.k, prec)
				f.put(" ")
				f.putF64((f.h-pt.Y)*f.k, prec)
				f.put(" l \n")
			}
		}
		// f.outf("%.5f %.5f l ", points[0].X*f.k, (f.h-points[0].Y)*f.k)
		f.putF64(points[0].X*f.k, prec)
		f.put(" ")
		f.putF64((f.h-points[0].Y)*f.k, prec)
		f.put(" l \n")
		f.DrawPath(styleStr)
	}
}

// Beziergon draws a closed figure defined by a series of cubic Bézier curve
// segments. The first point in the slice defines the starting point of the
// figure. Each three following points p1, p2, p3 represent a curve segment to
// the point p3 using p1 and p2 as the Bézier control points.
//
// The x and y fields of the points use the units established in New().
//
// styleStr can be "F" for filled, "D" for outlined only, or "DF" or "FD" for
// outlined and filled. An empty string will be replaced with "D". Drawing uses
// the current draw color and line width centered on the ellipse's perimeter.
// Filling uses the current fill color.
func (f *Fpdf) Beziergon(points []PointType, styleStr string) {

	// Thanks, Robert Lillack, for contributing this function.

	if len(points) < 4 {
		return
	}
	f.point(points[0].XY())

	points = points[1:]
	for len(points) >= 3 {
		cx0, cy0 := points[0].XY()
		cx1, cy1 := points[1].XY()
		x1, y1 := points[2].XY()
		f.curve(cx0, cy0, cx1, cy1, x1, y1)
		points = points[3:]
	}

	f.DrawPath(styleStr)
}

// point outputs current point
func (f *Fpdf) point(x, y float64) {
	// f.outf("%.2f %.2f m", x*f.k, (f.h-y)*f.k)
	f.putF64(x*f.k, 2)
	f.put(" ")
	f.putF64((f.h-y)*f.k, 2)
	f.put(" m\n")
}

// curve outputs a single cubic Bézier curve segment from current point
func (f *Fpdf) curve(cx0, cy0, cx1, cy1, x, y float64) {
	// Thanks, Robert Lillack, for straightening this out
	// f.outf("%.5f %.5f %.5f %.5f %.5f %.5f c", cx0*f.k, (f.h-cy0)*f.k, cx1*f.k,
	// 	(f.h-cy1)*f.k, x*f.k, (f.h-y)*f.k)
	const prec = 5
	f.putF64(cx0*f.k, prec)
	f.put(" ")
	f.putF64((f.h-cy0)*f.k, prec)
	f.put(" ")
	f.putF64(cx1*f.k, prec)
	f.put(" ")
	f.putF64((f.h-cy1)*f.k, prec)
	f.put(" ")
	f.putF64(x*f.k, prec)
	f.put(" ")
	f.putF64((f.h-y)*f.k, prec)
	f.put(" c\n")
}

// Curve draws a single-segment quadratic Bézier curve. The curve starts at
// the point (x0, y0) and ends at the point (x1, y1). The control point (cx,
// cy) specifies the curvature. At the start point, the curve is tangent to the
// straight line between the start point and the control point. At the end
// point, the curve is tangent to the straight line between the end point and
// the control point.
//
// styleStr can be "F" for filled, "D" for outlined only, or "DF" or "FD" for
// outlined and filled. An empty string will be replaced with "D". Drawing uses
// the current draw color, line width, and cap style centered on the curve's
// path. Filling uses the current fill color.
//
// The Circle() example demonstrates this method.
func (f *Fpdf) Curve(x0, y0, cx, cy, x1, y1 float64, styleStr string) {
	f.point(x0, y0)
	// f.outf("%.5f %.5f %.5f %.5f v %s", cx*f.k, (f.h-cy)*f.k, x1*f.k, (f.h-y1)*f.k,
	// 	fillDrawOp(styleStr))
	const prec = 5
	f.putF64(cx*f.k, prec)
	f.put(" ")
	f.putF64((f.h-cy)*f.k, prec)
	f.put(" ")
	f.putF64(x1*f.k, prec)
	f.put(" ")
	f.putF64((f.h-y1)*f.k, prec)
	f.put(" v " + fillDrawOp(styleStr) + "\n")
}

// CurveCubic draws a single-segment cubic Bézier curve. This routine performs
// the same function as CurveBezierCubic() but has a nonstandard argument order.
// It is retained to preserve backward compatibility.
func (f *Fpdf) CurveCubic(x0, y0, cx0, cy0, x1, y1, cx1, cy1 float64, styleStr string) {
	// f.point(x0, y0)
	// f.outf("%.5f %.5f %.5f %.5f %.5f %.5f c %s", cx0*f.k, (f.h-cy0)*f.k,
	// cx1*f.k, (f.h-cy1)*f.k, x1*f.k, (f.h-y1)*f.k, fillDrawOp(styleStr))
	f.CurveBezierCubic(x0, y0, cx0, cy0, cx1, cy1, x1, y1, styleStr)
}

// CurveBezierCubic draws a single-segment cubic Bézier curve. The curve starts at
// the point (x0, y0) and ends at the point (x1, y1). The control points (cx0,
// cy0) and (cx1, cy1) specify the curvature. At the start point, the curve is
// tangent to the straight line between the start point and the control point
// (cx0, cy0). At the end point, the curve is tangent to the straight line
// between the end point and the control point (cx1, cy1).
//
// styleStr can be "F" for filled, "D" for outlined only, or "DF" or "FD" for
// outlined and filled. An empty string will be replaced with "D". Drawing uses
// the current draw color, line width, and cap style centered on the curve's
// path. Filling uses the current fill color.
//
// This routine performs the same function as CurveCubic() but uses standard
// argument order.
//
// The Circle() example demonstrates this method.
func (f *Fpdf) CurveBezierCubic(x0, y0, cx0, cy0, cx1, cy1, x1, y1 float64, styleStr string) {
	f.point(x0, y0)
	//	f.outf("%.5f %.5f %.5f %.5f %.5f %.5f c %s", cx0*f.k, (f.h-cy0)*f.k,
	//		cx1*f.k, (f.h-cy1)*f.k, x1*f.k, (f.h-y1)*f.k, fillDrawOp(styleStr))
	const prec = 5
	f.putF64(cx0*f.k, prec)
	f.put(" ")
	f.putF64((f.h-cy0)*f.k, prec)
	f.put(" ")
	f.putF64(cx1*f.k, prec)
	f.put(" ")
	f.putF64((f.h-cy1)*f.k, prec)
	f.put(" ")
	f.putF64(x1*f.k, prec)
	f.put(" ")
	f.putF64((f.h-y1)*f.k, prec)
	f.put(" c " + fillDrawOp(styleStr) + "\n")
}

// Arc draws an elliptical arc centered at point (x, y). rx and ry specify its
// horizontal and vertical radii.
//
// degRotate specifies the angle that the arc will be rotated. degStart and
// degEnd specify the starting and ending angle of the arc. All angles are
// specified in degrees and measured counter-clockwise from the 3 o'clock
// position.
//
// styleStr can be "F" for filled, "D" for outlined only, or "DF" or "FD" for
// outlined and filled. An empty string will be replaced with "D". Drawing uses
// the current draw color, line width, and cap style centered on the arc's
// path. Filling uses the current fill color.
//
// The Circle() example demonstrates this method.
func (f *Fpdf) Arc(x, y, rx, ry, degRotate, degStart, degEnd float64, styleStr string) {
	f.arc(x, y, rx, ry, degRotate, degStart, degEnd, styleStr, false)
}

// GetAlpha returns the alpha blending channel, which consists of the
// alpha transparency value and the blend mode. See SetAlpha for more
// details.
func (f *Fpdf) GetAlpha() (alpha float64, blendModeStr string) {
	return f.alpha, f.blendMode
}

// SetAlpha sets the alpha blending channel. The blending effect applies to
// text, drawings and images.
//
// alpha must be a value between 0.0 (fully transparent) to 1.0 (fully opaque).
// Values outside of this range result in an error.
//
// blendModeStr must be one of "Normal", "Multiply", "Screen", "Overlay",
// "Darken", "Lighten", "ColorDodge", "ColorBurn","HardLight", "SoftLight",
// "Difference", "Exclusion", "Hue", "Saturation", "Color", or "Luminosity". An
// empty string is replaced with "Normal".
//
// To reset normal rendering after applying a blending mode, call this method
// with alpha set to 1.0 and blendModeStr set to "Normal".
func (f *Fpdf) SetAlpha(alpha float64, blendModeStr string) {
	if f.err != nil {
		return
	}
	var bl blendModeType
	switch blendModeStr {
	case "Normal", "Multiply", "Screen", "Overlay",
		"Darken", "Lighten", "ColorDodge", "ColorBurn", "HardLight", "SoftLight",
		"Difference", "Exclusion", "Hue", "Saturation", "Color", "Luminosity":
		bl.modeStr = blendModeStr
	case "":
		bl.modeStr = "Normal"
	default:
		f.err = Errf("unrecognized blend mode \"%s\"", blendModeStr)
		return
	}
	if alpha < 0.0 || alpha > 1.0 {
		f.err = Errf("alpha value (0.0 - 1.0) is out of range: %.3f", alpha)
		return
	}
	f.alpha = alpha
	f.blendMode = blendModeStr
	alphaStr := sprintf("%.3f", alpha)
	keyStr := sprintf("%s %s", alphaStr, blendModeStr)
	pos, ok := f.blendMap[keyStr]
	if !ok {
		pos = len(f.blendList) // at least 1
		f.blendList = append(f.blendList, blendModeType{alphaStr, alphaStr, blendModeStr, 0})
		f.blendMap[keyStr] = pos
	}
	if len(f.blendMap) > 0 && f.pdfVersion < pdfVers1_4 {
		f.pdfVersion = pdfVers1_4
	}
	f.outf("/GS%d gs", pos)
}

func (f *Fpdf) gradientClipStart(x, y, w, h float64) {
	{
		const prec = 2
		// Save current graphic state and set clipping area
		// f.outf("q %.2f %.2f %.2f %.2f re W n", x*f.k, (f.h-y)*f.k, w*f.k, -h*f.k)
		f.put("q ")
		f.putF64(x*f.k, prec)
		f.put(" ")
		f.putF64((f.h-y)*f.k, prec)
		f.put(" ")
		f.putF64(w*f.k, prec)
		f.put(" ")
		f.putF64(-h*f.k, prec)
		f.put(" re W n\n")
	}
	{
		const prec = 5
		// Set up transformation matrix for gradient
		// f.outf("%.5f 0 0 %.5f %.5f %.5f cm", w*f.k, h*f.k, x*f.k, (f.h-(y+h))*f.k)
		f.putF64(w*f.k, prec)
		f.put(" 0 0 ")
		f.putF64(h*f.k, prec)
		f.put(" ")
		f.putF64(x*f.k, prec)
		f.put(" ")
		f.putF64((f.h-(y+h))*f.k, prec)
		f.put(" cm\n")
	}
}

func (f *Fpdf) gradientClipEnd() {
	// Restore previous graphic state
	f.out("Q")
}

func (f *Fpdf) gradient(tp, r1, g1, b1, r2, g2, b2 int, x1, y1, x2, y2, r float64) {
	pos := len(f.gradientList)
	clr1 := f.rgbColorValue(r1, g1, b1, "", "")
	clr2 := f.rgbColorValue(r2, g2, b2, "", "")
	f.gradientList = append(f.gradientList, gradientType{tp, clr1.str, clr2.str,
		x1, y1, x2, y2, r, 0})
	f.outf("/Sh%d sh", pos)
}

// LinearGradient draws a rectangular area with a blending of one color to
// another. The rectangle is of width w and height h. Its upper left corner is
// positioned at point (x, y).
//
// Each color is specified with three component values, one each for red, green
// and blue. The values range from 0 to 255. The first color is specified by
// (r1, g1, b1) and the second color by (r2, g2, b2).
//
// The blending is controlled with a gradient vector that uses normalized
// coordinates in which the lower left corner is position (0, 0) and the upper
// right corner is (1, 1). The vector's origin and destination are specified by
// the points (x1, y1) and (x2, y2). In a linear gradient, blending occurs
// perpendicularly to the vector. The vector does not necessarily need to be
// anchored on the rectangle edge. Color 1 is used up to the origin of the
// vector and color 2 is used beyond the vector's end point. Between the points
// the colors are gradually blended.
func (f *Fpdf) LinearGradient(x, y, w, h float64, r1, g1, b1, r2, g2, b2 int, x1, y1, x2, y2 float64) {
	f.gradientClipStart(x, y, w, h)
	f.gradient(2, r1, g1, b1, r2, g2, b2, x1, y1, x2, y2, 0)
	f.gradientClipEnd()
}

// RadialGradient draws a rectangular area with a blending of one color to
// another. The rectangle is of width w and height h. Its upper left corner is
// positioned at point (x, y).
//
// Each color is specified with three component values, one each for red, green
// and blue. The values range from 0 to 255. The first color is specified by
// (r1, g1, b1) and the second color by (r2, g2, b2).
//
// The blending is controlled with a point and a circle, both specified with
// normalized coordinates in which the lower left corner of the rendered
// rectangle is position (0, 0) and the upper right corner is (1, 1). Color 1
// begins at the origin point specified by (x1, y1). Color 2 begins at the
// circle specified by the center point (x2, y2) and radius r. Colors are
// gradually blended from the origin to the circle. The origin and the circle's
// center do not necessarily have to coincide, but the origin must be within
// the circle to avoid rendering problems.
//
// The LinearGradient() example demonstrates this method.
func (f *Fpdf) RadialGradient(x, y, w, h float64, r1, g1, b1, r2, g2, b2 int, x1, y1, x2, y2, r float64) {
	f.gradientClipStart(x, y, w, h)
	f.gradient(3, r1, g1, b1, r2, g2, b2, x1, y1, x2, y2, r)
	f.gradientClipEnd()
}

// ClipRect begins a rectangular clipping operation. The rectangle is of width
// w and height h. Its upper left corner is positioned at point (x, y). outline
// is true to draw a border with the current draw color and line width centered
// on the rectangle's perimeter. Only the outer half of the border will be
// shown. After calling this method, all rendering operations (for example,
// Image(), LinearGradient(), etc) will be clipped by the specified rectangle.
// Call ClipEnd() to restore unclipped operations.
//
// This ClipText() example demonstrates this method.
func (f *Fpdf) ClipRect(x, y, w, h float64, outline bool) {
	f.clipNest++
	// f.outf("q %.2f %.2f %.2f %.2f re W %s", x*f.k, (f.h-y)*f.k, w*f.k, -h*f.k, strIf(outline, "S", "n"))
	const prec = 2
	f.put("q ")
	f.putF64(x*f.k, prec)
	f.put(" ")
	f.putF64((f.h-y)*f.k, prec)
	f.put(" ")
	f.putF64(w*f.k, prec)
	f.put(" ")
	f.putF64(-h*f.k, prec)
	f.put(" re W " + strIf(outline, "S", "n") + "\n")
}

// ClipText begins a clipping operation in which rendering is confined to the
// character string specified by txtStr. The origin (x, y) is on the left of
// the first character at the baseline. The current font is used. outline is
// true to draw a border with the current draw color and line width centered on
// the perimeters of the text characters. Only the outer half of the border
// will be shown. After calling this method, all rendering operations (for
// example, Image(), LinearGradient(), etc) will be clipped. Call ClipEnd() to
// restore unclipped operations.
func (f *Fpdf) ClipText(x, y float64, txtStr string, outline bool) {
	f.clipNest++
	// f.outf("q BT %.5f %.5f Td %d Tr (%s) Tj ET", x*f.k, (f.h-y)*f.k, intIf(outline, 5, 7), f.escape(txtStr))
	const prec = 5
	f.put("q BT ")
	f.putF64(x*f.k, prec)
	f.put(" ")
	f.putF64((f.h-y)*f.k, prec)
	f.put(" Td ")
	f.putInt(intIf(outline, 5, 7))
	f.put(" Tr (")
	f.put(f.escape(txtStr))
	f.put(") Tj ET\n")
}

func (f *Fpdf) clipArc(x1, y1, x2, y2, x3, y3 float64) {
	h := f.h
	// f.outf("%.5f %.5f %.5f %.5f %.5f %.5f c ", x1*f.k, (h-y1)*f.k,
	// 	x2*f.k, (h-y2)*f.k, x3*f.k, (h-y3)*f.k)
	const prec = 5
	f.putF64(x1*f.k, prec)
	f.put(" ")
	f.putF64((h-y1)*f.k, prec)
	f.put(" ")
	f.putF64(x2*f.k, prec)
	f.put(" ")
	f.putF64((h-y2)*f.k, prec)
	f.put(" ")
	f.putF64(x3*f.k, prec)
	f.put(" ")
	f.putF64((h-y3)*f.k, prec)
	f.put(" c \n")
}

// ClipRoundedRect begins a rectangular clipping operation. The rectangle is of
// width w and height h. Its upper left corner is positioned at point (x, y).
// The rounded corners of the rectangle are specified by radius r. outline is
// true to draw a border with the current draw color and line width centered on
// the rectangle's perimeter. Only the outer half of the border will be shown.
// After calling this method, all rendering operations (for example, Image(),
// LinearGradient(), etc) will be clipped by the specified rectangle. Call
// ClipEnd() to restore unclipped operations.
//
// This ClipText() example demonstrates this method.
func (f *Fpdf) ClipRoundedRect(x, y, w, h, r float64, outline bool) {
	f.ClipRoundedRectExt(x, y, w, h, r, r, r, r, outline)
}

// ClipRoundedRectExt behaves the same as ClipRoundedRect() but supports a
// different radius for each corner, given by rTL (top-left), rTR (top-right)
// rBR (bottom-right), rBL (bottom-left). See ClipRoundedRect() for more
// details. This method is demonstrated in the ClipText() example.
func (f *Fpdf) ClipRoundedRectExt(x, y, w, h, rTL, rTR, rBR, rBL float64, outline bool) {
	f.clipNest++
	f.roundedRectPath(x, y, w, h, rTL, rTR, rBR, rBL)
	f.outf(" W %s", strIf(outline, "S", "n"))
}

// add a rectangle path with rounded corners.
// routine shared by RoundedRect() and ClipRoundedRect(), which add the
// drawing operation
func (f *Fpdf) roundedRectPath(x, y, w, h, rTL, rTR, rBR, rBL float64) {
	k := f.k
	hp := f.h
	myArc := (4.0 / 3.0) * (math.Sqrt2 - 1.0)
	// f.outf("q %.5f %.5f m", (x+rTL)*k, (hp-y)*k)
	const prec = 5
	f.put("q ")
	f.putF64((x+rTL)*k, prec)
	f.put(" ")
	f.putF64((hp-y)*k, prec)
	f.put(" m\n")
	xc := x + w - rTR
	yc := y + rTR
	// f.outf("%.5f %.5f l", xc*k, (hp-y)*k)
	f.putF64(xc*k, prec)
	f.put(" ")
	f.putF64((hp-y)*k, prec)
	f.put(" l\n")
	if rTR != 0 {
		f.clipArc(xc+rTR*myArc, yc-rTR, xc+rTR, yc-rTR*myArc, xc+rTR, yc)
	}
	xc = x + w - rBR
	yc = y + h - rBR
	// f.outf("%.5f %.5f l", (x+w)*k, (hp-yc)*k)
	f.putF64((x+w)*k, prec)
	f.put(" ")
	f.putF64((hp-yc)*k, prec)
	f.put(" l\n")
	if rBR != 0 {
		f.clipArc(xc+rBR, yc+rBR*myArc, xc+rBR*myArc, yc+rBR, xc, yc+rBR)
	}
	xc = x + rBL
	yc = y + h - rBL
	// f.outf("%.5f %.5f l", xc*k, (hp-(y+h))*k)
	f.putF64(xc*k, prec)
	f.put(" ")
	f.putF64((hp-(y+h))*k, prec)
	f.put(" l\n")
	if rBL != 0 {
		f.clipArc(xc-rBL*myArc, yc+rBL, xc-rBL, yc+rBL*myArc, xc-rBL, yc)
	}
	xc = x + rTL
	yc = y + rTL
	// f.outf("%.5f %.5f l", x*k, (hp-yc)*k)
	f.putF64(x*k, prec)
	f.put(" ")
	f.putF64((hp-yc)*k, prec)
	f.put(" l\n")
	if rTL != 0 {
		f.clipArc(xc-rTL, yc-rTL*myArc, xc-rTL*myArc, yc-rTL, xc, yc-rTL)
	}
}

// ClipEllipse begins an elliptical clipping operation. The ellipse is centered
// at (x, y). Its horizontal and vertical radii are specified by rx and ry.
// outline is true to draw a border with the current draw color and line width
// centered on the ellipse's perimeter. Only the outer half of the border will
// be shown. After calling this method, all rendering operations (for example,
// Image(), LinearGradient(), etc) will be clipped by the specified ellipse.
// Call ClipEnd() to restore unclipped operations.
//
// This ClipText() example demonstrates this method.
func (f *Fpdf) ClipEllipse(x, y, rx, ry float64, outline bool) {
	f.clipNest++
	lx := (4.0 / 3.0) * rx * (math.Sqrt2 - 1)
	ly := (4.0 / 3.0) * ry * (math.Sqrt2 - 1)
	k := f.k
	h := f.h
	//	f.outf("q %.5f %.5f m %.5f %.5f %.5f %.5f %.5f %.5f c",
	//		(x+rx)*k, (h-y)*k,
	//		(x+rx)*k, (h-(y-ly))*k,
	//		(x+lx)*k, (h-(y-ry))*k,
	//		x*k, (h-(y-ry))*k)
	const prec = 5
	f.put("q ")
	f.putF64((x+rx)*k, prec)
	f.put(" ")
	f.putF64((h-y)*k, prec)
	f.put(" m ")
	f.putF64((x+rx)*k, prec)
	f.put(" ")
	f.putF64((h-(y-ly))*k, prec)
	f.put(" ")
	f.putF64((x+lx)*k, prec)
	f.put(" ")
	f.putF64((h-(y-ry))*k, prec)
	f.put(" ")
	f.putF64(x*k, prec)
	f.put(" ")
	f.putF64((h-(y-ry))*k, prec)
	f.put(" c\n")

	//	f.outf("%.5f %.5f %.5f %.5f %.5f %.5f c",
	//		(x-lx)*k, (h-(y-ry))*k,
	//		(x-rx)*k, (h-(y-ly))*k,
	//		(x-rx)*k, (h-y)*k)
	f.putF64((x-lx)*k, prec)
	f.put(" ")
	f.putF64((h-(y-ry))*k, prec)
	f.put(" ")
	f.putF64((x-rx)*k, prec)
	f.put(" ")
	f.putF64((h-(y-ly))*k, prec)
	f.put(" ")
	f.putF64((x-rx)*k, prec)
	f.put(" ")
	f.putF64((h-y)*k, prec)
	f.put(" c\n")

	//	f.outf("%.5f %.5f %.5f %.5f %.5f %.5f c",
	//		(x-rx)*k, (h-(y+ly))*k,
	//		(x-lx)*k, (h-(y+ry))*k,
	//		x*k, (h-(y+ry))*k)
	f.putF64((x-rx)*k, prec)
	f.put(" ")
	f.putF64((h-(y+ly))*k, prec)
	f.put(" ")
	f.putF64((x-lx)*k, prec)
	f.put(" ")
	f.putF64((h-(y+ry))*k, prec)
	f.put(" ")
	f.putF64(x*k, prec)
	f.put(" ")
	f.putF64((h-(y+ry))*k, prec)
	f.put(" c\n")

	//	f.outf("%.5f %.5f %.5f %.5f %.5f %.5f c W %s",
	//		(x+lx)*k, (h-(y+ry))*k,
	//		(x+rx)*k, (h-(y+ly))*k,
	//		(x+rx)*k, (h-y)*k,
	//		strIf(outline, "S", "n"))
	f.putF64((x+lx)*k, prec)
	f.put(" ")
	f.putF64((h-(y+ry))*k, prec)
	f.put(" ")
	f.putF64((x+rx)*k, prec)
	f.put(" ")
	f.putF64((h-(y+ly))*k, prec)
	f.put(" ")
	f.putF64((x+rx)*k, prec)
	f.put(" ")
	f.putF64((h-y)*k, prec)
	f.put(" c W " + strIf(outline, "S", "n") + "\n")
}

// ClipCircle begins a circular clipping operation. The circle is centered at
// (x, y) and has radius r. outline is true to draw a border with the current
// draw color and line width centered on the circle's perimeter. Only the outer
// half of the border will be shown. After calling this method, all rendering
// operations (for example, Image(), LinearGradient(), etc) will be clipped by
// the specified circle. Call ClipEnd() to restore unclipped operations.
//
// The ClipText() example demonstrates this method.
func (f *Fpdf) ClipCircle(x, y, r float64, outline bool) {
	f.ClipEllipse(x, y, r, r, outline)
}

// ClipPolygon begins a clipping operation within a polygon. The figure is
// defined by a series of vertices specified by points. The x and y fields of
// the points use the units established in New(). The last point in the slice
// will be implicitly joined to the first to close the polygon. outline is true
// to draw a border with the current draw color and line width centered on the
// polygon's perimeter. Only the outer half of the border will be shown. After
// calling this method, all rendering operations (for example, Image(),
// LinearGradient(), etc) will be clipped by the specified polygon. Call
// ClipEnd() to restore unclipped operations.
//
// The ClipText() example demonstrates this method.
func (f *Fpdf) ClipPolygon(points []PointType, outline bool) {
	f.clipNest++
	var s fmtBuffer
	h := f.h
	k := f.k
	s.printf("q ")
	for j, pt := range points {
		s.printf("%.5f %.5f %s ", pt.X*k, (h-pt.Y)*k, strIf(j == 0, "m", "l"))
	}
	s.printf("h W %s", strIf(outline, "S", "n"))
	f.out(s.String())
}

// ClipEnd ends a clipping operation that was started with a call to
// ClipRect(), ClipRoundedRect(), ClipText(), ClipEllipse(), ClipCircle() or
// ClipPolygon(). Clipping operations can be nested. The document cannot be
// successfully output while a clipping operation is active.
//
// The ClipText() example demonstrates this method.
func (f *Fpdf) ClipEnd() {
	if f.err == nil {
		if f.clipNest > 0 {
			f.clipNest--
			f.out("Q")
		} else {
			f.err = Errf("error attempting to end clip operation out of sequence")
		}
	}
}

// SetDashPattern sets the dash pattern that is used to draw lines. The
// dashArray elements are numbers that specify the lengths, in units
// established in New(), of alternating dashes and gaps. The dash phase
// specifies the distance into the dash pattern at which to start the dash. The
// dash pattern is retained from page to page. Call this method with an empty
// array to restore solid line drawing.
//
// The Beziergon() example demonstrates this method.
func (f *Fpdf) SetDashPattern(dashArray []float64, dashPhase float64) {
	scaled := make([]float64, len(dashArray))
	for i, value := range dashArray {
		scaled[i] = value * f.k
	}
	dashPhase *= f.k

	f.dashArray = scaled
	f.dashPhase = dashPhase
	if f.page > 0 {
		f.outputDashPattern()
	}

}

func (f *Fpdf) outputDashPattern() {
	var buf bytes.Buffer
	buf.WriteByte('[')
	for i, value := range f.dashArray {
		if i > 0 {
			buf.WriteByte(' ')
		}
		buf.WriteString(Convert(value).Round(2).String())
	}
	buf.WriteString("] ")
	buf.WriteString(Convert(f.dashPhase).Round(2).String())
	buf.WriteString(" d")
	f.outbuf(&buf)
}
