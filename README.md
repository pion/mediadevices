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

![](img/demo.gif)

## Interfaces

| Interface  | Linux | Mac | Windows |
| :--------: | :---: | :-: | :-----: |
|   Camera   |  ✔️   | ✔️  |   ✔️    |
| Microphone |  ✔️   | ✖️  |   ✔️    |
|   Screen   |  ✔️   | ✖️  |   ✖️    |

### Camera

|   OS    |                           Library/Interface                             |
| :-----: | :---------------------------------------------------------------------: |
|  Linux  |        [Video4Linux](https://en.wikipedia.org/wiki/Video4Linux)         |
|   Mac   |       [AVFoundation](https://developer.apple.com/av-foundation/)        |
| Windows | [DirectShow](https://docs.microsoft.com/en-us/windows/win32/directshow) |

|                     Pixel Format                      | Linux | Mac | Windows |
| :---------------------------------------------------: | :---: | :-: | :-----: |
| [YUY2](https://www.fourcc.org/pixel-format/yuv-yuy2/) |  ✔️   | ✖️  |   ✔️    |
| [UYVY](https://www.fourcc.org/pixel-format/yuv-uyvy/) |  ✔️   | ✔️  |   ✖️    |
| [I420](https://www.fourcc.org/pixel-format/yuv-i420/) |  ✔️   | ✖️  |   ✖️    |
| [NV21](https://www.fourcc.org/pixel-format/yuv-nv21/) |  ✔️   | ✔️  |   ✖️    |
|         [MJPEG](https://www.fourcc.org/mjpg/)         |  ✔️   | ✖️  |   ✖️    |

### Microphone

|   OS    |                            Library/Interface                            |
| :-----: | :---------------------------------------------------------------------: |
|  Linux  |         [PulseAudio](https://en.wikipedia.org/wiki/PulseAudio)          |
|   Mac   |                                   N/A                                   |
| Windows |  [waveIn](https://docs.microsoft.com/en-us/windows/win32/api/mmeapi/)   |

### Screen casting

|   OS    |                            Library/Interface                            |
| :-----: | :---------------------------------------------------------------------: |
|  Linux  |          [X11](https://en.wikipedia.org/wiki/X_Window_System)           |
|   Mac   |                                   N/A                                   |
| Windows |                                   N/A                                   |

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
