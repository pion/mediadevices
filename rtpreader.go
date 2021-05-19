package mediadevices

import (
	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/rtp"
)

type RTPReadCloser interface {
	Read() (pkts []*rtp.Packet, release func(), err error)
	Close() error
	codec.Controllable
}

type rtpReadCloserImpl struct {
	readFn       func() ([]*rtp.Packet, func(), error)
	closeFn      func() error
	controllerFn func() codec.EncoderController
}

func (r *rtpReadCloserImpl) Read() ([]*rtp.Packet, func(), error) {
	return r.readFn()
}

func (r *rtpReadCloserImpl) Close() error {
	return r.closeFn()
}

func (r *rtpReadCloserImpl) Controller() codec.EncoderController {
	return r.controllerFn()
}
