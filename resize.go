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
//     imgResized := resize.Resize(1000, -1, imgOld, Lanczos3)
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
func calcFactors(w, h int, wo, ho float32) (sx, sy float32) {
	if w <= 0 {
		w = -1
	}
	if h <= 0 {
		h = -1
	}

	if w == -1 {
		if h == -1 {
			sx = 1.0
			sy = 1.0
		} else {
			sy = ho / float32(h)
			sx = sy
		}
	} else {
		sx = wo / float32(w)
		if h == -1 {
			sy = sx
		} else {
			sy = ho / float32(h)
		}
	}
	return
}

// InterpolationFunction return a color for an arbitrary point inside
// an image
type InterpolationFunction func(float32, float32, image.Image) color.RGBA64

// Resize an image to new width w and height h using the interpolation function interp.
// A new image with the given dimensions will be returned.
// If one of the parameters w or h is set to -1, its size will be calculated so that
// the aspect ratio is that of the originating image.
// The resizing algorithm uses channels for parallel computation.
func Resize(w int, h int, img image.Image, interp InterpolationFunction) image.Image {
	b_old := img.Bounds()
	w_old := float32(b_old.Dx())
	h_old := float32(b_old.Dy())

	scaleX, scaleY := calcFactors(w, h, w_old, h_old)
	t := Trans2{scaleX, 0, float32(b_old.Min.X), 0, scaleY, float32(b_old.Min.Y)}

	m := image.NewRGBA64(image.Rect(0, 0, int(w_old/scaleX), int(h_old/scaleY)))
	b := m.Bounds()

	c := make(chan int, NCPU)
	for i := 0; i < NCPU; i++ {
		go func(b image.Rectangle, c chan int) {
			var u, v float32
			for y := b.Min.Y; y < b.Max.Y; y++ {
				for x := b.Min.X; x < b.Max.X; x++ {
					u, v = t.Eval(float32(x), float32(y))
					m.SetRGBA64(x, y, interp(u, v, img))
				}
			}
			c <- 1
		}(image.Rect(b.Min.X, b.Min.Y+i*(b.Dy())/4, b.Max.X, b.Min.Y+(i+1)*(b.Dy())/4), c)
	}

	for i := 0; i < NCPU; i++ {
		<-c
	}

	return m
}
