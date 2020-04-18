package wave

import (
	"math"
	"reflect"
	"testing"
)

func TestConvert(t *testing.T) {
	cases := map[string]struct {
		in       []Sample
		typ      SampleFormat
		expected []Sample
	}{
		"Int16ToFloat32": {
			in: []Sample{
				Int16Sample(-0x1000),
				Int16Sample(-0x100),
				Int16Sample(0x0),
				Int16Sample(0x100),
				Int16Sample(0x1000),
			},
			typ: Float32SampleFormat,
			expected: []Sample{
				Float32Sample(-math.Pow(2, -4)),
				Float32Sample(-math.Pow(2, -8)),
				Float32Sample(0.0),
				Float32Sample(math.Pow(2, -8)),
				Float32Sample(math.Pow(2, -4)),
			},
		},
		"Float32ToInt16": {
			in: []Sample{
				Float32Sample(-math.Pow(2, -4)),
				Float32Sample(-math.Pow(2, -8)),
				Float32Sample(0.0),
				Float32Sample(math.Pow(2, -8)),
				Float32Sample(math.Pow(2, -4)),
			},
			typ: Int16SampleFormat,
			expected: []Sample{
				Int16Sample(-0x1000),
				Int16Sample(-0x100),
				Int16Sample(0x0),
				Int16Sample(0x100),
				Int16Sample(0x1000),
			},
		},
	}
	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			for i := range c.in {
				s := c.typ.Convert(c.in[i])
				if !reflect.DeepEqual(c.expected[i], s) {
					t.Errorf("Convert result differs, expected: %v, got: %v", c.expected[i], s)
				}
			}
		})
	}
}
