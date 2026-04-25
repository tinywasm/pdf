package fpdf

import (
	"github.com/tinywasm/time"
)

// SetDefaultCreationDate sets the default value of the document creation date
// that will be used when initializing a new Fpdf instance. See
// SetCreationDate() for more details.
func SetDefaultCreationDate(tm int64) {
	gl.creationDate = pdfTime(tm)
}

// SetDefaultModificationDate sets the default value of the document modification date
// that will be used when initializing a new Fpdf instance. See
// SetCreationDate() for more details.
func SetDefaultModificationDate(tm int64) {
	gl.modDate = pdfTime(tm)
}

// GetCreationDate returns the document's internal CreationDate value.
func (f *Fpdf) GetCreationDate() int64 {
	return int64(f.creationDate)
}

// SetCreationDate fixes the document's internal CreationDate value. By
// default, the time when the document is generated is used for this value.
// This method is typically only used for testing purposes to facilitate PDF
// comparison. Specify a zero-value time to revert to the default behavior.
func (f *Fpdf) SetCreationDate(tm int64) {
	f.creationDate = pdfTime(tm)
}

// GetModificationDate returns the document's internal ModDate value.
func (f *Fpdf) GetModificationDate() int64 {
	return int64(f.modDate)
}

// SetModificationDate fixes the document's internal ModDate value.
// See `SetCreationDate` for more details.
func (f *Fpdf) SetModificationDate(tm int64) {
	f.modDate = pdfTime(tm)
}

// returns Now() if tm is zero
func timeOrNow(tm pdfTime) int64 {
	if tm == 0 {
		return time.Now()
	}
	return int64(tm)
}

func formatPDFDate(tm pdfTime) string {
	nano := timeOrNow(tm)
	s := time.FormatISO8601(nano)
	// ISO8601: "2026-04-02T15:30:45Z" -> "20260402153045"
	if len(s) < 19 {
		return "D:00000000000000"
	}
	return "D:" + s[0:4] + s[5:7] + s[8:10] + s[11:13] + s[14:16] + s[17:19]
}
