// Package codectest provides shared test for codec implementations.
package codectest

import (
	"image"
	"io"
	"testing"

	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/mediadevices/pkg/wave"
)

func AudioEncoderSimpleReadTest(t *testing.T, c codec.AudioEncoderBuilder, p prop.Media, w wave.Audio) {
	var eof bool
	enc, err := c.BuildAudioEncoder(audio.ReaderFunc(func() (wave.Audio, func(), error) {
		if eof {
			return nil, nil, io.EOF
		}
		return w, nil, nil
	}), p)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 16; i++ {
		b, release, err := enc.Read()
		if err != nil {
			t.Fatal(err)
		}
		if len(b) == 0 {
			t.Fatal("Encoded frame is empty")
		}
		release()
	}

	eof = true
	if _, _, err := enc.Read(); err != io.EOF {
		t.Fatalf("Expected EOF, got %v", err)
	}

	if err := enc.Close(); err != nil {
		t.Fatal(err)
	}
}

func VideoEncoderSimpleReadTest(t *testing.T, c codec.VideoEncoderBuilder, p prop.Media, img image.Image) {
	var eof bool
	enc, err := c.BuildVideoEncoder(video.ReaderFunc(func() (image.Image, func(), error) {
		if eof {
			return nil, nil, io.EOF
		}
		return img, nil, nil
	}), p)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 16; i++ {
		b, release, err := enc.Read()
		if err != nil {
			t.Fatal(err)
		}
		if len(b) == 0 {
			t.Errorf("Encoded frame is empty (%d)", i)
		}
		release()
	}

	eof = true
	if _, _, err := enc.Read(); err != io.EOF {
		t.Fatalf("Expected EOF, got %v", err)
	}

	if err := enc.Close(); err != nil {
		t.Fatal(err)
	}
}
