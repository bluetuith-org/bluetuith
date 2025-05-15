package views

import (
	"errors"
	"math/bits"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/atomic"

	"github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"

	"github.com/darkhz/bluetuith/ui/keybindings"
	"github.com/darkhz/bluetuith/ui/theme"
)

// mediaPlayer holds the media player view.
type mediaPlayer struct {
	isSupported atomic.Bool
	isOpen      atomic.Bool
	skip        atomic.Bool

	keyEvent               chan string
	stopEvent, buttonEvent chan struct{}
	currentMedia           bluetooth.MediaPlayer

	*Views

	sync.Mutex
}

// playerElements holds the individual player view display elements.
type playerElements struct {
	player                                *tview.Flex
	info, title, progress, track, buttons *tview.TextView
}

// Initialize initializes the media player.
func (m *mediaPlayer) Initialize() error {
	m.isSupported.Store(true)

	m.stopEvent = make(chan struct{})
	m.keyEvent = make(chan string, 1)
	m.buttonEvent = make(chan struct{}, 1)

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

	if m.isOpen.Load() {
		return
	}

	m.Lock()
	defer m.Unlock()

	device := m.device.getSelection(false)
	if device.Address.IsNil() {
		return
	}

	m.currentMedia = m.app.Session().MediaPlayer(device.Address)
	properties, err := m.currentMedia.Properties()
	if err != nil {
		m.status.ErrorMessage(err)
		return
	}

	go m.updateLoop(device, properties)
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

// renderProgress renders the player progress bar.
func (m *mediaPlayer) renderProgress(progressView *tview.TextView, media bluetooth.MediaData) {
	var length int

	_, _, width, _ := m.pages.GetRect()
	position := media.Position
	duration := media.Duration

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

	var sb strings.Builder

	sb.WriteString(" ")
	sb.WriteString(formatDuration(position))
	sb.WriteString(" |")
	sb.WriteString(strings.Repeat("█", length))
	sb.WriteString(strings.Repeat(" ", endlength))
	sb.WriteString("| ")
	sb.WriteString(formatDuration(duration))

	progressView.SetText(sb.String())
}

// renderButtons renders the player buttons.
func (m *mediaPlayer) renderButtons(buttonsView *tview.TextView, mediaStatus bluetooth.MediaStatus, skip bool) {
	const (
		mediaLeftButtons  = `["rewind"][::b]<<[""] ["prev"][::b]<[""] ["play"][::b]`
		mediaRightButtons = `[""] ["next"][::b]>[""] ["fastforward"][::b]>>[""]`
	)

	button := "|>"

	if !skip {
		switch mediaStatus {
		case bluetooth.MediaPlaying:
			button = "||"

		case bluetooth.MediaPaused:
			button = "|>"

		case bluetooth.MediaStopped:
			button = "[]"
		}
	}

	var sb strings.Builder

	sb.WriteString(mediaLeftButtons)
	sb.WriteString(button)
	sb.WriteString(mediaRightButtons)

	buttonsView.SetText(sb.String())
}

// renderTrackData renders the track details.
func (m *mediaPlayer) renderTrackData(infoView, titleView, trackView *tview.TextView, trackData bluetooth.TrackData) {
	number := strconv.FormatUint(uint64(trackData.TrackNumber), 10)
	total := strconv.FormatUint(uint64(trackData.TotalTracks), 10)

	track := "Track " + number + "/" + total

	titleView.SetText(trackData.Title)
	infoView.SetText(trackData.Artist + " - " + trackData.Album)
	trackView.SetText(track)
}

// renderPlayer renders the entire media player.
func (m *mediaPlayer) renderPlayer(cached bluetooth.MediaData, elements playerElements, track, progress, buttons bool) bool {
	if track {
		m.renderTrackData(elements.info, elements.title, elements.track, cached.TrackData)
	}

	if progress {
		m.renderProgress(elements.progress, cached)
	}

	if buttons {
		m.renderButtons(elements.buttons, cached.Status, m.skip.Load())
	}

	return track || progress || buttons
}

// updateLoop updates the media player.
func (m *mediaPlayer) updateLoop(device bluetooth.DeviceData, props bluetooth.MediaData) {
	mediaSub := bluetooth.MediaEvent().Subscribe()
	if !mediaSub.Subscribable {
		return
	}
	defer mediaSub.Unsubscribe()

	elements := m.setup(device.Name)
	go m.app.QueueDraw(func() {
		m.help.swapStatusHelp(elements.player, true)
	})
	defer m.app.QueueDraw(func() {
		m.help.swapStatusHelp(elements.player, false)
	})

	m.isOpen.Store(true)
	defer m.isOpen.Store(false)

	t := time.NewTicker(1 * time.Second)
	defer t.Stop()

	var cached bluetooth.MediaEventData
	var delta uint32 = 0

	cached.MediaData = props
	if cached.Title == "" {
		cached.Title = "<No media is playing>"
	}

	m.app.QueueDraw(func() {
		m.renderPlayer(props, elements, true, true, true)
	})

PlayerLoop:
	for {
		select {
		case <-m.stopEvent:
			break PlayerLoop

		case h := <-m.keyEvent:
			switch h {
			case "fastforward", "rewind":
				t.Reset(250 * time.Millisecond)
			default:
				t.Reset(1 * time.Second)
			}

			go m.app.QueueDraw(func() {
				elements.buttons.Highlight(h)
			})

		case mediaEvent, ok := <-mediaSub.C:
			if !ok {
				break PlayerLoop
			}

			if mediaEvent.Action != bluetooth.EventActionUpdated || mediaEvent.Data.Address != device.Address {
				continue
			}

			var track, progress, buttons bool

			if mediaEvent.Data.TrackData != (bluetooth.TrackData{}) && mediaEvent.Data.TrackData != cached.TrackData {
				cached.TrackData = mediaEvent.Data.TrackData
				track = true
			}

			if mediaEvent.Data.MediaData != (bluetooth.MediaData{}) && mediaEvent.Data.MediaData != cached.MediaData {
				if mediaEvent.Data.Status != "" && mediaEvent.Data.Status != cached.Status {
					cached.Status = mediaEvent.Data.Status
					buttons = true
				}

				if mediaEvent.Data.Position > 0 {
					switch cached.Status {
					case bluetooth.MediaForwardSeek, bluetooth.MediaReverseSeek:
						if mediaEvent.Data.Position > cached.Position {
							delta = mediaEvent.Data.Position - cached.Position
						} else {
							delta = cached.Position - mediaEvent.Data.Position
						}

					default:
						t.Reset(1 * time.Second)
						delta = 0
					}

					if mediaEvent.Data.Position != cached.Position {
						cached.Position = mediaEvent.Data.Position
						progress = true
					}
				}
			}

			data := cached.MediaData
			m.app.QueueDraw(func() {
				if m.renderPlayer(data, elements, track, progress, buttons) {
					t.Reset(1 * time.Second)
				}
			})

		case <-t.C:
			if cached.Position >= cached.Duration {
				continue
			}

			switch cached.Status {
			case bluetooth.MediaForwardSeek:
				t.Reset(250 * time.Millisecond)
				pos, c := bits.Add32(cached.Position, delta, 0)
				if c != 0 {
					cached.Position = 0
					break
				}
				cached.Position = pos

			case bluetooth.MediaReverseSeek:
				t.Reset(250 * time.Millisecond)
				pos, b := bits.Sub32(cached.Position, delta, 0)
				if b != 0 {
					cached.Position = 0
					break
				}
				cached.Position = pos

			case bluetooth.MediaPlaying:
				cached.Position += 1000

			default:
				continue
			}

			mediaData := cached.MediaData
			m.app.QueueDraw(func() {
				m.renderProgress(elements.progress, mediaData)
			})
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

	buttonNames := []string{
		"play",
		"next",
		"prev",
		"fastforward",
		"rewind",
	}

	buttons := tview.NewTextView()
	buttons.SetRegions(true)
	buttons.SetDynamicColors(true)
	buttons.SetTextAlign(tview.AlignCenter)
	buttons.SetTextColor(theme.GetColor(theme.ThemeText))
	buttons.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))
	buttons.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		if action != tview.MouseLeftClick {
			return action, event
		}

		var key keybindings.Key
		var region string

		x, y := event.Position()
		rectx, recty, _, _ := buttons.GetInnerRect()

		x -= rectx
		y -= recty

		for _, r := range buttonNames {
			start := buttons.GetRegionStart(r)
			if x == start || x == start+1 {
				region = r
				break
			}
		}
		if region == "" {
			return action, event
		}

		switch region {
		case "play":
			key = keybindings.KeyPlayerTogglePlay

		case "next":
			key = keybindings.KeyPlayerNext

		case "prev":
			key = keybindings.KeyPlayerPrevious

		case "fastforward":
			key = keybindings.KeyPlayerSeekForward

		case "rewind":
			key = keybindings.KeyPlayerSeekBackward

		default:
			return action, event
		}

		keyData := m.kb.Data(key)
		go m.keyEvents(tcell.NewEventKey(keyData.Kb.Key, keyData.Kb.Rune, keyData.Kb.Mod))

		return action, event
	})
	buttons.SetHighlightedFunc(func(added, re, rm []string) {
		if added == nil {
			return
		}

		if added[0] == "fastforward" || added[0] == "rewind" {
			return
		}

		go func() {
			time.Sleep(100 * time.Millisecond)
			m.app.QueueDraw(func() {
				buttons.Highlight()
			})
		}()
	})
	m.renderButtons(buttons, bluetooth.MediaPlaying, false)

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
func (m *mediaPlayer) keyEvents(event *tcell.EventKey) {
	if !m.isSupported.Load() || !m.isOpen.Load() {
		return
	}

	var nokey bool
	var highlight string

	switch m.kb.Key(event, keybindings.ContextDevice) {
	case keybindings.KeyPlayerSeekForward:
		highlight = "fastforward"
		m.currentMedia.FastForward()
		m.skip.Store(true)

	case keybindings.KeyPlayerSeekBackward:
		highlight = "rewind"
		m.currentMedia.Rewind()
		m.skip.Store(true)

	case keybindings.KeyPlayerPrevious:
		highlight = "prev"
		m.currentMedia.Previous()

	case keybindings.KeyPlayerNext:
		highlight = "next"
		m.currentMedia.Next()

	case keybindings.KeyPlayerStop:
		highlight = "play"
		m.currentMedia.Stop()

	case keybindings.KeyPlayerTogglePlay:
		highlight = "play"
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
		select {
		case m.keyEvent <- highlight:

		default:
		}
	}
}

// formatDuration converts a duration into a human-readable format.
func formatDuration(duration uint32) string {
	var durationtext strings.Builder

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
			durationtext.WriteString("0")
		}

		durationtext.WriteString(strconv.Itoa(int(h)))
		durationtext.WriteString(":")
	}

	if m > 0 {
		if m < 10 {
			durationtext.WriteString("0")
		}

		durationtext.WriteString(strconv.Itoa(int(m)))
	} else {
		durationtext.WriteString("00")
	}

	durationtext.WriteString(":")

	if s < 10 {
		durationtext.WriteString("0")
	}

	durationtext.WriteString(strconv.Itoa(int(s)))

	return durationtext.String()
}
