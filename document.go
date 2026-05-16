package pdf

import (
	"bytes"
	"io"

	. "github.com/tinywasm/fmt"
	"github.com/tinywasm/pdf/fpdf"
)

// fontEntry holds a font registration request.
type fontEntry struct {
	family    string
	style     string // "", "B", "I", "BI"
	path      string
	allStyles bool // when true, register the file for all four styles
}

// Document wraps the internal fpdf.Fpdf to provide a fluent API.
type Document struct {
	internal *fpdf.Fpdf
	logger   func(message ...any)

	// Resource registries
	fonts  []fontEntry
	images []KeyValue // name -> path

	fontFamily string
	theme      Theme
	err        error
}

// DefaultFontPath is the default path to the Arial UTF-8 font.
const DefaultFontPath = "fonts/Arial.ttf"

type Option func(*Document)

func WithLogger(fn func(...any)) Option {
	return func(d *Document) {
		d.logger = fn
	}
}

func WithPageSize(w, h float64, unit string) Option {
	return func(d *Document) {
		// unit is ignored for now as fpdf is initialized with mm by default
		// but we could recreate internal if needed.
		// For now let's assume mm.
	}
}

func WithMargins(left, top, right, bottom float64) Option {
	return func(d *Document) {
		d.internal.SetMargins(left, top, right)
		d.internal.SetAutoPageBreak(true, bottom)
	}
}

func WithDefaultFont(family, path string) Option {
	return func(d *Document) {
		d.fontFamily = family
		// We can't load it here easily because it might need d.readFile
		// which is initialized after options? No, it's initialized in initIO.
	}
}

// NewDocument creates a new Document instance with UTF-8 support.
func NewDocument(opts ...Option) *Document {
	d := &Document{
		fonts:      []fontEntry{},
		images:     []KeyValue{},
		fontFamily: "Arial",
		theme:      DefaultTheme,
	}
	d.initIO() // initializes logger + IO depending on build tag
	d.internal = fpdf.New(
		fpdf.WriteFileFunc(d.writeFile),
		fpdf.ReadFileFunc(d.readFile),
		fpdf.FileSizeFunc(d.fileSize),
	)
	d.internal.SetMargins(20, 20, 20)
	d.internal.SetAutoPageBreak(true, 20)

	for _, opt := range opts {
		opt(d)
	}

	d.loadDefaultFont()
	return d
}

// loadDefaultFont loads Arial as UTF-8 font so the default "Arial" supports unicode.
func (d *Document) loadDefaultFont() {
	data, err := d.readFile(DefaultFontPath)
	if err != nil {
		return // fallback to built-in Arial (Latin-1 only)
	}
	d.addUTF8FontAllStyles(d.fontFamily, data)
}

// SetTheme sets the document theme. Missing numeric fields (Sizes, Spacing)
// inherit from DefaultTheme so callers can specify only the colors they care
// about without producing zero-height text.
func (d *Document) SetTheme(theme Theme) *Document {
	if theme.Sizes.H1 == 0 {
		theme.Sizes.H1 = DefaultTheme.Sizes.H1
	}
	if theme.Sizes.H2 == 0 {
		theme.Sizes.H2 = DefaultTheme.Sizes.H2
	}
	if theme.Sizes.H3 == 0 {
		theme.Sizes.H3 = DefaultTheme.Sizes.H3
	}
	if theme.Sizes.Body == 0 {
		theme.Sizes.Body = DefaultTheme.Sizes.Body
	}
	if theme.Sizes.Small == 0 {
		theme.Sizes.Small = DefaultTheme.Sizes.Small
	}
	if theme.Spacing.Paragraph == 0 {
		theme.Spacing.Paragraph = DefaultTheme.Spacing.Paragraph
	}
	if theme.Spacing.Section == 0 {
		theme.Spacing.Section = DefaultTheme.Spacing.Section
	}
	if theme.Spacing.Page == 0 {
		theme.Spacing.Page = DefaultTheme.Spacing.Page
	}
	d.theme = theme
	if theme.FontFamily != "" {
		d.fontFamily = theme.FontFamily
	}
	if theme.Page.Width > 0 && theme.Page.Height > 0 {
		d.internal.SetPageSizeMM(theme.Page.Width, theme.Page.Height)
	}
	if theme.Margin.Top > 0 || theme.Margin.Right > 0 || theme.Margin.Bottom > 0 || theme.Margin.Left > 0 {
		m := theme.Margin
		d.internal.SetMargins(m.Left, m.Top, m.Right)
		d.internal.SetAutoPageBreak(true, m.Bottom)
	}
	return d
}

// Err returns the first accumulated error.
func (d *Document) Err() error {
	return d.err
}

func (d *Document) addError(err error) {
	if d.err == nil && err != nil {
		d.err = err
	}
}

// SetDefaultFont sets the default font family for the document.
func (d *Document) SetDefaultFont(family string) *Document {
	d.fontFamily = family
	return d
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

// kvGet retrieves a value from a KeyValue slice by key.
// Note: This helper is currently available for future use in resource lookups.
func kvGet(kv []KeyValue, key string) (string, bool) {
	for i := range kv {
		if kv[i].Key == key {
			return kv[i].Value, true
		}
	}
	return "", false
}

// RegisterFont registers a TTF font for all styles (regular, bold, italic, bold-italic).
func (d *Document) RegisterFont(family, path string) *Document {
	d.fonts = append(d.fonts, fontEntry{family: family, path: path, allStyles: true})
	return d
}

// RegisterFontStyle registers a TTF font for a specific style ("", "B", "I", "BI").
func (d *Document) RegisterFontStyle(family, style, path string) *Document {
	d.fonts = append(d.fonts, fontEntry{family: family, style: style, path: path})
	return d
}

// RegisterImage registers an image to be loaded.
func (d *Document) RegisterImage(name, path string) *Document {
	d.images = append(d.images, KeyValue{Key: name, Value: path})
	return d
}

// Load loads all registered resources.
func (d *Document) Load(cb func(error)) {
	for i := range d.fonts {
		fe := d.fonts[i]
		data, err := d.readFile(fe.path)
		if err != nil {
			cb(err)
			return
		}
		if fe.allStyles {
			d.addUTF8FontAllStyles(fe.family, data)
		} else {
			d.internal.AddUTF8FontFromBytes(fe.family, fe.style, data)
		}
	}

	for i := range d.images {
		name, path := d.images[i].Key, d.images[i].Value
		data, err := d.readFile(path)
		if err != nil {
			cb(err)
			return
		}

		ext := ""
		if idx := LastIndex(path, "."); idx != -1 {
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
	if d.err != nil {
		return d.err
	}
	err := d.internal.OutputFileAndClose(path)
	if d.err == nil {
		return err
	}
	return d.err
}

// OutputTo writes the generated PDF into the provided writer.
func (d *Document) OutputTo(w io.Writer) error {
	if d.err != nil {
		return d.err
	}
	err := d.internal.Output(w)
	if d.err == nil {
		return err
	}
	return d.err
}

// --- Base Components ---

// AddText adds a text paragraph in flow mode.
func (d *Document) AddText(text string) *TextElement {
	return &TextElement{doc: d, content: text, align: "L"}
}

// AddHeader1 adds a level 1 header.
func (d *Document) AddHeader1(text string) *Document {
	d.setTextColor(d.theme.Accent)
	d.internal.SetFont(d.fontFamily, "B", d.theme.Sizes.H1)
	d.internal.CellFormat(0, d.theme.Sizes.H1/2, text, "", 1, "L", false, 0, "")
	d.setTextColor(d.theme.Body)
	d.internal.Ln(d.theme.Spacing.Section)
	return d
}

// AddHeader2 adds a level 2 header.
func (d *Document) AddHeader2(text string) *Document {
	d.setTextColor(d.theme.Accent)
	d.internal.SetFont(d.fontFamily, "B", d.theme.Sizes.H2)
	d.internal.CellFormat(0, d.theme.Sizes.H2/2, text, "", 1, "L", false, 0, "")
	d.setTextColor(d.theme.Body)
	d.internal.Ln(d.theme.Spacing.Section / 2)
	return d
}

// AddHeader3 adds a level 3 header.
func (d *Document) AddHeader3(text string) *Document {
	d.setTextColor(d.theme.Accent)
	d.internal.SetFont(d.fontFamily, "B", d.theme.Sizes.H3)
	d.internal.CellFormat(0, d.theme.Sizes.H3/2, text, "", 1, "L", false, 0, "")
	d.setTextColor(d.theme.Body)
	d.internal.Ln(d.theme.Spacing.Section / 4)
	return d
}

// AddSpace adds vertical space.
func (d *Document) AddSpace(units float64) *Document {
	d.internal.Ln(units)
	return d
}

// SpaceBefore is an alias for AddSpace.
func (d *Document) SpaceBefore(u float64) *Document {
	return d.AddSpace(u)
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

	d.drawLineH(x, y+2, width, "#000000", 0.2)
	d.internal.Ln(5)
	return d
}

func (d *Document) drawLineH(x, y, width float64, color Color, thickness float64) {
	r, g, b, err := color.parse()
	if err != nil {
		d.addError(err)
		return
	}
	d.internal.SetDrawColor(r, g, b)
	d.internal.SetLineWidth(thickness)
	d.internal.Line(x, y, x+width, y)
	d.internal.SetDrawColor(0, 0, 0)
	d.internal.SetLineWidth(0.2)
}

// AddImage adds an image by name (must be registered/loaded) in flow mode.
func (d *Document) AddImage(name string) *ImageElement {
	return &ImageElement{doc: d, name: name, align: "L"}
}

func (d *Document) drawImageAt(name string, x, y, width float64) {
	d.internal.Image(name, x, y, width, 0, false, "", 0, "")
}

func (d *Document) setPosition(x, y float64) {
	d.internal.SetXY(x, y)
}

func (d *Document) setCursorY(y float64) {
	d.internal.SetY(y)
}

func (d *Document) setCursorX(x float64) {
	d.internal.SetX(x)
}

func (d *Document) getCursorY() float64 {
	return d.internal.GetY()
}

func (d *Document) getCursorX() float64 {
	return d.internal.GetX()
}

func (d *Document) drawFilledRect(x, y, w, h float64, color Color) {
	r, g, b, err := color.parse()
	if err != nil {
		d.addError(err)
		return
	}
	d.internal.SetFillColor(r, g, b)
	d.internal.Rect(x, y, w, h, "F")
	d.internal.SetFillColor(255, 255, 255)
}

func (d *Document) setTextColor(color Color) {
	r, g, b, err := color.parse()
	if err != nil {
		d.addError(err)
		return
	}
	d.internal.SetTextColor(r, g, b)
}

func (d *Document) drawTextAt(x, y float64, text, style string, size float64) {
	d.internal.SetXY(x, y)
	d.internal.SetFont(d.fontFamily, style, size)
	d.internal.Cell(0, size/2.8, text)
}

func (d *Document) cellAt(x, y, w, h float64, text, style string, size float64, align string) {
	d.internal.SetXY(x, y)
	d.internal.SetFont(d.fontFamily, style, size)
	d.internal.CellFormat(w, h, text, "", 0, align, false, 0, "")
}

func (d *Document) measureText(text, style string, size float64) (width, height float64) {
	d.internal.SetFont(d.fontFamily, style, size)
	return d.internal.GetStringWidth(text), size / 2.8
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
		d.internal.SetFont(d.fontFamily, "I", 8)
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
	leftText   string
	pageTotal  bool
}

func (d *Document) SetPageFooter() *PageFooter {
	pf := &PageFooter{doc: d}
	d.internal.SetFooterFunc(func() {
		d.internal.SetY(-15) // Standard footer position
		d.internal.SetFont(d.fontFamily, "I", 8)

		if pf.leftText != "" {
			d.internal.SetTextColor(130, 130, 130)
			d.internal.CellFormat(0, 10, pf.leftText, "", 0, "L", false, 0, "")
			d.internal.SetTextColor(0, 0, 0)
		}

		if pf.centerText != "" {
			d.internal.CellFormat(0, 10, pf.centerText, "", 0, "C", false, 0, "")
		}

		if pf.pageTotal {
			pageNo := d.internal.PageNo()
			pageStr := Sprintf("Página %s/{nb}", Convert(pageNo).String())
			d.internal.CellFormat(0, 10, pageStr, "", 0, "R", false, 0, "")
		}
		d.internal.SetTextColor(0, 0, 0)
	})
	return pf
}

func (pf *PageFooter) SetCenterText(t string) *PageFooter {
	pf.centerText = t
	return pf
}

func (pf *PageFooter) WithLeftRight(leftText string) *PageFooter {
	pf.leftText = leftText
	pf.pageTotal = true
	pf.doc.internal.AliasNbPages("")
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

const (
	FontBold    = "B"
	FontItalic  = "I"
	FontRegular = ""
)
