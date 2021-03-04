package frame

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
)

func decodeZ16(frame []byte, width, height int) (image.Image, func(), error) {
	expectedSize := 2 * (width * height)
	if expectedSize  != len(frame) {
		return nil, func() {}, fmt.Errorf("frame length (%d) not expected size (%d)", len(frame), expectedSize)
	}
	img := image.NewGray16(image.Rect(0, 0, width, height))
	/*
		v4l specifies images in terms of series of lines which is perplexing because the
		example in https://www.kernel.org/doc/html/v5.9/userspace-api/media/v4l/pixfmt-z16.html
		seems to swap the subscript of each depth value.

		Clear example:
		Width: 3, Height: 4
		[Z_low(x_0,y_0), Z_high(x_0,y_0), Z_low(x_1,y_0), Z_high(x_1,y_0), Z_low(x_2,y_0), Z_high(x_2,y_0),
		 Z_low(x_0,y_1), Z_high(x_0,y_1), Z_low(x_1,y_1), Z_high(x_1,y_1), Z_low(x_2,y_1), Z_high(x_2,y_1),
		 Z_low(x_0,y_2), Z_high(x_0,y_2), Z_low(x_1,y_2), Z_high(x_1,y_2), Z_low(x_2,y_2), Z_high(x_2,y_2),
		 Z_low(x_0,y_3), Z_high(x_0,y_3), Z_low(x_1,y_3), Z_high(x_1,y_3), Z_low(x_2,y_3), Z_high(x_2,y_3)]
	*/
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			idx := 2 * (x + (y * width))
			z := binary.LittleEndian.Uint16(frame[idx : idx+2])
			img.SetGray16(x, y, color.Gray16{Y: z})
		}
	}
	return img, func() {}, nil
}
