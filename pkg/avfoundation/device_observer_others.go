//go:build !darwin

package avfoundation

import (
	"fmt"
)

// StartObserver starts the device observer (stub for non-Darwin)
func StartObserver() error {
	return fmt.Errorf("not supported on this platform")
}

// StopObserver stops the device observer (stub for non-Darwin)
func StopObserver() error {
	return nil
}

// GetDevices returns the currently connected video devices (stub for non-Darwin)
func GetDevices() ([]Device, error) {
	return nil, fmt.Errorf("not supported on this platform")
}

