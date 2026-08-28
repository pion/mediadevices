//go:build vpxtest

package vpx

import (
	"image"
	"testing"

	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

func yuv420Frame(w, h int) *image.YCbCr {
	img := image.NewYCbCr(image.Rect(0, 0, w, h), image.YCbCrSubsampleRatio420)
	for i := range img.Y {
		img.Y[i] = 128
	}
	for i := range img.Cb {
		img.Cb[i] = 128
	}
	for i := range img.Cr {
		img.Cr[i] = 128
	}
	return img
}

// TestResolutionChangeDoesNotLeakOldCodecContext reproduces the leak where a
// resolution change reinitializes the encoder context (newCtx + enc_init)
// and then frees the old context with plain free() instead of
// vpx_codec_destroy(). libvpx allocates internal state (e.g. lookahead and
// reference buffers) inside enc_init that only vpx_codec_destroy releases,
// so every resolution change leaks that memory.
//
// The test measures the process default malloc heap (libvpx 1.15+ allocates
// directly from the system allocator) while alternating between two
// resolutions. After the fix, live bytes return to the baseline after each
// switch; before the fix, they grow with every reinitialization.
func TestResolutionChangeDoesNotLeakOldCodecContext(t *testing.T) {
	p, err := NewVP8Params()
	if err != nil {
		t.Fatal(err)
	}

	// Alternates 16x16 and 32x32 frames. Every change in dimensions
	// triggers the reinitialization path in encoder.Read.
	var frameNum int
	r := video.ReaderFunc(func() (image.Image, func(), error) {
		frameNum++
		if frameNum%2 == 1 {
			return yuv420Frame(16, 16), func() {}, nil
		}
		return yuv420Frame(32, 32), func() {}, nil
	})

	e, err := p.BuildVideoEncoder(r, prop.Media{
		Video: prop.Video{Width: 16, Height: 16},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()

	// Warm up with the initial 16x16 frame so the baseline below only
	// measures memory churn caused by resolution switches.
	if _, _, err := e.Read(); err != nil {
		t.Fatal(err)
	}

	baseline := heapInUse()
	const switches = 40
	for i := 0; i < switches; i++ {
		// 32x32 triggers reinit.
		if _, _, err := e.Read(); err != nil {
			t.Fatal(err)
		}
		// Back to 16x16 triggers reinit again.
		if _, _, err := e.Read(); err != nil {
			t.Fatal(err)
		}
	}

	leaked := heapInUse() - baseline
	perSwitch := leaked / switches
	t.Logf("live heap bytes after %d resolution switches: %d (%.1f bytes per switch)",
		switches, leaked, float64(perSwitch))

	// The threshold distinguishes the regression this test guards against
	// from allocator noise. Before the fix each reinit leaks the whole
	// internal state of one encoder context (~790KB per switch, measured
	// with libvpx 1.17). After the fix the measurement is 0 on macOS and
	// ~5KB per switch on Linux CI (libvpx 1.14's destroy leaves a few KB
	// behind, visible through mallinfo2). 64KB per switch keeps a 13x
	// margin below the pre-fix level while staying well above the noise.
	const maxPerSwitchBytes = 64 * 1024
	if perSwitch > maxPerSwitchBytes {
		t.Fatalf("old codec context memory leaked across resolution changes: %d bytes (%d per switch)",
			leaked, perSwitch)
	}
}
