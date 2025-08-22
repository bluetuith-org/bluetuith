package gonetworkmanager

import (
	"encoding/json"

	"github.com/godbus/dbus/v5"
)

const (
	DeviceDummyInterface = DeviceInterface + ".Dummy"

	/* Properties */
	DeviceDummyPropertyHwAddress = DeviceDummyInterface + ".HwAddress" // readable   s
)

type DeviceDummy interface {
	Device

	// Hardware address of the device.
	GetPropertyHwAddress() (string, error)
}

func NewDeviceDummy(objectPath dbus.ObjectPath) (DeviceDummy, error) {
	var d deviceDummy
	return &d, d.init(NetworkManagerInterface, objectPath)
}

type deviceDummy struct {
	device
}

func (d *deviceDummy) GetPropertyHwAddress() (string, error) {
	return d.getStringProperty(DeviceDummyPropertyHwAddress)
}

func (d *deviceDummy) MarshalJSON() ([]byte, error) {
	m, err := d.device.marshalMap()
	if err != nil {
		return nil, err
	}

	m["HwAddress"], _ = d.GetPropertyHwAddress()
	return json.Marshal(m)
}
