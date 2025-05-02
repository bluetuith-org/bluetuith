package theme

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// ColorWrap wraps the text content with the modifier element's color.
func ColorWrap(elementName Context, elementContent string, attributes ...string) string {
	attr := "::b"
	if attributes != nil {
		attr = attributes[0]
	}

	return fmt.Sprintf("[%s%s]%s[-:-:-]", ThemeConfig[elementName], attr, elementContent)
}

// ColorName returns the name of the provided color.
func ColorName(color tcell.Color) string {
	for n, h := range tcell.ColorNames {
		if color == h {
			return n
		}
	}

	return ""
}

// BackgroundColor checks whether the given color is a light
// or dark color, and returns the appropriate color that is
// visible on top of the given color.
func BackgroundColor(themeContext Context) tcell.Color {
	if isLightColor(GetColor(themeContext)) {
		return tcell.ColorBlack
	}

	return tcell.ColorWhite
}

// GetColor returns the color of the modifier element.
func GetColor(themeContext Context) tcell.Color {
	color := ThemeConfig[themeContext]
	if color == "black" {
		return tcell.Color16
	}

	return tcell.GetColor(color)
}

// isLightColor checks if the given color is a light color.
// Adapted from:
// https://github.com/bgrins/TinyColor/blob/master/tinycolor.js#L68
func isLightColor(color tcell.Color) bool {
	r, g, b := color.RGB()
	brightness := (r*299 + g*587 + b*114) / 1000

	return brightness > 130
}

// isValidElementColor returns whether the modifier-value pair is valid.
func isValidElementColor(color string) bool {
	if color == "transparent" ||
		tcell.GetColor(color) != tcell.ColorDefault {
		return true
	}

	return false
}
