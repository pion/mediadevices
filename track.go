package mediadevices

import (
	"fmt"
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
	t LocalTrack
	s *sampler

	onErrorHandler func(error)
	err            error
	mu             sync.Mutex
	endOnce        sync.Once
}

func newTrack(selectedCodec *webrtc.RTPCodec, trackGenerator TrackGenerator, d driver.Driver) (*track, error) {
	if selectedCodec == nil {
		panic("codec is required")
	}

	t, err := trackGenerator(
		selectedCodec.PayloadType,
		rand.Uint32(),
		d.ID(),
		selectedCodec.Type.String(),
		selectedCodec,
	)
	if err != nil {
		return nil, err
	}

	return &track{
		t: t,
		s: newSampler(t),
	}, nil
}

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

func (t *track) Track() *webrtc.Track {
	return t.t.(*webrtc.Track)
}

func (t *track) LocalTrack() LocalTrack {
	return t.t
}

type videoTrack struct {
	*track
	d           driver.Driver
	constraints MediaTrackConstraints
	encoder     codec.ReadCloser
}

var _ Tracker = &videoTrack{}

func newVideoTrack(opts *MediaDevicesOptions, d driver.Driver, constraints MediaTrackConstraints) (*videoTrack, error) {
	err := d.Open()
	if err != nil {
		return nil, err
	}

	vr := d.(driver.VideoRecorder)
	r, err := vr.VideoRecord(constraints.Media)
	if err != nil {
		return nil, err
	}

	if constraints.VideoTransform != nil {
		r = constraints.VideoTransform(r)
	}

	var vt *videoTrack
	rtpCodecs := opts.codecs[webrtc.RTPCodecTypeVideo]
	for _, codecBuilder := range constraints.VideoEncoderBuilders {
		var matchedRTPCodec *webrtc.RTPCodec
		for _, rtpCodec := range rtpCodecs {
			if rtpCodec.Name == codecBuilder.Name() {
				matchedRTPCodec = rtpCodec
				break
			}
		}

		if matchedRTPCodec == nil {
			continue
		}

		t, err := newTrack(matchedRTPCodec, opts.trackGenerator, d)
		if err != nil {
			continue
		}

		encoder, err := codecBuilder.BuildVideoEncoder(r, constraints.Media)
		if err != nil {
			continue
		}

		vt = &videoTrack{
			track:       t,
			d:           d,
			constraints: constraints,
			encoder:     encoder,
		}
		break
	}

	if vt == nil {
		d.Close()
		return nil, fmt.Errorf("failed to find a matching video codec")
	}

	go vt.start()
	return vt, nil
}

func (vt *videoTrack) start() {
	var n int
	var err error
	buff := make([]byte, 1024)
	for {
		n, err = vt.encoder.Read(buff)
		if err != nil {
			if e, ok := err.(*mio.InsufficientBufferError); ok {
				buff = make([]byte, 2*e.RequiredSize)
				continue
			}

			vt.track.onError(err)
			return
		}

		if err := vt.s.sample(buff[:n]); err != nil {
			vt.track.onError(err)
			return
		}
	}
}

func (vt *videoTrack) Stop() {
	vt.d.Close()
	vt.encoder.Close()
}

type audioTrack struct {
	*track
	d           driver.Driver
	constraints MediaTrackConstraints
	encoder     codec.ReadCloser
}

var _ Tracker = &audioTrack{}

func newAudioTrack(opts *MediaDevicesOptions, d driver.Driver, constraints MediaTrackConstraints) (*audioTrack, error) {
	err := d.Open()
	if err != nil {
		return nil, err
	}

	ar := d.(driver.AudioRecorder)
	r, err := ar.AudioRecord(constraints.Media)
	if err != nil {
		return nil, err
	}

	if constraints.AudioTransform != nil {
		r = constraints.AudioTransform(r)
	}

	var at *audioTrack
	rtpCodecs := opts.codecs[webrtc.RTPCodecTypeAudio]
	for _, codecBuilder := range constraints.AudioEncoderBuilders {
		var matchedRTPCodec *webrtc.RTPCodec
		for _, rtpCodec := range rtpCodecs {
			if rtpCodec.Name == codecBuilder.Name() {
				matchedRTPCodec = rtpCodec
				break
			}
		}

		if matchedRTPCodec == nil {
			continue
		}

		t, err := newTrack(matchedRTPCodec, opts.trackGenerator, d)
		if err != nil {
			continue
		}

		encoder, err := codecBuilder.BuildAudioEncoder(r, constraints.Media)
		if err != nil {
			continue
		}

		at = &audioTrack{
			track:       t,
			d:           d,
			constraints: constraints,
			encoder:     encoder,
		}
	}

	if at == nil {
		d.Close()
		return nil, fmt.Errorf("failed to find a matching audio codec")
	}

	go at.start()
	return at, nil
}

func (t *audioTrack) start() {
	buff := make([]byte, 1024)
	sampleSize := uint32(float64(t.constraints.SampleRate) * t.constraints.Latency.Seconds())
	for {
		n, err := t.encoder.Read(buff)
		if err != nil {
			t.track.onError(err)
			return
		}

		if err := t.t.WriteSample(media.Sample{
			Data:    buff[:n],
			Samples: sampleSize,
		}); err != nil {
			t.track.onError(err)
			return
		}
	}
}

func (t *audioTrack) Stop() {
	t.d.Close()
	t.encoder.Close()
}
