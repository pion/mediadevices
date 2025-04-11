package vnc

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"io"
)

// An Encoding implements a method for encoding pixel data that is
// sent by the server to the client.
type Encoding interface {
	// The number that uniquely identifies this encoding type.
	Type() int32

	// Read reads the contents of the encoded pixel data from the reader.
	// This should return a new Encoding implementation that contains
	// the proper data.
	Read(*ClientConn, *Rectangle, io.Reader) (Encoding, error)
}
type CursorEncoding struct {
}

func (*CursorEncoding) Type() int32 {
	return -239
}
func (*CursorEncoding) Read(c *ClientConn, rect *Rectangle, r io.Reader) (Encoding, error) {
	size := int(rect.Height) * int(rect.Width) * int(c.PixelFormat.BPP) / 8
	pixelBytes := make([]uint8, size)
	if _, err := io.ReadFull(r, pixelBytes); err != nil {
		return nil, err
	}
	mask := ((int(rect.Width) + 7) / 8) * int(rect.Height)
	maskBytes := make([]uint8, mask)
	if _, err := io.ReadFull(r, maskBytes); err != nil {
		return nil, err
	}
	return &CursorEncoding{}, nil
}

// RawEncoding is raw pixel data sent by the server.
//
// See RFC 6143 Section 7.7.1
type RawEncoding struct {
	Colors   []Color
	RawPixel []uint32 //RGBA
}

func (*RawEncoding) Type() int32 {
	return 0
}

func (*RawEncoding) Read(c *ClientConn, rect *Rectangle, r io.Reader) (Encoding, error) {
	//fmt.Println("RawEncoding")
	bytesPerPixel := c.PixelFormat.BPP / 8
	pixelBytes := make([]uint8, bytesPerPixel)

	var byteOrder binary.ByteOrder = binary.LittleEndian
	if c.PixelFormat.BigEndian {
		byteOrder = binary.BigEndian
	}

	colors := make([]Color, int(rect.Height)*int(rect.Width))
	rawPixels := make([]uint32, int(rect.Height)*int(rect.Width))
	for y := uint16(0); y < rect.Height; y++ {
		for x := uint16(0); x < rect.Width; x++ {
			if _, err := io.ReadFull(r, pixelBytes); err != nil {
				return nil, err
			}

			var rawPixel uint32
			if c.PixelFormat.BPP == 8 {
				rawPixel = uint32(pixelBytes[0])
			} else if c.PixelFormat.BPP == 16 {
				rawPixel = uint32(byteOrder.Uint16(pixelBytes))
			} else if c.PixelFormat.BPP == 32 {
				rawPixel = byteOrder.Uint32(pixelBytes)
			}
			//rawPixels[int(y)*int(rect.Width)+int(x)]=rawPixel
			color := &colors[int(y)*int(rect.Width)+int(x)]
			if c.PixelFormat.TrueColor {
				color.R = uint16((rawPixel >> c.PixelFormat.RedShift) & uint32(c.PixelFormat.RedMax))
				color.G = uint16((rawPixel >> c.PixelFormat.GreenShift) & uint32(c.PixelFormat.GreenMax))
				color.B = uint16((rawPixel >> c.PixelFormat.BlueShift) & uint32(c.PixelFormat.BlueMax))
				if c.PixelFormat.BPP == 16 {
					color.B = color.B<<3 | color.B>>2
					color.G = color.G<<2 | color.G>>2
					color.R = color.R<<3 | color.R>>2
				}
			} else {
				*color = c.ColorMap[rawPixel]
			}
			rawPixels[int(y)*int(rect.Width)+int(x)] = uint32(0xff)<<24 | uint32(color.B)<<16 | uint32(color.G)<<8 | uint32(color.R)
			//fmt.Printf("%x %x",rawPixel,rawPixels[int(y)*int(rect.Width)+int(x)])
		}
	}

	return &RawEncoding{colors, rawPixels}, nil
}

// ZlibEncoding is raw pixel data sent by the server compressed by Zlib.
//
// A single Zlib stream is created. There is only a single header for a framebuffer request response.
type ZlibEncoding struct {
	Colors   []Color
	RawPixel []uint32
	ZStream  *bytes.Buffer
	ZReader  io.ReadCloser
}

func (*ZlibEncoding) Type() int32 {
	return 6
}

func (ze *ZlibEncoding) Read(c *ClientConn, rect *Rectangle, r io.Reader) (Encoding, error) {
	//fmt.Println("ZlibEncoding")
	bytesPerPixel := c.PixelFormat.BPP / 8
	pixelBytes := make([]uint8, bytesPerPixel)

	var byteOrder binary.ByteOrder = binary.LittleEndian
	if c.PixelFormat.BigEndian {
		byteOrder = binary.BigEndian
	}

	// Format
	// 4 bytes        | uint32 | length
	// 'length' bytes | []byte | zlibData
	// Read zlib length
	var zipLength uint32
	err := binary.Read(r, binary.BigEndian, &zipLength)
	if err != nil {
		return nil, err
	}

	// Read all compressed data
	zBytes := make([]byte, zipLength)
	if _, err := io.ReadFull(r, zBytes); err != nil {
		return nil, err
	}

	// Create new zlib stream if needed
	if ze.ZStream == nil {
		// Create and save the buffer
		ze.ZStream = new(bytes.Buffer)
		ze.ZStream.Write(zBytes)

		// Create a reader for the buffer
		ze.ZReader, err = zlib.NewReader(ze.ZStream)
		if err != nil {
			return nil, err
		}

		// This is needed to avoid 'zlib missing header'
	} else {
		// Just append if already created
		ze.ZStream.Write(zBytes)
	}

	// Calculate zlib decompressed size
	sizeToRead := int(rect.Height) * int(rect.Width) * int(bytesPerPixel)

	// Create buffer for bytes
	colorBytes := make([]byte, sizeToRead)

	// Read all data from zlib stream
	read, err := io.ReadFull(ze.ZReader, colorBytes)
	if read != sizeToRead || err != nil {
		return nil, err
	}

	// Create buffer for raw encoding
	colorReader := bytes.NewReader(colorBytes)

	colors := make([]Color, int(rect.Height)*int(rect.Width))
	rawPixels := make([]uint32, int(rect.Height)*int(rect.Width))
	for y := uint16(0); y < rect.Height; y++ {
		for x := uint16(0); x < rect.Width; x++ {
			if _, err := io.ReadFull(colorReader, pixelBytes); err != nil {
				return nil, err
			}

			var rawPixel uint32
			if c.PixelFormat.BPP == 8 {
				rawPixel = uint32(pixelBytes[0])
			} else if c.PixelFormat.BPP == 16 {
				rawPixel = uint32(byteOrder.Uint16(pixelBytes))
			} else if c.PixelFormat.BPP == 32 {
				rawPixel = byteOrder.Uint32(pixelBytes)
			}

			color := &colors[int(y)*int(rect.Width)+int(x)]
			if c.PixelFormat.TrueColor {
				color.R = uint16((rawPixel >> c.PixelFormat.RedShift) & uint32(c.PixelFormat.RedMax))
				color.G = uint16((rawPixel >> c.PixelFormat.GreenShift) & uint32(c.PixelFormat.GreenMax))
				color.B = uint16((rawPixel >> c.PixelFormat.BlueShift) & uint32(c.PixelFormat.BlueMax))
				if c.PixelFormat.BPP == 16 {
					color.B = color.B<<3 | color.B>>2
					color.G = color.G<<2 | color.G>>2
					color.R = color.R<<3 | color.R>>2
				}
			} else {
				*color = c.ColorMap[rawPixel]
			}
			rawPixels[int(y)*int(rect.Width)+int(x)] = uint32(0xff)<<24 | uint32(color.B)<<16 | uint32(color.G)<<8 | uint32(color.R)
		}
	}

	return &ZlibEncoding{Colors: colors, RawPixel: rawPixels}, nil
}

func (ze *ZlibEncoding) Close() {
	if ze.ZStream != nil {
		ze.ZStream = nil
		ze.ZReader.Close()
		ze.ZReader = nil
	}
}
