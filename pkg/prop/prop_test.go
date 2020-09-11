package prop

import (
	"testing"
	"time"

	"github.com/pion/mediadevices/pkg/frame"
)

func TestCompareMatch(t *testing.T) {
	testDataSet := map[string]struct {
		a     MediaConstraints
		b     Media
		match bool
	}{
		"DeviceIDExactUnmatch": {
			MediaConstraints{
				DeviceID: StringExact("abc"),
			},
			Media{
				DeviceID: "cde",
			},
			false,
		},
		"DeviceIDExactMatch": {
			MediaConstraints{
				DeviceID: StringExact("abc"),
			},
			Media{
				DeviceID: "abc",
			},
			true,
		},
		"IntIdealUnmatch": {
			MediaConstraints{VideoConstraints: VideoConstraints{
				Width: Int(30),
			}},
			Media{Video: Video{
				Width: 50,
			}},
			true,
		},
		"IntIdealMatch": {
			MediaConstraints{VideoConstraints: VideoConstraints{
				Width: Int(30),
			}},
			Media{Video: Video{
				Width: 30,
			}},
			true,
		},
		"IntExactUnmatch": {
			MediaConstraints{VideoConstraints: VideoConstraints{
				Width: IntExact(30),
			}},
			Media{Video: Video{
				Width: 50,
			}},
			false,
		},
		"IntExactMatch": {
			MediaConstraints{VideoConstraints: VideoConstraints{
				Width: IntExact(30),
			}},
			Media{Video: Video{
				Width: 30,
			}},
			true,
		},
		"IntRangeUnmatch": {
			MediaConstraints{VideoConstraints: VideoConstraints{
				Width: IntRanged{Min: 30, Max: 40},
			}},
			Media{Video: Video{
				Width: 50,
			}},
			false,
		},
		"IntRangeMatch": {
			MediaConstraints{VideoConstraints: VideoConstraints{
				Width: IntRanged{Min: 30, Max: 40},
			}},
			Media{Video: Video{
				Width: 35,
			}},
			true,
		},
		"FrameFormatOneOfUnmatch": {
			MediaConstraints{VideoConstraints: VideoConstraints{
				FrameFormat: FrameFormatOneOf{frame.FormatYUYV, frame.FormatUYVY},
			}},
			Media{Video: Video{
				FrameFormat: frame.FormatYUYV,
			}},
			true,
		},
		"FrameFormatOneOfMatch": {
			MediaConstraints{VideoConstraints: VideoConstraints{
				FrameFormat: FrameFormatOneOf{frame.FormatYUYV, frame.FormatUYVY},
			}},
			Media{Video: Video{
				FrameFormat: frame.FormatMJPEG,
			}},
			false,
		},
		"DurationExactUnmatch": {
			MediaConstraints{AudioConstraints: AudioConstraints{
				Latency: DurationExact(time.Second),
			}},
			Media{Audio: Audio{
				Latency: time.Second + time.Millisecond,
			}},
			false,
		},
		"DurationExactMatch": {
			MediaConstraints{AudioConstraints: AudioConstraints{
				Latency: DurationExact(time.Second),
			}},
			Media{Audio: Audio{
				Latency: time.Second,
			}},
			true,
		},
		"DurationRangedUnmatch": {
			MediaConstraints{AudioConstraints: AudioConstraints{
				Latency: DurationRanged{Max: time.Second},
			}},
			Media{Audio: Audio{
				Latency: time.Second + time.Millisecond,
			}},
			false,
		},
		"DurationRangedMatch": {
			MediaConstraints{AudioConstraints: AudioConstraints{
				Latency: DurationRanged{Max: time.Second},
			}},
			Media{Audio: Audio{
				Latency: time.Millisecond,
			}},
			true,
		},
		"BoolExactUnmatch": {
			MediaConstraints{AudioConstraints: AudioConstraints{
				IsFloat: BoolExact(true),
			}},
			Media{Audio: Audio{
				IsFloat: false,
			}},
			false,
		},
		"BoolExactMatch": {
			MediaConstraints{AudioConstraints: AudioConstraints{
				IsFloat: BoolExact(true),
			}},
			Media{Audio: Audio{
				IsFloat: true,
			}},
			true,
		},
	}

	for name, testData := range testDataSet {
		testData := testData
		t.Run(name, func(t *testing.T) {
			_, match := testData.a.FitnessDistance(testData.b)
			if match != testData.match {
				t.Errorf("matching flag differs, expected: %v, got: %v", testData.match, match)
			}
		})
	}
}

func TestMergeWithZero(t *testing.T) {
	a := Media{
		Video: Video{
			Width: 30,
		},
	}

	b := Media{
		Video: Video{
			Height: 100,
		},
	}

	a.Merge(b)

	if a.Width == 0 {
		t.Error("expected a.Width to be 30, but got 0")
	}

	if a.Height == 0 {
		t.Error("expected a.Height to be 100, but got 0")
	}
}

func TestMergeWithSameField(t *testing.T) {
	a := Media{
		Video: Video{
			Width: 30,
		},
	}

	b := Media{
		Video: Video{
			Width: 100,
		},
	}

	a.Merge(b)

	if a.Width != 100 {
		t.Error("expected a.Width to be 100, but got 0")
	}
}

func TestMergeNested(t *testing.T) {
	type constraints struct {
		Media
	}

	a := constraints{
		Media{
			Video: Video{
				Width: 30,
			},
		},
	}

	b := Media{
		Video: Video{
			Width: 100,
		},
	}

	a.Merge(b)

	if a.Width != 100 {
		t.Error("expected a.Width to be 100, but got 0")
	}
}

func TestMergeConstraintsWithZero(t *testing.T) {
	a := Media{
		Video: Video{
			Width: 30,
		},
	}

	b := MediaConstraints{
		VideoConstraints: VideoConstraints{
			Height: Int(100),
		},
	}

	a.MergeConstraints(b)

	if a.Width == 0 {
		t.Error("expected a.Width to be 30, but got 0")
	}

	if a.Height == 0 {
		t.Error("expected a.Height to be 100, but got 0")
	}
}

func TestMergeConstraintsWithSameField(t *testing.T) {
	a := Media{
		Video: Video{
			Width: 30,
		},
	}

	b := MediaConstraints{
		VideoConstraints: VideoConstraints{
			Width: Int(100),
		},
	}

	a.MergeConstraints(b)

	if a.Width != 100 {
		t.Error("expected a.Width to be 100, but got 0")
	}
}

func TestMergeConstraintsNested(t *testing.T) {
	type constraints struct {
		Media
	}

	a := constraints{
		Media{
			Video: Video{
				Width: 30,
			},
		},
	}

	b := MediaConstraints{
		VideoConstraints: VideoConstraints{
			Width: Int(100),
		},
	}

	a.MergeConstraints(b)

	if a.Width != 100 {
		t.Error("expected a.Width to be 100, but got 0")
	}
}

func TestString(t *testing.T) {
	t.Run("IdealValues", func(t *testing.T) {
		t.Log("\n", &MediaConstraints{
			DeviceID: String("one"),
			VideoConstraints: VideoConstraints{
				Width:       Int(1920),
				FrameRate:   Float(30.0),
				FrameFormat: FrameFormat(frame.FormatI420),
			},
			AudioConstraints: AudioConstraints{
				Latency: Duration(time.Millisecond * 20),
			},
		})
	})

	t.Run("ExactValues", func(t *testing.T) {
		t.Log("\n", &MediaConstraints{
			DeviceID: StringExact("one"),
			VideoConstraints: VideoConstraints{
				Width:       IntExact(1920),
				FrameRate:   FloatExact(30.0),
				FrameFormat: FrameFormatExact(frame.FormatI420),
			},
			AudioConstraints: AudioConstraints{
				Latency:     DurationExact(time.Millisecond * 20),
				IsBigEndian: BoolExact(true),
			},
		})
	})

	t.Run("OneOfValues", func(t *testing.T) {
		t.Log("\n", &MediaConstraints{
			DeviceID: StringOneOf{"one", "two"},
			VideoConstraints: VideoConstraints{
				Width:       IntOneOf{1920, 1080},
				FrameRate:   FloatOneOf{30.0, 60.1234},
				FrameFormat: FrameFormatOneOf{frame.FormatI420, frame.FormatI444},
			},
			AudioConstraints: AudioConstraints{
				Latency: DurationOneOf{time.Millisecond * 20, time.Millisecond * 40},
			},
		})
	})

	t.Run("RangedValues", func(t *testing.T) {
		t.Log("\n", &MediaConstraints{
			VideoConstraints: VideoConstraints{
				Width:     &IntRanged{Min: 1080, Max: 1920, Ideal: 1500},
				FrameRate: &FloatRanged{Min: 30.123, Max: 60.12321312, Ideal: 45.12312312},
			},
			AudioConstraints: AudioConstraints{
				Latency: &DurationRanged{Min: time.Millisecond * 20, Max: time.Millisecond * 40, Ideal: time.Millisecond * 30},
			},
		})
	})
}
