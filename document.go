package pdf

import (
	"bytes"
	"io"
	"strings"

	. "github.com/tinywasm/fmt"
	"github.com/tinywasm/pdf/fpdf"
)

// Document wraps the internal fpdf.Fpdf to provide a fluent API.
type Document struct {
	internal *fpdf.Fpdf
	logger   func(message ...any)

	// Resource registries
	fonts  map[string]string // family -> path
	images map[string]string // name -> path
}

// DefaultFontPath is the default path to the Arial UTF-8 font.
const DefaultFontPath = "fonts/Arial.ttf"

// NewDocument creates a new Document instance with UTF-8 support.
func NewDocument() *Document {
	d := &Document{
		fonts:  make(map[string]string),
		images: make(map[string]string),
	}
	d.initIO() // initializes logger + IO depending on build tag
	d.internal = fpdf.New(
		fpdf.WriteFileFunc(d.writeFile),
		fpdf.ReadFileFunc(d.readFile),
		fpdf.FileSizeFunc(d.fileSize),
	)
	d.internal.SetMargins(20, 20, 20)
	d.internal.SetAutoPageBreak(true, 20)
	d.loadDefaultFont()
	return d
}

// loadDefaultFont loads Arial as UTF-8 font so the default "Arial" supports unicode.
func (d *Document) loadDefaultFont() {
	data, err := d.readFile(DefaultFontPath)
	if err != nil {
		return // fallback to built-in Arial (Latin-1 only)
	}
	d.addUTF8FontAllStyles("Arial", data)
}

// addUTF8FontAllStyles registers a TTF font for all styles (regular, bold, italic, bold-italic).
func (d *Document) addUTF8FontAllStyles(family string, data []byte) {
	for _, style := range []string{"", "B", "I", "BI"} {
		d.internal.AddUTF8FontFromBytes(family, style, data)
	}
}

// SetLog sets the logger function.
func (d *Document) SetLog(fn func(...any)) *Document {
	d.logger = fn
	return d
}

// Log writes a message to the logger.
func (d *Document) Log(message ...any) {
	if d.logger != nil {
		d.logger(message...)
	}
}

// RegisterFont registers a font to be loaded.
// path should be the path to the .ttf file.
func (d *Document) RegisterFont(family, path string) *Document {
	d.fonts[family] = path
	return d
}

// RegisterImage registers an image to be loaded.
func (d *Document) RegisterImage(name, path string) *Document {
	d.images[name] = path
	return d
}

// Load loads all registered resources.
func (d *Document) Load(cb func(error)) {
	for family, path := range d.fonts {
		data, err := d.readFile(path)
		if err != nil {
			cb(err)
			return
		}
		d.addUTF8FontAllStyles(family, data)
	}

	for name, path := range d.images {
		data, err := d.readFile(path)
		if err != nil {
			cb(err)
			return
		}

		ext := ""
		if idx := strings.LastIndex(path, "."); idx != -1 {
			ext = path[idx+1:]
		}

		opt := fpdf.ImageOptions{ImageType: ext, ReadDpi: true}
		d.internal.RegisterImageOptionsReader(name, opt, bytes.NewReader(data))
	}

	cb(nil)
}

// Draw is a placeholder for consistency, though currently operations draw immediately.
func (d *Document) Draw() *Document {
	return d
}

// WritePdf generates the PDF and writes it to the specified path.
func (d *Document) WritePdf(path string) error {
	return d.internal.OutputFileAndClose(path)
}

// OutputTo writes the generated PDF into the provided writer.
func (d *Document) OutputTo(w io.Writer) error {
	return d.internal.Output(w)
}

// --- Base Components ---

// AddText adds a text paragraph.
func (d *Document) AddText(text string) *TextComponent {
	return &TextComponent{
		doc:  d,
		text: text,
	}
}

// AddHeader1 adds a level 1 header.
func (d *Document) AddHeader1(text string) *Document {
	d.internal.SetFont("Arial", "B", 16)
	d.internal.CellFormat(0, 8, text, "", 1, "L", false, 0, "")
	d.internal.Ln(3)
	return d
}

// AddHeader2 adds a level 2 header.
func (d *Document) AddHeader2(text string) *Document {
	d.internal.SetFont("Arial", "B", 12)
	d.internal.CellFormat(0, 7, text, "", 1, "L", false, 0, "")
	d.internal.Ln(2)
	return d
}

// AddHeader3 adds a level 3 header.
func (d *Document) AddHeader3(text string) *Document {
	d.internal.SetFont("Arial", "B", 10)
	d.internal.CellFormat(0, 6, text, "", 1, "L", false, 0, "")
	d.internal.Ln(1)
	return d
}

// SpaceBefore adds vertical space.
func (d *Document) SpaceBefore(u float64) *Document {
	d.internal.Ln(u)
	return d
}

// AddPage adds a new page.
func (d *Document) AddPage() *Document {
	d.internal.AddPage()
	return d
}

// AddSeparator adds a horizontal line.
func (d *Document) AddSeparator() *Document {
	x := d.internal.GetX()
	y := d.internal.GetY()
	w, _ := d.internal.GetPageSize()
	lMargin, _, rMargin, _ := d.internal.GetMargins()
	width := w - lMargin - rMargin

	d.internal.Line(x, y+2, x+width, y+2)
	d.internal.Ln(5)
	return d
}

// AddImage adds an image by name (must be registered/loaded).
func (d *Document) AddImage(name string) *ImageComponent {
	return &ImageComponent{
		doc:  d,
		name: name,
	}
}

// DrawImageAt places an image at an absolute (x, y) position with given width (height auto).
func (d *Document) DrawImageAt(name string, x, y, width float64) *Document {
	d.internal.Image(name, x, y, width, 0, false, "", 0, "")
	return d
}

// SetPosition moves the cursor to an absolute (x, y) position.
func (d *Document) SetPosition(x, y float64) *Document {
	d.internal.SetXY(x, y)
	return d
}

// SetCursorY moves the cursor to an absolute Y position.
func (d *Document) SetCursorY(y float64) *Document {
	d.internal.SetY(y)
	return d
}

// GetCursorY returns the current Y position.
func (d *Document) GetCursorY() float64 {
	return d.internal.GetY()
}

// DrawFilledRect draws a filled rectangle at (x, y) with given width and height.
func (d *Document) DrawFilledRect(x, y, w, h float64, r, g, b int) *Document {
	d.internal.SetFillColor(r, g, b)
	d.internal.Rect(x, y, w, h, "F")
	d.internal.SetFillColor(255, 255, 255)
	return d
}

// SetTextColor sets the text color for subsequent AddText calls.
func (d *Document) SetTextColor(r, g, b int) *Document {
	d.internal.SetTextColor(r, g, b)
	return d
}

// DrawTextAt places a single-line text at absolute (x, y) with given font size.
func (d *Document) DrawTextAt(x, y float64, text, style string, size float64) *Document {
	d.internal.SetXY(x, y)
	d.internal.SetFont("Arial", style, size)
	d.internal.Cell(0, size/2.8, text)
	return d
}

// CellAt draws a cell at absolute (x,y) with given width, height, text, and alignment.
func (d *Document) CellAt(x, y, w, h float64, text, style string, size float64, align string) *Document {
	d.internal.SetXY(x, y)
	d.internal.SetFont("Arial", style, size)
	d.internal.CellFormat(w, h, text, "", 0, align, false, 0, "")
	return d
}

// --- Components Helpers ---

type TextComponent struct {
	doc   *Document
	text  string
	align string
	color [3]int
	bold  bool
	size  float64
}

func (t *TextComponent) Bold() *TextComponent {
	t.bold = true
	return t
}

func (t *TextComponent) AlignRight() *TextComponent {
	t.align = "R"
	return t
}

func (t *TextComponent) AlignCenter() *TextComponent {
	t.align = "C"
	return t
}

func (t *TextComponent) Justify() *TextComponent {
	t.align = "J"
	return t
}

func (t *TextComponent) SetColor(r, g, b int) *TextComponent {
	t.color = [3]int{r, g, b}
	return t
}

func (t *TextComponent) Draw() *Document {
	// Apply styles
	style := ""
	if t.bold {
		style = "B"
	}

	// Default font if not set
	family := t.doc.internal.GetFontFamily()
	if family == "" {
		family = "Arial" // Fallback
	}

	size := t.size
	if size == 0 {
		pt, _ := t.doc.internal.GetFontSize()
		if pt < 1 {
			pt = 10
		}
		size = pt
	}

	t.doc.internal.SetFont(family, style, size)
	t.doc.internal.SetTextColor(t.color[0], t.color[1], t.color[2])

	align := "L"
	if t.align != "" {
		align = t.align
	}

	t.doc.internal.MultiCell(0, 6, t.text, "", align, false)

	// Reset text color to black (optional, but good practice)
	t.doc.internal.SetTextColor(0, 0, 0)

	return t.doc
}

type ImageComponent struct {
	doc    *Document
	name   string
	width  float64
	height float64
	align  string
}

func (i *ImageComponent) Width(w float64) *ImageComponent {
	i.width = w
	return i
}

func (i *ImageComponent) Height(h float64) *ImageComponent {
	i.height = h
	return i
}

func (i *ImageComponent) AlignCenter() *ImageComponent {
	i.align = "C"
	return i
}

func (i *ImageComponent) Draw() *Document {
	// Logic to center image if needed
	x := i.doc.internal.GetX()
	y := i.doc.internal.GetY()

	if i.align == "C" {
		w, _ := i.doc.internal.GetPageSize()
		lMargin, _, rMargin, _ := i.doc.internal.GetMargins()
		pageWidth := w - lMargin - rMargin
		// We need to know image width. If 0, fpdf calculates it.
		// For centering we might need to know it beforehand or let fpdf handle it?
		// fpdf.Image doesn't support alignment directly.
		// We can use GetImageInfo to get dimensions.
		info := i.doc.internal.GetImageInfo(i.name)
		if info != nil {
			// Calculate aspect ratio width if i.width is set
			imgW := i.width
			if imgW == 0 {
				imgW = info.Width() // This returns width in user units (mm/pt)
				if i.height != 0 {
					imgW = i.height * info.Width() / info.Height()
				}
			}

			x = lMargin + (pageWidth-imgW)/2
		}
	}

	i.doc.internal.Image(i.name, x, y, i.width, i.height, false, "", 0, "")
	return i.doc
}

// --- Page Header/Footer ---

type PageHeader struct {
	doc       *Document
	leftText  string
	rightText string
}

func (d *Document) SetPageHeader() *PageHeader {
	ph := &PageHeader{doc: d}
	// Register the callback immediately, but it captures the struct so updates will reflect
	d.internal.SetHeaderFunc(func() {
		d.internal.SetY(10) // Standard header position
		d.internal.SetFont("Arial", "I", 8)
		if ph.leftText != "" {
			d.internal.Cell(0, 10, ph.leftText)
		}
		if ph.rightText != "" {
			// Align right
			// Calculate width? Or use CellFormat with align R?
			// Cell(0) goes to right margin.
			d.internal.CellFormat(0, 10, ph.rightText, "", 0, "R", false, 0, "")
		}
		d.internal.Ln(20) // Space after header
	})
	return ph
}

func (ph *PageHeader) SetLeftText(t string) *PageHeader {
	ph.leftText = t
	return ph
}

func (ph *PageHeader) SetRightText(t string) *PageHeader {
	ph.rightText = t
	return ph
}

type PageFooter struct {
	doc        *Document
	centerText string
	pageTotal  bool
}

func (d *Document) SetPageFooter() *PageFooter {
	pf := &PageFooter{doc: d}
	d.internal.SetFooterFunc(func() {
		d.internal.SetY(-15) // Standard footer position
		d.internal.SetFont("Arial", "I", 8)
		if pf.centerText != "" {
			d.internal.CellFormat(0, 10, pf.centerText, "", 0, "C", false, 0, "")
		}
		if pf.pageTotal {
			// We can use alias for total pages if enabled
			// d.internal.AliasNbPages("") should be called somewhere
			pageStr := Convert(d.internal.PageNo()).String()
			d.internal.CellFormat(0, 10, pageStr+" / {nb}", "", 0, "R", false, 0, "")
		}
	})
	return pf
}

func (pf *PageFooter) SetCenterText(t string) *PageFooter {
	pf.centerText = t
	return pf
}

func (pf *PageFooter) WithPageTotal(align string) *PageFooter {
	pf.pageTotal = true
	pf.doc.internal.AliasNbPages("")
	return pf
}

func (d *Document) SetFont(family string, size float64) *Document {
	d.internal.SetFont(family, "", size)
	return d
}

// --- Styles ---

type Style struct {
	FillColor Color
	TextColor Color
	Font      string // "B", "I", ""
	FontSize  float64
}

type Color struct {
	R, G, B int
}

func ColorRGB(r, g, b int) Color {
	return Color{r, g, b}
}

const (
	FontBold    = "B"
	FontItalic  = "I"
	FontRegular = ""
)
