/*
Copyright (c) 2012, Jan Schlicht <jan.schlicht@gmail.com>

Permission to use, copy, modify, and/or distribute this software for any purpose
with or without fee is hereby granted, provided that the above copyright notice
and this permission notice appear in all copies.

THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY AND
FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS
OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER
TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF
THIS SOFTWARE.
*/

// Package resize implements various image resizing methods.
//
// The package works with the Image interface described in the image package.
// Various interpolation methods are provided and multiple processors may be
// utilized in the computations.
//
// Example:
//     imgResized := resize.Resize(1000, 0, imgOld, resize.MitchellNetravali)
package resize

import (
	"image"
	"image/color"
	"runtime"
)

// Trans2 is a 2-dimensional linear transformation.
type Trans2 [6]float32

// Apply the transformation to a point (x,y).
func (t *Trans2) Eval(x, y float32) (u, v float32) {
	u = t[0]*x + t[1]*y + t[2]
	v = t[3]*x + t[4]*y + t[5]
	return
}

// Filter can interpolate at points (x,y)
type Filter interface {
	Interpolate(x, y float32) color.RGBA64
}

// InterpolationFunction return a Filter implementation
// that operates on an image. Two factors
// allow to scale the filter kernels in x- and y-direction
// to prevent moire patterns.
type InterpolationFunction func(image.Image, [2]float32) Filter

// Resize an image to new width and height using the interpolation function interp.
// A new image with the given dimensions will be returned.
// If one of the parameters width or height is set to 0, its size will be calculated so that
// the aspect ratio is that of the originating image.
// The resizing algorithm uses channels for parallel computation.
func Resize(width, height uint, img image.Image, interp InterpolationFunction) image.Image {
	oldBounds := img.Bounds()
	oldWidth := float32(oldBounds.Dx())
	oldHeight := float32(oldBounds.Dy())

	scaleX, scaleY := calcFactors(width, height, oldWidth, oldHeight)
	t := Trans2{scaleX, 0, float32(oldBounds.Min.X), 0, scaleY, float32(oldBounds.Min.Y)}

	resizedImg := image.NewRGBA64(image.Rect(0, 0, int(0.7+oldWidth/scaleX), int(0.7+oldHeight/scaleY)))
	b := resizedImg.Bounds()
	adjustX := 0.5 * ((oldWidth-1.0)/scaleX - float32(b.Dx()-1))
	adjustY := 0.5 * ((oldHeight-1.0)/scaleY - float32(b.Dy()-1))

	n := numJobs(b.Dy())
	c := make(chan int, n)
	for i := 0; i < n; i++ {
		go func(b image.Rectangle, c chan int) {
			filter := interp(img, [2]float32{clampFactor(scaleX), clampFactor(scaleY)})
			var u, v float32
			var color color.RGBA64
			for y := b.Min.Y; y < b.Max.Y; y++ {
				for x := b.Min.X; x < b.Max.X; x++ {
					u, v = t.Eval(float32(x)+adjustX, float32(y)+adjustY)
					color = filter.Interpolate(u, v)

					i := resizedImg.PixOffset(x, y)
					resizedImg.Pix[i+0] = uint8(color.R >> 8)
					resizedImg.Pix[i+1] = uint8(color.R)
					resizedImg.Pix[i+2] = uint8(color.G >> 8)
					resizedImg.Pix[i+3] = uint8(color.G)
					resizedImg.Pix[i+4] = uint8(color.B >> 8)
					resizedImg.Pix[i+5] = uint8(color.B)
					resizedImg.Pix[i+6] = uint8(color.A >> 8)
					resizedImg.Pix[i+7] = uint8(color.A)
				}
			}
			c <- 1
		}(image.Rect(b.Min.X, b.Min.Y+i*(b.Dy())/n, b.Max.X, b.Min.Y+(i+1)*(b.Dy())/n), c)
	}

	for i := 0; i < n; i++ {
		<-c
	}

	return resizedImg
}

// Calculate scaling factors using old and new image dimensions.
func calcFactors(width, height uint, oldWidth, oldHeight float32) (scaleX, scaleY float32) {
	if width == 0 {
		if height == 0 {
			scaleX = 1.0
			scaleY = 1.0
		} else {
			scaleY = oldHeight / float32(height)
			scaleX = scaleY
		}
	} else {
		scaleX = oldWidth / float32(width)
		if height == 0 {
			scaleY = scaleX
		} else {
			scaleY = oldHeight / float32(height)
		}
	}
	return
}

// Set filter scaling factor to avoid moire patterns.
// This is only useful in case of downscaling (factor>1).
func clampFactor(factor float32) float32 {
	if factor < 1 {
		factor = 1
	}
	return factor
}

// Set number of parallel jobs
// but prevent resize from doing too much work
// if #CPUs > width
func numJobs(d int) (n int) {
	n = runtime.NumCPU()
	if n > d {
		n = d
	}
	return
}
