package mediadevices

import (
	"errors"
	"fmt"
	"image"
	"io"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/driver"
	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/wave"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

const (
	rtpOutboundMTU = 1200
)

var (
	errInvalidDriverType      = errors.New("invalid driver type")
	errNotFoundPeerConnection = errors.New("failed to find given peer connection")
)

// Source is a generic representation of a media source
type Source interface {
	ID() string
	Close() error
}

// VideoSource is a specific type of media source that emits a series of video frames
type VideoSource interface {
	video.Reader
	Source
}

// AudioSource is a specific type of media source that emits a series of audio chunks
type AudioSource interface {
	audio.Reader
	Source
}

// Track is an interface that represent MediaStreamTrack
// Reference: https://w3c.github.io/mediacapture-main/#mediastreamtrack
type Track interface {
	Source
	// OnEnded registers a handler to receive an error from the media stream track.
	// If the error is already occured before registering, the handler will be
	// immediately called.
	OnEnded(func(error))
	Kind() webrtc.RTPCodecType
	// StreamID is the group this track belongs too. This must be unique
	StreamID() string
	// Bind binds the current track source to the given peer connection. In Pion/webrtc v3, the bind
	// call will happen automatically after the SDP negotiation. Users won't need to call this manually.
	Bind(webrtc.TrackLocalContext) (webrtc.RTPCodecParameters, error)
	// Unbind is the clean up operation that should be called after Bind. Similar to Bind, unbind will
	// be called automatically in Pion/webrtc v3.
	Unbind(webrtc.TrackLocalContext) error
	// NewRTPReader creates a new reader from the source. The reader will encode the source, and packetize
	// the encoded data in RTP format with given mtu size.
	NewRTPReader(codecName string, ssrc uint32, mtu int) (RTPReadCloser, error)
	// NewEncodedReader creates a EncodedReadCloser that reads the encoded data in codecName format
	NewEncodedReader(codecName string) (EncodedReadCloser, error)
	// NewEncodedReader creates a new Go standard io.ReadCloser that reads the encoded data in codecName format
	NewEncodedIOReader(codecName string) (io.ReadCloser, error)
}

type baseTrack struct {
	Source
	err                   error
	onErrorHandler        func(error)
	mu                    sync.Mutex
	endOnce               sync.Once
	kind                  MediaDeviceType
	selector              *CodecSelector
	activePeerConnections map[string]chan<- chan<- struct{}
}

func newBaseTrack(source Source, kind MediaDeviceType, selector *CodecSelector) *baseTrack {
	return &baseTrack{
		Source:                source,
		kind:                  kind,
		selector:              selector,
		activePeerConnections: make(map[string]chan<- chan<- struct{}),
	}
}

// Kind returns track's kind
func (track *baseTrack) Kind() webrtc.RTPCodecType {
	switch track.kind {
	case VideoInput:
		return webrtc.RTPCodecTypeVideo
	case AudioInput:
		return webrtc.RTPCodecTypeAudio
	default:
		panic("invalid track kind: only support VideoInput and AudioInput")
	}
}

func (track *baseTrack) StreamID() string {
	// TODO: StreamID should be used to group multiple tracks. Should get this information from mediastream instead.
	generator, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}

	return generator.String()
}

// OnEnded sets an error handler. When a track has been created and started, if an
// error occurs, handler will get called with the error given to the parameter.
func (track *baseTrack) OnEnded(handler func(error)) {
	track.mu.Lock()
	track.onErrorHandler = handler
	err := track.err
	track.mu.Unlock()

	if err != nil && handler != nil {
		// Already errored.
		track.endOnce.Do(func() {
			handler(err)
		})
	}
}

// onError is a callback when an error occurs
func (track *baseTrack) onError(err error) {
	track.mu.Lock()
	track.err = err
	handler := track.onErrorHandler
	track.mu.Unlock()

	if handler != nil {
		track.endOnce.Do(func() {
			handler(err)
		})
	}
}

func (track *baseTrack) bind(ctx webrtc.TrackLocalContext, specializedTrack Track) (webrtc.RTPCodecParameters, error) {
	track.mu.Lock()
	defer track.mu.Unlock()

	signalCh := make(chan chan<- struct{})
	track.activePeerConnections[ctx.ID()] = signalCh

	var encodedReader RTPReadCloser
	var selectedCodec webrtc.RTPCodecParameters
	var err error
	var errReasons []string
	for _, wantedCodec := range ctx.CodecParameters() {
		logger.Debugf("trying to build %s rtp reader", wantedCodec.MimeType)
		encodedReader, err = specializedTrack.NewRTPReader(wantedCodec.MimeType, uint32(ctx.SSRC()), rtpOutboundMTU)
		if err == nil {
			selectedCodec = wantedCodec
			break
		}

		errReasons = append(errReasons, fmt.Sprintf("%s: %s", wantedCodec.MimeType, err))
	}

	if encodedReader == nil {
		return webrtc.RTPCodecParameters{}, errors.New(strings.Join(errReasons, "\n\n"))
	}

	go func() {
		var doneCh chan<- struct{}
		writer := ctx.WriteStream()
		defer func() {
			encodedReader.Close()

			// When there's another call to unbind, it won't block since we mark the signalCh to be closed
			close(signalCh)
			if doneCh != nil {
				close(doneCh)
			}
		}()

		for {
			select {
			case doneCh = <-signalCh:
				return
			default:
			}

			pkts, _, err := encodedReader.Read()
			if err != nil {
				// explicitly ignore this error since the higher level should've reported this
				return
			}

			for _, pkt := range pkts {
				_, err = writer.WriteRTP(&pkt.Header, pkt.Payload)
				if err != nil {
					track.onError(err)
					return
				}
			}
		}
	}()

	return selectedCodec, nil
}

func (track *baseTrack) unbind(ctx webrtc.TrackLocalContext) error {
	track.mu.Lock()
	defer track.mu.Unlock()

	ch, ok := track.activePeerConnections[ctx.ID()]
	if !ok {
		return errNotFoundPeerConnection
	}

	doneCh := make(chan struct{})
	ch <- doneCh
	<-doneCh
	delete(track.activePeerConnections, ctx.ID())
	return nil
}

func newTrackFromDriver(d driver.Driver, constraints MediaTrackConstraints, selector *CodecSelector) (Track, error) {
	if err := d.Open(); err != nil {
		return nil, err
	}

	switch recorder := d.(type) {
	case driver.VideoRecorder:
		return newVideoTrackFromDriver(d, recorder, constraints, selector)
	case driver.AudioRecorder:
		return newAudioTrackFromDriver(d, recorder, constraints, selector)
	default:
		panic(errInvalidDriverType)
	}
}

// VideoTrack is a specific track type that contains video source which allows multiple readers to access, and manipulate.
type VideoTrack struct {
	*baseTrack
	*video.Broadcaster
}

// NewVideoTrack constructs a new VideoTrack
func NewVideoTrack(source VideoSource, selector *CodecSelector) Track {
	return newVideoTrackFromReader(source, source, selector)
}

func newVideoTrackFromReader(source Source, reader video.Reader, selector *CodecSelector) Track {
	base := newBaseTrack(source, VideoInput, selector)
	wrappedReader := video.ReaderFunc(func() (img image.Image, release func(), err error) {
		img, _, err = reader.Read()
		if err != nil {
			base.onError(err)
		}
		return img, func() {}, err
	})

	// TODO: Allow users to configure broadcaster
	broadcaster := video.NewBroadcaster(wrappedReader, nil)

	return &VideoTrack{
		baseTrack:   base,
		Broadcaster: broadcaster,
	}
}

// newVideoTrackFromDriver is an internal video track creation from driver
func newVideoTrackFromDriver(d driver.Driver, recorder driver.VideoRecorder, constraints MediaTrackConstraints, selector *CodecSelector) (Track, error) {
	reader, err := recorder.VideoRecord(constraints.selectedMedia)
	if err != nil {
		return nil, err
	}

	return newVideoTrackFromReader(d, reader, selector), nil
}

// Transform transforms the underlying source by applying the given fns in serial order
func (track *VideoTrack) Transform(fns ...video.TransformFunc) {
	src := track.Broadcaster.Source()
	track.Broadcaster.ReplaceSource(video.Merge(fns...)(src))
}

func (track *VideoTrack) Bind(ctx webrtc.TrackLocalContext) (webrtc.RTPCodecParameters, error) {
	return track.bind(ctx, track)
}

func (track *VideoTrack) Unbind(ctx webrtc.TrackLocalContext) error {
	return track.unbind(ctx)
}

func (track *VideoTrack) newEncodedReader(codecNames ...string) (EncodedReadCloser, *codec.RTPCodec, error) {
	reader := track.NewReader(false)
	inputProp, err := detectCurrentVideoProp(track.Broadcaster)
	if err != nil {
		return nil, nil, err
	}

	encodedReader, selectedCodec, err := track.selector.selectVideoCodecByNames(reader, inputProp, codecNames...)
	if err != nil {
		return nil, nil, err
	}

	sample := newVideoSampler(selectedCodec.ClockRate)

	return &encodedReadCloserImpl{
		readFn: func() (EncodedBuffer, func(), error) {
			data, release, err := encodedReader.Read()
			buffer := EncodedBuffer{
				Data:    data,
				Samples: sample(),
			}
			return buffer, release, err
		},
		closeFn: encodedReader.Close,
	}, selectedCodec, nil
}

func (track *VideoTrack) NewEncodedReader(codecName string) (EncodedReadCloser, error) {
	reader, _, err := track.newEncodedReader(codecName)
	return reader, err
}

func (track *VideoTrack) NewEncodedIOReader(codecName string) (io.ReadCloser, error) {
	encodedReader, _, err := track.newEncodedReader(codecName)
	if err != nil {
		return nil, err
	}
	return newEncodedIOReadCloserImpl(encodedReader), nil
}

func (track *VideoTrack) NewRTPReader(codecName string, ssrc uint32, mtu int) (RTPReadCloser, error) {
	encodedReader, selectedCodec, err := track.newEncodedReader(codecName)
	if err != nil {
		return nil, err
	}

	packetizer := rtp.NewPacketizer(mtu, uint8(selectedCodec.PayloadType), ssrc, selectedCodec.Payloader, rtp.NewRandomSequencer(), selectedCodec.ClockRate)

	return &rtpReadCloserImpl{
		readFn: func() ([]*rtp.Packet, func(), error) {
			encoded, release, err := encodedReader.Read()
			if err != nil {
				encodedReader.Close()
				track.onError(err)
				return nil, func() {}, err
			}
			defer release()

			pkts := packetizer.Packetize(encoded.Data, encoded.Samples)
			return pkts, release, err
		},
		closeFn: encodedReader.Close,
	}, nil
}

// AudioTrack is a specific track type that contains audio source which allows multiple readers to access, and
// manipulate.
type AudioTrack struct {
	*baseTrack
	*audio.Broadcaster
}

// NewAudioTrack constructs a new VideoTrack
func NewAudioTrack(source AudioSource, selector *CodecSelector) Track {
	return newAudioTrackFromReader(source, source, selector)
}

func newAudioTrackFromReader(source Source, reader audio.Reader, selector *CodecSelector) Track {
	base := newBaseTrack(source, AudioInput, selector)
	wrappedReader := audio.ReaderFunc(func() (chunk wave.Audio, release func(), err error) {
		chunk, _, err = reader.Read()
		if err != nil {
			base.onError(err)
		}
		return chunk, func() {}, err
	})

	// TODO: Allow users to configure broadcaster
	broadcaster := audio.NewBroadcaster(wrappedReader, nil)

	return &AudioTrack{
		baseTrack:   base,
		Broadcaster: broadcaster,
	}
}

// newAudioTrackFromDriver is an internal audio track creation from driver
func newAudioTrackFromDriver(d driver.Driver, recorder driver.AudioRecorder, constraints MediaTrackConstraints, selector *CodecSelector) (Track, error) {
	reader, err := recorder.AudioRecord(constraints.selectedMedia)
	if err != nil {
		return nil, err
	}

	return newAudioTrackFromReader(d, reader, selector), nil
}

// Transform transforms the underlying source by applying the given fns in serial order
func (track *AudioTrack) Transform(fns ...audio.TransformFunc) {
	src := track.Broadcaster.Source()
	track.Broadcaster.ReplaceSource(audio.Merge(fns...)(src))
}

func (track *AudioTrack) Bind(ctx webrtc.TrackLocalContext) (webrtc.RTPCodecParameters, error) {
	return track.bind(ctx, track)
}

func (track *AudioTrack) Unbind(ctx webrtc.TrackLocalContext) error {
	return track.unbind(ctx)
}

func (track *AudioTrack) newEncodedReader(codecNames ...string) (EncodedReadCloser, *codec.RTPCodec, error) {
	reader := track.NewReader(false)
	inputProp, err := detectCurrentAudioProp(track.Broadcaster)
	if err != nil {
		return nil, nil, err
	}

	encodedReader, selectedCodec, err := track.selector.selectAudioCodecByNames(reader, inputProp, codecNames...)
	if err != nil {
		return nil, nil, err
	}

	sample := newAudioSampler(selectedCodec.ClockRate, inputProp.Latency)

	return &encodedReadCloserImpl{
		readFn: func() (EncodedBuffer, func(), error) {
			data, release, err := encodedReader.Read()
			buffer := EncodedBuffer{
				Data:    data,
				Samples: sample(),
			}
			return buffer, release, err
		},
		closeFn: encodedReader.Close,
	}, selectedCodec, nil
}

func (track *AudioTrack) NewEncodedReader(codecName string) (EncodedReadCloser, error) {
	reader, _, err := track.newEncodedReader(codecName)
	return reader, err
}

func (track *AudioTrack) NewEncodedIOReader(codecName string) (io.ReadCloser, error) {
	encodedReader, _, err := track.newEncodedReader(codecName)
	if err != nil {
		return nil, err
	}
	return newEncodedIOReadCloserImpl(encodedReader), nil
}

func (track *AudioTrack) NewRTPReader(codecName string, ssrc uint32, mtu int) (RTPReadCloser, error) {
	encodedReader, selectedCodec, err := track.newEncodedReader(codecName)
	if err != nil {
		return nil, err
	}

	packetizer := rtp.NewPacketizer(mtu, uint8(selectedCodec.PayloadType), ssrc, selectedCodec.Payloader, rtp.NewRandomSequencer(), selectedCodec.ClockRate)

	return &rtpReadCloserImpl{
		readFn: func() ([]*rtp.Packet, func(), error) {
			encoded, release, err := encodedReader.Read()
			if err != nil {
				encodedReader.Close()
				track.onError(err)
				return nil, func() {}, err
			}
			defer release()

			pkts := packetizer.Packetize(encoded.Data, encoded.Samples)
			return pkts, release, err
		},
		closeFn: encodedReader.Close,
	}, nil
}
