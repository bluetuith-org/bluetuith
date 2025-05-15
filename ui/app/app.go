package app

import (
	"syscall"
	"time"

	"github.com/bluetuith-org/bluetooth-classic/api/appfeatures"
	"github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"

	"github.com/darkhz/bluetuith/ui/app/views"
	"github.com/darkhz/bluetuith/ui/config"
)

// Application holds an application with its views.
type Application struct {
	view *views.Views
}

// NewApplication returns a new application.
func NewApplication() *Application {
	return &Application{
		view: views.NewViews(),
	}
}

// Start starts the application.
func (a *Application) Start(session bluetooth.Session, featureSet *appfeatures.FeatureSet, cfg *config.Config) error {
	binder := &appBinder{
		session:     session,
		draws:       make(chan struct{}, 1),
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

	go binder.monitorQueuedDraws()

	return binder.SetRoot(appview.Layout, true).SetFocus(appview.InitialFocus).EnableMouse(true).Run()
}

// Authorizer returns the session's authorizer.
func (a *Application) Authorizer() bluetooth.SessionAuthorizer {
	return a.view.Authorizer()
}

// appBinder holds the bluetooth session and the application.
type appBinder struct {
	session       bluetooth.Session
	featureSet    *appfeatures.FeatureSet
	draws         chan struct{}
	shouldSuspend bool

	*tview.Application
}

// Session returns the current session.
func (a *appBinder) Session() bluetooth.Session {
	return a.session
}

// Features returns the current features of the session.
func (a *appBinder) Features() *appfeatures.FeatureSet {
	return a.featureSet
}

// InstantDraw instantly draws to the screen.
func (a *appBinder) InstantDraw(drawFunc func()) {
	a.QueueUpdateDraw(drawFunc)
}

// QueueDraw only queues the drawing.
func (a *appBinder) QueueDraw(drawFunc func()) {
	a.QueueUpdate(drawFunc)

	select {
	case a.draws <- struct{}{}:
	default:
	}
}

// Refresh refreshes the screen.
func (a *appBinder) Refresh() {
	a.Draw()
}

// GetFocused gets the currently focused primitive.
func (a *appBinder) GetFocused() tview.Primitive {
	return a.GetFocus()
}

// FocusPrimitive sets the focus on the provided primitive.
func (a *appBinder) FocusPrimitive(primitive tview.Primitive) {
	a.SetFocus(primitive)
}

// StartSuspend starts the application's suspend.
// [appbinder.Suspend] is called within the application's drawing handler
// once this function is called.
func (a *appBinder) StartSuspend() {
	a.shouldSuspend = true
}

// Suspend suspends the application.
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

// Close stops the application.
func (a *appBinder) Close() {
	a.Stop()
}

// monitorQueuedDraws monitors for any queued primitive draws and refrehses the screen.
func (a *appBinder) monitorQueuedDraws() {
	t := time.NewTicker(1 * time.Second)
	defer t.Stop()

	var queued bool

	for {
		select {
		case <-t.C:
			if queued {
				go a.Refresh()
				queued = false
				t.Reset(1 * time.Second)
			}

		case <-a.draws:
			queued = true
			t.Reset(50 * time.Millisecond)
		}
	}
}
