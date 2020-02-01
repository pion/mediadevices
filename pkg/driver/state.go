package driver

import "fmt"

// State represents driver's state
type State uint

const (
	// StateClosed means that the driver has not been opened. In this state,
	// all information related to the hardware are still unknown. For example,
	// if it's a video driver, the pixel format information is still unknown.
	StateClosed State = iota
	// StateOpened means that the driver is already opened and information about
	// the hardware are already known and may be extracted from the driver.
	StateOpened
	// StateStarted means that the driver has been sending data. The caller
	// who started the driver may start reading data from the hardware.
	StateStarted
	// StateStopped means that the driver is no longer sending data. In this state,
	// information about the hardware is still available.
	StateStopped
)

// Update updates current state, s, to next. If f fails to execute,
// s will stay unchanged. Otherwise, s will be updated to next
func (s *State) Update(next State, f func() error) error {
	type checkFunc func() error
	m := map[State]checkFunc{
		StateOpened:  s.toOpened,
		StateClosed:  s.toClosed,
		StateStarted: s.toStarted,
		StateStopped: s.toStopped,
	}

	err := m[next]()
	if err != nil {
		return err
	}

	err = f()
	if err == nil {
		*s = next
	}
	return err
}

func (s *State) toOpened() error {
	if *s != StateClosed {
		return fmt.Errorf("invalid state: driver is already opened")
	}
	return nil
}

func (s *State) toClosed() error {
	return nil
}

func (s *State) toStarted() error {
	if *s == StateClosed {
		return fmt.Errorf("invalid state: driver hasn't been opened")
	}

	if *s == StateStarted {
		return fmt.Errorf("invalid state: driver has been started")
	}

	return nil
}

func (s *State) toStopped() error {
	if *s != StateStarted {
		return fmt.Errorf("invalid state: driver hasn't been started")
	}

	return nil
}
