package main

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const (
	windowWidth  float32 = 550
	windowHeight float32 = 500
	appTitle             = "Aerospace Cheatsheet"
)

var (
	colorHeader   = color.NRGBA{R: 130, G: 170, B: 255, A: 255}
	colorText     = color.NRGBA{R: 212, G: 212, B: 212, A: 255}
	colorTitle    = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	colorSubtitle = color.NRGBA{R: 140, G: 140, B: 140, A: 255}
	colorError    = color.NRGBA{R: 255, G: 100, B: 100, A: 255}
)

func cheatsheetPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "aerospace", "cheatsheet.txt")
}

func loadCheatsheet() (string, error) {
	path := cheatsheetPath()
	if path == "" {
		return "", fmt.Errorf("could not determine home directory")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("could not read %s: %w", path, err)
	}
	return string(data), nil
}

func buildContent(text string) fyne.CanvasObject {
	text = strings.ReplaceAll(text, "\t", "    ")
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	objects := make([]fyne.CanvasObject, 0, len(lines))

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			spacer := canvas.NewRectangle(color.Transparent)
			spacer.SetMinSize(fyne.NewSize(0, 6))
			objects = append(objects, spacer)
			continue
		}

		if len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
			t := canvas.NewText(line, colorHeader)
			t.TextStyle.Bold = true
			t.TextSize = 14
			objects = append(objects, t)
		} else {
			t := canvas.NewText(line, colorText)
			t.TextStyle.Monospace = true
			t.TextSize = 13
			objects = append(objects, t)
		}
	}

	return container.NewVBox(objects...)
}

func buildErrorContent() fyne.CanvasObject {
	path := cheatsheetPath()
	if path == "" {
		path = "~/.config/aerospace/cheatsheet.txt"
	}

	errTitle := canvas.NewText("Cheatsheet file not found", colorError)
	errTitle.TextStyle.Bold = true
	errTitle.TextSize = 15

	msg := widget.NewLabel(fmt.Sprintf(
		"Create your cheatsheet file at:\n\n  %s\n\n"+
			"Lines without leading whitespace are treated as section headers.\n"+
			"Lines with leading whitespace are treated as shortcut entries.",
		path,
	))
	msg.Wrapping = fyne.TextWrapWord

	return container.NewVBox(errTitle, msg)
}

func main() {
	a := app.NewWithID("dev.chiark.aerospace-cheatsheet")
	a.Settings().SetTheme(&overlayTheme{})

	w := a.NewWindow(appTitle)
	w.Resize(fyne.NewSize(windowWidth, windowHeight))
	w.SetFixedSize(true)
	w.CenterOnScreen()

	// Title
	title := canvas.NewText(appTitle, colorTitle)
	title.TextStyle.Bold = true
	title.TextSize = 18
	title.Alignment = fyne.TextAlignCenter

	// Subtitle
	subtitle := canvas.NewText("Press ESC to close", colorSubtitle)
	subtitle.TextSize = 12
	subtitle.Alignment = fyne.TextAlignCenter

	// Header
	header := container.NewVBox(
		title,
		subtitle,
		widget.NewSeparator(),
	)

	// Content
	var body fyne.CanvasObject
	text, err := loadCheatsheet()
	if err != nil {
		body = buildErrorContent()
	} else {
		body = buildContent(text)
	}

	scroll := container.NewScroll(body)

	// Layout
	w.SetContent(container.NewPadded(
		container.NewBorder(header, nil, nil, nil, scroll),
	))

	// Escape to close
	w.Canvas().SetOnTypedKey(func(key *fyne.KeyEvent) {
		if key.Name == fyne.KeyEscape {
			a.Quit()
		}
	})

	// Platform-specific configuration
	configureWindow(w)

	w.ShowAndRun()
}
