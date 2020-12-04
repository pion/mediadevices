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

void rgbToYCbCrCGO(
    unsigned char* y,
    unsigned char* cb,
    unsigned char* cr,
    const unsigned char r,
    const unsigned char g,
    const unsigned char b)
{
  // ITU-R BT.601
  const int y2 = 77 * r + 150 * g + 29 * b;
  const int cb2 = -43 * r - 85 * g + 128 * b + 0x8000;
  const int cr2 = 128 * r - 107 * g - 21 * b + 0x8000;

  *y = y2 >> 8;
  *cb = cb2 >> 8;
  *cr = cr2 >> 8;
}

void repeatRGBToYCbCrCGO(
    const int n,
    unsigned char* y,
    unsigned char* cb,
    unsigned char* cr,
    const unsigned char r,
    const unsigned char g,
    const unsigned char b)
{
  int i;
  for (i = 0; i < n; ++i)
  {
    rgbToYCbCrCGO(y, cb, cr, r, g, b);
  }
}

void yCbCrToRGBCGO(
    unsigned char* r,
    unsigned char* g,
    unsigned char* b,
    const unsigned char y,
    const unsigned char cb,
    const unsigned char cr)
{
  const int cb2 = cb - 0x80;
  const int cr2 = cr - 0x80;

  // ITU-R BT.601
  int r2 = 256 * y + 359 * cr2;
  int g2 = 256 * y - 88 * cb2 - 183 * cr2;
  int b2 = 256 * y + 454 * cb2;

  if (r2 < 0)
    r2 = 0;
  else if (r2 > 0xFFFF)
    r2 = 0xFFFF;
  if (g2 < 0)
    g2 = 0;
  else if (g2 > 0xFFFF)
    g2 = 0xFFFF;
  if (b2 < 0)
    b2 = 0;
  else if (b2 > 0xFFFF)
    b2 = 0xFFFF;

  *r = r2 >> 8;
  *g = g2 >> 8;
  *b = b2 >> 8;
}

void repeatYCbCrToRGBCGO(
    const int n,
    unsigned char* r,
    unsigned char* g,
    unsigned char* b,
    const unsigned char y,
    const unsigned char cb,
    const unsigned char cr)
{
  int i;
  for (i = 0; i < n; ++i)
  {
    yCbCrToRGBCGO(r, g, b, y, cb, cr);
  }
}

void i444ToRGBACGO(
    unsigned char* rgb,
    const unsigned char* y,
    const unsigned char* cb,
    const unsigned char* cr,
    const int stride, const int h)
{
  int i;
  for (i = 0; i < stride * h; ++i)
  {
    yCbCrToRGBCGO(rgb, rgb + 1, rgb + 2, y[i], cb[i], cr[i]);
    rgb[3] = 0xFF;
    rgb += 4;
  }
}

void rgbaToI444(
    unsigned char* y,
    unsigned char* cb,
    unsigned char* cr,
    const unsigned char* rgb,
    const int stride, const int h)
{
  int i;
  for (i = 0; i < stride * h; ++i)
  {
    rgbToYCbCrCGO(&y[i], &cb[i], &cr[i], rgb[0], rgb[1], rgb[2]);
    rgb += 4;
  }
}
