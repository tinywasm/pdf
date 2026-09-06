//go:build wasm
// +build wasm

package ui

import (
	"syscall/js"

	"webtyp.com/pdf"
)

var (
	Doc       *pdf.Document
	textInput js.Value
	pdfEmbed  js.Value
)

// Setup inicializa y configura toda la interfaz de usuario
func Setup(doc *pdf.Document) {
	Doc = doc
	setupUI()
}

func setupUI() {
	document := js.Global().Get("document")
	body := document.Get("body")
	body.Set("innerHTML", "")

	// Crear contenedor principal
	container := document.Call("createElement", "div")
	container.Set("className", "container")

	// Título
	title := document.Call("createElement", "h1")
	title.Set("textContent", "Generador de PDF con Document API")
	container.Call("appendChild", title)

	// Sección del formulario
	formSection := document.Call("createElement", "div")
	formSection.Set("className", "form-section")

	// Campo de texto
	inputLabel := document.Call("createElement", "label")
	inputLabel.Set("textContent", "Título del documento:")
	formSection.Call("appendChild", inputLabel)

	textInput = document.Call("createElement", "input")
	textInput.Set("type", "text")
	textInput.Set("value", "Documento de Ejemplo")
	formSection.Call("appendChild", textInput)

	// Botón generar PDF
	btn := document.Call("createElement", "button")
	btn.Set("textContent", "Generar PDF")

	// Evento click del botón
	btn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		GeneratePDF()
		return nil
	}))
	formSection.Call("appendChild", btn)

	container.Call("appendChild", formSection)

	// Contenedor para el PDF
	pdfContainer := document.Call("createElement", "div")
	pdfContainer.Set("className", "pdf-container")
	pdfContainer.Set("id", "pdf-container")
	container.Call("appendChild", pdfContainer)

	body.Call("appendChild", container)

	// Cargar estilos CSS
	loadStyles()
}

func loadStyles() {
	document := js.Global().Get("document")
	head := document.Get("head")

	// Verificar si ya existe el link de estilos
	existingLink := document.Call("querySelector", "link[href='style.css']")
	if !existingLink.IsNull() {
		return
	}

	link := document.Call("createElement", "link")
	link.Set("rel", "stylesheet")
	link.Set("href", "style.css")
	head.Call("appendChild", link)
}

// ShowError muestra un mensaje de error en la UI
func ShowError(message string) {
	document := js.Global().Get("document")
	pdfContainer := document.Call("getElementById", "pdf-container")
	pdfContainer.Set("innerHTML", "")

	errorDiv := document.Call("createElement", "div")
	errorDiv.Set("className", "error-message")
	errorDiv.Set("textContent", message)

	pdfContainer.Call("appendChild", errorDiv)
}

// ShowPDF muestra el PDF en el embed
func ShowPDF(dataURL string) {
	Doc.Log("ShowPDF llamado con URL:", dataURL)

	document := js.Global().Get("document")
	pdfContainer := document.Call("getElementById", "pdf-container")

	if pdfContainer.IsNull() {
		Doc.Log("ERROR: pdf-container no encontrado")
		return
	}

	pdfContainer.Set("innerHTML", "")

	pdfEmbed = document.Call("createElement", "embed")
	pdfEmbed.Set("src", dataURL)
	pdfEmbed.Set("type", "application/pdf")
	pdfEmbed.Set("className", "pdf-embed")
	pdfEmbed.Set("width", "100%")
	pdfEmbed.Set("height", "600px")

	pdfContainer.Call("appendChild", pdfEmbed)

	Doc.Log("Embed creado y agregado al contenedor")
}

// GetTitleText obtiene el texto del campo de entrada
func GetTitleText() string {
	titleText := textInput.Get("value").String()
	if titleText == "" {
		titleText = "Documento sin título"
	}
	return titleText
}
