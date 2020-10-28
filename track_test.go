package mediadevices

import (
	"errors"
	"testing"
	"time"
)

func TestOnEnded(t *testing.T) {
	errExpected := errors.New("an error")

	t.Run("ErrorAfterRegister", func(t *testing.T) {
		tr := &baseTrack{}

		called := make(chan error, 1)
		tr.OnEnded(func(error) {
			called <- errExpected
		})
		select {
		case <-called:
			t.Error("OnEnded handler is unexpectedly called")
		case <-time.After(10 * time.Millisecond):
		}

		tr.onError(errExpected)

		select {
		case err := <-called:
			if err != errExpected {
				t.Errorf("Expected to receive error: %v, got: %v", errExpected, err)
			}
		case <-time.After(10 * time.Millisecond):
			t.Error("Timeout")
		}
	})

	t.Run("ErrorBeforeRegister", func(t *testing.T) {
		tr := &baseTrack{}

		tr.onError(errExpected)

		called := make(chan error, 1)
		tr.OnEnded(func(err error) {
			called <- errExpected
		})
		select {
		case err := <-called:
			if err != errExpected {
				t.Errorf("Expected to receive error: %v, got: %v", errExpected, err)
			}
		case <-time.After(10 * time.Millisecond):
			t.Error("Timeout")
		}
	})
}
