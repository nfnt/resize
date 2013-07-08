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
	factor [2]float32

	// for optimized access to image points
	converter

	// temporaries used by Interpolate
	tempRow, tempCol []colorArray
}

func (f *filterModel) convolution1d(x float32, p []colorArray, factor float32) colorArray {
	var k float32
	var sum float32 = 0
	c := colorArray{0.0, 0.0, 0.0, 0.0}

	for j := range p {
		k = f.kernel((x - float32(j)) / factor)
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

	for i := range f.tempCol {
		for j := range f.tempRow {
			f.tempRow[j] = f.at(xf+j, yf+i)
		}

		f.tempCol[i] = f.convolution1d(x, f.tempRow, f.factor[0])
	}

	c := f.convolution1d(y, f.tempCol, f.factor[1])
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

// Return a filter kernel that performs nearly identically to the provided
// kernel, but generates and uses a precomputed table rather than executing
// the kernel for each evaluation. The table is generated with tableSize
// values that cover the kernal domain from -maxX to +maxX. The input kernel
// is assumed to be symmetrical around 0, so the table only includes values
// from 0 to maxX.
func tableKernel(kernel func(float32) float32, tableSize int,
	maxX float32) func(float32) float32 {

	// precompute an array of filter coefficients
	weights := make([]float32, tableSize+1)
	for i := range weights {
		weights[i] = kernel(maxX * float32(i) / float32(tableSize))
	}
	weights[tableSize] = 0.0

	return func(x float32) float32 {
		if x < 0.0 {
			x = -x
		}
		indf := x / maxX * float32(tableSize)
		ind := int(indf)
		if ind >= tableSize {
			return 0.0
		}
		return weights[ind] + (weights[ind+1]-weights[ind])*(indf-float32(ind))
	}
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
	return createFilter(img, factor, 4, splineKernel(0, 0.5))
}

// Mitchell-Netravali interpolation
func MitchellNetravali(img image.Image, factor [2]float32) Filter {
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

const lanczosTableSize = 300

// Lanczos interpolation (a=2)
func Lanczos2(img image.Image, factor [2]float32) Filter {
	return createFilter(img, factor, 4, lanczosKernel(2))
}

// Lanczos interpolation (a=2) using a look-up table
// to speed up computation
func Lanczos2Lut(img image.Image, factor [2]float32) Filter {
	return createFilter(img, factor, 4,
		tableKernel(lanczosKernel(2), lanczosTableSize, 2.0))
}

// Lanczos interpolation (a=3)
func Lanczos3(img image.Image, factor [2]float32) Filter {
	return createFilter(img, factor, 6, lanczosKernel(3))
}

// Lanczos interpolation (a=3) using a look-up table
// to speed up computation
func Lanczos3Lut(img image.Image, factor [2]float32) Filter {
	return createFilter(img, factor, 6,
		tableKernel(lanczosKernel(3), lanczosTableSize, 3.0))
}
