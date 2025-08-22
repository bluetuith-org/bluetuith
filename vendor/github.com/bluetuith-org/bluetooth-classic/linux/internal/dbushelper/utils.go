//go:build linux

package dbushelper

import "github.com/godbus/dbus/v5"

// ListActivatableBusNames returns a list of bus names from the provided DBus connection.
func ListActivatableBusNames(conn *dbus.Conn) ([]string, error) {
	var names []string

	if err := conn.Object("org.freedesktop.DBus", "/org/freedesktop/DBus").
		Call("org.freedesktop.DBus.ListActivatableNames", 0).
		Store(&names); err != nil {
		return nil, err
	}

	return names, nil
}
