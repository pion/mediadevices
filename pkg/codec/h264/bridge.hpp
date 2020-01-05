#pragma once

#include <openh264/codec_api.h>

#ifdef __cplusplus
extern "C" {
#endif
typedef struct Slice {
  unsigned char *data;
  int data_len;
} Slice;

typedef struct Frame {
  void *y, *u, *v;
  int height;
  int width;
} Frame;

typedef struct EncoderOptions {
  int width, height;
  int target_bitrate, max_bitrate;
  float max_fps;
} EncoderOptions;

typedef struct Encoder {
  SEncParamExt params;
  ISVCEncoder *engine;
  unsigned char *buff;
  int buff_size;
} Encoder;

Encoder *enc_new(const EncoderOptions params);
void enc_free(Encoder *e);
Slice enc_encode(Encoder *e, Frame f);
#ifdef __cplusplus
}
#endif