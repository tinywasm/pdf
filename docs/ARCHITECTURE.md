# ARCHITECTURE — Ecosistema TinyPDF

```
                   +------------------------+
                   |   github.com/webtyp  |
                   |      (color, font)     |
                   +-----------+------------+
                               |
                               v
                   +------------------------+
                   |   github.com/webtyp  |
                   |         (pdf)          |
                   +-----------+------------+
                               ^
                               | import (usa pdf.Canvas)
                               |
                   +------------------------+
                   |   github.com/webtyp  |
                   |        (chart)         |
                   +------------------------+
```

La flecha va de `chart` hacia `pdf`, no al revés: `pdf` declara e implementa `Canvas`
(`canvas.go`, con `var _ Canvas = (*Document)(nil)` como guardia de compilación) sin
saber que `chart` existe. `chart` importa `pdf` y construye sus gráficos sobre
`pdf.Canvas`. `pdf` nunca importa `chart` — de hacerlo, crearía un ciclo.

## Principios de Diseño

1. **Identidad Tipográfica Única:** La identidad tipográfica la posee y gestiona `webtyp/font`. Este módulo únicamente consume y activa esas tipografías, garantizando consistencia completa entre Web y PDF.
2. **Modularidad y Responsabilidades Separadas:** El motor de PDF se enfoca exclusivamente en la maquetación y generación del documento (`Document` y `TableBuilder`). Los gráficos se delegan al módulo externo `webtyp/chart`.
3. **Superficie de Dibujo Abierta (`Canvas`):** Para permitir extensiones gráficas sin acoplar dependencias pesadas, `Document` expone e implementa un contrato `Canvas` limpio y tipado para que utilidades externas puedan dibujar de forma controlada.
4. **Cero Residuos de Formatos Heredados:** Se han depurado por completo los cargadores de fuentes Type1/AFM obsoletos, maximizando la eficiencia de compilación WebAssembly (WASM).
