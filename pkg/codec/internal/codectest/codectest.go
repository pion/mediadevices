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

func AudioEncoderReadAfterCloseTest(t *testing.T, c codec.AudioEncoderBuilder, p prop.Media, w wave.Audio) {
	enc, err := c.BuildAudioEncoder(audio.ReaderFunc(func() (wave.Audio, func(), error) {
		return w, nil, nil
	}), p)
	if err != nil {
		t.Fatal(err)
	}

	if err := assertNoPanic(t, enc.Close, "on Close()"); err != nil {
		t.Fatal(err)
	}
	if err := assertNoPanic(t, func() error {
		_, _, err := enc.Read()
		return err
	}, "on Read()"); err != io.EOF {
		t.Fatalf("Expected: %v, got: %v", io.EOF, err)
	}
}

func VideoEncoderReadAfterCloseTest(t *testing.T, c codec.VideoEncoderBuilder, p prop.Media, img image.Image) {
	enc, err := c.BuildVideoEncoder(video.ReaderFunc(func() (image.Image, func(), error) {
		return img, nil, nil
	}), p)
	if err != nil {
		t.Fatal(err)
	}

	if err := assertNoPanic(t, enc.Close, "on Close()"); err != nil {
		t.Fatal(err)
	}
	if err := assertNoPanic(t, func() error {
		_, _, err := enc.Read()
		return err
	}, "on Read()"); err != io.EOF {
		t.Fatalf("Expected: %v, got: %v", io.EOF, err)
	}
}
