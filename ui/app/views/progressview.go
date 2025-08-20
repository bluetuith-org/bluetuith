package views

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"sync"
	"time"

	"go.uber.org/atomic"

	"github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
	"github.com/puzpuzpuz/xsync/v3"
	"github.com/schollz/progressbar/v3"

	"github.com/darkhz/bluetuith/ui/keybindings"
	"github.com/darkhz/bluetuith/ui/theme"
)

const progressPage viewName = "progressview"

const progressViewButtonRegion = `["resume"][::b][Resume[][""] ["suspend"][::b][Pause[][""] ["cancel"][::b][Cancel[][""]`

// progressView describes a file transfer progress display.
type progressView struct {
	isSupported atomic.Bool

	view, statusProgress *tview.Table
	flex                 *tview.Flex

	total atomic.Uint32

	sessions *xsync.MapOf[bluetooth.MacAddress, *progressViewSession]

	*Views
}

type progressViewSession struct {
	sessionRemoved bool

	transferSession bluetooth.ObexObjectPush
	transfers       map[bluetooth.ObjectPushTransferID]struct{}

	mu sync.Mutex
}

// progressIndicator describes a progress indicator, which will display
// a description and a progress bar.
type progressIndicator struct {
	desc        *tview.TableCell
	progress    *tview.TableCell
	progressBar *progressbar.ProgressBar

	recv, drawn bool
	status      bluetooth.ObjectPushStatus

	deviceAddress bluetooth.MacAddress

	appDrawFunc func(func())
}

func (p *progressView) Initialize() error {
	title := tview.NewTextView()
	title.SetDynamicColors(true)
	title.SetTextAlign(tview.AlignLeft)
	title.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))
	title.SetText(theme.ColorWrap(theme.ThemeText, "Progress View", "::bu"))

	p.view = tview.NewTable()
	p.view.SetSelectable(true, false)
	p.view.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))
	p.view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch p.kb.Key(event, keybindings.ContextProgress) {
		case keybindings.KeyClose:
			if p.status.HasPage(progressPage.String()) && p.total.Load() > 0 {
				p.status.SwitchToPage(progressPage.String())
			}

			p.pages.SwitchToPage(devicePage.String())

		case keybindings.KeyProgressTransferCancel:
			p.cancelTransfer()

		case keybindings.KeyProgressTransferSuspend:
			p.suspendTransfer()

		case keybindings.KeyProgressTransferResume:
			p.resumeTransfer()

		case keybindings.KeyQuit:
			go p.actions.quit()
		}

		return ignoreDefaultEvent(event)
	})

	progressViewButtons := tview.NewTextView()
	progressViewButtons.SetRegions(true)
	progressViewButtons.SetDynamicColors(true)
	progressViewButtons.SetTextAlign(tview.AlignLeft)
	progressViewButtons.SetText(progressViewButtonRegion)
	progressViewButtons.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))
	progressViewButtons.SetHighlightedFunc(func(added, _, _ []string) {
		if added == nil {
			return
		}

		if slices.Contains(progressViewButtons.GetRegionIDs(), added[0]) {
			switch added[0] {
			case "resume":
				p.resumeTransfer()

			case "suspend":
				p.suspendTransfer()

			case "cancel":
				p.cancelTransfer()
			}

			progressViewButtons.Highlight("")
		}
	})

	p.flex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(title, 1, 0, false).
		AddItem(p.view, 0, 10, true).
		AddItem(progressViewButtons, 2, 0, false)

	p.statusProgress = tview.NewTable()
	p.statusProgress.SetSelectable(true, true)
	p.statusProgress.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	p.status.AddPage(progressPage.String(), p.statusProgress, true, false)

	p.isSupported.Store(true)
	p.sessions = xsync.NewMapOf[bluetooth.MacAddress, *progressViewSession]()

	go p.monitorTransfers()

	return nil
}

func (p *progressView) SetRootView(v *Views) {
	p.Views = v
}

// show displays the progress view.
func (p *progressView) show() {
	if !p.isSupported.Load() {
		p.status.ErrorMessage(errors.New("this operation is not supported"))
		return
	}

	if pg, _ := p.status.GetFrontPage(); pg == progressPage.String() {
		p.status.SwitchToPage(statusMessagesPage.String())
	}

	if p.total.Load() == 0 {
		p.status.InfoMessage("No transfers are in progress", false)
		return
	}

	p.pages.AddAndSwitchToPage(progressPage.String(), p.flex, true)
}

// showStatus displays the progress view in the status bar.
func (p *progressView) showStatus() {
	if !p.isSupported.Load() {
		p.status.ErrorMessage(errors.New("this operation is not supported"))
		return
	}

	if pg, _ := p.pages.GetFrontPage(); pg != progressPage.String() {
		p.status.SwitchToPage(progressPage.String())
	}
}

// newIndicator returns a new Progress.
func (p *progressView) newIndicator(props bluetooth.ObjectPushData, recv bool) *progressIndicator {
	var progress progressIndicator
	var progressText string

	if recv {
		progressText = "Receiving"
	} else {
		progressText = "Sending"
	}

	p.total.Add(1)

	name := props.Name
	if name == "" && props.Filename != "" {
		name = filepath.Base(filepath.Clean(props.Filename))
	}
	if props.Filename == "" && props.Name == "" {
		progressText = ""
		name = "Unknown file transfer"
	}

	title := fmt.Sprintf(" [::b]%s %s[-:-:-]", progressText, name)

	progress.recv = recv
	progress.deviceAddress = props.Address
	progress.appDrawFunc = p.app.QueueDraw

	progress.desc = tview.NewTableCell(title).
		SetExpansion(1).
		SetSelectable(false).
		SetAlign(tview.AlignLeft).
		SetTextColor(theme.GetColor(theme.ThemeProgressText))

	progress.progress = tview.NewTableCell("").
		SetExpansion(1).
		SetSelectable(false).
		SetReference(&progress).
		SetAlign(tview.AlignRight).
		SetTextColor(theme.GetColor(theme.ThemeProgressBar))

	progress.progressBar = progressbar.NewOptions64(
		int64(props.Size),
		progressbar.OptionSpinnerType(34),
		progressbar.OptionSetWriter(&progress),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionThrottle(200*time.Millisecond),
	)

	return &progress
}

// drawIndicator draws a progress indicator onto the screen.
func (p *progressView) drawIndicator(progress *progressIndicator, props bluetooth.ObjectPushEventData) {
	if progress.drawn {
		return
	}

	count := p.total.Load()
	p.app.QueueDraw(func() {
		p.showStatus()

		rows := p.view.GetRowCount()

		p.statusProgress.SetCell(0, 0, progress.desc)
		p.statusProgress.SetCell(0, 1, progress.progress)

		p.view.SetCell(rows+1, 0, tview.NewTableCell("#"+strconv.FormatUint(uint64(count), 10)).
			SetReference(props).
			SetAlign(tview.AlignCenter),
		)
		p.view.SetCell(rows+1, 1, progress.desc)
		p.view.SetCell(rows+1, 2, progress.progress)

		progress.drawn = true
	})
}

// monitorTransfers monitors all incoming and outgoing Object Push transfers.
func (p *progressView) monitorTransfers() {
	oppSub, ok := bluetooth.ObjectPushEvents().Subscribe()
	if !ok {
		return
	}
	defer oppSub.Unsubscribe()

	type transferProperty struct {
		indicator *progressIndicator

		bluetooth.ObjectPushData
	}

	sessionMap := map[bluetooth.ObjectPushSessionID]map[bluetooth.ObjectPushTransferID]*transferProperty{}

	getIndicator := func(ev bluetooth.ObjectPushEventData, remove bool) (*transferProperty, bool) {
		var prop *transferProperty

		tmap, ok := sessionMap[ev.SessionID]
		if !ok {
			return nil, false
		}

		prop, ok = tmap[ev.TransferID]
		if remove {
			if ok {
				delete(sessionMap[ev.SessionID], ev.TransferID)
			}

			if len(sessionMap[ev.SessionID]) == 0 {
				delete(sessionMap, ev.SessionID)
			}
		}

		if prop != nil {
			prop.ObjectPushEventData = ev
		}

		return prop, prop != nil
	}

	addIndicator := func(ev bluetooth.ObjectPushData) (*transferProperty, bool) {
		if ev.SessionID == "" || ev.TransferID == "" {
			return nil, false
		}

		indicator := p.newIndicator(ev, ev.Receiving)
		indicator.progressBar.Set64(int64(ev.Transferred))
		property := &transferProperty{indicator: indicator, ObjectPushData: ev}

		tmap, ok := sessionMap[ev.SessionID]
		if !ok {
			sessionMap[ev.SessionID] = make(map[bluetooth.ObjectPushTransferID]*transferProperty)
			tmap = sessionMap[ev.SessionID]
		}

		tmap[ev.TransferID] = property
		sessionMap[ev.SessionID] = tmap

		return property, true
	}

Transfer:
	for {
		select {
		case <-oppSub.Done:
			break Transfer

		case ev := <-oppSub.AddedEvents:
			_, _ = addIndicator(ev)

		case ev := <-oppSub.UpdatedEvents:
			property, ok := getIndicator(ev, false)
			if !ok {
				property, _ = addIndicator(bluetooth.ObjectPushData{ObjectPushEventData: ev})
			}

			p.drawIndicator(property.indicator, property.ObjectPushEventData)
			property.indicator.progressBar.Set64(int64(ev.Transferred))

			switch property.Status {
			case bluetooth.TransferError:
				p.status.ErrorMessage(fmt.Errorf("transfer could not be completed for %s", ev.Address.String()))
				fallthrough

			case bluetooth.TransferComplete:
				p.removeProgress(property.ObjectPushData)
				_, _ = getIndicator(ev, true)
			}

		case ev := <-oppSub.RemovedEvents:
			property, ok := getIndicator(ev, true)
			if ok {
				p.removeProgress(property.ObjectPushData)
			}
		}
	}
}

// startTransfer creates a new progress indicator, monitors the OBEX DBus interface for transfer events,
// and displays the progress on the screen. If the optional path parameter is provided, it means that
// a file is being received, and on transfer completion, the received file should be moved to a user-accessible
// directory.
func (p *progressView) startTransfer(address bluetooth.MacAddress, session bluetooth.ObexObjectPush, files []bluetooth.ObjectPushData) {
	psession := &progressViewSession{transferSession: session}
	psession.transfers = make(map[bluetooth.ObjectPushTransferID]struct{})
	for _, f := range files {
		psession.transfers[f.TransferID] = struct{}{}
	}

	p.sessions.Store(address, psession)
}

// suspendTransfer suspends the transfer.
// This does not work when a file is being received.
func (p *progressView) suspendTransfer() {
	transferProps, progress := p.transferData()
	if transferProps.Address.IsNil() {
		return
	}

	if progress.recv {
		p.status.InfoMessage("Cannot suspend receiving transfer", false)
		return
	}

	p.app.Session().Obex(progress.deviceAddress).ObjectPush().SuspendTransfer()
}

// resumeTransfer resumes the transfer.
// This does not work when a file is being received.
func (p *progressView) resumeTransfer() {
	transferProps, progress := p.transferData()
	if transferProps.Address.IsNil() {
		return
	}

	if progress.recv {
		p.status.InfoMessage("Cannot resume receiving transfer", false)
		return
	}

	p.app.Session().Obex(progress.deviceAddress).ObjectPush().ResumeTransfer()
}

// cancelTransfer cancels the transfer.
func (p *progressView) cancelTransfer() {
	transferProps, progress := p.transferData()
	if transferProps.Address.IsNil() {
		return
	}

	p.app.Session().Obex(progress.deviceAddress).ObjectPush().CancelTransfer()
}

// removeProgress removes the progress indicator from the screen.
func (p *progressView) removeProgress(transferProps bluetooth.ObjectPushData) {
	p.total.Add(^uint32(0))

	isComplete := transferProps.Status == bluetooth.TransferComplete
	path := transferProps.Filename

	if psession, ok := p.sessions.Load(transferProps.Address); ok {
		psession.mu.Lock()
		if !psession.sessionRemoved {
			delete(psession.transfers, transferProps.TransferID)

			if !isComplete || len(psession.transfers) == 0 {
				psession.sessionRemoved = true

				if psession.transferSession != nil {
					go psession.transferSession.RemoveSession()
				}

				p.sessions.Delete(transferProps.Address)
			}
		}
		psession.mu.Unlock()
	}

	p.app.QueueDraw(func() {
		for row := range p.view.GetRowCount() {
			cell := p.view.GetCell(row, 0)
			if cell == nil {
				continue
			}

			props, ok := cell.GetReference().(bluetooth.ObjectPushData)
			if !ok {
				continue
			}

			if props.TransferID == transferProps.TransferID {
				p.view.RemoveRow(row)
				p.view.RemoveRow(row - 1)

				break
			}
		}

		if p.total.Load() == 0 {
			p.statusProgress.Clear()
			p.status.SwitchToPage(statusMessagesPage.String())
		}
	})

	if path != "" && isComplete && transferProps.Receiving {
		go func() {
			if err := savefile(path, p.cfg.Values.ReceiveDir); err != nil {
				p.status.ErrorMessage(err)
			}
		}()
	}
}

// transferData gets the file transfer properties and the progress data
// from the current selection in the progress view.
func (p *progressView) transferData() (bluetooth.ObjectPushData, *progressIndicator) {
	row, _ := p.view.GetSelection()

	pathCell := p.view.GetCell(row, 0)
	if pathCell == nil {
		return bluetooth.ObjectPushData{}, nil
	}

	progCell := p.view.GetCell(row, 2)
	if progCell == nil {
		return bluetooth.ObjectPushData{}, nil
	}

	props, ok := pathCell.GetReference().(bluetooth.ObjectPushData)
	if !ok {
		return bluetooth.ObjectPushData{}, nil
	}

	progress, ok := progCell.GetReference().(*progressIndicator)
	if !ok {
		return bluetooth.ObjectPushData{}, nil
	}

	return props, progress
}

// Write is used by the progressbar to display the progress on the screen.
func (p *progressIndicator) Write(b []byte) (int, error) {
	p.appDrawFunc(func() {
		p.progress.SetText(string(b))
	})

	return 0, nil
}

// savefile moves a file from the obex cache to a specified user-accessible directory.
// If the directory is not specified, it automatically creates a directory in the
// user's home path and moves the file there.
func savefile(path string, userpath string) error {
	if userpath == "" {
		homedir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		userpath = filepath.Join(homedir, "bluetuith")

		if _, err := os.Stat(userpath); err != nil {
			err = os.Mkdir(userpath, 0700)
			if err != nil {
				return err
			}
		}
	}

	return os.Rename(path, filepath.Join(userpath, filepath.Base(path)))
}
