//go:build wasm

package main

import (
	"github.com/tinywasm/pdf"
	"github.com/tinywasm/pdf/web/ui"
)

func main() {
	// Cargar tipografía para el cliente Web (WASM)
	tf, err := pdf.LoadTypeface(
		"fonts/Roboto-Regular.ttf",
		"fonts/Roboto-Bold.ttf",
		"fonts/Roboto-Italic.ttf",
		"fonts/Roboto-BoldItalic.ttf",
	)
	if err != nil {
		panic("Error cargando tipografía Roboto: " + err.Error())
	}

	// Crear instancia de Document con la tipografía cargada
	doc := pdf.NewDocument(tf)

	doc.Log("Document inicializado con tipografía Roboto...")

	// Configurar UI
	ui.Setup(doc)

	doc.Log("Aplicación lista")

	// Mantener el programa ejecutándose
	select {}
}
