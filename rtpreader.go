package mediadevices

import "github.com/pion/rtp"

type RTPReadCloser interface {
	Read() (pkts []*rtp.Packet, release func(), err error)
	Close() error
}

type rtpReadCloserImpl struct {
	readFn  func() ([]*rtp.Packet, func(), error)
	closeFn func() error
}

func (r *rtpReadCloserImpl) Read() ([]*rtp.Packet, func(), error) {
	return r.readFn()
}

func (r *rtpReadCloserImpl) Close() error {
	return r.closeFn()
}
