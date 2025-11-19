//go:build !darwin

package avfoundation

import (
	"fmt"
)

func StartObserver() error {
	return fmt.Errorf("not supported on this platform")
}

func StopObserver() error {
	return nil
}
