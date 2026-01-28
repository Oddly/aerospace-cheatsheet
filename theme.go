package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type overlayTheme struct{}

var _ fyne.Theme = (*overlayTheme)(nil)

func (t *overlayTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 24, G: 24, B: 30, A: 255}
	case theme.ColorNameForeground:
		return color.NRGBA{R: 212, G: 212, B: 212, A: 255}
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 60, G: 60, B: 65, A: 255}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 80, G: 80, B: 90, A: 200}
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 30, G: 30, B: 36, A: 255}
	default:
		return theme.DefaultTheme().Color(name, theme.VariantDark)
	}
}

func (t *overlayTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *overlayTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *overlayTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
