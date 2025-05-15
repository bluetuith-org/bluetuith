package views

import (
	"errors"
	"fmt"
	"go.uber.org/atomic"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"slices"

	"github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	"github.com/darkhz/bluetuith/ui/keybindings"
	"github.com/darkhz/bluetuith/ui/theme"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
	"github.com/schollz/progressbar/v3"
)

const progressPage viewName = "progressview"
const progressViewButtonRegion = `["resume"][::b][Resume[][""] ["suspend"][::b][Pause[][""] ["cancel"][::b][Cancel[][""]`

// progressView describes a file transfer progress display.
type progressView struct {
	isSupported atomic.Bool

	view, statusProgress *tview.Table
	flex                 *tview.Flex

	total atomic.Uint32

	*Views
}

// progressIndicator describes a progress indicator, which will display
// a description and a progress bar.
type progressIndicator struct {
	desc        *tview.TableCell
	progress    *tview.TableCell
	progressBar *progressbar.ProgressBar

	recv   bool
	status bluetooth.FileTransferStatus

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
func (p *progressView) newIndicator(props bluetooth.FileTransferData, recv bool) *progressIndicator {
	var progress progressIndicator
	var progressText string

	if recv {
		progressText = "Receiving"
	} else {
		progressText = "Sending"
	}

	p.total.Add(1)

	count := p.total.Load()
	title := fmt.Sprintf(" [::b]%s %s[-:-:-]", progressText, props.Name)

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

	p.app.QueueDraw(func() {
		p.show()
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
	})

	return &progress
}

// startTransfer creates a new progress indicator, monitors the OBEX DBus interface for transfer events,
// and displays the progress on the screen. If the optional path parameter is provided, it means that
// a file is being received, and on transfer completion, the received file should be moved to a user-accessible
// directory.
func (p *progressView) startTransfer(transferProps bluetooth.FileTransferData, path ...string) bool {
	oppSub := bluetooth.FileTransferEvent().Subscribe()
	if !oppSub.Subscribable {
		p.status.ErrorMessage(fmt.Errorf("cannot subscribe to transfer event (file %s, device %s)", transferProps.Filename, transferProps.Address))
		return false
	}
	defer oppSub.Unsubscribe()

	progress := p.newIndicator(transferProps, path != nil)
	completed := false
	for fileTransferEvent := range oppSub.C {
		if fileTransferEvent.Data.Address != transferProps.Address {
			continue
		}
		if fileTransferEvent.Action == bluetooth.EventActionRemoved {
			p.removeProgress(fileTransferEvent.Data, completed, path...)
			return completed
		}

		completed = fileTransferEvent.Data.Status == bluetooth.TransferComplete
		switch fileTransferEvent.Data.Status {
		case bluetooth.TransferError:
			p.status.ErrorMessage(errors.New("Transfer has failed for " + transferProps.Name))
			fallthrough

		case bluetooth.TransferComplete:
			p.removeProgress(fileTransferEvent.Data, completed, path...)
			return completed
		}

		progress.progressBar.Set64(int64(fileTransferEvent.Data.Transferred))
	}

	return false
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

	p.app.Session().Obex(progress.deviceAddress).FileTransfer().SuspendTransfer()
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

	p.app.Session().Obex(progress.deviceAddress).FileTransfer().ResumeTransfer()
}

// cancelTransfer cancels the transfer.
// This does not work when a file is being received.
func (p *progressView) cancelTransfer() {
	transferProps, progress := p.transferData()
	if transferProps.Address.IsNil() {
		return
	}

	if progress.recv {
		p.status.InfoMessage("Cannot cancel receiving transfer", false)
		return
	}

	p.app.Session().Obex(progress.deviceAddress).FileTransfer().CancelTransfer()
}

func (p *progressView) removeProgress(transferProps bluetooth.FileTransferEventData, isComplete bool, path ...string) {
	p.total.Add(^uint32(0))

	p.app.QueueDraw(func() {
		for row := range p.view.GetRowCount() {
			cell := p.view.GetCell(row, 0)
			if cell == nil {
				continue
			}

			props, ok := cell.GetReference().(bluetooth.FileTransferData)
			if !ok {
				continue
			}

			if props.Address == transferProps.Address {
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

	if path != nil && isComplete {
		if err := savefile(path[0], p.cfg.Values.ReceiveDir); err != nil {
			p.status.ErrorMessage(err)
		}
	}
}

// transferData gets the file transfer properties and the progress data
// from the current selection in the progress view.
func (p *progressView) transferData() (bluetooth.FileTransferData, *progressIndicator) {
	row, _ := p.view.GetSelection()

	pathCell := p.view.GetCell(row, 0)
	if pathCell == nil {
		return bluetooth.FileTransferData{}, nil
	}

	progCell := p.view.GetCell(row, 2)
	if progCell == nil {
		return bluetooth.FileTransferData{}, nil
	}

	props, ok := pathCell.GetReference().(bluetooth.FileTransferData)
	if !ok {
		return bluetooth.FileTransferData{}, nil
	}

	progress, ok := progCell.GetReference().(*progressIndicator)
	if !ok {
		return bluetooth.FileTransferData{}, nil
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
