package frame

import (
	"image"
	"image/color"
)

type RGB24Img struct {
	// Pix holds the image's pixels, in R, G, B order. The pixel at
	// (x, y) starts at Pix[(y-Rect.Min.Y) + (x-Rect.Min.X)*3].
	Pix    []uint8
	Rect   image.Rectangle
	Stride int
}

func decodeRGB24(frame []byte, width, height int) (image.Image, func(), error) {
	return &RGB24Img{
		Pix: frame,
		Rect: image.Rectangle{
			Min: image.Point{
				X: 0,
				Y: 0,
			},
			Max: image.Point{
				X: width,
				Y: height,
			},
		},
		Stride: width * 3,
	}, func() {}, nil
}

func (p *RGB24Img) ColorModel() color.Model {
	return color.RGBAModel
}
func (p *RGB24Img) Bounds() image.Rectangle {
	return p.Rect
}

func (p *RGB24Img) PixOffset(x, y int) int {
	return (y-p.Rect.Min.Y)*p.Stride + (x-p.Rect.Min.X)*3
}
func (p *RGB24Img) At(x, y int) color.Color {
	if !(image.Point{x, y}.In(p.Rect)) {
		return color.RGBA{}
	}
	i := p.PixOffset(x, y)
	s := p.Pix[i : i+3 : i+3] // Small capacity improves performance, see https://golang.org/issue/27857
	return color.RGBA{s[0], s[1], s[2], 1}
}
