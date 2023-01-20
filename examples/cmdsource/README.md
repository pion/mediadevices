## Instructions

This example is nearly the same as the archive example, but uses the output of a shell command (in this case ffmpeg) as a video input instead of a camera. See the other examples for how to take this track and use or stream it.

### Install required codecs

In this example, we'll be using x264 as our video codec. We also use FFMPEG to generate the test video stream. (Note that cmdsource does not requre ffmpeg to run, it is just what this example uses) Therefore, we need to make sure that these are installed within our system.

Installation steps:

* [ffmpeg](https://ffmpeg.org/)
* [x264](https://github.com/pion/mediadevices#x264)

### Download cmdsource example

```
git clone https://github.com/pion/mediadevices.git
```

### Run cmdsource example

Run `cd mediadevices/examples/cmdsource && go build && ./cmdsource recorded.h264`

To stop recording, press `Ctrl+c` or send a SIGINT signal.

### Playback recorded video

Use ffplay (part of the ffmpeg project):
```
ffplay -f h264 recorded.h264
```

Or install GStreamer and run:
```
gst-launch-1.0 playbin uri=file://${PWD}/recorded.h264
```

Or run VLC media plyer:
```
vlc recorded.h264
```

A video should start playing in a window.

Congrats, you have used pion-MediaDevices! Now start building something cool
