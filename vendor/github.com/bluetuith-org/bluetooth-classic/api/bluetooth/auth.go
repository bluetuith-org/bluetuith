package bluetooth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// SessionAuthorizer describes an authentication interface for authorizing session functions.
type SessionAuthorizer interface {
	AuthorizeReceiveFile
	AuthorizeDevicePairing
}

// AuthTimeout describes an authentication timeout duration.
// The context value is created with 'context.WithTimeout()'.
type AuthTimeout struct {
	context.Context
	cancel context.CancelFunc
}

// NewAuthTimeout returns a new authentication timeout token.
func NewAuthTimeout(timeout time.Duration) AuthTimeout {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	return AuthTimeout{ctx, cancel}
}

// Cancel cancels the inner context.
func (a *AuthTimeout) Cancel() {
	if a.cancel != nil {
		a.cancel()
	}
}

// DefaultAuthorizer describes a default authentication handler.
type DefaultAuthorizer struct{}

// AuthorizeTransfer accepts all file transfer authorization requests.
func (DefaultAuthorizer) AuthorizeTransfer(AuthTimeout, ObjectPushData) error {
	return nil
}

// DisplayPinCode accepts all display pincode requests.
func (DefaultAuthorizer) DisplayPinCode(AuthTimeout, MacAddress, string) error {
	return nil
}

// DisplayPasskey accepts all display passkey requests.
func (DefaultAuthorizer) DisplayPasskey(AuthTimeout, MacAddress, uint32, uint16) error {
	return nil
}

// ConfirmPasskey accepts all passkey confirmation requests.
func (DefaultAuthorizer) ConfirmPasskey(AuthTimeout, MacAddress, uint32) error {
	return nil
}

// AuthorizePairing accepts all pairing authorization requests.
func (DefaultAuthorizer) AuthorizePairing(AuthTimeout, MacAddress) error {
	return nil
}

// AuthorizeService accepts all service (Bluetooth profile) authorization requests.
func (DefaultAuthorizer) AuthorizeService(AuthTimeout, MacAddress, uuid.UUID) error {
	return nil
}
