package video

import (
	"image"
)

// FrameBuffer is a buffer that can store any image format.
type FrameBuffer struct {
	buffer []uint8
	tmp    image.Image
}

// NewFrameBuffer creates a new FrameBuffer instance and initialize internal buffer
// with initialSize
func NewFrameBuffer(initialSize int) *FrameBuffer {
	return &FrameBuffer{
		buffer: make([]uint8, initialSize),
	}
}

func (buff *FrameBuffer) storeInOrder(srcs ...[]uint8) {
	var neededSize int

	for _, src := range srcs {
		neededSize += len(src)
	}

	if len(buff.buffer) < neededSize {
		if cap(buff.buffer) >= neededSize {
			buff.buffer = buff.buffer[:neededSize]
		} else {
			buff.buffer = make([]uint8, neededSize)
		}
	}

	var currentLen int
	for _, src := range srcs {
		copy(buff.buffer[currentLen:], src)
		currentLen += len(src)
	}
}

// Load loads the current owned image
func (buff *FrameBuffer) Load() image.Image {
	return buff.tmp
}

// StoreCopy makes a copy of src and store its copy. StoreCopy will reuse as much memory as it can
// from the previous copies. For example, if StoreCopy is given an image that has the same resolution
// and format from the previous call, StoreCopy will not allocate extra memory and only copy the content
// from src to the previous buffer.
func (buff *FrameBuffer) StoreCopy(src image.Image) {
	switch src := src.(type) {
	case *image.Alpha:
		clone, ok := buff.tmp.(*image.Alpha)
		if ok {
			*clone = *src
		} else {
			copied := *src
			clone = &copied
		}

		buff.storeInOrder(src.Pix)
		clone.Pix = buff.buffer[:len(src.Pix)]

		buff.tmp = clone
	case *image.Alpha16:
		clone, ok := buff.tmp.(*image.Alpha16)
		if ok {
			*clone = *src
		} else {
			copied := *src
			clone = &copied
		}

		buff.storeInOrder(src.Pix)
		clone.Pix = buff.buffer[:len(src.Pix)]

		buff.tmp = clone
	case *image.CMYK:
		clone, ok := buff.tmp.(*image.CMYK)
		if ok {
			*clone = *src
		} else {
			copied := *src
			clone = &copied
		}

		buff.storeInOrder(src.Pix)
		clone.Pix = buff.buffer[:len(src.Pix)]

		buff.tmp = clone
	case *image.Gray:
		clone, ok := buff.tmp.(*image.Gray)
		if ok {
			*clone = *src
		} else {
			copied := *src
			clone = &copied
		}

		buff.storeInOrder(src.Pix)
		clone.Pix = buff.buffer[:len(src.Pix)]

		buff.tmp = clone
	case *image.Gray16:
		clone, ok := buff.tmp.(*image.Gray16)
		if ok {
			*clone = *src
		} else {
			copied := *src
			clone = &copied
		}

		buff.storeInOrder(src.Pix)
		clone.Pix = buff.buffer[:len(src.Pix)]

		buff.tmp = clone
	case *image.NRGBA:
		clone, ok := buff.tmp.(*image.NRGBA)
		if ok {
			*clone = *src
		} else {
			copied := *src
			clone = &copied
		}

		buff.storeInOrder(src.Pix)
		clone.Pix = buff.buffer[:len(src.Pix)]

		buff.tmp = clone
	case *image.NRGBA64:
		clone, ok := buff.tmp.(*image.NRGBA64)
		if ok {
			*clone = *src
		} else {
			copied := *src
			clone = &copied
		}

		buff.storeInOrder(src.Pix)
		clone.Pix = buff.buffer[:len(src.Pix)]

		buff.tmp = clone
	case *image.RGBA:
		clone, ok := buff.tmp.(*image.RGBA)
		if ok {
			*clone = *src
		} else {
			copied := *src
			clone = &copied
		}

		buff.storeInOrder(src.Pix)
		clone.Pix = buff.buffer[:len(src.Pix)]

		buff.tmp = clone
	case *image.RGBA64:
		clone, ok := buff.tmp.(*image.RGBA64)
		if ok {
			*clone = *src
		} else {
			copied := *src
			clone = &copied
		}

		buff.storeInOrder(src.Pix)
		clone.Pix = buff.buffer[:len(src.Pix)]

		buff.tmp = clone
	case *image.NYCbCrA:
		clone, ok := buff.tmp.(*image.NYCbCrA)
		if ok {
			*clone = *src
		} else {
			copied := *src
			clone = &copied
		}

		var currentLen int
		buff.storeInOrder(src.Y, src.Cb, src.Cr, src.A)
		clone.Y = buff.buffer[currentLen : currentLen+len(src.Y) : currentLen+len(src.Y)]
		currentLen += len(src.Y)
		clone.Cb = buff.buffer[currentLen : currentLen+len(src.Cb) : currentLen+len(src.Cb)]
		currentLen += len(src.Cb)
		clone.Cr = buff.buffer[currentLen : currentLen+len(src.Cr) : currentLen+len(src.Cr)]
		currentLen += len(src.Cr)
		clone.A = buff.buffer[currentLen : currentLen+len(src.A) : currentLen+len(src.A)]

		buff.tmp = clone
	case *image.YCbCr:
		clone, ok := buff.tmp.(*image.YCbCr)
		if ok {
			*clone = *src
		} else {
			copied := *src
			clone = &copied
		}

		var currentLen int
		buff.storeInOrder(src.Y, src.Cb, src.Cr)
		clone.Y = buff.buffer[currentLen : currentLen+len(src.Y) : currentLen+len(src.Y)]
		currentLen += len(src.Y)
		clone.Cb = buff.buffer[currentLen : currentLen+len(src.Cb) : currentLen+len(src.Cb)]
		currentLen += len(src.Cb)
		clone.Cr = buff.buffer[currentLen : currentLen+len(src.Cr) : currentLen+len(src.Cr)]

		buff.tmp = clone
	default:
		var converted image.RGBA
		imageToRGBA(&converted, src)
		buff.StoreCopy(&converted)
	}
}
