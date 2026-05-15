# Plan: Professional Layout API

## Restricciones de la librería

- **Binario mínimo**: no introducir imports de stdlib que no existan ya. Usar `github.com/tinywasm/fmt` con dot-import (`. "github.com/tinywasm/fmt"`) — las funciones se usan sin prefijo: `LastIndex`, `Convert`, `Sprintf`, `Errf`.
- **TinyGo compatible**: no usar `map`. En su lugar usar `[]KeyValue` (tipo de `tinywasm/fmt`, disponible sin prefijo por dot-import).
- **Sin lógica de negocio**: la librería no conoce consumidores ni templates específicos.
- **Sin dependencias nuevas**: todo debe implementarse sobre la API existente de `fpdf`.

---

## Análisis de lo existente antes de proponer

| Funcionalidad | Ya existe | Ubicación |
|---|---|---|
| Footer con texto + página | Parcial | `SetPageFooter()` → solo center + pageTotal |
| Header por página | Sí | `SetPageHeader()` → left + right text |
| Línea horizontal | Sí (fpdf) | `fpdf.Line(x1,y1,x2,y2)` + `SetDrawColor` + `SetLineWidth` |
| Rectángulo relleno | Sí (fpdf) | `fpdf.Rect(x,y,w,h,"F")` + `SetFillColor` |
| Color de texto | Sí (fpdf) | `fpdf.SetTextColor` — expuesto en doc como `SetTextColor` ✅ |
| Posición absoluta | Sí (fpdf) | `fpdf.SetXY` — expuesto como `SetPosition` ✅ |
| Imagen en posición | Sí (fpdf) | `fpdf.Image` — expuesto como `DrawImageAt` ✅ |
| Total de páginas `{nb}` | Sí (fpdf) | `AliasNbPages("")` + `{nb}` en texto |

**Conclusión**: footer, línea con color y rect relleno son wrappers mínimos sobre fpdf, no código nuevo. No justifican imports adicionales.

---

## Cambios propuestos

### 1. `SetPageFooter` — agregar left+right (texto + página)

El `PageFooter` actual solo tiene `centerText` y `pageTotal`. Agregar:

```go
func (pf *PageFooter) WithLeftRight(leftText string) *PageFooter
```

- Izquierda: `leftText` en gris (130,130,130), fuente 8pt itálica
- Derecha: `Sprintf("Página %s/{nb}", Convert(pageNo).String())` — sin stdlib
- Reset color a negro al salir
- Llama `AliasNbPages("")` internamente
- **Conflicto con `SetCenterText`**: `WithLeftRight` y `SetCenterText` son mutuamente excluyentes. `WithLeftRight` ignora `centerText`. El `SetFooterFunc` se registra una vez; si se llaman ambos, el último registrado gana — documentar esto.

**Justificación**: reutiliza `SetFooterFunc` ya existente. Cero código nuevo en fpdf.

---

### 2. `AddHeader1/2/3` — respetar color activo, reset al salir

Actualmente no resetean el color de texto, lo que causa que el color seteado antes persista después del header. Fix mínimo: agregar `SetTextColor(0,0,0)` al final de cada AddHeader.

```go
func (d *Document) AddHeader2(text string) *Document {
    d.internal.SetFont(d.fontFamily, "B", 12)  // ver punto 4
    d.internal.CellFormat(0, 7, text, "", 1, "L", false, 0, "")
    d.internal.SetTextColor(0, 0, 0) // reset explícito
    d.internal.Ln(2)
    return d
}
```

**Flujo recomendado para el consumidor:**
```go
doc.SetTextColor(30, 60, 120)
doc.AddHeader2("1. Objeto del Contrato")  // imprime azul, sale en negro
doc.AddText("...").Draw()                 // ya en negro
```

---

### 3. `DrawLineH` — línea horizontal con color y grosor

Wrapper sobre `fpdf.Line` + `SetDrawColor` + `SetLineWidth`:

```go
func (d *Document) DrawLineH(y, width float64, r, g, b int, thickness float64) *Document
```

- Usa `d.internal.GetX()` y márgenes para calcular x automáticamente
- Reset `SetDrawColor(0,0,0)` y `SetLineWidth(0.2)` al salir
- **No agrega ningún import**: todo es fpdf interno

---

### 4. `fontFamily` — campo en `Document` para escalabilidad de fuente

Actualmente "Arial" está hardcodeado en 10+ lugares. Si el consumidor registra otra fuente como default, los headers siguen usando Arial.

```go
type Document struct {
    internal   *fpdf.Fpdf
    logger     func(message ...any)
    fonts      []KeyValue // Key: family, Value: path — compatible TinyGo (dot-import)
    images     []KeyValue // Key: name,   Value: path — compatible TinyGo (dot-import)
    fontFamily string     // default: "Arial", inicializado en NewDocument()
}
```

`NewDocument()` debe inicializar `fontFamily: "Arial"` explícitamente.

Búsqueda de valor por key con iteración simple (sin map):
```go
func kvGet(kv []KeyValue, key string) (string, bool) {
    for i := range kv {
        if kv[i].Key == key {
            return kv[i].Value, true
        }
    }
    return "", false
}
```

`RegisterFont` y `RegisterImage` hacen append al slice. Los registros son pocos (< 10), por lo que la búsqueda lineal es O(n) aceptable.

Agregar método:
```go
func (d *Document) SetDefaultFont(family string) *Document {
    d.fontFamily = family
    return d
}
```

Todos los `AddHeader1/2/3`, `AddText`, `AddSeparator` usan `d.fontFamily` en lugar de la string literal `"Arial"`. Incluye `loadDefaultFont()` que registra la fuente bajo el nombre `d.fontFamily`.

**Justificación de escalabilidad**: cambiar la fuente de todo el documento requiere un solo `doc.SetDefaultFont("MyFont")` en lugar de buscar y reemplazar.

---

## Eliminar import `strings` de `document.go`

Una sola línea usa `strings.LastIndex` en `document.go:104`:

```go
// Antes
if idx := strings.LastIndex(path, "."); idx != -1 {
    ext = path[idx+1:]
}

// Después — dot-import, sin prefijo
if idx := LastIndex(path, "."); idx != -1 {
    ext = path[idx+1:]
}
```

`LastIndex` ya existe en `tinywasm/fmt/operations.go`. El import `"strings"` se elimina completamente.

---

## Lo que NO se propone

- No se propone `SetDefaultFooter` como función nueva: `SetPageFooter().WithLeftRight(text)` cubre el caso con la API existente.
- No se propone ningún helper de layout de columnas: eso es responsabilidad del consumidor usando `SetPosition` + `CellAt`.
- No se introduce `fmt.Sprintf`, `strconv`, ni `strings` — todo via `tinywasm/fmt`.

---

## Corrección de errores de compilación `[js,wasm]`

### `undefined: fpdf.MakeFont` en tests y herramienta makefont

**Causa**: `MakeFont` está definida en `fpdf/font.go` con `//go:build !wasm`. Los archivos que la usan no tienen la misma restricción, por lo que el compilador la busca al compilar para `js,wasm` y falla.

**Archivos afectados**:

| Archivo | Línea | Fix |
|---|---|---|
| `fpdf/makefont/makefont.go` | 44 | Agregar `//go:build !wasm` en línea 1 |
| `fpdf/fpdf_test.go` | 735 | Agregar `//go:build !wasm` en línea 1 |
| `fpdf/issues_test.go` | 235 | Agregar `//go:build !wasm` en línea 1 |

**Fix**: agregar la build tag al inicio de cada archivo (antes del `package`):

```go
//go:build !wasm

package ...
```

**Justificación**: `makefont` es una herramienta CLI para generar definiciones de fuente — no tiene sentido en WASM. Los tests que la usan son tests de parsing de fuentes AFM/Type1, tampoco aplicables en WASM. No se pierde cobertura útil.

---

## Orden de ejecución

1. [ ] Agregar `//go:build !wasm` a `makefont/makefont.go`, `fpdf_test.go`, `issues_test.go`
2. [ ] Reemplazar `strings.LastIndex` por `fmt.LastIndex` y eliminar import `"strings"` en `document.go`
3. [ ] Reemplazar `map[string]string` por `[]fmt.KeyValue` en `Document`
4. [ ] Agregar campo `fontFamily` a `Document` y método `SetDefaultFont`
5. [ ] Reemplazar literales `"Arial"` en AddHeader/AddText/etc. por `d.fontFamily`
6. [ ] Agregar reset `SetTextColor(0,0,0)` al final de `AddHeader1/2/3`
7. [ ] Agregar `WithLeftRight(leftText string)` a `PageFooter`
8. [ ] Agregar `DrawLineH(y, width, r, g, b, thickness)`
