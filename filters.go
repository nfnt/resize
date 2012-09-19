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

// build rgba16 from an arbitrary color
func toRgba16(c color.Color) rgba16 {
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

type filterModel struct {
	src              image.Image
	factor           [2]float32
	kernel           func(float32) float32
	tempRow, tempCol []rgba16
}

func (f *filterModel) convolution1d(x float32, p []rgba16, isRow bool) (c rgba16) {
	var k float32
	var sum float32 = 0
	l := [4]float32{0.0, 0.0, 0.0, 0.0}

	var index uint
	if isRow {
		index = 0
	} else {
		index = 1
	}

	for j := range p {
		k = f.kernel((x - float32(j)) / f.factor[index])
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

func (f *filterModel) Interpolate(x, y float32) color.RGBA64 {
	xf, yf := int(x)-len(f.tempRow)/2+1, int(y)-len(f.tempCol)/2+1
	x -= float32(xf)
	y -= float32(yf)

	for i := 0; i < len(f.tempCol); i++ {
		for j := 0; j < len(f.tempRow); j++ {
			f.tempRow[j] = toRgba16(f.src.At(xf+j, yf+i))
		}
		f.tempCol[i] = f.convolution1d(x, f.tempRow, true)
	}

	c := f.convolution1d(y, f.tempCol, false)
	return color.RGBA64{c[0], c[1], c[2], c[3]}
}

func createFilter(img image.Image, factor [2]float32, size int, kernel func(float32) float32) Filter {
	sizeX := size * (int(math.Ceil(float64(factor[0]))))
	sizeY := size * (int(math.Ceil(float64(factor[1]))))
	return &filterModel{img, factor, kernel, make([]rgba16, sizeX), make([]rgba16, sizeY)}
}

// Nearest-neighbor interpolation
func NearestNeighbor(img image.Image, factor [2]float32) Filter {
	return createFilter(img, factor, 2, func(x float32) (y float32) {
		if x >= -0.5 && x < 0.5 {
			y = 1
		} else {
			y = 0
		}
		return
	})
}

// Bilinear interpolation
func Bilinear(img image.Image, factor [2]float32) Filter {
	return createFilter(img, factor, 2, func(x float32) float32 {
		return 1 - float32(math.Abs(float64(x)))
	})
}

// Bicubic interpolation (with cubic hermite spline)
func Bicubic(img image.Image, factor [2]float32) Filter {
	return createFilter(img, factor, 4, func(x float32) (y float32) {
		absX := float32(math.Abs(float64(x)))
		if absX <= 1 {
			y = absX*absX*(1.5*absX-2.5) + 1
		} else {
			y = absX*(absX*(2.5-0.5*absX)-4) + 2
		}
		return
	})
}

func MitchellNetravali(img image.Image, factor [2]float32) Filter {
	return createFilter(img, factor, 4, func(x float32) (y float32) {
		absX := float32(math.Abs(float64(x)))
		if absX <= 1 {
			y = absX*absX*(7*absX-12) + 16.0/3
		} else {
			y = -(absX - 2) * (absX - 2) / 3 * (7*absX - 8)
		}
		return
	})
}

func lanczosKernel(a uint) func(float32) float32 {
	return func(x float32) float32 {
		return float32(Sinc(float64(x))) * float32(Sinc(float64(x/float32(a))))
	}
}

// Lanczos interpolation (a=2).
func Lanczos2(img image.Image, factor [2]float32) Filter {
	return createFilter(img, factor, 4, lanczosKernel(2))
}

// Lanczos interpolation (a=3).
func Lanczos3(img image.Image, factor [2]float32) Filter {
	return createFilter(img, factor, 6, lanczosKernel(3))
}
