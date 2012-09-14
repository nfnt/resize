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
type RGBA [4]uint16

// build RGBA from an arbitrary color
func toRGBA(c color.Color) RGBA {
	r, g, b, a := c.RGBA()
	return RGBA{uint16(r), uint16(g), uint16(b), uint16(a)}
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

func convolution1d(x float32, kernel func(float32, int) float32, p []RGBA) (c RGBA) {
	x -= float32(int(x))

	var k float32
	var sum float32 = 0
	l := [4]float32{0.0, 0.0, 0.0, 0.0}

	for j := range p {
		k = kernel(x, j)
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

func filter(x, y float32, img image.Image, n int, kernel func(x float32, j int) float32) color.RGBA64 {
	xf, yf := int(x)-n/2+1, int(y)-n/2+1

	row := make([]RGBA, n)
	col := make([]RGBA, n)

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
	kernel := func(x float32, j int) (y float32) {
		if x+0.5 >= float32(j) && x+0.5 < float32(j)+1 {
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
	kernel := func(x float32, j int) float32 {
		xa := float32(math.Abs(float64(x - float32(j))))
		return 1 - xa
	}
	return filter(x, y, img, n, kernel)
}

// Bicubic interpolation
func Bicubic(x, y float32, img image.Image) color.RGBA64 {
	n := 4
	kernel := func(x float32, j int) (y float32) {
		xa := float32(math.Abs(float64(x - float32(j-1))))
		if xa <= 1 {
			y = 1.5*xa*xa*xa - 2.5*xa*xa + 1
		} else {
			y = -0.5*xa*xa*xa + 2.5*xa*xa - 4*xa + 2
		}
		return
	}
	return filter(x, y, img, n, kernel)
}

// Lanczos interpolation (a=2).
func Lanczos2(x, y float32, img image.Image) color.RGBA64 {
	n := 4
	kernel := func(x float32, j int) float32 {
		return float32(Sinc(float64(x-float32(j-1)))) * float32(Sinc(float64((x-float32(j-1))/float32(2))))
	}
	return filter(x, y, img, n, kernel)
}

// Lanczos interpolation (a=3).
func Lanczos3(x, y float32, img image.Image) color.RGBA64 {
	n := 6
	kernel := func(x float32, j int) float32 {
		return float32(Sinc(float64(x-float32(j-2)))) * float32(Sinc(float64((x-float32(j-2))/float32(3))))
	}
	return filter(x, y, img, n, kernel)
}
