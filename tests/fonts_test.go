package pdf_test

import (
	"os"
	"testing"

	"github.com/tinywasm/pdf"
)

// fontDir is relative to this test's working directory: readFile() is a plain
// os.ReadFile, so a registered path resolves against the package dir.
const fontDir = "../fpdf/fonts/"

// fontFamily is one family under test and the four files backing its styles.
// Italic is a real file for Inter and Roboto; DroidSans has no italic cut, so
// it points at the upright face — which is exactly why it is here, as the
// baseline the other two are meant to improve on.
type fontFamily struct {
	name                            string
	regular, bold, italic, boldItal string
	note                            string
}

var families = []fontFamily{
	{
		name: "Roboto", regular: "Roboto-Regular.ttf", bold: "Roboto-Bold.ttf",
		italic: "Roboto-Italic.ttf", boldItal: "Roboto-BoldItalic.ttf",
		note: "subset latino, cursivas reales",
	},
	{
		name: "Inter", regular: "Inter-Regular.ttf", bold: "Inter-Bold.ttf",
		italic: "Inter-Italic.ttf", boldItal: "Inter-BoldItalic.ttf",
		note: "subset latino, cursivas reales",
	},
	{
		name: "Droid", regular: "DroidSans.ttf", bold: "DroidSans-Bold.ttf",
		italic: "DroidSans.ttf", boldItal: "DroidSans-Bold.ttf",
		note: "sin cursiva real (apunta a la recta) y sin simbolo €",
	},
}

// The sample exercises what a quote document actually prints: Spanish
// diacritics, the inverted marks, typographic quotes and dashes, and the euro
// sign that the DroidSans subset is missing.
const sample = `Señor Muñoz — ítem “café orgánico” ¿1.250 €? ¡Sí! ÁÉÍÓÚ üÜ ñÑ ª º`

func register(doc *pdf.Document, f fontFamily) error {
	doc.RegisterFontStyle(f.name, "", fontDir+f.regular)
	doc.RegisterFontStyle(f.name, "B", fontDir+f.bold)
	doc.RegisterFontStyle(f.name, "I", fontDir+f.italic)
	doc.RegisterFontStyle(f.name, "BI", fontDir+f.boldItal)

	var loadErr error
	doc.Load(func(err error) { loadErr = err })
	return loadErr
}

// TestFonts_AllStyles renders every family in all four styles into a single
// PDF so the faces can be compared side by side. Open tests/test_fonts.pdf.
func TestFonts_AllStyles(t *testing.T) {
	doc := pdf.NewDocument()

	for _, f := range families {
		if err := register(doc, f); err != nil {
			t.Fatalf("registering %s: %v", f.name, err)
		}
	}

	// The family must be selected before ANY text is drawn. Until then
	// d.fontFamily is "Arial", and loadDefaultFont() resolves fonts/Arial.ttf
	// against the working directory — which does not exist here, so the
	// document silently falls back to the built-in Latin-1 Arial and every
	// accent written under it comes out mangled.
	doc.SetDefaultFont(families[0].name)
	doc.SetFont(families[0].name, 12)

	doc.AddPage()
	doc.AddHeader1("Comparativa de fuentes")

	for _, f := range families {
		// SetDefaultFont, not just SetFont: SetFont only touches the internal
		// fpdf state, while the text elements resolve their face through
		// d.fontFamily. Setting only the latter leaves every string on the
		// built-in Latin-1 Arial, which mangles any accent it is given.
		doc.SetDefaultFont(f.name)
		doc.SetFont(f.name, 12)
		doc.AddHeader2(f.name + " — " + f.note)

		doc.AddText("REGULAR   " + sample).Draw()
		doc.AddText("BOLD      " + sample).Bold().Draw()
		doc.AddText("ITALIC    " + sample).Italic().Draw()
		doc.AddText("BOLD+ITAL " + sample).Bold().Italic().Draw()

		doc.SpaceBefore(6)
		doc.AddSeparator()
		doc.SpaceBefore(6)
	}

	const out = "test_fonts.pdf"
	if err := doc.WritePdf(out); err != nil {
		t.Fatalf("WritePdf: %v", err)
	}

	st, err := os.Stat(out)
	if err != nil {
		t.Fatalf("stat %s: %v", out, err)
	}
	t.Logf("escrito %s (%d bytes) — abrelo para ver el resultado", out, st.Size())
}

// TestFonts_SubsetsAreSmallerThanDroid is the size claim, checked rather than
// asserted in prose: the subsets must come in under the DroidSans face they
// replace, despite carrying a real italic and the euro sign it lacks.
func TestFonts_SubsetsAreSmallerThanDroid(t *testing.T) {
	size := func(name string) int64 {
		st, err := os.Stat(fontDir + name)
		if err != nil {
			t.Fatalf("stat %s: %v", name, err)
		}
		return st.Size()
	}

	droid := size("DroidSans.ttf")
	for _, name := range []string{"Roboto-Regular.ttf", "Inter-Regular.ttf"} {
		got := size(name)
		if got >= droid {
			t.Errorf("%s pesa %d B, no mejora a DroidSans.ttf (%d B)", name, got, droid)
		}
		t.Logf("%-20s %6d B  (DroidSans.ttf %d B)", name, got, droid)
	}
}
