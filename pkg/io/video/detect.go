package video

import (
	"image"
	"time"

	"github.com/pion/mediadevices/pkg/prop"
)

// DetectChanges will detect frame and video property changes. For video property detection,
// since it's time related, interval will be used to determine the sample rate.
func DetectChanges(interval time.Duration, onChange func(prop.Media)) TransformFunc {
	return func(r Reader) Reader {
		var currentProp prop.Media
		var lastTaken time.Time
		var frames uint
		return ReaderFunc(func() (image.Image, func(), error) {
			var dirty bool

			img, _, err := r.Read()
			if err != nil {
				return nil, func() {}, err
			}

			bounds := img.Bounds()
			if currentProp.Width != bounds.Dx() {
				currentProp.Width = bounds.Dx()
				dirty = true
			}

			if currentProp.Height != bounds.Dy() {
				currentProp.Height = bounds.Dy()
				dirty = true
			}

			// TODO: maybe detect frame format? It probably doesn't make sense since some
			// formats only are about memory layout, e.g. YUV2 vs NV12.

			now := time.Now()
			elapsed := now.Sub(lastTaken)
			if elapsed >= interval {
				fps := float32(float64(frames) / elapsed.Seconds())
				// TODO: maybe add some epsilon so that small changes will not mark as dirty
				currentProp.FrameRate = fps
				frames = 0
				lastTaken = now
				dirty = true
			}

			if dirty {
				onChange(currentProp)
			}

			frames++
			return img, func() {}, nil
		})
	}
}
