# Fpdf Standard Library Migration Plan

## Objective

Migrate Fpdf library from Go standard library dependencies (`fmt`, `strconv`, `errors`, `strings`) to use exclusively `github.com/tinywasm/fmt` for maximum binary size reduction and TinyGo compatibility.

## Background

Fpdf currently imports standard library packages that significantly increase binary size, especially in WebAssembly builds. The goal is to achieve **zero standard library dependencies** for string operations, formatting, and error handling by leveraging the `github.com/tinywasm/fmt` library's comprehensive API.

## Migration Strategy

### Phase 1: Error Handling Migration (`fmt.Errorf` → `Errf`)

**Target Files:**
- [x] `spotcolor.go` (2 instances)
- [x] `drawing.go` (3 instances)
- [x] `document.go` (5 instances)
- [x] `png.go` (8 instances)
- [x] `fpdftrans.go` (4 instances)
- [x] `ttfparser.go` (12 instances)
- [x] `font.go` (8 instances)
- [x] `font_afm.go` (6 instances)
- [x] `utf8fontfile.go` (2 instances)
- [x] `fonts.go` (2 instances)
- [x] `svgbasic.go`
- [x] `htmlbasic.go`
- [x] `grid.go`
- [x] `embedded.go`
- [x] `makefont.go`
- [ ] `gofpdi/helper.go` 
- [ ] `gofpdi/importer.go` 
- [ ] `gofpdi/writer.go` 




**Migration Pattern:**
```go
// BEFORE
err = fmt.Errorf("invalid format: %s", value)

// AFTER  
err = Errf("invalid format: %s", value).Error()
```

**Implementation Steps:**
1. Add `github.com/tinywasm/fmt` import: `import . "github.com/tinywasm/fmt"`
2. Replace all `fmt.Errorf()` calls with `Errf().Error()`
3. Remove `"fmt"` import when no other fmt usage remains
4. Test functionality to ensure error messages remain consistent

### Phase 2: String Formatting Migration (`fmt.Sprintf` → `Sprintf`)

**Target Files:**
- `util.go` (1 instance)
- `document.go` (1 instance)
- `fonts.go` (1 instance)
- `gofpdi/` subdirectory (multiple instances)

**Migration Pattern:**
```go
// BEFORE
result := fmt.Sprintf("Hello %s, count: %d", name, count)

// AFTER
result := Sprintf("Hello %s, count: %d", name, count)
```

**Implementation Steps:**
1. Replace `fmt.Sprintf()` with `Sprintf()`
2. Replace `fmt.Sprint()` with `Convert().String()`
3. Test format specifiers compatibility (%s, %d, %f, %x, etc.)

### Phase 3: Debug Output Migration (`fmt.Printf` → Alternative)

**Target Files:**
- `fonts.go` (1 instance)
- `utf8fontfile.go` (multiple instances)
- Test files (optional, can remain for debugging)

**Migration Pattern:**
```go
// BEFORE
fmt.Printf("Debug: %s\n", message)

// AFTER - Option 1: Remove debug output
// (Remove entirely for production)

// AFTER - Option 2: Use println for critical debugging
println(Sprintf("Debug: %s", message))
```

### Phase 4: String Parsing Migration (`fmt.Sscanf` → Custom Parser)

**Target Files:**
- `util.go` (1 instance)
- `font.go` (1 instance)

**Migration Pattern:**
```go
// BEFORE
_, err = fmt.Sscanf(lineStr, "!%2X U+%4X %s", &cPos, &uPos, &nameStr)

// AFTER - Custom parsing using github.com/tinywasm/fmt
parts := Convert(lineStr).Split()
if len(parts) >= 3 && Contains(parts[0], "!") {
    // Parse hex values and strings manually
    cPos, err1 := Convert(parts[0][1:]).Int() // Remove "!" prefix
    uPos, err2 := Convert(parts[1][2:]).Int() // Remove "U+" prefix  
    nameStr = parts[2]
    // Handle errors appropriately
}
```

### Phase 5: String Operations Migration (if any `strings` package usage)

**Migration Patterns:**
```go
// strings.Contains
// BEFORE
if strings.Contains(text, "pattern") { }
// AFTER  
if Contains(text, "pattern") { }

// strings.Replace
// BEFORE
result := strings.Replace(text, "old", "new", -1)
// AFTER
result := Convert(text).Replace("old", "new").String()

// strings.Split
// BEFORE
parts := strings.Split(text, ",")
// AFTER
parts := Convert(text).Split(",")

// strings.Join
// BEFORE
result := strings.Join(slice, ",")
// AFTER
result := Convert(slice).Join(",").String()

// strings.ToLower/ToUpper
// BEFORE
result := strings.ToLower(text)
// AFTER
result := Convert(text).Low().String()
```

### Phase 6: Type Conversion Migration (`strconv` → `github.com/tinywasm/fmt`)

**Migration Patterns:**
```go
// strconv.Itoa
// BEFORE
str := strconv.Itoa(42)
// AFTER
str := Convert(42).String()

// strconv.Atoi
// BEFORE
num, err := strconv.Atoi("42")
// AFTER
num, err := Convert("42").Int()

// strconv.ParseFloat
// BEFORE
f, err := strconv.ParseFloat("3.14", 64)
// AFTER
f, err := Convert("3.14").Float64()

// strconv.FormatFloat
// BEFORE
str := strconv.FormatFloat(3.14159, 'f', 2, 64)
// AFTER
str := Convert(3.14159).Round(2).String()
```

### Phase 7: Import Cleanup

**Final Step:**
1. Remove all unused standard library imports:
   - `"fmt"`
   - `"strconv"`  
   - `"errors"`
   - `"strings"`
2. Ensure only essential imports remain (io, os, etc.)
3. Add single tinywasm/fmt import: `import . "github.com/tinywasm/fmt"`

## File-by-File Migration Checklist

### Core Files (Priority 1)
- [x] `util.go` - Remove `fmt.Sprintf`
- [x] `document.go` - Replace 5x `fmt.Errorf`, 1x `fmt.Sprintf`
- [x] `fonts.go` - Replace 2x `fmt.Errorf`, 1x `fmt.Printf`, 1x `fmt.Sprintf`
- [x] `drawing.go` - Replace 3x `fmt.Errorf`
- [x] `spotcolor.go` - Replace 2x `fmt.Errorf`

### Parser Files (Priority 2)  
- [x] `ttfparser.go` - Replace 12x `fmt.Errorf`
- [x] `font_afm.go` - Replace 6x `fmt.Errorf`
- [x] `font.go` - Replace 8x `fmt.Errorf`, 1x `fmt.Sscanf`, 1x `fmt.Fprintf`
- [x] `utf8fontfile.go` - Replace 2x `fmt.Errorf`, multiple `fmt.Printf`

### Image & Graphics (Priority 3)
- [x] `png.go` - Replace 8x `fmt.Errorf`
- [x] `fpdftrans.go` - Replace 4x `fmt.Errorf`

### External Dependencies (Priority 4)
- [ ] `gofpdi/` subdirectory - Multiple `fmt.Sprintf` instances
- [ ] `env/` subdirectory - Multiple `fmt.Errorf` instances

## Testing Strategy

### Functional Testing
1. Run existing test suite after each file migration
2. Verify error messages maintain same semantic meaning
3. Test format output matches previous fmt.Sprintf behavior
4. Validate numeric conversions produce identical results

### Binary Size Validation
```bash
# Before migration
go build -ldflags="-s -w" -o tinypdf-before ./cmd/example
ls -la tinypdf-before

# After migration  
go build -ldflags="-s -w" -o tinypdf-after ./cmd/example
ls -la tinypdf-after

# Calculate reduction
echo "Size reduction: $(($(stat -c%s tinypdf-before) - $(stat -c%s tinypdf-after))) bytes"
```

### TinyGo Compatibility Testing
```bash
# Test WebAssembly build
tinygo build -o main.wasm -target wasm ./cmd/example

# Test embedded target
tinygo build -target arduino ./cmd/example
```

## Expected Benefits

### Binary Size Reduction
- **WebAssembly**: ~50-70% reduction in binary size
- **Native binaries**: ~20-40% reduction depending on usage
- **Embedded targets**: Maximum compatibility with TinyGo constraints

### Performance Improvements  
- Reduced memory allocations through fmt's buffer pooling
- Faster string operations optimized for small devices
- Predictable memory usage patterns

### Multilingual Support
- Built-in error message translations (9 languages)
- Consistent error formatting across the library
- Future-proof internationalization support

## Risk Mitigation

### Compatibility Risks
- **Format specifier differences**: Test all `Sprintf()` calls against original `fmt.Sprintf()`
- **Error message changes**: Verify client code doesn't depend on exact error text
- **Numeric precision**: Validate floating-point conversion accuracy

### Migration Risks
- **Incremental approach**: Migrate one file at a time to isolate issues
- **Comprehensive testing**: Run full test suite after each change
- **Rollback capability**: Maintain git commits for easy rollback

## Success Criteria

1. ✅ Zero standard library imports for string/format/error operations
2. ✅ All existing tests pass without modification
3. ✅ Binary size reduction of minimum 20% (50%+ for WebAssembly)
4. ✅ Full TinyGo compatibility across all supported targets
5. ✅ No performance regression in critical paths
6. ✅ Maintains API compatibility for end users

## Implementation Timeline

### Week 1: Core Infrastructure
- Phase 1: Error handling migration (util.go, document.go, fonts.go)
- Phase 2: String formatting migration

### Week 2: Parser & Graphics
- Phase 3: Debug output cleanup  
- Phase 4: Custom parsing implementation
- Phase 5: String operations migration

### Week 3: Testing & Validation
- Phase 6: Type conversion migration
- Phase 7: Import cleanup
- Comprehensive testing and binary size validation

### Week 4: Documentation & Release
- Update documentation and examples
- Performance benchmarking
- Release preparation

---

**Note**: This migration eliminates ALL standard library dependencies for string operations, making Fpdf a truly minimal-footprint PDF library optimized for WebAssembly and embedded deployments.

### Update: Professional Layout API & Font Refactor
- Replaced hardcoded "Arial" with dynamic `fontFamily` in all components (Charts, Tables, Text, Headers, Footers).
- Refactored resource registries (`fonts`, `images`) from `map[string]string` to `[]KeyValue` for TinyGo compatibility.
- Implemented `DrawLineH` and enhanced `PageFooter` with `leftText` and standard "Página X/{nb}" numbering.
- Fixed WASM compilation by adding `//go:build !wasm` to font generation tools and tests.
- Eliminated `strings` package dependency from `document.go`.
- Implemented a new Flow-First Layout API centered around `TableBuilder`, `Element`, and `Theme`.
- Renamed all absolute positioning methods to be package-private to encourage flow-mode usage.
- Standardized color handling using the `Color` hex-string alias.
