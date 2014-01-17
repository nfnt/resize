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

type colorArray [4]float32

func replicateBorder1d(x, min, max int) int {
	if x < min {
		x = min
	} else if x >= max {
		x = max - 1
	}

	return x
}

func replicateBorder(x, y int, rect image.Rectangle) (xx, yy int) {
	xx = replicateBorder1d(x, rect.Min.X, rect.Max.X)
	yy = replicateBorder1d(y, rect.Min.Y, rect.Max.Y)
	return
}

// converter allows to retrieve a colorArray for points of an image.
// the idea is to speed up computation by providing optimized implementations
// for different image types instead of relying on image.Image.At().
type converter interface {
	at(x, y int, color *colorArray)
}

type genericConverter struct {
	src image.Image
}

func (c *genericConverter) at(x, y int, result *colorArray) {
	r, g, b, a := c.src.At(replicateBorder(x, y, c.src.Bounds())).RGBA()
	result[0] = float32(r)
	result[1] = float32(g)
	result[2] = float32(b)
	result[3] = float32(a)
	return
}

type rgbaConverter struct {
	src *image.RGBA
}

func (c *rgbaConverter) at(x, y int, result *colorArray) {
	i := c.src.PixOffset(replicateBorder(x, y, c.src.Rect))
	result[0] = float32(uint16(c.src.Pix[i+0])<<8 | uint16(c.src.Pix[i+0]))
	result[1] = float32(uint16(c.src.Pix[i+1])<<8 | uint16(c.src.Pix[i+1]))
	result[2] = float32(uint16(c.src.Pix[i+2])<<8 | uint16(c.src.Pix[i+2]))
	result[3] = float32(uint16(c.src.Pix[i+3])<<8 | uint16(c.src.Pix[i+3]))
	return
}

type rgba64Converter struct {
	src *image.RGBA64
}

func (c *rgba64Converter) at(x, y int, result *colorArray) {
	i := c.src.PixOffset(replicateBorder(x, y, c.src.Rect))
	result[0] = float32(uint16(c.src.Pix[i+0])<<8 | uint16(c.src.Pix[i+1]))
	result[1] = float32(uint16(c.src.Pix[i+2])<<8 | uint16(c.src.Pix[i+3]))
	result[2] = float32(uint16(c.src.Pix[i+4])<<8 | uint16(c.src.Pix[i+5]))
	result[3] = float32(uint16(c.src.Pix[i+6])<<8 | uint16(c.src.Pix[i+7]))
	return
}

type grayConverter struct {
	src *image.Gray
}

func (c *grayConverter) at(x, y int, result *colorArray) {
	i := c.src.PixOffset(replicateBorder(x, y, c.src.Rect))
	g := float32(uint16(c.src.Pix[i])<<8 | uint16(c.src.Pix[i]))
	result[0] = g
	result[1] = g
	result[2] = g
	result[3] = float32(0xffff)
	return
}

type gray16Converter struct {
	src *image.Gray16
}

func (c *gray16Converter) at(x, y int, result *colorArray) {
	i := c.src.PixOffset(replicateBorder(x, y, c.src.Rect))
	g := float32(uint16(c.src.Pix[i+0])<<8 | uint16(c.src.Pix[i+1]))
	result[0] = g
	result[1] = g
	result[2] = g
	result[3] = float32(0xffff)
	return
}

type ycbcrConverter struct {
	src *image.YCbCr
}

func (c *ycbcrConverter) at(x, y int, result *colorArray) {
	xx, yy := replicateBorder(x, y, c.src.Rect)
	yi := c.src.YOffset(xx, yy)
	ci := c.src.COffset(xx, yy)
	r, g, b := color.YCbCrToRGB(c.src.Y[yi], c.src.Cb[ci], c.src.Cr[ci])
	result[0] = float32(uint16(r) * 0x101)
	result[1] = float32(uint16(g) * 0x101)
	result[2] = float32(uint16(b) * 0x101)
	result[3] = float32(0xffff)
	return
}
