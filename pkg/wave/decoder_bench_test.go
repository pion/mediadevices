package wave

import (
	"encoding/binary"
	"fmt"
	"testing"
)

func BenchmarkDecoder(b *testing.B) {
	var nonHostEndian binary.ByteOrder
	if hostEndian == binary.BigEndian {
		nonHostEndian = binary.LittleEndian
	} else {
		nonHostEndian = binary.BigEndian
	}

	for format, decoder := range registeredDecoders {
		format := format
		decoder := decoder

		b.Run(fmt.Sprintf("%sHostEndian", format), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := decoder.Decode(hostEndian, make([]byte, 800), 2)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run(fmt.Sprintf("%sNonHostEndian", format), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := decoder.Decode(nonHostEndian, make([]byte, 800), 2)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
