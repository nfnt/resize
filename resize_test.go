package resize

import (
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

func Benchmark_BigResizeLanczos3(b *testing.B) {
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

// Benchmark resize of 16 MPix jpeg image to 800px width.
func jpegThumb(b *testing.B, interp InterpolationFunction) {
	input := image.NewYCbCr(image.Rect(0, 0, 4896, 3264), image.YCbCrSubsampleRatio422)

	var output image.Image
	for i := 0; i < b.N; i++ {
		output = Resize(800, 0, input, interp)
	}

	output.At(0, 0)
}

func Benchmark_LargeJpegThumbNearestNeighbor(b *testing.B) {
	jpegThumb(b, NearestNeighbor)
}

func Benchmark_LargeJpegThumbBilinear(b *testing.B) {
	jpegThumb(b, Bilinear)
}

func Benchmark_LargeJpegThumbBicubic(b *testing.B) {
	jpegThumb(b, Bicubic)
}

func Benchmark_LargeJpegThumbMitchellNetravali(b *testing.B) {
	jpegThumb(b, MitchellNetravali)
}

func Benchmark_LargeJpegThumbLanczos2(b *testing.B) {
	jpegThumb(b, Lanczos2)
}

func Benchmark_LargeJpegThumbLanczos3(b *testing.B) {
	jpegThumb(b, Lanczos3)
}
