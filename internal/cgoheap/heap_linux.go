//go:build linux && cgo

// Package cgoheap reports live heap usage of the process default malloc
// zone. It is intended for tests that assert on native allocations made by
// cgo libraries such as libvpx.
package cgoheap

/*
#include <malloc.h>

// mallinfo2 is glibc >= 2.33; mallinfo is glibc only. musl provides
// neither, so the probe degrades gracefully there.
#if defined(__GLIBC__) && defined(__GLIBC_PREREQ)
#if __GLIBC_PREREQ(2, 33)
#define PION_HAVE_MALLINFO2 1
#else
#define PION_HAVE_MALLINFO 1
#endif
#endif

static size_t cgoheap_in_use(void) {
#ifdef PION_HAVE_MALLINFO2
	struct mallinfo2 mi = mallinfo2();
	return mi.uordblks;
#elif defined(PION_HAVE_MALLINFO)
	// Approximate on older glibc; may lag behind frees.
	struct mallinfo mi = mallinfo();
	return (size_t)mi.uordblks;
#else
	return 0;
#endif
}

static int cgoheap_supported(void) {
#if defined(PION_HAVE_MALLINFO2) || defined(PION_HAVE_MALLINFO)
	return 1;
#else
	return 0;
#endif
}
*/
import "C"

// HeapInUse returns the number of live bytes in the process default malloc
// arena. It returns 0 when the libc does not expose heap accounting.
func HeapInUse() uint64 {
	return uint64(C.cgoheap_in_use())
}

// Supported reports whether HeapInUse returns meaningful data on this
// platform.
func Supported() bool {
	return C.cgoheap_supported() != 0
}
