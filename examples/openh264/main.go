package main

import (
	"fmt"
	"image"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/codec/openh264"
	_ "github.com/pion/mediadevices/pkg/driver/camera" // This is required to register camera adapter
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("usage: %s <path/to/file.h264>\n", os.Args[0])
		return
	}
	dest := os.Args[1]

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)

	params, err := openh264.NewParams()
	must(err)
	params.BitRate = 1_000_000 // 1mbps

	codecSelector := mediadevices.NewCodecSelector(
		mediadevices.WithVideoEncoders(&params),
	)

	mediaStream, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
		Video: func(c *mediadevices.MediaTrackConstraints) {
			c.FrameFormat = prop.FrameFormat(frame.FormatI420)
			c.Width = prop.Int(640)
			c.Height = prop.Int(480)
		},
		Codec: codecSelector,
	})
	must(err)

	videoTrack := mediaStream.GetVideoTracks()[0].(*mediadevices.VideoTrack)
	defer videoTrack.Close()

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

	reader, err := videoTrack.NewEncodedIOReader(params.RTPCodec().MimeType)
	must(err)
	defer reader.Close()

	out, err := os.Create(dest)
	must(err)
	fmt.Println("Recording... Press Ctrl+c to Set BitRate")
	go func() {
		_, err = io.Copy(out, reader)
	}()
	<-sigs
	if control, ok := reader.(codec.Controllable); ok {
		if ctrl, ok := control.Controller().(codec.KeyFrameController); ok {
			fmt.Println("Force Key")
			ctrl.ForceKeyFrame()
		}
		if ctrl, ok := control.Controller().(codec.BitRateController); ok {
			fmt.Println("SetBitRate")
			ctrl.SetBitRate(200_000)
		}
	}
	fmt.Println("Recording... Press Ctrl+c to stop")
	<-sigs
	must(err)
	fmt.Println("Your video has been recorded to", dest)
}
