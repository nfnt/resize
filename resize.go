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
//     imgResized := resize.Resize(1000, 0, imgOld, Lanczos3)
package resize

import (
	"image"
	"image/color"
	"runtime"
)

var (
	// NCPU holds the number of available CPUs at runtime.
	NCPU = runtime.NumCPU()
)

// Trans2 is a 2-dimensional linear transformation.
type Trans2 [6]float32

// Apply the transformation to a point (x,y).
func (t *Trans2) Eval(x, y float32) (u, v float32) {
	u = t[0]*x + t[1]*y + t[2]
	v = t[3]*x + t[4]*y + t[5]
	return
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

// InterpolationFunction return a color for an arbitrary point inside
// an image
type InterpolationFunction func(float32, float32, image.Image) color.RGBA64

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

	resizedImg := image.NewRGBA64(image.Rect(0, 0, int(oldWidth/scaleX), int(oldHeight/scaleY)))
	b := resizedImg.Bounds()

	// prevent resize from doing too much work
	// if #CPUs > width
	n := 1
	if NCPU < b.Dy() {
		n = NCPU
	} else {
		n = b.Dy()
	}

	c := make(chan int, n)
	for i := 0; i < n; i++ {
		go func(b image.Rectangle, c chan int) {
			var u, v float32
			for y := b.Min.Y; y < b.Max.Y; y++ {
				for x := b.Min.X; x < b.Max.X; x++ {
					u, v = t.Eval(float32(x), float32(y))
					resizedImg.SetRGBA64(x, y, interp(u, v, img))
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
