package codec

import (
	"io"

	"github.com/pion/mediadevices"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v2"
)

type RTPReader interface {
	ReadRTP() (*rtp.Packet, error)
}

type RTPReadCloser interface {
	RTPReader
	Close()
}

type EncoderBuilder interface {
	Codec() *webrtc.RTPCodec
	BuildEncoder(mediadevices.Track) (RTPReadCloser, error)
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
