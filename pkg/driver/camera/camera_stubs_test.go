//go:build linux || windows

package camera

import (
	"errors"
	"testing"

	"github.com/pion/mediadevices/pkg/driver/availability"
)

// TestSetupObserver tests the stub implementation of SetupObserver.
func TestSetupObserver(t *testing.T) {
	err := SetupObserver()
	if !errors.Is(err, availability.ErrUnimplemented) {
		t.Errorf("SetupObserver() should return ErrUnimplemented for stub implementation, got: %v", err)
	}
}

// TestStartObserver tests the stub implementation of StartObserver.
func TestStartObserver(t *testing.T) {
	err := StartObserver()
	if !errors.Is(err, availability.ErrUnimplemented) {
		t.Errorf("StartObserver() should return ErrUnimplemented for stub implementation, got: %v", err)
	}
}

// TestDestroyObserver tests the stub implementation of DestroyObserver.
func TestDestroyObserver(t *testing.T) {
	err := DestroyObserver()
	if !errors.Is(err, availability.ErrUnimplemented) {
		t.Errorf("DestroyObserver() should return ErrUnimplemented for stub implementation, got: %v", err)
	}
}

// TestObserverFunctionsIdempotent tests that observer functions can be called multiple times safely.
func TestObserverFunctionsIdempotent(t *testing.T) {
	for i := 0; i < 3; i++ {
		if err := SetupObserver(); !errors.Is(err, availability.ErrUnimplemented) {
			t.Errorf("SetupObserver() call %d should return ErrUnimplemented, got: %v", i+1, err)
		}
		if err := StartObserver(); !errors.Is(err, availability.ErrUnimplemented) {
			t.Errorf("StartObserver() call %d should return ErrUnimplemented, got: %v", i+1, err)
		}
	}

	for i := 0; i < 3; i++ {
		if err := DestroyObserver(); !errors.Is(err, availability.ErrUnimplemented) {
			t.Errorf("DestroyObserver() call %d should return ErrUnimplemented, got: %v", i+1, err)
		}
	}
}
