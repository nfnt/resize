Resize
======

Image resizing for the [Go programming language](http://golang.org) that includes a few interpolation methods.

Installation
------------

```bash
$ go get github.com/nfnt/resize
```

It's that easy!

Usage
-----

Import package with

```go
import "github.com/nfnt/resize"
```

Resize creates a scaled image with new dimensions (`width`, `height`) using the interpolation function interp.
If either `width` or `height` is set to 0, it will be set to an aspect ratio preserving value.

```go
resize.Resize(width, height uint, img image.Image, interp resize.InterpolationFunction) image.Image, error 
```

The provided interpolation functions are

- `NearestNeighbor`: [Nearest-neighbor interpolation](http://en.wikipedia.org/wiki/Nearest-neighbor_interpolation)
- `Bilinear`: [Bilinear interpolation](http://en.wikipedia.org/wiki/Bilinear_interpolation)
- `Bicubic`: [Bicubic interpolation](http://en.wikipedia.org/wiki/Bicubic_interpolation)
- `Lanczos3`: [Lanczos resampling](http://en.wikipedia.org/wiki/Lanczos_resampling) with a=3

Sample usage:

```go
package main

import (
	"github.com/nfnt/resize"
	"image/jpeg"
	"os"
)

func main() {
	// open "test.jpg"
	file, err := os.Open("test.jpg")
	if err != nil {
		return
	}

	// decode jpeg into image.Image
	img, err := jpeg.Decode(file)
	if err != nil {
		return
	}
	file.Close()

	// resize to width 1000 using Lanczos resampling
	// and preserve aspect ration
	m := resize.Resize(1000, 0, img, resize.Lanczos3)

	out, err := os.Create("test_resized.jpg")
	if err != nil {
		return
	}
	defer out.Close()

	// write new image to file
	jpeg.Encode(out, m, nil)
}
```

License
-------

Copyright (c) 2012 Jan Schlicht <janschlicht@gmail.com>
Resize is released under an MIT style license.
