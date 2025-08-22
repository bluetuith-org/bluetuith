//go:build linux

package dbushelper

import (
	"context"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/fmsg"
	"github.com/Southclaws/fault/ftag"
	bluetooth "github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	errorkinds "github.com/bluetuith-org/bluetooth-classic/api/errorkinds"
	"github.com/godbus/dbus/v5"
)

// PublishSignalError publishes an error message with DBus signal data to the error event stream.
func PublishSignalError(err error, signal *dbus.Signal, message string, metadata ...string) {
	bluetooth.ErrorEvents().PublishAdded(wrapSignalErrors(err, signal, message, metadata...))
}

// PublishError publishes an error to the error event stream.
func PublishError(err error, message string, metadata ...string) {
	bluetooth.ErrorEvents().PublishAdded(errorkinds.GenericError{
		Errors: fault.Wrap(err,
			fctx.With(context.Background(), metadata...),
			ftag.With(ftag.Internal),
			fmsg.With(message),
		),
	})
}

// wrapSignalErrors returns an ErrorData after wrapping the provided error and signal related data.
func wrapSignalErrors(err error, signal *dbus.Signal, message string, metadata ...string) errorkinds.GenericError {
	md := append([]string{"signal-name", signal.Name, "signal-path", string(signal.Path)}, metadata...)

	return errorkinds.GenericError{
		Errors: fault.Wrap(err,
			fctx.With(context.Background(), md...),
			ftag.With(ftag.Internal),
			fmsg.With(message),
		),
	}
}
