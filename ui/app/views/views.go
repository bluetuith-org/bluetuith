package views

import (
	"github.com/bluetuith-org/bluetooth-classic/api/appfeatures"
	"github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	"github.com/darkhz/bluetuith/ui/config"
	"github.com/darkhz/bluetuith/ui/keybindings"
	"github.com/darkhz/bluetuith/ui/theme"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
)

type ViewApp struct {
	// layout holds the layout of the application.
	Layout         *tview.Flex
	InitialFocus   *tview.Flex
	MouseFunc      func(event *tcell.EventMouse, action tview.MouseAction) (*tcell.EventMouse, tview.MouseAction)
	BeforeDrawFunc func(t tcell.Screen) bool
	InputCapture   func(event *tcell.EventKey) *tcell.EventKey
}

type Views struct {
	// pages holds and renders the different views, along with
	// any menu popups that will be added.
	pages  *viewPages
	layout *tview.Flex

	menu          *menuBarView
	help          *helpView
	status        *statusBarView
	modals        *modalViews
	device        *deviceView
	adapter       *adapterView
	filepicker    *filePickerView
	progress      *progressView
	player        *mediaPlayer
	audioProfiles *audioProfilesView
	network       *networkView

	actions *viewActions
	op      *viewOperation
	kb      *keybindings.Keybindings
	cfg     *config.Config

	app  AppBinder
	auth *authorizer
}

type AppBinder interface {
	Session() bluetooth.Session
	Features() *appfeatures.FeatureSet

	InstantDraw(drawFunc func())
	Refresh()
	FocusPrimitive(primitive tview.Primitive)

	Suspend(t tcell.Screen)
	StartSuspend()
	GetFocused() tview.Primitive
	Close()
}

type viewName string

func (v viewName) String() string {
	return string(v)
}

type viewManager interface {
	Initialize() error
	SetRootView(v *Views)
}

func NewViews() *Views {
	v := &Views{
		pages:         &viewPages{},
		menu:          &menuBarView{},
		help:          &helpView{},
		status:        &statusBarView{},
		modals:        &modalViews{},
		device:        &deviceView{},
		adapter:       &adapterView{},
		filepicker:    &filePickerView{},
		progress:      &progressView{},
		player:        &mediaPlayer{},
		audioProfiles: &audioProfilesView{},
		network:       &networkView{},
		actions:       &viewActions{},
		op:            &viewOperation{},
		kb:            &keybindings.Keybindings{},
	}

	v.auth = newAuthorizer(v)

	return v
}

func (v *Views) Initialize(binder AppBinder, cfg *config.Config) (*ViewApp, error) {
	v.app = binder
	v.cfg = cfg
	v.kb = v.cfg.Values.Kb

	v.actions = newViewActions(v)
	v.op = newViewOperation(v)

	v.pages = newViewPages()
	v.layout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(v.pages, 0, 10, true)

	initialize := []viewManager{
		v.menu,
		v.status,
		v.help,
		v.modals,
		v.adapter,
		v.device,
	}
	if v.app.Features().HasAny(appfeatures.FeatureSendFile, appfeatures.FeatureReceiveFile) {
		initialize = append(initialize, v.filepicker, v.progress)
	}
	if v.app.Features().Has(appfeatures.FeatureMediaPlayer) {
		initialize = append(initialize, v.player, v.audioProfiles)
	}
	if v.app.Features().Has(appfeatures.FeatureNetwork) {
		initialize = append(initialize, v.network)
	}
	for _, i := range initialize {
		i.SetRootView(v)
		if err := i.Initialize(); err != nil {
			return nil, err
		}
	}

	v.auth.setInitialized()

	return &ViewApp{
		Layout:       v.layout,
		InitialFocus: v.arrangeViews(),
		MouseFunc: func(event *tcell.EventMouse, action tview.MouseAction) (*tcell.EventMouse, tview.MouseAction) {
			return v.modals.modalMouseHandler(event, action)
		},
		BeforeDrawFunc: func(t tcell.Screen) bool {
			v.modals.ResizeModal()
			v.app.Suspend(t)

			return false
		},
		InputCapture: func(event *tcell.EventKey) *tcell.EventKey {
			operation := v.kb.Key(event)

			if e, ok := v.kb.IsNavigation(operation, event); ok {
				focused := v.app.GetFocused()
				if focused != nil && focused.InputHandler() != nil {
					focused.InputHandler()(e, nil)
					return nil
				}
			}

			switch operation {
			case keybindings.KeySuspend:
				v.app.StartSuspend()

			case keybindings.KeyCancel:
				v.op.cancelOperation(true)
			}

			return tcell.NewEventKey(event.Key(), event.Rune(), event.Modifiers())
		},
	}, nil
}

func (v *Views) Authorizer() bluetooth.SessionAuthorizer {
	return v.auth
}

func (v *Views) arrangeViews() *tview.Flex {
	box := tview.NewBox().
		SetBackgroundColor(theme.GetColor(theme.ThemeMenuBar))

	menuArea := tview.NewFlex().
		AddItem(v.menu.bar, 0, 1, false).
		AddItem(box, 1, 0, false).
		AddItem(v.adapter.topStatus, 0, 1, false)
	menuArea.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(menuArea, 1, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(v.device.table, 0, 10, true)
	flex.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	v.pages.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))
	v.pages.SetChangedFunc(func() {
		page, _ := v.pages.GetFrontPage()

		contexts := map[string]keybindings.Context{
			devicePage.String():     keybindings.ContextDevice,
			filePickerPage.String(): keybindings.ContextFiles,
			progressPage.String():   keybindings.ContextProgress,
		}

		switch page {
		case devicePage.String(), filePickerPage.String(), progressPage.String():
			v.pages.currentPage(page)
			v.pages.currentContext(contexts[page])

		default:
			v.pages.currentContext(keybindings.ContextApp)
		}

		v.help.showStatusHelp(page)
	})

	v.layout.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	v.pages.AddAndSwitchToPage(devicePage.String(), flex, true)
	v.status.InfoMessage("bluetuith is ready.", false)

	return flex
}

// getSelectionXY gets the coordinates of the current table selection.
func getSelectionXY(table *tview.Table) (int, int) {
	row, _ := table.GetSelection()

	cell := table.GetCell(row, 0)
	x, y, _ := cell.GetLastPosition()

	return x, y
}

// ignoreDefaultEvent ignores the default keyevents in the provided event.
func ignoreDefaultEvent(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyCtrlF, tcell.KeyCtrlB:
		return nil
	}

	switch event.Rune() {
	case 'g', 'G', 'j', 'k', 'h', 'l':
		return nil
	}

	return event
}

// horizontalLine returns a box with a thick horizontal line.
func horizontalLine() *tview.Box {
	return tview.NewBox().
		SetBackgroundColor(tcell.ColorDefault).
		SetDrawFunc(func(
			screen tcell.Screen,
			x, y, width, height int) (int, int, int, int) {
			centerY := y + height/2
			for cx := x; cx < x+width; cx++ {
				screen.SetContent(
					cx,
					centerY,
					tview.BoxDrawingsLightHorizontal,
					nil,
					tcell.StyleDefault.Foreground(tcell.ColorWhite),
				)
			}

			return x + 1,
				centerY + 1,
				width - 2,
				height - (centerY + 1 - y)
		})
}
