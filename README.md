<h1 align="center">
  <br>
  Pion MediaDevices
  <br>
</h1>
<h4 align="center">Go implementation of the <a href="https://developer.mozilla.org/en-US/docs/Web/API/MediaDevices">MediaDevices</a> API</h4>
<p align="center">
  <a href="https://pion.ly/slack"><img src="https://img.shields.io/badge/join-us%20on%20slack-gray.svg?longCache=true&logo=slack&colorB=brightgreen" alt="Slack Widget"></a>
  <a href="https://github.com/pion/mediadevices/actions"><img src="https://github.com/pion/mediadevices/workflows/CI/badge.svg?branch=master" alt="Build status"></a> 
  <a href="https://pkg.go.dev/github.com/pion/mediadevices"><img src="https://godoc.org/github.com/pion/mediadevices?status.svg" alt="GoDoc"></a>
  <a href="https://codecov.io/gh/pion/mediadevices"><img src="https://codecov.io/gh/pion/mediadevices/branch/master/graph/badge.svg" alt="Coverage Status"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License: MIT"></a>
</p>
<br>

`mediadevices` provides access to media input devices like cameras, microphones, and screen capture. It can also be used to encode your video/audio stream to various codec selections. `mediadevices` abstracts away the complexities of interacting with things like hardware and codecs allowing you to focus on building appilcations, interacting only with an amazingly simple, easy, and elegant API!

## Install

`go get -u github.com/pion/mediadevices`

## Usage

The following snippet shows how to capture a camera stream and store a frame as a jpeg image:

```go
package main

import (
	"image/jpeg"
	"os"

	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/prop"

	// This is required to register camera adapter
	_ "github.com/pion/mediadevices/pkg/driver/camera" 
	// Note: If you don't have a camera or your adapters are not supported,
	//       you can always swap your adapters with our dummy adapters below.
	// _ "github.com/pion/mediadevices/pkg/driver/videotest"
)

func main() {
	stream, _ := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
		Video: func(constraint *mediadevices.MediaTrackConstraints) {
			// Query for ideal resolutions
			constraint.Width = prop.Int(600)
			constraint.Height = prop.Int(400)
		},
	})

	// Since track can represent audio as well, we need to cast it to 
	// *mediadevices.VideoTrack to get video specific functionalities
	track := stream.GetVideoTracks()[0]
	videoTrack := track.(*mediadevices.VideoTrack)
	defer videoTrack.Close()

	// Create a new video reader to get the decoded frames. Release is used 
	// to return the buffer to hold frame back to the source so that the buffer 
	// can be reused for the next frames.
	videoReader := videoTrack.NewReader(false)
	frame, release, _ := videoReader.Read()
	defer release()

	// Since frame is the standard image.Image, it's compatible with Go standard 
	// library. For example, capturing the first frame and store it as a jpeg image.
	output, _ := os.Create("frame.jpg")
	jpeg.Encode(output, frame, nil)
}


```

## More Examples

* [Webrtc](/examples/webrtc) - Use Webrtc to create a realtime peer-to-peer video call
* [Face Detection](/examples/facedetection) - Use a machine learning algorithm to detect faces in a camera stream
* [RTP Stream](examples/rtp) - Capture camera stream, encode it in H264/VP8/VP9, and send it to a RTP server
* [HTTP Broadcast](/examples/http) - Broadcast camera stream through HTTP with MJPEG
* [Archive](/examples/archive) - Archive H264 encoded video stream from a camera

## Available Media Inputs

| Input  | Linux | Mac | Windows |
| :--------: | :---: | :-: | :-----: |
|   Camera   |  ✔️   | ✔️  |   ✔️    |
| Microphone |  ✔️   | ✔️  |   ✔️    |
|   Screen   |  ✔️   | ✔️  |   ✔️    |

By default, there's no media input registered. This decision was made to allow you to pay what you need. Therefore, you need to import the associated packages for the media inputs. For example, if you want to use a camera, you need to import the camera package as a side effect:

```go
import (
	...
	_ "github.com/pion/mediadevices/pkg/driver/camera"
)
```

## Available Codecs

In order to encode your video/audio, `mediadevices` needs to know what codecs that you want to use and their parameters. To do this, you need to import the associated packages for the codecs, and add them to the codec selector that you'll pass to `GetUserMedia`:

```go
package main

import (
	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/codec/x264"      // This is required to use H264 video encoder
	_ "github.com/pion/mediadevices/pkg/driver/camera" // This is required to register camera adapter
)

func main() {
	// configure codec specific parameters
	x264Params, _ := x264.NewParams()
	x264Params.Preset = x264.PresetMedium
	x264Params.BitRate = 1_000_000 // 1mbps

	codecSelector := mediadevices.NewCodecSelector(
		mediadevices.WithVideoEncoders(&x264Params),
	)
	
	mediaStream, _ := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
		Video: func(c *mediadevices.MediaTrackConstraints) {},
		Codec: codecSelector, // let GetUsermedia know available codecs
	})
}
```

Since `mediadevices` doesn't implement the video/audio codecs, it needs to call the codec libraries from the system through cgo. Therefore, you're required to install the codec libraries before you can use them in `mediadevices`. In the next section, it shows a list of available codecs, where the packages are defined (documentation linked), and installation instructions.

Note: we do not provide recommendations on choosing one codec or another as it is very complex and can be subjective.

### Video Codecs

#### x264
A free software library and application for encoding video streams into the H.264/MPEG-4 AVC compression format.

* Package: [github.com/pion/mediadevices/pkg/codec/x264](https://pkg.go.dev/github.com/pion/mediadevices/pkg/codec/x264)
* Installation:
  * Mac: `brew install x264`
  * Ubuntu: `apt install libx264-dev`
  
#### mmal
A framework to enable H264 hardware encoding for Raspberry Pi or boards that use VideoCore GPUs.

* Package: [github.com/pion/mediadevices/pkg/codec/mmal](https://pkg.go.dev/github.com/pion/mediadevices/pkg/codec/mmal)
* Installation: no installation needed, mmal should come built in Raspberry Pi devices

#### openh264
A codec library which supports H.264 encoding and decoding. It is suitable for use in real time applications.

* Package: [github.com/pion/mediadevices/pkg/codec/openh264](https://pkg.go.dev/github.com/pion/mediadevices/pkg/codec/openh264)
* Installation: no installation needed, included as a static binary

#### vpx
A free software video codec library from Google and the Alliance for Open Media that implements VP8/VP9 video coding formats.

* Package: [github.com/pion/mediadevices/pkg/codec/vpx](https://pkg.go.dev/github.com/pion/mediadevices/pkg/codec/vpx)
* Installation:
  * Mac: `brew install libvpx`
  * Ubuntu: `apt install libvpx-dev`
  
#### vaapi
An open source API that allows applications such as VLC media player or GStreamer to use hardware video acceleration capabilities (currently support VP8/VP9).

* Package: [github.com/pion/mediadevices/pkg/codec/vaapi](https://pkg.go.dev/github.com/pion/mediadevices/pkg/codec/vaapi)
* Installation:
  * Ubuntu: `apt install libva-dev`


### Audio Codecs

#### opus
A totally open, royalty-free, highly versatile audio codec.

* Package: [github.com/pion/mediadevices/pkg/codec/opus](https://pkg.go.dev/github.com/pion/mediadevices/pkg/codec/opus)
* Installation:
  * Mac: `brew install opus`
  * Ubuntu: `apt install libopus-dev`

## Benchmark

Result as of Nov 4, 2020 with Go 1.14 on a Raspberry pi 3, `mediadevices` can produce video, encode, send across network, and decode at **720p, 30 fps with < 500 ms latency**.  

The test was taken by capturing a camera stream, decoding the raw frames, encoding the video stream with mmal, and sending the stream through Webrtc.

## FAQ

### Failed to find the best driver that fits the constraints

`mediadevices` provides an automated driver discovery through `GetUserMedia` and `GetDisplayMedia`. The driver discover algorithm works something like:

1. Open all registered drivers
2. Get all properties (property describes what a driver is capable of, e.g. resolution, frame rate, etc.) from opened drivers
3. Find the best property that meets the criteria

So, when `mediadevices` returns `failed to find the best driver that fits the constraints` error, one of the following conditions might have occured:
* Driver was not imported as a side effect in your program, e.g. `import _ github.com/pion/mediadevices/pkg/driver/camera`
* Your constraint is too strict that there's no driver can fullfil your requirements. In this case, you can try to turn up the debug level by specifying the following environment variable: `export PION_LOG_DEBUG=all` to see what was too strict and tune that.
* Your driver is not supported/implemented. In this case, you can either let us know (file an issue) and wait for the maintainers to implement it. Or, you can implement it yourself and register it through `RegisterDriverAdapter`

### Failed to find vpx/x264/mmal/opus codecs

Since `mediadevices` uses cgo to access video/audio codecs, it needs to find these libraries from the system. To accomplish this, [pkg-config](https://www.freedesktop.org/wiki/Software/pkg-config/) is used for library discovery.

If you see the following error message at compile time:
```
# pkg-config --cflags  -- vpx
Package vpx was not found in the pkg-config search path.
Perhaps you should add the directory containing `vpx.pc'
to the PKG_CONFIG_PATH environment variable
No package 'vpx' found
pkg-config: exit status 1
```

There are 2 common problems:

* The required codec library is not installed (vpx in this example). In this case, please refer to the [available codecs](#available-codecs).
* Pkg-config fails to find the `.pc` files for this codec ([reference](https://people.freedesktop.org/~dbn/pkg-config-guide.html#using)). In this case, you need to find where the codec library's `.pc` is stored, and let pkg-config knows with: `export PKG_CONFIG_PATH=/path/to/directory`.


## Community
Pion has an active community on the [Slack](https://pion.ly/slack).

Follow the [Pion Twitter](https://twitter.com/_pion) for project updates and important WebRTC news.

We are always looking to support **your projects**. Please reach out if you have something to build!
If you need commercial support or don't want to use public methods you can contact us at [team@pion.ly](mailto:team@pion.ly)

## Contributing
Check out the **[contributing wiki](https://github.com/pion/webrtc/wiki/Contributing)** to join the group of amazing people making this project possible:

* [Lukas Herman](https://github.com/lherman-cs) - _Original Author_
* [Atsushi Watanabe](https://github.com/at-wat) - _VP8, Screencast, etc._

## License
MIT License - see [LICENSE](LICENSE) for full text
