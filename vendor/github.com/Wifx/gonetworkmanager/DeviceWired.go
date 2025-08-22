package gonetworkmanager

import (
	"encoding/json"

	"github.com/godbus/dbus/v5"
)

const (
	DeviceWiredInterface = DeviceInterface + ".Wired"

	// Properties
	DeviceWiredPropertyHwAddress       = DeviceWiredInterface + ".HwAddress"       // readable   s
	DeviceWiredPropertyPermHwAddress   = DeviceWiredInterface + ".PermHwAddress"   // readable   s
	DeviceWiredPropertySpeed           = DeviceWiredInterface + ".Speed"           // readable   u
	DeviceWiredPropertyS390Subchannels = DeviceWiredInterface + ".S390Subchannels" // readable   as
	DeviceWiredPropertyCarrier         = DeviceWiredInterface + ".Carrier"         // readable   b
)

type DeviceWired interface {
	Device

	// GetPropertyHwAddress Active hardware address of the device.
	GetPropertyHwAddress() (string, error)

	// GetPropertyPermHwAddress Permanent hardware address of the device.
	GetPropertyPermHwAddress() (string, error)

	// GetPropertySpeed Design speed of the device, in megabits/second (Mb/s).
	GetPropertySpeed() (uint32, error)

	// GetPropertyS390Subchannels Array of S/390 subchannels for S/390 or z/Architecture devices.
	GetPropertyS390Subchannels() ([]string, error)

	// GetPropertyCarrier Indicates whether the physical carrier is found (e.g. whether a cable is plugged in or not).
	GetPropertyCarrier() (bool, error)
}

func NewDeviceWired(objectPath dbus.ObjectPath) (DeviceWired, error) {
	var d deviceWired
	return &d, d.init(NetworkManagerInterface, objectPath)
}

type deviceWired struct {
	device
}

func (d *deviceWired) GetPropertyHwAddress() (string, error) {
	return d.getStringProperty(DeviceWiredPropertyHwAddress)
}

func (d *deviceWired) GetPropertyPermHwAddress() (string, error) {
	return d.getStringProperty(DeviceWiredPropertyPermHwAddress)
}

func (d *deviceWired) GetPropertySpeed() (uint32, error) {
	return d.getUint32Property(DeviceWiredPropertySpeed)
}

func (d *deviceWired) GetPropertyS390Subchannels() ([]string, error) {
	return d.getSliceStringProperty(DeviceWiredPropertyS390Subchannels)
}

func (d *deviceWired) GetPropertyCarrier() (bool, error) {
	return d.getBoolProperty(DeviceWiredPropertyCarrier)
}

func (d *deviceWired) MarshalJSON() ([]byte, error) {
	m, err := d.device.marshalMap()
	if err != nil {
		return nil, err
	}

	m["HwAddress"], _ = d.GetPropertyHwAddress()
	m["PermHwAddress"], _ = d.GetPropertyPermHwAddress()
	m["Speed"], _ = d.GetPropertySpeed()
	m["S390Subchannels"], _ = d.GetPropertyS390Subchannels()
	m["Carrier"], _ = d.GetPropertyCarrier()
	return json.Marshal(m)
}
