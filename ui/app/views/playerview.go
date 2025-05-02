package views

import (
	"errors"
	"go.uber.org/atomic"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	"github.com/darkhz/bluetuith/ui/keybindings"
	"github.com/darkhz/bluetuith/ui/theme"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
	"golang.org/x/sync/semaphore"
)

// mediaPlayer holds the media player view.
type mediaPlayer struct {
	isSupported atomic.Bool
	isOpen      atomic.Bool
	skip        atomic.Bool

	keyEvent               chan string
	stopEvent, buttonEvent chan struct{}
	currentMedia           bluetooth.MediaPlayer

	playerLock *semaphore.Weighted

	*Views

	sync.Mutex
}

// playerElements holds the idividual player view display elements.
type playerElements struct {
	player                                *tview.Flex
	info, title, progress, track, buttons *tview.TextView
}

const mediaButtons = `["rewind"][::b][<<][""] ["prev"][::b][<][""] ["play"][::b][|>][""] ["next"][::b][>][""] ["fastforward"][::b][>>][""]`

// Initialize initializes the media player.
func (m *mediaPlayer) Initialize() error {
	m.isSupported.Store(true)

	return nil
}

// SetRootView sets the root view for the media player.
func (m *mediaPlayer) SetRootView(v *Views) {
	m.Views = v
}

// show shows the media player.
func (m *mediaPlayer) show() {
	if !m.isSupported.Load() {
		m.status.ErrorMessage(errors.New("this operation is not supported"))
		return
	}

	m.Lock()
	defer m.Unlock()

	device := m.device.getSelection(false)
	if device.Address.IsNil() {
		return
	}

	m.currentMedia = m.app.Session().MediaPlayer(device.Address)
	_, err := m.currentMedia.Properties()
	if err != nil {
		m.status.ErrorMessage(err)
		return
	}

	if m.keyEvent == nil {
		m.stopEvent = make(chan struct{})
		m.keyEvent = make(chan string, 1)
		m.buttonEvent = make(chan struct{}, 1)

		m.playerLock = semaphore.NewWeighted(1)
	}

	go m.updateLoop(device.Name)
}

// close closes the media player.
func (m *mediaPlayer) close() {
	if !m.isSupported.Load() {
		return
	}

	select {
	case <-m.stopEvent:

	case m.stopEvent <- struct{}{}:

	default:
	}
}

// getProgress returns the title and progress of the currently playing media.
func (m *mediaPlayer) getProgress(media bluetooth.MediaData, buttons string, width int, skip bool) (title string, newbuttons string, track string, progress string) {
	var length int

	title = media.Title
	position := media.Position
	duration := media.Duration
	number := strconv.FormatUint(uint64(media.TrackNumber), 10)
	total := strconv.FormatUint(uint64(media.TotalTracks), 10)

	button := "|>"
	oldButton := button

	if !skip {
		switch media.Status {
		case "playing":
			button = "||"

		case "paused":
			button = "|>"

		case "stopped":
			button = "[]"
		}
	}

	buttons = strings.Replace(buttons, oldButton, button, 1)

	width /= 2
	if position >= duration {
		position = duration
	}

	if duration <= 0 {
		length = 0
	} else {
		length = width * int(position) / int(duration)
	}

	endlength := width - length
	if endlength < 0 {
		endlength = width
	}

	track = "Track " + number + "/" + total
	progress = " " + formatDuration(position) +
		" |" + strings.Repeat("█", length) + strings.Repeat(" ", endlength) + "| " +
		formatDuration(duration)

	return title, buttons, track, progress
}

// updateLoop updates the media player.
func (m *mediaPlayer) updateLoop(deviceName string) {
	if !m.playerLock.TryAcquire(1) {
		return
	}
	defer m.playerLock.Release(1)

	mediaSub := bluetooth.MediaEvent().Subscribe()
	if !mediaSub.Subscribable {
		return
	}
	defer mediaSub.Unsubscribe()

	elements := m.setup(deviceName)
	m.app.InstantDraw(func() {
		m.help.swapStatusHelp(elements.player, true)
	})
	defer m.app.InstantDraw(func() {
		m.help.swapStatusHelp(elements.player, false)
	})

	m.isOpen.Store(true)
	defer m.isOpen.Store(false)

	t := time.NewTicker(1 * time.Second)
	defer t.Stop()

PlayerLoop:
	for {
		media, err := m.currentMedia.Properties()
		if err != nil {
			break PlayerLoop
		}

		_, _, width, _ := m.pages.GetRect()
		title, buttons, tracknum, progress := m.getProgress(media, mediaButtons, width, m.skip.Load())

		m.app.InstantDraw(func() {
			elements.info.SetText(media.Artist + " - " + media.Album)
			elements.title.SetText(title)
			elements.track.SetText(tracknum)
			elements.buttons.SetText(buttons)
			elements.progress.SetText(progress)
		})

		select {
		case <-m.stopEvent:
			break PlayerLoop

		case highlight, ok := <-m.keyEvent:
			if !ok {
				break PlayerLoop
			}

			elements.buttons.Highlight(highlight)
			t.Reset(1 * time.Second)

		case <-m.buttonEvent:
			t.Reset(1 * time.Second)

		case _, ok := <-mediaSub.C:
			if !ok {
				break PlayerLoop
			}

			t.Reset(1 * time.Second)

		case <-t.C:
		}
	}
}

// setup sets up the media player elements.
func (m *mediaPlayer) setup(deviceName string) playerElements {
	info := tview.NewTextView()
	info.SetDynamicColors(true)
	info.SetTextAlign(tview.AlignCenter)
	info.SetTextColor(theme.GetColor(theme.ThemeText))
	info.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	title := tview.NewTextView()
	title.SetDynamicColors(true)
	title.SetTextAlign(tview.AlignCenter)
	title.SetTextColor(theme.GetColor(theme.ThemeText))
	title.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	progress := tview.NewTextView()
	progress.SetDynamicColors(true)
	progress.SetTextAlign(tview.AlignCenter)
	progress.SetTextColor(theme.GetColor(theme.ThemeText))
	progress.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	track := tview.NewTextView()
	track.SetDynamicColors(true)
	track.SetTextAlign(tview.AlignLeft)
	track.SetTextColor(theme.GetColor(theme.ThemeText))
	track.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	device := tview.NewTextView()
	device.SetText(deviceName)
	device.SetDynamicColors(true)
	device.SetTextAlign(tview.AlignRight)
	device.SetTextColor(theme.GetColor(theme.ThemeText))
	device.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	buttons := tview.NewTextView()
	buttons.SetRegions(true)
	buttons.SetText(mediaButtons)
	buttons.SetDynamicColors(true)
	buttons.SetTextAlign(tview.AlignCenter)
	buttons.SetTextColor(theme.GetColor(theme.ThemeText))
	buttons.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))
	buttons.SetHighlightedFunc(func(added, _, _ []string) {
		var ch rune
		var key tcell.Key

		if added == nil {
			return
		}

		switch added[0] {
		case "play":
			key, ch = tcell.KeyRune, ' '

		case "next":
			key, ch = tcell.KeyRune, '>'

		case "prev":
			key, ch = tcell.KeyRune, '<'

		case "fastforward":
			buttons.Highlight(added[0])
			key, ch = tcell.KeyRight, '-'

		case "rewind":
			buttons.Highlight(added[0])
			key, ch = tcell.KeyLeft, '-'

		default:
			return
		}

		m.keyEvents(tcell.NewEventKey(key, ch, tcell.ModNone), true)
	})

	buttonFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(track, 0, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(buttons, 0, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(device, 0, 1, false)
	buttonFlex.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	player := tview.NewFlex().
		AddItem(nil, 1, 0, false).
		AddItem(title, 1, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(info, 1, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(progress, 1, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(buttonFlex, 1, 0, false).
		SetDirection(tview.FlexRow)
	player.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	return playerElements{player, info, title, progress, track, buttons}
}

// keyEvents handles the media player events.
func (m *mediaPlayer) keyEvents(event *tcell.EventKey, button bool) {
	if !m.isSupported.Load() || !m.isOpen.Load() {
		return
	}

	var highlight string
	var nokey bool

	operation := m.kb.Key(event)

	switch operation {
	case keybindings.KeyPlayerSeekForward:
		m.currentMedia.FastForward()

		m.skip.Store(true)
		highlight = "fastforward"

	case keybindings.KeyPlayerSeekBackward:
		m.currentMedia.Rewind()

		m.skip.Store(true)
		highlight = "rewind"

	case keybindings.KeyPlayerPrevious:
		m.currentMedia.Previous()

	case keybindings.KeyPlayerNext:
		m.currentMedia.Next()

	case keybindings.KeyPlayerStop:
		m.currentMedia.Stop()

	case keybindings.KeyPlayerTogglePlay:
		if m.skip.Load() {
			m.currentMedia.Play()
			m.skip.Store(false)

			break
		}

		m.currentMedia.TogglePlayPause()

	default:
		nokey = true
	}

	if !nokey {
		if button {
			select {
			case m.buttonEvent <- struct{}{}:

			default:
			}

			return
		}

		select {
		case m.keyEvent <- highlight:

		default:
		}
	}
}

// formatDuration converts a duration into a human-readable format.
func formatDuration(duration uint32) string {
	var durationtext string

	input, err := time.ParseDuration(strconv.FormatUint(uint64(duration), 10) + "ms")
	if err != nil {
		return "00:00"
	}

	d := input.Round(time.Second)

	h := d / time.Hour
	d -= h * time.Hour

	m := d / time.Minute
	d -= m * time.Minute

	s := d / time.Second

	if h > 0 {
		if h < 10 {
			durationtext += "0"
		}

		durationtext += strconv.Itoa(int(h))
		durationtext += ":"
	}

	if m > 0 {
		if m < 10 {
			durationtext += "0"
		}

		durationtext += strconv.Itoa(int(m))
	} else {
		durationtext += "00"
	}

	durationtext += ":"

	if s < 10 {
		durationtext += "0"
	}

	durationtext += strconv.Itoa(int(s))

	return durationtext
}
