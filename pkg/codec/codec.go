package codec

import (
	"io"

	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

// AudioEncoderBuilder is the interface that wraps basic operations that are
// necessary to build the audio encoder.
//
// This interface is for codec implementors to provide codec specific params,
// but still giving generality for the users.
type AudioEncoderBuilder interface {
	// Name represents the codec name
	Name() string
	// BuildAudioEncoder builds audio encoder by given media params and audio input
	BuildAudioEncoder(r audio.Reader, p prop.Media) (ReadCloser, error)
}

// VideoEncoderBuilder is the interface that wraps basic operations that are
// necessary to build the video encoder.
//
// This interface is for codec implementors to provide codec specific params,
// but still giving generality for the users.
type VideoEncoderBuilder interface {
	// Name represents the codec name
	Name() string
	// BuildVideoEncoder builds video encoder by given media params and video input
	BuildVideoEncoder(r video.Reader, p prop.Media) (ReadCloser, error)
}

// ReadCloser is an io.ReadCloser with methods for rate limiting: SetBitRate and ForceKeyFrame
type ReadCloser interface {
	io.ReadCloser
	// SetBitRate sets current target bitrate, lower bitrate means smaller data will be transmitted
	// but this also means that the quality will also be lower.
	SetBitRate(int) error
	// ForceKeyFrame forces the next frame to be a keyframe, aka intra-frame.
	ForceKeyFrame() error
}

// BaseParams represents an codec's encoding properties
type BaseParams struct {
	// Target bitrate in bps.
	BitRate int

	// Expected interval of the keyframes in frames.
	KeyFrameInterval int
}
