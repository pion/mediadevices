//go:build darwin && cgo

// Package cgoheap reports live heap usage of the process default malloc
// zone. It is intended for tests that assert on native allocations made by
// cgo libraries such as libvpx.
package cgoheap

/*
#include <malloc/malloc.h>

static size_t cgoheap_in_use(void) {
	malloc_statistics_t st;
	malloc_zone_statistics(malloc_default_zone(), &st);
	return st.size_in_use;
}
*/
import "C"

// HeapInUse returns the number of live bytes in the process default malloc
// zone.
func HeapInUse() uint64 {
	return uint64(C.cgoheap_in_use())
}

// Supported reports whether HeapInUse returns meaningful data on this
// platform.
func Supported() bool {
	return true
}
