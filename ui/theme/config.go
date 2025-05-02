package theme

import (
	"fmt"
)

// ThemeContext describes the type of context to apply the color into.
type ThemeContext string

// The different context types for themes.
const (
	ThemeText                     ThemeContext = "Text"
	ThemeBorder                   ThemeContext = "Border"
	ThemeBackground               ThemeContext = "Background"
	ThemeStatusInfo               ThemeContext = "StatusInfo"
	ThemeStatusError              ThemeContext = "StatusError"
	ThemeAdapter                  ThemeContext = "Adapter"
	ThemeAdapterPowered           ThemeContext = "AdapterPowered"
	ThemeAdapterNotPowered        ThemeContext = "AdapterNotPowered"
	ThemeAdapterDiscoverable      ThemeContext = "AdapterDiscoverable"
	ThemeAdapterScanning          ThemeContext = "AdapterScanning"
	ThemeAdapterPairable          ThemeContext = "AdapterPairable"
	ThemeDevice                   ThemeContext = "Device"
	ThemeDeviceType               ThemeContext = "DeviceType"
	ThemeDeviceAlias              ThemeContext = "DeviceAlias"
	ThemeDeviceConnected          ThemeContext = "DeviceConnected"
	ThemeDeviceDiscovered         ThemeContext = "DeviceDiscovered"
	ThemeDeviceProperty           ThemeContext = "DeviceProperty"
	ThemeDevicePropertyConnected  ThemeContext = "DevicePropertyConnected"
	ThemeDevicePropertyDiscovered ThemeContext = "DevicePropertyDiscovered"
	ThemeMenu                     ThemeContext = "Menu"
	ThemeMenuBar                  ThemeContext = "MenuBar"
	ThemeMenuItem                 ThemeContext = "MenuItem"
	ThemeProgressBar              ThemeContext = "ProgressBar"
	ThemeProgressText             ThemeContext = "ProgressText"
)

// ThemeConfig stores a list of color for the modifier elements.
var ThemeConfig = map[ThemeContext]string{
	ThemeText:        "white",
	ThemeBorder:      "white",
	ThemeBackground:  "default",
	ThemeStatusInfo:  "white",
	ThemeStatusError: "red",

	ThemeAdapter:             "white",
	ThemeAdapterPowered:      "green",
	ThemeAdapterNotPowered:   "red",
	ThemeAdapterDiscoverable: "aqua",
	ThemeAdapterScanning:     "yellow",
	ThemeAdapterPairable:     "mediumorchid",

	ThemeDevice:                   "white",
	ThemeDeviceType:               "white",
	ThemeDeviceAlias:              "white",
	ThemeDeviceConnected:          "white",
	ThemeDeviceDiscovered:         "white",
	ThemeDeviceProperty:           "grey",
	ThemeDevicePropertyConnected:  "green",
	ThemeDevicePropertyDiscovered: "orange",

	ThemeMenu:     "white",
	ThemeMenuBar:  "default",
	ThemeMenuItem: "white",

	ThemeProgressBar:  "white",
	ThemeProgressText: "white",
}

// ParseThemeConfig parses the theme configuration.
func ParseThemeConfig(themeConfig map[string]string) error {
	for context, color := range themeConfig {
		if !isValidElementColor(color) {
			return fmt.Errorf("theme configuration is incorrect for %s (%s)", context, color)
		}

		switch color {
		case "black":
			color = "#000000"

		case "transparent":
			color = "default"
		}

		ThemeConfig[ThemeContext(context)] = color
	}

	return nil
}
