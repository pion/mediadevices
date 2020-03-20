## Instructions

### Download screenshare

```
go get github.com/pion/mediadevices/examples/screenshare
```

### Open example page

[jsfiddle.net](https://jsfiddle.net/gh/get/library/pure/pion/mediadevices/tree/master/examples/internal/jsfiddle/audio-and-video) you should see two text-areas and a 'Start Session' button

### Run screenshare with your browsers SessionDescription as stdin

In the jsfiddle the top textarea is your browser, copy that and:

#### Linux

Run `echo $BROWSER_SDP | screenshare`

### Input screenshare's SessionDescription into your browser

Copy the text that `screenshare` just emitted and copy into second text area

### Hit 'Start Session' in jsfiddle, enjoy your video!

A video should start playing in your browser above the input boxes, and will continue playing until you close the application.

Congrats, you have used pion-WebRTC! Now start building something cool
