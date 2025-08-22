package gonetworkmanager

import (
	"encoding/json"

	"github.com/godbus/dbus/v5"
)

const (
	DHCP6ConfigInterface = NetworkManagerInterface + ".DHCP6Config"

	// Properties
	DHCP6ConfigPropertyOptions = DHCP6ConfigInterface + ".Options"
)

type DHCP6Options map[string]interface{}

type DHCP6Config interface {
	// GetPropertyOptions GetOptions gets options map of configuration returned by the IPv4 DHCP server.
	GetPropertyOptions() (DHCP6Options, error)

	MarshalJSON() ([]byte, error)
}

func NewDHCP6Config(objectPath dbus.ObjectPath) (DHCP6Config, error) {
	var c dhcp6Config
	return &c, c.init(NetworkManagerInterface, objectPath)
}

type dhcp6Config struct {
	dbusBase
}

func (c *dhcp6Config) GetPropertyOptions() (DHCP6Options, error) {
	options, err := c.getMapStringVariantProperty(DHCP6ConfigPropertyOptions)
	rv := make(DHCP6Options)

	if err != nil {
		return rv, err
	}

	for k, v := range options {
		rv[k] = v.Value()
	}

	return rv, nil
}

func (c *dhcp6Config) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{})
	m["Options"], _ = c.GetPropertyOptions()

	return json.Marshal(m)
}
