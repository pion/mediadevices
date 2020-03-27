#pragma once

#ifdef HAS_VAAPI

#include <unistd.h>
#include <va/va.h>

int open2(char *path, int flags);
VAGenericValue genValInt(const int i);
VAStatus vaCreateBufferPtr(
    VADisplay d, VAContextID ctx, VABufferType type,
    unsigned int size, unsigned int n,
    size_t dataptr,
    VABufferID *buf_id);
VAStatus vaMapBufferSeg(VADisplay d, VABufferID buf_id, VACodedBufferSegment **seg);
void copyI420toNV12(
    void *nv12,
    const uint8_t *y, const uint8_t *cb, const uint8_t *cr,
    const unsigned int n);

#endif // HAS_VAAPI
