package driver

import "testing"

var noop = func() error { return nil }

func TestUpdate1(t *testing.T) {
	s := StateClosed
	s.Update(StateOpened, noop)

	if s != StateOpened {
		t.Fatalf("expected %s, got %s", StateOpened, s)
	}

	s.Update(StateClosed, noop)

	if s != StateClosed {
		t.Fatalf("expected %s, got %s", StateClosed, s)
	}

	s.Update(StateOpened, noop)

	if s != StateOpened {
		t.Fatalf("expected %s, got %s", StateOpened, s)
	}
}
