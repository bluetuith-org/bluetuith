package views

import (
	"errors"
	"strconv"
	"strings"

	"github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	"github.com/darkhz/bluetuith/ui/keybindings"
	"github.com/darkhz/bluetuith/ui/theme"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
)

const devicePage viewName = "devices"

// deviceView holds the devices view.
type deviceView struct {
	table *tview.Table

	*Views
}

// Initialize initializes the devices view.
func (d *deviceView) Initialize() error {
	d.table = tview.NewTable()
	d.table.SetSelectorWrap(true)
	d.table.SetSelectable(true, false)
	d.table.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))
	d.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch d.kb.Key(event) {
		case keybindings.KeyMenu:
			d.menu.highlight(menuAdapterName)
			return event

		case keybindings.KeyHelp:
			d.help.showHelp()
			return event
		}

		d.player.keyEvents(event)

		d.menu.inputHandler(event)

		return ignoreDefaultEvent(event)
	})
	d.table.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		if action == tview.MouseRightClick && d.table.HasFocus() {
			device := d.getSelection(false)
			if device.Address.IsNil() {
				return action, event
			}

			d.menu.setupSubMenu(0, 0, menuDeviceName, struct{}{})
		}

		return action, event
	})

	d.list()
	d.connectByAddress()
	go d.event()

	return nil
}

// SetRootView sets the root view of the devices view.
func (d *deviceView) SetRootView(v *Views) {
	d.Views = v
}

// list lists the devices belonging to the selected adapter within the devices view.
func (d *deviceView) list() {
	devices, err := d.adapter.currentSession().Devices()
	if err != nil {
		d.status.ErrorMessage(err)
		return
	}

	d.table.Clear()
	for i, device := range devices {
		d.setInfo(i, device)
	}
	d.table.Select(0, 0)
}

// connectByAddress connects to a device based on the provided address
// which was parsed from the "connect-bdaddr" command-line option.
func (d *deviceView) connectByAddress() {
	address := d.cfg.Values.AutoConnectDeviceAddr
	if address.IsNil() {
		return
	}

	go d.actions.connect(address.String())
}

// showDetailedInfo shows detailed information about a device.
func (d *deviceView) showDetailedInfo() {
	device := d.getSelection(false)
	if device.Address.IsNil() {
		return
	}

	yesno := func(val bool) string {
		if !val {
			return "no"
		}

		return "yes"
	}

	assocAdapter, err := d.app.Session().Adapter(device.AssociatedAdapter).Properties()
	if err != nil {
		d.status.ErrorMessage(err)
		return
	}

	device, err = d.app.Session().Device(device.Address).Properties()
	if err != nil {
		d.status.ErrorMessage(err)
		return
	}

	props := [][]string{
		{"Name", device.Name},
		{"Alias", device.Alias},
		{"Address", device.Address.String()},
		{"Class", strconv.FormatUint(uint64(device.Class), 10)},
		{"Adapter", assocAdapter.UniqueName},
		{"Connected", yesno(device.Connected)},
		{"Paired", yesno(device.Paired)},
		{"Bonded", yesno(device.Bonded)},
		{"Trusted", yesno(device.Trusted)},
		{"Blocked", yesno(device.Blocked)},
		{"LegacyPairing", yesno(device.LegacyPairing)},
	}
	props = append(props, []string{"UUIDs", ""})

	infoModal := d.modals.newModalWithTable("info", "Device Information", 40, 100)
	infoModal.table.SetSelectionChangedFunc(func(row, _ int) {
		_, _, _, height := infoModal.table.GetRect()
		infoModal.table.SetOffset(row-((height-1)/2), 0)
	})

	for i, prop := range props {
		propName := prop[0]
		propValue := prop[1]

		if propName == "Class" {
			propValue += " (" + device.Type + ")"
		}

		infoModal.table.SetCell(i, 0, tview.NewTableCell("[::b]"+propName+":").
			SetExpansion(1).
			SetAlign(tview.AlignLeft).
			SetTextColor(theme.GetColor(theme.ThemeText)).
			SetSelectedStyle(tcell.Style{}.
				Bold(true).
				Underline(true),
			),
		)

		infoModal.table.SetCell(i, 1, tview.NewTableCell(propValue).
			SetExpansion(1).
			SetAlign(tview.AlignLeft).
			SetTextColor(theme.GetColor(theme.ThemeText)),
		)
	}

	rows := infoModal.table.GetRowCount() - 1
	for i, serviceUUID := range device.UUIDs {
		serviceType := bluetooth.ServiceType(serviceUUID)
		serviceUUID = "(" + serviceUUID + ")"

		infoModal.table.SetCell(rows+i, 1, tview.NewTableCell(serviceType).
			SetExpansion(1).
			SetAlign(tview.AlignLeft).
			SetTextColor(theme.GetColor(theme.ThemeText)),
		)

		infoModal.table.SetCell(rows+i, 2, tview.NewTableCell(serviceUUID).
			SetExpansion(0).
			SetTextColor(theme.GetColor(theme.ThemeText)),
		)
	}

	infoModal.height = min(infoModal.table.GetRowCount()+4, 60)

	infoModal.show()
}

// getSelection retrieves device information from the current selection in the devices view.
func (d *deviceView) getSelection(lock bool) bluetooth.DeviceData {
	var device bluetooth.DeviceData

	getdevice := func() {
		row, _ := d.table.GetSelection()

		cell := d.table.GetCell(row, 0)
		if cell == nil {
			device = bluetooth.DeviceData{}
		}

		d, ok := cell.GetReference().(bluetooth.DeviceData)
		if !ok {
			device = bluetooth.DeviceData{}
		}

		device = d
	}

	if lock {
		d.app.QueueDraw(func() {
			getdevice()
		})

		return device
	}

	getdevice()

	return device
}

// getRowByAddress iterates through the devices view and checks
// if a device whose path matches the path parameter exists.
func (d *deviceView) getRowByAddress(address bluetooth.MacAddress) (int, bool) {
	for row := range d.table.GetRowCount() {
		cell := d.table.GetCell(row, 0)
		if cell == nil {
			continue
		}

		ref, ok := cell.GetReference().(bluetooth.DeviceData)
		if !ok {
			continue
		}

		if ref.Address == address {
			return row, true
		}
	}

	return -1, false
}

// setInfo writes device information into the specified row of the devices view.
func (d *deviceView) setInfo(row int, device bluetooth.DeviceData) {
	var sb strings.Builder

	name := device.Name
	if name == "" {
		name = device.Address.String()
	}
	sb.WriteString(name)
	sb.WriteString(" (")
	if device.Alias != "" && device.Alias != device.Name {
		sb.WriteString(theme.ColorWrap(theme.ThemeDeviceAlias, device.Alias))
		sb.WriteString(", ")
	}
	sb.WriteString(theme.ColorWrap(theme.ThemeDeviceType, device.Type))
	sb.WriteString(")")

	nameDisplay := sb.String()
	nameColor := theme.ThemeDevice

	d.table.SetCell(
		row, 0, tview.NewTableCell(nameDisplay).
			SetExpansion(1).
			SetReference(device).
			SetAlign(tview.AlignLeft).
			SetAttributes(tcell.AttrBold).
			SetTextColor(theme.GetColor(nameColor)).
			SetSelectedStyle(tcell.Style{}.
				Foreground(theme.GetColor(nameColor)).
				Background(theme.BackgroundColor(nameColor)),
			),
	)

	d.setPropertyInfo(row, device.DeviceEventData, false)
}

// setPropertyInfo writes the property information of a device into the specified row of the devices view.
func (d *deviceView) setPropertyInfo(row int, deviceEvent bluetooth.DeviceEventData, isPartialUpdate bool) {
	var sb strings.Builder

	nameColor := theme.ThemeDevice
	propColor := theme.ThemeDeviceProperty

	sb.WriteString("(")

	appendProperty := func(prop string) {
		if sb.Len() == 1 {
			sb.WriteString(prop)
			return
		}

		sb.WriteString(", ")
		sb.WriteString(prop)
	}

	if deviceEvent.Connected {
		appendProperty("Connected")

		nameColor = theme.ThemeDeviceConnected
		propColor = theme.ThemeDevicePropertyConnected

		if deviceEvent.RSSI < 0 {
			rssi := strconv.FormatInt(int64(deviceEvent.RSSI), 10)
			sb.WriteString(" [")
			sb.WriteString(rssi)
			sb.WriteString("[]")
		}

		if deviceEvent.Percentage > 0 {
			appendProperty("Battery ")
			sb.WriteString(strconv.Itoa(deviceEvent.Percentage))
			sb.WriteString("%")
		}
	}
	if deviceEvent.Trusted {
		appendProperty("Trusted")
	}
	if deviceEvent.Blocked {
		appendProperty("Blocked")
	}

	if deviceEvent.Bonded && deviceEvent.Paired {
		appendProperty("Bonded")
	} else if !deviceEvent.Bonded && deviceEvent.Paired {
		appendProperty("Paired")
	}

	if sb.Len() == 1 {
		sb.Reset()
		sb.WriteString("[New Device[]")
		nameColor = theme.ThemeDeviceDiscovered
		propColor = theme.ThemeDevicePropertyDiscovered
	} else {
		sb.WriteString(")")
	}

	deviceNameCell := d.table.GetCell(row, 0)
	if deviceNameCell != nil {
		deviceNameCell.SetTextColor(theme.GetColor(nameColor))
	}

	if isPartialUpdate {
		device, ok := deviceNameCell.GetReference().(bluetooth.DeviceData)
		if ok {
			device.DeviceEventData = deviceEvent
			deviceNameCell.SetReference(device)
		}

		devicePropertyCell := d.table.GetCell(row, 1)
		if devicePropertyCell != nil {
			devicePropertyCell.SetText(sb.String())
			devicePropertyCell.SetTextColor(theme.GetColor(propColor))
		}

		return
	}

	d.table.SetCell(
		row, 1, tview.NewTableCell(sb.String()).
			SetExpansion(1).
			SetAlign(tview.AlignRight).
			SetTextColor(theme.GetColor(propColor)).
			SetSelectedStyle(tcell.Style{}.
				Bold(true),
			),
	)
}

// event handles device-specific events.
func (d *deviceView) event() {
	deviceSub, ok := bluetooth.DeviceEvents().Subscribe()
	if !ok {
		d.status.ErrorMessage(errors.New("cannot subscribe to device events"))
		return
	}

	for {
		select {
		case <-deviceSub.Done:
			return

		case ev := <-deviceSub.AddedEvents:
			go d.app.QueueDraw(func() {
				deviceRow := d.table.GetRowCount()

				row, ok := d.getRowByAddress(ev.Address)
				if ok {
					deviceRow = row
				}
				d.setInfo(deviceRow, ev)
			})

		case ev := <-deviceSub.UpdatedEvents:
			go d.app.QueueDraw(func() {
				row, ok := d.getRowByAddress(ev.Address)
				if ok {
					d.setPropertyInfo(row, ev, true)

				}
			})

		case ev := <-deviceSub.RemovedEvents:
			go d.app.QueueDraw(func() {
				row, ok := d.getRowByAddress(ev.Address)
				if ok {
					d.table.RemoveRow(row)
					d.player.closeForDevice(ev.Address)
				}
			})
		}
	}
}
