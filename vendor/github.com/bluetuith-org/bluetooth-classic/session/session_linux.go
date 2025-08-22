//go:build linux

package session

import (
	"github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	"github.com/bluetuith-org/bluetooth-classic/linux"
)

// NewSession returns a Linux-specific session handler.
func NewSession() bluetooth.Session {
	return &linux.BluezSession{}
}
