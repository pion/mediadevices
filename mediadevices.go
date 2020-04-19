package mediadevices

import (
	"fmt"
	"math"

	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/driver"
	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/webrtc/v2"
)

var errNotFound = fmt.Errorf("failed to find the best driver that fits the constraints")

// MediaDevices is an interface that's defined on https://developer.mozilla.org/en-US/docs/Web/API/MediaDevices
type MediaDevices interface {
	GetDisplayMedia(constraints MediaStreamConstraints) (MediaStream, error)
	GetUserMedia(constraints MediaStreamConstraints) (MediaStream, error)
	EnumerateDevices() []MediaDeviceInfo
}

// NewMediaDevices creates MediaDevices interface that provides access to connected media input devices
// like cameras and microphones, as well as screen sharing.
// In essence, it lets you obtain access to any hardware source of media data.
func NewMediaDevices(pc *webrtc.PeerConnection, opts ...MediaDevicesOption) MediaDevices {
	codecs := make(map[webrtc.RTPCodecType][]*webrtc.RTPCodec)
	for _, kind := range []webrtc.RTPCodecType{
		webrtc.RTPCodecTypeAudio,
		webrtc.RTPCodecTypeVideo,
	} {
		codecs[kind] = pc.GetRegisteredRTPCodecs(kind)
	}
	return NewMediaDevicesFromCodecs(codecs, opts...)
}

// NewMediaDevicesFromCodecs creates MediaDevices interface from lists of the available codecs
// that provides access to connected media input devices like cameras and microphones,
// as well as screen sharing.
// In essence, it lets you obtain access to any hardware source of media data.
func NewMediaDevicesFromCodecs(codecs map[webrtc.RTPCodecType][]*webrtc.RTPCodec, opts ...MediaDevicesOption) MediaDevices {
	mdo := MediaDevicesOptions{
		codecs:         codecs,
		trackGenerator: defaultTrackGenerator,
	}
	for _, o := range opts {
		o(&mdo)
	}
	return &mediaDevices{
		MediaDevicesOptions: mdo,
	}
}

// TrackGenerator is a function to create new track.
type TrackGenerator func(payloadType uint8, ssrc uint32, id, label string, codec *webrtc.RTPCodec) (LocalTrack, error)

var defaultTrackGenerator = TrackGenerator(func(pt uint8, ssrc uint32, id, label string, codec *webrtc.RTPCodec) (LocalTrack, error) {
	return webrtc.NewTrack(pt, ssrc, id, label, codec)
})

type mediaDevices struct {
	MediaDevicesOptions
}

// MediaDevicesOptions stores parameters used by MediaDevices.
type MediaDevicesOptions struct {
	codecs               map[webrtc.RTPCodecType][]*webrtc.RTPCodec
	trackGenerator       TrackGenerator
	videoEncoderBuilders []codec.VideoEncoderBuilder
	audioEncoderBuilders []codec.AudioEncoderBuilder
	videoTransform       video.TransformFunc
	audioTransform       audio.TransformFunc
}

// MediaDevicesOption is a type of MediaDevices functional option.
type MediaDevicesOption func(*MediaDevicesOptions)

// WithTrackGenerator specifies a TrackGenerator to use customized track.
func WithTrackGenerator(gen TrackGenerator) MediaDevicesOption {
	return func(o *MediaDevicesOptions) {
		o.trackGenerator = gen
	}
}

// WithVideoEncoders specifies available video encoders.
// encoders are codec builders that are used for encoding the video
// and later being used for sending the appropriate RTP payload type.
//
// If one encoder builder fails to build the codec, the next builder will be used,
// repeating until a codec builds. If no builders build successfully, an error is returned.
func WithVideoEncoders(encoders ...codec.VideoEncoderBuilder) MediaDevicesOption {
	return func(o *MediaDevicesOptions) {
		o.videoEncoderBuilders = encoders
	}
}

// WithAudioEncoders specifies available audio encoders
// encoders are codec builders that are used for encoding the audio
// and later being used for sending the appropriate RTP payload type.
//
// If one encoder builder fails to build the codec, the next builder will be used,
// repeating until a codec builds. If no builders build successfully, an error is returned.
func WithAudioEncoders(encoders ...codec.AudioEncoderBuilder) MediaDevicesOption {
	return func(o *MediaDevicesOptions) {
		o.audioEncoderBuilders = encoders
	}
}

// WithVideoTransformers will be used to transform the video that's coming from the driver.
// So, basically it'll look like following: driver -> VideoTransform -> codec
func WithVideoTransformers(transformFuncs ...video.TransformFunc) MediaDevicesOption {
	return func(o *MediaDevicesOptions) {
		o.videoTransform = video.Merge(transformFuncs...)
	}
}

// WithAudioTransformers will be used to transform the audio that's coming from the driver.
// So, basically it'll look like following: driver -> AudioTransform -> code
func WithAudioTransformers(transformFuncs ...audio.TransformFunc) MediaDevicesOption {
	return func(o *MediaDevicesOptions) {
		o.audioTransform = audio.Merge(transformFuncs...)
	}
}

// GetDisplayMedia prompts the user to select and grant permission to capture the contents
// of a display or portion thereof (such as a window) as a MediaStream.
// Reference: https://developer.mozilla.org/en-US/docs/Web/API/MediaDevices/getDisplayMedia
func (m *mediaDevices) GetDisplayMedia(constraints MediaStreamConstraints) (MediaStream, error) {
	trackers := make([]Tracker, 0)

	cleanTrackers := func() {
		for _, t := range trackers {
			t.Stop()
		}
	}

	if constraints.Video != nil {
		var p prop.Media
		constraints.Video(&p)
		tracker, err := m.selectScreen(p)
		if err != nil {
			cleanTrackers()
			return nil, err
		}

		trackers = append(trackers, tracker)
	}

	s, err := NewMediaStream(trackers...)
	if err != nil {
		cleanTrackers()
		return nil, err
	}

	return s, nil
}

// GetUserMedia prompts the user for permission to use a media input which produces a MediaStream
// with tracks containing the requested types of media.
// Reference: https://developer.mozilla.org/en-US/docs/Web/API/MediaDevices/getUserMedia
func (m *mediaDevices) GetUserMedia(constraints MediaStreamConstraints) (MediaStream, error) {
	// TODO: It should return media stream based on constraints
	trackers := make([]Tracker, 0)

	cleanTrackers := func() {
		for _, t := range trackers {
			t.Stop()
		}
	}

	if constraints.Video != nil {
		var p prop.Media
		constraints.Video(&p)
		tracker, err := m.selectVideo(p)
		if err != nil {
			cleanTrackers()
			return nil, err
		}

		trackers = append(trackers, tracker)
	}

	if constraints.Audio != nil {
		var p prop.Media
		constraints.Audio(&p)
		tracker, err := m.selectAudio(p)
		if err != nil {
			cleanTrackers()
			return nil, err
		}

		trackers = append(trackers, tracker)
	}

	s, err := NewMediaStream(trackers...)
	if err != nil {
		cleanTrackers()
		return nil, err
	}

	return s, nil
}

func queryDriverProperties(filter driver.FilterFn) map[driver.Driver][]prop.Media {
	var needToClose []driver.Driver
	drivers := driver.GetManager().Query(filter)
	m := make(map[driver.Driver][]prop.Media)

	for _, d := range drivers {
		if d.Status() == driver.StateClosed {
			err := d.Open()
			if err != nil {
				// Skip this driver if we failed to open because we can't get the properties
				continue
			}
			needToClose = append(needToClose, d)
		}

		m[d] = d.Properties()
	}

	for _, d := range needToClose {
		// Since it was closed, we should close it to avoid a leak
		d.Close()
	}

	return m
}

// select implements SelectSettings algorithm.
// Reference: https://w3c.github.io/mediacapture-main/#dfn-selectsettings
func selectBestDriver(filter driver.FilterFn, constraints prop.Media) (driver.Driver, prop.Media, error) {
	var bestDriver driver.Driver
	var bestProp prop.Media
	minFitnessDist := math.Inf(1)

	driverProperties := queryDriverProperties(filter)
	for d, props := range driverProperties {
		priority := float64(d.Info().Priority)
		for _, p := range props {
			fitnessDist := constraints.FitnessDistance(p) - priority
			if fitnessDist < minFitnessDist {
				minFitnessDist = fitnessDist
				bestDriver = d
				bestProp = p
			}
		}
	}

	if bestDriver == nil {
		return nil, prop.Media{}, errNotFound
	}

	constraints.Merge(bestProp)
	return bestDriver, constraints, nil
}

func (m *mediaDevices) selectAudio(constraints prop.Media) (Tracker, error) {
	typeFilter := driver.FilterAudioRecorder()
	filter := typeFilter
	if constraints.DeviceID != "" {
		idFilter := driver.FilterID(constraints.DeviceID)
		filter = driver.FilterAnd(typeFilter, idFilter)
	}

	d, c, err := selectBestDriver(filter, constraints)
	if err != nil {
		return nil, err
	}

	return newTrack(&m.MediaDevicesOptions, d, c)
}

func (m *mediaDevices) selectVideo(constraints prop.Media) (Tracker, error) {
	typeFilter := driver.FilterVideoRecorder()
	notScreenFilter := driver.FilterNot(driver.FilterDeviceType(driver.Screen))
	filter := driver.FilterAnd(typeFilter, notScreenFilter)
	if constraints.DeviceID != "" {
		idFilter := driver.FilterID(constraints.DeviceID)
		filter = driver.FilterAnd(typeFilter, notScreenFilter, idFilter)
	}

	d, c, err := selectBestDriver(filter, constraints)
	if err != nil {
		return nil, err
	}

	return newTrack(&m.MediaDevicesOptions, d, c)
}

func (m *mediaDevices) selectScreen(constraints prop.Media) (Tracker, error) {
	typeFilter := driver.FilterVideoRecorder()
	screenFilter := driver.FilterDeviceType(driver.Screen)
	filter := driver.FilterAnd(typeFilter, screenFilter)
	if constraints.DeviceID != "" {
		idFilter := driver.FilterID(constraints.DeviceID)
		filter = driver.FilterAnd(typeFilter, screenFilter, idFilter)
	}

	d, c, err := selectBestDriver(filter, constraints)
	if err != nil {
		return nil, err
	}

	return newTrack(&m.MediaDevicesOptions, d, c)
}

func (m *mediaDevices) EnumerateDevices() []MediaDeviceInfo {
	drivers := driver.GetManager().Query(
		driver.FilterFn(func(driver.Driver) bool { return true }))
	info := make([]MediaDeviceInfo, 0, len(drivers))
	for _, d := range drivers {
		var kind MediaDeviceType
		switch {
		case driver.FilterVideoRecorder()(d):
			kind = VideoInput
		case driver.FilterAudioRecorder()(d):
			kind = AudioInput
		default:
			continue
		}
		driverInfo := d.Info()
		info = append(info, MediaDeviceInfo{
			DeviceID:   d.ID(),
			Kind:       kind,
			Label:      driverInfo.Label,
			DeviceType: driverInfo.DeviceType,
		})
	}
	return info
}
