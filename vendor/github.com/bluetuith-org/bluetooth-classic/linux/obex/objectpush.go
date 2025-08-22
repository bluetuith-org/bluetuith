//go:build linux

package obex

import (
	"context"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/fmsg"
	"github.com/Southclaws/fault/ftag"
	bluetooth "github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	errorkinds "github.com/bluetuith-org/bluetooth-classic/api/errorkinds"
	dbh "github.com/bluetuith-org/bluetooth-classic/linux/internal/dbushelper"
	"github.com/godbus/dbus/v5"
)

// fileTransfer describes a file transfer session.
type fileTransfer struct {
	Obex
}

// CreateSession creates a new Obex session with a device.
// The context (ctx) can be provided in case this function call
// needs to be cancelled, since this function call can take some time
// to complete.
func (o *fileTransfer) CreateSession(ctx context.Context) error {
	if err := o.check(); err != nil {
		return err
	}

	var sessionPath dbus.ObjectPath

	args := make(map[string]any, 1)
	args["Target"] = "opp"

	session := o.callClientAsync(ctx, "CreateSession", o.Address.String(), args)
	select {
	case <-ctx.Done():
		return fault.Wrap(
			context.Canceled,
			fctx.With(context.Background(),
				"error_at", "obex-createsession-cancelled",
				"address", o.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Session creation was cancelled"),
		)

	case call := <-session.Done:
		if call.Err != nil {
			return fault.Wrap(
				call.Err,
				fctx.With(context.Background(),
					"error_at", "obex-createsession-methodcall",
					"address", o.Address.String(),
				),
				ftag.With(ftag.Internal),
				fmsg.With("Cannot start a file transfer session"),
			)
		}

		if err := call.Store(&sessionPath); err != nil {
			return fault.Wrap(
				err,
				fctx.With(context.Background(),
					"error_at", "obex-createsession-path",
					"address", o.Address.String(),
				),
				ftag.With(ftag.Internal),
				fmsg.With("Cannot obtain file transfer session data"),
			)
		}
	}

	dbh.PathConverter.AddDbusPath(dbh.DbusPathObexSession, sessionPath, o.Address)

	return nil
}

// RemoveSession removes a created Obex session.
func (o *fileTransfer) RemoveSession() error {
	if err := o.check(); err != nil {
		return err
	}

	sessionPath, ok := dbh.PathConverter.DbusPath(dbh.DbusPathObexSession, o.Address)
	if !ok {
		return fault.Wrap(
			errorkinds.ErrPropertyDataParse,
			fctx.With(context.Background(),
				"error_at", "obex-removesession-path",
				"address", o.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Cannot obtain file transfer session data"),
		)
	}

	if err := o.callClient("RemoveSession", sessionPath).Store(); err != nil {
		return fault.Wrap(
			err,
			fctx.With(context.Background(),
				"error_at", "obex-removesession-methodcall",
				"address", o.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("An error occurred while removing the file transfer session"),
		)
	}

	return nil
}

// SendFile sends a file to the device. The 'filepath' must be a full path to the file.
func (o *fileTransfer) SendFile(filepath string) (bluetooth.ObjectPushData, error) {
	if err := o.check(); err != nil {
		return bluetooth.ObjectPushData{}, err
	}

	var transferPath dbus.ObjectPath
	var fileTransferObject obexTransferProperties

	sessionPath, ok := dbh.PathConverter.DbusPath(dbh.DbusPathObexSession, o.Address)
	if !ok {
		return bluetooth.ObjectPushData{},
			fault.Wrap(
				errorkinds.ErrPropertyDataParse,
				fctx.With(context.Background(),
					"error_at", "obex-sendfile-sessionpath",
					"address", o.Address.String(),
				),
				ftag.With(ftag.Internal),
				fmsg.With("Cannot obtain file transfer session data"),
			)
	}

	transferPropertyMap := make(map[string]dbus.Variant)
	if err := o.callObjectPush(sessionPath, "SendFile", filepath).
		Store(&transferPath, &transferPropertyMap); err != nil {
		return bluetooth.ObjectPushData{},
			fault.Wrap(
				err,
				fctx.With(context.Background(),
					"error_at", "obex-sendfile-methodcall",
					"address", o.Address.String(),
				),
				ftag.With(ftag.Internal),
				fmsg.With("Cannot send file: "+filepath),
			)
	}

	fileTransferObject.appendExtra(transferPath, o.Address)
	dbh.PathConverter.AddDbusPath(dbh.DbusPathObexTransfer, transferPath, o.Address)

	if err := dbh.DecodeVariantMap(transferPropertyMap, &fileTransferObject); err != nil {
		return bluetooth.ObjectPushData{},
			fault.Wrap(
				err,
				fctx.With(context.Background(),
					"error_at", "obex-sendfile-decode",
					"address", o.Address.String(),
				),
				ftag.With(ftag.Internal),
				fmsg.With("Cannot obtain file transfer data"),
			)
	}

	return fileTransferObject.ObjectPushData, nil
}

// CancelTransfer cancels the transfer.
func (o *fileTransfer) CancelTransfer() error {
	if err := o.check(); err != nil {
		return err
	}

	transferPath, ok := dbh.PathConverter.DbusPath(dbh.DbusPathObexTransfer, o.Address)
	if !ok {
		return fault.Wrap(
			errorkinds.ErrPropertyDataParse,
			fctx.With(context.Background(),
				"error_at", "obex-canceltransfer-path",
				"address", o.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Cannot obtain file transfer data"),
		)
	}

	if err := o.callTransfer(transferPath, "Cancel").Store(); err != nil {
		return fault.Wrap(
			err,
			fctx.With(context.Background(),
				"error_at", "obex-canceltransfer-call",
				"address", o.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Cannot cancel transfer"),
		)
	}

	return nil
}

// SuspendTransfer suspends the transfer.
func (o *fileTransfer) SuspendTransfer() error {
	if err := o.check(); err != nil {
		return err
	}

	transferPath, ok := dbh.PathConverter.DbusPath(dbh.DbusPathObexTransfer, o.Address)
	if !ok {
		return fault.Wrap(
			errorkinds.ErrPropertyDataParse,
			fctx.With(context.Background(),
				"error_at", "obex-suspendtransfer-path",
				"address", o.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Cannot obtain file transfer data"),
		)
	}

	if err := o.callTransfer(transferPath, "Suspend").Store(); err != nil {
		return fault.Wrap(
			err,
			fctx.With(context.Background(),
				"error_at", "obex-suspendtransfer-call",
				"address", o.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Cannot suspend transfer"),
		)
	}

	return nil
}

// ResumeTransfer resumes the transfer.
func (o *fileTransfer) ResumeTransfer() error {
	if err := o.check(); err != nil {
		return err
	}

	transferPath, ok := dbh.PathConverter.DbusPath(dbh.DbusPathObexTransfer, o.Address)
	if !ok {
		return fault.Wrap(
			errorkinds.ErrPropertyDataParse,
			fctx.With(context.Background(),
				"error_at", "obex-resumetransfer-path",
				"address", o.Address.String(),
			),
			ftag.With(ftag.NotFound),
			fmsg.With("Cannot obtain file transfer data"),
		)
	}

	if err := o.callTransfer(transferPath, "Resume").Store(); err != nil {
		return fault.Wrap(
			err,
			fctx.With(context.Background(),
				"error_at", "obex-resumetransfer-call",
				"address", o.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Cannot resume transfer"),
		)
	}

	return nil
}

// check checks whether the SessionBus was initialized.
func (o *fileTransfer) check() error {
	if o.SessionBus == nil {
		return fault.Wrap(errorkinds.ErrObexInitSession,
			fctx.With(context.Background(),
				"error_at", "obex-check-sessionbus",
				"address", o.Address.String(),
			),
			ftag.With(ftag.NotFound),
			fmsg.With("Cannot call file transfer method on session-bus"),
		)
	}

	_, ok := dbh.PathConverter.DbusPath(dbh.DbusPathDevice, o.Address)
	if !ok {
		return fault.Wrap(errorkinds.ErrDeviceNotFound,
			fctx.With(context.Background(),
				"error_at", "obex-check-device",
				"address", o.Address.String(),
			),
			ftag.With(ftag.NotFound),
			fmsg.With("Device does not exist"),
		)
	}

	return nil
}
