#include <stdint.h>

void decodeYUY2CGO(
    uint8_t* y,
    uint8_t* cb,
    uint8_t* cr,
    uint8_t* yuy2,
    int width, int height)
{
  const int l = width * height * 2;
  int i, fast = 0, slow = 0;
  for (i = 0; i < l; i += 4)
  {
    y[fast] = yuy2[i];
    cb[slow] = yuy2[i + 1];
    y[fast + 1] = yuy2[i + 2];
    cr[slow] = yuy2[i + 3];
    fast += 2;
    ++slow;
  }
}

void decodeUYVYCGO(
    uint8_t* y,
    uint8_t* cb,
    uint8_t* cr,
    uint8_t* uyvy,
    int width, int height)
{
  const int l = width * height * 2;
  int i, fast = 0, slow = 0;
  for (i = 0; i < l; i += 4)
  {
    cb[slow] = uyvy[i];
    y[fast] = uyvy[i+1];
    cr[slow] = uyvy[i + 2];
    y[fast + 1] = uyvy[i + 3];
    fast += 2;
    ++slow;
  }
}
