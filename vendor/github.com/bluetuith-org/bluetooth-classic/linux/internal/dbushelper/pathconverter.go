//go:build linux

package dbushelper

import (
	"path/filepath"

	bluetooth "github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	"github.com/godbus/dbus/v5"
	"github.com/puzpuzpuz/xsync/v3"
)

// DbusPathType represents the type of DBus path in the Bluez DBus service.
// For example, adapter paths will have a path type of DbusPathAdapter and will
// be mapped to an adapter address (/org/bluez/hci0 => DBusPathAdapter).
// For other DBus path types like DbusPathObexSession and DbusPathObexTransfer,
// their paths will be mapped to device addresses.
type DbusPathType int

// The different Bluez DBus path types.
const (
	DbusPathDevice DbusPathType = iota
	DbusPathAdapter
	DbusPathObexSession
	DbusPathObexTransfer
)

// dbusPath holds the Bluez DBus path and its type.
type dbusPath struct {
	pathType DbusPathType
	path     dbus.ObjectPath
}

// dbusPathConverter holds a list of Bluez DBus paths and maps them
// to their respective Bluetooth addresses.
type dbusPathConverter struct {
	paths *xsync.MapOf[dbusPath, bluetooth.MacAddress]
}

// PathConverter is used to obtain respective Bluetooth addresses that are mapped to
// Bluez DBus paths. This is mainly used to identify adapters and devices.
var PathConverter = dbusPathConverter{paths: xsync.NewMapOf[dbusPath, bluetooth.MacAddress]()}

// AddDbusPath adds a mapping of a Bluez DBus path and a Bluetooth address to the path converter.
func (d *dbusPathConverter) AddDbusPath(pathType DbusPathType, path dbus.ObjectPath, address bluetooth.MacAddress) {
	d.paths.Store(dbusPath{pathType: pathType, path: path}, address)
}

// RemoveDbusPath removes a mapping of a Bluez DBus path and a Bluetooth address from the path converter.
func (d *dbusPathConverter) RemoveDbusPath(pathType DbusPathType, path dbus.ObjectPath) {
	d.paths.Delete(dbusPath{pathType: pathType, path: path})
}

// RemoveAdapterDbusPath removes mappings of a Bluez DBus adapter path and its associated devices.
func (d *dbusPathConverter) RemoveAdapterDbusPath(path dbus.ObjectPath) {
	_, ok := d.Address(DbusPathAdapter, path)
	if !ok {
		return
	}

	d.RemoveDbusPath(DbusPathAdapter, path)
	d.paths.Range(func(p dbusPath, _ bluetooth.MacAddress) bool {
		if p.pathType != DbusPathDevice || filepath.Dir(string(p.path)) == string(p.path) {
			return true
		}

		d.RemoveDbusPath(DbusPathDevice, p.path)

		return true
	})
}

// Address returns a Bluetooth address that is mapped to the provided Bluez DBus path.
func (d *dbusPathConverter) Address(pathType DbusPathType, path dbus.ObjectPath) (bluetooth.MacAddress, bool) {
	return d.paths.Load(dbusPath{pathType: pathType, path: path})
}

// DbusPath returns a Bluez DBus path that is mapped to the provided Bluetooth address.
func (d *dbusPathConverter) DbusPath(pathType DbusPathType, address bluetooth.MacAddress) (dbus.ObjectPath, bool) {
	var dpath dbus.ObjectPath

	d.paths.Range(func(p dbusPath, addr bluetooth.MacAddress) bool {
		if address == addr && p.pathType == pathType {
			dpath = p.path

			return false
		}

		return true
	})

	return dpath, dpath != ""
}
