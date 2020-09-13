package video

import (
	"fmt"
	"image"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func BenchmarkBroadcast(b *testing.B) {
	var src Reader
	img := image.NewRGBA(image.Rect(0, 0, 1920, 1080))
	interval := time.NewTicker(time.Millisecond * 33) // 30 fps
	defer interval.Stop()
	src = ReaderFunc(func() (image.Image, error) {
		<-interval.C
		return img, nil
	})

	for n := 1; n <= 4096; n *= 16 {
		n := n

		b.Run(fmt.Sprintf("Readers-%d", n), func(b *testing.B) {
			b.SetParallelism(n)
			broadcaster := NewBroadcaster(src)
			b.RunParallel(func(pb *testing.PB) {
				reader := broadcaster.NewReader(false)
				for pb.Next() {
					reader.Read()
				}
			})
		})
	}
}

func TestBroadcast(t *testing.T) {
	// https://github.com/pion/mediadevices/issues/198
	if runtime.GOOS == "darwin" {
		t.Skip("Skipping because Darwin CI is not reliable for timing related tests.")
	}
	frames := make([]image.Image, 5*30) // 5 seconds worth of frames
	resolution := image.Rect(0, 0, 1920, 1080)
	for i := range frames {
		rgba := image.NewRGBA(resolution)
		rgba.Pix[0] = uint8(i >> 24)
		rgba.Pix[1] = uint8(i >> 16)
		rgba.Pix[2] = uint8(i >> 8)
		rgba.Pix[3] = uint8(i)
		frames[i] = rgba
	}

	for n := 1; n <= 256; n *= 16 {
		n := n

		t.Run(fmt.Sprintf("Readers-%d", n), func(t *testing.T) {
			var src Reader
			interval := time.NewTicker(time.Millisecond * 33) // 30 fps
			defer interval.Stop()
			frameCount := 0
			src = ReaderFunc(func() (image.Image, error) {
				<-interval.C
				frame := frames[frameCount]
				frameCount++
				return frame, nil
			})
			broadcaster := NewBroadcaster(src)
			var done uint32
			duration := time.Second * 3
			fpsChan := make(chan []float64)

			var wg sync.WaitGroup
			wg.Add(n)
			for i := 0; i < n; i++ {
				go func() {
					reader := broadcaster.NewReader(false)
					count := 0
					lastFrameCount := -1
					droppedFrames := 0
					wg.Done()
					wg.Wait()
					for atomic.LoadUint32(&done) == 0 {
						frame, err := reader.Read()
						if err != nil {
							t.Error(err)
						}
						rgba := frame.(*image.RGBA)
						var frameCount int
						frameCount |= int(rgba.Pix[0]) << 24
						frameCount |= int(rgba.Pix[1]) << 16
						frameCount |= int(rgba.Pix[2]) << 8
						frameCount |= int(rgba.Pix[3])

						droppedFrames += (frameCount - lastFrameCount - 1)
						lastFrameCount = frameCount
						count++
					}

					fps := float64(count) / duration.Seconds()
					if fps < 28 || fps > 32 {
						t.Fail()
					}

					droppedFramesPerSecond := float64(droppedFrames) / duration.Seconds()
					if droppedFramesPerSecond > 0.3 {
						t.Fail()
					}

					fpsChan <- []float64{fps, droppedFramesPerSecond}
				}()
			}

			time.Sleep(duration)
			atomic.StoreUint32(&done, 1)

			var fpsAvg float64
			var droppedFramesPerSecondAvg float64
			var count int
			for metric := range fpsChan {
				fps, droppedFramesPerSecond := metric[0], metric[1]
				fpsAvg += fps
				droppedFramesPerSecondAvg += droppedFramesPerSecond
				count++
				if count == n {
					break
				}
			}

			t.Log("Average FPS                      :", fpsAvg/float64(n))
			t.Log("Average dropped frames per second:", droppedFramesPerSecondAvg/float64(n))
		})
	}
}
