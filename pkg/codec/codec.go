package codec

import (
	"time"

	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v4"
)

// RTPCodec wraps webrtc.RTPCodec. RTPCodec might extend webrtc.RTPCodec in the future.
type RTPCodec struct {
	webrtc.RTPCodecParameters
	rtp.Payloader

	// Latency of static frame size codec.
	Latency time.Duration
}

// NewRTPH264Codec is a helper to create an H264 codec
func NewRTPH264Codec(clockrate uint32) *RTPCodec {
	return &RTPCodec{
		RTPCodecParameters: webrtc.RTPCodecParameters{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:     webrtc.MimeTypeH264,
				ClockRate:    90000,
				Channels:     0,
				SDPFmtpLine:  "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42e01f",
				RTCPFeedback: nil,
			},
			PayloadType: 125,
		},
		Payloader: &codecs.H264Payloader{},
	}
}

// NewRTPH265Codec is a helper to create an H265 codec
func NewRTPH265Codec(clockrate uint32) *RTPCodec {
	return &RTPCodec{
		RTPCodecParameters: webrtc.RTPCodecParameters{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:     webrtc.MimeTypeH265,
				ClockRate:    90000,
				Channels:     0,
				SDPFmtpLine:  "",
				RTCPFeedback: nil,
			},
			PayloadType: 116,
		},
		Payloader: &codecs.H265Payloader{},
	}
}

// NewRTPVP8Codec is a helper to create an VP8 codec
func NewRTPVP8Codec(clockrate uint32) *RTPCodec {
	return &RTPCodec{
		RTPCodecParameters: webrtc.RTPCodecParameters{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:     webrtc.MimeTypeVP8,
				ClockRate:    90000,
				Channels:     0,
				SDPFmtpLine:  "",
				RTCPFeedback: nil,
			},
			PayloadType: 96,
		},
		Payloader: &codecs.VP8Payloader{},
	}
}

// NewRTPVP9Codec is a helper to create an VP9 codec
func NewRTPVP9Codec(clockrate uint32) *RTPCodec {
	return &RTPCodec{
		RTPCodecParameters: webrtc.RTPCodecParameters{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:     webrtc.MimeTypeVP9,
				ClockRate:    90000,
				Channels:     0,
				SDPFmtpLine:  "",
				RTCPFeedback: nil,
			},
			PayloadType: 98,
		},
		Payloader: &codecs.VP9Payloader{},
	}
}

// NewRTPAV1Codec is a helper to create an AV1 codec
func NewRTPAV1Codec(clockrate uint32) *RTPCodec {
	return &RTPCodec{
		RTPCodecParameters: webrtc.RTPCodecParameters{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:     webrtc.MimeTypeAV1,
				ClockRate:    90000,
				Channels:     0,
				SDPFmtpLine:  "level-idx=5;profile=0;tier=0",
				RTCPFeedback: nil,
			},
			PayloadType: 99,
		},
		Payloader: &codecs.AV1Payloader{},
	}
}

// NewRTPOpusCodec is a helper to create an Opus codec
func NewRTPOpusCodec(clockrate uint32) *RTPCodec {
	return &RTPCodec{
		RTPCodecParameters: webrtc.RTPCodecParameters{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:     webrtc.MimeTypeOpus,
				ClockRate:    48000,
				Channels:     2,
				SDPFmtpLine:  "minptime=10;useinbandfec=1",
				RTCPFeedback: nil,
			},
			PayloadType: 111,
		},
		Payloader: &codecs.OpusPayloader{},
	}
}

// AudioEncoderBuilder is the interface that wraps basic operations that are
// necessary to build the audio encoder.
//
// This interface is for codec implementors to provide codec specific params,
// but still giving generality for the users.
type AudioEncoderBuilder interface {
	// RTPCodec represents the codec metadata
	RTPCodec() *RTPCodec
	// BuildAudioEncoder builds audio encoder by given media params and audio input
	BuildAudioEncoder(r audio.Reader, p prop.Media) (ReadCloser, error)
}

// VideoEncoderBuilder is the interface that wraps basic operations that are
// necessary to build the video encoder.
//
// This interface is for codec implementors to provide codec specific params,
// but still giving generality for the users.
type VideoEncoderBuilder interface {
	// RTPCodec represents the codec metadata
	RTPCodec() *RTPCodec
	// BuildVideoEncoder builds video encoder by given media params and video input
	BuildVideoEncoder(r video.Reader, p prop.Media) (ReadCloser, error)
}

// ReadCloser is an io.ReadCloser with a controller
type ReadCloser interface {
	Read() (b []byte, release func(), err error)
	Close() error
	Controllable
}

// EncoderController is the interface allowing to control the encoder behaviour after it's initialisation.
// It will possibly have common control method in the future.
// A controller can have optional methods represented by *Controller interfaces
type EncoderController interface{}

// Controllable is a interface representing a encoder which can be controlled
// after it's initialisation with an EncoderController
type Controllable interface {
	Controller() EncoderController
}

// KeyFrameController is a interface representing an encoder that can be forced to produce key frame on demand
type KeyFrameController interface {
	EncoderController
	// ForceKeyFrame forces the next frame to be a keyframe, aka intra-frame.
	ForceKeyFrame() error
}

// BitRateController is a interface representing an encoder which can have a variable bit rate
type BitRateController interface {
	EncoderController
	// SetBitRate sets current target bitrate, lower bitrate means smaller data will be transmitted
	// but this also means that the quality will also be lower.
	SetBitRate(int) error
}

// BaseParams represents an codec's encoding properties
type BaseParams struct {
	// Target bitrate in bps.
	BitRate int

	// Expected interval of the keyframes in frames.
	KeyFrameInterval int
}
