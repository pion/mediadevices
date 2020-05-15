package io

// Copy copies data from src to dst. If dst is not big enough, return an
// InsufficientBufferError.
func Copy(dst, src []byte) (n int, err error) {
	if len(dst) < len(src) {
		return 0, &InsufficientBufferError{len(src)}
	}

	return copy(dst, src), nil
}
