//go:build linux

package dbushelper

import (
	bluetooth "github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	errorkinds "github.com/bluetuith-org/bluetooth-classic/api/errorkinds"
	sstore "github.com/bluetuith-org/bluetooth-classic/api/helpers/sessionstore"
	"github.com/godbus/dbus/v5"
)

// PublishAdapterUpdateEvent publishes an adapter event after updating the session store.
func PublishAdapterUpdateEvent(store *sstore.SessionStore, signal *dbus.Signal, variants map[string]dbus.Variant) {
	go func() {
		address, ok := PathConverter.Address(DbusPathAdapter, signal.Path)
		if !ok {
			PublishSignalError(errorkinds.ErrAdapterNotFound, signal,
				"Bluez event handler error",
				"error_at", "pchanged-adapter-address",
			)

			return
		}

		updated, err := store.UpdateAdapter(address, DecodeAdapterFunc(variants))
		if err != nil {
			PublishSignalError(err, signal,
				"Bluez event handler error",
				"error_at", "pchanged-adapter-update",
			)

			return
		}

		bluetooth.AdapterEvents().PublishUpdated(updated)
	}()
}

// PublishDeviceUpdateEvent publishes a device event after updating the session store.
func PublishDeviceUpdateEvent(store *sstore.SessionStore, signal *dbus.Signal, variants map[string]dbus.Variant) {
	go func() {
		address, ok := PathConverter.Address(DbusPathDevice, signal.Path)
		if !ok {
			PublishSignalError(errorkinds.ErrDeviceNotFound, signal,
				"Bluez event handler error",
				"error_at", "pchanged-adapter-address",
			)

			return
		}

		updated, err := store.UpdateDevice(address, DecodeDeviceFunc(variants))
		if err != nil {
			PublishSignalError(err, signal,
				"Bluez event handler error",
				"error_at", "pchanged-adapter-update",
			)

			return
		}

		bluetooth.DeviceEvents().PublishUpdated(updated)
	}()
}
