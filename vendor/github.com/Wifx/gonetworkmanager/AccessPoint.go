package gonetworkmanager

import (
	"encoding/json"

	"github.com/godbus/dbus/v5"
)

const (
	AccessPointInterface = NetworkManagerInterface + ".AccessPoint"

	/* Properties */
	AccessPointPropertyFlags      = AccessPointInterface + ".Flags"      // readable   u
	AccessPointPropertyWpaFlags   = AccessPointInterface + ".WpaFlags"   // readable   u
	AccessPointPropertyRsnFlags   = AccessPointInterface + ".RsnFlags"   // readable   u
	AccessPointPropertySsid       = AccessPointInterface + ".Ssid"       // readable   ay
	AccessPointPropertyFrequency  = AccessPointInterface + ".Frequency"  // readable   u
	AccessPointPropertyHwAddress  = AccessPointInterface + ".HwAddress"  // readable   s
	AccessPointPropertyMode       = AccessPointInterface + ".Mode"       // readable   u
	AccessPointPropertyMaxBitrate = AccessPointInterface + ".MaxBitrate" // readable   u
	AccessPointPropertyStrength   = AccessPointInterface + ".Strength"   // readable   y
	AccessPointPropertyLastSeen   = AccessPointInterface + ".LastSeen"   // readable   i
)

type AccessPoint interface {
	GetPath() dbus.ObjectPath

	// GetPropertyFlags gets flags describing the capabilities of the access point.
	GetPropertyFlags() (uint32, error)

	// GetPropertyWPAFlags gets flags describing the access point's capabilities
	// according to WPA (Wi-Fi Protected Access).
	GetPropertyWPAFlags() (uint32, error)

	// GetPropertyRSNFlags gets flags describing the access point's capabilities
	// according to the RSN (Robust Secure Network) protocol.
	GetPropertyRSNFlags() (uint32, error)

	// GetPropertySSID returns the Service Set Identifier identifying the access point.
	GetPropertySSID() (string, error)

	// GetPropertyFrequency gets the radio channel frequency in use by the access point, in MHz.
	GetPropertyFrequency() (uint32, error)

	// GetPropertyHWAddress gets the hardware address (BSSID) of the access point.
	GetPropertyHWAddress() (string, error)

	// GetPropertyMode describes the operating mode of the access point.
	GetPropertyMode() (Nm80211Mode, error)

	// GetPropertyMaxBitrate gets the maximum bitrate this access point is capable of, in kilobits/second (Kb/s).
	GetPropertyMaxBitrate() (uint32, error)

	// GetPropertyStrength gets the current signal quality of the access point, in percent.
	GetPropertyStrength() (uint8, error)

	// GetPropertyLastSeen The timestamp (in CLOCK_BOOTTIME seconds) for the last  time the access point was found in scan results.
	// A value of -1 means the access point has never been found in scan results.
	GetPropertyLastSeen() (int32, error)

	MarshalJSON() ([]byte, error)
}

func NewAccessPoint(objectPath dbus.ObjectPath) (AccessPoint, error) {
	var a accessPoint
	return &a, a.init(NetworkManagerInterface, objectPath)
}

type accessPoint struct {
	dbusBase
}

func (a *accessPoint) GetPath() dbus.ObjectPath {
	return a.obj.Path()
}

func (a *accessPoint) GetPropertyFlags() (uint32, error) {
	return a.getUint32Property(AccessPointPropertyFlags)
}

func (a *accessPoint) GetPropertyWPAFlags() (uint32, error) {
	return a.getUint32Property(AccessPointPropertyWpaFlags)
}

func (a *accessPoint) GetPropertyRSNFlags() (uint32, error) {
	return a.getUint32Property(AccessPointPropertyRsnFlags)
}

func (a *accessPoint) GetPropertySSID() (string, error) {
	r, err := a.getSliceByteProperty(AccessPointPropertySsid)
	if err != nil {
		return "", err
	}
	return string(r), nil
}

func (a *accessPoint) GetPropertyFrequency() (uint32, error) {
	return a.getUint32Property(AccessPointPropertyFrequency)
}

func (a *accessPoint) GetPropertyHWAddress() (string, error) {
	return a.getStringProperty(AccessPointPropertyHwAddress)
}

func (a *accessPoint) GetPropertyMode() (Nm80211Mode, error) {
	r, err := a.getUint32Property(AccessPointPropertyMode)
	if err != nil {
		return Nm80211ModeUnknown, err
	}
	return Nm80211Mode(r), nil
}

func (a *accessPoint) GetPropertyMaxBitrate() (uint32, error) {
	return a.getUint32Property(AccessPointPropertyMaxBitrate)
}

func (a *accessPoint) GetPropertyStrength() (uint8, error) {
	return a.getUint8Property(AccessPointPropertyStrength)
}

func (a *accessPoint) GetPropertyLastSeen() (int32, error) {
	return a.getInt32Property(AccessPointPropertyLastSeen)
}

func (a *accessPoint) MarshalJSON() ([]byte, error) {
	Flags, err := a.GetPropertyFlags()
	if err != nil {
		return nil, err
	}
	WPAFlags, err := a.GetPropertyWPAFlags()
	if err != nil {
		return nil, err
	}
	RSNFlags, err := a.GetPropertyRSNFlags()
	if err != nil {
		return nil, err
	}
	SSID, err := a.GetPropertySSID()
	if err != nil {
		return nil, err
	}
	Frequency, err := a.GetPropertyFrequency()
	if err != nil {
		return nil, err
	}
	HWAddress, err := a.GetPropertyHWAddress()
	if err != nil {
		return nil, err
	}
	Mode, err := a.GetPropertyMode()
	if err != nil {
		return nil, err
	}
	MaxBitrate, err := a.GetPropertyMaxBitrate()
	if err != nil {
		return nil, err
	}
	Strength, err := a.GetPropertyStrength()
	if err != nil {
		return nil, err
	}
	LastSeen, err := a.GetPropertyLastSeen()
	if err != nil {
		return nil, err
	}

	return json.Marshal(map[string]interface{}{
		"Flags":      Flags,
		"WPAFlags":   WPAFlags,
		"RSNFlags":   RSNFlags,
		"SSID":       SSID,
		"Frequency":  Frequency,
		"HWAddress":  HWAddress,
		"Mode":       Mode.String(),
		"MaxBitrate": MaxBitrate,
		"Strength":   Strength,
		"LastSeen":   LastSeen,
	})
}
