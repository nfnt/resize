package resize

import (
	"image"
	"image/draw"
)

func Crop(width, height uint, img image.Image, interp InterpolationFunction) image.Image {
	ob := img.Bounds()
	var w, h uint
	rx := float64(ob.Dx()) / float64(width)
	ry := float64(ob.Dy()) / float64(height)

	if rx < ry {
		w = width
		h = uint(float64(ob.Dy()) / rx)
	} else {
		w = uint(float64(ob.Dx()) / ry)
		h = height
	}

	buf := Resize(w, h, img, interp)
	r := image.Rect(0, 0, int(width), int(height))
	dst := image.NewRGBA64(r)
	var pt image.Point
	if rx < ry {
		pt.Y = (int(h) - int(height)) / 2
	} else {
		pt.X = (int(w) - int(width)) / 2
	}

	draw.Draw(dst, r, buf, pt, draw.Src)

	return dst
}
