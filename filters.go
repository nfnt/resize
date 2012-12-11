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

// restrict an input float32 to the range of uint16 values
func clampToUint16(x float32) (y uint16) {
	y = uint16(x)
	if x < 0 {
		y = 0
	} else if x > float32(0xfffe) {
		// "else if x > float32(0xffff)" will cause overflows!
		y = 0xffff
	}
	return
}

func boolToUint(b bool) (i uint) {
	if b {
		i = 1
	} else {
		i = 0
	}

	return
}

// describe a resampling filter
type filterModel struct {
	// resampling is done by convolution with a (scaled) kernel
	kernel func(float32) float32

	// instead of blurring an image before downscaling to avoid aliasing,
	// to filter is scaled by a factor which leads to a similar effect
	factor [2]float32

	// for optimized access to image points
	converter

	// temporaries used by Interpolate
	tempRow, tempCol []colorArray
}

func (f *filterModel) convolution1d(x float32, p []colorArray, isCol bool) colorArray {
	var k float32
	var sum float32 = 0
	c := colorArray{0.0, 0.0, 0.0, 0.0}

	index := boolToUint(isCol)

	for j := range p {
		k = f.kernel((x - float32(j)) / f.factor[index])
		sum += k
		for i := range c {
			c[i] += p[j][i] * k
		}
	}

	// normalize values
	for i := range c {
		c[i] = c[i] / sum
	}
	return c
}

func (f *filterModel) Interpolate(x, y float32) color.RGBA64 {
	xf, yf := int(x)-len(f.tempRow)/2+1, int(y)-len(f.tempCol)/2+1
	x -= float32(xf)
	y -= float32(yf)

	for i := 0; i < len(f.tempCol); i++ {
		for j := 0; j < len(f.tempRow); j++ {
			f.tempRow[j] = f.at(xf+j, yf+i)
		}
		f.tempCol[i] = f.convolution1d(x, f.tempRow, false)
	}

	c := f.convolution1d(y, f.tempCol, true)
	return color.RGBA64{
		clampToUint16(c[0]),
		clampToUint16(c[1]),
		clampToUint16(c[2]),
		clampToUint16(c[3]),
	}
}

// createFilter tries to find an optimized converter for the given input image
// and initializes all filterModel members to their defaults
func createFilter(img image.Image, factor [2]float32, size int, kernel func(float32) float32) (f Filter) {
	sizeX := size * (int(math.Ceil(float64(factor[0]))))
	sizeY := size * (int(math.Ceil(float64(factor[1]))))

	switch img.(type) {
	default:
		f = &filterModel{
			kernel, factor,
			&genericConverter{img},
			make([]colorArray, sizeX), make([]colorArray, sizeY),
		}
	case *image.RGBA:
		f = &filterModel{
			kernel, factor,
			&rgbaConverter{img.(*image.RGBA)},
			make([]colorArray, sizeX), make([]colorArray, sizeY),
		}
	case *image.RGBA64:
		f = &filterModel{
			kernel, factor,
			&rgba64Converter{img.(*image.RGBA64)},
			make([]colorArray, sizeX), make([]colorArray, sizeY),
		}
	case *image.Gray:
		f = &filterModel{
			kernel, factor,
			&grayConverter{img.(*image.Gray)},
			make([]colorArray, sizeX), make([]colorArray, sizeY),
		}
	case *image.Gray16:
		f = &filterModel{
			kernel, factor,
			&gray16Converter{img.(*image.Gray16)},
			make([]colorArray, sizeX), make([]colorArray, sizeY),
		}
	case *image.YCbCr:
		f = &filterModel{
			kernel, factor,
			&ycbcrConverter{img.(*image.YCbCr)},
			make([]colorArray, sizeX), make([]colorArray, sizeY),
		}
	}
	return
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
	return createFilter(img, factor, 2, func(x float32) (y float32) {
		absX := float32(math.Abs(float64(x)))
		if absX <= 1 {
			y = 1 - absX
		} else {
			y = 0
		}
		return
	})
}

// Bicubic interpolation (with cubic hermite spline)
func Bicubic(img image.Image, factor [2]float32) Filter {
	return createFilter(img, factor, 4, func(x float32) (y float32) {
		absX := float32(math.Abs(float64(x)))
		if absX <= 1 {
			y = absX*absX*(1.5*absX-2.5) + 1
		} else if absX <= 2 {
			y = absX*(absX*(2.5-0.5*absX)-4) + 2
		} else {
			y = 0
		}
		return
	})
}

// Mitchell-Netravali interpolation
func MitchellNetravali(img image.Image, factor [2]float32) Filter {
	return createFilter(img, factor, 4, func(x float32) (y float32) {
		absX := float32(math.Abs(float64(x)))
		if absX <= 1 {
			y = absX*absX*(7*absX-12) + 16.0/3
		} else if absX <= 2 {
			y = -(absX - 2) * (absX - 2) / 3 * (7*absX - 8)
		} else {
			y = 0
		}
		return
	})
}

func lanczosKernel(a uint) func(float32) float32 {
	return func(x float32) (y float32) {
		if x > -float32(a) && x < float32(a) {
			y = float32(Sinc(float64(x))) * float32(Sinc(float64(x/float32(a))))
		} else {
			y = 0
		}
		return
	}
}

// Lanczos interpolation (a=2)
func Lanczos2(img image.Image, factor [2]float32) Filter {
	return createFilter(img, factor, 4, lanczosKernel(2))
}

// Lanczos interpolation (a=3)
func Lanczos3(img image.Image, factor [2]float32) Filter {
	return createFilter(img, factor, 6, lanczosKernel(3))
}
