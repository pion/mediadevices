# mediadevices

Go implementation of the [MediaDevices](https://developer.mozilla.org/en-US/docs/Web/API/MediaDevices) API.

![](img/demo.gif)

## Interfaces

| Interface  | Linux | Mac | Windows |
| :--------: | :---: | :-: | :-----: |
|   Camera   |  ✔️   | ✖️  |   ✖️    |
| Microphone |  ✖️   | ✖️  |   ✖️    |

### Camera

|   OS    |                    Library/Interface                     |
| :-----: | :------------------------------------------------------: |
|  Linux  | [Video4Linux](https://en.wikipedia.org/wiki/Video4Linux) |
|   Mac   |                           N/A                            |
| Windows |                           N/A                            |

|                     Pixel Format                      | Linux | Mac | Windows |
| :---------------------------------------------------: | :---: | :-: | :-----: |
| [YUY2](https://www.fourcc.org/pixel-format/yuv-yuy2/) |  ✔️   | ✖️  |   ✖️    |
| [I420](https://www.fourcc.org/pixel-format/yuv-i420/) |  ✔️   | ✖️  |   ✖️    |
| [NV21](https://www.fourcc.org/pixel-format/yuv-nv21/) |  ✔️   | ✖️  |   ✖️    |
|         [MJPEG](https://www.fourcc.org/mjpg/)         |  ✔️   | ✖️  |   ✖️    |

### Microphone

N/A

## Contributing

- [Lukas Herman](https://github.com/lherman-cs) - _Original Author_

## Project Status

[![Stargazers over time](https://starchart.cc/pion/mediadevices.svg)](https://starchart.cc/pion/mediadevices)

## References

- https://developer.mozilla.org/en-US/docs/Web/Media/Formats/WebRTC_codecs
- https://tools.ietf.org/html/rfc7742
