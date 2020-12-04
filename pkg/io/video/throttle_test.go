package video

import (
	"image"
	"runtime"
	"testing"
	"time"
)

func TestThrottle(t *testing.T) {
	// https://github.com/pion/mediadevices/issues/198
	if runtime.GOOS == "darwin" {
		t.Skip("Skipping because Darwin CI is not reliable for timing related tests.")
	}
	img := image.NewRGBA(image.Rect(0, 0, 640, 480))

	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()

	var cntPush int
	trans := Throttle(50)
	r := trans(ReaderFunc(func() (image.Image, func(), error) {
		<-ticker.C
		cntPush++
		return img, func() {}, nil
	}))

	for i := 0; i < 20; i++ {
		_, _, err := r.Read()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	}
	cntExpected := 20
	if cntPush < cntExpected*8/10 || cntExpected*12/10 < cntPush {
		t.Fatalf("Number of pushed images is expected to be %d, but pushed %d", cntExpected, cntPush)
	}
	t.Log(cntPush)
}
