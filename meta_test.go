package mediadevices

import (
	"image"
	"testing"

	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/wave"
)

func TestDetectCurrentVideoProp(t *testing.T) {
	resolution := image.Rect(0, 0, 4, 4)
	first := image.NewRGBA(resolution)
	first.Pix[0] = 1
	second := image.NewRGBA(resolution)
	second.Pix[0] = 2

	isFirst := true
	source := video.ReaderFunc(func() (image.Image, func(), error) {
		if isFirst {
			isFirst = true
			return first, func() {}, nil
		} else {
			return second, func() {}, nil
		}
	})

	broadcaster := video.NewBroadcaster(source, nil)

	currentProp, err := detectCurrentVideoProp(broadcaster)
	if err != nil {
		t.Fatal(err)
	}

	if currentProp.Width != resolution.Dx() {
		t.Fatalf("Expect the actual width to be %d, but got %d", currentProp.Width, resolution.Dx())
	}

	if currentProp.Height != resolution.Dy() {
		t.Fatalf("Expect the actual height to be %d, but got %d", currentProp.Height, resolution.Dy())
	}

	reader := broadcaster.NewReader(false)
	img, _, err := reader.Read()
	if err != nil {
		t.Fatal(err)
	}

	rgba := img.(*image.RGBA)
	if rgba.Pix[0] != 1 {
		t.Fatal("Expect the frame after reading the current prop is not the first frame")
	}
}

func TestDetectCurrentAudioProp(t *testing.T) {
	info := wave.ChunkInfo{
		Len:          4,
		Channels:     2,
		SamplingRate: 48000,
	}
	first := wave.NewInt16Interleaved(info)
	first.Data[0] = 1
	second := wave.NewInt16Interleaved(info)
	second.Data[0] = 2

	isFirst := true
	source := audio.ReaderFunc(func() (wave.Audio, func(), error) {
		if isFirst {
			isFirst = true
			return first, func() {}, nil
		} else {
			return second, func() {}, nil
		}
	})

	broadcaster := audio.NewBroadcaster(source, nil)

	currentProp, err := detectCurrentAudioProp(broadcaster)
	if err != nil {
		t.Fatal(err)
	}

	if currentProp.ChannelCount != info.Channels {
		t.Fatalf("Expect the actual channel count to be %d, but got %d", currentProp.ChannelCount, info.Channels)
	}

	reader := broadcaster.NewReader(false)
	chunk, _, err := reader.Read()
	if err != nil {
		t.Fatal(err)
	}

	realChunk := chunk.(*wave.Int16Interleaved)
	if realChunk.Data[0] != 1 {
		t.Fatal("Expect the chunk after reading the current prop is not the first chunk")
	}
}
