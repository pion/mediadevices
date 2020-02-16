package video

import (
	"image"
	"time"
)

// Throttle returns video throttling transform.
// This transform drops some of the incoming frames to achieve given framerate in fps.
func Throttle(rate float32) TransformFunc {
	return func(r Reader) Reader {
		ticker := time.NewTicker(time.Duration(int64(float64(time.Second) / float64(rate))))
		return ReaderFunc(func() (image.Image, error) {
			for {
				img, err := r.Read()
				if err != nil {
					ticker.Stop()
					return nil, err
				}
				select {
				case <-ticker.C:
					return img, nil
				default:
				}
			}
		})
	}
}
