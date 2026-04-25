# PLAN: Reduce WASM binary — Remaining optimizations

## Current Status: 423.5 KB (post-optimization, down from 738 KB — 43% reduction)

## Remaining target: ~390-400 KB (this round) → ~150 KB requires Round 2 (Stage 8 candidates)

---

## Post-execution twiggy analysis

| Dependency | Before | After | Status |
|---|---|---|---|
| encoding/json | 91 KB | 0 KB | Eliminated |
| time stdlib | 75 KB | ~1 KB | Eliminated |
| image/gif | 12 KB | 0 KB | Eliminated |
| compress/* | 38 KB | ~0.1 KB | Eliminated |
| stdlib fmt | 45 KB | 0 KB | Eliminated |
| crypto/* | 30 KB | **25 KB** | **Partial — attachments.go still imports crypto/md5** |
| runtime | 61 KB | 61 KB | Irreducible |

---

## STAGES

### Stage 1: Eliminate crypto/md5 from attachments.go — Savings ~25 KB
**Risk**: low | **Complexity**: low

**Context**: `fpdf/attachments.go` imports `crypto/md5` for the `checksum()` function (line 32-35).
This is used in `writeCompressedFileObject()` to embed the MD5 checksum of attachment content.
Attachments are NOT part of the WASM functional scope (generate-only, no embedded files needed in browser).

**Current code** (`fpdf/attachments.go:32-35`):
```go
func checksum(data []byte) string {
    sl := md5.Sum(data)
    return hex.EncodeToString(sl[:])
}
```

**What to do**:
1. Add `//go:build !wasm` to `fpdf/attachments.go`
2. Create `fpdf/attachments_wasm.go` (`//go:build wasm`) with:
   - The `Attachment` struct definition (needed for compilation)
   - Stub functions that set `f.err` with "attachments not supported in WASM"
   - The `annotationAttach` struct and `pageAttachments` type if referenced elsewhere

**Important**: before adding build tags, verify what symbols from `attachments.go` are referenced in other files that compile for WASM:
```bash
grep -rn "Attachment\|attachments\|pageAttachments\|annotationAttach\|putAttachments\|getEmbeddedFiles\|putAnnotationsAttachments\|putAttachmentAnnotationLinks\|AddAttachmentAnnotation\|SetAttachments" fpdf/*.go | grep -v _test.go | grep -v attachments.go
```
All referenced symbols must have stubs in the WASM file.

**Files**:
- `fpdf/attachments.go` → add `//go:build !wasm`
- `fpdf/attachments_wasm.go` → new, struct definitions + stub functions

**Validation**:
1. `wasmbuild` compiles without errors
2. `twiggy top ... | grep crypto` → 0 results
3. `go test ./...` backend tests still pass
4. Measure new size: `ls -lh web/public/client.wasm`

---

### Stage 2: Unify fontid — Eliminate duplicated code + crypto/sha1 from backend
**Risk**: low | **Complexity**: low

**Context**: `fpdf/fontid_back.go` and `fpdf/fontid_wasm.go` duplicate logic unnecessarily.
The backend version uses `crypto/sha1` + `encoding/json` for `generateFontID` and SHA1 hashing for `generateImageID`. The WASM version uses simple deterministic strings. The simple version is sufficient for both platforms — IDs only need to be unique map keys within a document session.

**Current state**:
- `fpdf/fontid_back.go` (!wasm): `crypto/sha1`, `encoding/json` — heavy dependencies for no real benefit
- `fpdf/fontid_wasm.go` (wasm): simple `Tp + "_" + Name` for fonts, `img_w_h_len(data)` for images

**Problem with WASM image ID**: `img_w_h_len(data)` can collide — two different images with same dimensions and data length get the same ID.

**What to do**:
1. Delete `fpdf/fontid_back.go` and `fpdf/fontid_wasm.go`
2. Create single `fpdf/fontid.go` (no build tags) with:
   ```go
   func generateFontID(fdt fontDefType) (string, error) {
       return fdt.Tp + "_" + fdt.Name, nil
   }
   ```
3. For `generateImageID`: use `tinywasm/unixid` — unique, thread-safe, no crypto dependency.
   ```go
   func generateImageID(info *ImageInfoType) (string, error) {
       var id string
       uid.SetNewID(&id)
       return id, nil
   }
   ```
   Where `uid` is a package-level `*unixid.UnixID` instance initialized once. IDs are unique per session, not deterministic — this is correct since image IDs only need to be unique map keys within a document lifecycle.

**Files**:
- Delete `fpdf/fontid_back.go`
- Delete `fpdf/fontid_wasm.go`
- Create `fpdf/fontid.go` — unified, no build tags, uses `tinywasm/unixid`

**Validation**:
1. `wasmbuild` compiles
2. `go test ./...` passes
3. `twiggy top ... | grep crypto/sha1` → 0 results
4. Multiple images in same PDF render correctly (no ID collisions)

---

### Stage 3: Unify fonts_json — Remove encoding/json duplication
**Risk**: low | **Complexity**: low

**Context**: `fpdf/fonts_json_back.go` uses `encoding/json.Unmarshal` and `fpdf/fonts_json_wasm.go` uses `tinywasm/json.Decode`. Since `tinywasm/json` is platform-agnostic, both can use it.

**Note**: `fpdf/font.go` also imports `encoding/json` for `MakeFont` (build tool, not runtime). That file gets `!wasm` build tag in Stage 4 and keeps `encoding/json` since it's backend-only tooling.

**What to do**:
1. Delete `fpdf/fonts_json_back.go` and `fpdf/fonts_json_wasm.go`
2. Create single `fpdf/fonts_json.go` (no build tags):
   ```go
   package fpdf

   import "github.com/tinywasm/json"

   func unmarshalFontDef(data []byte, def *fontDefType) error {
       return json.Decode(data, def)
   }
   ```

**Files**:
- Delete `fpdf/fonts_json_back.go`
- Delete `fpdf/fonts_json_wasm.go`
- Create `fpdf/fonts_json.go` — unified, no build tags

**Important**: `fontDefType` must conform to what `tinywasm/json.Decode` supports. If `fontDefType` has fields with types not supported by `tinywasm/json`, adjust the struct — the library adapts to `tinywasm/json`, not the other way around.

**Validation**:
1. `wasmbuild` compiles
2. `go test ./...` passes — fonts load correctly with tinywasm/json on backend
3. Generate PDF with multiple fonts to verify character widths (Cw) parse correctly
4. Verify all `fontDefType` fields decode correctly (compare output with `encoding/json` in a one-off test)

---

### Stage 4: Add `!wasm` build tag to font.go — Preventive
**Risk**: low | **Complexity**: low

**Context**: `fpdf/font.go` imports `encoding/json`, `compress/zlib`, `os` but currently has no build tag.
TinyGo eliminates it via dead code elimination, but this is fragile — any future call to `MakeFont` from WASM code would silently pull in ~130 KB of dependencies.

**What to do**:
1. Add `//go:build !wasm` to `fpdf/font.go`
2. Verify no WASM code calls any function from this file

**Validation**:
1. `wasmbuild` compiles
2. `go test ./...` passes

---

### Stage 5: Unify time — Remove time stdlib duplication
**Risk**: low | **Complexity**: low

**Context**: `fpdf/time_back.go` uses `time.Time` (stdlib) and `fpdf/time_wasm.go` uses `int64` (tinywasm/time). The underlying type `pdfTime` is also split: `types_back.go` defines it as `time.Time`, `types_wasm.go` as `int64`. The tinywasm ecosystem mandates `int64` (unix nano) for all time operations — this is a design constraint, not optional. Unify to `int64` only.

**Files to unify/delete**:
- Delete `fpdf/types_back.go` and `fpdf/types_wasm.go` → create `fpdf/types.go`: `type pdfTime int64`
- Delete `fpdf/time_back.go` and `fpdf/time_wasm.go` → create `fpdf/time.go` using `tinywasm/time` only:
  - API: `SetCreationDate(tm int64)`, `GetCreationDate() int64`, etc.
  - `timeOrNow(tm pdfTime) int64`: if 0 return `time.Now()`, else return `int64(tm)`
  - `formatPDFDate(tm pdfTime) string`: use `time.FormatISO8601` + string slicing (from current wasm version)

**Tests to update** (use `int64` unix nano instead of `time.Time`):
- `fpdf/getter_test.go:126,448` — `TestGetCreationDate`, `TestGetModificationDate`
- `fpdf/fpdf_test.go:2691` — `Test_SetModificationDate`
- `fpdf/exampleDir_test.go:55-56` — replace `time.Date(...)` with equivalent unix nano value

**Note**: once `tinywasm/time` has `FormatCompact` (see tinywasm/time PLAN.md), replace the ISO8601 string slicing with `time.FormatCompact(nano)`.

**Validation**:
1. `wasmbuild` compiles
2. `go test ./...` passes (tests updated to int64)
3. `grep -rn '"time"' fpdf/*.go | grep -v _test.go` → 0 results (only tinywasm/time)
4. PDF CreationDate/ModDate are correct in generated output

---

### Stage 6: Remove `os` from shared production files — Use io interfaces
**Risk**: low | **Complexity**: low

**Context**: `os` package should not be used in shared fpdf code. The library already has an injection pattern for file operations (`env.front.go`/`env.back.go`, `def.go:431 fileSize func`). Three production files still import `os`:

| File | Usage | Fix |
|---|---|---|
| `fpdf/svgbasic.go:302` | `os.ReadFile()` in `SVGBasicFileParse()` | Move file-path variant to `!wasm` file. `SVGBasicParse([]byte)` already exists in shared code |
| `fpdf/util.go:38,50` | `os.Stat()` in `fileExist()`, `fileSize()` | Only called from `font.go` (already `!wasm`). Move to `!wasm` file |
| `fpdf/list/list.go:37` | `filepath.Walk` + `os.FileInfo` | Backend-only tooling, add `!wasm` build tag |

**What to do**:
1. **svgbasic.go**: `SVGBasicFileParse` should use the injected `readFile` function (same pattern as `WriteFileFunc`/`ReadFileFunc`/`FileSizeFunc`). This requires `SVGBasicFileParse` to be a method on `Fpdf` (or receive the readFile func). Remove direct `os.ReadFile` import.
2. **util.go**: `fileExist()` and `fileSize()` are only called from `font.go` (already `!wasm` after Stage 4). Move them into `font.go` directly — no need for a separate file. Remove `os` import from `util.go`.
3. **list/list.go**: add `//go:build !wasm`.

**Files**:
- `fpdf/svgbasic.go` → replace `os.ReadFile` with injected `f.readFile` in `SVGBasicFileParse`
- `fpdf/util.go` → remove `os` import, remove `fileExist`/`fileSize`
- `fpdf/font.go` → absorb `fileExist`/`fileSize` (already `!wasm` from Stage 4)
- `fpdf/list/list.go` → add `//go:build !wasm`

**Validation**:
1. `wasmbuild` compiles
2. `go test ./...` passes
3. Verify: `grep -rn '"os"' fpdf/*.go | grep -v _test.go | grep -v _back.go` → only `font.go` (already `!wasm`)

---

### Stage 7: Verify xcompr_wasm.go imports compress/zlib — Cleanup / Verification
**Risk**: low | **Complexity**: low

**Context**: `fpdf/xcompr_wasm.go` currently imports `compress/zlib` for the `uncompress()` function.
The agent kept decompression functional in case fonts need it.
twiggy shows compress/* at only 0.1 KB, suggesting TinyGo may be eliminating most of it.

**What to do**:
1. Verify if `uncompress()` is actually called in WASM builds (grep for call sites)
2. If not called: remove `compress/zlib` import, make `uncompress()` return error
3. If called (e.g., for .z compressed fonts): keep it but document why

**Validation**:
1. `wasmbuild` compiles
2. Font loading works in WASM
3. `twiggy top ... | grep compress` → verify size impact

---

### Stage 8: Analyze remaining size for Round 2 candidates
**Risk**: N/A | **Complexity**: analysis only

After Stages 1-7, run full twiggy analysis to identify next optimization targets:
```bash
twiggy top web/public/client.wasm -n 50
```

Known Round 2 candidates (evaluate in separate plan):
1. `regexp` in `htmlbasic.go` and `ttfparser.go` → manual parsing
2. `image/jpeg` (26 KB) + `image/png` (17 KB) → decode in JS Canvas API
3. fpdf dead code: layers, gradients, spot colors, blending modes
4. Strip "function names" WASM subsection (~36 KB)
5. `web/ui.setupUI` (70 KB) — demo code, not the library

---

## Execution Order

```
Stage 1 (attachments crypto/md5) ── Main WASM savings (~25 KB)
Stage 2 (unify fontid)           ── Remove duplication + crypto/sha1
Stage 3 (unify fonts_json)       ── Remove encoding/json duplication
Stage 4 (font.go build tag)      ── Preventive
Stage 5 (unify time)             ── Remove time stdlib duplication
Stage 6 (remove os from shared)  ── Clean architecture
Stage 7 (xcompr_wasm.go verify)  ── Bug fix
Stage 8 (analysis)               ── Plan next round
```

## Validation per stage
Each stage MUST:
1. Compile: `wasmbuild`
2. Measure: `ls -lh web/public/client.wasm`
3. Analyze: `twiggy top web/public/client.wasm -n 30`
4. Functional: generate PDF with text, table, chart and image in WASM
5. Backend: `go test ./...`
