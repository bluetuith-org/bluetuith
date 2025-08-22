package bluetooth

import (
	"github.com/google/uuid"
)

// Device describes a function call interface to invoke device related functions.
type Device interface {
	// Pair will attempt to pair a bluetooth device that is in pairing mode.
	Pair() error

	// CancelPairing will cancel a pairing attempt.
	CancelPairing() error

	// Connect will attempt to connect an already paired bluetooth device
	// to an adapter.
	Connect() error

	// Disconnect will disconnect the bluetooth device from the adapter.
	Disconnect() error

	// ConnectProfile will attempt to connect an already paired bluetooth device
	// to an adapter, using a specific Bluetooth profile UUID .
	ConnectProfile(profileUUID uuid.UUID) error

	// DisconnectProfile will attempt to disconnect an already paired bluetooth device
	// to an adapter, using a specific Bluetooth profile UUID .
	DisconnectProfile(profileUUID uuid.UUID) error

	// Remove removes a device from its associated adapter.
	Remove() error

	// SetTrusted sets the device 'trust' status within its associated adapter.
	// Currently is valid only on Linux.
	SetTrusted(enable bool) error

	// SetBlocked sets the device 'blocked' status within its associated adapter.
	// Currently is valid only on Linux.
	SetBlocked(enable bool) error

	// Properties returns all the properties of the device.
	Properties() (DeviceData, error)
}

// AuthorizeDevicePairing describes an authentication interface, which is used
// to request authentication to pair a device.
type AuthorizeDevicePairing interface {
	DisplayPinCode(timeout AuthTimeout, address MacAddress, pincode string) error
	DisplayPasskey(timeout AuthTimeout, address MacAddress, passkey uint32, entered uint16) error
	ConfirmPasskey(timeout AuthTimeout, address MacAddress, passkey uint32) error
	AuthorizePairing(timeout AuthTimeout, address MacAddress) error
	AuthorizeService(timeout AuthTimeout, address MacAddress, serviceUUID uuid.UUID) error
}

// DeviceData holds the static bluetooth device information installed for a system.
type DeviceData struct {
	// Name holds the name of the device.
	Name string `json:"name,omitempty" codec:"Name,omitempty" doc:"The name of the device."`

	// Class holds the device type class specifier.
	Class uint32 `json:"class,omitempty" codec:"Class,omitempty" doc:"The device type class specifier."`

	// Type holds the type name of the device.
	// For example, type of the device can be "Phone", "Headset" etc.
	Type string `json:"type,omitempty" codec:"Type,omitempty" doc:"The type name of the device. For example, type of the device can be 'Phone', 'Headset' etc."`

	// Alias holds the optional or user-assigned name for the adapter.
	// Usually valid for Linux systems, may be empty or equate to "Name"
	// for other systems.
	Alias string `json:"alias,omitempty" codec:"Alias,omitempty" doc:"The optional or user-assigned name for the adapter. Usually valid for Linux systems, may be empty or equate to **name** for other systems."`

	// LegacyPairing indicates whether the device only supports the pre-2.1 pairing mechanism.
	// This property is useful during device discovery to anticipate whether
	// legacy or simple pairing will occur if pairing is initiated.
	LegacyPairing bool `json:"legacy_pairing,omitempty" codec:"LegacyPairing,omitempty" doc:"Indicates whether the device only supports the pre-2.1 pairing mechanism. This property is useful during device discovery to anticipate whether legacy or simple pairing will occur if pairing is initiated."`

	DeviceEventData
}

// DeviceEventData holds the dynamic (variable) bluetooth device information.
// This is primarily used to send device event related data.
type DeviceEventData struct {
	// Address holds the Bluetooth MAC address of the device.
	Address MacAddress `json:"address,omitempty" codec:"Address,omitempty" doc:"The Bluetooth MAC address of the device."`

	// AssociatedAdapter holds the Bluetooth MAC address of the adapter
	// the device is associated with.
	AssociatedAdapter MacAddress `json:"associated_adapter,omitempty" codec:"AssociatedAdapter,omitempty" doc:"The Bluetooth MAC address of the adapter the device is associated with."`

	// Paired indicates if the device is paired.
	Paired bool `json:"paired,omitempty" codec:"Paired,omitempty" doc:"Indicates if the device is paired."`

	// Connected indicates if the device is connected.
	Connected bool `json:"connected,omitempty" codec:"Connected,omitempty" doc:"Indicates if the device is connected."`

	// Trusted indicates if the device is marked as trusted.
	// Valid only on Linux systems, will equate to "true"
	// on other systems if the device is paired.
	Trusted bool `json:"trusted,omitempty" codec:"Trusted,omitempty" doc:"Indicates if the device is marked as trusted. Valid only on Linux systems, will equate to 'true' on other systems if the device is paired."`

	// Blocked indicates if the device is marked as blocked.
	// Valid only on Linux systems, will equate to "false"
	// on other systems.
	Blocked bool `json:"blocked,omitempty" codec:"Blocked,omitempty" doc:"Indicates if the device is marked as blocked. Valid only on Linux systems, will equate to 'false' on other systems."`

	// Bonded indicates if the device is bonded.
	Bonded bool `json:"bonded,omitempty" codec:"Bonded,omitempty" doc:"Indicates if the device is bonded."`

	// RSSI indicates the signal strength of the device.
	RSSI int16 `json:"rssi,omitempty" codec:"RSSI,omitempty" doc:"Indicates the signal strength of the device."`

	// Percentage holds the battery percentage of the device.
	Percentage int `json:"percentage,omitempty" codec:"Percentage,omitempty" minimum:"0" maximum:"100" doc:"The battery percentage of the device."`

	// UUIDs holds the device-supported Bluetooth profile UUIDs.
	UUIDs []string `json:"uuids,omitempty" codec:"UUIDs,omitempty" doc:"The device-supported Bluetooth profile UUIDs."`
}

// HaveService returns if the device advertises a specific service (Bluetooth profile).
func (d *DeviceData) HaveService(service uint32) bool {
	return ServiceExists(d.UUIDs, service)
}

// DeviceTypeFromClass parses the device class and returns its type.
//
//gocyclo:ignore
func DeviceTypeFromClass(class uint32) string {
	/*
		Adapted from:
		https://gitlab.freedesktop.org/upower/upower/-/blob/master/src/linux/up-device-bluez.c#L64
	*/
	switch (class & 0x1f00) >> 8 {
	case 0x01:
		return "Computer"

	case 0x02:
		switch (class & 0xfc) >> 2 {
		case 0x01, 0x02, 0x03, 0x05:
			return "Phone"

		case 0x04:
			return "Modem"
		}

	case 0x03:
		return "Network"

	case 0x04:
		switch (class & 0xfc) >> 2 {
		case 0x01, 0x02:
			return "Headset"

		case 0x05:
			return "Speakers"

		case 0x06:
			return "Headphones"

		case 0x0b, 0x0c, 0x0d:
			return "Video"

		default:
			return "Audio device"
		}

	case 0x05:
		switch (class & 0xc0) >> 6 {
		case 0x00:
			switch (class & 0x1e) >> 2 {
			case 0x01, 0x02:
				return "Gaming input"

			case 0x03:
				return "Remote control"
			}

		case 0x01:
			return "Keyboard"

		case 0x02:
			switch (class & 0x1e) >> 2 {
			case 0x05:
				return "Tablet"

			default:
				return "Mouse"
			}
		}

	case 0x06:
		if (class & 0x80) > 0 {
			return "Printer"
		}

		if (class & 0x40) > 0 {
			return "Scanner"
		}

		if (class & 0x20) > 0 {
			return "Camera"
		}

		if (class & 0x10) > 0 {
			return "Monitor"
		}

	case 0x07:
		return "Wearable"

	case 0x08:
		return "Toy"
	}

	return "Unknown"
}
