//go:build linux && cgo

// Package cgoheap reports live heap usage of the process default malloc
// zone. It is intended for tests that assert on native allocations made by
// cgo libraries such as libvpx.
package cgoheap

/*
#include <malloc.h>

static size_t cgoheap_in_use(void) {
	struct mallinfo2 mi = mallinfo2();
	return mi.uordblks;
}
*/
import "C"

// HeapInUse returns the number of live bytes in the process default malloc
// arena.
func HeapInUse() uint64 {
	return uint64(C.cgoheap_in_use())
}
