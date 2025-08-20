package views

import (
	"errors"
	"fmt"
	"strings"

	"go.uber.org/atomic"

	"github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	"github.com/darkhz/bluetuith/ui/theme"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
)

// networkView holds the network selector view.
type networkView struct {
	isSupported atomic.Bool

	*Views
}

// Initialize initializes the network selector view.
func (n *networkView) Initialize() error {
	n.isSupported.Store(true)

	return nil
}

// SetRootView sets the root view for the network selector view.
func (n *networkView) SetRootView(v *Views) {
	n.Views = v
}

// networkSelect shows a popup to select the network type.
func (n *networkView) networkSelect() {
	if !n.isSupported.Load() {
		n.status.ErrorMessage(errors.New("this operation is not supported"))
		return
	}

	type nwTypeDesc struct {
		connType bluetooth.NetworkType
		desc     string
	}

	var connTypes []nwTypeDesc

	device := n.device.getSelection(false)
	if device.Address.IsNil() {
		return
	}

	if device.HaveService(bluetooth.PanuServiceClass) {
		connTypes = append(connTypes, nwTypeDesc{
			bluetooth.NetworkPanu,
			"Personal Area Network",
		})
	}
	if device.HaveService(bluetooth.DialupNetServiceClass) {
		connTypes = append(connTypes, nwTypeDesc{
			bluetooth.NetworkDun,
			"Dialup Network",
		})
	}

	if connTypes == nil {
		n.status.InfoMessage("No network options exist for "+device.Name, false)
		return
	}

	n.menu.drawContextMenu(
		menuDeviceName.String(),
		func(networkMenu *tview.Table) {
			row, _ := networkMenu.GetSelection()

			cell := networkMenu.GetCell(row, 0)
			if cell == nil {
				return
			}

			connType, ok := cell.GetReference().(bluetooth.NetworkType)
			if !ok {
				return
			}

			go n.networkConnect(device, connType)

		}, nil,
		func(networkMenu *tview.Table) (int, int) {
			var width int

			for row, nw := range connTypes {
				ctype := nw.connType
				description := nw.desc

				if len(description) > width {
					width = len(description)
				}

				networkMenu.SetCell(row, 0, tview.NewTableCell(description).
					SetExpansion(1).
					SetReference(ctype).
					SetAlign(tview.AlignLeft).
					SetTextColor(theme.GetColor(theme.ThemeText)).
					SetSelectedStyle(tcell.Style{}.
						Foreground(theme.GetColor(theme.ThemeText)).
						Background(theme.BackgroundColor(theme.ThemeText)),
					),
				)
				networkMenu.SetCell(row, 1, tview.NewTableCell("("+strings.ToUpper(ctype.String())+")").
					SetAlign(tview.AlignRight).
					SetTextColor(theme.GetColor(theme.ThemeText)).
					SetSelectedStyle(tcell.Style{}.
						Foreground(theme.GetColor(theme.ThemeText)).
						Background(theme.BackgroundColor(theme.ThemeText)),
					),
				)
			}

			return width, 0
		},
	)
}

// networkConnect connects to the network with the selected network type.
func (n *networkView) networkConnect(device bluetooth.DeviceData, connType bluetooth.NetworkType) {
	info := fmt.Sprintf("%s (%s)",
		device.Name, strings.ToUpper(connType.String()),
	)

	n.op.startOperation(
		func() {
			n.status.InfoMessage("Connecting to "+info, true)
			err := n.app.Session().Network(device.Address).Connect(device.Name, connType)
			if err != nil {
				n.status.ErrorMessage(err)
				return
			}
			n.status.InfoMessage("Connected to "+info, false)
		},
		func() {
			err := n.app.Session().Network(device.Address).Disconnect()
			if err != nil {
				n.status.ErrorMessage(err)
				return
			}
			n.status.InfoMessage("Cancelled connection to "+info, false)
		},
	)
}
