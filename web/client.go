//go:build wasm

package main

import (
	"github.com/tinywasm/pdf"
	"github.com/tinywasm/pdf/web/ui"
)

func main() {
	// Crear instancia de Document
	doc := pdf.NewDocument()

	doc.Log("Document inicializado...")

	// Configurar UI
	ui.Setup(doc)

	doc.Log("Aplicación lista")

	// Mantener el programa ejecutándose
	select {}
}
