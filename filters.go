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
	} else if x > float32(0xffff) {
		y = 0xffff
	}
	return
}

// Nearest-neighbor interpolation.
// Approximates a value by returning the value of the nearest point.
func NearestNeighbor(x, y float32, img image.Image) color.RGBA64 {
	xn, yn := int(float32(int(x))+0.5), int(float32(int(y))+0.5)
	c := toRGBA(img.At(xn, yn))
	return color.RGBA64{c[0], c[1], c[2], c[3]}
}

// Linear interpolation.
func linearInterp(x float32, p *[2]RGBA) (c RGBA) {
	x -= float32(int(x))
	for i := range c {
		c[i] = clampToUint16(float32(p[0][i])*(1.0-x) + x*float32(p[1][i]))
	}
	return
}

// Bilinear interpolation.
func Bilinear(x, y float32, img image.Image) color.RGBA64 {
	xf, yf := int(x), int(y)

	var row [2]RGBA
	var col [2]RGBA
	for i := 0; i < 2; i++ {
		row = [2]RGBA{toRGBA(img.At(xf, yf+i)), toRGBA(img.At(xf+1, yf+i))}
		col[i] = linearInterp(x, &row)
	}

	c := linearInterp(y, &col)
	return color.RGBA64{c[0], c[1], c[2], c[3]}
}

// cubic interpolation
func cubicInterp(x float32, p *[4]RGBA) (c RGBA) {
	x -= float32(int(x))
	for i := range c {
		c[i] = clampToUint16(float32(p[1][i]) + 0.5*x*(float32(p[2][i])-float32(p[0][i])+x*(2.0*float32(p[0][i])-5.0*float32(p[1][i])+4.0*float32(p[2][i])-float32(p[3][i])+x*(3.0*(float32(p[1][i])-float32(p[2][i]))+float32(p[3][i])-float32(p[0][i])))))
	}
	return
}

// Bicubic interpolation.
func Bicubic(x, y float32, img image.Image) color.RGBA64 {
	xf, yf := int(x), int(y)

	var row [4]RGBA
	var col [4]RGBA
	for i := 0; i < 4; i++ {
		row = [4]RGBA{toRGBA(img.At(xf-1, yf+i-1)), toRGBA(img.At(xf, yf+i-1)), toRGBA(img.At(xf+1, yf+i-1)), toRGBA(img.At(xf+2, yf+i-1))}
		col[i] = cubicInterp(x, &row)
	}

	c := cubicInterp(y, &col)
	return color.RGBA64{c[0], c[1], c[2], c[3]}
}

// 1-d convolution with windowed sinc for a=2.
func lanczos2_x(x float32, p *[4]RGBA) (c RGBA) {
	x -= float32(int(x))

	var kernel float32
	var sum float32 = 0 // for kernel normalization
	l := [4]float32{0.0, 0.0, 0.0, 0.0}

	for j := range p {
		kernel = float32(Sinc(float64(x-float32(j-1)))) * float32(Sinc(float64((x-float32(j-1))/2.0)))
		sum += kernel
		for i := range c {
			l[i] += float32(p[j][i]) * kernel
		}
	}
	for i := range c {
		c[i] = clampToUint16(l[i] / sum)
	}
	return
}

// Lanczos interpolation (a=2).
func Lanczos2(x, y float32, img image.Image) color.RGBA64 {
	xf, yf := int(x), int(y)

	var row [4]RGBA
	var col [4]RGBA
	for i := range row {
		row = [4]RGBA{toRGBA(img.At(xf-1, yf+i-1)), toRGBA(img.At(xf, yf+i-1)), toRGBA(img.At(xf+1, yf+i-1)), toRGBA(img.At(xf+2, yf+i-1))}
		col[i] = lanczos2_x(x, &row)
	}

	c := lanczos2_x(y, &col)
	return color.RGBA64{c[0], c[1], c[2], c[3]}
}

// 1-d convolution with windowed sinc for a=3.
func lanczos3_x(x float32, p *[6]RGBA) (c RGBA) {
	x -= float32(int(x))

	var kernel float32
	var sum float32 = 0 // for kernel normalization
	l := [4]float32{0.0, 0.0, 0.0, 0.0}

	for j := range p {
		kernel = float32(Sinc(float64(x-float32(j-2)))) * float32(Sinc(float64((x-float32(j-2))/3.0)))
		sum += kernel
		for i := range c {
			l[i] += float32(p[j][i]) * kernel
		}
	}
	for i := range c {
		c[i] = clampToUint16(l[i] / sum)
	}
	return
}

// Lanczos interpolation (a=3).
func Lanczos3(x, y float32, img image.Image) color.RGBA64 {
	xf, yf := int(x), int(y)

	var row [6]RGBA
	var col [6]RGBA
	for i := range row {
		row = [6]RGBA{toRGBA(img.At(xf-2, yf+i-2)), toRGBA(img.At(xf-1, yf+i-2)), toRGBA(img.At(xf, yf+i-2)), toRGBA(img.At(xf+1, yf+i-2)), toRGBA(img.At(xf+2, yf+i-2)), toRGBA(img.At(xf+3, yf+i-2))}
		col[i] = lanczos3_x(x, &row)
	}

	c := lanczos3_x(y, &col)
	return color.RGBA64{c[0], c[1], c[2], c[3]}
}
