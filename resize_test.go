package resize

import (
	"image"
	"image/color"
	"testing"
)

var img = image.NewGray16(image.Rect(0, 0, 3, 3))

func Test_Nearest(t *testing.T) {
	img.Set(1, 1, color.White)

	m, err := Resize(6, -1, img, NearestNeighbor)

	if err != nil || m.At(2, 2) != m.At(3, 3) {
		t.Fail()
	}
}

func Test_Param1(t *testing.T) {
	m, err := Resize(-1, -1, img, NearestNeighbor)
	if err != nil || m.Bounds() != img.Bounds() {
		t.Fail()
	}
}

func Test_Param2(t *testing.T) {
	_, err := Resize(-100, -1, img, NearestNeighbor)
	if err == nil {
		t.Fail()
	}
}

func Test_Param3(t *testing.T) {
	m, err := Resize(0, -1, img, NearestNeighbor)
	if err != nil || m.Bounds() != image.Rect(0, 0, 0, 0) {
		t.Fail()
	}
}

func Test_ZeroImg(t *testing.T) {
	zeroImg := image.NewGray16(image.Rect(0, 0, 0, 0))

	m, err := Resize(-1, -1, zeroImg, NearestNeighbor)
	if err != nil || m.Bounds() != zeroImg.Bounds() {
		t.Fail()
	}
}
