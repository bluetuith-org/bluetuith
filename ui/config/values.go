package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	"github.com/darkhz/bluetuith/ui/keybindings"
	"github.com/darkhz/bluetuith/ui/theme"
)

// Values describes the possible configuration values that a user can
// modify and supply to the application.
type Values struct {
	Adapter       string            `koanf:"adapter"`
	ReceiveDir    string            `koanf:"receive-dir"`
	GsmApn        string            `koanf:"gsm-apn"`
	GsmNumber     string            `koanf:"gsm-number"`
	AdapterStates string            `koanf:"adapter-states"`
	ConnectAddr   string            `koanf:"connect-bdaddr"`
	NoWarning     bool              `koanf:"no-warning"`
	NoHelpDisplay bool              `koanf:"no-help-display"`
	ConfirmOnQuit bool              `koanf:"confirm-on-quit"`
	Theme         map[string]string `koanf:"theme"`
	Keybindings   map[string]string `koanf:"keybindings"`

	AdapterStatesMap      map[string]string
	SelectedAdapter       *bluetooth.AdapterData
	AutoConnectDeviceAddr bluetooth.MacAddress
	Kb                    *keybindings.Keybindings
}

// validateValues validates all configuration values.
func (v *Values) validateValues() error {
	for _, validate := range []func() error{
		v.validateKeybindings,
		v.validateAdapterStates,
		v.validateConnectBDAddr,
		v.validateReceiveDir,
		v.validateGsm,
		v.validateTheme,
	} {
		if err := validate(); err != nil {
			return err
		}
	}

	return nil
}

// validateSessionValues validates all configuration values that require a bluetooth session.
func (v *Values) validateSessionValues(session bluetooth.Session) error {
	for _, validate := range []func(bluetooth.Session) error{
		v.validateAdapter,
		v.validateDeviceExists,
	} {
		if err := validate(session); err != nil {
			return err
		}
	}

	return nil
}

// validateAdapter validates if the adapter specified by the user exists in the system.
func (v *Values) validateAdapter(session bluetooth.Session) error {
	adapters, err := session.Adapters()
	if err != nil {
		return fmt.Errorf("no adapters were found: %w", err)
	}

	if v.Adapter == "" {
		v.SelectedAdapter = &adapters[0]
		return nil
	}

	for _, adapter := range adapters {
		if adapter.UniqueName == v.Adapter || adapter.Address.String() == v.Adapter {
			v.SelectedAdapter = &adapter
			return nil
		}
	}

	return fmt.Errorf("%s: The adapter does not exist", v.Adapter)
}

// validateDeviceExists validates if a device specified by the user exists within any adapter in the system.
func (v *Values) validateDeviceExists(session bluetooth.Session) error {
	if v.AutoConnectDeviceAddr.IsNil() {
		return nil
	}

	deviceAddr := v.AutoConnectDeviceAddr

	adapters, err := session.Adapters()
	if err != nil {
		return fmt.Errorf("no adapters were found: %w", err)
	}

	adapterlist := make([]string, 0, len(adapters))
	for _, adapter := range adapters {
		if v.Adapter != "" && adapter.UniqueName != v.Adapter {
			continue
		}

		adapterlist = append(adapterlist, adapter.UniqueName)
		devices, err := session.Adapter(adapter.Address).Devices()
		if err != nil {
			continue
		}

		for _, device := range devices {
			if device.Address == deviceAddr {
				v.SelectedAdapter = &adapter
				return nil
			}
		}

		if v.Adapter != "" && adapter.UniqueName == v.Adapter {
			break
		}
	}

	return fmt.Errorf("no device with address %s found on adapters %s", deviceAddr.String(), strings.Join(adapterlist, ", "))
}

// validateKeybindings validates the keybindings.
func (v *Values) validateKeybindings() error {
	v.Kb = keybindings.NewKeybindings()
	if len(v.Keybindings) == 0 {
		return nil
	}

	return v.Kb.Validate(v.Keybindings)
}

// validateAdapterStates validates the adapter states to be set on application launch.
// The result is appended to the 'AdapterStatesMap' property, which includes:
//   - A 'sequence' key with a comma-separated value of states to be toggled in order.
//   - Each adapter state as subsequent keys, with values of "yes"/"no" to determine how each adapter state should be toggled.
func (v *Values) validateAdapterStates() error {
	if v.AdapterStates == "" {
		return nil
	}

	properties := make(map[string]string)
	propertyAndStates := strings.Split(v.AdapterStates, ",")

	propertyOptions := []string{
		"powered",
		"scan",
		"discoverable",
		"pairable",
	}

	stateOptions := []string{
		"yes", "no",
		"y", "n",
		"on", "off",
	}

	sequence := []string{}

	for _, ps := range propertyAndStates {
		property := strings.FieldsFunc(ps, func(r rune) bool {
			return r == ' ' || r == ':'
		})
		if len(property) != 2 {
			return fmt.Errorf(
				"provided property:state format '%s' is incorrect",
				ps,
			)
		}

		for _, prop := range propertyOptions {
			if property[0] == prop {
				goto CheckState
			}
		}
		return fmt.Errorf(
			"provided property '%s' is incorrect.\nValid properties are '%s'",
			property[0],
			strings.Join(propertyOptions, ", "),
		)

	CheckState:
		state := property[1]
		switch state {
		case "yes", "y", "on":
			state = "yes"

		case "no", "n", "off":
			state = "no"

		default:
			return fmt.Errorf(
				"provided state '%s' for property '%s' is incorrect.\nValid states are '%s'",
				state, property[0],
				strings.Join(stateOptions, ", "),
			)
		}

		properties[property[0]] = state
		sequence = append(sequence, property[0])
	}

	properties["sequence"] = strings.Join(sequence, ",")
	v.AdapterStatesMap = properties

	return nil
}

// validateConnectBDAddr validates the device address that has to be automatically connected to on application
// launch.
func (v *Values) validateConnectBDAddr() error {
	if v.ConnectAddr == "" {
		return nil
	}

	deviceAddr, err := bluetooth.ParseMAC(v.ConnectAddr)
	if err != nil {
		return fmt.Errorf("invalid address format: %s", v.ConnectAddr)
	}

	v.AutoConnectDeviceAddr = deviceAddr

	return nil
}

// validateReceiveDir validates the path to the download directory for received files
// via OBEX Object Push.
func (v *Values) validateReceiveDir() error {
	if v.ReceiveDir == "" {
		return nil
	}

	if statpath, err := os.Stat(v.ReceiveDir); err != nil || !statpath.IsDir() {
		return fmt.Errorf("%s: Directory is not accessible", v.ReceiveDir)
	}

	return nil
}

// validateGsm validates the GSM number and APN for the DUN network type.
func (v *Values) validateGsm() error {
	if v.GsmNumber == "" && v.GsmApn == "" {
		return nil
	}
	if v.GsmNumber == "" && v.GsmApn != "" {
		return fmt.Errorf("specify GSM Number")
	}

	if v.GsmNumber == "" {
		v.GsmNumber = "*99#"
	}

	return nil
}

// validateTheme validates the theme configuration.
func (v *Values) validateTheme() error {
	if len(v.Theme) == 0 {
		return nil
	}

	return theme.ParseThemeConfig(v.Theme)
}
