//go:build !linux

package shim

import (
	"github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	"github.com/bluetuith-org/bluetooth-classic/api/errorkinds"
)

// network holds the network manager and active connections.
type network struct {
}

// Connect connects to a specific device according to the provided NetworkType
// and assigns a name to the established connection.
func (n *network) Connect(string, bluetooth.NetworkType) error {
	return errorkinds.ErrNotSupported
}

// Disconnect disconnects from an established connection.
func (n *network) Disconnect() error {
	return errorkinds.ErrNotSupported
}
