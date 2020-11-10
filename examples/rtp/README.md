## Instructions

### Install required codecs

In this example, we'll be using x264 as our video codec. Therefore, we need to make sure that these codecs are installed within our system. 

Installation steps:

* [x264](https://github.com/pion/mediadevices#x264)

### Download rtp example

```
git clone https://github.com/pion/mediadevices.git
```

### Listen RTP

Install GStreamer and run:
```
gst-launch-1.0 udpsrc port=5000 caps=application/x-rtp,encode-name=H264 \
    ! rtph264depay ! avdec_h264 ! videoconvert ! autovideosink
```

Or run VLC media plyer:
```
vlc ./h264.sdp
```

### Run rtp

Run `cd mediadevices/examples/archive && go build && ./rtp localhost:5000`

A video should start playing in your GStreamer or VLC window.
It's not WebRTC, but pure RTP.

Congrats, you have used pion-MediaDevices! Now start building something cool

