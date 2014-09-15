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

func Test_SameColor(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 20, 20))
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			img.SetRGBA(x, y, color.RGBA{0x80, 0x80, 0x80, 0xFF})
		}
	}
	out := Resize(10, 10, img, Lanczos3)
	for y := out.Bounds().Min.Y; y < out.Bounds().Max.Y; y++ {
		for x := out.Bounds().Min.X; x < out.Bounds().Max.X; x++ {
			color := img.At(x, y).(color.RGBA)
			if color.R != 0x80 || color.G != 0x80 || color.B != 0x80 || color.A != 0xFF {
				t.Fail()
			}
		}
	}
}

func Test_Bounds(t *testing.T) {
	img := image.NewRGBA(image.Rect(20, 10, 200, 99))
	out := Resize(80, 80, img, Lanczos2)
	out.At(0, 0)
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
