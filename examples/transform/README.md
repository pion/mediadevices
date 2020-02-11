## Instructions

### Download transform example

```
go get github.com/pion/mediadevices/examples/transform
```

### Open example page

[jsfiddle.net](https://jsfiddle.net/z7ms3u5r/) you should see two text-areas and a 'Start Session' button

### Run transform with your browsers SessionDescription as stdin

In the jsfiddle the top textarea is your browser, copy that and:

#### Linux

Run `echo $BROWSER_SDP | transform`

### Input transform's SessionDescription into your browser

Copy the text that `transform` just emitted and copy into second text area

### Hit 'Start Session' in jsfiddle, enjoy your video!

A video should start playing in your browser above the input boxes, and will continue playing until you close the application.

Congrats, you have used pion-WebRTC! Now start building something cool
