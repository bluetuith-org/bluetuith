package gonetworkmanager

import (
	"fmt"

	"github.com/godbus/dbus/v5"
)

const (
	ActiveConnectionInterface = NetworkManagerInterface + ".Connection.Active"

	/* Properties */
	ActiveConnectionPropertyConnection     = ActiveConnectionInterface + ".Connection"     // readable   o
	ActiveConnectionPropertySpecificObject = ActiveConnectionInterface + ".SpecificObject" // readable   o
	ActiveConnectionPropertyId             = ActiveConnectionInterface + ".Id"             // readable   s
	ActiveConnectionPropertyUuid           = ActiveConnectionInterface + ".Uuid"           // readable   s
	ActiveConnectionPropertyType           = ActiveConnectionInterface + ".Type"           // readable   s
	ActiveConnectionPropertyDevices        = ActiveConnectionInterface + ".Devices"        // readable   ao
	ActiveConnectionPropertyState          = ActiveConnectionInterface + ".State"          // readable   u
	ActiveConnectionPropertyStateFlags     = ActiveConnectionInterface + ".StateFlags"     // readable   u
	ActiveConnectionPropertyDefault        = ActiveConnectionInterface + ".Default"        // readable   b
	ActiveConnectionPropertyIp4Config      = ActiveConnectionInterface + ".Ip4Config"      // readable   o
	ActiveConnectionPropertyDhcp4Config    = ActiveConnectionInterface + ".Dhcp4Config"    // readable   o
	ActiveConnectionPropertyDefault6       = ActiveConnectionInterface + ".Default6"       // readable   b
	ActiveConnectionPropertyIp6Config      = ActiveConnectionInterface + ".Ip6Config"      // readable   o
	ActiveConnectionPropertyDhcp6Config    = ActiveConnectionInterface + ".Dhcp6Config"    // readable   o
	ActiveConnectionPropertyVpn            = ActiveConnectionInterface + ".Vpn"            // readable   b
	ActiveConnectionPropertyMaster         = ActiveConnectionInterface + ".Master"         // readable   o

	/* Signals */
	ActiveConnectionSignalStateChanged = "StateChanged" // u state, u reason

)

type ActiveConnection interface {
	GetPath() dbus.ObjectPath

	// GetPropertyConnection gets connection object of the connection.
	GetPropertyConnection() (Connection, error)

	// GetPropertySpecificObject gets a specific object associated with the active connection.
	GetPropertySpecificObject() (AccessPoint, error)

	// GetPropertyID gets the ID of the connection.
	GetPropertyID() (string, error)

	// GetPropertyUUID gets the UUID of the connection.
	GetPropertyUUID() (string, error)

	// GetPropertyType gets the type of the connection.
	GetPropertyType() (string, error)

	// GetPropertyDevices gets array of device objects which are part of this active connection.
	GetPropertyDevices() ([]Device, error)

	// GetPropertyState gets the state of the connection.
	GetPropertyState() (NmActiveConnectionState, error)

	// GetPropertyStateFlags gets the state flags of the connection.
	GetPropertyStateFlags() (uint32, error)

	// GetPropertyDefault gets the default IPv4 flag of the connection.
	GetPropertyDefault() (bool, error)

	// GetPropertyIP4Config gets the IP4Config of the connection.
	GetPropertyIP4Config() (IP4Config, error)

	// GetPropertyDHCP4Config gets the DHCP6Config of the connection.
	GetPropertyDHCP4Config() (DHCP4Config, error)

	// GetPropertyDefault6 gets the default IPv6 flag of the connection.
	GetPropertyDefault6() (bool, error)

	// GetPropertyIP6Config gets the IP6Config of the connection.
	GetPropertyIP6Config() (IP6Config, error)

	// GetPropertyDHCP6Config gets the DHCP4Config of the connection.
	GetPropertyDHCP6Config() (DHCP6Config, error)

	// GetPropertyVPN gets the VPN flag of the connection.
	GetPropertyVPN() (bool, error)

	// GetPropertyMaster gets the master device of the connection.
	GetPropertyMaster() (Device, error)

	SubscribeState(receiver chan StateChange, exit chan struct{}) (err error)
}

func NewActiveConnection(objectPath dbus.ObjectPath) (ActiveConnection, error) {
	var a activeConnection
	return &a, a.init(NetworkManagerInterface, objectPath)
}

type activeConnection struct {
	dbusBase
}

func (a *activeConnection) GetPath() dbus.ObjectPath {
	return a.obj.Path()
}

func (a *activeConnection) GetPropertyConnection() (Connection, error) {
	path, err := a.getObjectProperty(ActiveConnectionPropertyConnection)
	if err != nil {
		return nil, err
	}
	con, err := NewConnection(path)
	if err != nil {
		return nil, err
	}
	return con, nil
}

func (a *activeConnection) GetPropertySpecificObject() (AccessPoint, error) {
	path, err := a.getObjectProperty(ActiveConnectionPropertySpecificObject)
	if err != nil {
		return nil, err
	}
	ap, err := NewAccessPoint(path)
	if err != nil {
		return nil, err
	}
	return ap, nil
}

func (a *activeConnection) GetPropertyID() (string, error) {
	return a.getStringProperty(ActiveConnectionPropertyId)
}

func (a *activeConnection) GetPropertyUUID() (string, error) {
	return a.getStringProperty(ActiveConnectionPropertyUuid)
}

func (a *activeConnection) GetPropertyType() (string, error) {
	return a.getStringProperty(ActiveConnectionPropertyType)
}

func (a *activeConnection) GetPropertyDevices() ([]Device, error) {
	paths, err := a.getSliceObjectProperty(ActiveConnectionPropertyDevices)
	if err != nil {
		return nil, err
	}
	devices := make([]Device, len(paths))
	for i, path := range paths {
		devices[i], err = DeviceFactory(path)
		if err != nil {
			return nil, err
		}
	}
	return devices, nil
}
func (a *activeConnection) GetPropertyState() (NmActiveConnectionState, error) {
	v, err := a.getUint32Property(ActiveConnectionPropertyState)
	return NmActiveConnectionState(v), err
}

func (a *activeConnection) GetPropertyStateFlags() (uint32, error) {
	return a.getUint32Property(ActiveConnectionPropertyStateFlags)
}

func (a *activeConnection) GetPropertyDefault() (bool, error) {
	b, err := a.getProperty(ActiveConnectionPropertyDefault)
	if err != nil {
		return false, err
	}
	return b.(bool), nil
}

func (a *activeConnection) GetPropertyIP4Config() (IP4Config, error) {
	path, err := a.getObjectProperty(ActiveConnectionPropertyIp4Config)
	if err != nil || path == "/" {
		return nil, err
	}
	return NewIP4Config(path)
}

func (a *activeConnection) GetPropertyDHCP4Config() (DHCP4Config, error) {
	path, err := a.getObjectProperty(ActiveConnectionPropertyDhcp4Config)
	if err != nil || path == "/" {
		return nil, err
	}
	return NewDHCP4Config(path)
}

func (a *activeConnection) GetPropertyDefault6() (bool, error) {
	return a.getBoolProperty(ActiveConnectionPropertyDefault6)
}

func (a *activeConnection) GetPropertyIP6Config() (IP6Config, error) {
	path, err := a.getObjectProperty(ActiveConnectionPropertyIp6Config)
	if err != nil || path == "/" {
		return nil, err
	}

	return NewIP6Config(path)
}

func (a *activeConnection) GetPropertyDHCP6Config() (DHCP6Config, error) {
	path, err := a.getObjectProperty(ActiveConnectionPropertyDhcp6Config)
	if err != nil || path == "/" {
		return nil, err
	}

	return NewDHCP6Config(path)
}

func (a *activeConnection) GetPropertyVPN() (bool, error) {
	ret, err := a.getProperty(ActiveConnectionPropertyVpn)
	if err != nil {
		return false, err
	}
	return ret.(bool), nil
}

func (a *activeConnection) GetPropertyMaster() (Device, error) {
	path, err := a.getObjectProperty(ActiveConnectionPropertyMaster)
	if err != nil || path == "/" {
		return nil, err
	}
	return DeviceFactory(path)
}

type StateChange struct {
	Path   dbus.ObjectPath
	State  NmActiveConnectionState
	Reason NmActiveConnectionStateReason
}

func (a *activeConnection) SubscribeState(receiver chan StateChange, exit chan struct{}) (err error) {

	channel := make(chan *dbus.Signal, 1)

	a.conn.Signal(channel)

	err = a.conn.AddMatchSignal(
		dbus.WithMatchInterface(ActiveConnectionInterface),
		dbus.WithMatchMember(ActiveConnectionSignalStateChanged),
		dbus.WithMatchObjectPath(a.GetPath()),
	)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case signal, ok := <-channel:
				if !ok {
					err = fmt.Errorf("connection closed for %s", ActiveConnectionSignalStateChanged)
					return
				}

				if signal.Name != ActiveConnectionInterface+"."+ActiveConnectionSignalStateChanged {
					continue
				}

				stateChange := StateChange{
					Path:   signal.Path,
					State:  NmActiveConnectionState(signal.Body[0].(uint32)),
					Reason: NmActiveConnectionStateReason(signal.Body[1].(uint32)),
				}

				receiver <- stateChange

			case <-exit:
				a.conn.RemoveSignal(channel)
				close(channel)
				return
			}
		}
	}()

	return
}
