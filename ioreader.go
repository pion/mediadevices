package mediadevices

import "github.com/pion/mediadevices/pkg/codec"

type EncodedBuffer struct {
	Data    []byte
	Samples uint32
}

type EncodedReadCloser interface {
	Read() (EncodedBuffer, func(), error)
	Close() error
	codec.Controllable
}

type encodedReadCloserImpl struct {
	readFn       func() (EncodedBuffer, func(), error)
	closeFn      func() error
	controllerFn func() codec.EncoderController
}

func (r *encodedReadCloserImpl) Read() (EncodedBuffer, func(), error) {
	return r.readFn()
}

func (r *encodedReadCloserImpl) Close() error {
	return r.closeFn()
}

func (r *encodedReadCloserImpl) Controller() codec.EncoderController {
	return r.controllerFn()
}

type encodedIOReadCloserImpl struct {
	readFn     func([]byte) (int, error)
	closeFn    func() error
	controller func() codec.EncoderController
}

func newEncodedIOReadCloserImpl(reader EncodedReadCloser) *encodedIOReadCloserImpl {
	var encoded EncodedBuffer
	release := func() {}
	return &encodedIOReadCloserImpl{
		readFn: func(b []byte) (int, error) {
			var err error

			if len(encoded.Data) == 0 {
				release()
				encoded, release, err = reader.Read()
				if err != nil {
					reader.Close()
					return 0, err
				}
			}

			n := copy(b, encoded.Data)
			encoded.Data = encoded.Data[n:]
			return n, nil
		},
		closeFn:    reader.Close,
		controller: reader.Controller,
	}
}

func (r *encodedIOReadCloserImpl) Read(b []byte) (int, error) {
	return r.readFn(b)
}

func (r *encodedIOReadCloserImpl) Close() error {
	return r.closeFn()
}

func (r *encodedIOReadCloserImpl) Controller() codec.EncoderController {
	return r.controller()
}
