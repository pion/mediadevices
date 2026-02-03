#pragma once

#include <EbSvtAv1.h>
#include <EbSvtAv1Enc.h>
#include <stdbool.h>

// For SVT-AV1 v4.0.0+ support
// See below for documentation of low delay value:
// https://gitlab.com/AOMediaCodec/SVT-AV1/-/blob/master/Docs/Parameters.md#gop-size-and-type-options
#ifndef SVT_AV1_PRED_LOW_DELAY
#define SVT_AV1_PRED_LOW_DELAY 1
#endif

#define ERR_INIT_ENC_HANDLER 1
#define ERR_SET_ENC_PARAM 2
#define ERR_ENC_INIT 3
#define ERR_SEND_PICTURE 4
#define ERR_GET_PACKET 5

typedef struct Encoder {
  EbSvtAv1EncConfiguration *param;
  EbComponentType *handle;
  EbBufferHeaderType *in_buf;

  bool force_keyframe;
} Encoder;

int enc_free(Encoder *e);
int enc_new(Encoder **e);
int enc_init(Encoder *e);
int enc_apply_param(Encoder *e);
int enc_force_keyframe(Encoder *e);
int enc_send_frame(Encoder *e, uint8_t *y, uint8_t *cb, uint8_t *cr, int ystride, int cstride);
int enc_get_packet(Encoder *e, EbBufferHeaderType **out);
void memcpy_uint8(uint8_t *dst, const uint8_t *src, size_t n);
