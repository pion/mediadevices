#include <interface/mmal/mmal.h>
#include <interface/mmal/util/mmal_default_components.h>
#include <interface/mmal/util/mmal_util_params.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdlib.h>

#define CHK(__status, __msg)                                                                                           \
  do {                                                                                                                 \
    status.code = __status;                                                                                            \
    if (status.code != MMAL_SUCCESS) {                                                                                 \
      status.msg = __msg;                                                                                              \
      goto CleanUp;                                                                                                    \
    }                                                                                                                  \
  } while (0)

typedef struct Status {
  MMAL_STATUS_T code;
  const char *msg;
} Status;

typedef struct Slice {
  uint8_t *data;
  int len;
} Slice;

typedef struct Params {
  int width, height;
  uint32_t bitrate;
  uint32_t key_frame_interval;
} Params;

typedef struct Encoder {
  MMAL_COMPONENT_T *component;
  MMAL_PORT_T *port_in, *port_out;
  MMAL_QUEUE_T *queue_out;
  MMAL_POOL_T *pool_in, *pool_out;
} Encoder;

Status enc_new(Params, Encoder *);
Status enc_encode(Encoder *, Slice y, Slice cb, Slice cr, MMAL_BUFFER_HEADER_T **);
Status enc_close(Encoder *);

static void encoder_in_cb(MMAL_PORT_T *port, MMAL_BUFFER_HEADER_T *buffer) { mmal_buffer_header_release(buffer); }

static void encoder_out_cb(MMAL_PORT_T *port, MMAL_BUFFER_HEADER_T *buffer) {
  MMAL_QUEUE_T *queue = (MMAL_QUEUE_T *)port->userdata;
  mmal_queue_put(queue, buffer);
}

Status enc_new(Params params, Encoder *encoder) {
  Status status = {0};
  bool created = false;

  memset(encoder, 0, sizeof(Encoder));

  CHK(mmal_component_create(MMAL_COMPONENT_DEFAULT_VIDEO_ENCODER, &encoder->component),
      "Failed to create video encoder component");
  created = true;

  encoder->port_in = encoder->component->input[0];
  encoder->port_in->format->type = MMAL_ES_TYPE_VIDEO;
  encoder->port_in->format->encoding = MMAL_ENCODING_I420;
  encoder->port_in->format->es->video.width = params.width;
  encoder->port_in->format->es->video.height = params.height;
  encoder->port_in->format->es->video.par.num = 1;
  encoder->port_in->format->es->video.par.den = 1;
  encoder->port_in->format->es->video.crop.x = 0;
  encoder->port_in->format->es->video.crop.y = 0;
  encoder->port_in->format->es->video.crop.width = params.width;
  encoder->port_in->format->es->video.crop.height = params.height;
  CHK(mmal_port_format_commit(encoder->port_in), "Failed to commit input port format");

  encoder->port_out = encoder->component->output[0];
  encoder->port_out->format->type = MMAL_ES_TYPE_VIDEO;
  encoder->port_out->format->encoding = MMAL_ENCODING_H264;
  encoder->port_out->format->bitrate = params.bitrate;
  CHK(mmal_port_format_commit(encoder->port_out), "Failed to commit output port format");

  MMAL_PARAMETER_VIDEO_PROFILE_T encoder_param_profile = {0};
  encoder_param_profile.hdr.id = MMAL_PARAMETER_PROFILE;
  encoder_param_profile.hdr.size = sizeof(encoder_param_profile);
  encoder_param_profile.profile[0].profile = MMAL_VIDEO_PROFILE_H264_BASELINE;
  encoder_param_profile.profile[0].level = MMAL_VIDEO_LEVEL_H264_42;
  CHK(mmal_port_parameter_set(encoder->port_out, &encoder_param_profile.hdr), "Failed to set encoder profile param");

  CHK(mmal_port_parameter_set_uint32(encoder->port_out, MMAL_PARAMETER_INTRAPERIOD, params.key_frame_interval),
      "Failed to set intra period param");

  MMAL_PARAMETER_VIDEO_RATECONTROL_T encoder_param_rate_control = {0};
  encoder_param_rate_control.hdr.id = MMAL_PARAMETER_RATECONTROL;
  encoder_param_rate_control.hdr.size = sizeof(encoder_param_rate_control);
  encoder_param_rate_control.control = MMAL_VIDEO_RATECONTROL_VARIABLE;
  CHK(mmal_port_parameter_set(encoder->port_out, &encoder_param_rate_control.hdr), "Failed to set rate control param");

  // Some decoders expect SPS/PPS headers to be added to every frame
  CHK(mmal_port_parameter_set_boolean(encoder->port_out, MMAL_PARAMETER_VIDEO_ENCODE_INLINE_HEADER, MMAL_TRUE),
      "Failed to set inline header param");

  CHK(mmal_port_parameter_set_boolean(encoder->port_out, MMAL_PARAMETER_VIDEO_ENCODE_HEADERS_WITH_FRAME, MMAL_TRUE),
      "Failed to set headers with frame param");

  /* FIXME: Somehow this flag is broken? When this flag is on, the encoder will get stuck.
  // Since our use case is mainly for real time streaming, the encoder should optimized for low latency
  CHK(mmal_port_parameter_set_boolean(encoder->port_out, MMAL_PARAMETER_VIDEO_ENCODE_H264_LOW_LATENCY, MMAL_TRUE),
      "Failed to set low latency param");
  */

  // Now we know the format of both ports and the requirements of the encoder, we can create
  // our buffer headers and their associated memory buffers. We use the buffer pool API for this.
  encoder->port_in->buffer_num = encoder->port_in->buffer_num_min;
  // mmal calculates recommended size that's big enough to store all of the pixels
  encoder->port_in->buffer_size = encoder->port_in->buffer_size_recommended;
  encoder->pool_in = mmal_pool_create(encoder->port_in->buffer_num, encoder->port_in->buffer_size);
  encoder->port_out->buffer_num = encoder->port_out->buffer_num_min;
  encoder->port_out->buffer_size = encoder->port_out->buffer_size_recommended;
  encoder->pool_out = mmal_pool_create(encoder->port_out->buffer_num, encoder->port_out->buffer_size);

  // Create a queue to store our encoded video frames. The callback we will get when
  // a frame has been encoded will put the frame into this queue.
  encoder->queue_out = mmal_queue_create();
  encoder->port_out->userdata = (void *)encoder->queue_out;

  // Enable all the input port and the output port.
  // The callback specified here is the function which will be called when the buffer header
  // we sent to the component has been processed.
  CHK(mmal_port_enable(encoder->port_in, encoder_in_cb), "Failed to enable input port");
  CHK(mmal_port_enable(encoder->port_out, encoder_out_cb), "Failed to enable output port");

  // Enable the component. Components will only process data when they are enabled.
  CHK(mmal_component_enable(encoder->component), "Failed to enable component");

CleanUp:

  if (status.code != MMAL_SUCCESS) {
    if (created) {
      enc_close(encoder);
    }
  }

  return status;
}

// enc_encode encodes y, cb, cr. The encoded frame is going to be stored in encoded_buffer.
// IMPORTANT: the caller is responsible to release the ownership of encoded_buffer
Status enc_encode(Encoder *encoder, Slice y, Slice cb, Slice cr, MMAL_BUFFER_HEADER_T **encoded_buffer) {
  Status status = {0};
  MMAL_BUFFER_HEADER_T *buffer;
  uint32_t required_size;

  // buffer should always be available since the encoding process is blocking
  buffer = mmal_queue_get(encoder->pool_in->queue);
  assert(buffer != NULL);
  // buffer->data should've been allocated with enough memory to contain a frame by pool_in
  required_size = y.len + cb.len + cr.len;
  assert(buffer->alloc_size >= required_size);
  memcpy(buffer->data, y.data, y.len);
  memcpy(buffer->data + y.len, cb.data, cb.len);
  memcpy(buffer->data + y.len + cb.len, cr.data, cr.len);
  buffer->length = required_size;
  CHK(mmal_port_send_buffer(encoder->port_in, buffer), "Failed to send filled buffer to input port");

  while (1) {
    // Send empty buffers to the output port to allow the encoder to start
    // producing frames as soon as it gets input data
    while ((buffer = mmal_queue_get(encoder->pool_out->queue)) != NULL) {
      CHK(mmal_port_send_buffer(encoder->port_out, buffer), "Failed to send empty buffers to output port");
    }

    while ((buffer = mmal_queue_wait(encoder->queue_out)) != NULL) {
      if ((buffer->flags & MMAL_BUFFER_HEADER_FLAG_FRAME_END) != 0) {
        *encoded_buffer = buffer;
        goto CleanUp;
      }

      mmal_buffer_header_release(buffer);
    }
  }

CleanUp:

  return status;
}

Status enc_close(Encoder *encoder) {
  Status status = {0};

  mmal_pool_destroy(encoder->pool_out);
  mmal_pool_destroy(encoder->pool_in);
  mmal_queue_destroy(encoder->queue_out);
  mmal_component_destroy(encoder->component);

CleanUp:

  return status;
}
