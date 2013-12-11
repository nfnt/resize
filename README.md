Resize
======

Image resizing for the [Go programming language](http://golang.org) with common interpolation methods.

[![Build Status](https://travis-ci.org/nfnt/resize.png)](https://travis-ci.org/nfnt/resize)

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

Resize creates a scaled image with new dimensions (`width`, `height`) using the interpolation function `interp`.
If either `width` or `height` is set to 0, it will be set to an aspect ratio preserving value.

```go
resize.Resize(width, height uint, img image.Image, interp resize.InterpolationFunction) image.Image 
```

Crop resizes the original image and crop it to fit `width` and `height`

```go
resize.Crop(width, height uint, img image.Image, interp resize.InterpolationFunction) image.Image 
```

The provided interpolation functions are (from fast to slow execution time)

- `NearestNeighbor`: [Nearest-neighbor interpolation](http://en.wikipedia.org/wiki/Nearest-neighbor_interpolation)
- `Bilinear`: [Bilinear interpolation](http://en.wikipedia.org/wiki/Bilinear_interpolation)
- `Bicubic`: [Bicubic interpolation](http://en.wikipedia.org/wiki/Bicubic_interpolation)
- `MitchellNetravali`: [Mitchell-Netravali interpolation](http://dl.acm.org/citation.cfm?id=378514)
- `Lanczos2Lut`: [Lanczos resampling](http://en.wikipedia.org/wiki/Lanczos_resampling) with a=2 using a look-up table for fast computation
- `Lanczos2`: [Lanczos resampling](http://en.wikipedia.org/wiki/Lanczos_resampling) with a=2
- `Lanczos3Lut`: [Lanczos resampling](http://en.wikipedia.org/wiki/Lanczos_resampling) with a=3 using a look-up table for fast computation
- `Lanczos3`: [Lanczos resampling](http://en.wikipedia.org/wiki/Lanczos_resampling) with a=3

Which of these methods gives the best results depends on your use case.

Sample usage:

```go
package main

import (
	"github.com/nfnt/resize"
	"image/jpeg"
	"log"
	"os"
)

func main() {
	// open "test.jpg"
	file, err := os.Open("test.jpg")
	if err != nil {
		log.Fatal(err)
	}

	// decode jpeg into image.Image
	img, err := jpeg.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	file.Close()

	// resize to width 1000 using Lanczos resampling
	// and preserve aspect ratio
	m := resize.Resize(1000, 0, img, resize.Lanczos3)

	out, err := os.Create("test_resized.jpg")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	// write new image to file
	jpeg.Encode(out, m, nil)
}
```

Downsizing Samples
-------

Downsizing is not as simple as it might look like. Images have to be filtered before they are scaled down, otherwise aliasing might occur.
Filtering is highly subjective: Applying too much will blur the whole image, too little will make aliasing become apparent.
Resize tries to provide sane defaults that should suffice in most cases.

### Artificial sample

Original image
![Rings](http://nfnt.github.com/img/rings_lg_orig.png)

<table>
<tr>
<th><img src="http://nfnt.github.com/img/rings_300_NearestNeighbor.png" /><br>Nearest-Neighbor</th>
<th><img src="http://nfnt.github.com/img/rings_300_Bilinear.png" /><br>Bilinear</th>
</tr>
<tr>
<th><img src="http://nfnt.github.com/img/rings_300_Bicubic.png" /><br>Bicubic</th>
<th><img src="http://nfnt.github.com/img/rings_300_MitchellNetravali.png" /><br>Mitchell-Netravali</th>
</tr>
<tr>
<th><img src="http://nfnt.github.com/img/rings_300_Lanczos2.png" /><br>Lanczos2</th>
<th><img src="http://nfnt.github.com/img/rings_300_Lanczos3.png" /><br>Lanczos3</th>
</tr>
</table>

### Real-Life sample

Original image  
![Original](http://nfnt.github.com/img/IMG_3694_720.jpg)

<table>
<tr>
<th><img src="http://nfnt.github.com/img/IMG_3694_300_NearestNeighbor.png" /><br>Nearest-Neighbor</th>
<th><img src="http://nfnt.github.com/img/IMG_3694_300_Bilinear.png" /><br>Bilinear</th>
</tr>
<tr>
<th><img src="http://nfnt.github.com/img/IMG_3694_300_Bicubic.png" /><br>Bicubic</th>
<th><img src="http://nfnt.github.com/img/IMG_3694_300_MitchellNetravali.png" /><br>Mitchell-Netravali</th>
</tr>
<tr>
<th><img src="http://nfnt.github.com/img/IMG_3694_300_Lanczos2.png" /><br>Lanczos2</th>
<th><img src="http://nfnt.github.com/img/IMG_3694_300_Lanczos3.png" /><br>Lanczos3</th>
</tr>
</table>


License
-------

Copyright (c) 2012 Jan Schlicht <janschlicht@gmail.com>
Resize is released under a MIT style license.
