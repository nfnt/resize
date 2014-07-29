package resize

import "testing"

func Test_FloatToUint8(t *testing.T) {
	var testData = []struct {
		in       float32
		expected uint8
	}{
		{0, 0},
		{255, 255},
		{128, 128},
		{1, 1},
		{256, 255},
	}
	for _, test := range testData {
		actual := floatToUint8(test.in)
		if actual != test.expected {
			t.Fail()
		}
	}
}

func Test_FloatToUint16(t *testing.T) {
	var testData = []struct {
		in       float32
		expected uint16
	}{
		{0, 0},
		{65535, 65535},
		{128, 128},
		{1, 1},
		{65536, 65535},
	}
	for _, test := range testData {
		actual := floatToUint16(test.in)
		if actual != test.expected {
			t.Fail()
		}
	}
}
