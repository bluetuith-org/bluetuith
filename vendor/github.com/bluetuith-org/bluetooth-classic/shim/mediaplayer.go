//go:build !linux

package shim

import (
	"github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	"github.com/bluetuith-org/bluetooth-classic/api/errorkinds"
)

// mediaPlayer describes a function call interface to invoke media control related
// functions.
type mediaPlayer struct {
}

func (m *mediaPlayer) AudioProfiles() ([]bluetooth.AudioProfile, error) {
	return nil, errorkinds.ErrNotSupported
}

func (m *mediaPlayer) SetAudioProfile(profile bluetooth.AudioProfile) error {
	return errorkinds.ErrNotSupported
}

// Properties gets the media properties of the currently playing track.
func (m *mediaPlayer) Properties() (bluetooth.MediaData, error) {
	return bluetooth.MediaData{}, errorkinds.ErrNotSupported
}

// Play starts the media playback.
func (m *mediaPlayer) Play() error {
	return errorkinds.ErrNotSupported
}

// Pause suspends the media playback.
func (m *mediaPlayer) Pause() error {
	return errorkinds.ErrNotSupported
}

// TogglePlayPause toggles the play/pause states.
func (m *mediaPlayer) TogglePlayPause() error {
	return errorkinds.ErrNotSupported
}

// Next switches to the next track.
func (m *mediaPlayer) Next() error {
	return errorkinds.ErrNotSupported
}

// Previous switches to the previous track.
func (m *mediaPlayer) Previous() error {
	return errorkinds.ErrNotSupported
}

// FastForward forward-skips the currently playing track.
func (m *mediaPlayer) FastForward() error {
	return errorkinds.ErrNotSupported
}

// Rewind backward-skips the currently playing track.
func (m *mediaPlayer) Rewind() error {
	return errorkinds.ErrNotSupported
}

// Stop halts the media playback.
func (m *mediaPlayer) Stop() error {
	return errorkinds.ErrNotSupported
}
