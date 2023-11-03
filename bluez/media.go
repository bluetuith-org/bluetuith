package bluez

import (
	"fmt"

	"github.com/godbus/dbus/v5"
	"github.com/pkg/errors"
)

const (
	dbusBluezMediaControlIface = "org.bluez.MediaControl1"
	dbusBluezMediaPlayerIface  = "org.bluez.MediaPlayer1"
)

// MediaProperties holds the media player information.
type MediaProperties struct {
	Status   string
	Position uint32
	Track    TrackProperties
}

// TrackProperties describes the track properties of
// the currently playing media.
type TrackProperties struct {
	Title       string
	Album       string
	Artist      string
	Duration    uint32
	TrackNumber uint32
	TotalTracks uint32
}

// InitMediaPlayer initializes the media player.
func (b *Bluez) InitMediaPlayer(devicePath string) error {
	mediaControl, err := b.GetMediaControlProperties(devicePath)
	if err != nil {
		return err
	}

	connected, ok := mediaControl["Connected"].Value().(bool)
	if !ok || !connected {
		return fmt.Errorf("Player is not connected")
	}

	playerPath, ok := mediaControl["Player"].Value().(dbus.ObjectPath)
	if !ok {
		return fmt.Errorf("Cannot get device's media player path")
	}

	b.SetCurrentPlayer(playerPath)

	return nil
}

// Play starts the media playback.
func (b *Bluez) Play() error {
	return b.CallMediaPlayer("Play")
}

// Pause suspends the media playback.
func (b *Bluez) Pause() error {
	return b.CallMediaPlayer("Pause")
}

// TogglePlayPause toggles the play/pause states.
func (b *Bluez) TogglePlayPause() error {
	status, err := b.GetMediaPlayerProperty("Status")
	if err != nil {
		return err
	}

	if status.(string) == "playing" {
		return b.Pause()
	} else if status.(string) == "paused" {
		return b.Play()
	}

	return nil
}

// Next switches to the next track.
func (b *Bluez) Next() error {
	return b.CallMediaPlayer("Next")
}

// Previous switches to the previous track.
func (b *Bluez) Previous() error {
	return b.CallMediaPlayer("Previous")
}

// FastForward forward-skips the currently playing track.
func (b *Bluez) FastForward() error {
	return b.CallMediaPlayer("FastForward")
}

// Rewind backward-skips the currently playing track.
func (b *Bluez) Rewind() error {
	return b.CallMediaPlayer("Rewind")
}

// Stop halts the media playback.
func (b *Bluez) Stop() error {
	return b.CallMediaPlayer("Stop")
}

// GetMediaProperties gets the media properties of the currently playing track.
func (b *Bluez) GetMediaProperties(values ...map[string]dbus.Variant) (MediaProperties, error) {
	var props MediaProperties
	var mediaPlayer map[string]dbus.Variant

	if values != nil {
		mediaPlayer = values[0]
	} else {
		mp, err := b.GetMediaPlayerProperties()
		if err != nil {
			return MediaProperties{}, err
		}

		mediaPlayer = mp
	}

	track := TrackProperties{
		Artist: "<Unknown Artist>",
		Album:  "<Unknown Album>",
	}
	if t, ok := mediaPlayer["Track"].Value().(map[string]dbus.Variant); ok {
		if err := DecodeVariantMap(t, &track); err != nil {
			return MediaProperties{}, err
		}

		if track.TrackNumber > 0 && track.TotalTracks == 0 {
			track.TotalTracks = track.TrackNumber
		}
	}
	delete(mediaPlayer, "Track")

	props.Track = track

	return props, DecodeVariantMap(mediaPlayer, &props)
}

// GetMediaPlayerProperties gets the media player properties.
func (b *Bluez) GetMediaPlayerProperties() (map[string]dbus.Variant, error) {
	player := b.GetCurrentPlayer()
	if player == "" {
		return nil, errors.New("No player path")
	}

	result := make(map[string]dbus.Variant)

	if err := b.conn.Object(dbusBluezName, player).
		Call(dbusPropertiesGetAllPath, 0, dbusBluezMediaPlayerIface).
		Store(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetMediaPlayerProperty gets the specified media player property.
func (b *Bluez) GetMediaPlayerProperty(property string) (interface{}, error) {
	var result interface{}

	player := b.GetCurrentPlayer()
	if player == "" {
		return nil, errors.New("No player path")
	}

	if err := b.conn.Object(dbusBluezName, player).
		Call(dbusPropertiesGetPath, 0, dbusBluezMediaPlayerIface, property).
		Store(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetMediaControlProperties gets the media control properties.
func (b *Bluez) GetMediaControlProperties(devicePath string) (map[string]dbus.Variant, error) {
	result := make(map[string]dbus.Variant)
	path := dbus.ObjectPath(devicePath)

	if err := b.conn.Object(dbusBluezName, path).
		Call(dbusPropertiesGetAllPath, 0, dbusBluezMediaControlIface).
		Store(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetCurrentPlayer gets the currently running player's path.
func (b *Bluez) GetCurrentPlayer() dbus.ObjectPath {
	b.PlayerLock.Lock()
	defer b.PlayerLock.Unlock()

	return b.CurrentPlayer
}

// SetCurrentPlayer sets the player path.
func (b *Bluez) SetCurrentPlayer(playerPath dbus.ObjectPath) {
	b.PlayerLock.Lock()
	defer b.PlayerLock.Unlock()

	b.CurrentPlayer = playerPath
}

// CallMediaPlayer is used to interact with the bluez MediaPlayer interface.
func (b *Bluez) CallMediaPlayer(command string) error {
	player := b.GetCurrentPlayer()
	if player == "" {
		return errors.New("No player path")
	}

	return b.conn.Object(dbusBluezName, player).
		Call(dbusBluezMediaPlayerIface+"."+command, 0).
		Store()
}
