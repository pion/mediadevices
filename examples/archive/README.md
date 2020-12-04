## Instructions

### Install required codecs

In this example, we'll be using x264 as our video codec. Therefore, we need to make sure that these codecs are installed within our system. 

Installation steps:

* [x264](https://github.com/pion/mediadevices#x264)

### Download archive examplee

```
git clone https://github.com/pion/mediadevices.git
```

### Run archive example

Run `cd mediadevices/examples/archive && go build && ./archive recorded.h264`

To stop recording, press `Ctrl+c` or send a SIGINT signal.

### Playback recorded video

Install GStreamer and run:
```
gst-launch-1.0 playbin uri=file://${PWD}/recorded.h264
```

Or run VLC media plyer:
```
vlc recorded.h264
```

A video should start playing in your GStreamer or VLC window.

Congrats, you have used pion-MediaDevices! Now start building something cool

