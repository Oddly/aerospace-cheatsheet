package main

import (
	"flag"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/fsnotify/fsnotify"
)

const (
	defaultWidth  float32 = 900
	defaultHeight float32 = 600
	scrollStep    float32 = 50
	appTitle              = "Aerospace Cheatsheet"
)

var (
	// Colors
	colorHeader      = color.NRGBA{R: 130, G: 170, B: 255, A: 255}
	colorKey         = color.NRGBA{R: 255, G: 203, B: 107, A: 255} // Gold for key combos
	colorDescription = color.NRGBA{R: 180, G: 180, B: 180, A: 255}
	colorTitle       = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	colorSubtitle    = color.NRGBA{R: 140, G: 140, B: 140, A: 255}
	colorError       = color.NRGBA{R: 255, G: 100, B: 100, A: 255}
	colorSearchBg    = color.NRGBA{R: 40, G: 40, B: 45, A: 255}

	// CLI flags
	configPath string
	numColumns int

	// Regex to split "key combo" from "description" (separated by 2+ spaces)
	shortcutRegex = regexp.MustCompile(`^(\s*)(\S+(?:\s+\S+)*?)\s{2,}(.+)$`)
)

func init() {
	flag.StringVar(&configPath, "config", "", "Path to cheatsheet file")
	flag.IntVar(&numColumns, "columns", 2, "Number of columns (1-3)")
}

// Section represents a group of shortcuts under a header
type Section struct {
	Header string
	Lines  []string
}

// CheatsheetApp holds the application state
type CheatsheetApp struct {
	fyneApp    fyne.App
	window     fyne.Window
	scroll     *container.Scroll
	searchBox  *widget.Entry
	searchRow  *fyne.Container
	sections   []Section
	configPath string
	numColumns int
	searchMode bool
}

func defaultCheatsheetPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "aerospace", "cheatsheet.txt")
}

func loadCheatsheet(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("no config path specified")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("could not read %s: %w", path, err)
	}
	return string(data), nil
}

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

// buildShortcutLine creates a line with syntax highlighting for key combos
func buildShortcutLine(line string) fyne.CanvasObject {
	matches := shortcutRegex.FindStringSubmatch(line)
	if matches == nil {
		// No match - render as plain text
		t := canvas.NewText(line, colorDescription)
		t.TextStyle.Monospace = true
		t.TextSize = 12
		return t
	}

	indent := matches[1]
	keyCombo := matches[2]
	description := matches[3]

	// Create colored segments
	indentText := canvas.NewText(indent, colorDescription)
	indentText.TextStyle.Monospace = true
	indentText.TextSize = 12

	keyText := canvas.NewText(keyCombo, colorKey)
	keyText.TextStyle.Monospace = true
	keyText.TextStyle.Bold = true
	keyText.TextSize = 12

	// Calculate spacing to maintain alignment
	spacing := strings.Repeat(" ", 2)
	spacerText := canvas.NewText(spacing, colorDescription)
	spacerText.TextStyle.Monospace = true
	spacerText.TextSize = 12

	descText := canvas.NewText(description, colorDescription)
	descText.TextStyle.Monospace = true
	descText.TextSize = 12

	return container.NewHBox(indentText, keyText, spacerText, descText)
}

func buildSection(section Section) fyne.CanvasObject {
	objects := make([]fyne.CanvasObject, 0, len(section.Lines)+1)

	header := canvas.NewText(section.Header, colorHeader)
	header.TextStyle.Bold = true
	header.TextSize = 14
	objects = append(objects, header)

	for _, line := range section.Lines {
		objects = append(objects, buildShortcutLine(line))
	}

	return container.NewVBox(objects...)
}

func buildColumns(sections []Section, numCols int) fyne.CanvasObject {
	if len(sections) == 0 {
		return container.NewVBox()
	}

	if numCols < 1 {
		numCols = 1
	}
	if numCols > 3 {
		numCols = 3
	}

	totalLines := 0
	for _, s := range sections {
		totalLines += len(s.Lines) + 2
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

	columnContainers := make([]fyne.CanvasObject, 0, numCols*2-1)
	for i, col := range columns {
		columnContainers = append(columnContainers, col)
		if i < numCols-1 {
			sep := canvas.NewRectangle(color.NRGBA{R: 50, G: 50, B: 55, A: 255})
			sep.SetMinSize(fyne.NewSize(1, 0))
			columnContainers = append(columnContainers, container.NewPadded(sep))
		}
	}

	return container.NewHBox(columnContainers...)
}

func (a *CheatsheetApp) filterSections(query string) []Section {
	if query == "" {
		return a.sections
	}

	query = strings.ToLower(query)
	var result []Section

	for _, s := range a.sections {
		if strings.Contains(strings.ToLower(s.Header), query) {
			result = append(result, s)
			continue
		}

		var matchedLines []string
		for _, line := range s.Lines {
			if strings.Contains(strings.ToLower(line), query) {
				matchedLines = append(matchedLines, line)
			}
		}
		if len(matchedLines) > 0 {
			result = append(result, Section{Header: s.Header, Lines: matchedLines})
		}
	}

	return result
}

func (a *CheatsheetApp) rebuildContent(query string) {
	filtered := a.filterSections(query)
	if len(filtered) == 0 {
		noResults := canvas.NewText("No matches found", colorSubtitle)
		noResults.TextSize = 14
		a.scroll.Content = container.NewCenter(noResults)
	} else {
		a.scroll.Content = buildColumns(filtered, a.numColumns)
	}
	a.scroll.Refresh()
}

func (a *CheatsheetApp) loadAndDisplay() {
	text, err := loadCheatsheet(a.configPath)
	if err != nil {
		a.scroll.Content = buildErrorContent(a.configPath)
	} else {
		a.sections = parseSections(text)
		query := ""
		if a.searchMode {
			query = a.searchBox.Text
		}
		a.rebuildContent(query)
	}
	a.scroll.Refresh()
}

func (a *CheatsheetApp) toggleSearch() {
	a.searchMode = !a.searchMode
	if a.searchMode {
		a.searchRow.Show()
		a.window.Canvas().Focus(a.searchBox)
	} else {
		a.searchRow.Hide()
		a.searchBox.SetText("")
		a.rebuildContent("")
	}
}

func (a *CheatsheetApp) watchConfigFile() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}
	defer watcher.Close()

	// Watch the directory to handle atomic saves
	dir := filepath.Dir(a.configPath)
	if err := watcher.Add(dir); err != nil {
		return
	}

	filename := filepath.Base(a.configPath)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if filepath.Base(event.Name) == filename {
				if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
					a.loadAndDisplay()
				}
			}
		case _, ok := <-watcher.Errors:
			if !ok {
				return
			}
		}
	}
}

func buildErrorContent(path string) fyne.CanvasObject {
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
	flag.Parse()

	// Resolve config path: CLI flag > env var > default
	path := configPath
	if path == "" {
		path = os.Getenv("AEROSPACE_CHEATSHEET")
	}
	if path == "" {
		path = defaultCheatsheetPath()
	}

	cols := numColumns
	if cols < 1 {
		cols = 1
	}
	if cols > 3 {
		cols = 3
	}

	myApp := &CheatsheetApp{
		configPath: path,
		numColumns: cols,
	}

	myApp.fyneApp = app.NewWithID("io.github.oddly.aerospace-cheatsheet")
	myApp.fyneApp.Settings().SetTheme(&overlayTheme{})

	myApp.window = myApp.fyneApp.NewWindow(appTitle)
	myApp.window.Resize(fyne.NewSize(defaultWidth, defaultHeight))
	myApp.window.CenterOnScreen()

	// Title
	title := canvas.NewText(appTitle, colorTitle)
	title.TextStyle.Bold = true
	title.TextSize = 18
	title.Alignment = fyne.TextAlignCenter

	// Subtitle
	subtitle := canvas.NewText("ESC/Q close • H/J/K/L scroll • G top/bottom • / search", colorSubtitle)
	subtitle.TextSize = 11
	subtitle.Alignment = fyne.TextAlignCenter

	// Search box (hidden by default)
	myApp.searchBox = widget.NewEntry()
	myApp.searchBox.SetPlaceHolder("Search shortcuts...")
	myApp.searchBox.OnChanged = func(text string) {
		myApp.rebuildContent(text)
	}

	searchLabel := canvas.NewText("/", colorKey)
	searchLabel.TextStyle.Bold = true
	searchLabel.TextSize = 14

	myApp.searchRow = container.NewBorder(nil, nil, searchLabel, nil, myApp.searchBox)
	myApp.searchRow.Hide()

	// Header
	header := container.NewVBox(
		title,
		subtitle,
		myApp.searchRow,
		widget.NewSeparator(),
	)

	// Load initial content
	text, err := loadCheatsheet(path)
	var body fyne.CanvasObject
	if err != nil {
		body = buildErrorContent(path)
	} else {
		myApp.sections = parseSections(text)
		body = buildColumns(myApp.sections, cols)
	}

	myApp.scroll = container.NewScroll(body)

	// Layout
	myApp.window.SetContent(container.NewPadded(
		container.NewBorder(header, nil, nil, nil, myApp.scroll),
	))

	// Close on focus loss (click outside behavior)
	myApp.window.Canvas().SetOnTypedKey(func(key *fyne.KeyEvent) {
		// If in search mode, let the search box handle most keys
		if myApp.searchMode {
			if key.Name == fyne.KeyEscape {
				myApp.toggleSearch()
			}
			return
		}

		switch key.Name {
		case fyne.KeyEscape:
			myApp.fyneApp.Quit()
		case fyne.KeyDown, fyne.KeyJ:
			pos := myApp.scroll.Offset
			pos.Y += scrollStep
			myApp.scroll.Offset = pos
			myApp.scroll.Refresh()
		case fyne.KeyUp, fyne.KeyK:
			pos := myApp.scroll.Offset
			pos.Y -= scrollStep
			if pos.Y < 0 {
				pos.Y = 0
			}
			myApp.scroll.Offset = pos
			myApp.scroll.Refresh()
		case fyne.KeyLeft, fyne.KeyH:
			pos := myApp.scroll.Offset
			pos.X -= scrollStep
			if pos.X < 0 {
				pos.X = 0
			}
			myApp.scroll.Offset = pos
			myApp.scroll.Refresh()
		case fyne.KeyRight, fyne.KeyL:
			pos := myApp.scroll.Offset
			pos.X += scrollStep
			myApp.scroll.Offset = pos
			myApp.scroll.Refresh()
		case fyne.KeyG:
			myApp.scroll.ScrollToTop()
		}
	})

	myApp.window.Canvas().SetOnTypedRune(func(r rune) {
		if myApp.searchMode {
			return // Let Entry handle input
		}

		switch r {
		case 'q', 'Q':
			myApp.fyneApp.Quit()
		case 'j':
			pos := myApp.scroll.Offset
			pos.Y += scrollStep
			myApp.scroll.Offset = pos
			myApp.scroll.Refresh()
		case 'k':
			pos := myApp.scroll.Offset
			pos.Y -= scrollStep
			if pos.Y < 0 {
				pos.Y = 0
			}
			myApp.scroll.Offset = pos
			myApp.scroll.Refresh()
		case 'h':
			pos := myApp.scroll.Offset
			pos.X -= scrollStep
			if pos.X < 0 {
				pos.X = 0
			}
			myApp.scroll.Offset = pos
			myApp.scroll.Refresh()
		case 'l':
			pos := myApp.scroll.Offset
			pos.X += scrollStep
			myApp.scroll.Offset = pos
			myApp.scroll.Refresh()
		case 'g':
			myApp.scroll.ScrollToTop()
		case 'G':
			myApp.scroll.ScrollToBottom()
		case '/':
			myApp.toggleSearch()
		}
	})

	// Platform-specific configuration
	configureWindow(myApp.window)

	// Start file watcher for hot reload
	go myApp.watchConfigFile()

	myApp.window.ShowAndRun()
}
