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
	windowWidth  float32 = 900
	windowHeight float32 = 600
	columnCount          = 2
	appTitle             = "Aerospace Cheatsheet"
)

var (
	colorHeader   = color.NRGBA{R: 130, G: 170, B: 255, A: 255}
	colorText     = color.NRGBA{R: 212, G: 212, B: 212, A: 255}
	colorTitle    = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	colorSubtitle = color.NRGBA{R: 140, G: 140, B: 140, A: 255}
	colorError    = color.NRGBA{R: 255, G: 100, B: 100, A: 255}
)

// Section represents a group of shortcuts under a header
type Section struct {
	Header string
	Lines  []string
}

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

// parseSections splits the cheatsheet text into sections
func parseSections(text string) []Section {
	text = strings.ReplaceAll(text, "\t", "    ")
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")

	var sections []Section
	var current *Section

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Line without leading whitespace = section header
		if len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
			if current != nil {
				sections = append(sections, *current)
			}
			current = &Section{Header: line, Lines: []string{}}
		} else if current != nil {
			current.Lines = append(current.Lines, line)
		}
	}

	if current != nil {
		sections = append(sections, *current)
	}

	return sections
}

// buildSection creates a visual section with header and content
func buildSection(section Section) fyne.CanvasObject {
	objects := make([]fyne.CanvasObject, 0, len(section.Lines)+1)

	header := canvas.NewText(section.Header, colorHeader)
	header.TextStyle.Bold = true
	header.TextSize = 14
	objects = append(objects, header)

	for _, line := range section.Lines {
		t := canvas.NewText(line, colorText)
		t.TextStyle.Monospace = true
		t.TextSize = 12
		objects = append(objects, t)
	}

	return container.NewVBox(objects...)
}

// buildColumns distributes sections across multiple columns
func buildColumns(sections []Section, numCols int) fyne.CanvasObject {
	if len(sections) == 0 {
		return container.NewVBox()
	}

	// Calculate approximate lines per column for balanced distribution
	totalLines := 0
	for _, s := range sections {
		totalLines += len(s.Lines) + 2 // +2 for header and spacing
	}
	targetLinesPerCol := (totalLines + numCols - 1) / numCols

	columns := make([]fyne.CanvasObject, numCols)
	for i := range columns {
		columns[i] = container.NewVBox()
	}

	colIdx := 0
	colLines := 0

	for _, section := range sections {
		sectionLines := len(section.Lines) + 2

		// Move to next column if current is full (but not if we're on the last column)
		if colLines > 0 && colLines+sectionLines > targetLinesPerCol && colIdx < numCols-1 {
			colIdx++
			colLines = 0
		}

		sectionWidget := buildSection(section)
		spacer := canvas.NewRectangle(color.Transparent)
		spacer.SetMinSize(fyne.NewSize(0, 12))

		col := columns[colIdx].(*fyne.Container)
		col.Add(sectionWidget)
		col.Add(spacer)

		colLines += sectionLines
	}

	// Create horizontal grid of columns with padding between
	columnContainers := make([]fyne.CanvasObject, 0, numCols*2-1)
	for i, col := range columns {
		columnContainers = append(columnContainers, col)
		if i < numCols-1 {
			// Add separator between columns
			sep := canvas.NewRectangle(color.NRGBA{R: 50, G: 50, B: 55, A: 255})
			sep.SetMinSize(fyne.NewSize(1, 0))
			columnContainers = append(columnContainers, container.NewPadded(sep))
		}
	}

	return container.NewHBox(columnContainers...)
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
	w.CenterOnScreen()

	// Title
	title := canvas.NewText(appTitle, colorTitle)
	title.TextStyle.Bold = true
	title.TextSize = 18
	title.Alignment = fyne.TextAlignCenter

	// Subtitle with VIM keys hint
	subtitle := canvas.NewText("Press ESC or Q to close  •  J/K to scroll", colorSubtitle)
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
		sections := parseSections(text)
		body = buildColumns(sections, columnCount)
	}

	scroll := container.NewScroll(body)

	// Layout
	w.SetContent(container.NewPadded(
		container.NewBorder(header, nil, nil, nil, scroll),
	))

	// Keyboard handling with VIM keys
	w.Canvas().SetOnTypedKey(func(key *fyne.KeyEvent) {
		switch key.Name {
		case fyne.KeyEscape:
			a.Quit()
		case fyne.KeyJ:
			// Scroll down
			scroll.ScrollToBottom()
			pos := scroll.Offset
			pos.Y += 50
			scroll.Offset = pos
			scroll.Refresh()
		case fyne.KeyK:
			// Scroll up
			pos := scroll.Offset
			pos.Y -= 50
			if pos.Y < 0 {
				pos.Y = 0
			}
			scroll.Offset = pos
			scroll.Refresh()
		case fyne.KeyG:
			// Scroll to top (gg in vim, but single g here for simplicity)
			scroll.ScrollToTop()
		}
	})

	w.Canvas().SetOnTypedRune(func(r rune) {
		switch r {
		case 'q', 'Q':
			a.Quit()
		case 'j':
			pos := scroll.Offset
			pos.Y += 50
			scroll.Offset = pos
			scroll.Refresh()
		case 'k':
			pos := scroll.Offset
			pos.Y -= 50
			if pos.Y < 0 {
				pos.Y = 0
			}
			scroll.Offset = pos
			scroll.Refresh()
		case 'g', 'G':
			if r == 'G' {
				scroll.ScrollToBottom()
			} else {
				scroll.ScrollToTop()
			}
		}
	})

	// Platform-specific configuration
	configureWindow(w)

	w.ShowAndRun()
}
