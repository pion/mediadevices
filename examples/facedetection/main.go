package main

import (
	"image"
	"io/ioutil"
	"log"
	"time"

	pigo "github.com/esimov/pigo/core"
	"github.com/pion/mediadevices"
	_ "github.com/pion/mediadevices/pkg/driver/camera" // This is required to register camera adapter
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/prop"
)

const (
	confidenceLevel = 5.0
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

func detectFace(frame *image.YCbCr) bool {
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

	for _, det := range dets {
		if det.Q >= confidenceLevel {
			return true
		}
	}

	return false
}

func main() {
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

	mediaStream, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
		Video: func(c *mediadevices.MediaTrackConstraints) {
			c.FrameFormat = prop.FrameFormatExact(frame.FormatUYVY)
			c.Width = prop.Int(640)
			c.Height = prop.Int(480)
		},
	})
	must(err)

	// since we're trying to access the raw data, we need to cast Track to its real type, *mediadevices.VideoTrack
	videoTrack := mediaStream.GetVideoTracks()[0].(*mediadevices.VideoTrack)
	defer videoTrack.Close()

	videoReader := videoTrack.NewReader(false)
	// To save resources, we can simply use 4 fps to detect faces.
	ticker := time.NewTicker(time.Millisecond * 250)
	defer ticker.Stop()

	for range ticker.C {
		frame, release, err := videoReader.Read()
		must(err)

		// Since we asked the frame format to be exactly YUY2 in GetUserMedia, we can guarantee that it must be YCbCr
		if detectFace(frame.(*image.YCbCr)) {
			log.Println("Detect a face")
		}

		release()
	}
}
