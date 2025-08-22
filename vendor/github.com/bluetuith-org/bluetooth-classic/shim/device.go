//go:build !linux

package shim

import (
	"context"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/fmsg"
	"github.com/Southclaws/fault/ftag"
	"github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	"github.com/bluetuith-org/bluetooth-classic/api/errorkinds"
	"github.com/bluetuith-org/bluetooth-classic/shim/internal/commands"
	"github.com/google/uuid"
)

// device describes a function call interface to invoke device related functions.
type device struct {
	s       *ShimSession
	Address bluetooth.MacAddress
}

// Pair will attempt to pair a bluetooth device that is in pairing mode.
func (d *device) Pair() error {
	_, err := commands.Pair(d.Address).ExecuteWith(d.s.executor)
	return err
}

// CancelPairing will cancel a pairing attempt.
func (d *device) CancelPairing() error {
	_, err := commands.CancelPairing(d.Address).ExecuteWith(d.s.executor)
	return err
}

// Connect will attempt to connect an already paired bluetooth device
// to an device.
func (d *device) Connect() error {
	_, err := commands.Connect(d.Address).ExecuteWith(d.s.executor)
	return err
}

// Disconnect will disconnect the bluetooth device from the device.
func (d *device) Disconnect() error {
	_, err := commands.Disconnect(d.Address).ExecuteWith(d.s.executor)
	return err
}

// ConnectProfile will attempt to connect an already paired bluetooth device
// to an device, using a specific Bluetooth profile UUID .
func (d *device) ConnectProfile(profileUUID uuid.UUID) error {
	_, err := commands.ConnectProfile(d.Address, profileUUID).ExecuteWith(d.s.executor)

	return err
}

// DisconnectProfile will attempt to disconnect an already paired bluetooth device
// to an device, using a specific Bluetooth profile UUID .
func (d *device) DisconnectProfile(profileUUID uuid.UUID) error {
	_, err := commands.DisconnectProfile(d.Address, profileUUID).ExecuteWith(d.s.executor)

	return err
}

// Remove removes a device from its associated device.
func (d *device) Remove() error {
	_, err := commands.Remove(d.Address).ExecuteWith(d.s.executor)
	return err
}

// SetTrusted sets the device 'trust' status within its associated adapter.
// Currently is valid only on Linux.
func (d *device) SetTrusted(enable bool) error {
	return errorkinds.ErrNotSupported
}

// SetBlocked sets the device 'blocked' status within its associated adapter.
// Currently is valid only on Linux.
func (d *device) SetBlocked(enable bool) error {
	return errorkinds.ErrNotSupported
}

// Properties returns all the properties of the device.
func (d *device) Properties() (bluetooth.DeviceData, error) {
	return d.check()
}

func (d *device) check() (bluetooth.DeviceData, error) {
	switch {
	case d.s == nil || d.s.sessionClosed.Load():
		return bluetooth.DeviceData{}, fault.Wrap(errorkinds.ErrSessionNotExist,
			fctx.With(context.Background(),
				"error_at", "device-check-bus",
				"address", d.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Error while fetching device data"),
		)
	}

	device, err := d.s.store.Device(d.Address)
	if err != nil {
		return device, fault.Wrap(err,
			fctx.With(context.Background(),
				"error_at", "device-check-store",
				"address", d.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Adapter does not exist"),
		)
	}

	return device, nil
}

// appendProperties appends any additional properties to the provided device and returns
// the new result.
func (d *device) appendProperties(device bluetooth.DeviceData, adapter bluetooth.AdapterData) (bluetooth.DeviceData, error) {
	device.AssociatedAdapter = adapter.Address
	device.Type = bluetooth.DeviceTypeFromClass(device.Class)

	return device, nil
}
