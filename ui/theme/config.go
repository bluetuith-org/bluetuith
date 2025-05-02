package theme

import (
	"fmt"
)

// Context describes the type of context to apply the color into.
type Context string

// The different context types for themes.
const (
	ThemeText                     Context = "Text"
	ThemeBorder                   Context = "Border"
	ThemeBackground               Context = "Background"
	ThemeStatusInfo               Context = "StatusInfo"
	ThemeStatusError              Context = "StatusError"
	ThemeAdapter                  Context = "Adapter"
	ThemeAdapterPowered           Context = "AdapterPowered"
	ThemeAdapterNotPowered        Context = "AdapterNotPowered"
	ThemeAdapterDiscoverable      Context = "AdapterDiscoverable"
	ThemeAdapterScanning          Context = "AdapterScanning"
	ThemeAdapterPairable          Context = "AdapterPairable"
	ThemeDevice                   Context = "Device"
	ThemeDeviceType               Context = "DeviceType"
	ThemeDeviceAlias              Context = "DeviceAlias"
	ThemeDeviceConnected          Context = "DeviceConnected"
	ThemeDeviceDiscovered         Context = "DeviceDiscovered"
	ThemeDeviceProperty           Context = "DeviceProperty"
	ThemeDevicePropertyConnected  Context = "DevicePropertyConnected"
	ThemeDevicePropertyDiscovered Context = "DevicePropertyDiscovered"
	ThemeMenu                     Context = "Menu"
	ThemeMenuBar                  Context = "MenuBar"
	ThemeMenuItem                 Context = "MenuItem"
	ThemeProgressBar              Context = "ProgressBar"
	ThemeProgressText             Context = "ProgressText"
)

// ThemeConfig stores a list of color for the modifier elements.
var ThemeConfig = map[Context]string{
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

		ThemeConfig[Context(context)] = color
	}

	return nil
}
