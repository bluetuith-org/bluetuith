package views

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	"github.com/google/uuid"
)

// authorizer holds a set of functions used to authenticate pairing and receiving
// file transfer requests. A new instance of this is supposed to be passed to
// [bluetooth.Session.Start] to handle any authorization requests.
type authorizer struct {
	v           *Views
	initialized bool

	alwaysAuthorize bool
}

// newAuthorizer returns a new authorizer.
func newAuthorizer(v *Views) *authorizer {
	return &authorizer{v: v}
}

// setInitialized sets the authorizer to the initialized state.
// This is called after all views have been initialized.
func (a *authorizer) setInitialized() {
	a.initialized = true
}

// AuthorizeTransfer asks the user to authorize a file transfer (Object Push) that is about to be sent
// from the remote device.
func (a *authorizer) AuthorizeTransfer(timeout bluetooth.AuthTimeout, props bluetooth.ObjectPushData) error {
	if !a.initialized {
		return nil
	}

	device, err := a.v.app.Session().Device(props.Address).Properties()
	if err != nil {
		return err
	}

	filename := props.Name
	if filename == "" {
		filename = filepath.Base(props.Filename)
	}

	reply := a.v.status.waitForInput(timeout, fmt.Sprintf("[::bu]%s[-:-:-]: Accept file '%s' (y/n/a)", device.Name, filename))
	switch reply {
	case "a":
		a.alwaysAuthorize = true
		fallthrough

	case "y":
		a.v.progress.showStatus()
		return nil
	}

	return errors.New("Cancelled")
}

// DisplayPinCode displays the pincode from the remote device to the user during a pairing authorization session.
func (a *authorizer) DisplayPinCode(timeout bluetooth.AuthTimeout, address bluetooth.MacAddress, pincode string) error {
	if !a.initialized {
		return nil
	}

	device, err := a.v.app.Session().Device(address).Properties()
	if err != nil {
		return err
	}

	msg := fmt.Sprintf(
		"The pincode for [::bu]%s[-:-:-] is:\n\n[::b]%s[-:-:-]",
		device.Name, pincode,
	)

	modal := a.generateDisplayModal(address, "pincode", "Pin Code", msg)
	modal.display(timeout)

	return nil
}

// DisplayPasskey only displays the passkey from the remote device to the user during a pairing authorization session.
// This can be called multiple times, since each time the user enters a number on the remote device, this function
// is called with the updated 'entered' value.
// TODO: Handle multiple calls/draws when this function is called.
func (a *authorizer) DisplayPasskey(timeout bluetooth.AuthTimeout, address bluetooth.MacAddress, passkey uint32, entered uint16) error {
	if !a.initialized {
		return nil
	}

	device, err := a.v.app.Session().Device(address).Properties()
	if err != nil {
		return err
	}

	msg := fmt.Sprintf(
		"The passkey for [::bu]%s[-:-:-] is:\n\n[::b]%d[-:-:-]",
		device.Name, passkey,
	)
	if entered > 0 {
		msg += fmt.Sprintf("\n\nYou have entered %d", entered)
	}

	modal := a.generateDisplayModal(address, "passkey-display", "Passkey Display", msg)
	modal.display(timeout)

	return nil
}

// ConfirmPasskey asks the user to authorize the pairing request using the provided passkey.
func (a *authorizer) ConfirmPasskey(timeout bluetooth.AuthTimeout, address bluetooth.MacAddress, passkey uint32) error {
	if !a.initialized {
		return nil
	}

	device, err := a.v.app.Session().Device(address).Properties()
	if err != nil {
		return err
	}

	msg := fmt.Sprintf(
		"Confirm passkey for [::bu]%s[-:-:-] is \n\n[::b]%d[-:-:-]",
		device.Name, passkey,
	)

	modal := a.generateConfirmModal(address, "passkey-confirm", "Passkey Confirmation", msg)
	reply := modal.getReply(timeout)
	if reply != "y" {
		return errors.New("Reply was: " + reply)
	}

	_ = a.v.app.Session().Device(address).SetTrusted(true)

	return nil
}

// AuthorizePairing asks the user to authorize a pairing request.
func (a *authorizer) AuthorizePairing(timeout bluetooth.AuthTimeout, address bluetooth.MacAddress) error {
	if !a.initialized {
		return nil
	}

	device, err := a.v.app.Session().Device(address).Properties()
	if err != nil {
		return err
	}
	msg := fmt.Sprintf("Confirm pairing with [::bu]%s[-:-:-]", device.Name)

	modal := a.generateConfirmModal(address, "pairing-confirm", "Pairing Confirmation", msg)
	reply := modal.getReply(timeout)
	if reply != "y" {
		return errors.New("Reply was: " + reply)
	}

	_ = a.v.app.Session().Device(address).SetTrusted(true)

	return nil
}

// AuthorizeService asks the user to authorize whether a specific Bluetooth Profile is allowed to be used.
func (a *authorizer) AuthorizeService(timeout bluetooth.AuthTimeout, address bluetooth.MacAddress, profileUUID uuid.UUID) error {
	if !a.initialized || a.alwaysAuthorize {
		return nil
	}

	serviceName := bluetooth.ServiceType(profileUUID.String())
	device, err := a.v.app.Session().Device(address).Properties()
	if err != nil {
		return err
	}

	reply := a.v.status.waitForInput(timeout, fmt.Sprintf("[::bu]%s[-:-:-]: Authorize service '%s' (y/n/a)", device.Name, serviceName))
	switch reply {
	case "a":
		a.alwaysAuthorize = true
		fallthrough

	case "y":
		return nil
	}

	return errors.New("Cancelled")
}

// generateConfirmModal generates a confirmation modal with the provided parameters.
func (a *authorizer) generateConfirmModal(address bluetooth.MacAddress, name, title, msg string) *confirmModalView {
	return a.v.modals.newConfirmModal(name+":"+address.String(), title, msg)
}

// generateDisplayModal generates a display modal with the provided parameters.
func (a *authorizer) generateDisplayModal(address bluetooth.MacAddress, name, title, msg string) *displayModalView {
	return a.v.modals.newDisplayModal(name+":"+address.String(), title, msg)
}
