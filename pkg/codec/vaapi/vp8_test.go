//go:build dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build dragonfly freebsd linux netbsd openbsd solaris

package vaapi

import (
	"github.com/pion/mediadevices/pkg/codec"
	"testing"
)

func TestVP8ShouldImplementBitRateControl(t *testing.T) {
	t.SkipNow() // TODO: Implement bit rate control

	e := &encoderVP8{}
	if _, ok := e.Controller().(codec.BitRateController); !ok {
		t.Error()
	}
}

func TestVP8ShouldImplementKeyFrameControl(t *testing.T) {
	e := &encoderVP8{}
	if _, ok := e.Controller().(codec.KeyFrameController); !ok {
		t.Error()
	}
}
