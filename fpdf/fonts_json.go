package fpdf

import (
	"webtyp.com/json"
)

func unmarshalFontDef(data []byte, def *fontDefType) error {
	return json.Decode(data, def)
}
