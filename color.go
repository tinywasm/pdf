package pdf

import (
	. "github.com/tinywasm/fmt"
)

// Color is a CSS-style hex string, e.g. "#RRGGBB" or "#RGB".
type Color string

// parse converts the hex string to (r, g, b) integers.
func (c Color) parse() (r, g, b int, err error) {
	s := string(c)
	if s == "" {
		return 0, 0, 0, nil
	}
	if s[0] == '#' {
		s = s[1:]
	}

	if len(s) == 3 {
		r3, err := Convert(s[0:1] + s[0:1]).Int(16)
		if err != nil {
			return 0, 0, 0, Errf("color", string(c), "invalid")
		}
		g3, err := Convert(s[1:2] + s[1:2]).Int(16)
		if err != nil {
			return 0, 0, 0, Errf("color", string(c), "invalid")
		}
		b3, err := Convert(s[2:3] + s[2:3]).Int(16)
		if err != nil {
			return 0, 0, 0, Errf("color", string(c), "invalid")
		}
		return r3, g3, b3, nil
	}

	if len(s) == 6 {
		r6, err := Convert(s[0:2]).Int(16)
		if err != nil {
			return 0, 0, 0, Errf("color", string(c), "invalid")
		}
		g6, err := Convert(s[2:4]).Int(16)
		if err != nil {
			return 0, 0, 0, Errf("color", string(c), "invalid")
		}
		b6, err := Convert(s[4:6]).Int(16)
		if err != nil {
			return 0, 0, 0, Errf("color", string(c), "invalid")
		}
		return r6, g6, b6, nil
	}

	return 0, 0, 0, Errf("color", string(c), "invalid")
}

type Theme struct {
	Accent     Color // color para texto de headers (H1, H2, H3)
	Brand      Color // color de marca para elementos decorativos (líneas, bandas)
	Header     Color
	Gray       Color
	Body       Color
	FontFamily string
	Sizes      struct {
		H1, H2, H3, Body, Small float64
	}
	Spacing struct {
		Paragraph, Section, Page float64
	}
	// Page size in mm. Zero values keep the fpdf default (A4).
	Page ThemePage
	// Margin in mm applied to all four sides. Zero values keep the default (20mm).
	Margin ThemeMargin
}

// ThemePage defines the page dimensions in millimetres.
type ThemePage struct {
	Width, Height float64
}

// ThemeMargin defines the four page margins in millimetres.
type ThemeMargin struct {
	Top, Right, Bottom, Left float64
}

var DefaultTheme = Theme{
	Accent:     "#1E3C78",
	Header:     "#F0F4FA",
	Gray:       "#646464",
	Body:       "#000000",
	FontFamily: "Arial",
	Sizes: struct {
		H1, H2, H3, Body, Small float64
	}{
		H1:    16,
		H2:    12,
		H3:    10,
		Body:  10,
		Small: 8,
	},
	Spacing: struct {
		Paragraph, Section, Page float64
	}{
		Paragraph: 2,
		Section:   5,
		Page:      20,
	},
}
