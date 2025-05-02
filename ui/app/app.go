package app

import (
	"syscall"

	"github.com/bluetuith-org/bluetooth-classic/api/appfeatures"
	"github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	"github.com/darkhz/bluetuith/ui/app/views"
	"github.com/darkhz/bluetuith/ui/config"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
)

type Application struct {
	view   *views.Views
}

// authorizer
func NewApplication() *Application {
	return &Application{
		view: views.NewViews(),
	}
}

func (a *Application) Start(session bluetooth.Session, featureSet *appfeatures.FeatureSet, cfg *config.Config) error {
	binder := &appBinder{
		session:     session,
		featureSet:  featureSet,
		Application: tview.NewApplication(),
	}

	appview, err := a.view.Initialize(binder, cfg)
	if err != nil {
		return err
	}

	binder.SetInputCapture(appview.InputCapture)
	binder.SetMouseCapture(appview.MouseFunc)
	binder.SetBeforeDrawFunc(appview.BeforeDrawFunc)

	return binder.SetRoot(appview.Layout, true).SetFocus(appview.InitialFocus).EnableMouse(true).Run()
}

func (a *Application) Authorizer() bluetooth.SessionAuthorizer {
	return a.view.Authorizer()
}

type appBinder struct {
	session       bluetooth.Session
	featureSet    *appfeatures.FeatureSet
	shouldSuspend bool

	*tview.Application
}

func (a *appBinder) Session() bluetooth.Session {
	return a.session
}

func (a *appBinder) Features() *appfeatures.FeatureSet {
	return a.featureSet
}

func (a *appBinder) InstantDraw(drawFunc func()) {
	a.QueueUpdateDraw(drawFunc)
}

func (a *appBinder) Refresh() {
	a.Draw()
}

func (a *appBinder) GetFocused() tview.Primitive {
	return a.GetFocus()
}

func (a *appBinder) FocusPrimitive(primitive tview.Primitive) {
	a.SetFocus(primitive)
}

func (a *appBinder) StartSuspend() {
	a.shouldSuspend = true
}

// Suspend suspends the alication.
func (a *appBinder) Suspend(t tcell.Screen) {
	if !a.shouldSuspend {
		return
	}

	a.shouldSuspend = false

	if err := t.Suspend(); err != nil {
		return
	}
	if err := syscall.Kill(syscall.Getpid(), syscall.SIGSTOP); err != nil {
		return
	}
	if err := t.Resume(); err != nil {
		return
	}
}

func (a *appBinder) Close() {
	a.Stop()
}
