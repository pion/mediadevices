package driver

import (
	"testing"
)

func filterTrue(d Driver) bool {
	return true
}
func filterFalse(d Driver) bool {
	return false
}

func TestFilterNot(t *testing.T) {
	if FilterNot(filterTrue)(nil) != false {
		t.Error("FilterNot(filterTrue)() must be false")
	}
	if FilterNot(filterFalse)(nil) != true {
		t.Error("FilterNot(filterFalse)() must be true")
	}
}

func TestFilterAnd(t *testing.T) {
	if FilterAnd(filterTrue, filterTrue)(nil) != true {
		t.Error("FilterAnd(filterTrue, filterTrue)() must be true")
	}
	if FilterAnd(filterTrue, filterFalse)(nil) != false {
		t.Error("FilterAnd(filterTrue, filterFalse)() must be false")
	}
	if FilterAnd(filterFalse, filterTrue)(nil) != false {
		t.Error("FilterAnd(filterFalse, filterTrue)() must be false")
	}
	if FilterAnd(filterFalse, filterFalse)(nil) != false {
		t.Error("FilterAnd(filterFalse, filterFalse)() must be false")
	}
	if FilterAnd(filterFalse, filterTrue, filterTrue)(nil) != false {
		t.Error("FilterAnd(filterFalse, filterTrue, filterTrue)() must be false")
	}
	if FilterAnd(filterTrue, filterTrue, filterTrue)(nil) != true {
		t.Error("FilterAnd(filterTrue, filterTrue, filterTrue)() must be true")
	}
}
