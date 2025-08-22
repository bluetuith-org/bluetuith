//go:build linux

package obex

import (
	"errors"
	"path/filepath"
	"time"

	bluetooth "github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	dbh "github.com/bluetuith-org/bluetooth-classic/linux/internal/dbushelper"
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
)

// agent describes an OBEX agent connection.
// Note that, all public methods are exported to the Obex Agent Manager
// via the session bus, and hence is called by the Agent Manager only.
// Any errors are published to the global error event stream.
type agent struct {
	authHandler bluetooth.AuthorizeReceiveFile

	ctx         bluetooth.AuthTimeout
	authTimeout time.Duration

	initialized bool

	*fileTransfer
}

// newAgent returns a new OBEX agent.
func newAgent(authHandler bluetooth.AuthorizeReceiveFile, authTimeout time.Duration, transferSession *fileTransfer) *agent {
	return &agent{
		authHandler:  authHandler,
		authTimeout:  authTimeout,
		fileTransfer: transferSession,
	}
}

// setup sets up an OBEX agent.
func (o *agent) setup() error {
	if o.authHandler == nil {
		return errors.New("no authorization handler interface specified")
	}

	err := o.SessionBus.Export(o, dbh.ObexAgentPath, dbh.ObexAgentIface)
	if err != nil {
		return err
	}

	node := &introspect.Node{
		Interfaces: []introspect.Interface{
			introspect.IntrospectData,
			{
				Name:    dbh.ObexAgentIface,
				Methods: introspect.Methods(o),
			},
		},
	}

	if err := o.SessionBus.Export(
		introspect.NewIntrospectable(node),
		dbh.ObexAgentPath,
		dbh.DbusIntrospectableIface,
	); err != nil {
		return err
	}

	if err := o.callObexAgentManager("RegisterAgent", dbh.ObexAgentPath).Store(); err != nil {
		return err
	}

	o.initialized = true

	return nil
}

// remove removes the OBEX agent.
func (o *agent) remove() error {
	if !o.initialized {
		return nil
	}

	return o.callObexAgentManager("UnregisterAgent", dbh.ObexAgentPath).Store()
}

// makeError creates a custom error.
func (o *agent) makeError() *dbus.Error {
	return &dbus.Error{
		Name: "org.bluez.obex.Error.Rejected",
		Body: []any{"Rejected"},
	}
}

// AuthorizePush asks for confirmation before receiving a transfer from the host device.
func (o *agent) AuthorizePush(transferPath dbus.ObjectPath) (string, *dbus.Error) {
	sessionPath := dbus.ObjectPath(filepath.Dir(string(transferPath)))

	sessionProperty, err := o.sessionProperties(sessionPath)
	if err != nil {
		dbh.PublishError(err,
			"OBEX agent error: Could not get session properties",
			"error_at", "authpush-session-properties",
		)

		return "", o.makeError()
	}

	transferProperty, err := o.transferProperties(transferPath)
	if err != nil {
		dbh.PublishError(err,
			"OBEX agent error: Could not get transfer properties",
			"error_at", "authpush-transfer-properties",
		)

		return "", o.makeError()
	}

	if sessionProperty.Root == "" {
		dbh.PublishError(err,
			"OBEX agent error: Session properties are empty",
			"error_at", "authpush-session-rootdest",
		)

		return "", o.makeError()
	}

	if transferProperty.Status == bluetooth.TransferError {
		dbh.PublishError(err,
			"OBEX agent error: Transfer property is empty",
			"error_at", "authpush-transfer-status",
		)

		return "", o.makeError()
	}

	bluetooth.ObjectPushEvents().PublishAdded(transferProperty.appendExtra(transferPath, sessionProperty.Destination, struct{}{}).ObjectPushData)

	path := filepath.Join(sessionProperty.Root, transferProperty.Name)
	o.ctx = bluetooth.NewAuthTimeout(o.authTimeout)
	defer o.Cancel()

	if err := o.authHandler.AuthorizeTransfer(o.ctx, transferProperty.ObjectPushData); err != nil {
		dbh.PublishError(err,
			"OBEX agent error: Transfer was not authorized",
			"error_at", "authpush-agent-authorize",
		)

		return "", o.makeError()
	}

	return path, nil
}

// Cancel is called when the OBEX agent request was cancelled.
func (o *agent) Cancel() *dbus.Error {
	o.ctx.Cancel()

	return nil
}

// Release is called when the OBEX agent is unregistered.
func (o *agent) Release() *dbus.Error {
	return nil
}

// callObexAgentManager calls the OBEX AgentManager1 interface with the provided arguments.
func (o *agent) callObexAgentManager(method string, args ...any) *dbus.Call {
	return o.SessionBus.Object(dbh.ObexBusName, dbh.ObexAgentManagerPath).
		Call(dbh.ObexAgentManagerIface+"."+method, 0, args...)
}
