package io

import "fmt"

// InsufficientBufferError tells the caller that the buffer provided is not sufficient/big
// enough to hold the whole data/sample.
type InsufficientBufferError struct {
	RequiredSize int
}

func (e *InsufficientBufferError) Error() string {
	return fmt.Sprintf("provided buffer doesn't meet the size requirement of length, %d", e.RequiredSize)
}
