package mediadevices

import (
	"fmt"

	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v2"
)

type RTPTracker struct {
	videoEncoders []codec.VideoEncoderBuilder
	audioEncoders []codec.AudioEncoderBuilder
}

type RTPTrackerOption func(*RTPTracker)

func WithVideoEncoders(codecs ...codec.VideoEncoderBuilder) func(*RTPTracker) {
	return func(tracker *RTPTracker) {
		tracker.videoEncoders = codecs
	}
}

func WithAudioEncoders(codecs ...codec.AudioEncoderBuilder) func(*RTPTracker) {
	return func(tracker *RTPTracker) {
		tracker.audioEncoders = codecs
	}
}

func NewRTPTracker(opts ...RTPTrackerOption) *RTPTracker {
	var tracker RTPTracker

	for _, opt := range opts {
		opt(&tracker)
	}

	return &tracker
}

func (tracker *RTPTracker) Track(track Track) *RTPTrack {
	rtpTrack := RTPTrack{
		Track: track,
	}

	return &rtpTrack
}

type RTPTrack struct {
	Track
	tracker        *RTPTracker
	currentEncoder codec.ReadCloser
	currentParams  RTPParameters
	lastProp       prop.Media
}

func (track *RTPTrack) SetParameters(params RTPParameters) error {
	var err error

	switch t := track.Track.(type) {
	case *VideoTrack:
		err = track.setParametersVideo(t, &params)
	case *AudioTrack:
		err = track.setParametersAudio(t, &params)
	default:
		err = fmt.Errorf("unsupported track type")
	}

	if err == nil {
		track.currentParams = params
	}

	return err
}

func (track *RTPTrack) setParametersVideo(videoTrack *VideoTrack, params *RTPParameters) error {
	if params.SelectedCodec.Type != webrtc.RTPCodecTypeVideo {
		return fmt.Errorf("invalid selected RTP codec type. Expected video but got audio")
	}

	video.DetectChanges(interval time.Duration, onChange func(prop.Media))
	return nil
}

func (track *RTPTrack) setParametersAudio(audioTrack *AudioTrack, params *RTPParameters) error {
	return nil
}

func (track *RTPTrack) ReadRTP() (*rtp.Packet, error) {
	if track.currentEncoder == nil {
		return nil, fmt.Errorf("Encoder has not been specified. Please call SetParameters to specify.")
	}

	return nil, nil
}

func (track *RTPTrack) WriteRTCP(packet rtcp.Packet) error {
	return nil
}
