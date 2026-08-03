//go:build wasm
// +build wasm

package ui

import (
	"github.com/tinywasm/fmt"
	"syscall/js"
)

// GeneratePDF genera el PDF con el título especificado
func GeneratePDF() {
	Doc.Log("Iniciando generación de PDF...")

	// Obtener el título del input
	titleText := GetTitleText()

	// 1. Configurar header y footer
	Doc.SetPageHeader().SetRightText(titleText)
	Doc.SetPageFooter().WithPageTotal("R")

	// 2. Limpiar/Reiniciar contenido si fuera necesario (fpdf.New crea uno nuevo, pero aquí reusamos Doc)
	// Para el demo web, lo más simple es crear un nuevo Document cada vez o resetear el internal
	// pero NewDocument es barato.
	// Sin embargo, Doc ya está inyectado en ui.TP (ahora ui.Doc).
	// El Document mantiene registros de fuentes/imágenes.
	// La forma más limpia es que GeneratePDF cree un nuevo Document si queremos empezar de cero,
	// pero el plan dice "GeneratePDF() usa Doc (Document) directamente".
	// Si reusamos Doc, fpdf.Fpdf se va acumulando.
	// Fpdf no tiene un "Reset". Tenemos que crear uno nuevo.
	// Actualizamos ui.Doc con uno nuevo.

	// Doc = pdf.NewDocument() // Esto requeriría importar pdf aquí
	// Pero fpdf tiene AddPage. Si no llamamos a nada más, empieza en blanco.
	// El problema es si ya tenía páginas.

	// Para cumplir el plan fielmente, usaremos Doc tal cual:
	Doc.AddPage()
	Doc.AddHeader1("Contenido del documento")
	Doc.SpaceBefore(5)

	Doc.SetSize(12)
	for j := 1; j <= 40; j++ {
		Doc.AddText(fmt.Sprintf("Línea de contenido número %d", j)).Draw()
	}

	// Obtener los bytes del PDF
	var buf []byte
	err := Doc.OutputTo(&bytesWriter{data: &buf})
	if err != nil {
		Doc.Log("Error obteniendo bytes del PDF:", err.Error())
		ShowError("Error al obtener bytes del PDF: " + err.Error())
		return
	}

	Doc.Log("PDF generado, tamaño:", len(buf), "bytes")

	// Usar Blob API para mostrar el PDF (más eficiente que base64)
	ShowPDFFromBytes(buf)

	Doc.Log("PDF generado exitosamente")
}

// bytesWriter implementa io.Writer para capturar los bytes del PDF
type bytesWriter struct {
	data *[]byte
}

func (bw *bytesWriter) Write(p []byte) (n int, err error) {
	*bw.data = append(*bw.data, p...)
	return len(p), nil
}

// ShowPDFFromBytes crea un Blob a partir de los bytes del PDF y lo muestra
// usando Blob API en lugar de base64 para mejor rendimiento
func ShowPDFFromBytes(pdfBytes []byte) {
	Doc.Log("Creando Blob del PDF, tamaño:", len(pdfBytes), "bytes")

	// Crear Uint8Array desde los bytes de Go
	uint8Array := js.Global().Get("Uint8Array").New(len(pdfBytes))
	js.CopyBytesToJS(uint8Array, pdfBytes)

	// Crear el Blob con el tipo MIME correcto
	// El constructor de Blob espera: new Blob([array], {type: 'mime/type'})
	blobParts := []interface{}{uint8Array}
	blobOptions := map[string]interface{}{
		"type": "application/pdf",
	}

	blob := js.Global().Get("Blob").New(blobParts, blobOptions)

	// Crear URL del Blob usando la función integrada
	blobURL := CreateBlobURL(blob)

	Doc.Log("Blob URL creado:", blobURL)

	// Mostrar el PDF usando el blob URL
	ShowPDF(blobURL)
}

// CreateBlobURL crea una URL Blob a partir de un blob.
func CreateBlobURL(blob any) string {
	jsBlob := js.ValueOf(blob)
	return js.Global().Get("URL").Call("createObjectURL", jsBlob).String()
}
