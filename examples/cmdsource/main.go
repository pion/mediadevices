package main

// !!!! This example requires ffmpeg to be installed !!!!

import (
	"errors"
	"fmt"
	"image"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/codec/x264" // This is required to use H264 video encoder
	"github.com/pion/mediadevices/pkg/driver"
	"github.com/pion/mediadevices/pkg/driver/cmdsource"
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

// handy correlation between the names of frame formats in pion media devices and the same -pix_fmt as passed to ffmpeg
var ffmpegFrameFormatMap = map[frame.Format]string{
	frame.FormatI420: "yuv420p",
	frame.FormatNV21: "nv21",
	frame.FormatNV12: "nv12",
	frame.FormatYUY2: "yuyv422",
	frame.FormatUYVY: "uyvy422",
	frame.FormatZ16:  "gray",
}

func ffmpegTestPatternCmd(width int, height int, frameRate float32, frameFormat frame.Format) string {
	// returns the (command-line) command to tell the ffmpeg program to output a test video stream with the given pixel format, size and framerate to stdout:
	return fmt.Sprintf("ffmpeg -hide_banner -f lavfi -i testsrc=size=%dx%d:rate=%f -vf realtime -f rawvideo -pix_fmt %s -", width, height, frameRate, ffmpegFrameFormatMap[frameFormat])
}

func getMediaDevicesDriverId(label string) (string, error) {
	drivers := driver.GetManager().Query(func(d driver.Driver) bool {
		return d.Info().Label == label
	})
	if len(drivers) == 0 {
		return "", errors.New("Failed to find the media devices driver for device label: " + label)
	}
	return drivers[0].ID(), nil
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("usage: %s <path/to/output_file.h264>\n", os.Args[0])
		return
	}
	dest := os.Args[1]

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)

	x264Params, err := x264.NewParams()
	must(err)
	x264Params.Preset = x264.PresetMedium
	x264Params.BitRate = 1_000_000 // 1mbps

	codecSelector := mediadevices.NewCodecSelector(
		mediadevices.WithVideoEncoders(&x264Params),
	)

	// configure source video properties (raw video stream format that we should expect the command to output)
	label := "My Cool Video"
	videoProps := prop.Media{
		Video: prop.Video{
			Width:       640,
			Height:      480,
			FrameFormat: frame.FormatI420,
			FrameRate:   30,
		},
		// OR Audio: prop.Audio{}
	}

	// Add the command source:
	cmdString := ffmpegTestPatternCmd(videoProps.Video.Width, videoProps.Video.Height, videoProps.Video.FrameRate, videoProps.FrameFormat)
	err = cmdsource.AddVideoCmdSource(label, cmdString, []prop.Media{videoProps}, 10, true)
	must(err)

	// Now your video command source will be a driver in mediaDevices:
	driverId, err := getMediaDevicesDriverId(label)
	must(err)

	mediaStream, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
		Video: func(c *mediadevices.MediaTrackConstraints) {
			c.DeviceID = prop.String(driverId)
		},
		Codec: codecSelector,
	})
	must(err)

	videoTrack := mediaStream.GetVideoTracks()[0].(*mediadevices.VideoTrack)
	defer videoTrack.Close()
	//// --- OR (if the track was setup as audio) --
	// audioTrack := mediaStream.GetAudioTracks()[0].(*mediadevices.AudioTrack)
	// defer audioTrack.Close()

	// Do whatever you want with the track, the rest of this example is the same as the archive example:
	// =================================================================================================

	videoTrack.Transform(video.TransformFunc(func(r video.Reader) video.Reader {
		return video.ReaderFunc(func() (img image.Image, release func(), err error) {
			// we send io.EOF signal to the encoder reader to stop reading. Therefore, io.Copy
			// will finish its execution and the program will finish
			select {
			case <-sigs:
				return nil, func() {}, io.EOF
			default:
			}

			return r.Read()
		})
	}))

	reader, err := videoTrack.NewEncodedIOReader(x264Params.RTPCodec().MimeType)
	must(err)
	defer reader.Close()

	out, err := os.Create(dest)
	must(err)

	fmt.Println("Recording... Press Ctrl+c to stop")
	_, err = io.Copy(out, reader)
	must(err)
	videoTrack.Close()                   // Ideally we should close the track before the io.Copy is done to save every last frame
	<-time.After(100 * time.Millisecond) // Give a bit of time for the ffmpeg stream to stop cleanly before the program exits
	fmt.Println("Your video has been recorded to", dest)
}
