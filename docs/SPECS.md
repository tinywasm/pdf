# SPECS — TinyPDF Specifications

## 1. Font System
- `LoadDeclared(font.Declaration)` replaces the legacy `LoadTypeface` string-based API.
- All font loading must occur via `webtyp.com/font` declarations, avoiding hardcoded string names in high-level calls.
- The four faces are mandatory: `LoadDeclared` fails with an error naming the missing face and its path when any is absent.

## 2. Color System
- Color management is fully unified under `webtyp.com/color`.
- Local `Color` hex parsing has been eliminated.
- Core rendering methods and theme definitions consume `color.Color` values directly.

## 3. Graphics & Canvas
- Native charts are removed and delegatively moved to `webtyp.com/chart`.
- This package exposes a clean `Canvas` interface on `Document` so external drawing engines can safely paint vector elements, text, and apply custom styling without reintroducing untyped font names.

## 4. AFM & Type1 Lineage
- All AFM files, Type1 fonts, binary `.pfb` loaders, and `.z` metrics compressors are completely eliminated from the library and WebAssembly builds.
