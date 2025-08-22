package gonetworkmanager

import (
	"encoding/json"

	"github.com/godbus/dbus/v5"
)

const (
	DeviceInterface = NetworkManagerInterface + ".Device"

	/* Methods */
	DeviceReapply              = DeviceInterface + ".Reapply"
	DeviceGetAppliedConnection = DeviceInterface + ".GetAppliedConnection"
	DeviceDisconnect           = DeviceInterface + ".Disconnect"
	DeviceDelete               = DeviceInterface + ".Delete"

	/* Properties */
	DevicePropertyUdi                  = DeviceInterface + ".Udi"                  // readable   s
	DevicePropertyInterface            = DeviceInterface + ".Interface"            // readable   s
	DevicePropertyIpInterface          = DeviceInterface + ".IpInterface"          // readable   s
	DevicePropertyDriver               = DeviceInterface + ".Driver"               // readable   s
	DevicePropertyDriverVersion        = DeviceInterface + ".DriverVersion"        // readable   s
	DevicePropertyFirmwareVersion      = DeviceInterface + ".FirmwareVersion"      // readable   s
	DevicePropertyCapabilities         = DeviceInterface + ".Capabilities"         // readable   u
	DevicePropertyState                = DeviceInterface + ".State"                // readable   u
	DevicePropertyStateReason          = DeviceInterface + ".StateReason"          // readable   (uu)
	DevicePropertyActiveConnection     = DeviceInterface + ".ActiveConnection"     // readable   o
	DevicePropertyIp4Config            = DeviceInterface + ".Ip4Config"            // readable   o
	DevicePropertyDhcp4Config          = DeviceInterface + ".Dhcp4Config"          // readable   o
	DevicePropertyIp6Config            = DeviceInterface + ".Ip6Config"            // readable   o
	DevicePropertyDhcp6Config          = DeviceInterface + ".Dhcp6Config"          // readable   o
	DevicePropertyManaged              = DeviceInterface + ".Managed"              // readwrite  b
	DevicePropertyAutoconnect          = DeviceInterface + ".Autoconnect"          // readwrite  b
	DevicePropertyFirmwareMissing      = DeviceInterface + ".FirmwareMissing"      // readable   b
	DevicePropertyNmPluginMissing      = DeviceInterface + ".NmPluginMissing"      // readable   b
	DevicePropertyDeviceType           = DeviceInterface + ".DeviceType"           // readable   u
	DevicePropertyAvailableConnections = DeviceInterface + ".AvailableConnections" // readable   ao
	DevicePropertyPhysicalPortId       = DeviceInterface + ".PhysicalPortId"       // readable   s
	DevicePropertyMtu                  = DeviceInterface + ".Mtu"                  // readable   u
	DevicePropertyMetered              = DeviceInterface + ".Metered"              // readable   u
	DevicePropertyLldpNeighbors        = DeviceInterface + ".LldpNeighbors"        // readable   aa{sv}
	DevicePropertyReal                 = DeviceInterface + ".Real"                 // readable   b
	DevicePropertyIp4Connectivity      = DeviceInterface + ".Ip4Connectivity"      // readable   u
)

func DeviceFactory(objectPath dbus.ObjectPath) (Device, error) {
	d, err := NewDevice(objectPath)
	if err != nil {
		return nil, err
	}

	deviceType, err := d.GetPropertyDeviceType()
	if err != nil {
		return nil, err
	}

	switch deviceType {
	case NmDeviceTypeDummy:
		return NewDeviceDummy(objectPath)
	case NmDeviceTypeGeneric:
		return NewDeviceGeneric(objectPath)
	case NmDeviceTypeIpTunnel:
		return NewDeviceIpTunnel(objectPath)
	case NmDeviceTypeEthernet:
		return NewDeviceWired(objectPath)
	case NmDeviceTypeWifi:
		return NewDeviceWireless(objectPath)
	}

	return d, nil
}

type Device interface {
	GetPath() dbus.ObjectPath

	// Reapply Attempts to update the configuration of a device without deactivating it. NetworkManager has the concept of connections, which are profiles that contain the configuration for a networking device. Those connections are exposed via D-Bus as individual objects that can be created, modified and deleted. When activating such a settings-connection on a device, the settings-connection is cloned to become an applied-connection and used to configure the device (see GetAppliedConnection). Subsequent modification of the settings-connection don't propagate automatically to the device's applied-connection (with exception of the firewall-zone and the metered property). For the changes to take effect, you can either re-activate the settings-connection, or call Reapply. The Reapply call allows you to directly update the applied-connection and reconfigure the device. Reapply can also be useful if the currently applied-connection is equal to the connection that is about to be reapplied. This allows to reconfigure the device and revert external changes like removing or adding an IP address (which NetworkManager doesn't revert automatically because it is assumed that the user made these changes intentionally outside of NetworkManager). Reapply can make the applied-connection different from the settings-connection, just like updating the settings-connection can make them different.
	// connection: The optional connection settings that will be reapplied on the device. If empty, the currently active settings-connection will be used. The connection cannot arbitrarly differ from the current applied-connection otherwise the call will fail. Only certain changes are supported, like adding or removing IP addresses.
	// versionId: If non-zero, the current version id of the applied-connection must match. The current version id can be retrieved via GetAppliedConnection. This optional argument allows to catch concurrent modifications between the GetAppliedConnection call and Reapply.
	// flags: Flags which would modify the behavior of the Reapply call. There are no flags defined currently and the users should use the value of 0.
	Reapply(connection Connection, versionId uint64, flags uint32) error

	// Disconnect a device and prevents the device from automatically activating further connections without user intervention.
	Disconnect() error

	// Delete a software device from NetworkManager and removes the interface from the system. The method returns an error when called for a hardware device.
	Delete() error

	// GetPropertyUdi Operating-system specific transient device hardware identifier. This is an opaque string representing the underlying hardware for the device, and shouldn't be used to keep track of individual devices. For some device types (Bluetooth, Modems) it is an identifier used by the hardware service (ie bluez or ModemManager) to refer to that device, and client programs use it get additional information from those services which NM does not provide. The Udi is not guaranteed to be consistent across reboots or hotplugs of the hardware. If you're looking for a way to uniquely track each device in your application, use the object path. If you're looking for a way to track a specific piece of hardware across reboot or hotplug, use a MAC address or USB serial number.
	GetPropertyUdi() (string, error)

	// GetPropertyInterface The name of the device's control (and often data) interface. Note that non UTF-8 characters are backslash escaped, so the resulting name may be longer then 15 characters. Use g_strcompress() to revert the escaping.
	GetPropertyInterface() (string, error)

	// GetPropertyIpInterface The name of the device's data interface when available. This property may not refer to the actual data interface until the device has successfully established a data connection, indicated by the device's State becoming ACTIVATED. Note that non UTF-8 characters are backslash escaped, so the resulting name may be longer then 15 characters. Use g_strcompress() to revert the escaping.
	GetPropertyIpInterface() (string, error)

	// GetPropertyDriver The driver handling the device. Non-UTF-8 sequences are backslash escaped. Use g_strcompress() to revert.
	GetPropertyDriver() (string, error)

	// GetPropertyDriverVersion The version of the driver handling the device. Non-UTF-8 sequences are backslash escaped. Use g_strcompress() to revert.
	GetPropertyDriverVersion() (string, error)

	// GetPropertyFirmwareVersion The firmware version for the device. Non-UTF-8 sequences are backslash escaped. Use g_strcompress() to revert.
	GetPropertyFirmwareVersion() (string, error)

	// GetPropertyState The current state of the device.
	GetPropertyState() (NmDeviceState, error)

	// GetPropertyActiveConnection Object path of an ActiveConnection object that "owns" this device during activation. The ActiveConnection object tracks the life-cycle of a connection to a specific network and implements the org.freedesktop.NetworkManager.Connection.Active D-Bus interface.
	GetPropertyActiveConnection() (ActiveConnection, error)

	// GetPropertyIP4Config Object path of the Ip4Config object describing the configuration of the device. Only valid when the device is in the NM_DEVICE_STATE_ACTIVATED state.
	GetPropertyIP4Config() (IP4Config, error)

	// GetPropertyDHCP4Config Object path of the Dhcp4Config object describing the DHCP options returned by the DHCP server. Only valid when the device is in the NM_DEVICE_STATE_ACTIVATED state.
	GetPropertyDHCP4Config() (DHCP4Config, error)

	// GetPropertyIP6Config Object path of the Ip6Config object describing the configuration of the device. Only valid when the device is in the NM_DEVICE_STATE_ACTIVATED state.
	GetPropertyIP6Config() (IP6Config, error)

	// GetPropertyDHCP6Config Object path of the Dhcp6Config object describing the DHCP options returned by the DHCP server. Only valid when the device is in the NM_DEVICE_STATE_ACTIVATED state.
	GetPropertyDHCP6Config() (DHCP6Config, error)

	// GetPropertyManaged Whether this device is managed by NetworkManager. Setting this property has a similar effect to configuring the device as unmanaged via the keyfile.unmanaged-devices setting in NetworkManager.conf. Changes to this value are not persistent and lost after NetworkManager restart.
	GetPropertyManaged() (bool, error)
	SetPropertyManaged(bool) error

	// GetPropertyAutoConnect If TRUE, indicates the device is allowed to autoconnect. If FALSE, manual intervention is required before the device will automatically connect to a known network, such as activating a connection using the device, or setting this property to TRUE. This property cannot be set to TRUE for default-unmanaged devices, since they never autoconnect.
	GetPropertyAutoConnect() (bool, error)

	// GetPropertyFirmwareMissing If TRUE, indicates the device is likely missing firmware necessary for its operation.
	GetPropertyFirmwareMissing() (bool, error)

	// GetPropertyNmPluginMissing If TRUE, indicates the NetworkManager plugin for the device is likely missing or misconfigured.
	GetPropertyNmPluginMissing() (bool, error)

	// GetPropertyDeviceType The general type of the network device; ie Ethernet, Wi-Fi, etc.
	GetPropertyDeviceType() (NmDeviceType, error)

	// GetPropertyAvailableConnections An array of object paths of every configured connection that is currently 'available' through this device.
	GetPropertyAvailableConnections() ([]Connection, error)

	// GetPropertyPhysicalPortId If non-empty, an (opaque) indicator of the physical network port associated with the device. This can be used to recognize when two seemingly-separate hardware devices are actually just different virtual interfaces to the same physical port.
	GetPropertyPhysicalPortId() (string, error)

	// GetPropertyMtu The device MTU (maximum transmission unit).
	GetPropertyMtu() (uint32, error)

	// GetPropertyReal True if the device exists, or False for placeholder devices that do not yet exist but could be automatically created by NetworkManager if one of their AvailableConnections was activated.
	GetPropertyReal() (bool, error)

	// The result of the last IPv4 connectivity check.
	GetPropertyIp4Connectivity() (NmConnectivity, error)

	MarshalJSON() ([]byte, error)
}

func NewDevice(objectPath dbus.ObjectPath) (Device, error) {
	var d device
	return &d, d.init(NetworkManagerInterface, objectPath)
}

type device struct {
	dbusBase
}

func (d *device) GetPath() dbus.ObjectPath {
	return d.obj.Path()
}

func (d *device) Reapply(connection Connection, versionId uint64, flags uint32) error {
	return d.call(DeviceReapply, connection, versionId, flags)
}

func (d *device) Disconnect() error {
	return d.call(DeviceDisconnect)
}

func (d *device) Delete() error {
	return d.call(DeviceDelete)
}

func (d *device) GetPropertyUdi() (string, error) {
	return d.getStringProperty(DevicePropertyUdi)
}

func (d *device) GetPropertyInterface() (string, error) {
	return d.getStringProperty(DevicePropertyInterface)
}

func (d *device) GetPropertyIpInterface() (string, error) {
	return d.getStringProperty(DevicePropertyIpInterface)
}

func (d *device) GetPropertyDriver() (string, error) {
	return d.getStringProperty(DevicePropertyDriver)
}

func (d *device) GetPropertyDriverVersion() (string, error) {
	return d.getStringProperty(DevicePropertyDriverVersion)
}

func (d *device) GetPropertyFirmwareVersion() (string, error) {
	return d.getStringProperty(DevicePropertyFirmwareVersion)
}

func (d *device) GetPropertyState() (NmDeviceState, error) {
	r, err := d.getUint32Property(DevicePropertyState)
	if err != nil {
		return NmDeviceStateFailed, err
	}
	return NmDeviceState(r), nil
}

func (d *device) GetPropertyActiveConnection() (ActiveConnection, error) {
	path, err := d.getObjectProperty(DevicePropertyActiveConnection)
	if err != nil || path == "/" {
		return nil, err
	}

	return NewActiveConnection(path)
}

func (d *device) GetPropertyIP4Config() (IP4Config, error) {
	path, err := d.getObjectProperty(DevicePropertyIp4Config)
	if err != nil || path == "/" {
		return nil, err
	}

	return NewIP4Config(path)
}

func (d *device) GetPropertyDHCP4Config() (DHCP4Config, error) {
	path, err := d.getObjectProperty(DevicePropertyDhcp4Config)
	if err != nil || path == "/" {
		return nil, err
	}

	return NewDHCP4Config(path)
}

func (d *device) GetPropertyIP6Config() (IP6Config, error) {
	path, err := d.getObjectProperty(DevicePropertyIp6Config)
	if err != nil || path == "/" {
		return nil, err
	}

	return NewIP6Config(path)
}

func (d *device) GetPropertyDHCP6Config() (DHCP6Config, error) {
	path, err := d.getObjectProperty(DevicePropertyDhcp6Config)
	if err != nil || path == "/" {
		return nil, err
	}

	return NewDHCP6Config(path)
}

func (d *device) GetPropertyManaged() (bool, error) {
	return d.getBoolProperty(DevicePropertyManaged)
}

func (d *device) SetPropertyManaged(managed bool) error {
	return d.setProperty(DevicePropertyManaged, managed)
}

func (d *device) GetPropertyAutoConnect() (bool, error) {
	return d.getBoolProperty(DevicePropertyAutoconnect)
}

func (d *device) GetPropertyFirmwareMissing() (bool, error) {
	return d.getBoolProperty(DevicePropertyFirmwareMissing)
}

func (d *device) GetPropertyNmPluginMissing() (bool, error) {
	return d.getBoolProperty(DevicePropertyNmPluginMissing)
}

func (d *device) GetPropertyDeviceType() (NmDeviceType, error) {
	v, err := d.getUint32Property(DevicePropertyDeviceType)
	return NmDeviceType(v), err
}

func (d *device) GetPropertyAvailableConnections() ([]Connection, error) {
	connPaths, err := d.getSliceObjectProperty(DevicePropertyAvailableConnections)
	if err != nil {
		return nil, err
	}

	conns := make([]Connection, len(connPaths))
	for i, path := range connPaths {
		conns[i], err = NewConnection(path)
		if err != nil {
			return conns, err
		}
	}

	return conns, nil
}

func (d *device) GetPropertyPhysicalPortId() (string, error) {
	return d.getStringProperty(DevicePropertyPhysicalPortId)
}

func (d *device) GetPropertyMtu() (uint32, error) {
	return d.getUint32Property(DevicePropertyMtu)
}

func (d *device) GetPropertyReal() (bool, error) {
	return d.getBoolProperty(DevicePropertyReal)
}

func (d *device) GetPropertyIp4Connectivity() (NmConnectivity, error) {
	u, err := d.getUint32Property(DevicePropertyIp4Connectivity)
	return NmConnectivity(u), err
}

func (d *device) marshalMap() (map[string]interface{}, error) {
	Interface, err := d.GetPropertyInterface()
	if err != nil {
		return nil, err
	}
	IpInterface, err := d.GetPropertyIpInterface()
	if err != nil {
		return nil, err
	}
	State, err := d.GetPropertyState()
	if err != nil {
		return nil, err
	}
	IP4Config, err := d.GetPropertyIP4Config()
	if err != nil {
		return nil, err
	}
	DHCP4Config, err := d.GetPropertyDHCP4Config()
	if err != nil {
		return nil, err
	}
	DeviceType, err := d.GetPropertyDeviceType()
	if err != nil {
		return nil, err
	}
	AvailableConnections, err := d.GetPropertyAvailableConnections()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"Interface":            Interface,
		"IP interface":         IpInterface,
		"State":                State.String(),
		"IP4Config":            IP4Config,
		"DHCP4Config":          DHCP4Config,
		"DeviceType":           DeviceType.String(),
		"AvailableConnections": AvailableConnections,
	}, nil
}

func (d *device) MarshalJSON() ([]byte, error) {
	m, err := d.marshalMap()
	if err != nil {
		return nil, err
	}
	return json.Marshal(m)
}
