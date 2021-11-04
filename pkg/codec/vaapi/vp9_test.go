//go:build dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build dragonfly freebsd linux netbsd openbsd solaris

package vaapi

func TestShouldImplementBitRateControl(t *testing.T) {
	t.SkipNow() // TODO: Implement bit rate control

	e := &encoderVP9{}
	if _, ok := e.Controller().(codec.BitRateController); !ok {
		t.Error()
	}
}

func TestShouldImplementKeyFrameControl(t *testing.T) {
	t.SkipNow() // TODO: Implement key frame control

	e := &encoderVP9{}
	if _, ok := e.Controller().(codec.KeyFrameController); !ok {
		t.Error()
	}
}
