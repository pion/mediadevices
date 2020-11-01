package mediadevices

import "github.com/pion/rtp"

type RTPReader interface {
	Read() (pkts []*rtp.Packet, release func(), err error)
	Close() error
}

type rtpReaderImpl struct {
	readFn  func() ([]*rtp.Packet, func(), error)
	closeFn func() error
}

func (r *rtpReaderImpl) Read() ([]*rtp.Packet, func(), error) {
	return r.readFn()
}

func (r *rtpReaderImpl) Close() error {
	return r.closeFn()
}
