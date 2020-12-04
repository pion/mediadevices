package video

import (
	"image"
	"reflect"
	"testing"
)

func TestBroadcast(t *testing.T) {
	resolution := image.Rect(0, 0, 1920, 1080)
	img := image.NewGray(resolution)
	source := ReaderFunc(func() (image.Image, func(), error) {
		return img, func() {}, nil
	})

	broadcaster := NewBroadcaster(source, nil)
	readerWithoutCopy1 := broadcaster.NewReader(false)
	readerWithoutCopy2 := broadcaster.NewReader(false)
	actualWithoutCopy1, _, err := readerWithoutCopy1.Read()
	if err != nil {
		t.Fatal(err)
	}
	actualWithoutCopy2, _, err := readerWithoutCopy2.Read()
	if err != nil {
		t.Fatal(err)
	}

	if &actualWithoutCopy1.(*image.Gray).Pix[0] != &actualWithoutCopy2.(*image.Gray).Pix[0] {
		t.Fatal("Expected underlying buffer for frame with copy to be the same from broadcaster's buffer")
	}

	if !reflect.DeepEqual(img, actualWithoutCopy1) {
		t.Fatal("Expected actual frame without copy to be the same with the original")
	}

	readerWithCopy := broadcaster.NewReader(true)
	actualWithCopy, _, err := readerWithCopy.Read()
	if err != nil {
		t.Fatal(err)
	}

	if &actualWithCopy.(*image.Gray).Pix[0] == &actualWithoutCopy1.(*image.Gray).Pix[0] {
		t.Fatal("Expected underlying buffer for frame with copy to be different from broadcaster's buffer")
	}

	if !reflect.DeepEqual(img, actualWithCopy) {
		t.Fatal("Expected actual frame without copy to be the same with the original")
	}
}
