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

// describe a resampling filter
type filterModel struct {
	// resampling is done by convolution with a (scaled) kernel
	kernel func(float32) float32

	// instead of blurring an image before downscaling to avoid aliasing,
	// the filter is scaled by a factor which leads to a similar effect
	factorInv float32

	// for optimized access to image points
	converter

	// temporary used by Interpolate
	tempRow []colorArray

	kernelWeight []float32
	weightSum    float32
}

func (f *filterModel) SetKernelWeights(u float32) {
	uf := int(u) - len(f.tempRow)/2 + 1
	u -= float32(uf)
	f.weightSum = 0

	for j := range f.tempRow {
		f.kernelWeight[j] = f.kernel((u - float32(j)) * f.factorInv)
		f.weightSum += f.kernelWeight[j]
	}
}

func (f *filterModel) convolution1d() (c colorArray) {
	for j := range f.tempRow {
		for i := range c {
			c[i] += f.tempRow[j][i] * f.kernelWeight[j]
		}
	}

	// normalize values
	for i := range c {
		c[i] = c[i] / f.weightSum
	}

	return
}

func (f *filterModel) Interpolate(u float32, y int) color.RGBA64 {
	uf := int(u) - len(f.tempRow)/2 + 1
	u -= float32(uf)

	for i := range f.tempRow {
		f.at(uf+i, y, &f.tempRow[i])
	}

	c := f.convolution1d()
	return color.RGBA64{
		clampToUint16(c[0]),
		clampToUint16(c[1]),
		clampToUint16(c[2]),
		clampToUint16(c[3]),
	}
}

// createFilter tries to find an optimized converter for the given input image
// and initializes all filterModel members to their defaults
func createFilter(img image.Image, factor float32, size int, kernel func(float32) float32) (f Filter) {
	sizeX := size * (int(math.Ceil(float64(factor))))

	switch img.(type) {
	default:
		f = &filterModel{
			kernel, 1. / factor,
			&genericConverter{img},
			make([]colorArray, sizeX),
			make([]float32, sizeX),
			0,
		}
	case *image.RGBA:
		f = &filterModel{
			kernel, 1. / factor,
			&rgbaConverter{img.(*image.RGBA)},
			make([]colorArray, sizeX),
			make([]float32, sizeX),
			0,
		}
	case *image.RGBA64:
		f = &filterModel{
			kernel, 1. / factor,
			&rgba64Converter{img.(*image.RGBA64)},
			make([]colorArray, sizeX),
			make([]float32, sizeX),
			0,
		}
	case *image.Gray:
		f = &filterModel{
			kernel, 1. / factor,
			&grayConverter{img.(*image.Gray)},
			make([]colorArray, sizeX),
			make([]float32, sizeX),
			0,
		}
	case *image.Gray16:
		f = &filterModel{
			kernel, 1. / factor,
			&gray16Converter{img.(*image.Gray16)},
			make([]colorArray, sizeX),
			make([]float32, sizeX),
			0,
		}
	case *image.YCbCr:
		f = &filterModel{
			kernel, 1. / factor,
			&ycbcrConverter{img.(*image.YCbCr)},
			make([]colorArray, sizeX),
			make([]float32, sizeX),
			0,
		}
	}

	return
}

// Nearest-neighbor interpolation
func NearestNeighbor(img image.Image, factor float32) Filter {
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
func Bilinear(img image.Image, factor float32) Filter {
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
func Bicubic(img image.Image, factor float32) Filter {
	return createFilter(img, factor, 4, splineKernel(0, 0.5))
}

// Mitchell-Netravali interpolation
func MitchellNetravali(img image.Image, factor float32) Filter {
	return createFilter(img, factor, 4, splineKernel(1.0/3.0, 1.0/3.0))
}

func splineKernel(B, C float32) func(float32) float32 {
	factorA := 2.0 - 1.5*B - C
	factorB := -3.0 + 2.0*B + C
	factorC := 1.0 - 1.0/3.0*B
	factorD := -B/6.0 - C
	factorE := B + 5.0*C
	factorF := -2.0*B - 8.0*C
	factorG := 4.0/3.0*B + 4.0*C
	return func(x float32) (y float32) {
		absX := float32(math.Abs(float64(x)))
		if absX <= 1 {
			y = absX*absX*(factorA*absX+factorB) + factorC
		} else if absX <= 2 {
			y = absX*(absX*(absX*factorD+factorE)+factorF) + factorG
		} else {
			y = 0
		}

		return
	}
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
func Lanczos2(img image.Image, factor float32) Filter {
	return createFilter(img, factor, 4, lanczosKernel(2))
}

// Lanczos interpolation (a=3)
func Lanczos3(img image.Image, factor float32) Filter {
	return createFilter(img, factor, 6, lanczosKernel(3))
}
