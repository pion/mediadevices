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
			broadcaster := NewBroadcaster(src, nil)
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

	routinePauseConds := []struct {
		src          bool
		dst          bool
		expectedFPS  float64
		expectedDrop float64
	}{
		{
			src:         false,
			dst:         false,
			expectedFPS: 30,
		},
		{
			src:          true,
			dst:          false,
			expectedFPS:  20,
			expectedDrop: 10,
		},
		{
			src:          false,
			dst:          true,
			expectedFPS:  20,
			expectedDrop: 10,
		},
	}

	for _, pauseCond := range routinePauseConds {
		pauseCond := pauseCond
		t.Run(fmt.Sprintf("SrcPause-%v/DstPause-%v", pauseCond.src, pauseCond.dst), func(t *testing.T) {
			for n := 1; n <= 256; n *= 16 {
				n := n

				t.Run(fmt.Sprintf("Readers-%d", n), func(t *testing.T) {
					var src Reader
					interval := time.NewTicker(time.Millisecond * 33) // 30 fps
					defer interval.Stop()
					frameCount := 0
					frameSent := 0
					lastSend := time.Now()
					src = ReaderFunc(func() (image.Image, error) {
						if pauseCond.src && frameSent == 30 {
							time.Sleep(time.Second)
						}
						<-interval.C

						now := time.Now()
						if interval := now.Sub(lastSend); interval > time.Millisecond*33*3/2 {
							// Source reader should drop frames to catch up the latest frame.
							drop := int(interval/(time.Millisecond*33)) - 1
							frameCount += drop
							t.Logf("Skipped %d frames", drop)
						}
						lastSend = now
						frame := frames[frameCount]
						frameCount++
						frameSent++
						return frame, nil
					})
					broadcaster := NewBroadcaster(src, nil)
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
								if pauseCond.dst && count == 30 {
									time.Sleep(time.Second)
								}
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
							if fps < pauseCond.expectedFPS-2 || fps > pauseCond.expectedFPS+2 {
								t.Fatal("Unexpected average FPS")
							}

							droppedFramesPerSecond := float64(droppedFrames) / duration.Seconds()
							if droppedFramesPerSecond < pauseCond.expectedDrop-2 || droppedFramesPerSecond > pauseCond.expectedDrop+2 {
								t.Fatal("Unexpected drop count")
							}

							fpsChan <- []float64{fps, droppedFramesPerSecond, float64(lastFrameCount)}
						}()
					}

					time.Sleep(duration)
					atomic.StoreUint32(&done, 1)

					var fpsAvg float64
					var droppedFramesPerSecondAvg float64
					var lastFrameCountAvg float64
					var count int
					for metric := range fpsChan {
						fps, droppedFramesPerSecond, lastFrameCount := metric[0], metric[1], metric[2]
						fpsAvg += fps
						droppedFramesPerSecondAvg += droppedFramesPerSecond
						lastFrameCountAvg += lastFrameCount
						count++
						if count == n {
							break
						}
					}

					t.Log("Average FPS                      :", fpsAvg/float64(n))
					t.Log("Average dropped frames per second:", droppedFramesPerSecondAvg/float64(n))
					t.Log("Last frame count (src)           :", frameCount)
					t.Log("Average last frame count (dst)   :", lastFrameCountAvg/float64(n))
				})
			}
		})
	}
}
