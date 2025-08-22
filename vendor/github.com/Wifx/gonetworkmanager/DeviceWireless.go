package gonetworkmanager

import (
	"encoding/json"

	"github.com/godbus/dbus/v5"
)

const (
	DeviceWirelessInterface = DeviceInterface + ".Wireless"

	// Methods
	DeviceWirelessGetAccessPoints    = DeviceWirelessInterface + ".GetAccessPoints"
	DeviceWirelessGetAllAccessPoints = DeviceWirelessInterface + ".GetAllAccessPoints"
	DeviceWirelessRequestScan        = DeviceWirelessInterface + ".RequestScan"

	// Properties
	DeviceWirelessPropertyHwAddress            = DeviceWirelessInterface + ".HwAddress"            // readable   s
	DeviceWirelessPropertyPermHwAddress        = DeviceWirelessInterface + ".PermHwAddress"        // readable   s
	DeviceWirelessPropertyMode                 = DeviceWirelessInterface + ".Mode"                 // readable   u
	DeviceWirelessPropertyBitrate              = DeviceWirelessInterface + ".Bitrate"              // readable   u
	DeviceWirelessPropertyAccessPoints         = DeviceWirelessInterface + ".AccessPoints"         // readable   ao
	DeviceWirelessPropertyActiveAccessPoint    = DeviceWirelessInterface + ".ActiveAccessPoint"    // readable   o
	DeviceWirelessPropertyWirelessCapabilities = DeviceWirelessInterface + ".WirelessCapabilities" // readable   u
	DeviceWirelessPropertyLastScan             = DeviceWirelessInterface + ".LastScan"             // readable   x
)

type DeviceWireless interface {
	Device

	// GetAccessPoints gets the list of access points visible to this device.
	// Note that this list does not include access points which hide their SSID.
	// To retrieve a list of all access points (including hidden ones) use the
	// GetAllAccessPoints() method.
	GetAccessPoints() ([]AccessPoint, error)

	// GetAllAccessPoints gets the list of all access points visible to this
	// device, including hidden ones for which the SSID is not yet known.
	GetAllAccessPoints() ([]AccessPoint, error)

	// RequestScan Request the device to scan. To know when the scan is finished, use the
	// "PropertiesChanged" signal from "org.freedesktop.DBus.Properties" to listen
	// to changes to the "LastScan" property.
	RequestScan() error

	// GetPropertyHwAddress The active hardware address of the device.
	GetPropertyHwAddress() (string, error)

	// GetPropertyPermHwAddress The permanent hardware address of the device.
	GetPropertyPermHwAddress() (string, error)

	// GetPropertyMode The operating mode of the wireless device.
	GetPropertyMode() (Nm80211Mode, error)

	// GetPropertyBitrate The bit rate currently used by the wireless device, in kilobits/second (Kb/s).
	GetPropertyBitrate() (uint32, error)

	// GetPropertyAccessPoints List of object paths of access point visible to this wireless device.
	GetPropertyAccessPoints() ([]AccessPoint, error)

	// GetPropertyActiveAccessPoint Object path of the access point currently used by the wireless device.
	GetPropertyActiveAccessPoint() (AccessPoint, error)

	// GetPropertyWirelessCapabilities The capabilities of the wireless device.
	GetPropertyWirelessCapabilities() (uint32, error)

	// GetPropertyLastScan The timestamp (in CLOCK_BOOTTIME milliseconds) for the last finished
	// network scan. A value of -1 means the device never scanned for access
	// points.
	GetPropertyLastScan() (int64, error)
}

func NewDeviceWireless(objectPath dbus.ObjectPath) (DeviceWireless, error) {
	var d deviceWireless
	return &d, d.init(NetworkManagerInterface, objectPath)
}

type deviceWireless struct {
	device
}

func (d *deviceWireless) GetAccessPoints() ([]AccessPoint, error) {
	var apPaths []dbus.ObjectPath
	err := d.callWithReturn(&apPaths, DeviceWirelessGetAccessPoints)

	if err != nil {
		return nil, err
	}

	aps := make([]AccessPoint, len(apPaths))

	for i, path := range apPaths {
		aps[i], err = NewAccessPoint(path)
		if err != nil {
			return aps, err
		}
	}

	return aps, nil
}

func (d *deviceWireless) GetAllAccessPoints() ([]AccessPoint, error) {
	var apPaths []dbus.ObjectPath
	err := d.callWithReturn(&apPaths, DeviceWirelessGetAllAccessPoints)

	if err != nil {
		return nil, err
	}

	aps := make([]AccessPoint, len(apPaths))

	for i, path := range apPaths {
		aps[i], err = NewAccessPoint(path)
		if err != nil {
			return aps, err
		}
	}

	return aps, nil
}

func (d *deviceWireless) RequestScan() error {
	var options map[string]interface{}
	return d.obj.Call(DeviceWirelessRequestScan, 0, options).Store()
}

func (d *deviceWireless) GetPropertyHwAddress() (string, error) {
	return d.getStringProperty(DeviceWirelessPropertyHwAddress)
}

func (d *deviceWireless) GetPropertyPermHwAddress() (string, error) {
	return d.getStringProperty(DeviceWirelessPropertyPermHwAddress)
}

func (d *deviceWireless) GetPropertyMode() (Nm80211Mode, error) {
	v, err := d.getUint32Property(DeviceWirelessPropertyMode)
	return Nm80211Mode(v), err
}

func (d *deviceWireless) GetPropertyBitrate() (uint32, error) {
	return d.getUint32Property(DeviceWirelessPropertyBitrate)
}

func (d *deviceWireless) GetPropertyAccessPoints() ([]AccessPoint, error) {
	apPaths, err := d.getSliceObjectProperty(DeviceWirelessPropertyAccessPoints)
	if err != nil {
		return nil, err
	}

	ap := make([]AccessPoint, len(apPaths))
	for i, path := range apPaths {
		ap[i], err = NewAccessPoint(path)
		if err != nil {
			return ap, err
		}
	}

	return ap, nil
}

func (d *deviceWireless) GetPropertyActiveAccessPoint() (AccessPoint, error) {
	path, err := d.getObjectProperty(DeviceWirelessPropertyActiveAccessPoint)
	if err != nil || path == "/" {
		return nil, err
	}

	return NewAccessPoint(path)
}

func (d *deviceWireless) GetPropertyWirelessCapabilities() (uint32, error) {
	return d.getUint32Property(DeviceWirelessPropertyWirelessCapabilities)
}

func (d *deviceWireless) GetPropertyLastScan() (int64, error) {
	return d.getInt64Property(DeviceWirelessPropertyLastScan)
}

func (d *deviceWireless) MarshalJSON() ([]byte, error) {
	m, err := d.device.marshalMap()
	if err != nil {
		return nil, err
	}

	m["AccessPoints"], _ = d.GetPropertyAccessPoints()
	m["HwAddress"], _ = d.GetPropertyHwAddress()
	m["PermHwAddress"], _ = d.GetPropertyPermHwAddress()
	m["Mode"], _ = d.GetPropertyMode()
	m["Bitrate"], _ = d.GetPropertyBitrate()
	m["ActiveAccessPoint"], _ = d.GetPropertyActiveAccessPoint()
	m["WirelessCapabilities"], _ = d.GetPropertyWirelessCapabilities()
	m["LastScan"], _ = d.GetPropertyLastScan()
	return json.Marshal(m)
}
