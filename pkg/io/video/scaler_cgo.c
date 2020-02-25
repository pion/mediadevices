#include <stdint.h>
#include <string.h>
#include "_cgo_export.h"

void fastNearestNeighbor(
    uint8_t* dst, const uint8_t* src,
    const int ch,
    const int dw, const int dh, const int dstride,
    const int sw, const int sh, const int sstride)
{
  for (int dy = 0; dy < dh; dy++)
  {
    const int sy = dy * sh / dh;
    const uint8_t* src2 = &src[sy * sstride];

    for (int dx = 0; dx < dw; dx++)
    {
      const int sx = ch * (dx * sw / dw);
      for (int c = 0; c < ch; c++)
        *(dst++) = src2[sx + c];
    }
  }
}

void fastBoxSampling(
    uint8_t* dst, const uint8_t* src,
    const int ch,
    const int dw, const int dh, const int dstride,
    const int sw, const int sh, const int sstride,
    uint32_t* tmp)
{
  memset(tmp, 0, dw * dh * ch);

  for (int sy = 0; sy < sh; sy++)
  {
    const uint8_t* src2 = &src[sy * sstride];
    int tx = 0;
    const int ty = sy * dh / sh;
    uint32_t* tmp2 = &tmp[ty * dstride];
    for (int sx = 0; sx < sw * ch; sx += ch)
    {
      if (tx * sw < sx * dw)
        tx += ch;

      for (int c = 0; c < ch; c++)
      {
        tmp2[tx + c] += 0x10000 | src2[sx + c];
      }
    }
  }

  for (int i = 0; i < dw * dh * ch; i++)
  {
    const uint32_t tmp2 = tmp[i];
    const uint16_t sum = tmp2 & 0xFFFF;
    const uint16_t num = tmp2 >> 16;
    if (num > 0)
      dst[i] = sum / num;
  }
}
