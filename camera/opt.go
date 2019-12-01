package camera

import (
	"github.com/pion/webrtc/v2"
)

type Options struct {
	PC            *webrtc.PeerConnection
	Codec         string
	Width, Height int
}
