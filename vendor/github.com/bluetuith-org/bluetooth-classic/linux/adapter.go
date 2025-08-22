//go:build linux

package linux

import (
	"context"
	"path/filepath"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/fmsg"
	"github.com/Southclaws/fault/ftag"
	bluetooth "github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	errorkinds "github.com/bluetuith-org/bluetooth-classic/api/errorkinds"
	dbh "github.com/bluetuith-org/bluetooth-classic/linux/internal/dbushelper"
	"github.com/godbus/dbus/v5"
)

// adapter describes a function call interface to invoke adapter related functions.
type adapter struct {
	b    *BluezSession
	path dbus.ObjectPath

	Address bluetooth.MacAddress
}

// StartDiscovery will put the adapter into "discovering" mode, which means
// the bluetooth device will be able to discover other bluetooth devices
// that are in pairing mode.
func (a *adapter) StartDiscovery() error {
	if _, err := a.check(); err != nil {
		return err
	}

	if err := a.callAdapter("StartDiscovery", 0).Store(); err != nil {
		return fault.Wrap(err,
			fctx.With(context.Background(),
				"error_at", "adapter-start-discovery",
				"address", a.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("An error occurred while starting device discovery"),
		)
	}

	return nil
}

// StopDiscovery will stop the  "discovering" mode, which means the bluetooth device will
// no longer be able to discover other bluetooth devices that are in pairing mode.
func (a *adapter) StopDiscovery() error {
	if _, err := a.check(); err != nil {
		return err
	}

	if err := a.callAdapter("StopDiscovery", 0).Store(); err != nil {
		return fault.Wrap(err,
			fctx.With(context.Background(),
				"error_at", "adapter-stop-discovery",
				"address", a.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("An error occurred while stopping device discovery"),
		)
	}

	return nil
}

// SetPoweredState sets the powered state of the adapter.
func (a *adapter) SetPoweredState(enable bool) error {
	if _, err := a.check(); err != nil {
		return err
	}

	if err := a.setAdapterProperty("Powered", enable); err != nil {
		return fault.Wrap(err,
			fctx.With(context.Background(),
				"error_at", "adapter-setpowered-state",
				"address", a.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("An error occurred on setting powered state"),
		)
	}

	return nil
}

// SetDiscoverableState sets the discoverable state of the adapter.
func (a *adapter) SetDiscoverableState(enable bool) error {
	if _, err := a.check(); err != nil {
		return err
	}

	if err := a.setAdapterProperty("Discoverable", enable); err != nil {
		return fault.Wrap(err,
			fctx.With(context.Background(),
				"error_at", "adapter-setdiscoverable-state",
				"address", a.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("An error occurred on setting discoverable state"),
		)
	}

	return nil
}

// SetPairableState sets the pairable state of the adapter.
func (a *adapter) SetPairableState(enable bool) error {
	if _, err := a.check(); err != nil {
		return err
	}

	if err := a.setAdapterProperty("Pairable", enable); err != nil {
		return fault.Wrap(err,
			fctx.With(context.Background(),
				"error_at", "adapter-setpairable-state",
				"address", a.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("An error occurred on setting pairable state"),
		)
	}

	return nil
}

// Properties returns all the properties of the adapter.
func (a *adapter) Properties() (bluetooth.AdapterData, error) {
	return a.check()
}

// Devices returns all the devices associated with the adapter.
func (a *adapter) Devices() ([]bluetooth.DeviceData, error) {
	if _, err := a.check(); err != nil {
		return nil, err
	}

	devices, err := a.b.store.AdapterDevices(a.Address)
	if err != nil {
		return nil,
			fault.Wrap(err,
				fctx.With(context.Background(),
					"error_at", "adapter-fetch-devices",
					"address", a.Address.String(),
				),
				ftag.With(ftag.Internal),
				fmsg.With("Error while fetching adapter devices"),
			)
	}

	return devices, nil
}

// check validates whether a valid DBus path is associated with the provided
// adapter's address ((*Adapter).Address), and checks whether the adapter
// properties are present within the global session store.
func (a *adapter) check() (bluetooth.AdapterData, error) {
	dbusPath, exists := dbh.PathConverter.DbusPath(dbh.DbusPathAdapter, a.Address)

	switch {
	case a.b == nil:
		return bluetooth.AdapterData{}, fault.Wrap(errorkinds.ErrAdapterNotFound,
			fctx.With(context.Background(),
				"error_at", "adapter-check-bus",
				"address", a.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Error while fetching adapter data"),
		)

	case !exists:
		return bluetooth.AdapterData{}, fault.Wrap(errorkinds.ErrAdapterNotFound,
			fctx.With(context.Background(),
				"error_at", "adapter-check-path",
				"address", a.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Error while fetching adapter data"),
		)
	}

	a.path = dbusPath

	adapter, err := a.b.store.Adapter(a.Address)
	if err != nil {
		return adapter, fault.Wrap(err,
			fctx.With(context.Background(),
				"error_at", "adapter-check-store",
				"address", a.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Adapter does not exist"),
		)
	}

	return adapter, nil
}

// callAdapter is used to interact with the bluez Adapter dbus interface.
// https://git.kernel.org/pub/scm/bluetooth/bluez.git/tree/doc/adapter-api.txt
func (a *adapter) callAdapter(method string, flags dbus.Flags, args ...any) *dbus.Call {
	return a.b.systemBus.Object(dbh.BluezBusName, a.path).
		Call(dbh.BluezAdapterIface+"."+method, flags, args...)
}

// adapterProperties gathers all the properties for a bluetooth adapter.
func (a *adapter) adapterProperties() (map[string]dbus.Variant, error) {
	result := make(map[string]dbus.Variant)
	if err := a.b.systemBus.Object(dbh.BluezBusName, a.path).
		Call(dbh.DbusGetAllPropertiesIface, 0, dbh.BluezAdapterIface).
		Store(&result); err != nil {
		return result, err
	}

	return result, nil
}

// setAdapterProperty can be used to set certain properties for a bluetooth adapter.
func (a *adapter) setAdapterProperty(key string, value any) error {
	return a.b.systemBus.Object(dbh.BluezBusName, a.path).Call(
		dbh.DbusSetPropertiesIface, 0, dbh.BluezAdapterIface,
		key, dbus.MakeVariant(value),
	).Store()
}

// convertAndStoreObjectseObjectseObjects converts a map of dbus objects to a common AdapterData structure.
func (a *adapter) convertAndStoreObjects(values map[string]dbus.Variant) (bluetooth.AdapterData, error) {
	/*
		/org/bluez/hci0
			org.bluez.Adapter1
					Discoverable => dbus.Variant{sig:dbus.Signature{str:"b"}, value:true}
					UUIDs => dbus.Variant{sig:dbus.Signature{str:"as"}, value:[]string{"00001112-0000-1000-8000-00805f9b34fb", "00001801-0000-1000-8000-00805f9b34fb", "0000110e-0000-1000-8000-00805f9b34fb", "00001800-0000-1000-8000-00805f9b34fb", "00001200-0000-1000-8000-00805f9b34fb", "0000110c-0000-1000-8000-00805f9b34fb", "0000110b-0000-1000-8000-00805f9b34fb", "0000110a-0000-1000-8000-00805f9b34fb"}}
					Modalias => dbus.Variant{sig:dbus.Signature{str:"s"}, value:"usb:v1D6Bp0246d0525"}
					Pairable => dbus.Variant{sig:dbus.Signature{str:"b"}, value:true}
					DiscoverableTimeout => dbus.Variant{sig:dbus.Signature{str:"u"}, value:0x0}
					PairableTimeout => dbus.Variant{sig:dbus.Signature{str:"u"}, value:0x0}
					Powered => dbus.Variant{sig:dbus.Signature{str:"b"}, value:true}
					Class => dbus.Variant{sig:dbus.Signature{str:"u"}, value:0xc010c}
					Discovering => dbus.Variant{sig:dbus.Signature{str:"b"}, value:true}
					Address => dbus.Variant{sig:dbus.Signature{str:"s"}, value:"9C:B6:D0:1C:BB:B0"}
					Name => dbus.Variant{sig:dbus.Signature{str:"s"}, value:"jonathan-Blade"}
					Alias => dbus.Variant{sig:dbus.Signature{str:"s"}, value:"jonathan-Blade"}

	*/
	var adapter bluetooth.AdapterData

	if err := dbh.DecodeVariantMap(values, &adapter, "Address"); err != nil {
		return adapter, fault.Wrap(err,
			fctx.With(context.Background(),
				"error_at", "adapter-map-decode",
				"address", adapter.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Error converting adapter data"),
		)
	}

	dbh.PathConverter.AddDbusPath(dbh.DbusPathAdapter, a.path, adapter.Address)
	adapter.UniqueName = filepath.Base(string(a.path))

	a.b.store.AddAdapter(adapter)

	return adapter, nil
}
