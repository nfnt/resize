Resize
======

Image resizing for the [Go programming language](http://golang.org) that provides a few interpolation methods.

Installation
============

```bash
$ go get github.com/nfnt/resize
```

It's that easy!

Usage
=====

Import package with

```go
import "github.com/nfnt/resize"
```

Resize creates a scaled image with new dimensions (w,h) using the interpolation function interp.

```go
resize.Resize(w int, h int, img image.Image, interp InterpolationFunction) image.Image 
```

The provided interpolation functions are

- NearestNeighbor: Nearest-neighbor interpolation
- Bilinear: Bilinear interpolation
- Bicubic: Bicubic interpolation
- Lanczos3: Convolution with windowed Sinc function, a=3

License
=======

Copyright (c) 2012 Jan Schlicht <janschlicht@gmail.com>
This software is released unter the ISC license.
The license text is at <http://www.isc.org/software/license>
