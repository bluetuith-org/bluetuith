//go:build linux

package linux

import (
	"errors"
	"time"

	bluetooth "github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	dbh "github.com/bluetuith-org/bluetooth-classic/linux/internal/dbushelper"
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/google/uuid"
)

// agent describes an Bluez agent connection.
// Note that, all public methods are exported to the Bluez Agent Manager
// via the system bus, and hence is called by the Agent Manager only.
// Any errors are published to the global error event stream.
type agent struct {
	systemBus *dbus.Conn

	authHandler bluetooth.SessionAuthorizer
	authTimeout time.Duration
	ctx         bluetooth.AuthTimeout

	initialized bool
}

const (
	agentPinCode        = "0000"
	agentPassKey uint32 = 1024
)

// newAgent returns a new BlueZ agent.
func newAgent(systemBus *dbus.Conn, authHandler bluetooth.SessionAuthorizer, authTimeout time.Duration) *agent {
	return &agent{
		systemBus:   systemBus,
		authHandler: authHandler,
		authTimeout: authTimeout,
	}
}

// setup creates a new BluezAgent, exports all its methods
// to the bluez DBus interface, and registers the agent.
func (b *agent) setup() error {
	if b.authHandler == nil {
		return errors.New("no authorization handler interface specified")
	}

	err := b.systemBus.Export(b, dbh.BluezAgentPath, dbh.BluezAgentIface)
	if err != nil {
		return err
	}

	node := &introspect.Node{
		Interfaces: []introspect.Interface{
			introspect.IntrospectData,
			{
				Name:    dbh.BluezAgentIface,
				Methods: introspect.Methods(b),
			},
		},
	}

	if err := b.systemBus.Export(introspect.NewIntrospectable(node), dbh.BluezAgentPath, dbh.DbusIntrospectableIface); err != nil {
		return err
	}

	if err := b.callAgentManager("RegisterAgent", dbh.BluezAgentPath, "KeyboardDisplay").Store(); err != nil {
		return err
	}

	if err := b.callAgentManager("RequestDefaultAgent", dbh.BluezAgentPath).Store(); err != nil {
		return err
	}

	b.initialized = true

	return nil
}

// remove removes the agent.
func (b *agent) remove() error {
	if !b.initialized {
		return nil
	}

	return b.callAgentManager("UnregisterAgent", dbh.BluezAgentPath).Store()
}

// RequestPinCode returns a predefined pincode to the agent's pincode request.
func (b *agent) RequestPinCode(_ dbus.ObjectPath) (string, *dbus.Error) {
	return agentPinCode, nil
}

// RequestPasskey returns a predefined passkey to the agent's passkey request.
func (b *agent) RequestPasskey(_ dbus.ObjectPath) (uint32, *dbus.Error) {
	return agentPassKey, nil
}

// DisplayPinCode displays a pincode from the device via the agent.
func (b *agent) DisplayPinCode(devicePath dbus.ObjectPath, pincode string) *dbus.Error {
	address, ok := dbh.PathConverter.Address(dbh.DbusPathDevice, devicePath)
	if !ok {
		dbh.PublishError(errors.New(string(devicePath)),
			"Bluez agent error: Device not found",
			"error_at", "displaypin-device-address",
		)

		return dbus.MakeFailedError(errors.New("address not found"))
	}

	b.ctx = bluetooth.NewAuthTimeout(b.authTimeout)
	defer b.Cancel()

	if err := b.authHandler.DisplayPinCode(b.ctx, address, pincode); err != nil {
		dbh.PublishError(err,
			"Bluez agent error: Authorization callback returned an error",
			"error_at", "displaypin-device-address",
		)

		return dbus.MakeFailedError(err)
	}

	return nil
}

// DisplayPasskey displays a passkey from the device via the agent.
func (b *agent) DisplayPasskey(devicePath dbus.ObjectPath, passkey uint32, entered uint16) *dbus.Error {
	address, ok := dbh.PathConverter.Address(dbh.DbusPathDevice, devicePath)
	if !ok {
		dbh.PublishError(errors.New(string(devicePath)),
			"Bluez agent error: Device not found",
			"error_at", "displaypk-device-address",
		)

		return dbus.MakeFailedError(errors.New("address not found"))
	}

	b.ctx = bluetooth.NewAuthTimeout(b.authTimeout)
	defer b.Cancel()

	if err := b.authHandler.DisplayPasskey(b.ctx, address, passkey, entered); err != nil {
		dbh.PublishError(err,
			"Bluez agent error: Authorization callback returned an error",
			"error_at", "displaypk-device-address",
		)

		return dbus.MakeFailedError(err)
	}

	return nil
}

// RequestConfirmation requests confirmation to pair with the device using the provided passkey.
func (b *agent) RequestConfirmation(devicePath dbus.ObjectPath, passkey uint32) *dbus.Error {
	address, ok := dbh.PathConverter.Address(dbh.DbusPathDevice, devicePath)
	if !ok {
		dbh.PublishError(errors.New(string(devicePath)),
			"Bluez agent error: Device not found",
			"error_at", "authpk-device-address",
		)

		return dbus.MakeFailedError(errors.New("address not found"))
	}

	b.ctx = bluetooth.NewAuthTimeout(b.authTimeout)
	defer b.Cancel()

	if err := b.authHandler.ConfirmPasskey(b.ctx, address, passkey); err != nil {
		dbh.PublishError(err,
			"Bluez agent error: Authorization callback returned an error",
			"error_at", "authpk-device-address",
		)

		return dbus.MakeFailedError(err)
	}

	return nil
}

// RequestAuthorization requests authorization to pair with a device.
func (b *agent) RequestAuthorization(devicePath dbus.ObjectPath) *dbus.Error {
	address, ok := dbh.PathConverter.Address(dbh.DbusPathDevice, devicePath)
	if !ok {
		dbh.PublishError(errors.New(string(devicePath)),
			"Bluez agent error: Device not found",
			"error_at", "authpairing-device-address",
		)

		return dbus.MakeFailedError(errors.New("address not found"))
	}

	b.ctx = bluetooth.NewAuthTimeout(b.authTimeout)
	defer b.Cancel()

	if err := b.authHandler.AuthorizePairing(b.ctx, address); err != nil {
		dbh.PublishError(err,
			"Bluez agent error: Authorization callback returned an error",
			"error_at", "authpairing-device-address",
		)

		return dbus.MakeFailedError(err)
	}

	return nil
}

// AuthorizeService requests authorization of a Bluetooth service using its profile UUID.
func (b *agent) AuthorizeService(devicePath dbus.ObjectPath, uuidstr string) *dbus.Error {
	address, ok := dbh.PathConverter.Address(dbh.DbusPathDevice, devicePath)
	if !ok {
		dbh.PublishError(errors.New(string(devicePath)),
			"Bluez agent error: Device not found",
			"error_at", "authservice-device-address",
		)

		return dbus.MakeFailedError(errors.New("address not found"))
	}

	u, _ := uuid.Parse(uuidstr)
	b.ctx = bluetooth.NewAuthTimeout(b.authTimeout)
	defer b.Cancel()

	if err := b.authHandler.AuthorizeService(b.ctx, address, u); err != nil {
		dbh.PublishError(err,
			"Bluez agent error: Authorization callback returned an error",
			"error_at", "authservice-device-address",
		)

		return dbus.MakeFailedError(err)
	}

	return nil
}

// Cancel is called when the Bluez agent request was cancelled.
func (b *agent) Cancel() *dbus.Error {
	b.ctx.Cancel()

	return nil
}

// Release is called when the Bluez agent is unregistered.
func (b *agent) Release() *dbus.Error {
	return nil
}

// callAgentManager calls the AgentManager1 interface with the provided arguments.
func (b *agent) callAgentManager(method string, args ...any) *dbus.Call {
	return b.systemBus.Object(dbh.BluezBusName, dbh.BluezAgentManagerPath).Call(dbh.BluezAgentManagerIface+"."+method, 0, args...)
}
