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

Resize creates a scaled image with new dimensions (w,h) using the interpolation function interp.

```go
resize.Resize(w int, h int, img image.Image, interp resize.InterpolationFunction) image.Image 
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
	m := resize.Resize(1000, -1, img, resize.Lanczos3)

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
