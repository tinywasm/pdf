---
PLAN: "fix: una cara ausente falla nombrando el archivo, no se sustituye en silencio"
TAG: v0.1.2
---
## Antes de escribir código: lee [CONSTRUCTION_HARNESS.md](CONSTRUCTION_HARNESS.md)

**Es vinculante, no orientativo.**

| # | Principio | Cómo se aplica aquí |
|---|---|---|
| 6 | Fail at compile time, not runtime | Y si el fallo es inevitable en runtime, que **sea** un fallo: no un documento mal compuesto. |
| 9 | Lego pieces, never forks | `LoadDeclared` está parcheando la derivación de `tinywasm/font` con nombres alternativos. Eso es un fork con nombre amable. |
| 4 | One way to do each thing | Una cara, un archivo, un nombre. |

**Prerrequisito: `tinywasm/font` v0.1.0 publicado** (`font/docs/PLAN.md`), donde
`Face(Regular)` pasa a devolver `Roboto-Regular`. Sin eso, borrar el fallback de la línea
41 rompe la carga.

---

## 1. Tres sustituciones silenciosas en `LoadDeclared`

```go
// document.go:41
regPathAlt := dir + string(f) + "-Regular.ttf"   // si Face(Regular) no existe

// document.go:58
it, err = readFile(regPath)                      // si falta la itálica → carga la RECTA

// document.go:72
bi, err = readFile(bldPath)                      // si falta la negrita itálica → la NEGRITA
```

Las tres devuelven `nil` como error. `docs/SPECS.md:6` las documenta como una capacidad:
«supports automatic style fallbacks».

No lo son, y cada una falla de una manera distinta:

**La línea 41** acepta un segundo nombre válido para la misma cara. Mientras `pdf` fuera
el único lector daba igual; ahora `assetmin` va a copiar y publicar esos mismos archivos
usando **una sola** derivación, así que el PDF encontrará la fuente y la web no — o al
revés. El defecto real está en `font` y allí se arregla; aquí sólo hay que dejar de
taparlo.

**Las líneas 58 y 72** son peores porque no son un problema de nombres: producen un
documento **sin cursivas** y nadie se entera. No hay error, no hay log, y el resultado
parece correcto hasta que alguien compara con el diseño. El comentario justifica ambas
«for fonts like DroidSans that don't have italics» — DroidSans está descontinuada desde
Android 4.0, no tiene glifo `€` y ya fue excluida del ecosistema por eso mismo. El caso
de uso que justificaba el fallback ya no existe.

Un documento al que le falta una cara no es un documento degradado: es un producto roto.
Y el navegador, si lo dejamos llegar hasta el `@font-face`, sintetizará la cursiva
inclinando la recta — que es justo lo que la tipografía real evita.

---

## 2. Cambios

1. **`document.go`, `LoadDeclared`**: borrar las tres ramas de fallback (41-47, 57-67,
   71-77). Las cuatro caras se leen igual: `dir + f.Face(s) + ".ttf"`, y el primer fallo
   devuelve error.
2. **El error nombra el archivo.** `readFile` devuelve el error del sistema; envolverlo
   con la ruta completa que se intentó y la cara que faltaba, para que el mensaje diga qué
   archivo poner y dónde. Es la única pista que el dev va a tener.
3. **`go.mod`**: `github.com/tinywasm/font` a v0.1.0.
4. **`docs/SPECS.md:6`**: la línea de «automatic style fallbacks» se sustituye por la
   regla nueva — las cuatro caras son obligatorias y su ausencia es un error que las
   nombra.
5. **`fpdf/fonts/`**: verificar que las cuatro caras de Roboto e Inter coinciden con la
   derivación nueva. **No hay que renombrar nada** — ya se llaman `-Regular`.

### Lo que NO se toca

- **`Typeface` y `NewDocument` no cambian de firma.** Esto no es un rediseño: es borrar
  código que oculta fallos.
- **DejaVu, calligra y DroidSans siguen en `fpdf/fonts/`** como material de test; que
  DroidSans no tenga cursiva deja de ser un caso a soportar y pasa a ser, si se quiere,
  un caso de test del error.
- **`cmd/demo/`** está en `.gitignore` (patrón `demo`), así que ningún commit lo alcanza.
  Usa `font.Declare("Roboto", "fpdf/fonts/")` y seguirá funcionando: esas caras ya llevan
  el nombre nuevo.

---

## 3. Verificación

1. Con las cuatro caras presentes, `LoadDeclared` carga igual que hoy y el PDF sale
   idéntico.
2. Faltando `Roboto-Italic.ttf`, `LoadDeclared` **devuelve error** y el mensaje contiene
   esa ruta. Hoy devuelve `nil` y un documento sin cursivas: ése es el test que prueba el
   cambio.
3. Faltando `Roboto-BoldItalic.ttf`, lo mismo.
4. `grep -n "Alt\|Fallback" document.go` no devuelve nada.
5. Un documento con texto en cursiva **renderiza con la cara itálica real**, no con la
   recta inclinada por el motor.
6. `gotest` — incluido `vet`, que es lo que atrapa el `cmd/demo` roto si se rompiera.
