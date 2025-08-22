package gonetworkmanager

import (
	"encoding/json"
	"errors"

	"github.com/godbus/dbus/v5"
)

const (
	IP6ConfigInterface = NetworkManagerInterface + ".IP6Config"

	/* Properties */
	IP6ConfigPropertyAddresses   = IP6ConfigInterface + ".Addresses"   // readable   a(ayuay)
	IP6ConfigPropertyAddressData = IP6ConfigInterface + ".AddressData" // readable   aa{sv}
	IP6ConfigPropertyGateway     = IP6ConfigInterface + ".Gateway"     // readable   s
	IP6ConfigPropertyRoutes      = IP6ConfigInterface + ".Routes"      // readable   a(ayuayu)
	IP6ConfigPropertyRouteData   = IP6ConfigInterface + ".RouteData"   // readable   aa{sv}
	IP6ConfigPropertyNameservers = IP6ConfigInterface + ".Nameservers" // readable   aay
	IP6ConfigPropertyDomains     = IP6ConfigInterface + ".Domains"     // readable   as
	IP6ConfigPropertySearches    = IP6ConfigInterface + ".Searches"    // readable   as
	IP6ConfigPropertyDnsOptions  = IP6ConfigInterface + ".DnsOptions"  // readable   as
	IP6ConfigPropertyDnsPriority = IP6ConfigInterface + ".DnsPriority" // readable   i
)

// Deprecated: use IP6AddressData instead
type IP6Address struct {
	Address string
	Prefix  uint8
	Gateway string
}

type IP6AddressData struct {
	Address string
	Prefix  uint8
}

// Deprecated: use IP6RouteData instead
type IP6Route struct {
	Route   string
	Prefix  uint8
	NextHop string
	Metric  uint8
}

type IP6RouteData struct {
	Destination          string
	Prefix               uint8
	NextHop              string
	Metric               uint8
	AdditionalAttributes map[string]string
}

type IP6Config interface {

	// GetPropertyAddressData Array of IP address data objects. All addresses will include "address" (an IP address string), and "prefix" (a uint). Some addresses may include additional attributes.
	GetPropertyAddressData() ([]IP6AddressData, error)

	// GetPropertyGateway The gateway in use.
	GetPropertyGateway() (string, error)

	// GetPropertyRouteData Array of IP route data objects. All routes will include "dest" (an IP address string) and "prefix" (a uint). Some routes may include "next-hop" (an IP address string), "metric" (a uint), and additional attributes.
	GetPropertyRouteData() ([]IP6RouteData, error)

	// GetPropertyNameservers GetNameservers gets the nameservers in use.
	GetPropertyNameservers() ([][]byte, error)

	// GetPropertyDomains A list of domains this address belongs to.
	GetPropertyDomains() ([]string, error)

	// GetPropertySearches A list of dns searches.
	GetPropertySearches() ([]string, error)

	// GetPropertyDnsOptions A list of DNS options that modify the behavior of the DNS resolver. See resolv.conf(5) manual page for the list of supported options.
	GetPropertyDnsOptions() ([]string, error)

	// GetPropertyDnsPriority The relative priority of DNS servers.
	GetPropertyDnsPriority() (uint32, error)

	MarshalJSON() ([]byte, error)
}

func NewIP6Config(objectPath dbus.ObjectPath) (IP6Config, error) {
	var c ip6Config
	return &c, c.init(NetworkManagerInterface, objectPath)
}

type ip6Config struct {
	dbusBase
}

func (c *ip6Config) GetPropertyAddressData() ([]IP6AddressData, error) {
	addresses, err := c.getSliceMapStringVariantProperty(IP6ConfigPropertyAddressData)
	ret := make([]IP6AddressData, len(addresses))

	if err != nil {
		return ret, err
	}

	for i, address := range addresses {
		prefix, ok := address["prefix"].Value().(uint32)
		if !ok {
			return ret, errors.New("unexpected variant type for prefix")
		}

		address, ok := address["address"].Value().(string)
		if !ok {
			return ret, errors.New("unexpected variant type for address")
		}

		ret[i] = IP6AddressData{
			Address: address,
			Prefix:  uint8(prefix),
		}
	}

	return ret, nil
}

func (c *ip6Config) GetPropertyGateway() (string, error) {
	return c.getStringProperty(IP6ConfigPropertyGateway)
}

func (c *ip6Config) GetPropertyRouteData() ([]IP6RouteData, error) {
	routesData, err := c.getSliceMapStringVariantProperty(IP6ConfigPropertyRouteData)
	routes := make([]IP6RouteData, len(routesData))

	if err != nil {
		return routes, err
	}

	for index, routeData := range routesData {

		route := IP6RouteData{}

		for routeDataAttributeName, routeDataAttribute := range routeData {
			switch routeDataAttributeName {
			case "dest":
				destination, ok := routeDataAttribute.Value().(string)
				if !ok {
					return routes, errors.New("unexpected variant type for dest")
				}
				route.Destination = destination
			case "prefix":
				prefix, ok := routeDataAttribute.Value().(uint32)
				if !ok {
					return routes, errors.New("unexpected variant type for prefix")
				}
				route.Prefix = uint8(prefix)
			case "next-hop":
				nextHop, ok := routeDataAttribute.Value().(string)
				if !ok {
					return routes, errors.New("unexpected variant type for next-hop")
				}
				route.NextHop = nextHop
			case "metric":
				metric, ok := routeDataAttribute.Value().(uint32)
				if !ok {
					return routes, errors.New("unexpected variant type for metric")
				}
				route.Metric = uint8(metric)
			default:
				if route.AdditionalAttributes == nil {
					route.AdditionalAttributes = make(map[string]string)
				}
				route.AdditionalAttributes[routeDataAttributeName] = routeDataAttribute.String()
			}
		}

		routes[index] = route
	}
	return routes, nil
}

func (c *ip6Config) GetPropertyNameservers() ([][]byte, error) {
	nameservers, err := c.getSliceSliceByteProperty(IP6ConfigPropertyNameservers)
	ret := make([][]byte, len(nameservers))

	if err != nil {
		return ret, err
	}

	for i, nameserver := range nameservers {
		ret[i] = nameserver
	}

	return ret, nil
}

func (c *ip6Config) GetPropertyDomains() ([]string, error) {
	return c.getSliceStringProperty(IP6ConfigPropertyDomains)
}

func (c *ip6Config) GetPropertySearches() ([]string, error) {
	return c.getSliceStringProperty(IP6ConfigPropertySearches)
}

func (c *ip6Config) GetPropertyDnsOptions() ([]string, error) {
	return c.getSliceStringProperty(IP6ConfigPropertyDnsOptions)
}

func (c *ip6Config) GetPropertyDnsPriority() (uint32, error) {
	return c.getUint32Property(IP6ConfigPropertyDnsPriority)
}

func (c *ip6Config) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{})

	m["Addresses"], _ = c.GetPropertyAddressData()
	m["Routes"], _ = c.GetPropertyRouteData()
	m["Nameservers"], _ = c.GetPropertyNameservers()
	m["Domains"], _ = c.GetPropertyDomains()

	return json.Marshal(m)
}
