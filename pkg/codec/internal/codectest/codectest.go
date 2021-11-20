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

func assertNoPanic(t *testing.T, fn func() error, msg string) error {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("panic: %v: %s", r, msg)
		}
	}()
	return fn()
}

func AudioEncoderCloseTwiceTest(t *testing.T, c codec.AudioEncoderBuilder, p prop.Media) {
	enc, err := c.BuildAudioEncoder(audio.ReaderFunc(func() (wave.Audio, func(), error) {
		return nil, nil, io.EOF
	}), p)
	if err != nil {
		t.Fatal(err)
	}

	if err := assertNoPanic(t, enc.Close, "on first Close()"); err != nil {
		t.Fatal(err)
	}
	if err := assertNoPanic(t, enc.Close, "on second Close()"); err != nil {
		t.Fatal(err)
	}
}

func VideoEncoderCloseTwiceTest(t *testing.T, c codec.VideoEncoderBuilder, p prop.Media) {
	enc, err := c.BuildVideoEncoder(video.ReaderFunc(func() (image.Image, func(), error) {
		return nil, nil, io.EOF
	}), p)
	if err != nil {
		t.Fatal(err)
	}

	if err := assertNoPanic(t, enc.Close, "on first Close()"); err != nil {
		t.Fatal(err)
	}
	if err := assertNoPanic(t, enc.Close, "on second Close()"); err != nil {
		t.Fatal(err)
	}
}
