package gonetworkmanager

import (
	"encoding/json"

	"github.com/godbus/dbus/v5"
)

const (
	DeviceGenericInterface = DeviceInterface + ".Generic"

	// Properties
	DeviceGenericPropertyHwAddress       = DeviceGenericInterface + ".HwAddress"       // readable   s
	DeviceGenericPropertyTypeDescription = DeviceGenericInterface + ".TypeDescription" // readable   s
)

type DeviceGeneric interface {
	Device

	// GetPropertyHwAddress Active hardware address of the device.
	GetPropertyHwAddress() (string, error)

	// GetPropertyTypeDescription A (non-localized) description of the interface type, if known.
	GetPropertyTypeDescription() (string, error)
}

func NewDeviceGeneric(objectPath dbus.ObjectPath) (DeviceGeneric, error) {
	var d deviceGeneric
	return &d, d.init(NetworkManagerInterface, objectPath)
}

type deviceGeneric struct {
	device
}

func (d *deviceGeneric) GetPropertyHwAddress() (string, error) {
	return d.getStringProperty(DeviceGenericPropertyHwAddress)
}

func (d *deviceGeneric) GetPropertyTypeDescription() (string, error) {
	return d.getStringProperty(DeviceGenericPropertyTypeDescription)
}

func (d *deviceGeneric) MarshalJSON() ([]byte, error) {
	m, err := d.device.marshalMap()
	if err != nil {
		return nil, err
	}

	m["HwAddress"], _ = d.GetPropertyHwAddress()
	m["TypeDescription"], _ = d.GetPropertyTypeDescription()
	return json.Marshal(m)
}
