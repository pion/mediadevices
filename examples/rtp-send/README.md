## Instructions

### Download rtp-send example

```
go get github.com/pion/mediadevices/examples/rtp-send
```

### Listen RTP

Install GStreamer and run:
```
gst-launch-1.0 udpsrc port=5000 caps=application/x-rtp,encode-name=VP8 \
    ! rtpvp8depay ! vp8dec ! videoconvert ! autovideosink
```

Or run VLC media plyer:
```
vlc ./vp8.sdp
```

### Run rtp-send

Run `rtp-send localhost:5000`

A video should start playing in your GStreamer or VLC window.
It's not WebRTC, but pure RTP.

Congrats, you have used pion-MediaDevices! Now start building something cool
