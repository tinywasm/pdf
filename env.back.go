//go:build !wasm
// +build !wasm

package pdf

import (
	"github.com/tinywasm/fmt"
	"os"
)

// initIO inicializa las funciones de IO para entorno backend (no-wasm)
func (d *Document) initIO() {
	// Inicializar logger para backend usando fmt.Println
	d.logger = func(message ...any) {
		fmt.Println(message...)
	}
}

// writeFile escribe un archivo en el sistema de archivos usando os
func (d *Document) writeFile(filePath string, content []byte) error {
	return os.WriteFile(filePath, content, 0644)
}

// readFile lee un archivo del sistema de archivos usando os
func readFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

// readFile lee un archivo del sistema de archivos usando os
func (d *Document) readFile(filePath string) ([]byte, error) {
	return readFile(filePath)
}

// fileSize obtiene el tamaño de un archivo usando os.Stat
func (d *Document) fileSize(filePath string) (int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}
