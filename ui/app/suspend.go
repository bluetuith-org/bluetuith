//go:build !windows

package app

import (
	"syscall"

	"github.com/gdamore/tcell/v2"
)

func suspendApp(t tcell.Screen) {
	if err := t.Suspend(); err != nil {
		return
	}
	if err := syscall.Kill(syscall.Getpid(), syscall.SIGSTOP); err != nil {
		return
	}
	if err := t.Resume(); err != nil {
		return
	}
}
