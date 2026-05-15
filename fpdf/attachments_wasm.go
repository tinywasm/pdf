//go:build wasm

package fpdf

import (
	. "github.com/tinywasm/fmt"
)

type Attachment struct {
	Content      []byte
	Filename     string
	Description  string
	objectNumber int
}

type annotationAttach struct {
	*Attachment
	x, y, w, h float64
}

func (f *Fpdf) SetAttachments(as []Attachment) {
	f.err = Err("attachments", "unsupported", "in WASM")
}

func (f *Fpdf) AddAttachmentAnnotation(a *Attachment, x, y, w, h float64) {
	f.err = Err("attachments", "unsupported", "in WASM")
}

func (f *Fpdf) putAttachments() {
	// no-op in WASM
}

func (f *Fpdf) putAnnotationsAttachments() {
	// no-op in WASM
}

func (f *Fpdf) getEmbeddedFiles() string {
	return ""
}

func (f *Fpdf) putAttachmentAnnotationLinks(out *fmtBuffer, page int) {
	// no-op in WASM
}
