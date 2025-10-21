#include <EbSvtAv1.h>
#include <EbSvtAv1Enc.h>
#include <EbSvtAv1ErrorCodes.h>
#include <stdbool.h>
#include <stdint.h>
#include <string.h>

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

int enc_free(Encoder *e) {
  free(e->in_buf->p_buffer);
  free(e->in_buf);
  free(e->param);
  free(e);

  return 0;
}

int enc_new(Encoder **e) {
  *e = malloc(sizeof(Encoder));
  (*e)->param = malloc(sizeof(EbSvtAv1EncConfiguration));
  (*e)->in_buf = malloc(sizeof(EbBufferHeaderType));

  memset((*e)->in_buf, 0, sizeof(EbBufferHeaderType));
  (*e)->in_buf->p_buffer = malloc(sizeof(EbSvtIOFormat));
  (*e)->in_buf->size = sizeof(EbBufferHeaderType);

#if SVT_AV1_CHECK_VERSION(3, 0, 0)
  const EbErrorType sret = svt_av1_enc_init_handle(&(*e)->handle, (*e)->param);
#else
  const EbErrorType sret = svt_av1_enc_init_handle(&(*e)->handle, NULL, (*e)->param);
#endif
  if (sret != EB_ErrorNone) {
    enc_free(*e);
    return ERR_INIT_ENC_HANDLER;
  }

  return 0;
}

int enc_init(Encoder *e) {
  EbErrorType sret;

  e->param->encoder_bit_depth = 8;
  e->param->encoder_color_format = EB_YUV420;

  sret = svt_av1_enc_set_parameter(e->handle, e->param);
  if (sret != EB_ErrorNone) {
    return ERR_SET_ENC_PARAM;
  }

  sret = svt_av1_enc_init(e->handle);
  if (sret != EB_ErrorNone) {
    return ERR_ENC_INIT;
  }

  return 0;
}

int enc_apply_param(Encoder *e) {
  const EbErrorType sret = svt_av1_enc_set_parameter(e->handle, e->param);
  if (sret != EB_ErrorNone) {
    return ERR_SET_ENC_PARAM;
  }

  return 0;
}

int enc_force_keyframe(Encoder *e) {
  e->force_keyframe = true;
  return 0;
}

int enc_send_frame(Encoder *e, uint8_t *y, uint8_t *cb, uint8_t *cr, int ystride, int cstride) {
  EbSvtIOFormat *in_data = (EbSvtIOFormat *)e->in_buf->p_buffer;
  in_data->luma = y;
  in_data->cb = cb;
  in_data->cr = cr;
  in_data->y_stride = ystride;
  in_data->cb_stride = cstride;
  in_data->cr_stride = cstride;

  e->in_buf->pic_type = EB_AV1_INVALID_PICTURE; // auto
  if (e->force_keyframe) {
    e->in_buf->pic_type = EB_AV1_KEY_PICTURE;
    e->force_keyframe = false;
  }
  e->in_buf->flags = 0;
  e->in_buf->pts++;
  e->in_buf->n_filled_len = ystride * e->param->source_height;
  e->in_buf->n_filled_len += 2 * cstride * e->param->source_height / 2;

  const EbErrorType sret = svt_av1_enc_send_picture(e->handle, e->in_buf);
  if (sret != EB_ErrorNone) {
    return ERR_SEND_PICTURE;
  }
  return 0;
}

int enc_get_packet(Encoder *e, EbBufferHeaderType **out) {
  const EbErrorType sret = svt_av1_enc_get_packet(e->handle, out, 1);
  if (sret == EB_NoErrorEmptyQueue) {
    return 0;
  }
  if (sret != EB_ErrorNone) {
    return ERR_GET_PACKET;
  }

  return 0;
}
