//go:build !darwin

package xpe

import "errors"

var ErrPlatformNotSupported = errors.New("xpe: unsupported on this platform")

// GetCPU returns information about the CPU, but is not currently supported on this platform.
//
// It returns [ErrPlatformNotSupported] to indicate that this functionality is not available.
func GetCPU() (*CPU, error) {
	return nil, ErrPlatformNotSupported
}
