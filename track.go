package mediadevices

import (
	"errors"
	"math/rand"
	"sync"

	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/driver"
	mio "github.com/pion/mediadevices/pkg/io"
	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media"
)

// Tracker is an interface that represent MediaStreamTrack
// Reference: https://w3c.github.io/mediacapture-main/#mediastreamtrack
type Tracker interface {
	Track() *webrtc.Track
	LocalTrack() LocalTrack
	Stop()
	// OnEnded registers a handler to receive an error from the media stream track.
	// If the error is already occured before registering, the handler will be
	// immediately called.
	OnEnded(func(error))
}

type LocalTrack interface {
	WriteSample(s media.Sample) error
	Codec() *webrtc.RTPCodec
	ID() string
	Kind() webrtc.RTPCodecType
}

type track struct {
	localTrack LocalTrack
	d          driver.Driver
	sample     samplerFunc
	encoder    codec.ReadCloser

	onErrorHandler func(error)
	err            error
	mu             sync.Mutex
	endOnce        sync.Once
}

func newTrack(opts *MediaDevicesOptions, d driver.Driver, constraints MediaTrackConstraints) (*track, error) {
	var encoderBuilders []encoderBuilder
	var rtpCodecs []*webrtc.RTPCodec
	var buildSampler func(t LocalTrack) samplerFunc
	var err error

	err = d.Open()
	if err != nil {
		return nil, err
	}

	switch r := d.(type) {
	case driver.VideoRecorder:
		rtpCodecs = opts.codecs[webrtc.RTPCodecTypeVideo]
		buildSampler = newVideoSampler
		encoderBuilders, err = newVideoEncoderBuilders(r, constraints)
	case driver.AudioRecorder:
		rtpCodecs = opts.codecs[webrtc.RTPCodecTypeAudio]
		buildSampler = func(t LocalTrack) samplerFunc {
			return newAudioSampler(t, constraints.selectedMedia.Latency)
		}
		encoderBuilders, err = newAudioEncoderBuilders(r, constraints)
	default:
		err = errors.New("newTrack: invalid driver type")
	}

	if err != nil {
		d.Close()
		return nil, err
	}

	for _, builder := range encoderBuilders {
		var matchedRTPCodec *webrtc.RTPCodec
		for _, rtpCodec := range rtpCodecs {
			if rtpCodec.Name == builder.name {
				matchedRTPCodec = rtpCodec
				break
			}
		}

		if matchedRTPCodec == nil {
			continue
		}

		localTrack, err := opts.trackGenerator(
			matchedRTPCodec.PayloadType,
			rand.Uint32(),
			d.ID(),
			matchedRTPCodec.Type.String(),
			matchedRTPCodec,
		)
		if err != nil {
			continue
		}

		encoder, err := builder.build()
		if err != nil {
			continue
		}

		t := track{
			localTrack: localTrack,
			sample:     buildSampler(localTrack),
			d:          d,
			encoder:    encoder,
		}
		go t.start()
		return &t, nil
	}

	d.Close()
	return nil, errors.New("newTrack: failed to find a matching codec")
}

// OnEnded sets an error handler. When a track has been created and started, if an
// error occurs, handler will get called with the error given to the parameter.
func (t *track) OnEnded(handler func(error)) {
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
func (t *track) onError(err error) {
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

// start starts the data flow from the driver all the way to the localTrack
func (t *track) start() {
	var n int
	var err error
	buff := make([]byte, 1024)
	for {
		n, err = t.encoder.Read(buff)
		if err != nil {
			if e, ok := err.(*mio.InsufficientBufferError); ok {
				buff = make([]byte, 2*e.RequiredSize)
				continue
			}

			t.onError(err)
			return
		}

		if err := t.sample(buff[:n]); err != nil {
			t.onError(err)
			return
		}
	}
}

// Stop stops the underlying driver and encoder
func (t *track) Stop() {
	t.d.Close()
	t.encoder.Close()
}

func (t *track) Track() *webrtc.Track {
	return t.localTrack.(*webrtc.Track)
}

func (t *track) LocalTrack() LocalTrack {
	return t.localTrack
}

// encoderBuilder is a generic encoder builder that acts as a delegator for codec.VideoEncoderBuilder and
// codec.AudioEncoderBuilder. The idea of having a delegator is to reduce redundant codes that are being
// duplicated for managing video and audio.
type encoderBuilder struct {
	name  string
	build func() (codec.ReadCloser, error)
}

// newVideoEncoderBuilders transforms video given by VideoRecorder with the video transformer that is passed through
// constraints and create a list of generic encoder builders
func newVideoEncoderBuilders(vr driver.VideoRecorder, constraints MediaTrackConstraints) ([]encoderBuilder, error) {
	r, err := vr.VideoRecord(constraints.selectedMedia)
	if err != nil {
		return nil, err
	}

	if constraints.VideoTransform != nil {
		r = constraints.VideoTransform(r)
	}

	encoderBuilders := make([]encoderBuilder, len(constraints.VideoEncoderBuilders))
	for i, b := range constraints.VideoEncoderBuilders {
		encoderBuilders[i].name = b.RTPCodec().Name
		encoderBuilders[i].build = func() (codec.ReadCloser, error) {
			return b.BuildVideoEncoder(r, constraints.selectedMedia)
		}
	}
	return encoderBuilders, nil
}

// newAudioEncoderBuilders transforms audio given by AudioRecorder with the audio transformer that is passed through
// constraints and create a list of generic encoder builders
func newAudioEncoderBuilders(ar driver.AudioRecorder, constraints MediaTrackConstraints) ([]encoderBuilder, error) {
	r, err := ar.AudioRecord(constraints.selectedMedia)
	if err != nil {
		return nil, err
	}

	if constraints.AudioTransform != nil {
		r = constraints.AudioTransform(r)
	}

	encoderBuilders := make([]encoderBuilder, len(constraints.AudioEncoderBuilders))
	for i, b := range constraints.AudioEncoderBuilders {
		encoderBuilders[i].name = b.RTPCodec().Name
		encoderBuilders[i].build = func() (codec.ReadCloser, error) {
			return b.BuildAudioEncoder(r, constraints.selectedMedia)
		}
	}
	return encoderBuilders, nil
}
