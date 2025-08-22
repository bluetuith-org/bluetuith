//go:build linux

package mediaplayer

import (
	"context"
	"errors"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/fmsg"
	"github.com/Southclaws/fault/ftag"
	bluetooth "github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	errorkinds "github.com/bluetuith-org/bluetooth-classic/api/errorkinds"
	dbh "github.com/bluetuith-org/bluetooth-classic/linux/internal/dbushelper"
	"github.com/godbus/dbus/v5"
	"github.com/mafik/pulseaudio"
)

// MediaPlayer describes a function call interface to invoke media control related
// functions.
type MediaPlayer struct {
	SystemBus *dbus.Conn
	Address   bluetooth.MacAddress
}

// AudioProfiles lists all available audio profiles for use with a device.
func (m *MediaPlayer) AudioProfiles() ([]bluetooth.AudioProfile, error) {
	var profiles []bluetooth.AudioProfile

	client, err := pulseaudio.NewClient()
	if err != nil {
		return nil, fault.Wrap(
			err,
			fctx.With(context.Background(),
				"error_at", "media-audio-profiles",
				"address", m.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Cannot fetch audio profiles of device"),
		)
	}
	defer client.Close()

	cards, err := client.Cards()
	if err != nil {
		return nil, fault.Wrap(
			err,
			fctx.With(context.Background(),
				"error_at", "media-audio-profiles",
				"address", m.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Cannot fetch audio profiles of device"),
		)
	}

	for _, card := range cards {
		if addr, ok := card.PropList["device.string"]; ok {
			if addr != m.Address.String() {
				continue
			}

			for profileName, profile := range card.Profiles {
				if profile.Available != 1 {
					continue
				}

				profiles = append(profiles, bluetooth.AudioProfile{
					Index:       card.Index,
					Name:        profileName,
					Description: profile.Description,
					Active:      profile.Name == card.ActiveProfile.Name,
				})
			}

			return profiles, nil
		}
	}

	return nil, fault.Wrap(
		errors.New("profile set empty"),
		fctx.With(context.Background(),
			"error_at", "media-audio-profiles",
			"address", m.Address.String(),
		),
		ftag.With(ftag.Internal),
		fmsg.With("Cannot fetch audio profiles of device"),
	)
}

// SetAudioProfile sets the audio profile for the device.
func (m *MediaPlayer) SetAudioProfile(profile bluetooth.AudioProfile) error {
	client, err := pulseaudio.NewClient()
	if err != nil {
		return fault.Wrap(
			err,
			fctx.With(context.Background(),
				"error_at", "media-audio-profiles-set",
				"address", m.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Cannot set audio profile of device"),
		)
	}
	defer client.Close()

	if err := client.SetCardProfile(profile.Index, profile.Name); err != nil {
		return fault.Wrap(
			err,
			fctx.With(context.Background(),
				"error_at", "media-audio-profiles-set",
				"address", m.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Cannot set audio profile of device"),
		)
	}

	return nil
}

// Play starts the media playback.
func (m *MediaPlayer) Play() error {
	playerPath, err := m.check()
	if err != nil {
		return err
	}

	if err := m.callMediaPlayer(playerPath, "Play"); err != nil {
		return fault.Wrap(
			err,
			fctx.With(context.Background(),
				"error_at", "media-control-play",
				"address", m.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Cannot send 'play' media control command to device"),
		)
	}

	return nil
}

// Pause suspends the media playback.
func (m *MediaPlayer) Pause() error {
	playerPath, err := m.check()
	if err != nil {
		return err
	}

	if err := m.callMediaPlayer(playerPath, "Pause"); err != nil {
		return fault.Wrap(
			err,
			fctx.With(context.Background(),
				"error_at", "media-control-pause",
				"address", m.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Cannot send 'pause' media control command to device"),
		)
	}

	return nil
}

// TogglePlayPause toggles the play/pause states.
func (m *MediaPlayer) TogglePlayPause() error {
	playerPath, err := m.check()
	if err != nil {
		return err
	}

	status, err := m.mediaPlayerProperty(playerPath, "Status")
	if err != nil {
		return err
	}
	st, ok := status.(string)
	if !ok {
		return errors.New("invalid status identifier")
	}

	switch st {
	case "playing":
		return m.Pause()
	case "paused":
		return m.Play()
	}

	return nil
}

// Next switches to the next track.
func (m *MediaPlayer) Next() error {
	playerPath, err := m.check()
	if err != nil {
		return err
	}

	if err := m.callMediaPlayer(playerPath, "Next"); err != nil {
		return fault.Wrap(
			err,
			fctx.With(context.Background(),
				"error_at", "media-control-next",
				"address", m.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Cannot send 'next' media control command to device"),
		)
	}

	return nil
}

// Previous switches to the previous track.
func (m *MediaPlayer) Previous() error {
	playerPath, err := m.check()
	if err != nil {
		return err
	}

	if err := m.callMediaPlayer(playerPath, "Previous"); err != nil {
		return fault.Wrap(
			err,
			fctx.With(context.Background(),
				"error_at", "media-control-previous",
				"address", m.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Cannot send 'previous' media control command to device"),
		)
	}

	return nil
}

// FastForward forward-skips the currently playing track.
func (m *MediaPlayer) FastForward() error {
	playerPath, err := m.check()
	if err != nil {
		return err
	}

	if err := m.callMediaPlayer(playerPath, "FastForward"); err != nil {
		return fault.Wrap(
			err,
			fctx.With(context.Background(),
				"error_at", "media-control-fastForward",
				"address", m.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Cannot send 'fastForward' media control command to device"),
		)
	}

	return nil
}

// Rewind backward-skips the currently playing track.
func (m *MediaPlayer) Rewind() error {
	playerPath, err := m.check()
	if err != nil {
		return err
	}

	if err := m.callMediaPlayer(playerPath, "Rewind"); err != nil {
		return fault.Wrap(
			err,
			fctx.With(context.Background(),
				"error_at", "media-control-rewind",
				"address", m.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Cannot send 'rewind' media control command to device"),
		)
	}

	return nil
}

// Stop halts the media playback.
func (m *MediaPlayer) Stop() error {
	playerPath, err := m.check()
	if err != nil {
		return err
	}

	if err := m.callMediaPlayer(playerPath, "Stop"); err != nil {
		return fault.Wrap(
			err,
			fctx.With(context.Background(),
				"error_at", "media-control-stop",
				"address", m.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Cannot send 'stop' media control command to device"),
		)
	}

	return nil
}

// Properties gets the media properties of the currently playing track.
func (m *MediaPlayer) Properties() (bluetooth.MediaData, error) {
	playerPath, err := m.check()
	if err != nil {
		return bluetooth.MediaData{}, err
	}

	propertyMap, err := m.mediaPlayerProperties(playerPath)
	if err != nil {
		return bluetooth.MediaData{},
			fault.Wrap(err,
				fctx.With(context.Background(),
					"error_at", "media-prop-path",
					"address", m.Address.String(),
				),
				ftag.With(ftag.Internal),
				fmsg.With("Media player properties were not found for device"),
			)
	}

	properties, err := m.ParseMap(propertyMap)
	if err != nil {
		return bluetooth.MediaData{},
			fault.Wrap(err,
				fctx.With(context.Background(),
					"error_at", "media-prop-track",
					"address", m.Address.String(),
				),
				ftag.With(ftag.Internal),
				fmsg.With("Media track data cannot be parsed"),
			)
	}

	return properties, nil
}

// ParseMap parses a variant map of mediaplayer properties.
func (m *MediaPlayer) ParseMap(values map[string]dbus.Variant) (bluetooth.MediaData, error) {
	var props bluetooth.MediaData

	track := bluetooth.TrackData{}
	if t, ok := values["Track"].Value().(map[string]dbus.Variant); ok {
		if err := dbh.DecodeVariantMap(t, &track); err != nil {
			return bluetooth.MediaData{}, err
		}

		if track.TrackNumber > 0 && track.TotalTracks == 0 {
			track.TotalTracks = track.TrackNumber
		}
	}

	delete(values, "Track")

	props.TrackData = track

	if err := dbh.DecodeVariantMap(values, &props); err != nil {
		return bluetooth.MediaData{}, err
	}

	return props, nil
}

// check checks if the device supports media control and playback.
func (m *MediaPlayer) check() (dbus.ObjectPath, error) {
	devicePath, ok := dbh.PathConverter.DbusPath(dbh.DbusPathDevice, m.Address)
	if !ok {
		return "", fault.Wrap(errorkinds.ErrDeviceNotFound,
			fctx.With(context.Background(),
				"error_at", "device-check-store",
				"address", m.Address.String(),
			),
			ftag.With(ftag.NotFound),
			fmsg.With("Device does not exist"),
		)
	}

	mediaControl, err := m.mediaControlProperties(devicePath)
	if err != nil {
		return "", fault.Wrap(errorkinds.ErrPropertyDataParse,
			fctx.With(context.Background(),
				"error_at", "media-player-props",
				"address", m.Address.String(),
			),
			ftag.With(ftag.NotFound),
			fmsg.With("Cannot find/parse player properties"),
		)
	}

	connected, ok := mediaControl["Connected"].Value().(bool)
	if !ok || !connected {
		return "",
			fault.Wrap(errorkinds.ErrMediaPlayerNotConnected,
				fctx.With(context.Background(),
					"error_at", "media-player-conn",
					"address", m.Address.String(),
				),
				ftag.With(ftag.Internal),
				fmsg.With("Player is not connected"),
			)
	}

	playerPath, ok := mediaControl["Player"].Value().(dbus.ObjectPath)
	if !ok {
		return "",
			fault.Wrap(err,
				fctx.With(context.Background(),
					"error_at", "media-player-path",
					"address", m.Address.String(),
				),
				ftag.With(ftag.Internal),
				fmsg.With("Cannot get device's media player path"),
			)
	}

	return playerPath, nil
}

// mediaPlayerProperties gets the media player properties.
func (m *MediaPlayer) mediaPlayerProperties(player dbus.ObjectPath) (map[string]dbus.Variant, error) {
	result := make(map[string]dbus.Variant)

	if err := m.SystemBus.Object(dbh.BluezBusName, player).
		Call(dbh.DbusGetAllPropertiesIface, 0, dbh.BluezMediaPlayerIface).
		Store(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// mediaPlayerProperty gets the specified media player property.
func (m *MediaPlayer) mediaPlayerProperty(player dbus.ObjectPath, property string) (any, error) {
	var result any

	if err := m.SystemBus.Object(dbh.BluezBusName, player).
		Call(dbh.DbusGetPropertiesIface, 0, dbh.BluezMediaPlayerIface, property).
		Store(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// mediaControlProperties gets the media control properties.
func (m *MediaPlayer) mediaControlProperties(devicePath dbus.ObjectPath) (map[string]dbus.Variant, error) {
	result := make(map[string]dbus.Variant)

	if err := m.SystemBus.Object(dbh.BluezBusName, devicePath).
		Call(dbh.DbusGetAllPropertiesIface, 0, dbh.BluezMediaControlIface).
		Store(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// callMediaPlayer is used to interact with the bluez MediaPlayer interface.
func (m *MediaPlayer) callMediaPlayer(player dbus.ObjectPath, command string) error {
	return m.SystemBus.Object(dbh.BluezBusName, player).
		Call(dbh.BluezMediaPlayerIface+"."+command, 0).
		Store()
}
