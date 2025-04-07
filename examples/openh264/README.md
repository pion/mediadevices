## Instructions

### Install required codecs

In this example, we'll be using openh264 as our video codec. Therefore, we need to make sure that these codecs are installed within our system. 

Installation steps:

* [openh264](https://github.com/pion/mediadevices#openh264)

### Download archive examplee

```
git clone https://github.com/pion/mediadevices.git
```

### Run openh264 example

Run `cd mediadevices/examples/openh264 && go build && ./openh264 recorded.h264`
set bitrate ,first press `Ctrl+c` or send a SIGINT signal.
To stop recording,second press `Ctrl+c` or send a SIGINT signal.

