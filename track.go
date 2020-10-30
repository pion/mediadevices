package mediadevices

import (
	"errors"
	"fmt"
	"image"
	"math/rand"
	"sync"

	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/driver"
	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/wave"
	"github.com/pion/webrtc/v2"
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
	Kind() MediaDeviceType
	// Bind binds the current track source to the given peer connection. In Pion/webrtc v3, the bind
	// call will happen automatically after the SDP negotiation. Users won't need to call this manually.
	Bind(*webrtc.PeerConnection) (*webrtc.Track, error)
	// Unbind is the clean up operation that should be called after Bind. Similar to Bind, unbind will
	// be called automatically in the future.
	Unbind(*webrtc.PeerConnection) error
}

type baseTrack struct {
	Source
	err                   error
	onErrorHandler        func(error)
	mu                    sync.Mutex
	endOnce               sync.Once
	kind                  MediaDeviceType
	selector              *CodecSelector
	activePeerConnections map[*webrtc.PeerConnection]chan<- chan<- struct{}
}

func newBaseTrack(source Source, kind MediaDeviceType, selector *CodecSelector) *baseTrack {
	return &baseTrack{
		Source:                source,
		kind:                  kind,
		selector:              selector,
		activePeerConnections: make(map[*webrtc.PeerConnection]chan<- chan<- struct{}),
	}
}

// Kind returns track's kind
func (track *baseTrack) Kind() MediaDeviceType {
	return track.kind
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

func (track *baseTrack) bind(pc *webrtc.PeerConnection, encodedReader codec.ReadCloser, selectedCodec *codec.RTPCodec, sampler func(*webrtc.Track) samplerFunc) (*webrtc.Track, error) {
	track.mu.Lock()
	defer track.mu.Unlock()

	webrtcTrack, err := pc.NewTrack(selectedCodec.PayloadType, rand.Uint32(), track.ID(), selectedCodec.MimeType)
	if err != nil {
		return nil, err
	}

	sample := sampler(webrtcTrack)
	signalCh := make(chan chan<- struct{})
	track.activePeerConnections[pc] = signalCh

	fmt.Println("Binding")

	go func() {
		var doneCh chan<- struct{}
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

			buff, _, err := encodedReader.Read()
			if err != nil {
				track.onError(err)
				return
			}

			if err := sample(buff); err != nil {
				track.onError(err)
				return
			}
		}
	}()

	return webrtcTrack, nil
}

func (track *baseTrack) unbind(pc *webrtc.PeerConnection) error {
	track.mu.Lock()
	defer track.mu.Unlock()

	ch, ok := track.activePeerConnections[pc]
	if !ok {
		return errNotFoundPeerConnection
	}

	doneCh := make(chan struct{})
	ch <- doneCh
	<-doneCh
	delete(track.activePeerConnections, pc)
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

func (track *VideoTrack) Bind(pc *webrtc.PeerConnection) (*webrtc.Track, error) {
	reader := track.NewReader(false)
	inputProp, err := detectCurrentVideoProp(track.Broadcaster)
	if err != nil {
		return nil, err
	}

	wantCodecs := pc.GetRegisteredRTPCodecs(webrtc.RTPCodecTypeVideo)
	fmt.Println(wantCodecs)
	fmt.Println(&inputProp)
	encodedReader, selectedCodec, err := track.selector.selectVideoCodec(wantCodecs, reader, inputProp)
	if err != nil {
		return nil, err
	}

	return track.bind(pc, encodedReader, selectedCodec, newVideoSampler)
}

func (track *VideoTrack) Unbind(pc *webrtc.PeerConnection) error {
	return track.unbind(pc)
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

func (track *AudioTrack) Bind(pc *webrtc.PeerConnection) (*webrtc.Track, error) {
	reader := track.NewReader(false)
	inputProp, err := detectCurrentAudioProp(track.Broadcaster)
	if err != nil {
		return nil, err
	}

	wantCodecs := pc.GetRegisteredRTPCodecs(webrtc.RTPCodecTypeAudio)
	encodedReader, selectedCodec, err := track.selector.selectAudioCodec(wantCodecs, reader, inputProp)
	if err != nil {
		return nil, err
	}

	return track.bind(pc, encodedReader, selectedCodec, func(t *webrtc.Track) samplerFunc { return newAudioSampler(t, inputProp.Latency) })
}

func (track *AudioTrack) Unbind(pc *webrtc.PeerConnection) error {
	return track.unbind(pc)
}
