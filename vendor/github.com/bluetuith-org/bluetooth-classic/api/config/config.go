package config

import (
	"time"
)

const (
	// DefaultAuthTimeout is the default timeout duration for authentication requests.
	DefaultAuthTimeout = 10 * time.Second
)

// Configuration describes a general configuration.
type Configuration struct {
	// SocketPath holds the path to the socket used to interface with the shim.
	SocketPath string

	// AuthTimeout holds the timeout for authentication requests.
	AuthTimeout time.Duration
}

// New returns a new configuration with the default authentication timeout.
func New() Configuration {
	return Configuration{
		AuthTimeout: DefaultAuthTimeout,
	}
}
