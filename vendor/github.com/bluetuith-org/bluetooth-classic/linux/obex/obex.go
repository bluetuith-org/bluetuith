//go:build linux

package obex

import (
	"context"
	"errors"
	"path/filepath"
	"time"

	"github.com/godbus/dbus/v5"

	ac "github.com/bluetuith-org/bluetooth-classic/api/appfeatures"
	bluetooth "github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	errorkinds "github.com/bluetuith-org/bluetooth-classic/api/errorkinds"
	dbh "github.com/bluetuith-org/bluetooth-classic/linux/internal/dbushelper"
)

// Obex describes a Bluez Obex session.
type Obex struct {
	SessionBus *dbus.Conn
	Address    bluetooth.MacAddress
}

// ObexManager holds an OBEX session and agent.
//
//revive:disable
type ObexManager struct {
	agent *agent

	Obex
}

//revive:enable

// NewManager returns a new ObexManager.
func NewManager(SessionBus *dbus.Conn) *ObexManager {
	return &ObexManager{
		nil, Obex{SessionBus: SessionBus},
	}
}

// obexSessionProperties holds properties for a created Obex session.
type obexSessionProperties struct {
	Root        string
	Target      string
	Source      string
	Destination bluetooth.MacAddress
}

// obexTransferProperties holds the properties for a created Obex transfer.
type obexTransferProperties struct {
	bluetooth.ObjectPushData
}

// Initialize attempts to initialize the Obex Agent, and returns the capabilities of the
// obex session.
func (o *ObexManager) Initialize(auth bluetooth.AuthorizeReceiveFile, authTimeout time.Duration) (ac.Features, *ac.Error) {
	var capabilities ac.Features

	serviceNames, err := dbh.ListActivatableBusNames(o.SessionBus)
	if err != nil {
		return capabilities,
			ac.NewError(ac.FeatureSendFile|ac.FeatureReceiveFile, err)
	}

	for _, name := range serviceNames {
		if name == dbh.ObexBusName {
			goto SetupAgent
		}
	}

	return capabilities,
		ac.NewError(
			ac.FeatureSendFile|ac.FeatureReceiveFile,
			errors.New("OBEX Service does not exist"),
		)

SetupAgent:
	go o.watchObexSessionBus()

	capabilities = ac.FeatureSendFile

	o.agent = newAgent(auth, authTimeout, &fileTransfer{Obex{SessionBus: o.SessionBus}})
	if err := o.agent.setup(); err != nil {
		return capabilities,
			ac.NewError(ac.FeatureReceiveFile, err)
	}

	capabilities |= ac.FeatureReceiveFile

	return capabilities, nil
}

// Stop removes the obex agent and closes the obex session.
func (o *ObexManager) Stop() error {
	return o.agent.remove()
}

// ObjectPush returns a function call interface to invoke device file transfer
// related functions.
func (o *Obex) ObjectPush() bluetooth.ObexObjectPush {
	return &fileTransfer{Obex{SessionBus: o.SessionBus, Address: o.Address}}
}

// watchObexSessionBus will register a signal and watch for events from the OBEX DBus interface.
func (o *ObexManager) watchObexSessionBus() {
	signalMatch := "type='signal', sender='org.bluez.obex'"
	o.SessionBus.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, signalMatch)

	ch := make(chan *dbus.Signal, 10)
	o.SessionBus.Signal(ch)

	for signal := range ch {
		o.parseSignalData(signal)
	}
}

// parseSignalData parses OBEX DBus signal data.
func (o *ObexManager) parseSignalData(signal *dbus.Signal) {
	switch signal.Name {
	case dbh.DbusSignalInterfacesAddedIface:
		objectPath, ok := signal.Body[0].(dbus.ObjectPath)
		if !ok {
			return
		}

		nestedPropertyMap, ok := signal.Body[1].(map[string]map[string]dbus.Variant)
		if !ok {
			return
		}

		for iftype := range nestedPropertyMap {
			switch iftype {
			case dbh.ObexSessionIface:
			case dbh.ObexTransferIface:
				var props obexTransferProperties
				if err := dbh.DecodeVariantMap(nestedPropertyMap[iftype], &props.ObjectPushData); err != nil {
					continue
				}

				sessionProps, err := o.sessionProperties(dbus.ObjectPath(props.SessionID))
				if err != nil {
					continue
				}

				dbh.PathConverter.AddDbusPath(dbh.DbusPathObexSession, dbus.ObjectPath(props.SessionID), sessionProps.Destination)
				dbh.PathConverter.AddDbusPath(dbh.DbusPathObexTransfer, dbus.ObjectPath(objectPath), sessionProps.Destination)

				if props.Filename != "" {
					props.appendExtra(objectPath, sessionProps.Destination)
					bluetooth.ObjectPushEvents().PublishAdded(props.ObjectPushData)
				}
			}
		}

	case dbh.DbusSignalPropertyChangedIface:
		objectInterfaceName, ok := signal.Body[0].(string)
		if !ok {
			return
		}

		propertyMap, ok := signal.Body[1].(map[string]dbus.Variant)
		if !ok {
			return
		}

		switch objectInterfaceName {
		case dbh.ObexSessionIface:
		case dbh.ObexTransferIface:
			address, ok := dbh.PathConverter.Address(dbh.DbusPathObexTransfer, signal.Path)

			if !ok {
				dbh.PublishSignalError(errorkinds.ErrDeviceNotFound, signal,
					"Obex event handler error",
					"error_at", "pchanged-obex-address",
				)

				return
			}

			transferData := obexTransferProperties{}
			transferData.appendExtra(signal.Path, address)

			if err := dbh.DecodeVariantMap(
				propertyMap, &transferData,
				"Status", "Transferred",
			); err != nil {
				dbh.PublishSignalError(err, signal,
					"Obex event handler error",
					"error_at", "pchanged-obex-decode",
				)

				return
			}

			bluetooth.ObjectPushEvents().PublishUpdated(transferData.ObjectPushEventData)
		}

	case dbh.DbusSignalInterfacesRemovedIface:
		objectPath, ok := signal.Body[0].(dbus.ObjectPath)
		if !ok {
			return
		}

		ifaceNames, ok := signal.Body[1].([]string)
		if !ok {
			return
		}

		for _, ifaceName := range ifaceNames {
			switch ifaceName {
			case dbh.ObexSessionIface:
				dbh.PathConverter.RemoveDbusPath(dbh.DbusPathObexSession, objectPath)

			case dbh.ObexTransferIface:
				address, ok := dbh.PathConverter.Address(dbh.DbusPathObexTransfer, objectPath)
				if !ok {
					dbh.PublishSignalError(errorkinds.ErrDeviceNotFound, signal,
						"Obex event handler error",
						"error_at", "premoved-obex-address",
					)

					return
				}

				var props obexTransferProperties
				props.appendExtra(objectPath, address)

				bluetooth.ObjectPushEvents().PublishRemoved(props.ObjectPushEventData)

				dbh.PathConverter.RemoveDbusPath(dbh.DbusPathObexTransfer, objectPath)
			}
		}
	}
}

// callClient calls the Client1 interface with the provided method.
func (o *Obex) callClient(method string, args ...any) *dbus.Call {
	return o.SessionBus.Object(dbh.ObexBusName, dbh.ObexBusPath).
		Call(dbh.ObexClientIface+"."+method, 0, args...)
}

// callClientAsync calls the Client1 interface asynchronously with the provided method.
func (o *Obex) callClientAsync(ctx context.Context, method string, args ...any) *dbus.Call {
	return o.SessionBus.Object(dbh.ObexBusName, dbh.ObexBusPath).
		GoWithContext(ctx, dbh.ObexClientIface+"."+method, 0, nil, args...)
}

// callObjectPush calls the ObjectPush1 interface with the provided method.
func (o *Obex) callObjectPush(sessionPath dbus.ObjectPath, method string, args ...any) *dbus.Call {
	return o.SessionBus.Object(dbh.ObexBusName, sessionPath).
		Call(dbh.ObexObjectPushIface+"."+method, 0, args...)
}

// callTransfer calls the Transfer1 interface with the provided method.
func (o *Obex) callTransfer(transferPath dbus.ObjectPath, method string, args ...any) *dbus.Call {
	return o.SessionBus.Object(dbh.ObexBusName, transferPath).
		Call(dbh.ObexTransferIface+"."+method, 0, args...)
}

// sessionProperties converts a map of OBEX session properties to ObexSessionProperties.
func (o *Obex) sessionProperties(sessionPath dbus.ObjectPath) (obexSessionProperties, error) {
	var sessionProperties obexSessionProperties

	props := make(map[string]dbus.Variant)
	if err := o.SessionBus.Object(dbh.ObexBusName, sessionPath).
		Call(dbh.DbusGetAllPropertiesIface, 0, dbh.ObexSessionIface).
		Store(&props); err != nil {
		return obexSessionProperties{}, err
	}

	return sessionProperties, dbh.DecodeVariantMap(props, &sessionProperties)
}

// transferProperties converts a map of OBEX transfer properties to ObjectPushData.
func (o *Obex) transferProperties(transferPath dbus.ObjectPath) (obexTransferProperties, error) {
	var transferProperties obexTransferProperties

	props := make(map[string]dbus.Variant)
	if err := o.SessionBus.Object(dbh.ObexBusName, transferPath).
		Call(dbh.DbusGetAllPropertiesIface, 0, dbh.ObexTransferIface).
		Store(&props); err != nil {
		return obexTransferProperties{}, err
	}

	return *transferProperties.appendExtra(transferPath, bluetooth.MacAddress{}), dbh.DecodeVariantMap(props, &transferProperties)
}

// appendExtra appends extra properties to the transfer item.
func (t *obexTransferProperties) appendExtra(transferPath dbus.ObjectPath, address bluetooth.MacAddress, receiving ...struct{}) *obexTransferProperties {
	t.TransferID = bluetooth.ObjectPushTransferID(string(transferPath))
	t.SessionID = bluetooth.ObjectPushSessionID(filepath.Dir(t.TransferID.String()))
	if !address.IsNil() {
		t.Address = address
	}

	t.Receiving = receiving != nil

	return t
}
