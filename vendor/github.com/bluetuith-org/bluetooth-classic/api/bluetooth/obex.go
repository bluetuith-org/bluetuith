package bluetooth

import (
	"context"
)

// Obex describes a function call interface to invoke Obex related functions
// on specified devices.
type Obex interface {
	// ObjectPush returns a function call interface to invoke device file transfer
	// related functions.
	ObjectPush() ObexObjectPush
}

// ObexObjectPush describes a function call interface to manage file-transfer
// related functions on specified devices.
type ObexObjectPush interface {
	// CreateSession creates a new Obex session with a device.
	// The context (ctx) can be provided in case this function call
	// needs to be cancelled, since this function call can take some time
	// to complete.
	CreateSession(ctx context.Context) error

	// RemoveSession removes a created Obex session.
	RemoveSession() error

	// SendFile sends a file to the device. The 'filepath' must be a full path to the file.
	SendFile(filepath string) (ObjectPushData, error)

	// CancelTransfer cancels the transfer.
	CancelTransfer() error

	// SuspendTransfer suspends the transfer.
	SuspendTransfer() error

	// ResumeTransfer resumes the transfer.
	ResumeTransfer() error
}

// ObjectPushStatus describes the status of the file transfer.
type ObjectPushStatus string

// The different transfer status types.
const (
	TransferQueued    ObjectPushStatus = "queued"
	TransferActive    ObjectPushStatus = "active"
	TransferSuspended ObjectPushStatus = "suspended"
	TransferComplete  ObjectPushStatus = "complete"
	TransferError     ObjectPushStatus = "error"
)

type (
	objectPushID string

	// ObjectPushSessionID represents a session ID of an Object push object.
	ObjectPushSessionID objectPushID

	// ObjectPushTransferID represents a transfer ID of an Object Push object.
	ObjectPushTransferID objectPushID
)

// String returns a string version of the session ID.
func (o ObjectPushSessionID) String() string {
	return string(o)
}

// String returns a string version of the transfer ID.
func (o ObjectPushTransferID) String() string {
	return string(o)
}

// ObjectPushData holds the static file transfer data for a device.
type ObjectPushData struct {
	// Name is the name of the object being transferred.
	Name string `json:"name,omitempty" codec:"Name,omitempty" doc:"The name of the object being transferred."`

	// Type is the type of the file (mime-type).
	Type string `json:"type,omitempty" codec:"Type,omitempty" doc:"The type of the file (mime-type)."`

	// Filename is the complete name of the file.
	Filename string `json:"filename,omitempty" codec:"Filename,omitempty" doc:"The complete name of the file."`

	// Receiving specifies whether this transfer is being received or not.
	Receiving bool `json:"receiving,omitempty" codec:"" doc:"Specifies whether this transfer is being received or not."`

	ObjectPushEventData
}

// ObjectPushEventData holds the dynamic (variable) file transfer data for a device.
// This is primarily used to send file transfer event related data.
type ObjectPushEventData struct {
	// Address holds the Bluetooth MAC address of the device.
	Address MacAddress `json:"address,omitempty" codec:"Address,omitempty" doc:"The Bluetooth MAC address of the device."`

	// Status indicates the file transfer status.
	Status ObjectPushStatus `json:"status,omitempty" codec:"Status,omitempty" enum:"queued,active,suspended,complete,error" doc:"Indicates the file transfer status."`

	// Size holds the total size of the file in bytes.
	Size uint64 `json:"size,omitempty" codec:"Size,omitempty" doc:"The total size of the file in bytes."`

	// Transferred holds the current number of bytes that was sent to the receiver.
	Transferred uint64 `json:"transferred,omitempty" codec:"Transferred,omitempty" doc:"The current number of bytes that was sent to the receiver."`

	// TransferID holds the ID of the transferring item.
	TransferID ObjectPushTransferID `json:"transfer_id,omitempty" codec:"" doc:"The ID of the transferring item."`

	// SessionID holds the ID of the session that this transferring item belongs to.
	SessionID ObjectPushSessionID `json:"session_id,omitempty" codec:"Session,omitempty" doc:"The ID of the session that this transferring item belongs to."`
}

// AuthorizeReceiveFile describes an authentication interface, which is used
// to authorize a file transfer being received, before starting the transfer.
type AuthorizeReceiveFile interface {
	AuthorizeTransfer(timeout AuthTimeout, props ObjectPushData) error
}
