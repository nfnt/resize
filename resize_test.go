package resize

import (
	"image"
	"image/color"
	"testing"
)

func Test_Nearest(t *testing.T) {
	img := image.NewGray16(image.Rect(0,0, 3,3))
	img.Set(1,1, color.White)
	
	m := Resize(6,-1, img, NearestNeighbor)
	
	if m.At(2,2) != m.At(3,3) {
		t.Fail()
	}
}
