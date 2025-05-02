package views

import (
	"strings"

	"github.com/darkhz/bluetuith/ui/keybindings"
	"github.com/darkhz/bluetuith/ui/theme"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
)

// helpView holds the help view.
type helpView struct {
	popup modalView
	page  string
	area  *tview.Flex

	topics map[string][]HelpData

	*Views
}

// Initialize initializes the help view.
func (h *helpView) Initialize() error {
	if !h.cfg.Values.NoHelpDisplay {
		h.area = tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(horizontalLine(), 1, 0, false).
			AddItem(h.status.Help, 1, 0, false)
	}

	h.initHelpData()
	h.statusHelpArea(true)

	return nil
}

// SetRootView sets the root view for the help view.
func (h *helpView) SetRootView(v *Views) {
	h.Views = v
}

// statusHelpArea shows or hides the status help text.
func (h *helpView) statusHelpArea(add bool) {
	if h.cfg.Values.NoHelpDisplay {
		return
	}

	if !add && h.area != nil {
		h.layout.RemoveItem(h.area)
		return
	}

	h.layout.AddItem(h.area, 2, 0, false)
}

// swapStatusHelp adds or removes the provided primitive from the layout and displays
// the help text below the statusbar.
func (h *helpView) swapStatusHelp(primitive tview.Primitive, add bool) {
	h.statusHelpArea(false)
	defer h.statusHelpArea(true)

	if add {
		h.layout.AddItem(primitive, 8, 0, false)
	} else {
		h.layout.RemoveItem(primitive)
	}
}

// showStatusHelp shows a condensed help text for the currently focused screen below the statusbar.
func (h *helpView) showStatusHelp(page string) {
	if h.cfg.Values.NoHelpDisplay || h.page == page {
		return
	}

	h.page = page
	pages := map[string]string{
		devicePage.String():     "Device Screen",
		filePickerPage.String(): "File Picker",
		progressPage.String():   "Progress View",
	}

	items, ok := h.topics[pages[page]]
	if !ok {
		h.status.Help.Clear()
		return
	}

	groups := map[string][]HelpData{}

	for _, item := range items {
		if !item.ShowInStatus {
			continue
		}

		var group string

		for _, key := range item.Keys {
			switch key {
			case keybindings.KeyMenu, keybindings.KeySwitch:
				group = "Open"

			case keybindings.KeyFilebrowserSelect, keybindings.KeyFilebrowserInvertSelection, keybindings.KeyFilebrowserSelectAll:
				group = "Select"

			case keybindings.KeyProgressTransferSuspend, keybindings.KeyProgressTransferResume, keybindings.KeyProgressTransferCancel:
				group = "Transfer"

			case keybindings.KeyDeviceConnect, keybindings.KeyDevicePair, keybindings.KeyAdapterToggleScan, keybindings.KeyAdapterTogglePower:
				group = "Toggle"
			}
		}
		if group == "" {
			group = item.Title
		}

		helpItem := groups[group]
		if helpItem == nil {
			helpItem = []HelpData{}
		}

		helpItem = append(helpItem, item)
		groups[group] = helpItem
	}

	text := ""
	count := 0
	for group, items := range groups {
		var names, keys []string

		for _, item := range items {
			if item.Title != group {
				names = append(names, item.Title)
			}
			for _, k := range item.Keys {
				keys = append(keys, h.kb.Name(h.kb.Data(k).Kb))
			}
		}
		if names != nil {
			group += " " + strings.Join(names, "/")
		}

		helpKeys := strings.Join(keys, "/")
		if count < len(groups)-1 {
			helpKeys += ", "
		}

		title := theme.ColorWrap(theme.ThemeText, group, "::bu")
		helpKeys = theme.ColorWrap(theme.ThemeText, ": "+helpKeys)

		text += title + helpKeys
		count++
	}

	h.status.Help.SetText(text)
}

// showHelp displays a modal with the help items for all the screens.
func (h *helpView) showHelp() {
	var row int

	helpModal := h.modals.newModalWithTable("help", "Help", 40, 60)
	helpModal.table.SetSelectionChangedFunc(func(row, _ int) {
		if row == 1 {
			helpModal.table.ScrollToBeginning()
		}
	})
	helpModal.table.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		if action == tview.MouseScrollUp {
			helpModal.table.InputHandler()(tcell.NewEventKey(tcell.KeyUp, ' ', tcell.ModNone), nil)
		}

		return action, event
	})

	for title, helpItems := range h.topics {
		helpModal.table.SetCell(row, 0, tview.NewTableCell("[::bu]"+title).
			SetSelectable(false).
			SetAlign(tview.AlignCenter).
			SetTextColor(theme.GetColor(theme.ThemeText)),
		)

		row++

		for _, item := range helpItems {
			var names []string

			for _, k := range item.Keys {
				names = append(names, h.kb.Name(h.kb.Data(k).Kb))
			}

			keybinding := strings.Join(names, "/")

			helpModal.table.SetCell(row, 0, tview.NewTableCell(theme.ColorWrap(theme.ThemeText, item.Description)).
				SetExpansion(1).
				SetAlign(tview.AlignLeft).
				SetTextColor(theme.GetColor(theme.ThemeText)).
				SetSelectedStyle(tcell.Style{}.
					Foreground(theme.GetColor(theme.ThemeText)).
					Background(theme.BackgroundColor(theme.ThemeText)),
				),
			)

			helpModal.table.SetCell(row, 1, tview.NewTableCell(theme.ColorWrap(theme.ThemeText, keybinding)).
				SetExpansion(0).
				SetAlign(tview.AlignLeft).
				SetTextColor(theme.GetColor(theme.ThemeText)).
				SetSelectedStyle(tcell.Style{}.
					Foreground(theme.GetColor(theme.ThemeText)).
					Background(theme.BackgroundColor(theme.ThemeText)),
				),
			)

			row++
		}

		row++

	}

	helpModal.show()
}

// HelpData describes the help item.
type HelpData struct {
	Title, Description string
	Keys               []keybindings.Key
	ShowInStatus       bool
}

// initHelpData initializes the help data for all the specified screens.
func (h *helpView) initHelpData() {
	h.topics = map[string][]HelpData{
		"Device Screen": {
			{"Menu", "Open the menu", []keybindings.Key{keybindings.KeyMenu}, true},
			{"Switch", "Navigate between menus", []keybindings.Key{keybindings.KeySwitch}, true},
			{"Navigation", "Navigate between devices/options", []keybindings.Key{keybindings.KeyNavigateUp, keybindings.KeyNavigateDown}, true},
			{"Power", "Toggle adapter power state", []keybindings.Key{keybindings.KeyAdapterTogglePower}, true},
			{"Discoverable", "Toggle discoverable state", []keybindings.Key{keybindings.KeyAdapterToggleDiscoverable}, false},
			{"Pairable", "Toggle pairable state", []keybindings.Key{keybindings.KeyAdapterTogglePairable}, false},
			{"Scan", "Toggle scan (discovery state)", []keybindings.Key{keybindings.KeyAdapterToggleScan}, true},
			{"Adapter", "Change adapter", []keybindings.Key{keybindings.KeyAdapterChange}, true},
			{"Send", "Send files", []keybindings.Key{keybindings.KeyDeviceSendFiles}, true},
			{"Network", "Connect to network", []keybindings.Key{keybindings.KeyDeviceNetwork}, false},
			{"Progress", "Progress view", []keybindings.Key{keybindings.KeyProgressView}, false},
			{"Player", "Show/Hide player", []keybindings.Key{keybindings.KeyPlayerShow, keybindings.KeyPlayerHide}, false},
			{"Device Info", "Show device information", []keybindings.Key{keybindings.KeyDeviceInfo}, false},
			{"Connect", "Toggle connection with selected device", []keybindings.Key{keybindings.KeyDeviceConnect}, true},
			{"Pair", "Toggle pair with selected device", []keybindings.Key{keybindings.KeyDevicePair}, true},
			{"Trust", "Toggle trust with selected device", []keybindings.Key{keybindings.KeyDeviceTrust}, false},
			{"Remove", "Remove device from adapter", []keybindings.Key{keybindings.KeyDeviceRemove}, false},
			{"Cancel", "Cancel operation", []keybindings.Key{keybindings.KeyCancel}, false},
			{"Help", "Show help", []keybindings.Key{keybindings.KeyHelp}, true},
			{"Quit", "Quit", []keybindings.Key{keybindings.KeyQuit}, false},
		},
		"File Picker": {
			{"Navigation", "Navigate between directory entries", []keybindings.Key{keybindings.KeyNavigateUp, keybindings.KeyNavigateDown}, true},
			{"ChgDir Fwd/Back", "Enter/Go back a directory", []keybindings.Key{keybindings.KeyNavigateRight, keybindings.KeyNavigateLeft}, true},
			{"One", "Select one file", []keybindings.Key{keybindings.KeyFilebrowserSelect}, true},
			{"Invert", "Invert file selection", []keybindings.Key{keybindings.KeyFilebrowserInvertSelection}, true},
			{"All", "Select all files", []keybindings.Key{keybindings.KeyFilebrowserSelectAll}, true},
			{"Refresh", "Refresh current directory", []keybindings.Key{keybindings.KeyFilebrowserRefresh}, false},
			{"Hidden", "Toggle hidden files", []keybindings.Key{keybindings.KeyFilebrowserToggleHidden}, false},
			{"Confirm", "Confirm file(s) selection", []keybindings.Key{keybindings.KeyFilebrowserConfirmSelection}, true},
			{"Exit", "Exit", []keybindings.Key{keybindings.KeyClose}, false},
		},
		"Progress View": {
			{"Navigation", "Navigate between transfers", []keybindings.Key{keybindings.KeyNavigateUp, keybindings.KeyNavigateDown}, true},
			{"Suspend", "Suspend transfer", []keybindings.Key{keybindings.KeyProgressTransferSuspend}, true},
			{"Resume", "Resume transfer", []keybindings.Key{keybindings.KeyProgressTransferResume}, true},
			{"Cancel", "Cancel transfer", []keybindings.Key{keybindings.KeyProgressTransferCancel}, true},
			{"Exit", "Exit", []keybindings.Key{keybindings.KeyClose}, true},
		},
		"Media Player": {
			{"Play/Pause", "Toggle play/pause", []keybindings.Key{keybindings.KeyNavigateUp, keybindings.KeyNavigateDown}, false},
			{"Next", "Next", []keybindings.Key{keybindings.KeyPlayerNext}, false},
			{"Previous", "Previous", []keybindings.Key{keybindings.KeyPlayerPrevious}, false},
			{"Rewind", "Rewind", []keybindings.Key{keybindings.KeyPlayerSeekBackward}, false},
			{"Forward", "Fast forward", []keybindings.Key{keybindings.KeyPlayerSeekForward}, false},
			{"Stop", "Stop", []keybindings.Key{keybindings.KeyPlayerStop}, false},
		},
	}
}
