package prop

import (
	"testing"
)

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
