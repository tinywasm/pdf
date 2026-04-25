package fpdf

import (
	"github.com/tinywasm/unixid"
)

var uid, _ = unixid.NewUnixID()

func generateImageID(info *ImageInfoType) (string, error) {
	var id string
	uid.SetNewID(&id)
	return id, nil
}

// generateFontID generates a font Id from the font definition
func generateFontID(fdt fontDefType) (string, error) {
	// Simple deterministic ID
	return fdt.Tp + "_" + fdt.Name, nil
}
