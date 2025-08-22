package errorkinds

import "errors"

// The different general error types.
var (
	ErrSessionStart    = errors.New("cannot start session")
	ErrSessionStop     = errors.New("cannot stop session")
	ErrSessionNotExist = errors.New("session does not exist")
	ErrMethodCall      = errors.New("cannot call method")
	ErrMethodCanceled  = errors.New("method call was cancelled")
	ErrMethodTimeout   = errors.New("timeout on method response")

	ErrInvalidAddress  = errors.New("invalid Bluetooth address")
	ErrAdapterNotFound = errors.New("adapter not found")
	ErrDeviceNotFound  = errors.New("device not found")

	ErrObexInitSession    = errors.New("obex session is not initialized")
	ErrNetworkInitSession = errors.New("network session is not initialized")

	ErrNetworkAlreadyActive  = errors.New("network is already active")
	ErrNetworkEstablishError = errors.New("network connection cannot be established")

	ErrMediaPlayerNotConnected = errors.New("media player is not connected")

	ErrPropertyDataParse = errors.New("error parsing property data")
	ErrEventDataParse    = errors.New("error parsing event data")

	ErrNotSupported = errors.New("this functionality is not supported")
)

// GenericError represents a standard error message.
type GenericError struct {
	// Errors stores all associated errors.
	Errors error `json:"errors,omitempty" doc:"A set of generic errors."`
}

// Error returns the formatted error as string.
func (e GenericError) Error() string {
	return e.Errors.Error()
}

// Unwrap unwraps all errors associated with this error.
func (e GenericError) Unwrap() error {
	return e.Errors
}
