package mediadevices

import (
	"fmt"
	"image"
	"math/rand"
	"sync"

	"github.com/pion/mediadevices/pkg/driver"
	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/wave"
)

// TrackKind represents content type of a track
type TrackKind string

const (
	TrackKindVideo TrackKind = "video"
	TrackKindAudio TrackKind = "audio"
)

// Track is an interface that represent MediaStreamTrack
// Reference: https://w3c.github.io/mediacapture-main/#mediastreamtrack
type Track interface {
	ID() string
	SSRC() uint32
	Kind() TrackKind
	Stop()
	// OnEnded registers a handler to receive an error from the media stream track.
	// If the error is already occured before registering, the handler will be
	// immediately called.
	OnEnded(func(error))
}

// VideoTrack is a specialized track for video
type VideoTrack struct {
	baseTrack
	src         video.Reader
	transformed video.Reader
	mux         sync.Mutex
	frameCount  int
	lastFrame   image.Image
	lastErr     error
}

func newVideoTrack(d driver.Driver, constraints MediaTrackConstraints) (*VideoTrack, error) {
	err := d.Open()
	if err != nil {
		return nil, err
	}

	recorder, ok := d.(driver.VideoRecorder)
	if !ok {
		d.Close()
		return nil, fmt.Errorf("driver is not an video recorder")
	}

	r, err := recorder.VideoRecord(constraints.selectedMedia)
	if err != nil {
		d.Close()
		return nil, err
	}

	return &VideoTrack{
		baseTrack:   newBaseTrack(d, constraints),
		src:         r,
		transformed: r,
	}, nil
}

// Kind returns track's kind
func (track *VideoTrack) Kind() TrackKind {
	return TrackKindVideo
}

// NewReader returns a reader to read frames from the source. You may create multiple
// readers and read from them in different goroutines.
//
// In the case of multiple readers, reading from the source will only get triggered
// when the reader has the latest frame from the source
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
			if err != nil {
				track.onErrorHandler(err)
			}
		}

		curFrameCount = track.frameCount
		return
	})
}

// TODO: implement copy in place
func copyFrame(dst, src image.Image) image.Image { return src }

// Transform transforms the underlying source. The transformation will reflect to
// all readers
func (track *VideoTrack) Transform(fns ...video.TransformFunc) {
	track.mux.Lock()
	defer track.mux.Unlock()
	track.transformed = video.Merge(fns...)(track.src)
}

// AudioTrack is a specialized track for audio
type AudioTrack struct {
	baseTrack
	src         audio.Reader
	transformed audio.Reader
	mux         sync.Mutex
	chunkCount  int
	lastChunks  wave.Audio
	lastErr     error
}

func newAudioTrack(d driver.Driver, constraints MediaTrackConstraints) (*AudioTrack, error) {
	err := d.Open()
	if err != nil {
		return nil, err
	}

	recorder, ok := d.(driver.AudioRecorder)
	if !ok {
		d.Close()
		return nil, fmt.Errorf("driver is not an audio recorder")
	}

	r, err := recorder.AudioRecord(constraints.selectedMedia)
	if err != nil {
		d.Close()
		return nil, err
	}

	return &AudioTrack{
		baseTrack:   newBaseTrack(d, constraints),
		src:         r,
		transformed: r,
	}, nil
}

func (track *AudioTrack) Kind() TrackKind {
	return TrackKindAudio
}

// NewReader returns a reader to read audio chunks from the source. You may create multiple
// readers and read from them in different goroutines.
//
// In the case of multiple readers, reading from the source will only get triggered
// when the reader has the latest chunk from the source
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
			if err != nil {
				track.onErrorHandler(err)
			}
		}

		currChunkCount = track.chunkCount
		return
	})
}

// TODO: implement copy in place
func copyChunks(dst, src wave.Audio) wave.Audio { return src }

// Transform transforms the underlying source. The transformation will reflect to
// all readers
func (track *AudioTrack) Transform(fns ...audio.TransformFunc) {
	track.mux.Lock()
	defer track.mux.Unlock()
	track.transformed = audio.Merge(fns...)(track.src)
}

type baseTrack struct {
	d           driver.Driver
	constraints MediaTrackConstraints
	ssrc        uint32

	onErrorHandler func(error)
	err            error
	mu             sync.Mutex
	endOnce        sync.Once
}

func newBaseTrack(d driver.Driver, constraints MediaTrackConstraints) baseTrack {
	return baseTrack{d: d, constraints: constraints, ssrc: rand.Uint32()}
}

func (t *baseTrack) ID() string {
	return t.d.ID()
}

func (t *baseTrack) SSRC() uint32 {
	return t.ssrc
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
