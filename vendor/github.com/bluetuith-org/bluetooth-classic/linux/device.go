//go:build linux

package linux

import (
	"context"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/fmsg"
	"github.com/Southclaws/fault/ftag"
	bluetooth "github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	errorkinds "github.com/bluetuith-org/bluetooth-classic/api/errorkinds"
	dbh "github.com/bluetuith-org/bluetooth-classic/linux/internal/dbushelper"
	"github.com/godbus/dbus/v5"
	"github.com/google/uuid"
)

// device describes a function call interface to invoke device related functions.
type device struct {
	b    *BluezSession
	path dbus.ObjectPath

	Address bluetooth.MacAddress
}

// Pair will attempt to pair a bluetooth device that is in pairing mode.
func (d *device) Pair() error {
	if _, err := d.check(); err != nil {
		return err
	}

	if err := d.callDevice("Pair", 0).Store(); err != nil {
		return fault.Wrap(err,
			fctx.With(context.Background(),
				"error_at", "device-pair",
				"address", d.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Cannot pair with device"),
		)
	}

	return nil
}

// CancelPairing will cancel a pairing attempt.
func (d *device) CancelPairing() error {
	if _, err := d.check(); err != nil {
		return err
	}

	if err := d.callDevice("CancelPairing", 0).Store(); err != nil {
		return fault.Wrap(err,
			fctx.With(context.Background(),
				"error_at", "device-cancelpairing",
				"address", d.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("An error occurred while cancelling pairing"),
		)
	}

	return nil
}

// Connect will attempt to connect an already paired bluetooth device
// to an adapter.
func (d *device) Connect() error {
	if _, err := d.check(); err != nil {
		return err
	}

	if err := d.callDevice("Connect", 0).Store(); err != nil {
		return fault.Wrap(err,
			fctx.With(context.Background(),
				"error_at", "device-connect",
				"address", d.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Cannot connect to device"),
		)
	}

	return nil
}

// Disconnect will disconnect the bluetooth device from the adapter.
func (d *device) Disconnect() error {
	if _, err := d.check(); err != nil {
		return err
	}

	if err := d.callDevice("Disconnect", 0).Store(); err != nil {
		return fault.Wrap(err,
			fctx.With(context.Background(),
				"error_at", "device-disconnect",
				"address", d.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Cannot disconnect from device"),
		)
	}

	return nil
}

// ConnectProfile will attempt to connect an already paired bluetooth device
// to an adapter, using a specific Bluetooth profile UUID .
func (d *device) ConnectProfile(profileUUID uuid.UUID) error {
	if _, err := d.check(); err != nil {
		return err
	}

	if err := d.callDevice("ConnectProfile", 0, profileUUID.String()).Store(); err != nil {
		return fault.Wrap(err,
			fctx.With(context.Background(),
				"error_at", "device-connect-profile",
				"address", d.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Cannot connect to device with profile"),
		)
	}

	return nil
}

// DisconnectProfile will attempt to disconnect an already paired bluetooth device
// to an adapter, using a specific Bluetooth profile UUID .
func (d *device) DisconnectProfile(profileUUID uuid.UUID) error {
	if _, err := d.check(); err != nil {
		return err
	}

	if err := d.callDevice("DisconnectProfile", 0, profileUUID.String()).Store(); err != nil {
		return fault.Wrap(err,
			fctx.With(context.Background(),
				"error_at", "device-disconnect-profile",
				"address", d.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Cannot disconnect from device with profile"),
		)
	}

	return nil
}

// Remove removes a device from its associated adapter.
func (d *device) Remove() error {
	device, err := d.check()
	if err != nil {
		return err
	}

	adapterPath, ok := dbh.PathConverter.DbusPath(dbh.DbusPathAdapter, device.AssociatedAdapter)
	if !ok {
		return fault.Wrap(errorkinds.ErrAdapterNotFound,
			fctx.With(context.Background(),
				"error_at", "device-remove-adapterpath",
				"address", d.Address.String(),
			),
			ftag.With(ftag.NotFound),
			fmsg.With("Adapter does not exist"),
		)
	}

	if err := d.b.adapterInternal(adapterPath).callAdapter("RemoveDevice", 0, d.path).Store(); err != nil {
		return fault.Wrap(err,
			fctx.With(context.Background(),
				"error_at", "device-remove-methodcall",
				"address", d.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Cannot remove device"),
		)
	}

	return nil
}

// SetTrusted sets the device 'trust' status within its associated adapter.
// Currently is valid only on Linux.
func (d *device) SetTrusted(enable bool) error {
	if _, err := d.check(); err != nil {
		return err
	}

	if err := d.setDeviceProperty(d.path, "Trusted", enable); err != nil {
		return fault.Wrap(err,
			fctx.With(context.Background(),
				"error_at", "device-trust-method",
				"address", d.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Cannot set device 'trust' status"),
		)
	}

	return nil
}

// SetBlocked sets the device 'blocked' status within its associated adapter.
// Currently is valid only on Linux.
func (d *device) SetBlocked(enable bool) error {
	if _, err := d.check(); err != nil {
		return err
	}

	if err := d.setDeviceProperty(d.path, "Blocked", enable); err != nil {
		return fault.Wrap(err,
			fctx.With(context.Background(),
				"error_at", "device-blocked-method",
				"address", d.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Cannot set device 'blocked' status"),
		)
	}

	return nil
}

// Properties returns all the properties of the device.
func (d *device) Properties() (bluetooth.DeviceData, error) {
	return d.check()
}

// check validates whether a valid DBus path is associated with the provided
// device's address ((*Device).Address), and checks whether the device
// properties are present within the global session store.
func (d *device) check() (bluetooth.DeviceData, error) {
	dbusPath, exists := dbh.PathConverter.DbusPath(dbh.DbusPathDevice, d.Address)

	switch {
	case d.b == nil:
		return bluetooth.DeviceData{}, fault.Wrap(errorkinds.ErrDeviceNotFound,
			fctx.With(context.Background(),
				"error_at", "device-check-bus",
				"address", d.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Error while fetching device data"),
		)

	case !exists:
		return bluetooth.DeviceData{}, fault.Wrap(errorkinds.ErrDeviceNotFound,
			fctx.With(context.Background(),
				"error_at", "device-check-path",
				"address", d.Address.String(),
			),
			ftag.With(ftag.NotFound),
			fmsg.With("Error while fetching device data"),
		)
	}

	d.path = dbusPath

	device, err := d.b.store.Device(d.Address)
	if err != nil {
		return bluetooth.DeviceData{}, fault.Wrap(err,
			fctx.With(context.Background(),
				"error_at", "device-check-store",
				"address", d.Address.String(),
			),
			ftag.With(ftag.NotFound),
			fmsg.With("Device does not exist"),
		)
	}

	return device, nil
}

// callDevice is used to interact with the bluez Device dbus interface.
// https://git.kernel.org/pub/scm/bluetooth/bluez.git/tree/doc/device-api.txt
func (d *device) callDevice(method string, flags dbus.Flags, args ...any) *dbus.Call {
	return d.b.systemBus.Object(dbh.BluezBusName, d.path).
		Call(dbh.BluezDeviceIface+"."+method, flags, args...)
}

// setDeviceProperty can be used to set certain properties for a bluetooth device.
func (d *device) setDeviceProperty(devicePath dbus.ObjectPath, key string, value any) error {
	return d.b.systemBus.Object(dbh.BluezBusName, devicePath).Call(dbh.DbusSetPropertiesIface, 0, dbh.BluezDeviceIface, key, dbus.MakeVariant(value)).Store()
}

// convertAndStoreObjects converts a map of dbus objects to a common DeviceData structure.
func (d *device) convertAndStoreObjects(values map[string]dbus.Variant) (bluetooth.DeviceData, error) {
	/*
		org.bluez.Device1
			Icon => dbus.Variant{sig:dbus.Signature{str:"s"}, value:"audio-card"}
			LegacyPairing => dbus.Variant{sig:dbus.Signature{str:"b"}, value:false}
			Address => dbus.Variant{sig:dbus.Signature{str:"s"}, value:"2C:41:A1:49:37:CF"}
			Trusted => dbus.Variant{sig:dbus.Signature{str:"b"}, value:false}
			Connected => dbus.Variant{sig:dbus.Signature{str:"b"}, value:true}
			Paired => dbus.Variant{sig:dbus.Signature{str:"b"}, value:true}
			RSSI => dbus.Variant{sig:dbus.Signature{str:"n"}, value:-36}
			Modalias => dbus.Variant{sig:dbus.Signature{str:"s"}, value:"bluetooth:v009Ep4020d0251"}
			Name => dbus.Variant{sig:dbus.Signature{str:"s"}, value:"Bose QC35 II"}
			UUIDs => dbus.Variant{sig:dbus.Signature{str:"as"}, value:[]string{"00000000-deca-fade-deca-deafdecacaff", "00001101-0000-1000-8000-00805f9b34fb", "00001108-0000-1000-8000-00805f9b34fb", "0000110b-0000-1000-8000-00805f9b34fb", "0000110c-0000-1000-8000-00805f9b34fb", "0000110e-0000-1000-8000-00805f9b34fb", "0000111e-0000-1000-8000-00805f9b34fb", "00001200-0000-1000-8000-00805f9b34fb", "81c2e72a-0591-443e-a1ff-05f988593351", "f8d1fbe4-7966-4334-8024-ff96c9330e15"}}
			Adapter => dbus.Variant{sig:dbus.Signature{str:"o"}, value:"/org/bluez/hci0"}
			Blocked => dbus.Variant{sig:dbus.Signature{str:"b"}, value:false}
			Alias => dbus.Variant{sig:dbus.Signature{str:"s"}, value:"Bose QC35 II"}
			Class => dbus.Variant{sig:dbus.Signature{str:"u"}, value:0x240418}

	*/
	device := struct {
		Adapter dbus.ObjectPath
		bluetooth.DeviceData
	}{}

	if err := dbh.DecodeVariantMap(values, &device, "Name", "Address"); err != nil {
		return device.DeviceData, fault.Wrap(err,
			fctx.With(context.Background(),
				"error_at", "device-map-decode",
				"address", d.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Error converting device data"),
		)
	}

	adapterMap, err := d.b.adapterInternal(device.Adapter).adapterProperties()
	if err != nil {
		return device.DeviceData, fault.Wrap(errorkinds.ErrAdapterNotFound,
			fctx.With(context.Background(),
				"error_at", "device-adapter-map",
				"address", d.Address.String(),
			),
			ftag.With(ftag.NotFound),
			fmsg.With("Error parsing device adapter data"),
		)
	}

	addr := adapterMap["Address"]

	adapterMac, err := bluetooth.ParseMAC(addr.Value().(string))
	if err != nil {
		return device.DeviceData, fault.Wrap(errorkinds.ErrPropertyDataParse,
			fctx.With(context.Background(),
				"error_at", "device-adapter-mac",
				"address", d.Address.String(),
			),
			ftag.With(ftag.NotFound),
			fmsg.With("Error parsing device adapter address"),
		)
	}

	device.AssociatedAdapter = adapterMac
	device.Type = bluetooth.DeviceTypeFromClass(device.Class)

	if p, err := d.batteryPercentage(); err == nil {
		device.Percentage = int(p)
	}

	dbh.PathConverter.AddDbusPath(dbh.DbusPathDevice, d.path, device.Address)
	d.b.store.AddDevice(device.DeviceData)

	return device.DeviceData, nil
}

// batteryPercentage gets the battery percentage of a device.
func (d *device) batteryPercentage() (byte, error) {
	var result byte

	if err := d.b.systemBus.Object(dbh.BluezBusName, d.path).
		Call(dbh.DbusGetPropertiesIface, 0, dbh.BluezBatteryIface, "Percentage").
		Store(&result); err != nil {
		return result, err
	}

	return result, nil
}
