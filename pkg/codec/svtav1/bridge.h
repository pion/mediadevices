#include <EbSvtAv1.h>
#include <EbSvtAv1Enc.h>
#include <EbSvtAv1ErrorCodes.h>
#include <stdint.h>

#define ERR_INIT_ENC_HANDLER 1
#define ERR_SET_ENC_PARAM 2
#define ERR_ENC_INIT 3

typedef struct Encoder {
  EbSvtAv1EncConfiguration *param;
  EbComponentType *handle;
} Encoder;

typedef struct Buffer {
  unsigned char *data;
  int len;
} Buffer;

int enc_new(Encoder **e) {
  EbErrorType sret;
  *e = malloc(sizeof(Encoder));
  (*e)->param = malloc(sizeof(EbSvtAv1EncConfiguration));

#if SVT_AV1_CHECK_VERSION(3, 0, 0)
  sret = svt_av1_enc_init_handle(&(*e)->handle, (*e)->param);
#else
  sret = svt_av1_enc_init_handle(&(*e)->handle, NULL, (*e)->param);
#endif
  if (sret != EB_ErrorNone) {
    free((*e)->param);
    free(*e);
    return ERR_INIT_ENC_HANDLER;
  }

  return 0;
}

int enc_init(Encoder *e) {
  EbErrorType sret;

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
  EbErrorType sret = svt_av1_enc_set_parameter(e->handle, e->param);
  if (sret != EB_ErrorNone) {
    return ERR_SET_ENC_PARAM;
  }

  return 0;
}

unsigned char dummy[] = {0, 1, 2, 3};

int enc_encode(Encoder *e, Buffer *out, uint8_t *y, uint8_t *cb, uint8_t *cr) {
  // TODO: implement
  out->data = dummy;
  out->len = 4;

  return 0;
}

int enc_close(Encoder *e) {
  free(e->param);
  free(e);

  return 0;
}

int enc_is_force_keyframe_supported() {
#if SVT_AV1_CHECK_VERSION(1, 8, 0)
  return 1;
#endif
  return 0;
}

