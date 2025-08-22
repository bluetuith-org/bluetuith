package gonetworkmanager

import (
	"github.com/godbus/dbus/v5"
)

const (
	SettingsInterface  = NetworkManagerInterface + ".Settings"
	SettingsObjectPath = NetworkManagerObjectPath + "/Settings"

	/* Methods */
	SettingsListConnections      = SettingsInterface + ".ListConnections"
	SettingsGetConnectionByUUID  = SettingsInterface + ".GetConnectionByUuid"
	SettingsAddConnection        = SettingsInterface + ".AddConnection"
	SettingsAddConnectionUnsaved = SettingsInterface + ".AddConnectionUnsaved"
	SettingsLoadConnections      = SettingsInterface + ".LoadConnections"
	SettingsReloadConnections    = SettingsInterface + ".ReloadConnections"
	SettingsSaveHostname         = SettingsInterface + ".SaveHostname"

	/* Properties */
	SettingsPropertyConnections = SettingsInterface + ".Connections" // readable   ao
	SettingsPropertyHostname    = SettingsInterface + ".Hostname"    // readable   s
	SettingsPropertyCanModify   = SettingsInterface + ".CanModify"   // readable   b
)

type Settings interface {
	// ListConnections gets list the saved network connections known to NetworkManager
	ListConnections() ([]Connection, error)

	// ReloadConnections tells NetworkManager to reload all connection files from disk, including noticing any added or deleted connection files.
	ReloadConnections() error

	// GetConnectionByUUID gets the connection, given that connection's UUID.
	GetConnectionByUUID(uuid string) (Connection, error)

	// AddConnection adds new connection and save it to disk.
	AddConnection(settings ConnectionSettings) (Connection, error)

	// AddConnectionUnsaved Add new connection but do not save it to disk immediately. This operation does not start the network connection unless (1) device is idle and able to connect to the network described by the new connection, and (2) the connection is allowed to be started automatically. Use the 'Save' method on the connection to save these changes to disk. Note that unsaved changes will be lost if the connection is reloaded from disk (either automatically on file change or due to an explicit ReloadConnections call).
	AddConnectionUnsaved(settings ConnectionSettings) (Connection, error)

	// SaveHostname Save the hostname to persistent configuration.
	SaveHostname(hostname string) error

	// GetPropertyCanModify If true, adding and modifying connections is supported.
	GetPropertyCanModify() (bool, error)

	// GetPropertyHostname The machine hostname stored in persistent configuration.
	GetPropertyHostname() (string, error)
}

func NewSettings() (Settings, error) {
	var s settings
	return &s, s.init(NetworkManagerInterface, SettingsObjectPath)
}

type settings struct {
	dbusBase
}

func (s *settings) ListConnections() ([]Connection, error) {
	var connectionPaths []dbus.ObjectPath

	err := s.callWithReturn(&connectionPaths, SettingsListConnections)
	if err != nil {
		return nil, err
	}

	connections := make([]Connection, len(connectionPaths))

	for i, path := range connectionPaths {
		connections[i], err = NewConnection(path)
		if err != nil {
			return connections, err
		}
	}

	return connections, nil
}

// ReloadConnections tells NetworkManager to reload (and apply) configuration files
// from disk taking notice of any added or removed connections.
func (s *settings) ReloadConnections() error {
	return s.call(SettingsReloadConnections)
}

// GetConnectionByUUID gets the connection, given that connection's UUID.
func (s *settings) GetConnectionByUUID(uuid string) (Connection, error) {
	var path dbus.ObjectPath
	err := s.callWithReturn(&path, SettingsGetConnectionByUUID, uuid)
	if err != nil {
		return nil, err
	}

	return NewConnection(path)
}

func (s *settings) AddConnection(settings ConnectionSettings) (Connection, error) {
	var path dbus.ObjectPath
	err := s.callWithReturn(&path, SettingsAddConnection, settings)
	if err != nil {
		return nil, err
	}

	return NewConnection(path)
}

func (s *settings) AddConnectionUnsaved(settings ConnectionSettings) (Connection, error) {
	var path dbus.ObjectPath
	err := s.callWithReturn(&path, SettingsAddConnectionUnsaved, settings)

	if err != nil {
		return nil, err
	}

	return NewConnection(path)
}

func (s *settings) SaveHostname(hostname string) error {
	return s.call(SettingsSaveHostname, hostname)
}

func (s *settings) GetPropertyHostname() (string, error) {
	return s.getStringProperty(SettingsPropertyHostname)
}

func (s *settings) GetPropertyCanModify() (bool, error) {
	return s.getBoolProperty(SettingsPropertyCanModify)
}
