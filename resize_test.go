package resize

import (
	"fmt"
	"image"
	"image/color"
	"runtime"
	"testing"
)

var img = image.NewGray16(image.Rect(0, 0, 3, 3))

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	img.Set(1, 1, color.White)
}

func Test_Nearest(t *testing.T) {
	m := Resize(6, 0, img, NearestNeighbor)
	if m.At(1, 1) == m.At(2, 2) {
		t.Fail()
	}
}

func Test_Param1(t *testing.T) {
	m := Resize(0, 0, img, NearestNeighbor)
	if m.Bounds() != img.Bounds() {
		t.Fail()
	}
}

func Test_Param2(t *testing.T) {
	m := Resize(100, 0, img, NearestNeighbor)
	if m.Bounds() != image.Rect(0, 0, 100, 100) {
		t.Fail()
	}
}

func Test_ZeroImg(t *testing.T) {
	zeroImg := image.NewGray16(image.Rect(0, 0, 0, 0))

	m := Resize(0, 0, zeroImg, NearestNeighbor)
	if m.Bounds() != zeroImg.Bounds() {
		t.Fail()
	}
}

func Test_CorrectResize(t *testing.T) {
	zeroImg := image.NewGray16(image.Rect(0, 0, 256, 256))

	m := Resize(60, 0, zeroImg, NearestNeighbor)
	if m.Bounds() != image.Rect(0, 0, 60, 60) {
		t.Fail()
	}
}

func Test_ResizeGamma(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 256, 256))

	// Alternating white and black pixels
	for y := 0; y < 256; y += 1 {
		for x := 0; x < 256; x += 1 {
			if (x+y)&1 == 0 {
				img.Set(x, y, color.White)
			} else {
				img.Set(x, y, color.Black)
			}
		}
	}

	// The expected result is 50% gray in energy-linear space.
	// In 8-bit sRGB, that is 186.
	expected := color.RGBA{188, 188, 188, 255}

	m := Resize(64, 0, img, NearestNeighbor)
	for y := 0; y < 64; y += 1 {
		for x := 0; x < 64; x += 1 {
			// Compare with 8-bit precision (as used by expected).
			r1, _, _, _ := expected.RGBA()
			r2, _, _, _ := m.At(x, y).RGBA()
			if r1>>8 != r2>>8 {
				fmt.Printf("expected: %d\n", r1)
				fmt.Printf("actual:  %d\n", r2)
				t.Fail()
				return
			}
		}
	}
}

func Test_ResizeAlpha(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 256, 256))

	// Alternating black and 100% transparent pixels
	for y := 0; y < 256; y += 1 {
		for x := 0; x < 256; x += 1 {
			if (x+y)&1 == 0 {
				img.Set(x, y, color.Black)
			} else {
				img.Set(x, y, color.Transparent)
			}
		}
	}

	// The expected result is 50% alpha with 0 RGB.
	expected := color.RGBA{0, 0, 0, 127}

	m := Resize(64, 0, img, NearestNeighbor)
	for y := 0; y < 64; y += 1 {
		for x := 0; x < 64; x += 1 {
			// Compare with 8-bit precision (as used by expected).
			r1, _, _, a1 := expected.RGBA()
			r2, _, _, a2 := m.At(x, y).RGBA()
			if r1>>8 != r2>>8 || a1>>8 != a2>>8 {
				fmt.Println("expected", r1, a1)
				fmt.Println("actual", r2, a2)
				r,g,b,a := color.Transparent.RGBA()
				fmt.Println(r,g,b,a)
				t.Fail()
				return
			}
		}
	}
}

func Benchmark_BigResize(b *testing.B) {
	var m image.Image
	for i := 0; i < b.N; i++ {
		m = Resize(1000, 1000, img, Lanczos3)
	}
	m.At(0, 0)
}

func Benchmark_Reduction(b *testing.B) {
	largeImg := image.NewRGBA(image.Rect(0, 0, 1000, 1000))

	var m image.Image
	for i := 0; i < b.N; i++ {
		m = Resize(300, 300, largeImg, Bicubic)
	}
	m.At(0, 0)
}
