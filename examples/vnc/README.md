## Instructions

### Install required codecs

In this example, we'll be using x264 and opus as our video and audio codecs. Therefore, we need to make sure that these codecs are installed within our system. 

Installation steps:

* [x264](https://github.com/pion/mediadevices#x264)

### Download vnc example

```
git clone https://github.com/pion/mediadevices.git
```

#### Compile vnc example

```
cd mediadevices/examples/vnc && go build
```

### Open example page

[jsfiddle.net](https://jsfiddle.net/gh/get/library/pure/pion/mediadevices/tree/master/examples/internal/jsfiddle/audio-and-video) you should see two text-areas and a 'Start Session' button

### Run the webrtc example with your browsers SessionDescription as stdin

In the jsfiddle the top textarea is your browser, copy that, and store the session description in an environment variable, `export SDP=<put_the_sdp_here>`

Run `echo $SDP | ./vnc`

In Windows

```powershell
type sdp.txt| .\vnc.exe
```
### Input webrtc's SessionDescription into your browser

Copy the text that `./webrtc` just emitted and copy into second text area

### Hit 'Start Session' in jsfiddle, enjoy your video!

A video should start playing in your browser above the input boxes, and will continue playing until you close the application.

Congrats, you have used pion-MediaDevices! Now start building something cool
