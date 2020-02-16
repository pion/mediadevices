package screen

import (
	"testing"
)

func TestAlign64(t *testing.T) {
	if ret := cAlign64(0x00010008); ret != 0x00010008 {
		t.Errorf("Wrong alignment, expected %x, got %x", 0x00010008, ret)
	}
	if ret := cAlign64(0x00010006); ret != 0x00010008 {
		t.Errorf("Wrong alignment, expected %x, got %x", 0x00010008, ret)
	}
	if ret := cAlign64(0x00010009); ret != 0x00010010 {
		t.Errorf("Wrong alignment, expected %x, got %x", 0x00010010, ret)
	}
}
