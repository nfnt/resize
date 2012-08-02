package resize

import (
	"fmt"
	"math"
	"testing"
)

const limit = 1e-12

func Test_SincOne(t *testing.T) {
	zero := Sinc(1)
	if zero >= limit {
		t.Error("Sinc(1) != 0")
	}
}

func Test_SincZero(t *testing.T) {
	one := Sinc(0)
	if math.Abs(one-1) >= limit {
		t.Error("Sinc(0) != 1")
	}
}

func Test_SincDotOne(t *testing.T) {
	res := Sinc(0.1)
	if math.Abs(res-0.983631643083466) >= limit {
		t.Error("Sinc(0.1) wrong")
	}
}

func Test_SincNearZero(t *testing.T) {
	res := Sinc(0.000001)
	if math.Abs(res-0.9999999999983551) >= limit {
		fmt.Println(res)
		t.Error("Sinc near zero not stable")
	}
}
