package cmd

import (
	"github.com/fatih/color"
)

// printWarn prints a warning to the screen.
func printWarn(message string) {
	message = "[-] " + message

	color.New(color.FgYellow, color.Bold).Println(message)
}

// printError prints an error to the screen.
func printError(err error) {
	message := "[!] " + err.Error()

	color.New(color.FgRed, color.Bold).Println(message)
}
