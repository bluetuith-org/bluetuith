package platforminfo

import "runtime"

// PlatformInfo describes platform-specific information.
type PlatformInfo struct {
	OS    string `json:"os_info,omitempty"`
	Stack string `json:"stack,omitempty"`
}

// NewPlatformInfo returns a new PlatformInfo.
func NewPlatformInfo(stack string) PlatformInfo {
	return PlatformInfo{
		OS:    runtime.GOOS + " (" + runtime.GOARCH + ")",
		Stack: stack,
	}
}
