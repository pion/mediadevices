package mediadevices

type encodedReadCloserImpl struct {
	readFn  func([]byte) (int, error)
	closeFn func() error
}

func (r *encodedReadCloserImpl) Read(b []byte) (int, error) {
	return r.readFn(b)
}

func (r *encodedReadCloserImpl) Close() error {
	return r.closeFn()
}
