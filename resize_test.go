package resize

import (
	"image"
	"image/color"
	"testing"
)

var img = image.NewGray16(image.Rect(0, 0, 3, 3))

func Test_Nearest(t *testing.T) {
	img.Set(1, 1, color.White)

	m := Resize(6, -1, img, NearestNeighbor)

	if m.At(2, 2) != m.At(3, 3) {
		t.Fail()
	}
}

func Test_Param1(t *testing.T) {
	m := Resize(-1, -1, img, NearestNeighbor)
	if m.Bounds() != img.Bounds() {
		t.Fail()
	}
}

func Test_Param2(t *testing.T) {
	m := Resize(-100, -1, img, NearestNeighbor)
	if m.Bounds() != img.Bounds() {
		t.Fail()
	}
}

func Test_Param3(t *testing.T) {
	m := Resize(0, -1, img, NearestNeighbor)
	if m.Bounds() != img.Bounds() {
		t.Fail()
	}
}

func Test_ZeroImg(t *testing.T) {
	zeroImg := image.NewGray16(image.Rect(0, 0, 0, 0))

	m := Resize(-1, -1, zeroImg, NearestNeighbor)
	if m.Bounds() != zeroImg.Bounds() {
		t.Fail()
	}
}
