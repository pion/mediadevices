package main

import (
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"log"

	"github.com/disintegration/imaging"
	pigo "github.com/esimov/pigo/core"
)

var (
	cascade    []byte
	err        error
	classifier *pigo.Pigo
)

func imgToGrayscale(img image.Image) []uint8 {
	bounds := img.Bounds()
	flatten := bounds.Dy() * bounds.Dx()
	grayImg := make([]uint8, flatten)

	i := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pix := img.At(x, y)
			grayPix := color.GrayModel.Convert(pix).(color.Gray)
			grayImg[i] = grayPix.Y
			i++
		}
	}
	return grayImg
}

// clusterDetection runs Pigo face detector core methods
// and returns a cluster with the detected faces coordinates.
func clusterDetection(img image.Image) []pigo.Detection {
	grayscale := imgToGrayscale(img)
	bounds := img.Bounds()
	cParams := pigo.CascadeParams{
		MinSize:     100,
		MaxSize:     600,
		ShiftFactor: 0.15,
		ScaleFactor: 1.1,
		ImageParams: pigo.ImageParams{
			Pixels: grayscale,
			Rows:   bounds.Dy(),
			Cols:   bounds.Dx(),
			Dim:    bounds.Dx(),
		},
	}

	if len(cascade) == 0 {
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
	}

	// Run the classifier over the obtained leaf nodes and return the detection results.
	// The result contains quadruplets representing the row, column, scale and detection score.
	dets := classifier.RunCascade(cParams, 0.0)

	// Calculate the intersection over union (IoU) of two clusters.
	dets = classifier.ClusterDetections(dets, 0)

	return dets
}

func drawCircle(img draw.Image, x0, y0, r int, c color.Color) {
	x, y, dx, dy := r-1, 0, 1, 1
	err := dx - (r * 2)

	for x > y {
		img.Set(x0+x, y0+y, c)
		img.Set(x0+y, y0+x, c)
		img.Set(x0-y, y0+x, c)
		img.Set(x0-x, y0+y, c)
		img.Set(x0-x, y0-y, c)
		img.Set(x0-y, y0-x, c)
		img.Set(x0+y, y0-x, c)
		img.Set(x0+x, y0-y, c)

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

func markFaces(img image.Image) image.Image {
	nrgba := imaging.Clone(img)
	dets := clusterDetection(img)
	for _, det := range dets {
		if det.Q < 5.0 {
			continue
		}

		drawCircle(nrgba, det.Col, det.Row, det.Scale/2, color.Black)
	}
	return nrgba
}
