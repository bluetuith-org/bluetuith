//go:build !linux

package shim

import (
	"context"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/fmsg"
	"github.com/Southclaws/fault/ftag"
	"github.com/bluetuith-org/bluetooth-classic/api/appfeatures"
	"github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	"github.com/bluetuith-org/bluetooth-classic/api/errorkinds"
	"github.com/bluetuith-org/bluetooth-classic/shim/internal/commands"
)

// Obex describes an Obex session.
type obex struct {
	s       *ShimSession
	Address bluetooth.MacAddress
}

// obexObjectPush describes a file transfer session.
type obexObjectPush struct {
	*obex
}

// ObjectPush returns a function call interface to invoke device file transfer
// related functions.
func (o *obex) ObjectPush() bluetooth.ObexObjectPush {
	return &obexObjectPush{o}
}

// CreateSession creates a new Obex session with a device.
// The context (ctx) can be provided in case this function call
// needs to be cancelled, since this function call can take some time
// to complete.
func (o *obexObjectPush) CreateSession(ctx context.Context) error {
	if err := o.check(); err != nil {
		return err
	}

	_, err := commands.CreateSession(o.Address).ExecuteWith(o.s.executor)
	if ctx.Err() == context.Canceled {
		o.RemoveSession()
	}

	return err
}

// RemoveSession removes a created Obex session.
func (o *obexObjectPush) RemoveSession() error {
	if err := o.check(); err != nil {
		return err
	}

	_, err := commands.RemoveSession(o.Address).ExecuteWith(o.s.executor)
	return err
}

// SendFile sends a file to the device. The 'filepath' must be a full path to the file.
func (o *obexObjectPush) SendFile(filepath string) (bluetooth.ObjectPushData, error) {
	if err := o.check(); err != nil {
		return bluetooth.ObjectPushData{}, err
	}

	filetransfer, err := commands.SendFile(o.Address, filepath).ExecuteWith(o.s.executor)

	return filetransfer, err
}

// CancelTransfer cancels the transfer.
func (o *obexObjectPush) CancelTransfer() error {
	if err := o.check(); err != nil {
		return err
	}

	_, err := commands.CancelTransfer(o.Address).ExecuteWith(o.s.executor)
	return err
}

// SuspendTransfer suspends the transfer.
func (o *obexObjectPush) SuspendTransfer() error {
	if err := o.check(); err != nil {
		return err
	}

	_, err := commands.SuspendTransfer(o.Address).ExecuteWith(o.s.executor)
	return err
}

// ResumeTransfer resumes the transfer.
func (o *obexObjectPush) ResumeTransfer() error {
	if err := o.check(); err != nil {
		return err
	}

	_, err := commands.ResumeTransfer(o.Address).ExecuteWith(o.s.executor)
	return err
}

func (o *obexObjectPush) check() error {
	switch {
	case o.s == nil || o.s.sessionClosed.Load():
		return fault.Wrap(errorkinds.ErrSessionNotExist,
			fctx.With(context.Background(),
				"error_at", "obex-check-bus",
				"address", o.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("Error while fetching obex data"),
		)

	case !o.s.features.Has(appfeatures.FeatureSendFile):
		return fault.Wrap(errorkinds.ErrNotSupported,
			fctx.With(context.Background(),
				"error_at", "obex-check-features",
				"address", o.Address.String(),
			),
			ftag.With(ftag.Internal),
			fmsg.With("The provider does not support sending files"),
		)
	}

	return nil
}
