# PLAN — `tinywasm/font` como única fuente de verdad, y cerrar el linaje Type1

## Antes de escribir código: lee [CONSTRUCTION_HARNESS.md](CONSTRUCTION_HARNESS.md)

**Es vinculante, no orientativo.** Los principios que gobiernan este trabajo:

| # | Principio | Cómo se aplica aquí |
|---|---|---|
| 1 | Typed over `any` | Las cuatro rutas de `LoadTypeface` son `string` sueltos. Pasan a derivarse de un tipo. |
| 4 | One way to do each thing | Sobreviven dos sistemas de fuentes: el UTF-8 nuevo y el linaje Type1/AFM. Queda uno. |
| 9 | Lego pieces, never forks | La identidad tipográfica la posee `tinywasm/font`. Este módulo la consume, no la reinventa. |

Depende de `github.com/tinywasm/font` **v0.0.3**, ya publicado.

---

## 1. De dónde viene esto

`v0.1.0` cerró el harness de tipografía: `LoadTypeface` + `NewDocument(t)` +
`AddTypeface`/`Use`, `coreFonts` eliminado, `fpdf/internal/` borrado, los 870 KB de
assets no-UTF8 fuera, y 12.082 líneas menos. El fallo silencioso que escribía `Señor`
como `SeÃ±or` ya no es representable.

Quedan cuatro cosas: dos que el plan anterior dejó sin terminar, y dos que no miraba
—los gráficos y el color, que nunca fueron responsabilidad de este módulo.

---

## 2. Defecto 1 — El nombre de la cara se sigue escribiendo a mano

### Síntoma

```go
tf, err := pdf.LoadTypeface(
    "fonts/Roboto-Regular.ttf",
    "fonts/Roboto-Bold.ttf",
    "fonts/Roboto-Italic.ttf",
    "fonts/Roboto-BoldItalic.ttf",
)
```

Cuatro rutas tecleadas. Nada impide escribir `Roboto-Italic.ttf` dos veces —los
consumidores lo hacían con DroidSans, y por eso no tenían cursiva—, ni que el CSS de la
misma aplicación declare otra familia.

**El objetivo del proyecto era que la tipografía se decidiera una vez.** Mientras el
nombre se escriba aquí, hay dos sitios donde escribirlo: éste y `config/css.go`.

### Qué principios rompe

| Principio | Cómo |
|---|---|
| 1 · Typed over `any` | Una ruta es un `string`: cualquier cadena compila, sólo algunas existen. |
| 4 · One way to do each thing | La familia se nombra aquí **y** en el CSS. Dos orígenes que pueden divergir. |

### Corrección

`tinywasm/font` ya deriva los nombres, y ya está publicado:

```go
d := font.Declare("Roboto", "fonts/")
d.Family().Face(font.Bold)   // "Roboto-Bold"
```

Este módulo añade el puente:

```go
// LoadDeclared carga las cuatro caras que la declaración nombra.
// La extensión .ttf la pone este paquete: es del medio, no de la identidad.
func LoadDeclared(d font.Declaration) (Typeface, error)
```

Con eso, `config/fonts.go` es el único sitio donde vive el nombre. `css` toma la
familia; `pdf` deriva las caras; ninguno lo teclea.

### Qué pasa con `LoadTypeface`

**Se elimina.** No se conserva "por flexibilidad": dos formas de cargar una tipografía
es exactamente lo que el principio 4 prohíbe, y la que se queda debe ser la que no
permite escribir mal un nombre.

Si un caso legítimo necesitara caras con nombres arbitrarios, la respuesta es ampliar
`font.Declaration` —donde vive la identidad—, no reabrir un camino de strings aquí.

### Alcance

- `document.go:27` — `LoadTypeface` → `LoadDeclared`.
- `go.mod` — añadir `github.com/tinywasm/font v0.0.3`.
- `web/client.go:12` — el consumidor WASM del propio repo.
- `tests/fonts_test.go` — hoy construye las rutas con un `struct` local de tres campos
  por familia; pasa a declarar y derivar.
- Consumidores externos: `veltylabs/cotizaciones/print/doc.go`,
  `veltylabs/contracts/print/doc.go`.

---

## 3. Defecto 2 — El linaje Type1/AFM sigue vivo a medias

`v0.1.0` borró los paquetes que alimentaban el sistema viejo —`makefont`,
`internal/files`, los `.json`/`.z`/`.pfb`/`.afm`/`.map`— pero **no las funciones que los
consumían**:

| Símbolo | Sitio |
|---|---|
| `MakeFont` | `fpdf/font.go:377` |
| `AddFontFromBytes` | `fpdf/fonts.go:28` |
| `AddFontFromReader` | `fpdf/fonts.go:182` |
| `AddFont` | `fpdf/fonts.go:403` |
| `SetFontLocation` | `fpdf/fonts.go:544` |
| el parser AFM completo | `fpdf/font_afm.go` (212 líneas) |

Son código muerto que aún enlaza: `MakeFont` genera un formato cuyos consumidores ya no
existen, y `AddFont` lee de un directorio que ninguna ruta configura. La medida lo
confirma — el `client.wasm` sólo bajó de 3.824.043 a **3.785.258 B**, un 1%, cuando se
borraron 12.000 líneas.

Es el principio 4 otra vez: **dos sistemas de fuentes conviviendo**, y uno no funciona.

### Corrección

Borrar los seis, sin deprecación. Este módulo compila a WASM y un símbolo exportado que
nadie llama sigue enlazando su código.

**Verificación con número, no con opinión:** medir `web/public/client.wasm` antes y
después. Si no baja de forma apreciable, el linaje no se fue del todo.

---

## 4. Defecto 3 — Los gráficos no son responsabilidad de este módulo

`chart.go`, `chart_bar.go`, `chart_line.go` y `chart_pie.go` suman 398 líneas. Que la
funcionalidad exista está bien —hay documentos que necesitan gráficos—; que viva aquí,
no: este módulo genera documentos, y dibujar una tarta con porcentajes es otro concern.

Y hay un coste medible. La fábrica cuelga del documento:

```go
func (d *Document) Chart() *ChartFactory
func (f *ChartFactory) Bar() *BarChart   // …Line(), Pie()
```

Alcanzar `Bar()` obliga a enlazar `ChartFactory`, y con ella `Line` y `Pie`. **Un
informe que sólo dibuja barras carga los tres**, y en un binario WASM eso se paga en
cada carga de página.

### Corrección

Se trasladan a `github.com/tinywasm/chart`, ya creado, con el contrato en la raíz y un
subpaquete por tipo. Su plan es `chart/docs/PLAN.md`. Aquí toca:

1. **Exportar la superficie de dibujo.** Los gráficos usan 20 métodos de `d.internal`
   —`Rect`, `Circle`, `ArcTo`, `DrawPath`, `SetFillColor`, `Text`…— y
   `getActiveFontName()`, todos privados. Un módulo aparte no los alcanza, y eso es un
   contrato que falta *aquí*, no un problema del consumidor.

   **Restricción:** la superficie **no puede** exponer `SetFont(familyStr, styleStr, …)`.
   `v0.1.0` eliminó los nombres de fuente en `string` precisamente porque cualquier
   cadena compilaba; reintroducirlos en el contrato de dibujo reabre el agujero por la
   puerta de atrás. La familia ya la fijó `NewDocument(Typeface)`: un gráfico sólo
   necesita pedir «negrita, 12pt».

2. **Borrar los cuatro archivos y `(*Document).Chart()`.** Sin deprecación.
3. **`tests/chart_test.go`** se traslada al módulo nuevo.

---

## 5. Defecto 4 — El color se parsea aquí y en `css`

`pdf/color.go:11` tiene `Color.parse()`. `css/contrast_test.go:20-22` tiene su propio
parser con `strconv.ParseUint`, tres veces. Dos implementaciones del mismo cálculo en el
mismo ecosistema — principio 4.

`github.com/tinywasm/color` **ya existe** (v0.0.1) con el tipo trasladado desde aquí,
exportado como `RGB()` y con tests. Falta el lado de acá.

### Corrección

1. **Borrar `Color` y `parse()` de `pdf/color.go`** e importar
   `github.com/tinywasm/color`.
2. **Sustituir los usos por `color.Color`. Sin alias.**

   Una versión anterior de este plan proponía `type Color = color.Color` «para no tocar
   123 referencias». Las dos mitades de esa frase estaban mal:

   - **No son 123, son 20.** Aquella cifra contaba `SetFillColor`, `TextColor` y otros
     nombres de método. Los usos reales del tipo son 41, y de ésos 8 están en el
     `color.go` que se borra y 7 en los `chart_*.go` que se van a `tinywasm/chart`.
     Quedan **20, en cuatro archivos**: `element.go` (10), `document.go` (5),
     `table_layout.go` (4) y `examples/contract_layout/main.go` (1). Eso es un
     buscar-y-reemplazar, no un proyecto de migración.
   - **Un alias deja dos nombres para el mismo tipo.** `pdf.Color` y `color.Color`
     conviviendo es literalmente el ítem del checklist *«More than one way to do the
     same thing → collapse to one»*. Y este módulo acaba de romper su API pública sin
     shims por principio: un alias **es** un shim, sólo que con mejor prensa.

3. **`Theme` NO se mueve.** Lleva `Sizes` (H1/H2/H3), `Spacing`, `Page` y `Margin` en
   **milímetros**: es el tema de un documento, no color. Que contenga colores no lo
   convierte en uno, y llevarlo allí metería tamaños de página en una librería de color.

### Falsos positivos al buscar

`fpdf/drawing.go` y `fpdf/def.go` aparecen si se busca `Color`, pero **no usan este
tipo**: son un `"ColorDodge"` dentro de un comentario y los tipos `spotColorType` /
`cmykColorType`, que son del modelo CMYK interno de fpdf. No tocarlos.

`fpdf/png.go` y `fpdf/spotcolor.go` sí manejan color, pero del stdlib. Si algún archivo
acaba importando ambos, el identificador `color` colisiona y hay que renombrar uno en el
import.

---

## 6. Documentación y demo

No es cierre opcional: `docs/SPECS.md` es la autoridad de este repo, y si el código y
SPECS discrepan, ambos están mal.

| Documento | Qué cambia |
|---|---|
| `docs/SPECS.md` | `LoadDeclared` sustituye a `LoadTypeface`; desaparecen `Chart()` y los tipos de gráfico; se van los símbolos del linaje Type1 |
| `docs/ARCHITECTURE.md` | La identidad tipográfica la posee `tinywasm/font`; los gráficos, `tinywasm/chart`. Este módulo genera documentos y nada más |
| `README.md` | El ejemplo de inicio pasa a `font.Declare` → `LoadDeclared`; indexa los documentos |

**El demo** (`cmd/demo/`) se actualiza al mismo tiempo. Ojo con un detalle que se
descubrió tarde: está en `.gitignore` (patrón `demo`, línea 41), así que **no viaja en
el repositorio** — ningún cambio de este plan lo alcanza y quien lo tenga en local lo
migra a mano. Si se quiere que el demo sea la prueba viva de la API, hay que sacarlo del
`.gitignore` primero; mientras siga ignorado no puede cumplir esa función, y conviene
decidirlo explícitamente en vez de dejarlo a medias.

`web/` sí está versionado y es hoy el único consumidor real dentro del repo: su
`client.go` debe pasar a `LoadDeclared` y su `ui/pdf.go` dejar de referirse a gráficos.

---

## 7. Lo que este plan NO hace

- **No entrega los archivos al navegador.** Es de `assetmin`
  (`assetmin/docs/PLAN.md`): posee el pipeline de assets y decide las URLs.
- **No emite el `@font-face`.** También de `assetmin`, por lo mismo.
- **No elige la tipografía.** Eso es `config/fonts.go` en cada proyecto.
- **No recorta fuentes.** Eso es `font/docs/GOFONT_PLAN.md`.

---

## 8. Verificación

1. `LoadTypeface` no existe; `LoadDeclared(font.Declaration)` sí.
2. Un documento generado desde una `Declaration` embebe las cuatro caras como fuentes
   distintas, y su texto acentuado con `€` se extrae correcto — no basta con que
   compile. `tests/fonts_test.go` ya hace esta comprobación y debe seguir pasándola.
3. `grep -rn "MakeFont\|AddFontFromBytes\|AddFontFromReader\|SetFontLocation" fpdf/`
   no encuentra nada, y `fpdf/font_afm.go` no existe.
4. `web/public/client.wasm` **baja de 3.785.258 B**. Entre el linaje Type1 residual y
   los 398 KB de gráficos hay margen real; si no baja, algo no se fue.
5. `pdf/color.go` no declara ningún tipo `Color` ni parser de hex, y **no existe
   ningún alias**: el único nombre del tipo es `color.Color`.
6. `Theme` sigue en este módulo, con sus milímetros.
7. Este módulo no exporta ningún tipo de gráfico ni el método `Chart()`.
8. Un proyecto declara `font.Declare("Roboto", "fonts/")` una vez y de ahí salen el
   `--font-sans` del CSS y las caras del PDF, sin repetir el nombre en ningún sitio.
9. `gotest` en verde, y `go build ./...` limpio.

`docs/SPECS.md` y `docs/ARCHITECTURE.md` se actualizan en el mismo commit.

---

## 9. Orden en el conjunto

| Orden | Repo | Estado |
|---|---|---|
| — | `tinywasm/font` | ✅ publicado v0.0.3 |
| — | `tinywasm/color` | ✅ v0.0.1 — el tipo ya está allí; falta que este módulo lo consuma |
| 1 | **`tinywasm/pdf`** | este documento — consume `font` y `color`, suelta los gráficos |
| 1 | `tinywasm/css` | `css/docs/PLAN.md` — `FontSans` alimentado por `font.Family` |
| 2 | `tinywasm/chart` | `chart/docs/PLAN.md` — necesita el `Canvas` que expone este plan |
| 3 | `tinywasm/assetmin` | `assetmin/docs/PLAN.md` — entrega los `.ttf` y el `@font-face` |

Los dos pasos 1 son independientes entre sí. `chart` va después porque no puede compilar
hasta que exista el `Canvas`; `assetmin` al final, cuando ambos consumidores de `font`
estén definidos.

El estado del conjunto vive en `app-releases/docs/TYPOGRAPHY_MASTER_PLAN.md`.

`tinywasm/ssr` **no participa**: `config/fonts.go` y `config/css.go` son el mismo
paquete Go, así que `RootCSS()` llama a `Fonts()` directamente y el valor viaja dentro
de la extracción que ya existe.
