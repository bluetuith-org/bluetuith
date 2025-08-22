package bluetooth

// Network describes a function call interface to invoke Network related functions
// on specified devices.
type Network interface {
	// Connect connects to a specific device according to the provided NetworkType
	// and assigns a name to the established connection.
	Connect(name string, nt NetworkType) error

	// Disconnect disconnects from an established connection.
	Disconnect() error
}

// NetworkDunSettings holds the DUN-specific connection settings.
type NetworkDunSettings struct {
	APN    string
	Number string
}

// NetworkType specifies the network type.
type NetworkType string

// The different Bluetooth supported network types.
const (
	NetworkPanu NetworkType = "panu"
	NetworkDun  NetworkType = "dun"
)

// String converts the NetworkType to a string.
func (n NetworkType) String() string {
	return string(n)
}
