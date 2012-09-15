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

package resize

import (
	"image"
	"image/color"
	"math"
)

// color.RGBA64 as array
type rgba16 [4]uint16

// build RGBA from an arbitrary color
func toRGBA(c color.Color) rgba16 {
	r, g, b, a := c.RGBA()
	return rgba16{uint16(r), uint16(g), uint16(b), uint16(a)}
}

func clampToUint16(x float32) (y uint16) {
	y = uint16(x)
	if x < 0 {
		y = 0
	} else if x > float32(0xfffe) {
		y = 0xffff
	}
	return
}

func convolution1d(x float32, kernel func(float32) float32, p []rgba16) (c rgba16) {
	x -= float32(int(x))
	
	m := float32(len(p)/2-1)

	var k float32
	var sum float32 = 0
	l := [4]float32{0.0, 0.0, 0.0, 0.0}

	for j := range p {
		k = kernel(x+m-float32(j))
		sum += k
		for i := range c {
			l[i] += float32(p[j][i]) * k
		}
	}
	for i := range c {
		c[i] = clampToUint16(l[i] / sum)
	}
	return
}

func filter(x, y float32, img image.Image, n int, kernel func(x float32) float32) color.RGBA64 {
	xf, yf := int(x)-n/2+1, int(y)-n/2+1

	row := make([]rgba16, n)
	col := make([]rgba16, n)

	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			row[j] = toRGBA(img.At(xf+j, yf+i))
		}
		col[i] = convolution1d(x, kernel, row)
	}

	c := convolution1d(y, kernel, col)
	return color.RGBA64{c[0], c[1], c[2], c[3]}
}

// Nearest-neighbor interpolation.
// Approximates a value by returning the value of the nearest point.
func NearestNeighbor(x, y float32, img image.Image) color.RGBA64 {
	n := 2
	kernel := func(x float32) (y float32) {
		if x >= -0.5 && x < 0.5 {
			y = 1
		} else {
			y = 0
		}
		return
	}
	return filter(x, y, img, n, kernel)
}

// Bicubic interpolation
func Bilinear(x, y float32, img image.Image) color.RGBA64 {
	n := 2
	kernel := func(x float32) float32 {
		return 1 - float32(math.Abs(float64(x)))
	}
	return filter(x, y, img, n, kernel)
}

// Bicubic interpolation
func Bicubic(x, y float32, img image.Image) color.RGBA64 {
	n := 4
	kernel := func(x float32) (y float32) {
		absX := float32(math.Abs(float64(x)))
		if absX <= 1 {
			y = absX*absX*(1.5*absX-2.5) + 1
		} else {
			y = absX*(absX*(2.5-0.5*absX)-4) + 2
		}
		return
	}
	return filter(x, y, img, n, kernel)
}

// Lanczos interpolation (a=2).
func Lanczos2(x, y float32, img image.Image) color.RGBA64 {
	n := 4
	kernel := func(x float32) float32 {
		return float32(Sinc(float64(x))) * float32(Sinc(float64((x)/float32(2))))
	}
	return filter(x, y, img, n, kernel)
}

// Lanczos interpolation (a=3).
func Lanczos3(x, y float32, img image.Image) color.RGBA64 {
	n := 6
	kernel := func(x float32) float32 {
		return float32(Sinc(float64(x))) * float32(Sinc(float64((x)/float32(3))))
	}
	return filter(x, y, img, n, kernel)
}
