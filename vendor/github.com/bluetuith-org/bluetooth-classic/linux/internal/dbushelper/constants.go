//go:build linux

package dbushelper

import (
	"github.com/godbus/dbus/v5"
	"github.com/rs/xid"
)

// The DBus specific bus and property names.
const (
	DbusGetPropertiesIface    = "org.freedesktop.DBus.Properties.Get"
	DbusGetAllPropertiesIface = "org.freedesktop.DBus.Properties.GetAll"
	DbusSetPropertiesIface    = "org.freedesktop.DBus.Properties.Set"
	DbusObjectManagerIface    = "org.freedesktop.DBus.ObjectManager.GetManagedObjects"
	DbusIntrospectableIface   = "org.freedesktop.DBus.Introspectable"

	DbusSignalAddMatchIface          = "org.freedesktop.DBus.AddMatch"
	DbusSignalPropertyChangedIface   = "org.freedesktop.DBus.Properties.PropertiesChanged"
	DbusSignalInterfacesAddedIface   = "org.freedesktop.DBus.ObjectManager.InterfacesAdded"
	DbusSignalInterfacesRemovedIface = "org.freedesktop.DBus.ObjectManager.InterfacesRemoved"

	BluezBusName           = "org.bluez"
	BluezAdapterIface      = "org.bluez.Adapter1"
	BluezDeviceIface       = "org.bluez.Device1"
	BluezBatteryIface      = "org.bluez.Battery1"
	BluezMediaControlIface = "org.bluez.MediaControl1"
	BluezMediaPlayerIface  = "org.bluez.MediaPlayer1"

	BluezAgentIface        = "org.bluez.Agent1"
	BluezAgentManagerIface = "org.bluez.AgentManager1"
	BluezAgentManagerPath  = dbus.ObjectPath("/org/bluez")

	ObexBusName         = "org.bluez.obex"
	ObexClientIface     = "org.bluez.obex.Client1"
	ObexSessionIface    = "org.bluez.obex.Session1"
	ObexTransferIface   = "org.bluez.obex.Transfer1"
	ObexObjectPushIface = "org.bluez.obex.ObjectPush1"
	ObexBusPath         = dbus.ObjectPath("/org/bluez/obex")

	ObexAgentIface        = "org.bluez.obex.Agent1"
	ObexAgentManagerIface = "org.bluez.obex.AgentManager1"
	ObexAgentManagerPath  = dbus.ObjectPath("/org/bluez/obex")
)

// ObexAgentPath is a randomized path for registering a Bluez Obex Agent.
var ObexAgentPath = dbus.ObjectPath("/org/bluez/obex/agent/obexagent" + xid.New().String())

// BluezAgentPath is a randomized path for registering a Bluez Agent.
var BluezAgentPath = dbus.ObjectPath("/org/bluez/agent/blueagent" + xid.New().String())
