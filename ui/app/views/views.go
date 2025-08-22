package views

import (
	"github.com/bluetuith-org/bluetooth-classic/api/appfeatures"
	"github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"

	"github.com/darkhz/bluetuith/ui/config"
	"github.com/darkhz/bluetuith/ui/keybindings"
	"github.com/darkhz/bluetuith/ui/theme"
)

// AppData holds all the necessary layout and event handling data for the root application to initialize.
// This is passed to the application once all the views are initialized using [Views.Initialize].
type AppData struct {
	// layout holds the layout of the application.
	Layout         *tview.Flex
	InitialFocus   *tview.Flex
	MouseFunc      func(event *tcell.EventMouse, action tview.MouseAction) (*tcell.EventMouse, tview.MouseAction)
	BeforeDrawFunc func(t tcell.Screen) bool
	InputCapture   func(event *tcell.EventKey) *tcell.EventKey
}

// AppBinder binds all the root application's functions to the views manager ([Views]).
type AppBinder interface {
	Session() bluetooth.Session
	Features() *appfeatures.FeatureSet

	QueueDraw(drawFunc func())
	InstantDraw(drawFunc func())
	Refresh()
	FocusPrimitive(primitive tview.Primitive)

	Suspend(t tcell.Screen)
	StartSuspend()
	GetFocused() tview.Primitive
	Close()
}

// viewInitializer represents an initializer for a view.
// All views must implement this interface.
type viewInitializer interface {
	Initialize() error
	SetRootView(v *Views)
}

// Views holds all the views as well as different managers for
// the view layouts, operations and actions.
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

// NewViews returns a new Views instance.
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

// Initialize initializes all the views.
func (v *Views) Initialize(binder AppBinder, cfg *config.Config) (*AppData, error) {
	v.app = binder
	v.cfg = cfg
	v.kb = v.cfg.Values.Kb

	v.actions = newViewActions(v)
	v.op = newViewOperation(v)

	v.pages = newViewPages()
	v.layout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(v.pages, 0, 10, true)

	initializers := []viewInitializer{
		v.menu,
		v.status,
		v.help,
		v.modals,
		v.adapter,
		v.device,
		v.filepicker, v.progress,
		v.player, v.audioProfiles,
		v.network,
	}

	dontInit := map[viewInitializer]struct{}{}
	if !v.app.Features().HasAny(appfeatures.FeatureSendFile, appfeatures.FeatureReceiveFile) {
		dontInit[v.filepicker] = struct{}{}
		dontInit[v.progress] = struct{}{}
	}
	if !v.app.Features().Has(appfeatures.FeatureMediaPlayer) {
		dontInit[v.player] = struct{}{}
		dontInit[v.audioProfiles] = struct{}{}
	}
	if !v.app.Features().Has(appfeatures.FeatureNetwork) {
		dontInit[v.network] = struct{}{}
	}

	for _, i := range initializers {
		i.SetRootView(v)
		if _, ok := dontInit[i]; ok {
			continue
		}

		if err := i.Initialize(); err != nil {
			return nil, err
		}
	}

	v.menu.setHeader("")

	v.kb.Initialize()
	v.auth.setInitialized()

	return &AppData{
		Layout:       v.layout,
		InitialFocus: v.arrangeViews(),
		MouseFunc: func(event *tcell.EventMouse, action tview.MouseAction) (*tcell.EventMouse, tview.MouseAction) {
			return v.modals.modalMouseHandler(event, action)
		},
		BeforeDrawFunc: func(t tcell.Screen) bool {
			v.modals.resizeModal()
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

// Authorizer returns an authorization manager.
func (v *Views) Authorizer() bluetooth.SessionAuthorizer {
	return v.auth
}

// arrangeViews arranges all the views and their layouts.
func (v *Views) arrangeViews() *tview.Flex {
	box := tview.NewBox().
		SetBackgroundColor(theme.GetColor(theme.ThemeMenuBar))

	menuArea := tview.NewFlex().
		AddItem(v.adapter.topAdapterName, 0, 1, false).
		AddItem(v.menu.bar, len(v.menu.bar.GetText(true)), 1, false).
		AddItem(box, 1, 1, false).
		AddItem(v.adapter.topStatus, 0, 4, false)
	menuArea.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))
	menuArea.SetDrawFunc(func(_ tcell.Screen, x, y, width, height int) (int, int, int, int) {
		w := len(v.adapter.topAdapterName.GetText(true))
		resize := min(w, width/8)

		menuArea.ResizeItem(v.adapter.topAdapterName, resize, 0)

		return x, y, width, height
	})

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

// viewName represents the name of a particular view.
type viewName string

// String returns the string representation of the view's name.
func (v viewName) String() string {
	return string(v)
}

// getSelectionXY gets the coordinates of the current table selection.
func getSelectionXY(table *tview.Table) (x int, y int) {
	row, _ := table.GetSelection()

	cell := table.GetCell(row, 0)
	x, y, _ = cell.GetLastPosition()

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
