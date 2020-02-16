#include "_cgo_export.h"

void fastNearestNeighbor(
    unsigned char* dst, unsigned const char* src,
    const int ch,
    const int dw, const int dh, const int dstride,
    const int sw, const int sh, const int sstride)
{
  for (int dy = 0; dy < dh; dy++)
  {
    const int sy = dy * sh / dh;
    const unsigned char* src2 = &src[sy * sstride];

    for (int dx = 0; dx < dw; dx++)
    {
      const int sx = ch * (dx * sw / dw);
      for (int c = 0; c < ch; c++)
        *(dst++) = src2[sx + c];
    }
  }
}
