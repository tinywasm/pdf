package pdf

import (
	"testing"

	"webtyp.com/fmt"
)

// Test para verificar el problema con pdfVersion.String()
func TestPDFVersionFormatting(t *testing.T) {
	// Verificar que pdfVersion tenga un valor
	// pdfVersion es un tipo interno, pero podemos probar el formateo

	// Simular lo que hace putheader()
	result := fmt.Sprintf("%%PDF-%s", "1.3")
	expected := "%PDF-1.3"

	if result != expected {
		t.Errorf("Fmt con string literal falló. Expected: %q, got: %q", expected, result)
	}

	t.Logf("Resultado con string literal: %q", result)

	// Ahora probar con un tipo custom que tiene String()
	type customVersion string

	ver := customVersion("1.4")
	result2 := fmt.Sprintf("%%PDF-%s", ver)
	t.Logf("Resultado con custom type: %q", result2)

	if result2 == "%PDF-" {
		t.Error("Fmt no manejó el custom type correctamente, devolvió: %PDF-")
	}
}
