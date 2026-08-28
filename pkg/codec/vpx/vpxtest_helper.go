//go:build vpxtest

package vpx

/*
#include <stdlib.h>

#ifdef __APPLE__
#include <malloc/malloc.h>
#else
#include <malloc.h>
#endif

// pion_heap_in_use reports the number of live bytes in the process default
// malloc heap. libvpx 1.15+ allocates all internal memory directly from the
// system allocator (vpx_mem_set_functions was removed), so the encoder
// context's internal state is visible here.
size_t pion_heap_in_use(void) {
#ifdef __APPLE__
    malloc_statistics_t st;
    malloc_zone_statistics(malloc_default_zone(), &st);
    return st.size_in_use;
#else
    struct mallinfo2 mi = mallinfo2();
    return mi.uordblks;
#endif
}
*/
import "C"

func heapInUse() uint64 {
	return uint64(C.pion_heap_in_use())
}
