package bluetooth

// Taken from: https://github.com/tinygo-org/bluetooth/blob/release/mac.go

import (
	"bytes"

	"github.com/bluetuith-org/bluetooth-classic/api/errorkinds"
)

// MacAddress represents a Bluetooth address.
type MacAddress [NumAddressBytes]byte

const (
	// MaxAddressStringLength is the maximum length of a Bluetooth address string (with ':').
	MaxAddressStringLength = 17

	// MaxAddressBytesLength is the maximum length of a Bluetooth address string
	// as a byte array.
	MaxAddressBytesLength = 11

	// NumAddressBytes is the total number of bytes in a MacAddress byte array.
	NumAddressBytes = 6
)

// ParseMAC parses the given MAC address, which must be in 11:22:33:AA:BB:CC
// format. If it cannot be parsed, an error is returned.
func ParseMAC(s string) (mac MacAddress, err error) {
	return parseMacFromBuffer(bytes.NewBufferString(s))
}

// String returns a human-readable version of this MAC address, such as
// 11:22:33:AA:BB:CC.
func (m *MacAddress) String() string {
	return m.byteBuffer().String()
}

// IsNil checks if the MacAddress byte array is empty.
func (m *MacAddress) IsNil() bool {
	var numZeros int

	for _, b := range m {
		if b == 0 {
			numZeros++
		}
	}

	return numZeros == NumAddressBytes
}

// MarshalText implements encoding.TextMarshaler.
// This is never called within go-codec, it is defined to
// implement the TextMarshaler and TextUnmarshaler interfaces
// so that go-codec calls the (*MacAddress).UnmarshalText method
// to decode a Bluetooth address string to a MacAddress.
func (m MacAddress) MarshalText() (data []byte, err error) {
	return m.byteBuffer().Bytes(), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
// This is mainly used to unmarshal values within go-codec.
// For example, mapping a formatted Bluetooth address string
// to a MacAddress within a struct.
func (m *MacAddress) UnmarshalText(data []byte) error {
	mac, err := parseMacFromBuffer(bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	*m = mac

	return nil
}

// byteBuffer returns a byte buffer with the string representation of the MacAddress.
func (m *MacAddress) byteBuffer() *bytes.Buffer {
	s := bytes.NewBuffer(make([]byte, 0, MaxAddressStringLength))

	for i := 5; i >= 0; i-- {
		c := m[i]
		// Insert a hyphen at the correct locations.
		if i != 5 {
			s.WriteString(":")
		}

		// First nibble.
		nibble := c >> 4
		if nibble <= 9 {
			s.WriteByte(nibble + '0')
		} else {
			s.WriteByte(nibble + 'A' - 10)
		}

		// Second nibble.
		nibble = c & 0x0f
		if nibble <= 9 {
			s.WriteByte(nibble + '0')
		} else {
			s.WriteByte(nibble + 'A' - 10)
		}
	}

	return s
}

// parseMacFromBuffer parses a Bluetooth address string from a byte buffer.
func parseMacFromBuffer(b *bytes.Buffer) (MacAddress, error) {
	var mac MacAddress

	macIndex := MaxAddressBytesLength

	for {
		c, err := b.ReadByte()
		if err != nil {
			break
		}

		if c == ':' {
			continue
		}

		var nibble byte
		switch {
		case c >= '0' && c <= '9':
			nibble = c - '0' + 0x0
		case c >= 'A' && c <= 'F':
			nibble = c - 'A' + 0xA
		case c >= 'a' && c <= 'f':
			nibble = c - 'a' + 0xA
		default:
			return mac, errorkinds.ErrInvalidAddress
		}

		if macIndex < -1 {
			return mac, errorkinds.ErrInvalidAddress
		}

		if macIndex%2 == 0 {
			mac[macIndex/2] |= nibble
		} else {
			mac[macIndex/2] |= nibble << 4
		}

		macIndex--
	}

	if macIndex != -1 {
		return mac, errorkinds.ErrInvalidAddress
	}

	return mac, nil
}
