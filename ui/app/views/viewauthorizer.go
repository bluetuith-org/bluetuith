package views

import (
	"github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	"github.com/google/uuid"
)

type authorizer struct {
	v           *Views
	initialized bool
}

func newAuthorizer(v *Views) *authorizer {
	return &authorizer{v, false}
}

func (a *authorizer) setInitialized() {
	a.initialized = true
}

func (a *authorizer) AuthorizeTransfer(timeout bluetooth.AuthTimeout, props bluetooth.FileTransferData) error {
	panic("not implemented") // TODO: Implement
}
func (a *authorizer) DisplayPinCode(timeout bluetooth.AuthTimeout, address bluetooth.MacAddress, pincode string) error {
	panic("not implemented") // TODO: Implement
}
func (a *authorizer) DisplayPasskey(timeout bluetooth.AuthTimeout, address bluetooth.MacAddress, passkey uint32, entered uint16) error {
	panic("not implemented") // TODO: Implement
}
func (a *authorizer) ConfirmPasskey(timeout bluetooth.AuthTimeout, address bluetooth.MacAddress, passkey uint32) error {
	panic("not implemented") // TODO: Implement
}
func (a *authorizer) AuthorizePairing(timeout bluetooth.AuthTimeout, address bluetooth.MacAddress) error {
	panic("not implemented") // TODO: Implement
}
func (a *authorizer) AuthorizeService(timeout bluetooth.AuthTimeout, address bluetooth.MacAddress, uuid uuid.UUID) error {
	panic("not implemented") // TODO: Implement
}
