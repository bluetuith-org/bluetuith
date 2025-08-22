package gonetworkmanager

import (
	"encoding/json"

	"github.com/godbus/dbus/v5"
)

const (
	DeviceIpTunnelInterface = DeviceInterface + ".IPTunnel"

	// Properties
	DeviceIpTunnelPropertyHwAddress          = DeviceIpTunnelInterface + "HwAddress"           // readable   s
	DeviceIpTunnelPropertyMode               = DeviceIpTunnelInterface + ".Mode"               // readable   u
	DeviceIpTunnelPropertyParent             = DeviceIpTunnelInterface + ".Parent"             // readable   o
	DeviceIpTunnelPropertyLocal              = DeviceIpTunnelInterface + ".Local"              // readable   s
	DeviceIpTunnelPropertyRemote             = DeviceIpTunnelInterface + ".Remote"             // readable   s
	DeviceIpTunnelPropertyTtl                = DeviceIpTunnelInterface + ".Ttl"                // readable   y
	DeviceIpTunnelPropertyTos                = DeviceIpTunnelInterface + ".Tos"                // readable   y
	DeviceIpTunnelPropertyPathMtuDiscovery   = DeviceIpTunnelInterface + ".PathMtuDiscovery"   // readable   b
	DeviceIpTunnelPropertyInputKey           = DeviceIpTunnelInterface + ".InputKey"           // readable   s
	DeviceIpTunnelPropertyOutputKey          = DeviceIpTunnelInterface + ".OutputKey"          // readable   s
	DeviceIpTunnelPropertyEncapsulationLimit = DeviceIpTunnelInterface + ".EncapsulationLimit" // readable   y
	DeviceIpTunnelPropertyFlowLabel          = DeviceIpTunnelInterface + ".FlowLabel"          // readable   u
	DeviceIpTunnelPropertyFlags              = DeviceIpTunnelInterface + ".Flags"              // readable   u

)

type DeviceIpTunnel interface {
	Device

	// GetPropertyMode The tunneling mode
	GetPropertyMode() (uint32, error)

	// GetPropertyParent The object path of the parent device.
	GetPropertyParent() (Device, error)

	// GetPropertyLocal The local endpoint of the tunnel.
	GetPropertyLocal() (string, error)

	// GetPropertyRemote The remote endpoint of the tunnel.
	GetPropertyRemote() (string, error)

	// GetPropertyTtl The TTL assigned to tunneled packets. 0 is a special value meaning that packets inherit the TTL value
	GetPropertyTtl() (uint8, error)

	// GetPropertyTos The type of service (IPv4) or traffic class (IPv6) assigned to tunneled packets.
	GetPropertyTos() (uint8, error)

	// GetPropertyPathMtuDiscovery Whether path MTU discovery is enabled on this tunnel.
	GetPropertyPathMtuDiscovery() (bool, error)

	// GetPropertyInputKey The key used for incoming packets.
	GetPropertyInputKey() (string, error)

	// GetPropertyOutputKey The key used for outgoing packets.
	GetPropertyOutputKey() (string, error)

	// GetPropertyEncapsulationLimit How many additional levels of encapsulation are permitted to be prepended to packets. This property applies only to IPv6 tunnels.
	GetPropertyEncapsulationLimit() (uint8, error)

	// GetPropertyFlowLabel The flow label to assign to tunnel packets. This property applies only to IPv6 tunnels.
	GetPropertyFlowLabel() (uint32, error)

	// GetPropertyFlags Tunnel flags.
	GetPropertyFlags() (uint32, error)
}

func NewDeviceIpTunnel(objectPath dbus.ObjectPath) (DeviceIpTunnel, error) {
	var d deviceIpTunnel
	return &d, d.init(NetworkManagerInterface, objectPath)
}

type deviceIpTunnel struct {
	device
}

func (d *deviceIpTunnel) GetPropertyMode() (uint32, error) {
	return d.getUint32Property(DeviceIpTunnelPropertyMode)
}

func (d *deviceIpTunnel) GetPropertyParent() (Device, error) {
	path, err := d.getObjectProperty(DeviceIpTunnelPropertyParent)
	if err != nil || path == "/" {
		return nil, err
	}

	return DeviceFactory(path)
}

func (d *deviceIpTunnel) GetPropertyLocal() (string, error) {
	return d.getStringProperty(DeviceIpTunnelPropertyLocal)
}

func (d *deviceIpTunnel) GetPropertyRemote() (string, error) {
	return d.getStringProperty(DeviceIpTunnelPropertyRemote)
}

func (d *deviceIpTunnel) GetPropertyTtl() (uint8, error) {
	return d.getUint8Property(DeviceIpTunnelPropertyTtl)
}

func (d *deviceIpTunnel) GetPropertyTos() (uint8, error) {
	return d.getUint8Property(DeviceIpTunnelPropertyTos)
}

func (d *deviceIpTunnel) GetPropertyPathMtuDiscovery() (bool, error) {
	return d.getBoolProperty(DeviceIpTunnelPropertyPathMtuDiscovery)
}

func (d *deviceIpTunnel) GetPropertyInputKey() (string, error) {
	return d.getStringProperty(DeviceIpTunnelPropertyInputKey)
}

func (d *deviceIpTunnel) GetPropertyOutputKey() (string, error) {
	return d.getStringProperty(DeviceIpTunnelPropertyOutputKey)
}

func (d *deviceIpTunnel) GetPropertyEncapsulationLimit() (uint8, error) {
	return d.getUint8Property(DeviceIpTunnelPropertyEncapsulationLimit)
}

func (d *deviceIpTunnel) GetPropertyFlowLabel() (uint32, error) {
	return d.getUint32Property(DeviceIpTunnelPropertyFlowLabel)
}

func (d *deviceIpTunnel) GetPropertyFlags() (uint32, error) {
	return d.getUint32Property(DeviceIpTunnelPropertyFlags)
}

func (d *deviceIpTunnel) MarshalJSON() ([]uint8, error) {
	m, err := d.device.marshalMap()
	if err != nil {
		return nil, err
	}

	m["Mode"], _ = d.GetPropertyMode()
	m["Parent"], _ = d.GetPropertyParent()
	m["Local"], _ = d.GetPropertyLocal()
	m["Remote"], _ = d.GetPropertyRemote()
	m["Ttl"], _ = d.GetPropertyTtl()
	m["Tos"], _ = d.GetPropertyTos()
	m["PathMtuDiscovery"], _ = d.GetPropertyPathMtuDiscovery()
	m["InputKey"], _ = d.GetPropertyInputKey()
	m["OutputKey"], _ = d.GetPropertyOutputKey()
	m["EncapsulationLimit"], _ = d.GetPropertyEncapsulationLimit()
	m["FlowLabel"], _ = d.GetPropertyFlowLabel()
	m["Flags"], _ = d.GetPropertyFlags()
	return json.Marshal(m)
}
