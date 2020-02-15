# mediadevices

Go implementation of the [MediaDevices](https://developer.mozilla.org/en-US/docs/Web/API/MediaDevices) API.

![](img/demo.gif)

## Interfaces

| Interface  | Linux | Mac | Windows |
| :--------: | :---: | :-: | :-----: |
|   Camera   |  ✔️   | ✖️  |   ✖️    |
| Microphone |  ✔️   | ✖️  |   ✖️    |
|   Screen   |  ✔️   | ✖️  |   ✖️    |

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

|   OS    |                    Library/Interface                     |
| :-----: | :------------------------------------------------------: |
|  Linux  | [PulseAudio](https://en.wikipedia.org/wiki/PulseAudio)   |
|   Mac   |                           N/A                            |
| Windows |                           N/A                            |

### Screen casting

|   OS    |                    Library/Interface                     |
| :-----: | :------------------------------------------------------: |
|  Linux  | [X11](https://en.wikipedia.org/wiki/X_Window_System)     |
|   Mac   |                           N/A                            |
| Windows |                           N/A                            |

## Codecs

| Audio Codec |                    Library/Interface                     |
| :---------: | :------------------------------------------------------: |
|    OPUS     | [libopus](http://opus-codec.org/)                        |

| Video Codec |                    Library/Interface                     |
| :---------: | :------------------------------------------------------: |
|    H.264    | [OpenH264](https://www.openh264.org/)                    |
|     VP8     | [libvpx](https://www.webmproject.org/code/)              |
|     VP9     | [libvpx](https://www.webmproject.org/code/)              |

## Usage

[Wiki](https://github.com/pion/mediadevices/wiki)

## Contributing

- [Lukas Herman](https://github.com/lherman-cs) - _Original Author_
* [Atsushi Watanabe](https://github.com/at-wat) - _VP8, Screencast, etc._

## Project Status

[![Stargazers over time](https://starchart.cc/pion/mediadevices.svg)](https://starchart.cc/pion/mediadevices)

## References

- https://developer.mozilla.org/en-US/docs/Web/Media/Formats/WebRTC_codecs
- https://tools.ietf.org/html/rfc7742
