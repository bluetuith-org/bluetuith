//go:build !linux

package session

import (
	"github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	"github.com/bluetuith-org/bluetooth-classic/shim"
)

// NewSession returns a platform-specific session handler.
func NewSession() bluetooth.Session {
	return &shim.ShimSession{}
}
