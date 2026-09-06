package httpimg

import (
	"errors"
	"io"
	"net/http"

	fpdf "webtyp.com/pdf/fpdf"
)

// httpimgPdf is a partial interface that only implements the functions we need
// from the PDF generator to put the HTTP images on the PDF.
type httpimgPdf interface {
	GetImageInfo(imageStr string) *fpdf.ImageInfoType
	ImageTypeFromMime(mimeStr string) string
	RegisterImageReader(imgName, tp string, r io.Reader) *fpdf.ImageInfoType
	SetError(err error)
}

// Register registers a HTTP image. Downloading the image from the provided URL
// and adding it to the PDF but not adding it to the page. Use Image() with the
// same URL to add the image to the page.
func Register(f httpimgPdf, urlStr, tp string) (info *fpdf.ImageInfoType) {
	info = f.GetImageInfo(urlStr)

	if info != nil {
		return
	}

	resp, err := http.Get(urlStr)

	if err != nil {
		f.SetError(err)
		return
	}

	defer resp.Body.Close()

	if tp == "" {
		contentType := resp.Header.Get("Content-Type")
		if contentType != "" {
			tp = f.ImageTypeFromMime(contentType)
		} else {
			// If Content-Type is missing, cannot determine image type
			f.SetError(errors.New("missing Content-Type header, cannot determine image type"))
			return
		}
	}

	return f.RegisterImageReader(urlStr, tp, resp.Body)
}
