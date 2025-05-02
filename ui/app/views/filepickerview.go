package views

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"go.uber.org/atomic"

	"slices"

	"github.com/darkhz/bluetuith/ui/keybindings"
	"github.com/darkhz/bluetuith/ui/theme"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
)

const filePickerPage viewName = "filepicker"

// filePickerView holds the file picker view.
type filePickerView struct {
	isSupported bool

	table          *tview.Table
	title, buttons *tview.TextView
	pickerFlex     *tview.Flex

	prevDir, currentPath string
	isHidden             atomic.Bool

	listChan     chan []string
	prevFileInfo fs.DirEntry

	selectedFiles map[string]fs.DirEntry
	selectMu      sync.Mutex

	mu sync.Mutex

	*Views
}

const filePickButtonRegion = `["ok"][::b][OK[][""] ["cancel"][::b][Cancel[][""] ["hidden"][::b][Toggle hidden[][""] ["invert"][Invert selection[][""] ["all"][Select All[][""]`

// Initialize initializes the file picker.
func (f *filePickerView) Initialize() error {
	f.reset()

	infoTitle := tview.NewTextView()
	infoTitle.SetDynamicColors(true)
	infoTitle.SetTextAlign(tview.AlignCenter)
	infoTitle.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))
	infoTitle.SetText(theme.ColorWrap(theme.ThemeText, "Select files to send", "::bu"))

	f.pickerFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(infoTitle, 1, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(f.filePickerTitle(), 1, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(f.filePickerTable(), 0, 10, true).
		AddItem(nil, 1, 0, false).
		AddItem(f.filePickerButtons(), 2, 0, false)
	f.pickerFlex.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	if !f.getHidden() {
		f.toggleHidden()
	}

	f.isSupported = true

	return nil
}

// SetRootView sets the root view for the file picker.
func (f *filePickerView) SetRootView(v *Views) {
	f.Views = v
}

// Show shows a file picker, and returns
// a list of all the selected files.
func (f *filePickerView) Show() ([]string, error) {
	if !f.isSupported {
		return nil, errors.New("the filepicker cannot be opened since sending files is not supported")
	}

	f.reset()
	f.app.InstantDraw(func() {
		f.pages.AddAndSwitchToPage(filePickerPage.String(), f.pickerFlex, true)
		go f.changeDir(false, false)
	})

	return <-f.listChan, nil
}

// reset resets the list of selected files.
func (f *filePickerView) reset() {
	f.listChan = make(chan []string)
	f.selectedFiles = make(map[string]fs.DirEntry)
}

// filePickerTable sets up and returns the filepicker table.
func (f *filePickerView) filePickerTable() *tview.Table {
	f.table = tview.NewTable()
	f.table.SetSelectorWrap(true)
	f.table.SetSelectable(true, false)
	f.table.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))
	f.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch f.kb.Key(event, keybindings.ContextFiles) {
		case keybindings.KeyFilebrowserDirForward:
			go f.changeDir(true, false)

		case keybindings.KeyFilebrowserDirBack:
			go f.changeDir(false, true)

		case keybindings.KeyFilebrowserToggleHidden:
			f.toggleHidden()
			fallthrough

		case keybindings.KeyFilebrowserRefresh:
			go f.changeDir(false, false)

		case keybindings.KeyFilebrowserConfirmSelection:
			f.sendFileList()
			fallthrough

		case keybindings.KeyClose:
			f.buttonHandler("cancel")

		case keybindings.KeyQuit:
			go f.actions.quit()

		case keybindings.KeyHelp:
			f.help.showHelp()

		case keybindings.KeyFilebrowserSelectAll, keybindings.KeyFilebrowserInvertSelection, keybindings.KeyFilebrowserSelect:
			f.selectFile(event.Rune())
		}

		return ignoreDefaultEvent(event)
	})

	return f.table
}

// filePickerTitle sets up and returns the filepicker title area.
// This will be used to show the current directory path.
func (f *filePickerView) filePickerTitle() *tview.TextView {
	f.title = tview.NewTextView()
	f.title.SetDynamicColors(true)
	f.title.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))
	f.title.SetTextAlign(tview.AlignLeft)

	return f.title
}

// filePickerButtons sets up and returns the filepicker buttons.
func (f *filePickerView) filePickerButtons() *tview.TextView {
	f.buttons = tview.NewTextView()
	f.buttons.SetRegions(true)
	f.buttons.SetDynamicColors(true)
	f.buttons.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))
	f.buttons.SetHighlightedFunc(func(added, _, _ []string) {
		if added == nil {
			return
		}

		if slices.Contains(f.buttons.GetRegionIDs(), added[0]) {
			f.buttonHandler(added[0])
		}
	})

	f.buttons.SetTextAlign(tview.AlignLeft)
	f.buttons.SetText(theme.ColorWrap(theme.ThemeText, filePickButtonRegion))

	return f.buttons
}

// sendFileList sends a slice of all the selected files
// to the file list channel, which is received by filePicker().
func (f *filePickerView) sendFileList() {
	f.selectMu.Lock()
	defer f.selectMu.Unlock()

	var fileList []string

	for path := range f.selectedFiles {
		fileList = append(fileList, path)
	}

	f.listChan <- fileList
}

// selectFile sets the parameters for the file selection handler.
func (f *filePickerView) selectFile(key rune) {
	var all, inverse bool

	switch key {
	case 'A':
		all = true
		inverse = false

	case 'a':
		all = false
		inverse = true

	case ' ':
		all = false
		inverse = false
	}

	f.selectFileHandler(all, inverse)
}

// changeDir changes to a directory and lists its contents.
func (f *filePickerView) changeDir(cdFwd bool, cdBack bool) {
	var testPath string

	f.mu.Lock()
	defer f.mu.Unlock()

	if f.currentPath == "" {
		var err error

		f.currentPath, err = os.UserHomeDir()
		if err != nil {
			f.status.ErrorMessage(err)
			return
		}
	}

	testPath = f.currentPath

	row, _ := f.table.GetSelection()
	cell := f.table.GetCell(row, 1)
	if cell == nil {
		return
	}
	if cdFwd && tview.Escape(cell.Text) == "../" {
		cdFwd = false
		cdBack = true
	}

	switch {
	case cdFwd:
		entry, ok := cell.GetReference().(fs.DirEntry)
		if !ok {
			return
		}

		testPath = f.trimPath(testPath, false)
		testPath = filepath.Join(testPath, entry.Name())

	case cdBack:
		f.prevDir = filepath.Base(testPath)
		testPath = f.trimPath(testPath, cdBack)
	}

	dlist, listed := f.dirList(filepath.FromSlash(testPath))
	if !listed {
		return
	}

	sort.Slice(dlist, func(i, j int) bool {
		if dlist[i].IsDir() != dlist[j].IsDir() {
			return dlist[i].IsDir()
		}

		return dlist[i].Name() < dlist[j].Name()
	})

	f.currentPath = testPath

	f.createDirList(dlist, cdBack)
}

// createDirList displays the contents of the directory in the f.
func (f *filePickerView) createDirList(dlist []fs.DirEntry, cdBack bool) {
	f.app.InstantDraw(func() {
		var pos int

		prevrow := -1
		rowpadding := -1

		f.table.SetSelectable(false, false)
		f.table.Clear()

		for row, entry := range dlist {
			var attr tcell.AttrMask
			var entryColor tcell.Color

			info, err := entry.Info()
			if err != nil {
				continue
			}

			name := info.Name()
			fileTotalSize := formatSize(info.Size())
			fileModifiedTime := info.ModTime().Format("02 Jan 2006 03:04 PM")
			permissions := strings.ToLower(entry.Type().String())
			if len(permissions) > 10 {
				permissions = permissions[1:]
			}

			if entry.IsDir() {
				if f.currentPath != "/" {
					rowpadding = 0

					if cdBack && name == f.prevDir {
						pos = row
					}

					if entry == f.prevFileInfo {
						name = ".."
						prevrow = row

						row = 0
						f.table.InsertRow(0)

						if f.table.GetRowCount() > 0 {
							pos++
						}
					}
				} else if f.currentPath == "/" && name == "/" {
					rowpadding = -1
					continue
				}

				attr = tcell.AttrBold
				entryColor = tcell.ColorBlue
				name += string(os.PathSeparator)
			} else {
				entryColor = theme.GetColor(theme.ThemeText)
			}

			f.table.SetCell(row+rowpadding, 0, tview.NewTableCell(" ").
				SetSelectable(false),
			)

			f.table.SetCell(row+rowpadding, 1, tview.NewTableCell(tview.Escape(name)).
				SetExpansion(1).
				SetReference(entry).
				SetAttributes(attr).
				SetTextColor(entryColor).
				SetAlign(tview.AlignLeft).
				SetOnClickedFunc(f.cellHandler).
				SetSelectedStyle(tcell.Style{}.
					Bold(true).
					Foreground(entryColor).
					Background(theme.BackgroundColor(theme.ThemeText)),
				),
			)

			for col, text := range []string{
				permissions,
				fileTotalSize,
				fileModifiedTime,
			} {
				f.table.SetCell(row+rowpadding, col+2, tview.NewTableCell(text).
					SetAlign(tview.AlignRight).
					SetTextColor(tcell.ColorGrey).
					SetSelectedStyle(tcell.Style{}.
						Bold(true),
					),
				)
			}

			if prevrow > -1 {
				row = prevrow
				prevrow = -1
			}

			f.markFileSelection(row, entry, f.checkFileSelected(filepath.Join(f.currentPath, name)))
		}

		f.title.SetText(theme.ColorWrap(theme.ThemeText, "Directory: "+f.currentPath))

		f.table.ScrollToBeginning()
		f.table.SetSelectable(true, false)
		f.table.Select(pos, 0)
	})
}

// dirList lists a directory's contents.
func (f *filePickerView) dirList(testPath string) ([]fs.DirEntry, bool) {
	var dlist []fs.DirEntry

	_, err := os.Lstat(testPath)
	if err != nil {
		return nil, false
	}

	dir, err := os.Lstat(f.trimPath(testPath, true))
	if err != nil {
		return nil, false
	}

	list, err := os.ReadDir(testPath)
	if err != nil {
		return nil, false
	}

	dirEntry := fs.FileInfoToDirEntry(dir)

	f.prevFileInfo = dirEntry
	dlist = append(dlist, dirEntry)

	for _, entry := range list {
		if f.getHidden() && strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		dlist = append(dlist, entry)
	}

	return dlist, true
}

// cellHandler handles on-click events for a table cell.
func (f *filePickerView) cellHandler(_ *tview.Table, row, _ int) {
	f.selectFileHandler(false, false, row)
}

// buttonHandler handles button on-click events for the filepicker buttons.
func (f *filePickerView) buttonHandler(button string) {
	switch button {
	case "ok":
		f.sendFileList()
		fallthrough

	case "cancel":
		close(f.listChan)

		f.pages.RemovePage(filePickerPage.String())
		f.pages.SwitchToPage(devicePage.String())

	case "hidden":
		f.toggleHidden()
		go f.changeDir(false, false)

	case "invert":
		f.selectFileHandler(false, true)

	case "all":
		f.selectFileHandler(true, false)
	}

	f.buttons.Highlight("")
}

// selectFileHandler iterates over the f.table's rows,
// determines the type of selection to be made (single, inverse or all),
// and marks the selections.
func (f *filePickerView) selectFileHandler(all, inverse bool, row ...int) {
	var pos int

	singleSelection := !all && !inverse
	inverseSelection := !all && inverse

	if row != nil {
		pos = row[0]
	} else {
		pos, _ = f.table.GetSelection()
	}
	totalrows := f.table.GetRowCount()

	for i := 0; i < totalrows; i++ {
		var checkSelected bool
		var userSelected []struct{}

		if singleSelection {
			i = pos
			userSelected = append(userSelected, struct{}{})
		}

		cell := f.table.GetCell(i, 1)
		if cell == nil {
			return
		}

		entry, ok := cell.GetReference().(fs.DirEntry)
		if !ok {
			return
		}

		fullpath := filepath.Join(f.currentPath, entry.Name())
		if singleSelection || inverseSelection {
			checkSelected = f.checkFileSelected(fullpath)
		}
		if !checkSelected {
			f.addFileSelection(fullpath, entry)
		} else {
			f.removeFileSelection(fullpath)
		}

		f.markFileSelection(i, entry, !checkSelected, userSelected...)

		if singleSelection {
			if i+1 < totalrows {
				f.table.Select(i+1, 0)
				return
			}

			break
		}
	}

	f.table.Select(pos, 0)
}

// addFileSelection adds a file to the f.selectedFiles list.
func (f *filePickerView) addFileSelection(path string, info fs.DirEntry) {
	f.selectMu.Lock()
	defer f.selectMu.Unlock()

	if !info.Type().IsRegular() {
		return
	}

	f.selectedFiles[path] = info
}

// removeFileSelection removes a file from the f.selectedFiles list.
func (f *filePickerView) removeFileSelection(path string) {
	f.selectMu.Lock()
	defer f.selectMu.Unlock()

	delete(f.selectedFiles, path)
}

// checkFileSelected checks if a file is selected.
func (f *filePickerView) checkFileSelected(path string) bool {
	f.selectMu.Lock()
	defer f.selectMu.Unlock()

	_, selected := f.selectedFiles[path]

	return selected
}

// markFileSelection marks the selection for files only, directories are skipped.
func (f *filePickerView) markFileSelection(row int, info fs.DirEntry, selected bool, userSelected ...struct{}) {
	if !info.Type().IsRegular() {
		if info.IsDir() && userSelected != nil {
			go f.changeDir(true, false)
		}

		return
	}

	cell := f.table.GetCell(row, 0)

	if selected {
		cell.Text = "+"
	} else {
		cell.Text = " "
	}

	cell.Text = theme.ColorWrap(theme.ThemeText, cell.Text)
}

// trimPath trims a given path and appends a path separator where appropriate.
func (f *filePickerView) trimPath(testPath string, cdBack bool) string {
	testPath = filepath.Clean(testPath)

	if cdBack {
		testPath = filepath.Dir(testPath)
	}

	return filepath.FromSlash(testPath)
}

// getHidden checks if hidden files can be shown or not.
func (f *filePickerView) getHidden() bool {
	return f.isHidden.Load()
}

// toggleHidden toggles the hidden files mode.
func (f *filePickerView) toggleHidden() {
	f.isHidden.Store(!f.isHidden.Load())
}

// formatSize returns the human readable form of a size value in bytes.
// Adapted from: https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/
func formatSize(size int64) string {
	const unit = 1000
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "kMGTPE"[exp])
}
