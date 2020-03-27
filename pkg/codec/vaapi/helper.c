#ifdef HAS_VAAPI

#include <fcntl.h>
#include <string.h>
#include <stdlib.h>
#include <unistd.h>
#include <va/va.h>
#include <va/va_drm.h>

#include "helper.h"

int open2(char *path, int flags)
{
  return open(path, flags);
}

VAGenericValue genValInt(const int i)
{
  VAGenericValue val;
  val.type = VAGenericValueTypeInteger;
  val.value.i = i;
  return val;
}

VAStatus vaCreateBufferPtr(
    VADisplay d, VAContextID ctx, VABufferType type,
    unsigned int size, unsigned int n,
    size_t dataptr,
    VABufferID *buf_id)
{
  return vaCreateBuffer(d, ctx, type, size, n, (void *)dataptr, buf_id);
}

VAStatus vaMapBufferSeg(VADisplay d, VABufferID buf_id, VACodedBufferSegment **seg)
{
  return vaMapBuffer(d, buf_id, (void **)seg);
}

void copyI420toNV12(
    void *nv12,
    const uint8_t *y, const uint8_t *cb, const uint8_t *cr,
    const unsigned int n)
{
  unsigned int i, j;
  memcpy(nv12, y, n);
  uint8_t *p = &((uint8_t *)nv12)[n];
  for (i = 0; i < n / 4; i++)
  {
    p[i * 2] = cb[i];
    p[i * 2 + 1] = cr[i];
  }
}

#endif // HAS_VAAPI
