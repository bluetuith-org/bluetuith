package views

import (
	"context"

	"github.com/darkhz/bluetuith/ui/keybindings"
	"github.com/darkhz/bluetuith/ui/theme"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
)

// modalViews holds all the displayed modals on the screen.
type modalViews struct {
	modals []*modalView

	rv *Views
}

// Initialize initializes the modals view.
func (m *modalViews) Initialize() error {
	m.modals = make([]*modalView, 0, 10)

	return nil
}

// SetRootView sets the root view of the modals view.
func (m *modalViews) SetRootView(v *Views) {
	m.rv = v
}

// newModal returns a modal.
func (m *modalViews) newModal(name, title string, item tview.Primitive, height, width int) *modalView {
	var modal *modalView

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
	closeButton.SetHighlightedFunc(func(added, _, _ []string) {
		if added == nil {
			return
		}

		modal.remove(false)
	})

	titleFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(box, 0, 1, false).
		AddItem(modalTitle, 0, 10, false).
		AddItem(closeButton, 0, 1, false)

	flex := tview.NewFlex()
	flex.SetBorder(true)
	flex.SetDirection(tview.FlexRow)

	flex.AddItem(titleFlex, 1, 0, false)
	flex.AddItem(horizontalLine(), 1, 0, false)
	flex.AddItem(item, 0, 1, true)
	flex.SetBorderColor(theme.GetColor(theme.ThemeBorder))
	flex.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	modal = &modalView{
		name: name,
		flex: flex,

		height: height,
		width:  width,

		closeButton: closeButton,
		mgr:         m,
	}

	return modal
}

// newModalWithTable returns a new modal with an embedded table.
func (m *modalViews) newModalWithTable(name, title string, height, width int) *tableModalView {
	var modal *modalView

	table := tview.NewTable()
	table.SetSelectorWrap(true)
	table.SetSelectable(true, false)
	table.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))
	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if m.rv.kb.Key(event) == keybindings.KeyClose {
			modal.remove(false)
		}

		return ignoreDefaultEvent(event)
	})

	modal = m.newModal(name, title, table, height, width)

	return &tableModalView{
		table:     table,
		modalView: modal,
	}
}

// newMenuModal returns a menu modal.
func (m *modalViews) newMenuModal(name string, regionX, regionY int) *tableModalView {
	table := tview.NewTable()
	table.SetBorder(true)
	table.SetSelectable(true, false)
	table.SetBackgroundColor(tcell.ColorDefault)
	table.SetBorderColor(theme.GetColor(theme.ThemeBorder))
	table.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(table, 0, 1, true)

	return &tableModalView{
		table: table,
		modalView: &modalView{
			name: name,
			flex: flex,

			isMenu:  true,
			regionX: regionX,
			regionY: regionY,
			mgr:     m,
		},
	}

}

// newDisplayModal returns a display modal.
func (m *modalViews) newDisplayModal(name, title, message string) *displayModalView {
	message += "\n\nPress any key or click the 'X' button to close this dialog."

	width, height := m.getModalDimensions(message, "")

	textview := tview.NewTextView()
	textview.SetText(message)
	textview.SetDynamicColors(true)
	textview.SetTextAlign(tview.AlignCenter)
	textview.SetTextColor(theme.GetColor(theme.ThemeText))
	textview.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	return &displayModalView{
		textview:  textview,
		modalView: m.newModal(name, title, textview, width, height),
	}
}

// newConfirmModal returns a confirmation modal.
func (m *modalViews) newConfirmModal(name, title, message string) *confirmModalView {
	message += "\n\nPress y/n to Confirm/Cancel, click the required button or click the 'X' button to close this dialog."
	buttonsText := `["confirm"][::b][Confirm[] ["cancel"][::b][Cancel[]`

	width, height := m.getModalDimensions(message, buttonsText)

	buttons := tview.NewTextView()
	buttons.SetRegions(true)
	buttons.SetText(buttonsText)
	buttons.SetDynamicColors(true)
	buttons.SetTextAlign(tview.AlignCenter)
	buttons.SetTextColor(theme.GetColor(theme.ThemeText))
	buttons.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

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

	return &confirmModalView{
		textview:  textview,
		buttons:   buttons,
		modalView: m.newModal(name, title, flex, height, width),
	}
}

// isModalDisplayed checks if the specified modal is displayed.
func (m *modalViews) isModalDisplayed(modalName string) bool {
	for _, modal := range m.modals {
		if modal.name == modalName {
			return true
		}
	}

	return false
}

// removeModal removes the specified modal from the screen.
func (m *modalViews) removeModal(currentModal *modalView, focusInput bool) {
	m.rv.pages.RemovePage(currentModal.name)

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

	m.setPrimaryFocus()
}

// displayModal displays the specified modal to the screen.
func (m *modalViews) displayModal(modal *modalView) {
	m.rv.pages.AddAndSwitchToPage(modal.name, modal.x, true)
	for _, modal := range m.modals {
		m.rv.pages.ShowPage(modal.name)
	}
	m.rv.pages.ShowPage(m.rv.pages.currentPage())
	m.rv.app.FocusPrimitive(modal.flex)

	m.modals = append(m.modals, modal)

	m.resizeModal()
}

// resizeModal resizes all the displayed modals according to the current screen dimensions.
func (m *modalViews) resizeModal() {
	var drawn bool

	for _, modal := range m.modals {
		_, _, pageWidth, pageHeight := m.rv.layout.GetInnerRect()

		if modal == nil || !modal.isOpen ||
			(modal.pageHeight == pageHeight && modal.pageWidth == pageWidth) {
			continue
		}

		modal.pageHeight = pageHeight
		modal.pageWidth = pageWidth

		height := modal.height
		width := modal.width
		if height >= pageHeight {
			height = pageHeight
		}
		if width >= pageWidth {
			width = pageWidth
		}

		var x, y int

		if modal.isMenu {
			x, y = modal.regionX, modal.regionY
		} else {
			x = (pageWidth - modal.width) / 2
			y = (pageHeight - modal.height) / 2
		}

		modal.y.ResizeItem(modal.flex, height, 0)
		modal.y.ResizeItem(nil, y, 0)

		modal.x.ResizeItem(modal.y, width, 0)
		modal.x.ResizeItem(nil, x, 0)

		drawn = true
	}

	if drawn {
		go m.rv.app.Refresh()
	}
}

// setPrimaryFocus sets the focus to the appropriate primitive.
func (m *modalViews) setPrimaryFocus() {
	if pg, _ := m.rv.status.GetFrontPage(); pg == statusInputPage.String() {
		m.rv.app.FocusPrimitive(m.rv.status.InputField)
		return
	}

	if len(m.modals) > 0 {
		m.rv.app.FocusPrimitive(m.modals[len(m.modals)-1].flex)
		return
	}

	m.rv.app.FocusPrimitive(m.rv.pages)
}

// getModal returns whether the getModal with the given name is displayed.
func (m *modalViews) getModal(name string) (*modalView, bool) {
	for _, modal := range m.modals {
		if modal.name == name {
			return modal, true
		}
	}

	return nil, false
}

// modalMouseHandler handles mouse events for a modal.
func (m *modalViews) modalMouseHandler(event *tcell.EventMouse, action tview.MouseAction) (*tcell.EventMouse, tview.MouseAction) {
	for _, modal := range m.modals {
		if modal == nil || !modal.isOpen {
			continue
		}

		x, y := event.Position()

		switch action {
		case tview.MouseRightClick:
			if m.rv.menu.bar.InRect(x, y) {
				return nil, action
			}

		case tview.MouseLeftClick:
			if modal.flex.InRect(x, y) {
				m.rv.app.FocusPrimitive(modal.flex)
			} else {
				if modal.isMenu {
					m.rv.menu.exit()
					continue
				}

				modal.remove(false)
			}
		}
	}

	return event, action
}

// getModalDimensions returns the height and width to set for the modal
// according to the provided text.
// Adapted from: https://github.com/rivo/tview/blob/1b91b8131c43011d923fe59855b4de3571dac997/modal.go#L156
func (m *modalViews) getModalDimensions(text, buttons string) (width int, height int) {
	_, _, screenWidth, _ := m.rv.pages.GetRect()
	buttonWidth := tview.TaggedStringWidth(buttons) - 2

	width = max(screenWidth/3, buttonWidth)

	padding := 6
	if buttonWidth < 0 {
		padding -= 2
	}

	return width + 4, len(tview.WordWrap(text, width)) + padding
}

// modalView stores a layout to display a floating modal.
type modalView struct {
	name           string
	isOpen, isMenu bool

	height, width         int
	pageHeight, pageWidth int

	y, x             *tview.Flex
	regionX, regionY int

	flex        *tview.Flex
	closeButton *tview.TextView

	mgr *modalViews
}

// show shows the modal onto the screen.
func (m *modalView) show() {
	var x, y, xprop, xattach int

	if m.mgr.isModalDisplayed(m.name) {
		return
	}

	switch {
	case m.isMenu:
		xprop = 1
		x, y = m.regionX, m.regionY

	default:
		xattach = 1
	}

	m.isOpen = true

	m.y = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, y, 0, false).
		AddItem(m.flex, m.height, 0, true).
		AddItem(nil, 1, 0, false)

	m.x = tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(nil, x, xattach, false).
		AddItem(m.y, m.width, 0, true).
		AddItem(nil, xprop, xattach, false)

	m.mgr.displayModal(m)
}

// remove removes the modal from the screen.
func (m *modalView) remove(focusInput bool) {
	if m == nil {
		return
	}

	m.isOpen = false
	m.pageWidth = 0
	m.pageHeight = 0

	m.mgr.removeModal(m, focusInput)
}

// tableModalView holds a modal view with an embedded table.
type tableModalView struct {
	table *tview.Table

	*modalView
}

// textModalView holds a modal with an embedded textview and buttons.
type textModalView struct {
	textview *tview.TextView
	buttons  *tview.TextView

	*modalView
}

// displayModalView holds a modal which only displays certain text.
type displayModalView textModalView

// display displays the modal, and waits for the provided context to be completed or a user input
// to close the modal.
func (d *displayModalView) display(ctx context.Context) {
	reply := make(chan struct{}, 1)

	d.textview.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		d.remove(false)

		return event
	})

	go d.mgr.rv.app.QueueDraw(func() {
		if m, ok := d.mgr.getModal(d.name); ok {
			m.remove(false)
		}

		d.show()
	})

	select {
	case <-ctx.Done():
		d.closeButton.InputHandler()(nil, nil)

	case <-reply:
		d.closeButton.InputHandler()(nil, nil)
	}
}

// confirmModalView holds a modal which displays a message and asks for confirmation from the user.
type confirmModalView textModalView

// getReply displays a modal, and waits for the user to confirm or the provided context to be completed
// to close the modal.
func (c *confirmModalView) getReply(ctx context.Context) string {
	reply := make(chan string, 1)

	send := func(msg string) {
		c.remove(false)

		reply <- msg
	}

	c.buttons.SetHighlightedFunc(func(added, _, _ []string) {
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
	c.buttons.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'y', 'n':
			send(string(event.Rune()))
		}

		if c.mgr.rv.kb.Key(event) == keybindings.KeyClose {
			send("n")
		}

		return event
	})
	c.closeButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		send("n")
		return event
	})

	go c.mgr.rv.app.QueueDraw(func() {
		if m, ok := c.mgr.getModal(c.name); ok {
			m.remove(false)
		}
		c.show()
	})

	select {
	case <-ctx.Done():
		c.closeButton.InputHandler()(nil, nil)

	case r := <-reply:
		c.closeButton.InputHandler()(nil, nil)
		return r
	}

	return ""
}
