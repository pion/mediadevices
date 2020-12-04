#include "bridge.hpp"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/time.h>

Encoder *enc_new(const EncoderOptions opts, int *eresult) {
  int rv;
  ISVCEncoder *engine;
  SEncParamExt params;

  rv = WelsCreateSVCEncoder(&engine);
  if (rv != 0) {
    *eresult = rv;
    return NULL;
  }

  rv = engine->GetDefaultParams(&params);
  if (rv != 0) {
    *eresult = rv;
    return NULL;
  }

  // TODO: Remove hardcoded values
  params.iUsageType = CAMERA_VIDEO_REAL_TIME;
  params.iPicWidth = opts.width;
  params.iPicHeight = opts.height;
  params.iTargetBitrate = opts.target_bitrate;
  params.iMaxBitrate = opts.target_bitrate;
  params.iRCMode = RC_BITRATE_MODE;
  params.fMaxFrameRate = opts.max_fps;
  params.bEnableFrameSkip = true;
  params.uiMaxNalSize = 0;
  params.uiIntraPeriod = 30;
  // set to 0, so that it'll automatically use multi threads when needed
  params.iMultipleThreadIdc = 0;
  // The base spatial layer 0 is the only one we use.
  params.sSpatialLayers[0].iVideoWidth = params.iPicWidth;
  params.sSpatialLayers[0].iVideoHeight = params.iPicHeight;
  params.sSpatialLayers[0].fFrameRate = params.fMaxFrameRate;
  params.sSpatialLayers[0].iSpatialBitrate = params.iTargetBitrate;
  params.sSpatialLayers[0].iMaxSpatialBitrate = params.iTargetBitrate;
  // Single NAL unit mode
  params.sSpatialLayers[0].sSliceArgument.uiSliceNum = 1;
  params.sSpatialLayers[0].sSliceArgument.uiSliceMode = SM_SIZELIMITED_SLICE;
  params.sSpatialLayers[0].sSliceArgument.uiSliceSizeConstraint = 12800;

  rv = engine->InitializeExt(&params);
  if (rv != 0) {
    *eresult = rv;
    return NULL;
  }

  Encoder *encoder = (Encoder *)malloc(sizeof(Encoder));
  encoder->engine = engine;
  encoder->params = params;
  encoder->buff = (unsigned char *)malloc(opts.width * opts.height);
  encoder->buff_size = opts.width * opts.height;
  return encoder;
}

void enc_free(Encoder *e, int *eresult) {
  int rv = e->engine->Uninitialize();
  if (rv != 0) {
    *eresult = rv;
    return;
  }

  WelsDestroySVCEncoder(e->engine);

  free(e->buff);
  free(e);
}

// There's a good reference from ffmpeg in using the encode_frame
// Reference: https://ffmpeg.org/doxygen/2.6/libopenh264enc_8c_source.html
Slice enc_encode(Encoder *e, Frame f, int *eresult) {
  int rv;
  SSourcePicture pic = {0};
  SFrameBSInfo info = {0};
  Slice payload = {0};

  pic.iPicWidth = f.width;
  pic.iPicHeight = f.height;
  pic.iColorFormat = videoFormatI420;
  // We always received I420 format
  pic.iStride[0] = pic.iPicWidth;
  pic.iStride[1] = pic.iStride[2] = pic.iPicWidth / 2;
  pic.pData[0] = (unsigned char *)f.y;
  pic.pData[1] = (unsigned char *)f.u;
  pic.pData[2] = (unsigned char *)f.v;

  rv = e->engine->EncodeFrame(&pic, &info);
  if (rv != 0) {
    *eresult = rv;
    return payload;
  }

  int *layer_size = (int *)calloc(sizeof(int), info.iLayerNum);
  int size = 0;
  for (int layer = 0; layer < info.iLayerNum; layer++) {
    for (int i = 0; i < info.sLayerInfo[layer].iNalCount; i++)
      layer_size[layer] += info.sLayerInfo[layer].pNalLengthInByte[i];

    size += layer_size[layer];
  }

  if (e->buff_size < size) {
    e->buff = (unsigned char *)malloc(size);
    e->buff_size = size;
  }
  size = 0;
  for (int layer = 0; layer < info.iLayerNum; layer++) {
    memcpy(e->buff + size, info.sLayerInfo[layer].pBsBuf, layer_size[layer]);
    size += layer_size[layer];
  }
  free(layer_size);

  payload.data = e->buff;
  payload.data_len = size;
  return payload;
}
