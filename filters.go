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
	n := color.RGBA64Model.Convert(c).(color.RGBA64)
	return RGBA{n.R, n.G, n.B, n.A}
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
	xn, yn := int(x), int(y)
	c := toRGBA(img.At(xn, yn))
	return color.RGBA64{c[0], c[1], c[2], c[3]}
}

// Linear interpolation.
func linearInterp(x float32, p *[2]RGBA) (c RGBA) {
	x -= float32(math.Floor(float64(x)))
	for i := range c {
		c[i] = clampToUint16(float32(p[0][i])*(1.0-x) + x*float32(p[1][i]))
	}
	return
}

// Bilinear interpolation.
func Bilinear(x, y float32, img image.Image) color.RGBA64 {
	xf, yf := int(math.Floor(float64(x))), int(math.Floor(float64(y)))

	var row [2]RGBA
	var col [2]RGBA
	row = [2]RGBA{toRGBA(img.At(xf, yf)), toRGBA(img.At(xf+1, yf))}
	col[0] = linearInterp(x, &row)
	row = [2]RGBA{toRGBA(img.At(xf, yf+1)), toRGBA(img.At(xf+1, yf+1))}
	col[1] = linearInterp(x, &row)

	c := linearInterp(y, &col)
	return color.RGBA64{c[0], c[1], c[2], c[3]}
}

// cubic interpolation
func cubicInterp(x float32, p *[4]RGBA) (c RGBA) {
	x -= float32(math.Floor(float64(x)))
	for i := range c {
		c[i] = clampToUint16(float32(p[1][i]) + 0.5*x*(float32(p[2][i])-float32(p[0][i])+x*(2.0*float32(p[0][i])-5.0*float32(p[1][i])+4.0*float32(p[2][i])-float32(p[3][i])+x*(3.0*(float32(p[1][i])-float32(p[2][i]))+float32(p[3][i])-float32(p[0][i])))))
	}
	return
}

// Bicubic interpolation.
func Bicubic(x, y float32, img image.Image) color.RGBA64 {
	xf, yf := int(math.Floor(float64(x))), int(math.Floor(float64(y)))

	var row [4]RGBA
	var col [4]RGBA
	row = [4]RGBA{toRGBA(img.At(xf-1, yf-1)), toRGBA(img.At(xf, yf-1)), toRGBA(img.At(xf+1, yf-1)), toRGBA(img.At(xf+2, yf-1))}
	col[0] = cubicInterp(x, &row)
	row = [4]RGBA{toRGBA(img.At(xf-1, yf)), toRGBA(img.At(xf, yf)), toRGBA(img.At(xf+1, yf)), toRGBA(img.At(xf+2, yf))}
	col[1] = cubicInterp(x, &row)
	row = [4]RGBA{toRGBA(img.At(xf-1, yf+1)), toRGBA(img.At(xf, yf+1)), toRGBA(img.At(xf+1, yf+1)), toRGBA(img.At(xf+2, yf+1))}
	col[2] = cubicInterp(x, &row)
	row = [4]RGBA{toRGBA(img.At(xf-1, yf+2)), toRGBA(img.At(xf, yf+2)), toRGBA(img.At(xf+1, yf+2)), toRGBA(img.At(xf+2, yf+2))}
	col[3] = cubicInterp(x, &row)

	c := cubicInterp(y, &col)
	return color.RGBA64{c[0], c[1], c[2], c[3]}
}

// 1-d convolution with windowed sinc for a=3.
func lanczos_x(x float32, p *[6]RGBA) (c RGBA) {
	x -= float32(math.Floor(float64(x)))
	var v float32
	l := [4]float32{0.0, 0.0, 0.0, 0.0}
	for j := range p {
		v = float32(Sinc(float64(x-float32(j-2)))) * float32(Sinc(float64((x-float32(j-2))/3.0)))
		for i := range c {
			l[i] += float32(p[j][i]) * v
		}
	}
	for i := range c {
		c[i] = clampToUint16(l[i])
	}
	return
}

// Lanczos interpolation (a=3).
func Lanczos3(x, y float32, img image.Image) color.RGBA64 {
	xf, yf := int(math.Floor(float64(x))), int(math.Floor(float64(y)))

	var row [6]RGBA
	var col [6]RGBA
	row = [6]RGBA{toRGBA(img.At(xf-2, yf-2)), toRGBA(img.At(xf-1, yf-2)), toRGBA(img.At(xf, yf-2)), toRGBA(img.At(xf+1, yf-2)), toRGBA(img.At(xf+2, yf-2)), toRGBA(img.At(xf+3, yf-2))}
	col[0] = lanczos_x(x, &row)
	row = [6]RGBA{toRGBA(img.At(xf-2, yf-1)), toRGBA(img.At(xf-1, yf-1)), toRGBA(img.At(xf, yf-1)), toRGBA(img.At(xf+1, yf-1)), toRGBA(img.At(xf+2, yf-1)), toRGBA(img.At(xf+3, yf-1))}
	col[1] = lanczos_x(x, &row)
	row = [6]RGBA{toRGBA(img.At(xf-2, yf)), toRGBA(img.At(xf-1, yf)), toRGBA(img.At(xf, yf)), toRGBA(img.At(xf+1, yf)), toRGBA(img.At(xf+2, yf)), toRGBA(img.At(xf+3, yf))}
	col[2] = lanczos_x(x, &row)
	row = [6]RGBA{toRGBA(img.At(xf-2, yf+1)), toRGBA(img.At(xf-1, yf+1)), toRGBA(img.At(xf, yf+1)), toRGBA(img.At(xf+1, yf+1)), toRGBA(img.At(xf+2, yf+1)), toRGBA(img.At(xf+3, yf+1))}
	col[3] = lanczos_x(x, &row)
	row = [6]RGBA{toRGBA(img.At(xf-2, yf+2)), toRGBA(img.At(xf-1, yf+2)), toRGBA(img.At(xf, yf+2)), toRGBA(img.At(xf+1, yf+2)), toRGBA(img.At(xf+2, yf+2)), toRGBA(img.At(xf+3, yf+2))}
	col[4] = lanczos_x(x, &row)
	row = [6]RGBA{toRGBA(img.At(xf-2, yf+3)), toRGBA(img.At(xf-1, yf+3)), toRGBA(img.At(xf, yf+3)), toRGBA(img.At(xf+1, yf+3)), toRGBA(img.At(xf+2, yf+3)), toRGBA(img.At(xf+3, yf+3))}
	col[5] = lanczos_x(x, &row)

	c := lanczos_x(y, &col)
	return color.RGBA64{c[0], c[1], c[2], c[3]}
}
