package views

import (
	"github.com/darkhz/bluetuith/ui/keybindings"
	"github.com/darkhz/bluetuith/ui/theme"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
)

// modalView stores a layout to display a floating modal.
type modalView struct {
	Name          string
	Open          bool
	Height, Width int

	menu                                    bool
	regionX, regionY, pageHeight, pageWidth int

	Flex  *tview.Flex
	Table *tview.Table

	y      *tview.Flex
	x      *tview.Flex
	button *tview.TextView

	mgr *modalViews
}

// Show shows the modal.
func (m *modalView) Show() {
	var x, y, xprop, xattach int

	if m.mgr.checkIfDisplayed(m.Name) {
		return
	}

	switch {
	case m.menu:
		xprop = 1
		x, y = m.regionX, m.regionY

	default:
		xattach = 1
	}

	m.Open = true

	m.y = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, y, 0, false).
		AddItem(m.Flex, m.Height, 0, true).
		AddItem(nil, 1, 0, false)

	m.x = tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(nil, x, xattach, false).
		AddItem(m.y, m.Width, 0, true).
		AddItem(nil, xprop, xattach, false)

	m.mgr.displayModal(m)
}

// Exit exits the modal.
func (m *modalView) Exit(focusInput bool) {
	if m == nil {
		return
	}

	m.Open = false
	m.pageWidth = 0
	m.pageHeight = 0

	m.mgr.removeModal(m, focusInput)
}

type modalViews struct {
	modals []*modalView

	rv *Views
}

func (m *modalViews) Initialize() error {
	m.modals = make([]*modalView, 0, 10)

	return nil
}

func (m *modalViews) SetRootView(v *Views) {
	m.rv = v
}

// NewModal returns a modal. If a primitive is not provided,
// a table is attach to it.
func (m *modalViews) NewModal(name, title string, item tview.Primitive, height, width int) *modalView {
	var modal *modalView
	var table *tview.Table

	box := tview.NewBox()
	box.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	modalTitle := tview.NewTextView()
	modalTitle.SetDynamicColors(true)
	modalTitle.SetText("[::bu]" + title)
	modalTitle.SetTextAlign(tview.AlignCenter)
	modalTitle.SetTextColor(theme.GetColor(theme.ThemeText))
	modalTitle.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	closeButton := tview.NewTextView()
	closeButton.SetRegions(true)
	closeButton.SetDynamicColors(true)
	closeButton.SetText(`["close"][::b][X[]`)
	closeButton.SetTextAlign(tview.AlignRight)
	closeButton.SetTextColor(theme.GetColor(theme.ThemeText))
	closeButton.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))
	closeButton.SetHighlightedFunc(func(added, removed, remaining []string) {
		if added == nil {
			return
		}

		modal.Exit(false)
	})

	titleFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(box, 0, 1, false).
		AddItem(modalTitle, 0, 10, false).
		AddItem(closeButton, 0, 1, false)

	if item == nil {
		table = tview.NewTable()
		table.SetSelectorWrap(true)
		table.SetSelectable(true, false)
		table.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))
		table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch m.rv.kb.Key(event) {
			case keybindings.KeyClose:
				modal.Exit(false)
			}

			return ignoreDefaultEvent(event)
		})

		item = table
	}

	flex := tview.NewFlex()
	flex.SetBorder(true)
	flex.SetDirection(tview.FlexRow)

	flex.AddItem(titleFlex, 1, 0, false)
	flex.AddItem(horizontalLine(), 1, 0, false)
	flex.AddItem(item, 0, 1, true)
	flex.SetBorderColor(theme.GetColor(theme.ThemeBorder))
	flex.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	modal = &modalView{
		Name:  name,
		Flex:  flex,
		Table: table,

		Height: height,
		Width:  width,

		button: closeButton,
		mgr:    m,
	}

	return modal
}

// NewMenuModal returns a menu modal.
func (m *modalViews) NewMenuModal(name string, regionX, regionY int) *modalView {
	table := tview.NewTable()
	table.SetBorder(true)
	table.SetSelectable(true, false)
	table.SetBackgroundColor(tcell.ColorDefault)
	table.SetBorderColor(theme.GetColor(theme.ThemeBorder))
	table.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(table, 0, 1, true)

	return &modalView{
		Name:  name,
		Table: table,
		Flex:  flex,

		menu:    true,
		regionX: regionX,
		regionY: regionY,
		mgr:     m,
	}
}

// NewDisplayModal displays a modal with a message.
func (m *modalViews) NewDisplayModal(name, title, message string) {
	message += "\n\nPress any key or click the 'X' button to close this dialog."

	width, height := m.getModalDimensions(message, "")

	textview := tview.NewTextView()
	textview.SetText(message)
	textview.SetDynamicColors(true)
	textview.SetTextAlign(tview.AlignCenter)
	textview.SetTextColor(theme.GetColor(theme.ThemeText))
	textview.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	modal := m.NewModal(name, title, textview, width, height)
	textview.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		modal.Exit(false)

		return event
	})

	go m.rv.app.InstantDraw(func() {
		if m, ok := m.ModalExists(name); ok {
			m.Exit(false)
		}

		modal.Show()
	})
}

// NewConfirmModal displays a modal, shows a message and asks for confirmation.
func (m *modalViews) NewConfirmModal(name, title, message string) string {
	var modal *modalView

	message += "\n\nPress y/n to Confirm/Cancel, click the required button or click the 'X' button to close this dialog."
	buttonsText := `["confirm"][::b][Confirm[] ["cancel"][::b][Cancel[]`

	reply := make(chan string, 10)

	send := func(msg string) {
		modal.Exit(false)

		reply <- msg
	}

	width, height := m.getModalDimensions(message, buttonsText)

	buttons := tview.NewTextView()
	buttons.SetRegions(true)
	buttons.SetText(buttonsText)
	buttons.SetDynamicColors(true)
	buttons.SetTextAlign(tview.AlignCenter)
	buttons.SetTextColor(theme.GetColor(theme.ThemeText))
	buttons.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))
	buttons.SetHighlightedFunc(func(added, removed, remaining []string) {
		if added == nil {
			return
		}

		switch added[0] {
		case "confirm":
			send("y")

		case "cancel":
			send("n")
		}
	})

	textview := tview.NewTextView()
	textview.SetText(message)
	textview.SetDynamicColors(true)
	textview.SetTextAlign(tview.AlignCenter)
	textview.SetTextColor(theme.GetColor(theme.ThemeText))
	textview.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(textview, 0, 1, false).
		AddItem(buttons, 1, 0, true)

	modal = m.NewModal(name, title, flex, height, width)
	buttons.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'y', 'n':
			send(string(event.Rune()))
		}

		switch m.rv.kb.Key(event) {
		case keybindings.KeyClose:
			send("n")
		}

		return event
	})
	modal.button.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		send("n")
		return event
	})

	go m.rv.app.InstantDraw(func() {
		if m, ok := m.ModalExists(name); ok {
			m.Exit(false)
		}

		modal.Show()
	})

	return <-reply
}

func (m *modalViews) checkIfDisplayed(modalName string) bool {
	return len(m.modals) > 0 && m.modals[len(m.modals)-1].Name == modalName
}

func (m *modalViews) removeModal(currentModal *modalView, focusInput bool) {
	m.rv.pages.RemovePage(currentModal.Name)

	for i, modal := range m.modals {
		if modal == currentModal {
			m.modals[i] = m.modals[len(m.modals)-1]
			m.modals = m.modals[:len(m.modals)-1]

			break
		}
	}

	if focusInput {
		m.rv.app.FocusPrimitive(m.rv.status.InputField)
		return
	}

	m.SetPrimaryFocus()
}

func (m *modalViews) displayModal(modal *modalView) {
	m.rv.pages.AddAndSwitchToPage(modal.Name, modal.x, true)
	for _, modal := range m.modals {
		m.rv.pages.ShowPage(modal.Name)
	}
	m.rv.pages.ShowPage(m.rv.pages.currentPage())
	m.rv.app.FocusPrimitive(modal.Flex)

	m.modals = append(m.modals, modal)

	m.ResizeModal()
}

// ResizeModal resizes the modal according to the current screen dimensions.
func (m *modalViews) ResizeModal() {
	var drawn bool

	for _, modal := range m.modals {
		_, _, pageWidth, pageHeight := m.rv.layout.GetInnerRect()

		if modal == nil || !modal.Open ||
			(modal.pageHeight == pageHeight && modal.pageWidth == pageWidth) {
			continue
		}

		modal.pageHeight = pageHeight
		modal.pageWidth = pageWidth

		height := modal.Height
		width := modal.Width
		if height >= pageHeight {
			height = pageHeight
		}
		if width >= pageWidth {
			width = pageWidth
		}

		var x, y int

		if modal.menu {
			x, y = modal.regionX, modal.regionY
		} else {
			x = (pageWidth - modal.Width) / 2
			y = (pageHeight - modal.Height) / 2
		}

		modal.y.ResizeItem(modal.Flex, height, 0)
		modal.y.ResizeItem(nil, y, 0)

		modal.x.ResizeItem(modal.y, width, 0)
		modal.x.ResizeItem(nil, x, 0)

		drawn = true
	}

	if drawn {
		go m.rv.app.Refresh()
	}
}

// SetPrimaryFocus sets the focus to the appropriate primitive.
func (m *modalViews) SetPrimaryFocus() {
	if pg, _ := m.rv.status.GetFrontPage(); pg == statusInputPage.String() {
		m.rv.app.FocusPrimitive(m.rv.status.InputField)
		return
	}

	if len(m.modals) > 0 {
		m.rv.app.FocusPrimitive(m.modals[len(m.modals)-1].Flex)
		return
	}

	m.rv.app.FocusPrimitive(m.rv.pages)
}

// ModalExists returns whether the modal with the given name is displayed.
func (m *modalViews) ModalExists(name string) (*modalView, bool) {
	for _, modal := range m.modals {
		if modal.Name == name {
			return modal, true
		}
	}

	return nil, false
}

// modalMouseHandler handles mouse events for a modal.
func (m *modalViews) modalMouseHandler(event *tcell.EventMouse, action tview.MouseAction) (*tcell.EventMouse, tview.MouseAction) {
	for _, modal := range m.modals {
		if modal == nil || !modal.Open {
			continue
		}

		x, y := event.Position()

		switch action {
		case tview.MouseRightClick:
			if m.rv.menu.bar.InRect(x, y) {
				return nil, action
			}

		case tview.MouseLeftClick:
			if modal.Flex.InRect(x, y) {
				m.rv.app.FocusPrimitive(modal.Flex)
			} else {
				if modal.menu {
					m.rv.menu.exit()
					continue
				}

				modal.Exit(false)
			}
		}
	}

	return event, action
}

// getModalDimensions returns the height and width to set for the modal
// according to the provided text.
// Adapted from: https://github.com/rivo/tview/blob/1b91b8131c43011d923fe59855b4de3571dac997/modal.go#L156
func (m *modalViews) getModalDimensions(text, buttons string) (int, int) {
	_, _, screenWidth, _ := m.rv.pages.GetRect()
	buttonWidth := tview.TaggedStringWidth(buttons) - 2

	width := max(screenWidth/3, buttonWidth)

	padding := 6
	if buttonWidth < 0 {
		padding -= 2
	}

	return width + 4, len(tview.WordWrap(text, width)) + padding
}
