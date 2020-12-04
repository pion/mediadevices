package wave

// Int64Sample is a 64-bits signed integer audio sample.
type Int64Sample int64

func (s Int64Sample) Int() int64 {
	return int64(s)
}
