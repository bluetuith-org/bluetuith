package views

import (
	"errors"
	"sort"

	"go.uber.org/atomic"

	"github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	"github.com/darkhz/bluetuith/ui/theme"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
)

// audioProfilesView holds the audio profiles viewer.
type audioProfilesView struct {
	isSupported atomic.Bool

	*Views
}

// audioDevice describes a single audio device.
type audioDevice struct {
	address bluetooth.MacAddress
	profile bluetooth.AudioProfile
}

// Initialize initializes the audio profiles viewer.
func (a *audioProfilesView) Initialize() error {
	a.isSupported.Store(true)

	return nil
}

// SetRootView sets the root view for the audio profiles viewer.
func (a *audioProfilesView) SetRootView(v *Views) {
	a.Views = v
}

// audioProfiles shows a popup to select the audio profile.
func (a *audioProfilesView) audioProfiles() {
	if !a.isSupported.Load() {
		a.status.ErrorMessage(errors.New("this operation is not supported"))
		return
	}

	device := a.device.getSelection(false)
	if device.Address.IsNil() {
		return
	}

	profiles, err := a.app.Session().MediaPlayer(device.Address).AudioProfiles()
	if err != nil {
		a.status.ErrorMessage(err)
		return
	}
	sort.Slice(profiles, func(i, _ int) bool {
		return profiles[i].Name == "off"
	})

	a.menu.drawContextMenu(
		menuDeviceName.String(),
		func(profileMenu *tview.Table) {
			row, _ := profileMenu.GetSelection()

			a.setProfile(profileMenu, row, 0)

		}, nil,
		func(profileMenu *tview.Table) (int, int) {
			var width, index int

			profileMenu.SetSelectorWrap(true)

			for row, profile := range profiles {
				if profile.Active {
					index = row
				}

				if len(profile.Description) > width {
					width = len(profile.Description)
				}

				profileMenu.SetCellSimple(row, 0, "")

				profileMenu.SetCell(row, 1, tview.NewTableCell(profile.Description).
					SetExpansion(1).
					SetReference(audioDevice{device.Address, profile}).
					SetAlign(tview.AlignLeft).
					SetOnClickedFunc(a.setProfile).
					SetTextColor(theme.GetColor(theme.ThemeText)).
					SetSelectedStyle(tcell.Style{}.
						Foreground(theme.GetColor(theme.ThemeText)).
						Background(theme.BackgroundColor(theme.ThemeText)),
					),
				)

			}

			a.markActiveProfile(profileMenu, index)

			return width - 16, index
		},
	)
}

// setProfile sets the selected audio profile.
func (a *audioProfilesView) setProfile(profileMenu *tview.Table, row, _ int) {
	cell := profileMenu.GetCell(row, 1)
	if cell == nil {
		return
	}

	device, ok := cell.GetReference().(audioDevice)
	if !ok {
		return
	}

	if err := a.app.Session().MediaPlayer(device.address).SetAudioProfile(device.profile); err != nil {
		a.status.ErrorMessage(err)
		return
	}

	a.markActiveProfile(profileMenu, row)
}

// markActiveProfile marks the active profile in the profiles list.
func (a *audioProfilesView) markActiveProfile(profileMenu *tview.Table, index ...int) {
	for i := range profileMenu.GetRowCount() {
		var activeIndicator string

		if i == index[0] {
			activeIndicator = string('\u2022')
		} else {
			activeIndicator = ""
		}

		profileMenu.SetCell(i, 0, tview.NewTableCell(activeIndicator).
			SetSelectable(false).
			SetTextColor(theme.GetColor(theme.ThemeText)).
			SetSelectedStyle(tcell.Style{}.
				Foreground(theme.GetColor(theme.ThemeText)).
				Background(theme.BackgroundColor(theme.ThemeText)),
			),
		)
	}
}
