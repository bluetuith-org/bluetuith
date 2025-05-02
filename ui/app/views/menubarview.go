package views

import (
	"sync"

	"github.com/darkhz/bluetuith/ui/keybindings"
	"github.com/darkhz/bluetuith/ui/theme"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
)

const (
	MenuBarName           viewName = "menu"
	menuAdapterChangeName viewName = "adapterchange"
	menuAdapterName       viewName = "adapter"
	menuDeviceName        viewName = "device"
)

const menuBarRegions = `["adapter"][::b][Adapter[][""] ["device"][::b][Device[][""]`

// menuBarView describes a region to display menu items.
type menuBarView struct {
	bar   *tview.TextView
	modal *modalView

	menuOptions map[string][]menuOption
	optionByKey map[keybindings.Key]*menuOptionState

	sync.RWMutex

	*Views
}

// menuOption describes an option layout for a submenu.
type menuOption struct {
	key keybindings.Key

	enabledText, disabledText         string
	initBeforeInvoke, checkVisibility bool
}

type menuOptionState struct {
	index    int
	menuName string

	displayText  string
	toggledState bool

	kdata *keybindings.KeyData

	*menuOption
}

func (m *menuBarView) Initialize() error {
	m.initOrderedOptions()
	m.setOptions()

	m.bar = tview.NewTextView()
	m.bar.SetRegions(true)
	m.bar.SetDynamicColors(true)
	m.bar.SetBackgroundColor(theme.GetColor(theme.ThemeMenuBar))
	m.bar.SetHighlightedFunc(func(added, removed, remaining []string) {
		if added == nil {
			return
		}

		for _, region := range m.bar.GetRegionIDs() {
			if region == added[0] {
				if region == menuAdapterChangeName.String() {
					m.adapter.change()
					m.bar.Highlight("")
				} else {
					m.setupSubMenu(m.bar.GetRegionStart(region), 1, viewName(added[0]))
				}

				break
			}
		}
	})

	m.modal = m.modals.NewMenuModal(MenuBarName.String(), 0, 0)

	return nil
}

func (m *menuBarView) SetRootView(v *Views) {
	m.Views = v
}

func (m *menuBarView) highlight(name viewName) {
	m.bar.Highlight(name.String())
}

// setOptions sets up the menu options with its attributes.
func (m *menuBarView) setOptions() {
	m.optionByKey = make(map[keybindings.Key]*menuOptionState)

	for menuName, keybindings := range m.menuOptions {
		for index, option := range keybindings {
			m.optionByKey[option.key] = &menuOptionState{
				index:        index,
				menuName:     menuName,
				displayText:  "",
				toggledState: false,
				kdata:        m.kb.Data(option.key),
				menuOption:   &option,
			}
		}
	}
}

// setItemToggle sets the toggled state of the specified menu item using
// the menu's name and the submenuenu's ID.
func (m *menuBarView) setItemToggle(menuKey keybindings.Key, toggle bool) {
	var name, displayText string
	var index int

	m.Lock()
	optstate, ok := m.optionByKey[menuKey]
	if ok {
		item := m.toggleMenuItem(optstate, toggle)
		name, displayText, index = item.menuName, item.displayText, item.index
	}
	m.Unlock()
	if !ok {
		return
	}

	m.app.InstantDraw(func() {
		highlighted := m.bar.GetHighlights()

		if m.modal.Open && highlighted != nil && highlighted[0] == name {
			cell := m.modal.Table.GetCell(index, 0)
			if cell == nil {
				return
			}

			cell.Text = displayText
		}
	})
}

func (m *menuBarView) toggleMenuItem(menuItem *menuOptionState, toggle bool) *menuOptionState {
	title := menuItem.kdata.Title
	switch {
	case menuItem.disabledText == "":
		menuItem.displayText = title
		return menuItem

	case menuItem.enabledText == "" && menuItem.disabledText != "":
		if toggle {
			menuItem.displayText = menuItem.disabledText
		} else {
			menuItem.displayText = title
		}

		menuItem.toggledState = toggle

		return menuItem
	}

	if toggle {
		menuItem.displayText = title + " " + menuItem.disabledText
	} else {
		menuItem.displayText = title + " " + menuItem.enabledText
	}

	menuItem.toggledState = toggle

	return menuItem
}

// setupSubMenu sets up a submenu for the specified menu.
func (m *menuBarView) setupSubMenu(x, y int, menuID viewName, device ...struct{}) {
	var width, skipped int

	modal := m.modal
	modal.Table.SetSelectorWrap(false)
	modal.Table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch m.kb.Key(event) {
		case keybindings.KeyClose:
			m.exit()
			return event

		case keybindings.KeySwitch:
			m.next()
			return event

		case keybindings.KeySelect:
			m.exit()
		}

		m.inputHandler(event)

		return ignoreDefaultEvent(event)
	})
	modal.Table.SetSelectedFunc(func(row, col int) {
		cell := m.modal.Table.GetCell(row, 0)
		if cell == nil {
			return
		}

		ref, ok := cell.GetReference().(*menuOption)
		if !ok || ref == nil {
			return
		}

		m.actions.handler(ref.key, actionInvoke)()
	})

	m.Lock()
	defer m.Unlock()

	modal.Table.Clear()
	for index, menuopt := range m.menuOptions[menuID.String()] {
		if menuopt.checkVisibility && !m.actions.handler(menuopt.key, actionVisibility)() {
			skipped++
			continue
		}

		optstate, ok := m.optionByKey[menuopt.key]
		if !ok {
			continue
		}

		toggle := optstate.toggledState
		if menuopt.initBeforeInvoke {
			if initializer := m.actions.handler(menuopt.key, actionInitializer); initializer != nil {
				toggle = initializer()
			}
		}

		newstate := m.toggleMenuItem(optstate, toggle)

		display := newstate.displayText
		keybinding := m.kb.Name(newstate.kdata.Kb)
		clickedFunc := func() bool {
			m.exit()
			return m.actions.handler(menuopt.key, actionInvoke)()
		}

		displayWidth := len(display) + len(keybinding) + 6
		if displayWidth > width {
			width = displayWidth
		}

		modal.Table.SetCell(index-skipped, 0, tview.NewTableCell(display).
			SetExpansion(1).
			SetReference(&menuopt).
			SetAlign(tview.AlignLeft).
			SetTextColor(theme.GetColor(theme.ThemeMenuItem)).
			SetClickedFunc(clickedFunc).
			SetSelectedStyle(tcell.Style{}.
				Foreground(theme.GetColor(theme.ThemeMenuItem)).
				Background(theme.BackgroundColor(theme.ThemeMenuItem)),
			),
		)
		modal.Table.SetCell(index-skipped, 1, tview.NewTableCell(keybinding).
			SetExpansion(1).
			SetAlign(tview.AlignRight).
			SetClickedFunc(clickedFunc).
			SetTextColor(theme.GetColor(theme.ThemeMenuItem)).
			SetSelectedStyle(tcell.Style{}.
				Foreground(theme.GetColor(theme.ThemeMenuItem)).
				Background(theme.BackgroundColor(theme.ThemeMenuItem)),
			),
		)
	}

	modal.Table.Select(0, 0)

	m.drawSubMenu(x, y, width, device != nil)
}

// drawContextMenu sets up a selector menu.
func (m *menuBarView) drawContextMenu(
	menuID string,
	selected func(table *tview.Table),
	changed func(table *tview.Table, row, col int),
	listContents func(table *tview.Table) (int, int),
) {
	var changeEnabled bool

	x, y := 0, 1

	modal := m.modal
	modal.Table.Clear()
	modal.Table.SetSelectorWrap(false)
	modal.Table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch m.kb.Key(event) {
		case keybindings.KeySelect:
			if selected != nil {
				selected(modal.Table)
			}

			fallthrough

		case keybindings.KeyClose:
			m.exit()
		}

		return ignoreDefaultEvent(event)
	})
	modal.Table.SetSelectionChangedFunc(func(row, col int) {
		if changed == nil {
			return
		}

		if !changeEnabled {
			changeEnabled = true
			return
		}

		changed(modal.Table, row, col)
	})
	modal.Table.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		if action == tview.MouseLeftClick && modal.Table.InRect(event.Position()) {
			m.exit()
		}

		return action, event
	})

	modal.Name = menuID
	width, index := listContents(modal.Table)

	modal.Table.Select(index, 0)
	m.drawSubMenu(x, y, width+20, menuID == menuDeviceName.String())
}

// drawSubMenu draws the submenu.
func (m *menuBarView) drawSubMenu(x, y, width int, device bool) {
	if m.modal.Open {
		m.exit(struct{}{})
	}

	if device {
		_, _, _, tableHeight := m.device.table.GetInnerRect()
		deviceX, deviceY := getSelectionXY(m.device.table)

		x = deviceX + 10
		if deviceY >= tableHeight-6 {
			y = deviceY - m.modal.Table.GetRowCount()
		} else {
			y = deviceY + 1
		}
	}

	m.modal.Height = m.modal.Table.GetRowCount() + 2
	m.modal.Width = width

	m.modal.regionX = x
	m.modal.regionY = y

	m.modal.Show()
}

// setHeader appends the header text with
// the menu bar's regions.
func (m *menuBarView) setHeader(header string) {
	m.bar.SetText(header + "[-:-:-] " + theme.ColorWrap(theme.ThemeMenu, menuBarRegions))
}

// next switches between menus.
func (m *menuBarView) next() {
	highlighted := m.bar.GetHighlights()
	if highlighted == nil {
		return
	}

	for _, region := range m.bar.GetRegionIDs() {
		if highlighted[0] != region && highlighted[0] != menuAdapterChangeName.String() {
			m.bar.Highlight(region)
		}
	}
}

// exit exits the menu.
func (m *menuBarView) exit(highlight ...struct{}) {
	m.modal.Exit(false)

	if highlight == nil {
		m.modal.Name = MenuBarName.String()
		m.bar.Highlight("")
	}

	m.app.FocusPrimitive(m.device.table)
}

// inputHandler handles key events for a submenu.
func (m *menuBarView) inputHandler(event *tcell.EventKey) {
	key := m.kb.Key(event, m.pages.currentContext(), keybindings.ContextProgress)
	if key == "" {
		return
	}

	option, ok := m.optionByKey[key]
	if !ok {
		return
	}

	if option.checkVisibility && !m.actions.handler(key, actionVisibility)() {
		return
	}

	m.actions.handler(key, actionInvoke)()
}

func (m *menuBarView) initOrderedOptions() {
	m.menuOptions = map[string][]menuOption{
		menuAdapterName.String(): {
			{
				key:              keybindings.KeyAdapterTogglePower,
				enabledText:      "On",
				disabledText:     "Off",
				initBeforeInvoke: true,
			},
			{
				key:              keybindings.KeyAdapterToggleDiscoverable,
				enabledText:      "On",
				disabledText:     "Off",
				initBeforeInvoke: true,
			},
			{
				key:              keybindings.KeyAdapterTogglePairable,
				enabledText:      "On",
				disabledText:     "Off",
				initBeforeInvoke: true,
			},
			{
				key:          keybindings.KeyAdapterToggleScan,
				disabledText: "Stop Scan",
			},
			{
				key: keybindings.KeyAdapterChange,
			},
			{
				key: keybindings.KeyProgressView,
			},
			{
				key: keybindings.KeyPlayerHide,
			},
			{
				key: keybindings.KeyQuit,
			},
		},
		menuDeviceName.String(): {
			{
				key:              keybindings.KeyDeviceConnect,
				disabledText:     "Disconnect",
				initBeforeInvoke: true,
			},
			{
				key: keybindings.KeyDevicePair,
			},
			{
				key:              keybindings.KeyDeviceTrust,
				disabledText:     "Untrust",
				initBeforeInvoke: true,
			},
			{
				key:              keybindings.KeyDeviceBlock,
				disabledText:     "Unblock",
				initBeforeInvoke: true,
			},
			{
				key:             keybindings.KeyDeviceSendFiles,
				checkVisibility: true,
			},
			{
				key:             keybindings.KeyDeviceNetwork,
				checkVisibility: true,
			},
			{
				key:             keybindings.KeyDeviceAudioProfiles,
				checkVisibility: true,
			},
			{
				key:             keybindings.KeyPlayerShow,
				checkVisibility: true,
			},
			{
				key: keybindings.KeyDeviceInfo,
			},
			{
				key: keybindings.KeyDeviceRemove,
			},
		},
	}
}
