//go:build !linux

package cmd

import "github.com/urfave/cli/v2"

func getPlatformSpecificFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "alt-library-path",
			EnvVars: []string{"BLUETUITH_LIBRARY_PATH"},
			Usage:   "Specify an alternate path for the 'libhbluetooth' dynamic library.",
		},
		&cli.StringFlag{
			Name:    "alt-daemon-socket-path",
			EnvVars: []string{"BLUETUITH_DAEMON_SOCKET_PATH"},
			Usage:   "Specify an alternate socket path for the 'haraltd' daemon.",
		},
	}
}
