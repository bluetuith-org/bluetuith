//go:build !linux

package commands

import (
	"strconv"
	"time"

	"github.com/bluetuith-org/bluetooth-classic/api/appfeatures"
	"github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	"github.com/bluetuith-org/bluetooth-classic/api/errorkinds"
	"github.com/bluetuith-org/bluetooth-classic/api/platforminfo"
	"github.com/bluetuith-org/bluetooth-classic/shim/internal/serde"
	"github.com/google/uuid"
)

// Session commands.
// GetFeatureFlags invokes the "rpc feature-flags" command.
func GetFeatureFlags() *Command[appfeatures.Features] {
	return &Command[appfeatures.Features]{cmd: "rpc feature-flags"}
}

// GetAdapters invokes the "adapter list" command.
func GetAdapters() *Command[[]bluetooth.AdapterData] {
	return &Command[[]bluetooth.AdapterData]{cmd: "adapter list"}
}

// GetPlatformInfo invokes the "rpc platform-info" command.
func GetPlatformInfo() *Command[platforminfo.PlatformInfo] {
	return &Command[platforminfo.PlatformInfo]{cmd: "rpc platform-info"}
}

// AuthenticationReply invokes the "rpc auth" command.
func AuthenticationReply(id int, input string) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "rpc auth"}).WithOptions(func(am OptionMap) {
		am[AuthenticationIdOption] = strconv.FormatInt(int64(id), 10)
		am[ResponseOption] = input
	})
}

// RegisterAgent registers the specified authentication agent with the daemon.
func RegisterAgent(agent RpcAgent) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "rpc agent register"}).WithOption(AgentOption, agent.String())
}

// Adapter commands.
// AdapterProperties invokes the "adapter properties" command.
func AdapterProperties(Address bluetooth.MacAddress) *Command[bluetooth.AdapterData] {
	return (&Command[bluetooth.AdapterData]{cmd: "adapter properties"}).WithOption(AddressOption, Address.String())
}

// GetPairedDevices invokes the "adapter get-paired-devices" command.
func GetPairedDevices(Address bluetooth.MacAddress) *Command[[]bluetooth.DeviceData] {
	return (&Command[[]bluetooth.DeviceData]{cmd: "adapter get-paired-devices"}).WithOption(AddressOption, Address.String())
}

// SetPairableState invokes the "adapter set-pairable-state" command.
func SetPairableState(Address bluetooth.MacAddress, State bool) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "adapter set-pairable-state"}).WithOptions(func(am OptionMap) {
		am[AddressOption] = Address.String()
		am[StateOption] = StateOptionValue(State)
	})
}

// SetDiscoverableState invokes the "adapter set-discoverable-state" command.
func SetDiscoverableState(Address bluetooth.MacAddress, State bool) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "adapter set-discoverable-state"}).WithOptions(func(am OptionMap) {
		am[AddressOption] = Address.String()
		am[StateOption] = StateOptionValue(State)
	})
}

// SetPoweredState invokes the "adapter set-powered-state" command.
func SetPoweredState(Address bluetooth.MacAddress, State bool) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "adapter set-powered-state"}).WithOptions(func(am OptionMap) {
		am[AddressOption] = Address.String()
		am[StateOption] = StateOptionValue(State)
	})
}

// StartDiscovery invokes the "adapter discovery start" command.
func StartDiscovery(Address bluetooth.MacAddress) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "adapter discovery start"}).WithOption(AddressOption, Address.String())
}

// StopDiscovery invokes the "adapter discovery stop" command.
func StopDiscovery(Address bluetooth.MacAddress) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "adapter discovery stop"}).WithOption(AddressOption, Address.String())
}

// Device commands.
// DeviceProperties invokes the "device properties" command.
func DeviceProperties(Address bluetooth.MacAddress) *Command[bluetooth.DeviceData] {
	return (&Command[bluetooth.DeviceData]{cmd: "device properties"}).WithOption(AddressOption, Address.String())
}

// Pair invokes the "device pair" command.
func Pair(Address bluetooth.MacAddress) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "device pair"}).WithOption(AddressOption, Address.String())
}

// CancelPairing invokes the "device pair cancel" command.
func CancelPairing(Address bluetooth.MacAddress) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "device pair cancel"}).WithOption(AddressOption, Address.String())
}

// Connect invokes the "device connect" command.
func Connect(Address bluetooth.MacAddress) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "device connect"}).WithOption(AddressOption, Address.String())
}

// Disconnect invokes the "device disconnect" command.
func Disconnect(Address bluetooth.MacAddress) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "device disconnect"}).WithOption(AddressOption, Address.String())
}

// ConnectProfile invokes the "device connect profile" command.
func ConnectProfile(Address bluetooth.MacAddress, Profile uuid.UUID) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "device connect profile"}).WithOptions(func(am OptionMap) {
		am[AddressOption] = Address.String()
		am[ProfileOption] = Profile.String()
	})
}

// DisconnectProfile invokes the "device disconnect profile" command.
func DisconnectProfile(Address bluetooth.MacAddress, Profile uuid.UUID) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "device disconnect profile"}).WithOptions(func(am OptionMap) {
		am[AddressOption] = Address.String()
		am[ProfileOption] = Profile.String()
	})
}

// Remove invokes the "device remove" command.
func Remove(Address bluetooth.MacAddress) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "device remove"}).WithOption(AddressOption, Address.String())
}

// Obex commands.
// CreateSession invokes the "device opp start-session" command.
func CreateSession(Address bluetooth.MacAddress) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "device opp start-session"}).WithOption(AddressOption, Address.String())
}

// RemoveSession invokes the "device opp stop-session" command.
func RemoveSession(Address bluetooth.MacAddress) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "device opp stop-session"}).WithOption(AddressOption, Address.String())
}

// SendFile invokes the "device opp send-file" command.
func SendFile(Address bluetooth.MacAddress, File string) *Command[bluetooth.ObjectPushData] {
	return (&Command[bluetooth.ObjectPushData]{cmd: "device opp send-file"}).WithOptions(func(am OptionMap) {
		am[AddressOption] = Address.String()
		am[FileOption] = File
	})
}

// CancelTransfer invokes the "device opp cancel-transfer" command.
func CancelTransfer(Address bluetooth.MacAddress) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "device opp cancel-transfer"}).WithOption(AddressOption, Address.String())
}

// SuspendTransfer invokes the "device opp suspend-transfer" command.
func SuspendTransfer(Address bluetooth.MacAddress) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "device opp suspend-transfer"}).WithOption(AddressOption, Address.String())
}

// ResumeTransfer invokes the "device opp resume-transfer" command.
func ResumeTransfer(Address bluetooth.MacAddress) *Command[NoResult] {
	return (&Command[NoResult]{cmd: "device opp resume-transfer"}).WithOption(AddressOption, Address.String())
}

// ExecuteWith invokes a command on the server, and listens for and returns the result of the command invocation.
func (c *Command[T]) ExecuteWith(fn ExecuteFunc, timeoutSeconds ...int) (T, error) {
	var result T

	var timeout = CommandReplyTimeout
	if timeoutSeconds != nil {
		timeout = time.Duration(timeoutSeconds[0] * int(time.Second))
	}

	responseChan, commandErr := fn(c.Slice())
	if commandErr != nil {
		return result, commandErr
	}

	commandErr = errorkinds.ErrSessionStop

	select {
	case response, ok := <-responseChan:
		if !ok {
			break
		}

		if response.Status == "error" {
			return result, response.Error
		}

		if response.Status == "ok" {
			switch any(result).(type) {
			case NoResult:
				return result, nil
			}

			reply := make(map[string]T, 1)
			if err := serde.UnmarshalJson(response.Data, &reply); err != nil {
				return result, err
			}

			for _, mv := range reply {
				result = mv
			}

			commandErr = nil
		}

	case <-time.After(timeout):
		commandErr = errorkinds.ErrMethodTimeout
	}

	return result, commandErr
}
