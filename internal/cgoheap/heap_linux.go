//go:build linux && cgo

// Package cgoheap reports live heap usage of the process default malloc
// zone. It is intended for tests that assert on native allocations made by
// cgo libraries such as libvpx.
package cgoheap

/*
#include <malloc.h>

// mallinfo2 is glibc >= 2.33 only; musl and older glibc only provide
// mallinfo.
#if defined(__GLIBC__) && defined(__GLIBC_PREREQ)
#if __GLIBC_PREREQ(2, 33)
#define PION_HAVE_MALLINFO2 1
#endif
#endif

static size_t cgoheap_in_use(void) {
#ifdef PION_HAVE_MALLINFO2
	struct mallinfo2 mi = mallinfo2();
	return mi.uordblks;
#else
	// The uordblks field of mallinfo is 0 or approximate on musl, which
	// degrades the measurement but keeps this package buildable on any
	// libc.
	struct mallinfo mi = mallinfo();
	return (size_t)mi.uordblks;
#endif
}
*/
import "C"

// HeapInUse returns the number of live bytes in the process default malloc
// arena.
func HeapInUse() uint64 {
	return uint64(C.cgoheap_in_use())
}
