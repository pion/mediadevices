#include <stdint.h>

#include "_cgo_export.h"

void i444ToI420CGO(
    unsigned char* cb,
    unsigned char* cr,
    const int stride, const int h)
{
  int isrc0 = 0;
  int isrc1 = stride;
  int idst = 0;
  for (int y = 0; y < h / 2; y++)
  {
    for (int x = 0; x < stride / 2; x++)
    {
      const uint8_t cb2 =
          ((uint16_t)cb[isrc0] + (uint16_t)cb[isrc1] +
           (uint16_t)cb[isrc0 + 1] + (uint16_t)cb[isrc1 + 1]) /
          4;
      const uint8_t cr2 =
          ((uint16_t)cr[isrc0] + (uint16_t)cr[isrc1] +
           (uint16_t)cr[isrc0 + 1] + (uint16_t)cr[isrc1 + 1]) /
          4;
      cb[idst] = cb2;
      cr[idst] = cr2;
      isrc0 += 2;
      isrc1 += 2;
      idst++;
    }
    isrc0 += stride;
    isrc1 += stride;
  }
}

void i422ToI420CGO(
    unsigned char* cb,
    unsigned char* cr,
    const int stride, const int h)
{
  int isrc = 0;
  int idst = 0;
  for (int y = 0; y < h / 2; y++)
  {
    for (int x = 0; x < stride; x++)
    {
      const uint8_t cb2 = ((uint16_t)cb[isrc] + (uint16_t)cb[isrc + stride]) / 2;
      const uint8_t cr2 = ((uint16_t)cr[isrc] + (uint16_t)cr[isrc + stride]) / 2;
      cb[idst] = cb2;
      cr[idst] = cr2;
      isrc++;
      idst++;
    }
    isrc += stride;
  }
}
