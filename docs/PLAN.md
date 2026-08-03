---
PLAN: "feat: Cerrar el harness de tipografía"
TAG: v0.1.0
STATUS: running
SESSION: 10102235131514896367
---

## Antes de escribir código: lee [CONSTRUCTION_HARNESS.md](CONSTRUCTION_HARNESS.md)

**Es vinculante, no orientativo.** Este plan no repite sus reglas: las aplica. Cada decisión
de diseño de aquí abajo se justifica citando uno de sus principios numerados, y una
implementación que los contradiga es incorrecta aunque los tests pasen.

Los cuatro que gobiernan este trabajo en concreto:

| # | Principio | Cómo se aplica aquí |
|---|---|---|
| 1 | Typed over `any` | La familia deja de nombrarse con `string`; se usa un handle opaco. |
| 3 | Illegal states unrepresentable | Un documento sin tipografía usable no debe poder construirse. |
| 4 | One way to do each thing | Hoy hay dos sistemas de fuentes y cuatro entradas para elegir una. Queda una de cada. |
| 6 | Fail at compile time | El orden es `compile error → loud diagnostic → (never) silent`. Si la corrección termina en un `if` que devuelve error, no se cerró el harness. |

Y la regla que decide si el trabajo está terminado:

> An API is not published until a consumer-shaped test, inside the library itself, proves it.

Ese test es `tests/fonts_test.go` (ver Verificación).

**Restricciones del ecosistema que este plan da por sentadas:** rompe API pública sin
retrocompatibilidad ni shims de deprecación —el paquete compila a WASM y cada símbolo muerto
enlaza código—, sin mapas, y sin carpetas `internal/`.

---

## Por qué se reescribió este plan

La primera versión proponía *propagar el error de fpdf en `WritePdf`*. Eso deja el fallo
en runtime, que es el **último** escalón del orden de preferencia del harness:

> compile error → loud development diagnostic → (never) silent failure

La pregunta correcta no era «cómo aviso del fallo» sino «por qué es representable». Un
documento sin tipografía usable no debería poder existir.

---

## Defecto 1 — La familia se nombra con un `string`

### Síntoma

```go
doc := pdf.NewDocument()
doc.RegisterFontStyle("Roboto", "", "fonts/Roboto-Regular.ttf")
doc.Load(func(error) {})
doc.AddPage()
doc.AddText("Señor Muñoz € ítem").Draw()   // falta SetDefaultFont
doc.WritePdf("out.pdf")                     // -> nil
```

El PDF sale con `Señor Muñoz € ítem` escrito como `SeÃ±or MuÃ±oz â‚¬ Ã­tem`, dibujado con
el Helvetica core Latin-1. `WritePdf()` devuelve `nil`. `Document.Err()` devuelve `nil`.
Reproducido en `tests/encoding_test.go`.

### Qué principios rompe

| Principio del harness | Cómo se rompe hoy |
|---|---|
| 1 · Typed over `any` | `family string` es un agujero genérico: **cualquier** string compila, sólo algunos funcionan. `SetFont("Robotoo", 12)` compila. |
| 3 · Illegal states unrepresentable | Un `Document` sin tipografía UTF-8 usable es construible y dibuja basura. |
| 4 · One way to do each thing | Dos formas de registrar (`RegisterFont` / `RegisterFontStyle`) y dos de seleccionar (`SetDefaultFont` / `SetFont`) — y de las dos de seleccionar **sólo una** afecta al texto. |
| 6 · Fail at compile time | Fallo silencioso puro: ni error, ni log, ni efecto visible salvo abrir el PDF. |
| 8 · Closed by default | El estado inseguro es el que se obtiene **sin escribir nada**: `NewDocument()` deja `fontFamily = "Arial"` y cae al core Latin-1. |

Y el ítem del checklist que lo nombra directamente:

> **Things you "have to remember".** Any mandatory step the author must remember to call
> (call order, "don't forget X") → that is a hole in the harness.

La secuencia obligatoria hoy es `Register…` → `Load` → `SetDefaultFont` → `SetFont` →
`AddPage` → `AddText`. Olvidar el tercer paso **es** este bug.

### Causa mecánica

1. `NewDocument()` deja `d.fontFamily = "Arial"` (`document.go:72`).
2. `loadDefaultFont()` (`document.go:93`) resuelve `"fonts/Arial.ttf"` —ruta relativa al
   directorio de trabajo— y ante error hace `return` mudo.
3. `fpdf.SetFont()` (`fpdf/fonts.go:264`) no halla `arial`, remapea a `helvetica` core y
   deja `f.isCurrentUTF8 = false`: el texto se emite byte a byte como Latin-1.
4. `WritePdf` (`document.go:251`) sólo consulta `d.err`, nunca el `f.err` de fpdf.

Propagar (4) habría hecho ruidoso el fallo. No habría hecho **imposible** llegar ahí.

---

## Corrección — la tipografía se exige para construir el documento

### La API

```go
// Typeface es completa por construcción: las cuatro caras son parámetros,
// así que ninguna se puede olvidar.
type Typeface struct{ /* campos sin exportar */ }

// Carga y valida antes de que exista documento alguno.
func LoadTypeface(regular, bold, italic, boldItalic string) (Typeface, error)

// Un documento no se puede construir sin tipografía utilizable.
func NewDocument(t Typeface) *Document

// Documentos con varias tipografías: el registro devuelve un handle opaco.
// No hay forma de nombrar una tipografía que no se añadió.
func (d *Document) AddTypeface(t Typeface) TypefaceID
func (d *Document) Use(id TypefaceID) *Document
```

### Por qué cierra el agujero

- **El `string` desaparece del punto de selección.** `Use` sólo acepta un `TypefaceID`, y
  el único origen de un `TypefaceID` es `AddTypeface`. Nombrar una tipografía inexistente
  deja de compilar; la clase de error «undefined font» se extingue.
- **El valor cero es seguro.** `TypefaceID{}` es la tipografía primaria —la que
  `NewDocument` exigió—, que siempre existe. Escribir nada da el estado correcto
  (principio 8), en vez del Latin-1 de hoy.
- **Desaparece el orden obligatorio.** No hay `Load` que recordar: `LoadTypeface` carga
  antes, y su `error` es un valor de retorno normal, no un callback. No hay
  `SetDefaultFont` que olvidar: sin tipografía no hay documento.
- **Una sola forma de cada cosa** (principio 4): una de cargar, una de añadir, una de
  seleccionar.

### Se elimina, sin sustituto ni deprecación

```
RegisterFont(family, path)          RegisterFontStyle(family, style, path)
SetDefaultFont(family)              SetFont(family, size)
Load(func(error))                   DefaultFontPath
loadDefaultFont()                   Theme.FontFamily
```

`SetFont` se parte en lo que de verdad hacía: `Use(id)` para la cara y `SetSize(pt)` para
el tamaño. `Theme.FontFamily` sale del `Theme` porque nombra una familia por string — la
tipografía viaja en el documento, no en el tema.

`RegisterImage(name, path)` tiene exactamente el mismo agujero (`name string` sin tipar) y
debe recibir el mismo tratamiento —`ImageID` opaco— en este mismo cambio: arreglar sólo la
mitad deja la API inconsistente, que es lo que el principio 4 prohíbe.

### Restricciones WASM/TinyGo respetadas

- **Sin mapas.** Las tipografías se guardan en un slice y `TypefaceID` es su índice. Un
  `map[string]*Typeface` sería la implementación obvia y está vetada en el ecosistema por
  el tamaño que añade al binario.
- **Sin retrocompatibilidad.** Los símbolos de la lista de arriba se borran, no se marcan
  deprecados. Un alias que nadie llama sigue enlazando código.
- **Menos binario que hoy.** Se va la búsqueda familia-por-string de la capa `Document`,
  con su `Convert().ToLower()` y su tabla de core fonts asociada.
- **`readFile` puede ser función de paquete.** Ni la versión backend (`os.ReadFile`) ni la
  WASM (`fetch`, que bloquea sobre un canal — `env.front.go:50`) usan estado del
  `Document`. Ese bloqueo es lo que permite que `LoadTypeface` devuelva `error` en vez de
  arrastrar el callback.

### Alcance del cambio

El trabajo real no son los cuatro métodos públicos, sino los puntos internos que hoy leen
`d.fontFamily` como string y que pasarán a resolver la cara desde el `TypefaceID` activo:

| Archivo | Puntos |
|---|---|
| `element.go` | `:38`, `:125`, `:146` — las tres rutas de dibujo de texto |
| `chart_bar.go` | `:67`, `:112` |
| `chart_pie.go` | `:60`, `:108` |
| `chart_line.go` | `:58` |

Los ocho llaman `internal.SetFont(d.fontFamily, style, size)`. Ese es el punto único donde
hoy se materializa el fallo silencioso, y donde el handle tipado lo vuelve imposible.

Consumidores externos a actualizar en el mismo cambio, ambos con el mismo patrón de cuatro
`RegisterFontStyle`:

- `veltylabs/cotizaciones/print/doc.go:16-19`
- `veltylabs/contracts/print/doc.go:16-19`

Los dos apuntan `"I"`/`"BI"` a los archivos rectos, así que hoy no tienen cursiva real. Con
`LoadTypeface(regular, bold, italic, boldItalic)` esa carencia deja de ser invisible: son
cuatro parámetros distintos, y pasar el mismo archivo dos veces se ve al escribirlo.

---

## Defecto 2 — Conviven dos sistemas de fuentes, y uno no sirve

### Qué hay en `fpdf/fonts/`

El directorio mezcla dos linajes incompatibles. Sólo uno produce texto UTF-8:

| Formato | Archivos | Peso | ¿UTF-8? |
|---|---|---|---|
| `.ttf` | 20 | 4,0 MB | **sí** — `AddUTF8FontFromBytes` lee `glyf`/`loca` |
| `.z` | 5 | 410.893 B | no — datos comprimidos del linaje Type1 |
| `.map` | 19 | 86.488 B | no — tablas de codepage de 8 bits |
| `.pfb` | 2 | 81.488 B | no — Type1 binario de Adobe |
| `.afm` | 2 | 22.339 B | no — métricas Type1 |
| `.json` | 6 | 10.709 B | no — definiciones de `makefont` |
| `.otf` | 1 | 258.992 B | **no** — ver abajo |

**870.909 B (850 KB), el 18 % del directorio, no puede generar texto acentuado.**

`Inter-Regular.otf` merece mención aparte: su cabecera es `OTTO`, es decir contornos CFF,
sin tabla `glyf`. El parser UTF-8 de gofpdf lee `glyf`/`loca`, así que **ese archivo no es
cargable por ninguna ruta del paquete**. Tiene 0 referencias en Go. Es peso muerto puro, y
un `.otf` junto a un `.ttf` del mismo nombre es exactamente la ambigüedad que el principio 4
prohíbe.

### No son sólo archivos: es un subsistema

La ruta legacy vive en código de librería, no en tests:

| Símbolo | Sitio | Qué es |
|---|---|---|
| `AddFontFromBytes(family, style, jsonBytes, zBytes)` | `fpdf/fonts.go:28` | carga JSON + Z |
| `AddFont(family, style, file)` | `fpdf/fonts.go:429` | carga desde disco |
| `AddFontFromReader` | `fpdf/fonts.go:182` | idem vía `io.Reader` |
| `SetFontLocation(dir)` | `fpdf/fonts.go:570` | directorio global mutable |
| `MakeFont(...)` | `fpdf/font.go:377` | genera los `.json`/`.z` |
| `coreFonts` | `fpdf/fpdf.go:127`, `def.go:434` | **`map[string]bool`** |
| paquete `makefont` | `fpdf/makefont/` | utilidad que produce el formato legacy |
| paquete `internal/files` | `fpdf/internal/files/` | **carpeta `internal/`** |

Dos de esas líneas son violaciones directas del harness:

- **`coreFonts` es un `map[string]bool`.** Los mapas están vetados en el ecosistema por lo
  que añaden al binario de TinyGo. Y no es un mapa cualquiera: **es el mecanismo del
  Defecto 1**. `SetFont` consulta `coreFonts` y por eso puede degradar a Helvetica Latin-1
  en silencio (`fpdf/fonts.go:302`).
- **`internal/files/`** — «No `internal/` folders. They are the signature of a forked or
  duplicated dependency instead of a contribution upstream.»

### Por qué es el mismo refactor que el Defecto 1

Al eliminar el linaje legacy desaparece `coreFonts`, y sin `coreFonts` **no existe destino
al que degradar**. La corrección del Defecto 1 deja de necesitar una comprobación en
runtime: no hay fallback silencioso porque no hay fallback. Los dos defectos son el mismo
problema —dos sistemas de fuentes conviviendo— visto desde dos alturas.

### Alcance

**Assets a borrar:** todos los `.json`, `.z`, `.pfb`, `.afm`, `.map` y el `.otf`.

**Assets `.ttf` sin referencias, mismo criterio:** `Arial_Black`, `Arial_Bold`,
`Arial_Italic`, `Arial_Bold_Italic` (834.472 B, 0 referencias en Go). El único Arial con uso
es `Arial.ttf`, y desaparece junto con `DefaultFontPath` según el Defecto 1.

**A revisar antes de decidir:** los cuatro `DejaVuSansCondensed*.ttf` pesan **2,4 MB** —más
de la mitad del directorio— con una sola referencia. Son UTF-8 válidos, así que no caen por
este criterio, pero conviene confirmar si esa referencia justifica el peso o si se
sustituyen por un subset como los de Inter/Roboto (21-25 KB por cara).

**Código a borrar:** los símbolos de la tabla de arriba, el paquete `makefont`, el paquete
`internal/files`, y los tests que los ejercitan (`fpdf_test.go`, `issues_test.go`,
`ttfparser_test.go`, `makefont_test.go` y `contrib/ghostscript`, que llama
`SetFont("Calligrapher", ...)`).

Sin deprecación ni shims: el paquete compila a WASM y un símbolo exportado que nadie llama
sigue enlazando su código.

---

## Defecto 3 — Los PDF de inspección se quedan obsoletos sin avisar

### Síntoma

Con `tests/` en verde y sin cambios en sus fuentes, borrar `tests/test_fonts.pdf` y correr
`gotest` **no lo regenera**. El runner informa `tests ✅` y el archivo no reaparece.

### Causa

Go cachea los resultados de test exitosos con clave en fuentes y argumentos. Si nada
cambió, devuelve el veredicto cacheado **sin ejecutar el test**. El veredicto se cachea; el
efecto lateral que produce el PDF, no.

### Por qué importa

Es el defecto caro, y es un fallo silencioso en el sentido del principio 6. Durante la
sesión que originó este plan se inspeccionó varias veces un `test_fonts.pdf` de una versión
anterior creyéndolo actual: el runner decía verde, el archivo no cambiaba, y el diagnóstico
apuntó un buen rato a un bug de fuentes inexistente.

### Corrección

Un test cuyo entregable es un archivo para revisión humana no debe poder servirse desde la
caché. La vía barata y alineada: **que el test verifique su propio artefacto**. Un
`os.Stat` tras `WritePdf` convierte la ausencia en fallo, y los fallos no se cachean.
`TestFonts_AllStyles` ya lo hace; `api_test.go`, `chart_test.go` y `table_test.go` no.

---

## Verificación

1. **`tests/encoding_test.go` deja de compilar.** Es el criterio de éxito, no un problema:
   `TestUnselectedFamilyIsReportedNotSilentlyDowngraded` sólo puede existir mientras el
   estado ilegal sea representable. Se borra en el commit que cierra el harness, y su
   desaparición **es** la prueba.
2. **`tests/fonts_test.go` compila y pasa sin tocar su cuerpo salvo la construcción.** Es
   el test *consumer-shaped* que exige el harness («An API is not published until a
   consumer-shaped test, inside the library itself, proves it»): usa tres tipografías
   reales, los cuatro estilos, y escribe un PDF que se mira. Si la nueva API lo vuelve
   incómodo de escribir, la API es incómoda de usar.
3. **El test ácido.** Un agente sin contexto, guiado sólo por autocompletado, debe producir
   un documento con texto acentuado correcto. Hoy produce el bug: `NewDocument()` no pide
   nada y `AddText` acepta cualquier cosa.
4. Borrar todos los `tests/*.pdf`, correr `gotest`, y que reaparezcan.
5. **`fpdf/fonts/` contiene sólo `.ttf` y las licencias.** Cualquier `.json`, `.z`, `.pfb`,
   `.afm`, `.map` u `.otf` que sobreviva es una ruta legacy que no se cerró.
6. **`grep -rn "map\[" fpdf/` no encuentra `coreFonts`.** Es la comprobación de que el
   fallback silencioso desapareció por construcción y no por un `if`.
7. **No existe `fpdf/internal/`.**
8. `gotest` en verde.
