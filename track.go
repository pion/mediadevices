package mediadevices

import (
	"fmt"
	"image"
	"sync"

	"github.com/pion/mediadevices/pkg/driver"
	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/mediadevices/pkg/wave"
)

type TrackKind string

const (
	TrackKindVideo TrackKind = "video"
	TrackKindAudio TrackKind = "audio"
)

// Track is an interface that represent MediaStreamTrack
// Reference: https://w3c.github.io/mediacapture-main/#mediastreamtrack
type Track interface {
	ID() string
	Kind() TrackKind
	Stop()
	// OnEnded registers a handler to receive an error from the media stream track.
	// If the error is already occured before registering, the handler will be
	// immediately called.
	OnEnded(func(error))
}

type VideoTrack struct {
	baseTrack
	src         video.Reader
	transformed video.Reader
	mux         sync.Mutex
	frameCount  int
	lastFrame   image.Image
	lastErr     error
}

func newVideoTrack(d driver.Driver, constraints prop.Media) (*VideoTrack, error) {
	err := d.Open()
	if err != nil {
		return nil, err
	}

	recorder, ok := d.(driver.VideoRecorder)
	if !ok {
		d.Close()
		return nil, fmt.Errorf("driver is not an video recorder")
	}

	r, err := recorder.VideoRecord(constraints)
	if err != nil {
		d.Close()
		return nil, err
	}

	return &VideoTrack{
		baseTrack:   baseTrack{d: d},
		src:         r,
		transformed: r,
	}, nil
}

func (track *VideoTrack) Kind() TrackKind {
	return TrackKindVideo
}

func (track *VideoTrack) NewReader() video.Reader {
	var curFrameCount int
	return video.ReaderFunc(func() (img image.Image, err error) {
		track.mux.Lock()
		defer track.mux.Unlock()

		if curFrameCount != track.frameCount {
			img = copyFrame(img, track.lastFrame)
			err = track.lastErr
		} else {
			img, err = track.transformed.Read()
			track.lastFrame = img
			track.lastErr = err
			track.frameCount++
		}

		curFrameCount = track.frameCount
		return
	})
}

// TODO: implement copy in place
func copyFrame(dst, src image.Image) image.Image { return src }

func (track *VideoTrack) Transform(fns ...video.TransformFunc) {
	track.mux.Lock()
	defer track.mux.Unlock()
	track.transformed = video.Merge(fns...)(track.src)
}

type AudioTrack struct {
	baseTrack
	src         audio.Reader
	transformed audio.Reader
	mux         sync.Mutex
	chunkCount  int
	lastChunks  wave.Audio
	lastErr     error
}

func newAudioTrack(d driver.Driver, constraints prop.Media) (*AudioTrack, error) {
	err := d.Open()
	if err != nil {
		return nil, err
	}

	recorder, ok := d.(driver.AudioRecorder)
	if !ok {
		d.Close()
		return nil, fmt.Errorf("driver is not an audio recorder")
	}

	r, err := recorder.AudioRecord(constraints)
	if err != nil {
		d.Close()
		return nil, err
	}

	return &AudioTrack{
		baseTrack:   baseTrack{d: d},
		src:         r,
		transformed: r,
	}, nil
}

func (track *AudioTrack) Kind() TrackKind {
	return TrackKindAudio
}

func (track *AudioTrack) NewReader() audio.Reader {
	var currChunkCount int
	return audio.ReaderFunc(func() (chunks wave.Audio, err error) {
		track.mux.Lock()
		defer track.mux.Unlock()

		if currChunkCount != track.chunkCount {
			chunks = copyChunks(chunks, track.lastChunks)
			err = track.lastErr
		} else {
			chunks, err = track.transformed.Read()
			track.lastChunks = chunks
			track.lastErr = err
			track.chunkCount++
		}

		currChunkCount = track.chunkCount
		return
	})
}

// TODO: implement copy in place
func copyChunks(dst, src wave.Audio) wave.Audio { return src }

func (track *AudioTrack) Transform(fns ...audio.TransformFunc) {
	track.mux.Lock()
	defer track.mux.Unlock()
	track.transformed = audio.Merge(fns...)(track.src)
}

type baseTrack struct {
	d driver.Driver

	onErrorHandler func(error)
	err            error
	mu             sync.Mutex
	endOnce        sync.Once
}

func (t *baseTrack) ID() string {
	return t.d.ID()
}

// OnEnded sets an error handler. When a track has been created and started, if an
// error occurs, handler will get called with the error given to the parameter.
func (t *baseTrack) OnEnded(handler func(error)) {
	t.mu.Lock()
	t.onErrorHandler = handler
	err := t.err
	t.mu.Unlock()

	if err != nil && handler != nil {
		// Already errored.
		t.endOnce.Do(func() {
			handler(err)
		})
	}
}

// onError is a callback when an error occurs
func (t *baseTrack) onError(err error) {
	t.mu.Lock()
	t.err = err
	handler := t.onErrorHandler
	t.mu.Unlock()

	if handler != nil {
		t.endOnce.Do(func() {
			handler(err)
		})
	}
}

func (t *baseTrack) Stop() {
	t.d.Close()
}
