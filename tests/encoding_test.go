package pdf_test

import (
	"os"
	"strings"
	"testing"

	"github.com/tinywasm/pdf"
)

// usesCoreFont reports whether the PDF fell back to one of the built-in
// Latin-1 base-14 faces. A document that registered a UTF-8 family and drew
// only with it never needs them.
func usesCoreFont(t *testing.T, path string) bool {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return strings.Contains(string(b), "/Helvetica")
}

// TestUnselectedFamilyIsReportedNotSilentlyDowngraded reproduces the silent
// corruption path.
//
// Registering a UTF-8 font is not the same as selecting it: until
// SetDefaultFont runs, d.fontFamily is still "Arial", and loadDefaultFont()
// resolves DefaultFontPath ("fonts/Arial.ttf") against the working directory.
// When that file is not there the document keeps the built-in Latin-1
// Helvetica, and every accented byte drawn under it is emitted raw — "Señor"
// reaches the page as "SeÃ±or".
//
// The defect is not the fallback, it is that nothing says so: WritePdf
// returns nil and Document.Err() stays nil, so the only way to find out is to
// open the PDF and look at it. fpdf sets its own f.err on an undefined font,
// but that error is never surfaced through Document.
//
// This test fails today and must pass once the fix lands.
func TestUnselectedFamilyIsReportedNotSilentlyDowngraded(t *testing.T) {
	const out = "test_encoding_unselected.pdf"
	t.Cleanup(func() { os.Remove(out) })

	doc := pdf.NewDocument()
	doc.RegisterFontStyle("Roboto", "", fontDir+"Roboto-Regular.ttf")
	doc.Load(func(error) {})

	doc.AddPage()
	// No SetDefaultFont: the family is loaded but never selected.
	doc.AddText("Señor Muñoz — ítem “café” 1.250 €").Draw()

	writeErr := doc.WritePdf(out)

	if !usesCoreFont(t, out) {
		return // the family was honoured; nothing to report
	}
	if writeErr == nil && doc.Err() == nil {
		t.Errorf("the document fell back to the Latin-1 core font and mangled every accent, "+
			"yet WritePdf returned %v and Document.Err() returned %v — the failure is invisible to the caller",
			writeErr, doc.Err())
	}
}

// TestSelectedFamilyIsHonoured is the control: the same document with the
// family selected must not touch a core font at all. It passes today and
// guards against a fix that silences the report by disabling the fallback.
func TestSelectedFamilyIsHonoured(t *testing.T) {
	const out = "test_encoding_selected.pdf"
	t.Cleanup(func() { os.Remove(out) })

	doc := pdf.NewDocument()
	doc.RegisterFontStyle("Roboto", "", fontDir+"Roboto-Regular.ttf")
	doc.Load(func(error) {})

	doc.SetDefaultFont("Roboto")
	doc.SetFont("Roboto", 12)
	doc.AddPage()
	doc.AddText("Señor Muñoz — ítem “café” 1.250 €").Draw()

	if err := doc.WritePdf(out); err != nil {
		t.Fatalf("WritePdf: %v", err)
	}
	if usesCoreFont(t, out) {
		t.Errorf("a document drawn entirely with a registered UTF-8 family must not embed a core Latin-1 face")
	}
}
