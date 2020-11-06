package main

import (
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"net"
	"os"

	pigo "github.com/esimov/pigo/core"
	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/codec/vpx"       // This is required to use h264 video encoder
	_ "github.com/pion/mediadevices/pkg/driver/camera" // This is required to register camera adapter
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

const (
	confidenceLevel = 9.5
	mtu             = 1000
	thickness       = 5
)

var (
	cascade    []byte
	classifier *pigo.Pigo
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func detectFaces(frame *image.YCbCr) []pigo.Detection {
	bounds := frame.Bounds()
	cascadeParams := pigo.CascadeParams{
		MinSize:     100,
		MaxSize:     600,
		ShiftFactor: 0.15,
		ScaleFactor: 1.1,
		ImageParams: pigo.ImageParams{
			Pixels: frame.Y, // Y in YCbCr should be enough to detect faces
			Rows:   bounds.Dy(),
			Cols:   bounds.Dx(),
			Dim:    bounds.Dx(),
		},
	}

	// Run the classifier over the obtained leaf nodes and return the detection results.
	// The result contains quadruplets representing the row, column, scale and detection score.
	dets := classifier.RunCascade(cascadeParams, 0.0)

	// Calculate the intersection over union (IoU) of two clusters.
	dets = classifier.ClusterDetections(dets, 0)
	return dets
}

func drawRect(frame *image.YCbCr, x0, y0, size int) {
	if x0 < 0 {
		x0 = 0
	}

	if y0 < 0 {
		y0 = 0
	}

	width := frame.Bounds().Dx()
	height := frame.Bounds().Dy()
	x1 := x0 + size
	y1 := y0 + size

	if x1 >= width {
		x1 = width - 1
	}

	if y1 >= height {
		y1 = height - 1
	}

	convert := func(x, y int) int {
		return y*width + x
	}

	for x := x0; x < x1; x++ {
		for t := 0; t < thickness; t++ {
			frame.Y[convert(x, y0+t)] = 0
			frame.Y[convert(x, y1-t)] = 0
		}
	}

	for y := y0; y < y1; y++ {
		for t := 0; t < thickness; t++ {
			frame.Y[convert(x0+t, y)] = 0
			frame.Y[convert(x1-t, y)] = 0
		}
	}
}

func detectFace(r video.Reader) video.Reader {
	return video.ReaderFunc(func() (img image.Image, release func(), err error) {
		img, release, err = r.Read()
		if err != nil {
			return
		}

		yuv := img.(*image.YCbCr)
		dets := detectFaces(yuv)
		for _, det := range dets {
			if det.Q < confidenceLevel {
				continue
			}

			drawRect(yuv, det.Col-det.Scale/2, det.Row-det.Scale/2, det.Scale)
		}
		return
	})
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("usage: %s host:port\n", os.Args[0])
		return
	}
	dest := os.Args[1]

	// prepare face detector
	var err error
	cascade, err = ioutil.ReadFile("facefinder")
	if err != nil {
		log.Fatalf("Error reading the cascade file: %s", err)
	}
	p := pigo.NewPigo()

	// Unpack the binary file. This will return the number of cascade trees,
	// the tree depth, the threshold and the prediction from tree's leaf nodes.
	classifier, err = p.Unpack(cascade)
	if err != nil {
		log.Fatalf("Error unpacking the cascade file: %s", err)
	}

	vp8Params, err := vpx.NewVP8Params()
	must(err)
	vp8Params.BitRate = 1_000_000 // 100kbps

	codecSelector := mediadevices.NewCodecSelector(
		mediadevices.WithVideoEncoders(&vp8Params),
	)

	mediaStream, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
		Video: func(c *mediadevices.MediaTrackConstraints) {
			c.FrameFormat = prop.FrameFormatExact(frame.FormatUYVY)
			c.Width = prop.Int(640)
			c.Height = prop.Int(480)
		},
		Codec: codecSelector,
	})
	must(err)

	// since we're trying to access the raw data, we need to cast Track to its real type, *mediadevices.VideoTrack
	videoTrack := mediaStream.GetVideoTracks()[0].(*mediadevices.VideoTrack)
	defer videoTrack.Close()

	videoTrack.Transform(detectFace)

	rtpReader, err := videoTrack.NewRTPReader(vp8Params.RTPCodec().Name, mtu)
	must(err)

	addr, err := net.ResolveUDPAddr("udp", dest)
	must(err)
	conn, err := net.DialUDP("udp", nil, addr)
	must(err)

	buff := make([]byte, mtu)
	for {
		pkts, release, err := rtpReader.Read()
		must(err)

		for _, pkt := range pkts {
			n, err := pkt.MarshalTo(buff)
			must(err)

			_, err = conn.Write(buff[:n])
			must(err)
		}

		release()
	}
}
