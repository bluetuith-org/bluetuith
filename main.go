package main

import (
	"os"

	"github.com/darkhz/bluetuith/cmd"
)

func main() {
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}
