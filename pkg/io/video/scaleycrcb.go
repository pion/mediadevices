package video

import (
	"image"
	"image/color"
)

type rgbLikeYCbCr struct {
	y  *image.Gray
	cb *image.Gray
	cr *image.Gray
}

func (p *rgbLikeYCbCr) ColorModel() color.Model {
	return color.RGBAModel
}

func (p *rgbLikeYCbCr) Bounds() image.Rectangle {
	return p.y.Rect
}

func (p *rgbLikeYCbCr) At(x, y int) color.Color {
	var yy, cb, cr uint8
	yy = p.y.GrayAt(x, y).Y
	if (image.Point{x, y}.In(p.cb.Rect)) {
		cb = p.cb.GrayAt(x, y).Y
		cr = p.cr.GrayAt(x, y).Y
	}
	return color.RGBA{yy, cb, cr, 255}
}

func (p *rgbLikeYCbCr) Set(x, y int, c color.Color) {
	rgb := c.(*color.RGBA64)
	p.y.SetGray(x, y, color.Gray{uint8(rgb.R / 0x100)})
	if (image.Point{x, y}.In(p.cb.Rect)) {
		p.cb.SetGray(x, y, color.Gray{uint8(rgb.G / 0x100)})
		p.cr.SetGray(x, y, color.Gray{uint8(rgb.B / 0x100)})
	}
}
