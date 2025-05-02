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

type deviceView struct {
	table *tview.Table

	*Views
}

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

		d.player.keyEvents(event, false)

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

func (d *deviceView) SetRootView(v *Views) {
	d.Views = v
}

// list lists the devices belonging to the selected adapter.
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

// getInfo shows information about a device.
func (d *deviceView) getInfo() {
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

	infoModal := d.modals.NewModal("info", "Device Information", nil, 40, 100)
	infoModal.Table.SetSelectionChangedFunc(func(row, col int) {
		_, _, _, height := infoModal.Table.GetRect()
		infoModal.Table.SetOffset(row-((height-1)/2), 0)
	})

	for i, prop := range props {
		propName := prop[0]
		propValue := prop[1]

		switch propName {
		case "Class":
			propValue += " (" + device.Type + ")"
		}

		infoModal.Table.SetCell(i, 0, tview.NewTableCell("[::b]"+propName+":").
			SetExpansion(1).
			SetAlign(tview.AlignLeft).
			SetTextColor(theme.GetColor(theme.ThemeText)).
			SetSelectedStyle(tcell.Style{}.
				Bold(true).
				Underline(true),
			),
		)

		infoModal.Table.SetCell(i, 1, tview.NewTableCell(propValue).
			SetExpansion(1).
			SetAlign(tview.AlignLeft).
			SetTextColor(theme.GetColor(theme.ThemeText)),
		)
	}

	rows := infoModal.Table.GetRowCount() - 1
	for i, serviceUUID := range device.UUIDs {
		serviceType := bluetooth.ServiceType(serviceUUID)
		serviceUUID = "(" + serviceUUID + ")"

		infoModal.Table.SetCell(rows+i, 1, tview.NewTableCell(serviceType).
			SetExpansion(1).
			SetAlign(tview.AlignLeft).
			SetTextColor(theme.GetColor(theme.ThemeText)),
		)

		infoModal.Table.SetCell(rows+i, 2, tview.NewTableCell(serviceUUID).
			SetExpansion(0).
			SetTextColor(theme.GetColor(theme.ThemeText)),
		)
	}

	infoModal.Height = min(infoModal.Table.GetRowCount()+4, 60)

	infoModal.Show()
}

// getSelection retrieves device information from
// the current selection in the d.table.
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
		d.app.InstantDraw(func() {
			getdevice()
		})

		return device
	}

	getdevice()

	return device
}

// checkd.table iterates through the d.table and checks
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

// setInfo writes device information into the
// specified row of the d.table.
func (d *deviceView) setInfo(row int, device bluetooth.DeviceData) {
	var props string

	data := []string{
		theme.ColorWrap(theme.ThemeDeviceType, device.Type),
	}
	name := device.Name
	if name == "" {
		name = device.Address.String()
	}
	if device.Alias != device.Name {
		data = append(
			[]string{theme.ColorWrap(theme.ThemeDeviceAlias, device.Alias)},
			data...,
		)
	}
	name += " (" + strings.Join(data, ", ") + ")"

	nameColor := theme.ThemeDevice
	propColor := theme.ThemeDeviceProperty

	if device.Connected {
		props += "Connected"

		nameColor = theme.ThemeDeviceConnected
		propColor = theme.ThemeDevicePropertyConnected

		if device.RSSI < 0 {
			rssi := strconv.FormatInt(int64(device.RSSI), 10)
			props += "[" + rssi + "[]"
		}

		if device.Percentage > 0 {
			props += ", Battery " + strconv.Itoa(device.Percentage) + "%"
		}

		props += ", "
	}

	if device.Trusted {
		props += "Trusted, "
	}
	if device.Blocked {
		props += "Blocked, "
	}
	if device.Bonded && device.Paired {
		props += "Bonded, "
	} else if !device.Bonded && device.Paired {
		props += "Paired, "
	}

	if props != "" {
		props = "(" + strings.TrimRight(props, ", ") + ")"
	} else {
		props = "[New Device[]"
		nameColor = theme.ThemeDeviceDiscovered
		propColor = theme.ThemeDevicePropertyDiscovered
	}

	d.table.SetCell(
		row, 0, tview.NewTableCell(name).
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
	d.table.SetCell(
		row, 1, tview.NewTableCell(props).
			SetExpansion(1).
			SetAlign(tview.AlignRight).
			SetTextColor(theme.GetColor(propColor)).
			SetSelectedStyle(tcell.Style{}.
				Bold(true),
			),
	)
}

// event handles device-specific events.
// TODO: Partial updates.
func (d *deviceView) event() {
	deviceSub := bluetooth.DeviceEvent().Subscribe()
	if !deviceSub.Subscribable {
		d.status.ErrorMessage(errors.New("cannot subscribe to device events"))
		return
	}

	for deviceEvent := range deviceSub.C {
		switch deviceEvent.Action {
		case bluetooth.EventActionUpdated:
			d.app.InstantDraw(func() {
				row, ok := d.getRowByAddress(deviceEvent.Data.Address)
				if ok {
					device, err := d.app.Session().Device(deviceEvent.Data.Address).Properties()
					if err == nil {
						d.setInfo(row, device)
					}
				}
			})

		case bluetooth.EventActionAdded:
			if deviceEvent.Data.AssociatedAdapter != d.adapter.getAdapter().Address {
				continue
			}

			d.app.InstantDraw(func() {
				deviceRow := d.table.GetRowCount()

				row, ok := d.getRowByAddress(deviceEvent.Data.Address)
				if ok {
					deviceRow = row
				}

				device, err := d.app.Session().Device(deviceEvent.Data.Address).Properties()
				if err == nil {
					d.setInfo(deviceRow, device)
				}
			})

		case bluetooth.EventActionRemoved:
			d.app.InstantDraw(func() {
				row, ok := d.getRowByAddress(deviceEvent.Data.Address)
				if ok {
					d.table.RemoveRow(row)
				}
			})
		}
	}
}
