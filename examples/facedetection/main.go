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
	confidenceLevel = 5.0
	mtu             = 1000
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

func drawCircle(frame *image.YCbCr, x0, y0, r int) {
	width := frame.Bounds().Dx()
	x, y, dx, dy := r-1, 0, 1, 1
	err := dx - (r * 2)

	convert := func(x, y int) int {
		return y*width + x
	}

	for x > y {
		frame.Y[convert(x0+x, y0+y)] = 0
		frame.Y[convert(x0+y, y0+x)] = 0
		frame.Y[convert(x0-y, y0+x)] = 0
		frame.Y[convert(x0-x, y0+y)] = 0
		frame.Y[convert(x0-x, y0-y)] = 0
		frame.Y[convert(x0-y, y0-x)] = 0
		frame.Y[convert(x0+y, y0-x)] = 0
		frame.Y[convert(x0+x, y0-y)] = 0

		if err <= 0 {
			y++
			err += dy
			dy += 2
		}
		if err > 0 {
			x--
			dx += 2
			err += dx - (r * 2)
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

			drawCircle(yuv, det.Col, det.Row, det.Scale/2)
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
