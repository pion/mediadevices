package io

import (
	"log"
	"testing"
)

func TestCopy(t *testing.T) {
	var dst []byte
	src := make([]byte, 4)

	n, err := Copy(dst, src)
	if err == nil {
		t.Fatal("expected err to be non-nill")
	}

	if n != 0 {
		t.Fatalf("expected n to be 0, but got %d", n)
	}

	e, ok := err.(*InsufficientBufferError)
	if !ok {
		t.Fatalf("expected error to be InsufficientBufferError")
	}

	if e.RequiredSize != len(src) {
		t.Fatalf("expected required size to be %d, but got %d", len(src), e.RequiredSize)
	}

	dst = make([]byte, 2*e.RequiredSize)
	n, err = Copy(dst, src)
	if err != nil {
		t.Fatalf("expected to not get an error after expanding the buffer")
	}

	if n != len(src) {
		t.Fatalf("expected n to be %d, but got %d", len(src), n)
	}

	for i := 0; i < len(src); i++ {
		if src[i] != dst[i] {
			log.Fatalf("expected value at %d to be %d, but got %d", i, src[i], dst[i])
		}
	}
}
