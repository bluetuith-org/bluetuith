package views

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	"github.com/darkhz/bluetuith/ui/theme"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
	"go.uber.org/atomic"
)

type adapterView struct {
	topStatus      *tview.TextView
	currentAdapter atomic.Pointer[bluetooth.AdapterData]

	*Views
}

func (a *adapterView) Initialize() error {
	a.topStatus = tview.NewTextView()
	a.topStatus.SetRegions(true)
	a.topStatus.SetDynamicColors(true)
	a.topStatus.SetTextAlign(tview.AlignRight)
	a.topStatus.SetBackgroundColor(theme.GetColor(theme.ThemeMenuBar))

	a.setAdapter(a.cfg.Values.SelectedAdapter)
	a.updateTopStatus()
	a.setStates()

	go a.event()
	if err := a.adapter.currentSession().SetPairableState(false); err == nil {
		a.adapter.currentSession().SetPairableState(true)
	}

	return nil
}

func (a *adapterView) SetRootView(v *Views) {
	a.Views = v
}

func (a *adapterView) refreshHeader() {
	props, err := a.currentSession().Properties()
	if err != nil {
		a.status.ErrorMessage(err)
		return
	}

	headerText := fmt.Sprintf("[\"%s\"]%s (%s)[\"\"]",
		menuAdapterChangeName.String(),
		props.Name,
		props.UniqueName,
	)
	a.menu.setHeader(theme.ColorWrap(theme.ThemeAdapter, headerText, "::bu"))
}

func (a *adapterView) getAdapter() *bluetooth.AdapterData {
	return a.currentAdapter.Load()
}

func (a *adapterView) currentSession() bluetooth.Adapter {
	return a.app.Session().Adapter(a.currentAdapter.Load().Address)
}

func (a *adapterView) setAdapter(adapter *bluetooth.AdapterData) {
	a.currentAdapter.Swap(adapter)
	a.refreshHeader()
}

// change launches a popup with a list of adapters.
// Changing the selection will change the currently selected adapter.
func (a *adapterView) change() {
	a.menu.drawContextMenu(
		menuAdapterName.String(), nil,
		func(adapterMenu *tview.Table, row, col int) {
			cell := adapterMenu.GetCell(row, 0)
			if cell == nil {
				return
			}

			adapter, ok := cell.GetReference().(bluetooth.AdapterData)
			if !ok {
				return
			}

			if a.actions.scan("no") {
				a.status.InfoMessage("Scanning stopped on "+a.getAdapter().UniqueName, false)
			}

			a.setAdapter(&adapter)
			a.updateTopStatus()

			a.op.cancelOperation(false)
			a.device.list()
		},
		func(adapterMenu *tview.Table) (int, int) {
			var width, index int

			adapters, err := a.app.Session().Adapters()
			if err != nil {
				a.status.ErrorMessage(err)
				return -1, -1
			}

			sort.Slice(adapters, func(i, j int) bool {
				return adapters[i].UniqueName < adapters[j].UniqueName
			})

			for row, adapter := range adapters {
				if len(adapter.Name) > width {
					width = len(adapter.Name)
				}

				if adapter.UniqueName == a.getAdapter().UniqueName {
					index = row
				}

				adapterMenu.SetCell(row, 0, tview.NewTableCell(adapter.Name).
					SetExpansion(1).
					SetReference(adapter).
					SetAlign(tview.AlignLeft).
					SetTextColor(theme.GetColor(theme.ThemeAdapter)).
					SetSelectedStyle(tcell.Style{}.
						Foreground(theme.GetColor(theme.ThemeAdapter)).
						Background(theme.BackgroundColor(theme.ThemeAdapter)),
					),
				)
				adapterMenu.SetCell(row, 1, tview.NewTableCell("("+adapter.UniqueName+")").
					SetAlign(tview.AlignRight).
					SetTextColor(theme.GetColor(theme.ThemeAdapter)).
					SetSelectedStyle(tcell.Style{}.
						Foreground(theme.GetColor(theme.ThemeAdapter)).
						Background(theme.BackgroundColor(theme.ThemeAdapter)),
					),
				)
			}

			return width, index
		})
}

// updateTopStatus updates the adapter status display.
func (a *adapterView) updateTopStatus() {
	var state string

	props, err := a.currentSession().Properties()
	if err != nil {
		a.status.ErrorMessage(err)
		return
	}

	properties := map[string]bool{
		"Powered":      props.Powered,
		"Discovering":  props.Discovering,
		"Discoverable": props.Discoverable,
		"Pairable":     props.Pairable,
	}

	for _, status := range []struct {
		Title   string
		Enabled bool
		Color   theme.ThemeContext
	}{
		{
			Title:   "Powered",
			Enabled: properties["Powered"],
			Color:   theme.ThemeAdapterPowered,
		},
		{
			Title:   "Scanning",
			Enabled: properties["Discovering"],
			Color:   theme.ThemeAdapterScanning,
		},
		{
			Title:   "Discoverable",
			Enabled: properties["Discoverable"],
			Color:   theme.ThemeAdapterDiscoverable,
		},
		{
			Title:   "Pairable",
			Enabled: properties["Pairable"],
			Color:   theme.ThemeAdapterPairable,
		},
	} {
		if !status.Enabled {
			if status.Title == "Powered" {
				status.Title = "Not " + status.Title
				status.Color = "AdapterNotPowered"
			} else {
				continue
			}
		}

		textColor := theme.ColorName(theme.BackgroundColor(status.Color))
		bgColor := theme.ThemeConfig[status.Color]

		region := strings.ToLower(status.Title)
		state += fmt.Sprintf("[\"%s\"][%s:%s:b] %s [-:-:-][\"\"] ", region, textColor, bgColor, status.Title)
	}

	a.topStatus.SetText(state)
}

// setStates sets the adapter states which were parsed from
// the "adapter-states" command-line option.
func (a *adapterView) setStates() {
	var lock sync.Mutex

	properties := a.cfg.Values.AdapterStatesMap
	if len(properties) == 0 {
		return
	}

	seq, ok := properties["sequence"]
	if !ok {
		a.status.InfoMessage("Cannot get adapter states", false)
		return
	}

	sequence := strings.SplitSeq(seq, ",")
	for property := range sequence {
		var handler func(set ...string) bool

		state, ok := properties[property]
		if !ok {
			a.status.InfoMessage("Cannot set adapter "+property+" state", false)
			return
		}

		switch property {
		case "powered":
			handler = a.actions.power

		case "scan":
			handler = a.actions.scan

		case "discoverable":
			handler = a.actions.discoverable

		case "pairable":
			handler = a.actions.pairable

		default:
			continue
		}

		go func() {
			lock.Lock()
			defer lock.Unlock()

			handler(state)
		}()
	}
}

// event handles adapter-specific events.
func (a *adapterView) event() {
	adapterSub := bluetooth.AdapterEvent().Subscribe()
	if !adapterSub.Subscribable {
		a.status.ErrorMessage(errors.New("cannot subscribe to adapter events"))
		return
	}

	for adapterEvent := range adapterSub.C {
		switch adapterEvent.Action {
		case bluetooth.EventActionUpdated:
			if adapterEvent.Data.Address == a.currentAdapter.Load().Address {
				a.app.InstantDraw(func() {
					a.updateTopStatus()
				})
			}

		case bluetooth.EventActionAdded:
			if adapterEvent.Data.Address == a.currentAdapter.Load().Address {
				a.device.list()
			}

			fallthrough

		case bluetooth.EventActionRemoved:
			a.app.InstantDraw(func() {
				if modal, ok := a.modals.ModalExists(menuAdapterName.String()); ok {
					modal.Exit(false)
					a.change()
				}
			})
		}
	}
}
