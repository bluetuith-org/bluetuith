package gonetworkmanager

import (
	"github.com/godbus/dbus/v5"
)

const (
	VpnConnectionInterface = NetworkManagerInterface + ".VPN.Connection"

	/* Properties */
	VpnConnectionPropertyVpnState = VpnConnectionInterface + ".VpnState" // readable   u
	VpnConnectionPropertyBanner   = VpnConnectionInterface + ".Banner"   // readable   s
)

type VpnConnection interface {
	GetPath() dbus.ObjectPath

	// GetPropertyVpnState The VPN-specific state of the connection.
	GetPropertyVpnState() (NmVpnConnectionState, error)

	// GetPropertyBanner The banner string of the VPN connection.
	GetPropertyBanner() (string, error)
}

func NewVpnConnection(objectPath dbus.ObjectPath) (VpnConnection, error) {
	var a vpnConnection
	return &a, a.init(NetworkManagerInterface, objectPath)
}

type vpnConnection struct {
	dbusBase
}

func (a *vpnConnection) GetPath() dbus.ObjectPath {
	return a.obj.Path()
}

func (a *vpnConnection) GetPropertyVpnState() (NmVpnConnectionState, error) {
	v, err := a.getUint32Property(VpnConnectionPropertyVpnState)
	return NmVpnConnectionState(v), err
}

func (a *vpnConnection) GetPropertyBanner() (string, error) {
	return a.getStringProperty(VpnConnectionPropertyBanner)
}
