package gonetworkmanager

import (
	"encoding/json"

	"github.com/godbus/dbus/v5"
)

const (
	CheckpointInterface = NetworkManagerInterface + ".Checkpoint"

	/* Properties */
	CheckpointPropertyDevices         = CheckpointInterface + ".Devices"         // readable   ao
	CheckpointPropertyCreated         = CheckpointInterface + ".Created"         // readable   x
	CheckpointPropertyRollbackTimeout = CheckpointInterface + ".RollbackTimeout" // readable   u
)

type Checkpoint interface {
	GetPath() dbus.ObjectPath

	// GetPropertyDevices Array of object paths for devices which are part of this checkpoint.
	GetPropertyDevices() ([]Device, error)

	// GetPropertyCreated The timestamp (in CLOCK_BOOTTIME milliseconds) of checkpoint creation.
	GetPropertyCreated() (int64, error)

	// GetPropertyRollbackTimeout Timeout in seconds for automatic rollback, or zero.
	GetPropertyRollbackTimeout() (uint32, error)

	MarshalJSON() ([]byte, error)
}

func NewCheckpoint(objectPath dbus.ObjectPath) (Checkpoint, error) {
	var c checkpoint
	return &c, c.init(NetworkManagerInterface, objectPath)
}

type checkpoint struct {
	dbusBase
}

func (c *checkpoint) GetPropertyDevices() ([]Device, error) {
	devicesPaths, err := c.getSliceObjectProperty(CheckpointPropertyDevices)
	if err != nil {
		return nil, err
	}

	devices := make([]Device, len(devicesPaths))
	for i, path := range devicesPaths {
		devices[i], err = NewDevice(path)
		if err != nil {
			return devices, err
		}
	}

	return devices, nil
}

func (c *checkpoint) GetPropertyCreated() (int64, error) {
	return c.getInt64Property(CheckpointPropertyCreated)
}

func (c *checkpoint) GetPropertyRollbackTimeout() (uint32, error) {
	return c.getUint32Property(CheckpointPropertyRollbackTimeout)
}

func (c *checkpoint) GetPath() dbus.ObjectPath {
	return c.obj.Path()
}

func (c *checkpoint) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{})

	m["Devices"], _ = c.GetPropertyDevices()
	m["Created"], _ = c.GetPropertyCreated()
	m["RollbackTimeout"], _ = c.GetPropertyRollbackTimeout()

	return json.Marshal(m)
}
